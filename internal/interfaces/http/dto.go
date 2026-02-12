package http

import (
	"time"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CreateUserRequest struct {
	FirstName       string   `json:"first_name" binding:"required"`
	LastName        string   `json:"last_name" binding:"required"`
	Patronymic      string   `json:"patronymic"`
	Phone           string   `json:"phone"`
	Telegram        string   `json:"telegram"`
	Email           string   `json:"email" binding:"omitempty,email"`
	Gender          string   `json:"gender"`
	Country         string   `json:"country"`
	Region          string   `json:"region"`
	Nationality     string   `json:"nationality"`
	ProfilePicURL   string   `json:"profile_pic_url"`
	Status          string   `json:"status"`
	TariffType      string   `json:"tariff_type"`
	JobStatus       string   `json:"job_status"`
	ActivityType    string   `json:"activity_type"`
	Specializations []string `json:"specializations"`
}

type UpdateUserRequest struct {
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	Patronymic      string   `json:"patronymic"`
	Phone           string   `json:"phone"`
	Telegram        string   `json:"telegram"`
	Email           string   `json:"email" binding:"omitempty,email"`
	Gender          string   `json:"gender"`
	Country         string   `json:"country"`
	Region          string   `json:"region"`
	Nationality     string   `json:"nationality"`
	ProfilePicURL   string   `json:"profile_pic_url"`
	Status          string   `json:"status"`
	TariffType      string   `json:"tariff_type"`
	JobStatus       string   `json:"job_status"`
	ActivityType    string   `json:"activity_type"`
	Specializations []string `json:"specializations"`
}

type UserResponse struct {
	ID              string   `json:"id"`
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	Patronymic      string   `json:"patronymic,omitempty"`
	Phone           string   `json:"phone,omitempty"`
	Telegram        string   `json:"telegram,omitempty"`
	Email           string   `json:"email,omitempty"`
	Gender          string   `json:"gender,omitempty"`
	Country         string   `json:"country,omitempty"`
	Region          string   `json:"region,omitempty"`
	Nationality     string   `json:"nationality,omitempty"`
	ProfilePicURL   string   `json:"profile_pic_url,omitempty"`
	Status          string   `json:"status"`
	TariffType      string   `json:"tariff_type"`
	JobStatus       string   `json:"job_status,omitempty"`
	ActivityType    string   `json:"activity_type,omitempty"`
	Specializations []string `json:"specializations"`
	CreatedAt       string   `json:"created_at"`
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
	specs := u.Specializations
	if specs == nil {
		specs = []string{}
	}
	return UserResponse{
		ID:              u.ID.String(),
		FirstName:       u.FirstName,
		LastName:        u.LastName,
		Patronymic:      u.Patronymic,
		Phone:           u.Phone,
		Telegram:        u.Telegram,
		Email:           u.Email,
		Gender:          u.Gender,
		Country:         u.Country,
		Region:          u.Region,
		Nationality:     u.Nationality,
		ProfilePicURL:   u.ProfilePicURL,
		Status:          u.Status,
		TariffType:      u.TariffType,
		JobStatus:       u.JobStatus,
		ActivityType:    u.ActivityType,
		Specializations: specs,
		CreatedAt:       u.CreatedAt.Format(time.RFC3339),
	}
}
