package handlers

import (
	"encoding/json"
	"net/http"

	"user-management-api/internal/models"
	"user-management-api/internal/service"
	"user-management-api/internal/validator"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	service   *service.UserService
	validator *validator.Validator
}

func NewUserHandler(service *service.UserService, validator *validator.Validator) *UserHandler {
	return &UserHandler{
		service:   service,
		validator: validator,
	}
}

// CreateUser creates a new user
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User to create"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	
	// Decode JSON body into struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, models.NewBadRequestError("Invalid request body"))
		return
	}
	
	// Validate request
	if validationErrors := h.validator.ValidateStruct(req); validationErrors != nil {
		h.sendValidationError(w, validationErrors)
		return
	}
	
	// Call service layer
	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	// Send successful response
	h.sendJSON(w, http.StatusCreated, user)
}

// GetUser retrieves a user by ID
// @Summary Get a user by ID
// @Description Get a single user by their UUID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {

	userID := chi.URLParam(r, "id")
	
	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.sendJSON(w, http.StatusOK, user)
}

// ListUsers retrieves all users
// @Summary List all users
// @Description Get a list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} models.ListUsersResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.sendJSON(w, http.StatusOK, users)
}

// UpdateUser updates an existing user
// @Summary Update a user
// @Description Update a user's information by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param user body models.UpdateUserRequest true "User fields to update"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [patch]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, models.NewBadRequestError("Invalid request body"))
		return
	}
	
	// Validate request
	if validationErrors := h.validator.ValidateStruct(req); validationErrors != nil {
		h.sendValidationError(w, validationErrors)
		return
	}
	
	user, err := h.service.UpdateUser(r.Context(), userID, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.sendJSON(w, http.StatusOK, user)
}

// DeleteUser deletes a user
// @Summary Delete a user
// @Description Delete a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	
	err := h.service.DeleteUser(r.Context(), userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.sendJSON(w, http.StatusOK, models.SuccessResponse{
		Message: "User deleted successfully",
	})
}