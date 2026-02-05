package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// wrapper around the validator library - so that we can easily swap it out later if needed
type Validator struct {
	validate *validator.Validate
}

// creating the validator instance
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

func (v *Validator) ValidateStruct(s interface{}) map[string]string {

	errors := make(map[string]string) // {fieldName: errorMessage}
	err := v.validate.Struct(s)

	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)

	if !ok {
		errors["_error"] = "Validation failed"
		return errors
	}

	// Iterate over validation errors
	for _, fieldError := range validationErrors {
		// Convert field name from PascalCase to camelCase
		fieldName := firstCharToLowercase(fieldError.Field())
		errors[fieldName] = formatValidationError(fieldError)
	}

	return errors
}

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

func firstCharToLowercase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
