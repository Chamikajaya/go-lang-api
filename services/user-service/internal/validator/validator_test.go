package validator

import (
	"testing"

	"user-service/internal/models"
)

// Helper

func strPtr(s string) *string                          { return &s }
func intPtr(i int) *int                                { return &i }
func statusPtr(s models.UserStatus) *models.UserStatus { return &s }

// =============================================================================
// CreateUserRequest Validation Tests
// =============================================================================

func TestCreateUser_ValidRequest(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestCreateUser_ValidRequestAllFields(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     strPtr("+12025551234"),
		Age:       intPtr(25),
		Status:    models.UserStatusActive,
	}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestCreateUser_MissingFirstName(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		LastName: "Doe",
		Email:    "john@example.com",
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["firstName"]; !ok {
		t.Errorf("expected error on firstName, got %v", errs)
	}
}

func TestCreateUser_MissingLastName(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		Email:     "john@example.com",
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["lastName"]; !ok {
		t.Errorf("expected error on lastName, got %v", errs)
	}
}

func TestCreateUser_MissingEmail(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["email"]; !ok {
		t.Errorf("expected error on email, got %v", errs)
	}
}

func TestCreateUser_MissingAllRequired(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	for _, field := range []string{"firstName", "lastName", "email"} {
		if _, ok := errs[field]; !ok {
			t.Errorf("expected error on %s, missing from %v", field, errs)
		}
	}
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "not-an-email",
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["email"]; !ok {
		t.Errorf("expected error on email, got %v", errs)
	}
}

func TestCreateUser_FirstNameTooShort(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "J",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["firstName"]; !ok {
		t.Errorf("expected error on firstName (min=2), got %v", errs)
	}
}

func TestCreateUser_FirstNameTooLong(t *testing.T) {
	v := NewValidator()

	longName := ""
	for i := 0; i < 51; i++ {
		longName += "a"
	}

	req := models.CreateUserRequest{
		FirstName: longName,
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["firstName"]; !ok {
		t.Errorf("expected error on firstName (max=50), got %v", errs)
	}
}

func TestCreateUser_LastNameTooShort(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "D",
		Email:     "john@example.com",
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["lastName"]; !ok {
		t.Errorf("expected error on lastName (min=2), got %v", errs)
	}
}

func TestCreateUser_InvalidPhoneFormat(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     strPtr("123-456-7890"),
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["phone"]; !ok {
		t.Errorf("expected error on phone (e164), got %v", errs)
	}
}

func TestCreateUser_ValidPhoneE164(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     strPtr("+12025551234"),
	}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no validation errors for valid E.164 phone, got %v", errs)
	}
}

func TestCreateUser_AgeZero(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Age:       intPtr(0),
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["age"]; !ok {
		t.Errorf("expected error on age (gt=0), got %v", errs)
	}
}

func TestCreateUser_AgeNegative(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Age:       intPtr(-5),
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["age"]; !ok {
		t.Errorf("expected error on age (gt=0), got %v", errs)
	}
}

func TestCreateUser_AgePositive(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Age:       intPtr(25),
	}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no validation errors for positive age, got %v", errs)
	}
}

func TestCreateUser_InvalidStatus(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Status:    models.UserStatus("Unknown"),
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["status"]; !ok {
		t.Errorf("expected error on status (oneof), got %v", errs)
	}
}

func TestCreateUser_ValidStatusActive(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Status:    models.UserStatusActive,
	}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no errors for Active status, got %v", errs)
	}
}

func TestCreateUser_ValidStatusInactive(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Status:    models.UserStatusInactive,
	}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no errors for Inactive status, got %v", errs)
	}
}

// =============================================================================
// UpdateUserRequest Validation Tests
// =============================================================================

func TestUpdateUser_EmptyRequest(t *testing.T) {
	v := NewValidator()

	req := models.UpdateUserRequest{}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no validation errors for empty update, got %v", errs)
	}
}

func TestUpdateUser_ValidPartialUpdate(t *testing.T) {
	v := NewValidator()

	req := models.UpdateUserRequest{
		FirstName: strPtr("Jane"),
	}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestUpdateUser_InvalidEmail(t *testing.T) {
	v := NewValidator()

	req := models.UpdateUserRequest{
		Email: strPtr("bad-email"),
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["email"]; !ok {
		t.Errorf("expected error on email, got %v", errs)
	}
}

func TestUpdateUser_FirstNameTooShort(t *testing.T) {
	v := NewValidator()

	req := models.UpdateUserRequest{
		FirstName: strPtr("A"),
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["firstName"]; !ok {
		t.Errorf("expected error on firstName (min=2), got %v", errs)
	}
}

func TestUpdateUser_LastNameTooLong(t *testing.T) {
	v := NewValidator()

	longName := ""
	for i := 0; i < 51; i++ {
		longName += "b"
	}

	req := models.UpdateUserRequest{
		LastName: strPtr(longName),
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["lastName"]; !ok {
		t.Errorf("expected error on lastName (max=50), got %v", errs)
	}
}

func TestUpdateUser_InvalidPhone(t *testing.T) {
	v := NewValidator()

	req := models.UpdateUserRequest{
		Phone: strPtr("not-a-phone"),
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["phone"]; !ok {
		t.Errorf("expected error on phone (e164), got %v", errs)
	}
}

func TestUpdateUser_AgeZero(t *testing.T) {
	v := NewValidator()

	req := models.UpdateUserRequest{
		Age: intPtr(0),
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["age"]; !ok {
		t.Errorf("expected error on age (gt=0), got %v", errs)
	}
}

func TestUpdateUser_InvalidStatus(t *testing.T) {
	v := NewValidator()

	badStatus := models.UserStatus("Banned")
	req := models.UpdateUserRequest{
		Status: &badStatus,
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["status"]; !ok {
		t.Errorf("expected error on status (oneof), got %v", errs)
	}
}

func TestUpdateUser_ValidFullUpdate(t *testing.T) {
	v := NewValidator()

	req := models.UpdateUserRequest{
		FirstName: strPtr("Jane"),
		LastName:  strPtr("Smith"),
		Email:     strPtr("jane@example.com"),
		Phone:     strPtr("+442071234567"),
		Age:       intPtr(30),
		Status:    statusPtr(models.UserStatusInactive),
	}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestUpdateUser_MultipleInvalidFields(t *testing.T) {
	v := NewValidator()

	badStatus := models.UserStatus("Deleted")
	req := models.UpdateUserRequest{
		FirstName: strPtr("A"),
		Email:     strPtr("bad"),
		Age:       intPtr(-1),
		Status:    &badStatus,
	}

	errs := v.ValidateStruct(req)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	expectedFields := []string{"firstName", "email", "age", "status"}
	for _, field := range expectedFields {
		if _, ok := errs[field]; !ok {
			t.Errorf("expected error on %s, missing from %v", field, errs)
		}
	}
}

// =============================================================================
// Boundary Value Tests
// =============================================================================

func TestCreateUser_FirstNameExactlyMinLength(t *testing.T) {
	v := NewValidator()

	req := models.CreateUserRequest{
		FirstName: "Jo",
		LastName:  "Doe",
		Email:     "jo@example.com",
	}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no errors for firstName with exactly 2 chars, got %v", errs)
	}
}

func TestCreateUser_FirstNameExactlyMaxLength(t *testing.T) {
	v := NewValidator()

	name50 := ""
	for i := 0; i < 50; i++ {
		name50 += "a"
	}

	req := models.CreateUserRequest{
		FirstName: name50,
		LastName:  "Doe",
		Email:     "test@example.com",
	}

	errs := v.ValidateStruct(req)
	if errs != nil {
		t.Errorf("expected no errors for firstName with exactly 50 chars, got %v", errs)
	}
}
