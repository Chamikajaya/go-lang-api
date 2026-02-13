package client

import (
	"api-gateway/internal/models"
)

/* NATS RPC request wrappers - adds ActorID context to requests */

type CreateUserRPCRequest struct {
	ActorID string                   `json:"actor_id"` // to identify who made the request - could be like jwt subject or user ID
	Data    models.CreateUserRequest `json:"data"`
}

type GetUserRPCRequest struct {
	ActorID string `json:"actor_id"`
	UserID  string `json:"user_id"`
}

type ListUsersRPCRequest struct {
	ActorID string `json:"actor_id"`
}

type UpdateUserRPCRequest struct {
	ActorID string                   `json:"actor_id"`
	UserID  string                   `json:"user_id"`
	Data    models.UpdateUserRequest `json:"data"`
}

type DeleteUserRPCRequest struct {
	ActorID string `json:"actor_id"`
	UserID  string `json:"user_id"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}
