package models

import "time"

/*  HTTP API request/response (external API contract) */

type CreateUserRequest struct {
	FirstName string `json:"firstName" example:"John"`
	LastName  string `json:"lastName" example:"Doe"`
	Email     string `json:"email" example:"john.doe@example.com"`
	Phone     string `json:"phone" example:"+1234567890"`
	Age       int32  `json:"age" example:"30"`
	Status    string `json:"status" example:"Active"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"firstName,omitempty" example:"Jane"`
	LastName  *string `json:"lastName,omitempty" example:"Doe"`
	Email     *string `json:"email,omitempty" example:"jane.doe@example.com"`
	Phone     *string `json:"phone,omitempty" example:"+1234567890"`
	Age       *int32  `json:"age,omitempty" example:"25"`
	Status    *string `json:"status,omitempty" example:"Inactive"`
}

type UserResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	FirstName string    `json:"firstName" example:"John"`
	LastName  string    `json:"lastName" example:"Doe"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	Phone     string    `json:"phone" example:"+1234567890"`
	Age       int32     `json:"age" example:"30"`
	Status    string    `json:"status" example:"Active"`
	CreatedAt time.Time `json:"createdAt" example:"2024-01-15T10:30:00Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2024-01-15T10:30:00Z"`
}

type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total" example:"10"`
}

type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation completed successfully"`
}
