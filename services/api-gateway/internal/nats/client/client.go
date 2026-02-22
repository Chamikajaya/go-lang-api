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
	timeout time.Duration // * for how long apigw will wait for a response from user-service over nats before giving up.
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

// * the wrapper response from user-service
type RPCResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// error in RPC response
type RPCError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// ! TODO: FIX CODE DUPLICATION
// sends an RPC request and waits for a response
// * CONTEXT -> to pass deadlines, cancellation signals, and other request-scoped values across API boundaries and between processes.
func (c *Client) request(ctx context.Context, subject string, req interface{}) (*nats.Msg, error) {
	data, err := json.Marshal(req) // go struct → json
	if err != nil {
		return nil, err
	}

	// taking the incoming context and wrapping it in a new one, to enforce a limit on how long nats request can take
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel() // to release resources associated with the context once the request is done

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

// doRequest sends RPC request and handles common response parsing
func (c *Client) doRequest(ctx context.Context, subject string, req interface{}) (json.RawMessage, *ErrorResponse, error) {
	// ! TODO: implement one error parameter, instead of returning 2 types of errors
	msg, err := c.request(ctx, subject, req)
	if err != nil {
		return nil, nil, err
	}

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

	return rpcResp.Data, nil, nil
}

func (c *Client) CreateUser(ctx context.Context, req *CreateUserRPCRequest) (*models.UserResponse, *ErrorResponse, error) {
	data, errResp, err := c.doRequest(ctx, SubjectUserCreate, req)
	if err != nil || errResp != nil {
		return nil, errResp, err
	}

	var resp models.UserResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, nil, err
	}
	return &resp, nil, nil
}

func (c *Client) GetUser(ctx context.Context, req *GetUserRPCRequest) (*models.UserResponse, *ErrorResponse, error) {
	data, errResp, err := c.doRequest(ctx, SubjectUserGet, req)
	if err != nil || errResp != nil {
		return nil, errResp, err
	}

	var resp models.UserResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, nil, err
	}
	return &resp, nil, nil
}

func (c *Client) ListUsers(ctx context.Context, req *ListUsersRPCRequest) (*models.ListUsersResponse, *ErrorResponse, error) {
	data, errResp, err := c.doRequest(ctx, SubjectUserList, req)
	if err != nil || errResp != nil {
		return nil, errResp, err
	}

	var resp models.ListUsersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, nil, err
	}
	return &resp, nil, nil
}

func (c *Client) UpdateUser(ctx context.Context, req *UpdateUserRPCRequest) (*models.UserResponse, *ErrorResponse, error) {
	data, errResp, err := c.doRequest(ctx, SubjectUserUpdate, req)
	if err != nil || errResp != nil {
		return nil, errResp, err
	}

	var resp models.UserResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, nil, err
	}
	return &resp, nil, nil
}

func (c *Client) DeleteUser(ctx context.Context, req *DeleteUserRPCRequest) (*models.SuccessResponse, *ErrorResponse, error) {
	_, errResp, err := c.doRequest(ctx, SubjectUserDelete, req)
	if err != nil || errResp != nil {
		return nil, errResp, err
	}

	return &models.SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil, nil
}

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

// Predefined errors - Sentinel errors (like named constants for errors) - need to use errors.Is to compare these errors
var (
	ErrTimeout            = errors.New("request timeout")
	ErrServiceUnavailable = errors.New("service unavailable")
)
