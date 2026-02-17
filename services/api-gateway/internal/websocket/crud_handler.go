package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"api-gateway/internal/models"
	"api-gateway/internal/nats/client"
)

// CRUDHandler processes incoming WebSocket CRUD requests by forwarding them
// to the user-service via the NATS RPC client and sending responses back
// over the WebSocket connection.
type CRUDHandler struct {
	rpcClient *client.Client
	manager   *Manager
}

// NewCRUDHandler creates a new handler with the shared NATS RPC client.
func NewCRUDHandler(rpcClient *client.Client, manager *Manager) *CRUDHandler {
	return &CRUDHandler{
		rpcClient: rpcClient,
		manager:   manager,
	}
}

// HandleMessage routes an incoming WSRequest to the appropriate CRUD handler
// and sends the WSResponse back to the originating client.
func (h *CRUDHandler) HandleMessage(clientUserID string, raw []byte) {
	var req WSRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		h.sendError(clientUserID, "", "unknown", "Bad Request", "Invalid JSON message", nil)
		return
	}

	if req.Action == "" {
		h.sendError(clientUserID, req.RequestID, "unknown", "Bad Request", "action field is required", nil)
		return
	}

	log.Printf("WebSocket CRUD: user %s, action=%s, request_id=%s", clientUserID, req.Action, req.RequestID)

	switch req.Action {
	case ActionCreateUser:
		h.handleCreateUser(clientUserID, &req)
	case ActionGetUser:
		h.handleGetUser(clientUserID, &req)
	case ActionListUsers:
		h.handleListUsers(clientUserID, &req)
	case ActionUpdateUser:
		h.handleUpdateUser(clientUserID, &req)
	case ActionDeleteUser:
		h.handleDeleteUser(clientUserID, &req)
	default:
		h.sendError(clientUserID, req.RequestID, req.Action, "Bad Request", "Unknown action: "+req.Action, nil)
	}
}

func (h *CRUDHandler) handleCreateUser(clientUserID string, req *WSRequest) {
	var payload CreateUserPayload
	if err := json.Unmarshal(req.Data, &payload); err != nil {
		h.sendError(clientUserID, req.RequestID, req.Action, "Bad Request", "Invalid data for user.create", nil)
		return
	}

	rpcReq := &client.CreateUserRPCRequest{
		ActorID: clientUserID,
		Data: models.CreateUserRequest{
			FirstName: payload.FirstName,
			LastName:  payload.LastName,
			Email:     payload.Email,
			Phone:     payload.Phone,
			Age:       payload.Age,
			Status:    payload.Status,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, errResp, err := h.rpcClient.CreateUser(ctx, rpcReq)
	if err != nil {
		h.sendRPCError(clientUserID, req.RequestID, req.Action, err)
		return
	}
	if errResp != nil {
		h.sendError(clientUserID, req.RequestID, req.Action, errResp.Error, errResp.Message, errResp.Details)
		return
	}

	h.sendSuccess(clientUserID, req.RequestID, req.Action, resp)
}

func (h *CRUDHandler) handleGetUser(clientUserID string, req *WSRequest) {
	var payload GetUserPayload
	if err := json.Unmarshal(req.Data, &payload); err != nil {
		h.sendError(clientUserID, req.RequestID, req.Action, "Bad Request", "Invalid data for user.get", nil)
		return
	}

	if payload.UserID == "" {
		h.sendError(clientUserID, req.RequestID, req.Action, "Bad Request", "user_id is required", nil)
		return
	}

	rpcReq := &client.GetUserRPCRequest{
		ActorID: clientUserID,
		UserID:  payload.UserID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, errResp, err := h.rpcClient.GetUser(ctx, rpcReq)
	if err != nil {
		h.sendRPCError(clientUserID, req.RequestID, req.Action, err)
		return
	}
	if errResp != nil {
		h.sendError(clientUserID, req.RequestID, req.Action, errResp.Error, errResp.Message, errResp.Details)
		return
	}

	h.sendSuccess(clientUserID, req.RequestID, req.Action, resp)
}

func (h *CRUDHandler) handleListUsers(clientUserID string, req *WSRequest) {
	rpcReq := &client.ListUsersRPCRequest{
		ActorID: clientUserID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, errResp, err := h.rpcClient.ListUsers(ctx, rpcReq)
	if err != nil {
		h.sendRPCError(clientUserID, req.RequestID, req.Action, err)
		return
	}
	if errResp != nil {
		h.sendError(clientUserID, req.RequestID, req.Action, errResp.Error, errResp.Message, errResp.Details)
		return
	}

	h.sendSuccess(clientUserID, req.RequestID, req.Action, resp)
}

func (h *CRUDHandler) handleUpdateUser(clientUserID string, req *WSRequest) {
	var payload UpdateUserPayload
	if err := json.Unmarshal(req.Data, &payload); err != nil {
		h.sendError(clientUserID, req.RequestID, req.Action, "Bad Request", "Invalid data for user.update", nil)
		return
	}

	if payload.UserID == "" {
		h.sendError(clientUserID, req.RequestID, req.Action, "Bad Request", "user_id is required", nil)
		return
	}

	rpcReq := &client.UpdateUserRPCRequest{
		ActorID: clientUserID,
		UserID:  payload.UserID,
		Data: models.UpdateUserRequest{
			FirstName: payload.FirstName,
			LastName:  payload.LastName,
			Email:     payload.Email,
			Phone:     payload.Phone,
			Age:       payload.Age,
			Status:    payload.Status,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, errResp, err := h.rpcClient.UpdateUser(ctx, rpcReq)
	if err != nil {
		h.sendRPCError(clientUserID, req.RequestID, req.Action, err)
		return
	}
	if errResp != nil {
		h.sendError(clientUserID, req.RequestID, req.Action, errResp.Error, errResp.Message, errResp.Details)
		return
	}

	h.sendSuccess(clientUserID, req.RequestID, req.Action, resp)
}

func (h *CRUDHandler) handleDeleteUser(clientUserID string, req *WSRequest) {
	var payload DeleteUserPayload
	if err := json.Unmarshal(req.Data, &payload); err != nil {
		h.sendError(clientUserID, req.RequestID, req.Action, "Bad Request", "Invalid data for user.delete", nil)
		return
	}

	if payload.UserID == "" {
		h.sendError(clientUserID, req.RequestID, req.Action, "Bad Request", "user_id is required", nil)
		return
	}

	rpcReq := &client.DeleteUserRPCRequest{
		ActorID: clientUserID,
		UserID:  payload.UserID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, errResp, err := h.rpcClient.DeleteUser(ctx, rpcReq)
	if err != nil {
		h.sendRPCError(clientUserID, req.RequestID, req.Action, err)
		return
	}
	if errResp != nil {
		h.sendError(clientUserID, req.RequestID, req.Action, errResp.Error, errResp.Message, errResp.Details)
		return
	}

	h.sendSuccess(clientUserID, req.RequestID, req.Action, resp)
}

// --- Response helpers ---

func (h *CRUDHandler) sendSuccess(clientUserID, requestID, action string, data interface{}) {
	resp := WSResponse{
		Action:    action,
		RequestID: requestID,
		Success:   true,
		Data:      data,
	}
	h.sendResponse(clientUserID, &resp)
}

func (h *CRUDHandler) sendError(clientUserID, requestID, action, code, message string, details map[string]string) {
	resp := WSResponse{
		Action:    action,
		RequestID: requestID,
		Success:   false,
		Error: &WSError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	h.sendResponse(clientUserID, &resp)
}

func (h *CRUDHandler) sendRPCError(clientUserID, requestID, action string, err error) {
	code := "Internal Server Error"
	message := "An unexpected error occurred"

	if err == client.ErrTimeout {
		code = "Gateway Timeout"
		message = "Request to user service timed out"
	} else if err == client.ErrServiceUnavailable {
		code = "Service Unavailable"
		message = "User service is not available"
	}

	h.sendError(clientUserID, requestID, action, code, message, nil)
}

func (h *CRUDHandler) sendResponse(clientUserID string, resp *WSResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("WebSocket CRUD: failed to marshal response for user %s: %v", clientUserID, err)
		return
	}

	if !h.manager.SendToClient(clientUserID, data) {
		log.Printf("WebSocket CRUD: failed to send response to user %s (disconnected?)", clientUserID)
	}
}
