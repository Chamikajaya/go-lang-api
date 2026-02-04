package models

import (
	"fmt"
	"net/http"
)

type AppError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Err        error  `json:"-"`  // internal error, not exposed to clients in json responses
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewBadRequestError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}


func NewNotFoundError(message string) *AppError {
	return &AppError{
		StatusCode:    http.StatusNotFound,
		Message: message,
	}
}

func NewInternalServerError(message string, err error) *AppError {
	return &AppError{
		StatusCode:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		StatusCode:    http.StatusConflict,
		Message: message,
	}
}

// ErrorResponse is the JSON structure sent to clients
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"` // omitempty = exclude if empty
}

// ValidationError represents validation failures
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}