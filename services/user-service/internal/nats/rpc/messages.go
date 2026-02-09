package rpc

import "user-service/internal/models"

// RPC Subjects for request/response communication
const (
	SubjectUserCreate = "user.create"
	SubjectUserGet    = "user.get"
	SubjectUserList   = "user.list"
	SubjectUserUpdate = "user.update"
	SubjectUserDelete = "user.delete"
)

// RPCRequest represents a generic RPC request wrapper
type RPCRequest struct {
	ActorID string `json:"actor_id"` // ID of user making the request
}

type CreateUserRPCRequest struct {
	RPCRequest
	Data models.CreateUserRequest `json:"data"`
}

type GetUserRPCRequest struct {
	RPCRequest
	UserID string `json:"user_id"`
}

type ListUsersRPCRequest struct {
	RPCRequest
}

type UpdateUserRPCRequest struct {
	RPCRequest
	UserID string                   `json:"user_id"`
	Data   models.UpdateUserRequest `json:"data"`
}

type DeleteUserRPCRequest struct {
	RPCRequest
	UserID string `json:"user_id"`
}

type RPCResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func NewSuccessResponse(data interface{}) *RPCResponse {
	return &RPCResponse{
		Success: true,
		Data:    data,
	}
}

func NewErrorResponse(code int, message string, details map[string]string) *RPCResponse {
	return &RPCResponse{
		Success: false,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
