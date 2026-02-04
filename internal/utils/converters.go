package utils

import (
	database "user-management-api/db/sqlc"
	"user-management-api/internal/models"

	"github.com/jackc/pgx/v5/pgtype"
)

// ConvertToUserResponse converts database user to API response
func ConvertToUserResponse(user database.User) *models.UserResponse {
	return &models.UserResponse{
		UserID:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     ConvertTextToStringPtr(user.Phone),
		Age:       ConvertInt4ToIntPtr(user.Age),
		Status:    models.UserStatus(user.Status),
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}
}

// ConvertStringPtrToText converts *string to pgtype.Text
func ConvertStringPtrToText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// ConvertTextToStringPtr converts pgtype.Text to *string
func ConvertTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// ConvertIntPtrToInt4 converts *int to pgtype.Int4
func ConvertIntPtrToInt4(i *int) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*i), Valid: true}
}

// ConvertInt4ToIntPtr converts pgtype.Int4 to *int
func ConvertInt4ToIntPtr(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	val := int(i.Int32)
	return &val
}
