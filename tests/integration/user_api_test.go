// Package integration contains integration tests for the API.
//
// INTEGRATION TESTS vs UNIT TESTS
// ================================
// Unit Tests: Test one small piece in isolation (fast, use mocks)
// Integration Tests: Test multiple pieces working together (slower, use real DB)
//
// These tests use testcontainers to spin up a real PostgreSQL database
// in a Docker container. This means:
// - Tests are slower (container startup + DB operations)
// - Tests are more realistic (catch issues that mocks might miss)
// - Docker must be running to run these tests
//
// To skip integration tests when Docker isn't available, we use build tags.
// Run with: go test -tags=integration ./tests/integration/...
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	database "user-management-api/db/sqlc"
	"user-management-api/internal/handlers"
	"user-management-api/internal/middleware"
	"user-management-api/internal/models"
	"user-management-api/internal/service"
	"user-management-api/internal/validator"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================================
// Test Setup
// ============================================================================

// TestApp holds all the dependencies needed for testing
type TestApp struct {
	Pool    *pgxpool.Pool
	Handler *handlers.UserHandler
	Router  *chi.Mux
}

// setupTestApp creates a test application with real database connection.
// This function should be called at the start of integration tests.
//
// NOTE: This requires a running PostgreSQL database.
// For CI/CD, you would use testcontainers-go to spin up a container.
func setupTestApp(t *testing.T) *TestApp {
	t.Helper() // Marks this as a helper function (better error messages)

	// Connect to test database
	// In a real setup, you would:
	// 1. Use testcontainers-go to start PostgreSQL in Docker
	// 2. Or use a separate test database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use environment variable or default to local test database
	dbURL := "postgres://postgres:postgres@localhost:5432/user_management?sslmode=disable"

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to database: %v", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Skipping integration test: cannot ping database: %v", err)
	}

	// Clean up database before tests
	cleanupDatabase(t, pool)

	// Create dependencies
	queries := database.New(pool)
	userService := service.NewUserService(pool, queries)
	validatorInstance := validator.NewValidator()
	userHandler := handlers.NewUserHandler(userService, validatorInstance)

	// Create router with middleware (same as production)
	router := chi.NewRouter()
	router.Use(middleware.ContentTypeJSON)

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.CreateUser)
			r.Get("/", userHandler.ListUsers)
			r.Get("/{id}", userHandler.GetUser)
			r.Patch("/{id}", userHandler.UpdateUser)
			r.Delete("/{id}", userHandler.DeleteUser)
		})
	})

	return &TestApp{
		Pool:    pool,
		Handler: userHandler,
		Router:  router,
	}
}

// cleanupDatabase removes all test data
func cleanupDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	_, err := pool.Exec(ctx, "DELETE FROM users")
	if err != nil {
		t.Logf("Warning: failed to cleanup database: %v", err)
	}
}

// teardown cleans up after tests
func (app *TestApp) teardown(t *testing.T) {
	t.Helper()
	cleanupDatabase(t, app.Pool)
	app.Pool.Close()
}

// ============================================================================
// Integration Tests
// ============================================================================

// TestIntegration_CreateAndGetUser tests the full create and get flow.
// This is a common pattern: test a complete user journey.
func TestIntegration_CreateAndGetUser(t *testing.T) {
	app := setupTestApp(t)
	defer app.teardown(t)

	// Step 1: Create a user
	createBody := models.CreateUserRequest{
		FirstName: "Integration",
		LastName:  "Test",
		Email:     "integration@test.com",
	}
	body, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Create user failed: expected %d, got %d. Body: %s",
			http.StatusCreated, rr.Code, rr.Body.String())
	}

	// Parse the created user
	var createdUser models.UserResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &createdUser); err != nil {
		t.Fatalf("Failed to parse created user: %v", err)
	}

	t.Logf("Created user with ID: %s", createdUser.UserID)

	// Step 2: Get the user by ID
	getReq := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/users/%s", createdUser.UserID), nil)
	getRr := httptest.NewRecorder()

	app.Router.ServeHTTP(getRr, getReq)

	if getRr.Code != http.StatusOK {
		t.Fatalf("Get user failed: expected %d, got %d. Body: %s",
			http.StatusOK, getRr.Code, getRr.Body.String())
	}

	var fetchedUser models.UserResponse
	if err := json.Unmarshal(getRr.Body.Bytes(), &fetchedUser); err != nil {
		t.Fatalf("Failed to parse fetched user: %v", err)
	}

	// Verify the data matches
	if fetchedUser.FirstName != "Integration" {
		t.Errorf("Expected FirstName 'Integration', got '%s'", fetchedUser.FirstName)
	}
	if fetchedUser.Email != "integration@test.com" {
		t.Errorf("Expected Email 'integration@test.com', got '%s'", fetchedUser.Email)
	}
}

// TestIntegration_CreateDuplicateEmail tests that duplicate emails are rejected.
func TestIntegration_CreateDuplicateEmail(t *testing.T) {
	app := setupTestApp(t)
	defer app.teardown(t)

	// Create first user
	createBody := models.CreateUserRequest{
		FirstName: "First",
		LastName:  "User",
		Email:     "duplicate@test.com",
	}
	body, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("First create failed: %s", rr.Body.String())
	}

	// Try to create second user with same email
	createBody2 := models.CreateUserRequest{
		FirstName: "Second",
		LastName:  "User",
		Email:     "duplicate@test.com", // Same email!
	}
	body2, _ := json.Marshal(createBody2)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()

	app.Router.ServeHTTP(rr2, req2)

	// Should get 409 Conflict
	if rr2.Code != http.StatusConflict {
		t.Errorf("Expected status %d for duplicate email, got %d. Body: %s",
			http.StatusConflict, rr2.Code, rr2.Body.String())
	}
}

// TestIntegration_UpdateUser tests updating a user.
func TestIntegration_UpdateUser(t *testing.T) {
	app := setupTestApp(t)
	defer app.teardown(t)

	// Create a user first
	createBody := models.CreateUserRequest{
		FirstName: "Original",
		LastName:  "Name",
		Email:     "update@test.com",
	}
	body, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	var createdUser models.UserResponse
	json.Unmarshal(rr.Body.Bytes(), &createdUser)

	// Update the user
	newFirstName := "Updated"
	updateBody := models.UpdateUserRequest{
		FirstName: &newFirstName,
	}
	updateBodyBytes, _ := json.Marshal(updateBody)

	updateReq := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/api/v1/users/%s", createdUser.UserID), bytes.NewReader(updateBodyBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRr := httptest.NewRecorder()

	app.Router.ServeHTTP(updateRr, updateReq)

	if updateRr.Code != http.StatusOK {
		t.Fatalf("Update failed: expected %d, got %d. Body: %s",
			http.StatusOK, updateRr.Code, updateRr.Body.String())
	}

	var updatedUser models.UserResponse
	json.Unmarshal(updateRr.Body.Bytes(), &updatedUser)

	if updatedUser.FirstName != "Updated" {
		t.Errorf("Expected FirstName 'Updated', got '%s'", updatedUser.FirstName)
	}

	// Original last name should be unchanged
	if updatedUser.LastName != "Name" {
		t.Errorf("Expected LastName 'Name' (unchanged), got '%s'", updatedUser.LastName)
	}
}

// TestIntegration_DeleteUser tests deleting a user.
func TestIntegration_DeleteUser(t *testing.T) {
	app := setupTestApp(t)
	defer app.teardown(t)

	// Create a user first
	createBody := models.CreateUserRequest{
		FirstName: "ToDelete",
		LastName:  "User",
		Email:     "delete@test.com",
	}
	body, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	var createdUser models.UserResponse
	json.Unmarshal(rr.Body.Bytes(), &createdUser)

	// Delete the user
	deleteReq := httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/api/v1/users/%s", createdUser.UserID), nil)
	deleteRr := httptest.NewRecorder()

	app.Router.ServeHTTP(deleteRr, deleteReq)

	if deleteRr.Code != http.StatusOK {
		t.Fatalf("Delete failed: expected %d, got %d", http.StatusOK, deleteRr.Code)
	}

	// Try to get the deleted user - should be 404
	getReq := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/users/%s", createdUser.UserID), nil)
	getRr := httptest.NewRecorder()

	app.Router.ServeHTTP(getRr, getReq)

	if getRr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for deleted user, got %d", getRr.Code)
	}
}

// TestIntegration_ListUsers tests listing all users.
func TestIntegration_ListUsers(t *testing.T) {
	app := setupTestApp(t)
	defer app.teardown(t)

	// Create multiple users
	users := []models.CreateUserRequest{
		{FirstName: "User1", LastName: "Test", Email: "user1@test.com"},
		{FirstName: "User2", LastName: "Test", Email: "user2@test.com"},
		{FirstName: "User3", LastName: "Test", Email: "user3@test.com"},
	}

	for _, user := range users {
		body, _ := json.Marshal(user)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		app.Router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("Failed to create user %s: %s", user.Email, rr.Body.String())
		}
	}

	// List all users
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	listRr := httptest.NewRecorder()

	app.Router.ServeHTTP(listRr, listReq)

	if listRr.Code != http.StatusOK {
		t.Fatalf("List users failed: expected %d, got %d", http.StatusOK, listRr.Code)
	}

	var response models.ListUsersResponse
	if err := json.Unmarshal(listRr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Total != 3 {
		t.Errorf("Expected 3 users, got %d", response.Total)
	}
}

// TestIntegration_GetNonExistentUser tests getting a user that doesn't exist.
func TestIntegration_GetNonExistentUser(t *testing.T) {
	app := setupTestApp(t)
	defer app.teardown(t)

	// Try to get a user with a random UUID that doesn't exist
	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/users/550e8400-e29b-41d4-a716-446655440000", nil)
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent user, got %d", rr.Code)
	}
}
