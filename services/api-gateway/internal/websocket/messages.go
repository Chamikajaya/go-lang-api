package websocket

import "encoding/json"

// WebSocket CRUD action types (client → server)
const (
	ActionCreateUser = "user.create"
	ActionGetUser    = "user.get"
	ActionListUsers  = "user.list"
	ActionUpdateUser = "user.update"
	ActionDeleteUser = "user.delete"
)

// WSRequest represents an incoming WebSocket CRUD request from a client.
// The client sends a JSON message with an action, a request_id for correlation,
// and a data payload containing the operation-specific fields.
type WSRequest struct {
	Action    string          `json:"action"`
	RequestID string          `json:"request_id"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// WSResponse is the response sent back to the requesting WebSocket client
// after a CRUD operation is processed.
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

// CreateUserPayload is the data payload for ActionCreateUser.
type CreateUserPayload struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	Age       int32  `json:"age,omitempty"`
	Status    string `json:"status,omitempty"`
}

// GetUserPayload is the data payload for ActionGetUser.
type GetUserPayload struct {
	UserID string `json:"user_id"`
}

// UpdateUserPayload is the data payload for ActionUpdateUser.
type UpdateUserPayload struct {
	UserID    string  `json:"user_id"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Age       *int32  `json:"age,omitempty"`
	Status    *string `json:"status,omitempty"`
}

// DeleteUserPayload is the data payload for ActionDeleteUser.
type DeleteUserPayload struct {
	UserID string `json:"user_id"`
}
