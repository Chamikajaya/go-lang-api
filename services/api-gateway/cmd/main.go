package main

import (
	"context"
	"fmt"
	"log/slog"
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
// @version 1.0
// @description API Gateway for User Management Microservices
// @host localhost:8080
// @BasePath /
// @schemes http https
// ! TODO: NEED TO UNDERSTAND MAIN() FUNC
func main() {
	cfg := config.LoadConfig()

	nc, err := connectNATS(cfg)
	if err != nil {
		slog.Error("Failed to connect to NATS", "error", err)
		os.Exit(1) // failing fast if we can't connect to NATS, since it's critical for the API Gateway's functionality
	}
	defer nc.Close()
	slog.Info("Successfully connected to NATS")

	rpcClient := client.NewClient(nc, cfg.RPCTimeout)
	userHandler := handlers.NewUserHandler(rpcClient)

	wsManager := ws.NewManager()
	go wsManager.Run()
	slog.Info("WebSocket manager started")

	wsCRUDHandler := ws.NewCRUDHandler(rpcClient, wsManager)
	slog.Info("WebSocket CRUD handler initialized")

	// Initialize and start NATS event subscriber (bridges NATS events → WebSocket)
	eventSubscriber := subscriber.NewSubscriber(nc, wsManager)
	if err := eventSubscriber.Start(); err != nil {
		slog.Error("Failed to start event subscriber", "error", err)
	}
	slog.Info("NATS event subscriber started")

	// Setup router
	router := setupRouter(userHandler, wsManager, wsCRUDHandler)

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
		slog.Info("API Gateway starting", "port", cfg.ServerPort)
		slog.Info("Swagger docs available", "url", fmt.Sprintf("http://localhost:%s/swagger/index.html", cfg.ServerPort))
		slog.Info("WebSocket endpoint available", "url", fmt.Sprintf("ws://localhost:%s/ws?userId={uuid}", cfg.ServerPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
		}
	}()

	// Graceful shutdown
	gracefulShutdown(server, nc, eventSubscriber, cfg.ShutdownTimeout)
}

func connectNATS(cfg *config.Config) (*nats.Conn, error) {

	opts := []nats.Option{
		nats.Name(cfg.NATS_CONNECTION_NAME),
		nats.ReconnectWait(cfg.RECONNECT_WAIT), // how long to wait before attempting to reconnect after a disconnect
		nats.MaxReconnects(-1),                 // Unlimited reconnects

		/* EVENT CALLBACKS */
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				slog.Error("NATS disconnected", "error", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			slog.Info("NATS reconnected", "url", nc.ConnectedUrl())
		}),
		// catch all for async errors
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			slog.Error("NATS error", "error", err)
		}),
	}

	nc, err := nats.Connect(cfg.NatsURL, opts...) // ... unpacks the slice into individual arguments
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err) // %w -> error wrapping
	}

	return nc, nil
}

func setupRouter(userHandler *handlers.UserHandler, wsManager *ws.Manager, wsCRUDHandler *ws.CRUDHandler) *chi.Mux { // dependency injection
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID) // to generate a unique request ID for each incoming HTTP request
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer) //  If a handler panics, Recoverer catches it, logs the stack trace, and returns a 500 Internal Server Error instead of crashing the whole server process.
	r.Use(middleware.CORS)
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// WebSocket endpoint (upgrading standard http GET to persistent WebSocket) -  supports CRUD + notifications
	r.Get("/ws", ws.HandleWebSocket(wsManager, wsCRUDHandler))

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

func gracefulShutdown(server *http.Server, nc *nats.Conn, eventSub *subscriber.Subscriber, timeout time.Duration) {

	quit := make(chan os.Signal, 1)                      // shape of the pipe is os.Signal messages , 1 means buffer can hold just one os.Signal message
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // sigint -> ctrl + c && sigterm -> standard polite method to kill a process - routing those signals to the channel created

	sig := <-quit // code paueses here and wait until the signal generates
	slog.Info("Received signal", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), timeout) // wrapping the slate context with a timeout, so that if the shutdown takes too long, it will forcefully exit after the timeout duration.

	defer cancel()

	// Stop event subscriber first
	if err := eventSub.Stop(); err != nil {
		slog.Error("Error stopping event subscriber", "error", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Error shutting down HTTP server", "error", err)
	}

	if err := nc.Drain(); err != nil {
		slog.Error("Error draining NATS connection", "error", err) // draining -> nats pushes out any pending final messages it was holding
	}

	slog.Info("API Gateway stopped gracefully")
}
