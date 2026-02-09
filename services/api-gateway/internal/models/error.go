package models

// ErrorResponse is the JSON structure sent to clients for errors
type ErrorResponse struct {
	Error   string            `json:"error" example:"Bad Request"`
	Message string            `json:"message" example:"Invalid request body"`
	Details map[string]string `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err, message string) *ErrorResponse {
	return &ErrorResponse{
		Error:   err,
		Message: message,
	}
}

// NewValidationErrorResponse creates a new validation error response
func NewValidationErrorResponse(details map[string]string) *ErrorResponse {
	return &ErrorResponse{
		Error:   "Validation Failed",
		Message: "One or more fields failed validation",
		Details: details,
	}
}
