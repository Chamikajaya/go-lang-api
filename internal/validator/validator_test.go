// Package validator_test contains unit tests for the validator package.
// Notice the package name ends with _test - this is a Go convention for "black box" testing.
// It means we test the package from the outside, like a real user would.
package validator_test

import (
	"testing"

	"user-management-api/internal/models"
	"user-management-api/internal/validator"
)

// TestNewValidator tests that we can create a new validator instance.
// Test function names in Go MUST start with "Test" followed by the function name.
func TestNewValidator(t *testing.T) {
	// Create a new validator
	v := validator.NewValidator()

	// Check it's not nil (basic sanity check)
	// t.Fatal stops the test immediately if it fails
	if v == nil {
		t.Fatal("Expected validator to not be nil")
	}
}

// TestValidateStruct_ValidCreateUserRequest tests validation with valid data.
// The naming convention is: Test<FunctionName>_<Scenario>
func TestValidateStruct_ValidCreateUserRequest(t *testing.T) {
	v := validator.NewValidator()

	// Create a valid request with all required fields
	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	// ValidateStruct returns nil when validation passes
	errors := v.ValidateStruct(req)

	// If errors is not nil, validation failed when it shouldn't have
	if errors != nil {
		t.Errorf("Expected no validation errors, got: %v", errors)
	}
}

// TestValidateStruct_MissingRequiredFields tests that required fields are enforced.
func TestValidateStruct_MissingRequiredFields(t *testing.T) {
	v := validator.NewValidator()

	// Empty request - all required fields missing
	req := models.CreateUserRequest{}

	errors := v.ValidateStruct(req)

	// We expect errors because required fields are missing
	if errors == nil {
		t.Fatal("Expected validation errors for empty request")
	}

	// Check specific fields have errors
	// In Go, we access map values with mapName[key]
	if _, exists := errors["firstName"]; !exists {
		t.Error("Expected error for firstName field")
	}
	if _, exists := errors["lastName"]; !exists {
		t.Error("Expected error for lastName field")
	}
	if _, exists := errors["email"]; !exists {
		t.Error("Expected error for email field")
	}
}

// TestValidateStruct_InvalidEmail tests email format validation.
func TestValidateStruct_InvalidEmail(t *testing.T) {
	v := validator.NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "not-an-email", // Invalid email format
	}

	errors := v.ValidateStruct(req)

	if errors == nil {
		t.Fatal("Expected validation error for invalid email")
	}

	if _, exists := errors["email"]; !exists {
		t.Error("Expected error for email field")
	}
}

// TestValidateStruct_NameTooShort tests minimum length validation.
func TestValidateStruct_NameTooShort(t *testing.T) {
	v := validator.NewValidator()

	req := models.CreateUserRequest{
		FirstName: "A", // Too short (min is 2)
		LastName:  "B", // Too short
		Email:     "test@example.com",
	}

	errors := v.ValidateStruct(req)

	if errors == nil {
		t.Fatal("Expected validation errors for short names")
	}

	// Check both name fields have errors
	if _, exists := errors["firstName"]; !exists {
		t.Error("Expected error for firstName field")
	}
	if _, exists := errors["lastName"]; !exists {
		t.Error("Expected error for lastName field")
	}
}

// TestValidateStruct_InvalidPhone tests phone number validation (E.164 format).
func TestValidateStruct_InvalidPhone(t *testing.T) {
	v := validator.NewValidator()

	invalidPhone := "123-456-7890" // Not E.164 format (should be like +14155551234)
	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     &invalidPhone,
	}

	errors := v.ValidateStruct(req)

	if errors == nil {
		t.Fatal("Expected validation error for invalid phone")
	}

	if _, exists := errors["phone"]; !exists {
		t.Error("Expected error for phone field")
	}
}

// TestValidateStruct_ValidPhone tests valid phone number passes.
func TestValidateStruct_ValidPhone(t *testing.T) {
	v := validator.NewValidator()

	validPhone := "+14155551234" // Valid E.164 format
	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     &validPhone,
	}

	errors := v.ValidateStruct(req)

	if errors != nil {
		t.Errorf("Expected no validation errors, got: %v", errors)
	}
}

// TestValidateStruct_InvalidAge tests that age must be greater than 0.
func TestValidateStruct_InvalidAge(t *testing.T) {
	v := validator.NewValidator()

	invalidAge := 0 // Must be > 0
	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Age:       &invalidAge,
	}

	errors := v.ValidateStruct(req)

	if errors == nil {
		t.Fatal("Expected validation error for invalid age")
	}

	if _, exists := errors["age"]; !exists {
		t.Error("Expected error for age field")
	}
}

// TestValidateStruct_InvalidStatus tests that status must be "Active" or "Inactive".
func TestValidateStruct_InvalidStatus(t *testing.T) {
	v := validator.NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Status:    "InvalidStatus", // Not "Active" or "Inactive"
	}

	errors := v.ValidateStruct(req)

	if errors == nil {
		t.Fatal("Expected validation error for invalid status")
	}

	if _, exists := errors["status"]; !exists {
		t.Error("Expected error for status field")
	}
}

// TestValidateStruct_UpdateUserRequest_AllOptional tests that UpdateUserRequest
// allows all fields to be nil (partial update).
func TestValidateStruct_UpdateUserRequest_AllOptional(t *testing.T) {
	v := validator.NewValidator()

	// Empty update request - all fields are optional
	req := models.UpdateUserRequest{}

	errors := v.ValidateStruct(req)

	// Should pass because all fields in UpdateUserRequest are optional
	if errors != nil {
		t.Errorf("Expected no validation errors for empty update request, got: %v", errors)
	}
}

// TestValidateStruct_UpdateUserRequest_PartialUpdate tests partial updates.
func TestValidateStruct_UpdateUserRequest_PartialUpdate(t *testing.T) {
	v := validator.NewValidator()

	// Only updating email
	newEmail := "new.email@example.com"
	req := models.UpdateUserRequest{
		Email: &newEmail,
	}

	errors := v.ValidateStruct(req)

	if errors != nil {
		t.Errorf("Expected no validation errors, got: %v", errors)
	}
}

// Table-driven test example - a very common Go pattern!
// This tests multiple scenarios in a single test function.
func TestValidateStruct_TableDriven(t *testing.T) {
	v := validator.NewValidator()

	// Define test cases as a slice of structs
	// Each struct represents one test scenario
	tests := []struct {
		name        string                    // Description of the test
		request     models.CreateUserRequest  // Input
		expectError bool                      // Expected outcome
		errorField  string                    // Which field should have error (if any)
	}{
		{
			name: "valid request",
			request: models.CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			expectError: false,
		},
		{
			name: "missing first name",
			request: models.CreateUserRequest{
				LastName: "Doe",
				Email:    "john@example.com",
			},
			expectError: true,
			errorField:  "firstName",
		},
		{
			name: "invalid email format",
			request: models.CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "not-valid",
			},
			expectError: true,
			errorField:  "email",
		},
	}

	// Loop through each test case
	for _, tt := range tests {
		// t.Run creates a subtest with the given name
		// This makes it easy to see which specific test failed
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidateStruct(tt.request)

			if tt.expectError {
				if errors == nil {
					t.Error("Expected validation error, got none")
				} else if tt.errorField != "" {
					if _, exists := errors[tt.errorField]; !exists {
						t.Errorf("Expected error for field %s, but it was not found", tt.errorField)
					}
				}
			} else {
				if errors != nil {
					t.Errorf("Expected no errors, got: %v", errors)
				}
			}
		})
	}
}
