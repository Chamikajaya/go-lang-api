// Package models_test tests the error model and its helper functions.
package models_test

import (
	"errors"
	"net/http"
	"testing"

	"user-management-api/internal/models"
)

// ============================================================================
// Testing AppError creation functions
// ============================================================================

func TestNewBadRequestError(t *testing.T) {
	// Create a bad request error
	err := models.NewBadRequestError("Invalid input")

	// Check status code
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, err.StatusCode)
	}

	// Check message
	if err.Message != "Invalid input" {
		t.Errorf("Expected message 'Invalid input', got '%s'", err.Message)
	}

	// Check that Err is nil (no wrapped error)
	if err.Err != nil {
		t.Error("Expected Err to be nil")
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := models.NewNotFoundError("User not found")

	if err.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, err.StatusCode)
	}

	if err.Message != "User not found" {
		t.Errorf("Expected message 'User not found', got '%s'", err.Message)
	}
}

func TestNewConflictError(t *testing.T) {
	err := models.NewConflictError("Email already exists")

	if err.StatusCode != http.StatusConflict {
		t.Errorf("Expected status code %d, got %d", http.StatusConflict, err.StatusCode)
	}

	if err.Message != "Email already exists" {
		t.Errorf("Expected message 'Email already exists', got '%s'", err.Message)
	}
}

func TestNewInternalServerError(t *testing.T) {
	// Create an underlying error
	underlyingErr := errors.New("database connection failed")

	// Create internal server error with wrapped error
	err := models.NewInternalServerError("Failed to process request", underlyingErr)

	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, err.StatusCode)
	}

	if err.Message != "Failed to process request" {
		t.Errorf("Expected message 'Failed to process request', got '%s'", err.Message)
	}

	// Check that the underlying error is stored
	if err.Err != underlyingErr {
		t.Error("Expected Err to contain the underlying error")
	}
}

// ============================================================================
// Testing AppError.Error() method
// AppError implements the error interface by having an Error() method
// ============================================================================

func TestAppError_Error_WithWrappedError(t *testing.T) {
	underlyingErr := errors.New("connection refused")
	err := models.NewInternalServerError("Database error", underlyingErr)

	// The Error() method should combine message and underlying error
	expected := "Database error: connection refused"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestAppError_Error_WithoutWrappedError(t *testing.T) {
	err := models.NewBadRequestError("Invalid JSON")

	// Without underlying error, just return the message
	if err.Error() != "Invalid JSON" {
		t.Errorf("Expected 'Invalid JSON', got '%s'", err.Error())
	}
}

// ============================================================================
// Testing that AppError satisfies the error interface
// This is important because we use type assertions in handlers
// ============================================================================

func TestAppError_ImplementsErrorInterface(t *testing.T) {
	// Create an AppError
	appErr := models.NewBadRequestError("test")

	// Try to assign it to the error interface
	// If this compiles, AppError implements error
	var err error = appErr

	// Use the error to avoid "unused variable" warning
	if err == nil {
		t.Error("Expected non-nil error")
	}
}

// ============================================================================
// Testing UserStatus constants
// ============================================================================

func TestUserStatusConstants(t *testing.T) {
	// Check that the status values are what we expect
	if models.UserStatusActive != "Active" {
		t.Errorf("Expected UserStatusActive to be 'Active', got '%s'", models.UserStatusActive)
	}

	if models.UserStatusInactive != "Inactive" {
		t.Errorf("Expected UserStatusInactive to be 'Inactive', got '%s'", models.UserStatusInactive)
	}
}

// ============================================================================
// Table-driven tests for all error types
// ============================================================================

func TestErrorCreators_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		createFunc     func() *models.AppError
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "BadRequest",
			createFunc: func() *models.AppError {
				return models.NewBadRequestError("bad request")
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "bad request",
		},
		{
			name: "NotFound",
			createFunc: func() *models.AppError {
				return models.NewNotFoundError("not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "not found",
		},
		{
			name: "Conflict",
			createFunc: func() *models.AppError {
				return models.NewConflictError("conflict")
			},
			expectedStatus: http.StatusConflict,
			expectedMsg:    "conflict",
		},
		{
			name: "InternalServerError",
			createFunc: func() *models.AppError {
				return models.NewInternalServerError("internal error", nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createFunc()

			if err.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, err.StatusCode)
			}

			if err.Message != tt.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, err.Message)
			}
		})
	}
}
