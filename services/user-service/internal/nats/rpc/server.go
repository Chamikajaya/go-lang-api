package rpc

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"user-service/internal/models"
	"user-service/internal/service"
	"user-service/internal/validator"

	"github.com/nats-io/nats.go"
)

// Server handles NATS RPC requests
type Server struct {
	nc        *nats.Conn
	service   *service.UserService
	validator *validator.Validator
	subs      []*nats.Subscription
}

func NewServer(nc *nats.Conn, svc *service.UserService, v *validator.Validator) *Server {
	return &Server{
		nc:        nc,
		service:   svc,
		validator: v,
		subs:      make([]*nats.Subscription, 0),
	}
}

// Start subscribes to all RPC subjects
func (s *Server) Start() error {
	handlers := map[string]nats.MsgHandler{
		SubjectUserCreate: s.handleCreateUser,
		SubjectUserGet:    s.handleGetUser,
		SubjectUserList:   s.handleListUsers,
		SubjectUserUpdate: s.handleUpdateUser,
		SubjectUserDelete: s.handleDeleteUser,
	}

	for subject, handler := range handlers {
		sub, err := s.nc.Subscribe(subject, handler)
		if err != nil {
			return err
		}
		s.subs = append(s.subs, sub)
		log.Printf("Subscribed to RPC subject: %s", subject)
	}

	return nil
}

// Stop unsubscribes from all RPC subjects
func (s *Server) Stop() error {
	for _, sub := range s.subs {
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("Error unsubscribing from %s: %v", sub.Subject, err)
		}
	}
	return nil
}

// handleCreateUser handles user creation RPC requests
func (s *Server) handleCreateUser(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var req CreateUserRPCRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		s.respondWithError(msg, http.StatusBadRequest, "Invalid request format", nil)
		return
	}

	// Validate request
	if validationErrors := s.validator.ValidateStruct(req.Data); validationErrors != nil {
		s.respondWithError(msg, http.StatusBadRequest, "Validation failed", validationErrors)
		return
	}

	// Call service
	user, err := s.service.CreateUser(ctx, req.Data, req.ActorID)
	if err != nil {
		s.handleServiceError(msg, err)
		return
	}

	s.respondWithSuccess(msg, user)
}

// handleGetUser handles get user by ID RPC requests
func (s *Server) handleGetUser(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var req GetUserRPCRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		s.respondWithError(msg, http.StatusBadRequest, "Invalid request format", nil)
		return
	}

	user, err := s.service.GetUserByID(ctx, req.UserID)
	if err != nil {
		s.handleServiceError(msg, err)
		return
	}

	s.respondWithSuccess(msg, user)
}

// list all users RPC requests
func (s *Server) handleListUsers(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := s.service.ListUsers(ctx)
	if err != nil {
		s.handleServiceError(msg, err)
		return
	}

	s.respondWithSuccess(msg, users)
}

// handleUpdateUser handles user update RPC requests
func (s *Server) handleUpdateUser(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var req UpdateUserRPCRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		s.respondWithError(msg, http.StatusBadRequest, "Invalid request format", nil)
		return
	}

	// Validate request
	if validationErrors := s.validator.ValidateStruct(req.Data); validationErrors != nil {
		s.respondWithError(msg, http.StatusBadRequest, "Validation failed", validationErrors)
		return
	}

	user, err := s.service.UpdateUser(ctx, req.UserID, req.Data, req.ActorID)
	if err != nil {
		s.handleServiceError(msg, err)
		return
	}

	s.respondWithSuccess(msg, user)
}

// handleDeleteUser handles user deletion RPC requests
func (s *Server) handleDeleteUser(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var req DeleteUserRPCRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		s.respondWithError(msg, http.StatusBadRequest, "Invalid request format", nil)
		return
	}

	err := s.service.DeleteUser(ctx, req.UserID, req.ActorID)
	if err != nil {
		s.handleServiceError(msg, err)
		return
	}

	s.respondWithSuccess(msg, models.DeleteUserResponse{
		Message: "User deleted successfully",
	})
}

// respondWithSuccess sends a successful RPC response
func (s *Server) respondWithSuccess(msg *nats.Msg, data interface{}) {
	response := NewSuccessResponse(data)
	s.respond(msg, response)
}

// respondWithError sends an error RPC response
func (s *Server) respondWithError(msg *nats.Msg, code int, message string, details map[string]string) {
	response := NewErrorResponse(code, message, details)
	s.respond(msg, response)
}

// handleServiceError converts service errors to RPC errors
func (s *Server) handleServiceError(msg *nats.Msg, err error) {
	if appErr, ok := err.(*models.AppError); ok {
		s.respondWithError(msg, appErr.StatusCode, appErr.Message, nil)
		return
	}
	s.respondWithError(msg, http.StatusInternalServerError, "An unexpected error occurred", nil)
}

// respond sends an RPC response
func (s *Server) respond(msg *nats.Msg, response *RPCResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal RPC response: %v", err)
		return
	}

	if err := msg.Respond(data); err != nil {
		log.Printf("Failed to send RPC response: %v", err)
	}
}
