// Package utils_test tests the converter utility functions.
// These are "pure functions" - they take input and return output with no side effects.
// Pure functions are the EASIEST to test!
package utils_test

import (
	"testing"
	"time"

	database "user-management-api/db/sqlc"
	"user-management-api/internal/models"
	"user-management-api/internal/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ============================================================================
// Testing ConvertStringPtrToText - converts Go *string to PostgreSQL Text
// ============================================================================

func TestConvertStringPtrToText_WithValue(t *testing.T) {
	// Create a string and get its pointer
	value := "hello"
	ptr := &value

	// Convert to pgtype.Text
	result := utils.ConvertStringPtrToText(ptr)

	// Check the result
	if !result.Valid {
		t.Error("Expected Valid to be true")
	}
	if result.String != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result.String)
	}
}

func TestConvertStringPtrToText_WithNil(t *testing.T) {
	// Pass nil (no value)
	result := utils.ConvertStringPtrToText(nil)

	// When input is nil, Valid should be false
	// This represents NULL in the database
	if result.Valid {
		t.Error("Expected Valid to be false for nil input")
	}
}

// ============================================================================
// Testing ConvertTextToStringPtr - converts PostgreSQL Text to Go *string
// ============================================================================

func TestConvertTextToStringPtr_WithValidValue(t *testing.T) {
	// Create a valid pgtype.Text (has a value)
	text := pgtype.Text{
		String: "world",
		Valid:  true,
	}

	result := utils.ConvertTextToStringPtr(text)

	// Result should not be nil
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	// Dereference pointer with * to get the value
	if *result != "world" {
		t.Errorf("Expected 'world', got '%s'", *result)
	}
}

func TestConvertTextToStringPtr_WithInvalidValue(t *testing.T) {
	// Create an invalid pgtype.Text (NULL in database)
	text := pgtype.Text{
		Valid: false, // This means NULL
	}

	result := utils.ConvertTextToStringPtr(text)

	// Result should be nil (Go's way of representing "no value")
	if result != nil {
		t.Error("Expected nil result for invalid/NULL text")
	}
}

// ============================================================================
// Testing ConvertIntPtrToInt4 - converts Go *int to PostgreSQL Int4
// ============================================================================

func TestConvertIntPtrToInt4_WithValue(t *testing.T) {
	age := 25
	ptr := &age

	result := utils.ConvertIntPtrToInt4(ptr)

	if !result.Valid {
		t.Error("Expected Valid to be true")
	}
	// Int4 stores as int32, so we compare as int32
	if result.Int32 != 25 {
		t.Errorf("Expected 25, got %d", result.Int32)
	}
}

func TestConvertIntPtrToInt4_WithNil(t *testing.T) {
	result := utils.ConvertIntPtrToInt4(nil)

	if result.Valid {
		t.Error("Expected Valid to be false for nil input")
	}
}

// ============================================================================
// Testing ConvertInt4ToIntPtr - converts PostgreSQL Int4 to Go *int
// ============================================================================

func TestConvertInt4ToIntPtr_WithValidValue(t *testing.T) {
	int4 := pgtype.Int4{
		Int32: 30,
		Valid: true,
	}

	result := utils.ConvertInt4ToIntPtr(int4)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if *result != 30 {
		t.Errorf("Expected 30, got %d", *result)
	}
}

func TestConvertInt4ToIntPtr_WithInvalidValue(t *testing.T) {
	int4 := pgtype.Int4{
		Valid: false, // NULL
	}

	result := utils.ConvertInt4ToIntPtr(int4)

	if result != nil {
		t.Error("Expected nil result for invalid/NULL int4")
	}
}

// ============================================================================
// Testing ConvertToUserResponse - converts database User to API response
// ============================================================================

func TestConvertToUserResponse_FullUser(t *testing.T) {
	// Create a mock database user with all fields populated
	userID := uuid.New()
	now := time.Now()

	dbUser := database.User{
		UserID:    userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone: pgtype.Text{
			String: "+14155551234",
			Valid:  true,
		},
		Age: pgtype.Int4{
			Int32: 28,
			Valid: true,
		},
		Status: "Active",
		CreatedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
	}

	// Convert to API response
	result := utils.ConvertToUserResponse(dbUser)

	// Verify all fields are correctly mapped
	if result.UserID != userID {
		t.Errorf("UserID mismatch: expected %v, got %v", userID, result.UserID)
	}
	if result.FirstName != "John" {
		t.Errorf("FirstName mismatch: expected 'John', got '%s'", result.FirstName)
	}
	if result.LastName != "Doe" {
		t.Errorf("LastName mismatch: expected 'Doe', got '%s'", result.LastName)
	}
	if result.Email != "john@example.com" {
		t.Errorf("Email mismatch")
	}
	if result.Phone == nil || *result.Phone != "+14155551234" {
		t.Errorf("Phone mismatch")
	}
	if result.Age == nil || *result.Age != 28 {
		t.Errorf("Age mismatch")
	}
	if result.Status != models.UserStatusActive {
		t.Errorf("Status mismatch: expected Active, got %s", result.Status)
	}
}

func TestConvertToUserResponse_NullableFieldsAreNil(t *testing.T) {
	// Create a user with nullable fields set to NULL
	dbUser := database.User{
		UserID:    uuid.New(),
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane@example.com",
		Phone: pgtype.Text{
			Valid: false, // NULL
		},
		Age: pgtype.Int4{
			Valid: false, // NULL
		},
		Status: "Inactive",
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	result := utils.ConvertToUserResponse(dbUser)

	// Nullable fields should be nil in the response
	if result.Phone != nil {
		t.Error("Expected Phone to be nil")
	}
	if result.Age != nil {
		t.Error("Expected Age to be nil")
	}
}

// ============================================================================
// Benchmark test example - measures performance
// Benchmark functions start with "Benchmark" instead of "Test"
// Run with: go test -bench=. ./internal/utils/
// ============================================================================

func BenchmarkConvertStringPtrToText(b *testing.B) {
	value := "test string"
	ptr := &value

	// b.N is automatically set by the testing framework
	// It runs the code enough times to get accurate measurements
	for i := 0; i < b.N; i++ {
		utils.ConvertStringPtrToText(ptr)
	}
}

func BenchmarkConvertToUserResponse(b *testing.B) {
	dbUser := database.User{
		UserID:    uuid.New(),
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     pgtype.Text{String: "+1234567890", Valid: true},
		Age:       pgtype.Int4{Int32: 30, Valid: true},
		Status:    "Active",
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	for i := 0; i < b.N; i++ {
		utils.ConvertToUserResponse(dbUser)
	}
}
