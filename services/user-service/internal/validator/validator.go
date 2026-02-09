package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator library
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// ValidateStruct validates a struct and returns a map of field errors
func (v *Validator) ValidateStruct(s interface{}) map[string]string {
	errors := make(map[string]string)
	err := v.validate.Struct(s)

	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		errors["_error"] = "Validation failed"
		return errors
	}

	for _, fieldError := range validationErrors {
		fieldName := firstCharToLowercase(fieldError.Field())
		errors[fieldName] = formatValidationError(fieldError)
	}

	return errors
}

// formatValidationError creates a user-friendly error message
func formatValidationError(fe validator.FieldError) string {
	field := firstCharToLowercase(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "e164":
		return fmt.Sprintf("%s must be a valid phone number in E.164 format", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// firstCharToLowercase converts the first character to lowercase
func firstCharToLowercase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
