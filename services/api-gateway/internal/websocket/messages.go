package websocket

import "encoding/json"

// WebSocket CRUD action types
const (
	ActionCreateUser = "user.create"
	ActionGetUser    = "user.get"
	ActionListUsers  = "user.list"
	ActionUpdateUser = "user.update"
	ActionDeleteUser = "user.delete"
)

// incoming WebSocket CRUD request from a client.
type WSRequest struct {
	Action    string          `json:"action"`
	RequestID string          `json:"request_id"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// response sent back to the requesting WebSocket client after a CRUD operation is processed.
type WSResponse struct {
	Action    string      `json:"action"`
	RequestID string      `json:"request_id"`
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *WSError    `json:"error,omitempty"`
}

// WSError contains error details for a failed WebSocket CRUD operation.
type WSError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Payloads for individual CRUD operations embedded in WSRequest.Data

type CreateUserPayload struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	Age       int32  `json:"age,omitempty"`
	Status    string `json:"status,omitempty"`
}

type GetUserPayload struct {
	UserID string `json:"user_id"`
}

type UpdateUserPayload struct {
	UserID    string  `json:"user_id"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Age       *int32  `json:"age,omitempty"`
	Status    *string `json:"status,omitempty"`
}

type DeleteUserPayload struct {
	UserID string `json:"user_id"`
}
