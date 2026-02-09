package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"api-gateway/internal/models"
	"api-gateway/internal/nats/client"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	client *client.Client
}

func NewUserHandler(client *client.Client) *UserHandler {
	return &UserHandler{
		client: client,
	}
}

// getActorID extracts actor ID from request header
// ! NEEDS TO BE MANUALLY SET IN REQUESTS FOR NOW
func getActorID(r *http.Request) string {
	actorID := r.Header.Get("X-Actor-ID")
	if actorID == "" {
		actorID = "00000000-0000-0000-0000-000000000000"
	}
	return actorID
}

// handleRPCError handles RPC client errors
func handleRPCError(w http.ResponseWriter, err error) {
	if errors.Is(err, client.ErrTimeout) {
		RespondError(w, http.StatusGatewayTimeout, "Gateway Timeout", "Request to user service timed out")
		return
	}
	if errors.Is(err, client.ErrServiceUnavailable) {
		RespondError(w, http.StatusServiceUnavailable, "Service Unavailable", "User service is not available")
		return
	}
	log.Printf("RPC error: %v", err)
	RespondError(w, http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred")
}

// handleServiceError handles error responses from the service
func handleServiceError(w http.ResponseWriter, errResp *client.ErrorResponse) {
	status := http.StatusInternalServerError
	switch errResp.Error {
	case "Not Found":
		status = http.StatusNotFound
	case "Bad Request", "Validation Failed":
		status = http.StatusBadRequest
	case "Conflict":
		status = http.StatusConflict
	}
	RespondJSON(w, status, errResp)
}

// CreateUser handles POST /users
// @Summary Create a new user
// @Description Create a new user with the provided details
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User to create"
// @Param X-Actor-ID header string false "Actor ID (UUID)"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Bad Request", "Invalid request body")
		return
	}

	// Build RPC request
	rpcReq := &client.CreateUserRPCRequest{
		ActorID: getActorID(r),
		Data:    req,
	}

	resp, errResp, err := h.client.CreateUser(r.Context(), rpcReq)
	if err != nil {
		handleRPCError(w, err)
		return
	}
	if errResp != nil {
		handleServiceError(w, errResp)
		return
	}

	RespondJSON(w, http.StatusCreated, resp)
}

// GetUser handles GET /users/{id}
// @Summary Get a user by ID
// @Description Get a user's details by their ID
// @Tags users
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param X-Actor-ID header string false "Actor ID (UUID)"
// @Success 200 {object} models.UserResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		RespondError(w, http.StatusBadRequest, "Bad Request", "User ID is required")
		return
	}

	rpcReq := &client.GetUserRPCRequest{
		ActorID: getActorID(r),
		UserID:  userID,
	}

	resp, errResp, err := h.client.GetUser(r.Context(), rpcReq)
	if err != nil {
		handleRPCError(w, err)
		return
	}
	if errResp != nil {
		handleServiceError(w, errResp)
		return
	}

	RespondJSON(w, http.StatusOK, resp)
}

// ListUsers handles GET /users
// @Summary List all users
// @Description Get a list of all users
// @Tags users
// @Produce json
// @Param X-Actor-ID header string false "Actor ID (UUID)"
// @Success 200 {object} models.ListUsersResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rpcReq := &client.ListUsersRPCRequest{
		ActorID: getActorID(r),
	}

	resp, errResp, err := h.client.ListUsers(r.Context(), rpcReq)
	if err != nil {
		handleRPCError(w, err)
		return
	}
	if errResp != nil {
		handleServiceError(w, errResp)
		return
	}

	RespondJSON(w, http.StatusOK, resp)
}

// UpdateUser handles PATCH /users/{id}
// @Summary Update a user
// @Description Update a user's details
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param user body models.UpdateUserRequest true "Fields to update"
// @Param X-Actor-ID header string false "Actor ID (UUID)"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [patch]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		RespondError(w, http.StatusBadRequest, "Bad Request", "User ID is required")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Bad Request", "Invalid request body")
		return
	}

	rpcReq := &client.UpdateUserRPCRequest{
		ActorID: getActorID(r),
		UserID:  userID,
		Data:    req,
	}

	resp, errResp, err := h.client.UpdateUser(r.Context(), rpcReq)
	if err != nil {
		handleRPCError(w, err)
		return
	}
	if errResp != nil {
		handleServiceError(w, errResp)
		return
	}

	RespondJSON(w, http.StatusOK, resp)
}

// DeleteUser handles DELETE /users/{id}
// @Summary Delete a user
// @Description Delete a user by their ID
// @Tags users
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param X-Actor-ID header string false "Actor ID (UUID)"
// @Success 200 {object} models.SuccessResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		RespondError(w, http.StatusBadRequest, "Bad Request", "User ID is required")
		return
	}

	rpcReq := &client.DeleteUserRPCRequest{
		ActorID: getActorID(r),
		UserID:  userID,
	}

	resp, errResp, err := h.client.DeleteUser(r.Context(), rpcReq)
	if err != nil {
		handleRPCError(w, err)
		return
	}
	if errResp != nil {
		handleServiceError(w, errResp)
		return
	}

	RespondJSON(w, http.StatusOK, resp)
}
