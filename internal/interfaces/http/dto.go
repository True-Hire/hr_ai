package http

import (
	"time"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CreateUserRequest struct {
	Phone         string `json:"phone" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
	ProfilePicURL string `json:"profile_pic_url"`
}

type UpdateUserRequest struct {
	Phone         string `json:"phone"`
	Email         string `json:"email" binding:"omitempty,email"`
	ProfilePicURL string `json:"profile_pic_url"`
}

type UserResponse struct {
	ID            string `json:"id"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	ProfilePicURL string `json:"profile_pic_url,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type PaginatedUsersResponse struct {
	Users    []UserResponse `json:"users"`
	Total    int64          `json:"total"`
	Page     int32          `json:"page"`
	PageSize int32          `json:"page_size"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:            u.ID.String(),
		Phone:         u.Phone,
		Email:         u.Email,
		ProfilePicURL: u.ProfilePicURL,
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
	}
}
