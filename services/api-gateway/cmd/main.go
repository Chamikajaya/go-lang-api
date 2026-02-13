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

	_ "api-gateway/docs" // Swagger generated docs
	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/middleware"
	"api-gateway/internal/nats/client"
	"api-gateway/internal/nats/subscriber"
	ws "api-gateway/internal/websocket"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title User Management API
// @version 2.0
// @description API Gateway for User Management Microservices
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @schemes http https
func main() {
	cfg := config.LoadConfig()

	nc, err := connectNATS(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Successfully connected to NATS")

	// Initialize NATS RPC client
	rpcClient := client.NewClient(nc, cfg.RPCTimeout)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(rpcClient)

	// Initialize WebSocket manager
	wsManager := ws.NewManager()
	go wsManager.Run()
	log.Println("WebSocket manager started")

	// Initialize and start NATS event subscriber (bridges NATS events → WebSocket)
	eventSubscriber := subscriber.NewSubscriber(nc, wsManager)
	if err := eventSubscriber.Start(); err != nil {
		log.Fatalf("Failed to start event subscriber: %v", err)
	}
	log.Println("NATS event subscriber started")

	// Setup router
	router := setupRouter(userHandler, wsManager)

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
		log.Printf("API Gateway starting on port %s", cfg.ServerPort)
		log.Printf("Swagger docs available at http://localhost:%s/swagger/index.html", cfg.ServerPort)
		log.Printf("WebSocket endpoint available at ws://localhost:%s/ws?userId={uuid}", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	gracefulShutdown(server, nc, eventSubscriber)
}

func connectNATS(cfg *config.Config) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name("api-gateway"),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1), // Unlimited reconnects
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Printf("NATS disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			log.Printf("NATS error: %v", err)
		}),
	}

	nc, err := nats.Connect(cfg.NatsURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return nc, nil
}

func setupRouter(userHandler *handlers.UserHandler, wsManager *ws.Manager) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS)

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// WebSocket endpoint
	r.Get("/ws", ws.HandleWebSocket(wsManager))

	// REST API routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.ContentTypeJSON)

		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.CreateUser)
			r.Get("/", userHandler.ListUsers)
			r.Get("/{id}", userHandler.GetUser)
			r.Patch("/{id}", userHandler.UpdateUser)
			r.Delete("/{id}", userHandler.DeleteUser)
		})
	})

	return r
}

func gracefulShutdown(server *http.Server, nc *nats.Conn, eventSub *subscriber.Subscriber) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Received signal: %v. Shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop event subscriber first
	if err := eventSub.Stop(); err != nil {
		log.Printf("Error stopping event subscriber: %v", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	if err := nc.Drain(); err != nil {
		log.Printf("Error draining NATS connection: %v", err)
	}

	log.Println("API Gateway stopped gracefully")
}
