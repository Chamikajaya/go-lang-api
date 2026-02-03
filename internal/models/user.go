package models

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "Active"
	UserStatusInactive UserStatus = "Inactive"
)

// Requests
type CreateUserRequest struct {
	FirstName string     `json:"firstName" validate:"required,min=2,max=50"`
	LastName  string     `json:"lastName" validate:"required,min=2,max=50"`
	Email     string     `json:"email" validate:"required,email"`
	Phone     *string    `json:"phone,omitempty" validate:"omitempty,e164"` // Pointer = optional field
	Age       *int       `json:"age,omitempty" validate:"omitempty,gt=0"`
	Status    UserStatus `json:"status,omitempty" validate:"omitempty,oneof=Active Inactive"`
}

type UpdateUserRequest struct {
	FirstName *string     `json:"firstName,omitempty" validate:"omitempty,min=2,max=50"`
	LastName  *string     `json:"lastName,omitempty" validate:"omitempty,min=2,max=50"`
	Email     *string     `json:"email,omitempty" validate:"omitempty,email"`
	Phone     *string     `json:"phone,omitempty" validate:"omitempty,e164"`
	Age       *int        `json:"age,omitempty" validate:"omitempty,gt=0"`
	Status    *UserStatus `json:"status,omitempty" validate:"omitempty,oneof=Active Inactive"`
}

// Responses
type UserResponse struct {
	UserID    uuid.UUID  `json:"userId"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Email     string     `json:"email"`
	Phone     *string    `json:"phone,omitempty"`
	Age       *int       `json:"age,omitempty"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}


type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // interface{} = any type 
}