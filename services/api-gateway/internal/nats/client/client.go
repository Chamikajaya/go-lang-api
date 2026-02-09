package client

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"api-gateway/internal/models"

	"github.com/nats-io/nats.go"
)

// NATS RPC communication
type Client struct {
	nc      *nats.Conn
	timeout time.Duration
}

func NewClient(nc *nats.Conn, timeout time.Duration) *Client {
	return &Client{
		nc:      nc,
		timeout: timeout,
	}
}

// Subjects for NATS RPC
const (
	SubjectUserCreate = "user.create"
	SubjectUserGet    = "user.get"
	SubjectUserList   = "user.list"
	SubjectUserUpdate = "user.update"
	SubjectUserDelete = "user.delete"
)

// RPCResponse represents the wrapper response from user-service
type RPCResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents an error in RPC response
type RPCError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Request sends an RPC request and waits for a response
func (c *Client) Request(ctx context.Context, subject string, request interface{}) (*nats.Msg, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Send request
	msg, err := c.nc.RequestWithContext(ctx, subject, data)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, nats.ErrTimeout) {
			return nil, ErrTimeout
		}
		if errors.Is(err, nats.ErrNoResponders) {
			return nil, ErrServiceUnavailable
		}
		return nil, err
	}

	return msg, nil
}

// CreateUser sends a request to create a user
func (c *Client) CreateUser(ctx context.Context, req *CreateUserRPCRequest) (*models.UserResponse, *ErrorResponse, error) {
	msg, err := c.Request(ctx, SubjectUserCreate, req)
	if err != nil {
		return nil, nil, err
	}

	// Parse the RPC wrapper response
	var rpcResp RPCResponse
	if err := json.Unmarshal(msg.Data, &rpcResp); err != nil {
		return nil, nil, err
	}

	if !rpcResp.Success && rpcResp.Error != nil {
		return nil, &ErrorResponse{
			Error:   getErrorName(rpcResp.Error.Code),
			Message: rpcResp.Error.Message,
			Details: rpcResp.Error.Details,
		}, nil
	}

	// Parse success response data
	var resp models.UserResponse
	if err := json.Unmarshal(rpcResp.Data, &resp); err != nil {
		return nil, nil, err
	}

	return &resp, nil, nil
}

// GetUser sends a request to get a user by ID
func (c *Client) GetUser(ctx context.Context, req *GetUserRPCRequest) (*models.UserResponse, *ErrorResponse, error) {
	msg, err := c.Request(ctx, SubjectUserGet, req)
	if err != nil {
		return nil, nil, err
	}

	// Parse the RPC wrapper response
	var rpcResp RPCResponse
	if err := json.Unmarshal(msg.Data, &rpcResp); err != nil {
		return nil, nil, err
	}

	if !rpcResp.Success && rpcResp.Error != nil {
		return nil, &ErrorResponse{
			Error:   getErrorName(rpcResp.Error.Code),
			Message: rpcResp.Error.Message,
			Details: rpcResp.Error.Details,
		}, nil
	}

	// Parse success response data
	var resp models.UserResponse
	if err := json.Unmarshal(rpcResp.Data, &resp); err != nil {
		return nil, nil, err
	}

	return &resp, nil, nil
}

// ListUsers sends a request to list all users
func (c *Client) ListUsers(ctx context.Context, req *ListUsersRPCRequest) (*models.ListUsersResponse, *ErrorResponse, error) {
	msg, err := c.Request(ctx, SubjectUserList, req)
	if err != nil {
		return nil, nil, err
	}

	// Parse the RPC wrapper response
	var rpcResp RPCResponse
	if err := json.Unmarshal(msg.Data, &rpcResp); err != nil {
		return nil, nil, err
	}

	if !rpcResp.Success && rpcResp.Error != nil {
		return nil, &ErrorResponse{
			Error:   getErrorName(rpcResp.Error.Code),
			Message: rpcResp.Error.Message,
			Details: rpcResp.Error.Details,
		}, nil
	}

	// Parse success response data - it's a ListUsersResponse object with users array
	var resp models.ListUsersResponse
	if err := json.Unmarshal(rpcResp.Data, &resp); err != nil {
		return nil, nil, err
	}

	return &resp, nil, nil
}

// UpdateUser sends a request to update a user
func (c *Client) UpdateUser(ctx context.Context, req *UpdateUserRPCRequest) (*models.UserResponse, *ErrorResponse, error) {
	msg, err := c.Request(ctx, SubjectUserUpdate, req)
	if err != nil {
		return nil, nil, err
	}

	// Parse the RPC wrapper response
	var rpcResp RPCResponse
	if err := json.Unmarshal(msg.Data, &rpcResp); err != nil {
		return nil, nil, err
	}

	if !rpcResp.Success && rpcResp.Error != nil {
		return nil, &ErrorResponse{
			Error:   getErrorName(rpcResp.Error.Code),
			Message: rpcResp.Error.Message,
			Details: rpcResp.Error.Details,
		}, nil
	}

	// Parse success response data
	var resp models.UserResponse
	if err := json.Unmarshal(rpcResp.Data, &resp); err != nil {
		return nil, nil, err
	}

	return &resp, nil, nil
}

// DeleteUser sends a request to delete a user
func (c *Client) DeleteUser(ctx context.Context, req *DeleteUserRPCRequest) (*models.SuccessResponse, *ErrorResponse, error) {
	msg, err := c.Request(ctx, SubjectUserDelete, req)
	if err != nil {
		return nil, nil, err
	}

	// Parse the RPC wrapper response
	var rpcResp RPCResponse
	if err := json.Unmarshal(msg.Data, &rpcResp); err != nil {
		return nil, nil, err
	}

	if !rpcResp.Success && rpcResp.Error != nil {
		return nil, &ErrorResponse{
			Error:   getErrorName(rpcResp.Error.Code),
			Message: rpcResp.Error.Message,
			Details: rpcResp.Error.Details,
		}, nil
	}

	return &models.SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil, nil
}

// getErrorName converts HTTP status codes to error names
func getErrorName(code int) string {
	switch code {
	case 400:
		return "Bad Request"
	case 404:
		return "Not Found"
	case 409:
		return "Conflict"
	case 500:
		return "Internal Server Error"
	default:
		return "Error"
	}
}

// Predefined errors
var (
	ErrTimeout            = errors.New("request timeout")
	ErrServiceUnavailable = errors.New("service unavailable")
)
