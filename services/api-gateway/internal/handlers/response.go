package handlers

import (
	"encoding/json"
	"net/http"

	"api-gateway/internal/models"
)

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, status int, err, message string) {
	RespondJSON(w, status, models.NewErrorResponse(err, message))
}

// RespondValidationError sends a validation error response
func RespondValidationError(w http.ResponseWriter, details map[string]string) {
	RespondJSON(w, http.StatusBadRequest, models.NewValidationErrorResponse(details))
}
