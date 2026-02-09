package client

import (
	"api-gateway/internal/models"
)

/* NATS RPC request wrappers - adds ActorID context to requests */

// CreateUserRPCRequest wraps CreateUserRequest with actor context
type CreateUserRPCRequest struct {
	ActorID string                   `json:"actor_id"`
	Data    models.CreateUserRequest `json:"data"`
}

// GetUserRPCRequest represents a request to get a user by ID
type GetUserRPCRequest struct {
	ActorID string `json:"actor_id"`
	UserID  string `json:"user_id"`
}

// ListUsersRPCRequest represents a request to list all users
type ListUsersRPCRequest struct {
	ActorID string `json:"actor_id"`
}

// UpdateUserRPCRequest wraps UpdateUserRequest with actor context
type UpdateUserRPCRequest struct {
	ActorID string                   `json:"actor_id"`
	UserID  string                   `json:"user_id"`
	Data    models.UpdateUserRequest `json:"data"`
}

// DeleteUserRPCRequest represents a request to delete a user
type DeleteUserRPCRequest struct {
	ActorID string `json:"actor_id"`
	UserID  string `json:"user_id"`
}

// ErrorResponse represents an error response from the service
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}
