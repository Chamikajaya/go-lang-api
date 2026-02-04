package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	database "user-management-api/db/sqlc"
	_ "user-management-api/docs" // Swagger generated docs
	"user-management-api/internal/config"
	"user-management-api/internal/handlers"
	"user-management-api/internal/middleware"
	"user-management-api/internal/service"
	"user-management-api/internal/validator"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	pool, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close() // Close when main() exits

	log.Println("Successfully connected to database")

	// Initialize dependencies
	queries := database.New(pool)
	userService := service.NewUserService(pool, queries)
	validatorInstance := validator.NewValidator()
	userHandler := handlers.NewUserHandler(userService, validatorInstance)

	// Setup router
	router := setupRouter(userHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	gracefulShutdown(server)
}

func connectDB(cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func setupRouter(userHandler *handlers.UserHandler) *chi.Mux {
	// Create new Chi router
	r := chi.NewRouter()

	// Global middleware (applies to all routes)
	r.Use(chimiddleware.RequestID)   // Adds request ID for tracing
	r.Use(middleware.Logger)         // custom logger
	r.Use(middleware.Recovery)       // Recover from panics
	r.Use(middleware.CORS)           // CORS headers
	r.Use(middleware.ContentTypeJSON) // Set JSON content type
	r.Use(chimiddleware.Timeout(60 * time.Second)) // Request timeout

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Swagger documentation
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/docs/doc.json"),
	))

	// API routes under /api/v1
	r.Route("/api/v1", func(r chi.Router) {
		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.CreateUser)       // POST /api/v1/users
			r.Get("/", userHandler.ListUsers)         // GET /api/v1/users
			r.Get("/{id}", userHandler.GetUser)       // GET /api/v1/users/{id}
			r.Patch("/{id}", userHandler.UpdateUser)  // PATCH /api/v1/users/{id}
			r.Delete("/{id}", userHandler.DeleteUser) // DELETE /api/v1/users/{id}
		})
	})

	return r
}

// gracefulShutdown handles graceful shutdown on SIGINT/SIGTERM
func gracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}