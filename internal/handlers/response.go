package handlers

import (
	"encoding/json"
	"net/http"

	"user-management-api/internal/models"
)

// sendJSON sends a JSON response
func (h *UserHandler) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	// json.NewEncoder writes to w (io.Writer)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, log it (response already started)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// sendError sends an error response
func (h *UserHandler) sendError(w http.ResponseWriter, appErr *models.AppError) {
	response := models.ErrorResponse{
		Error:   http.StatusText(appErr.StatusCode),
		Message: appErr.Message,
	}
	h.sendJSON(w, appErr.StatusCode, response)
}

// sendValidationError sends validation error response
func (h *UserHandler) sendValidationError(w http.ResponseWriter, errors map[string]string) {
	response := models.ErrorResponse{
		Error:   "Validation Failed",
		Message: "One or more fields failed validation",
		Details: errors,
	}
	h.sendJSON(w, http.StatusBadRequest, response)
}

// handleServiceError converts service errors to HTTP responses
func (h *UserHandler) handleServiceError(w http.ResponseWriter, err error) {
	// Type assertion to check if it's our custom error
	if appErr, ok := err.(*models.AppError); ok {
		h.sendError(w, appErr)
		return
	}

	// Unknown error - return 500
	h.sendError(w, models.NewInternalServerError("An unexpected error occurred", err))
}
