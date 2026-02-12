package http

import (
	"time"

	"github.com/google/uuid"

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

func (r *CreateUserRequest) ToDomain() *domain.User {
	return &domain.User{
		FirstName:       r.FirstName,
		LastName:        r.LastName,
		Patronymic:      r.Patronymic,
		Phone:           r.Phone,
		Telegram:        r.Telegram,
		Email:           r.Email,
		Gender:          r.Gender,
		Country:         r.Country,
		Region:          r.Region,
		Nationality:     r.Nationality,
		ProfilePicURL:   r.ProfilePicURL,
		Status:          r.Status,
		TariffType:      r.TariffType,
		JobStatus:       r.JobStatus,
		ActivityType:    r.ActivityType,
		Specializations: r.Specializations,
	}
}

func (r *UpdateUserRequest) ToDomain(id uuid.UUID) *domain.User {
	return &domain.User{
		ID:              id,
		FirstName:       r.FirstName,
		LastName:        r.LastName,
		Patronymic:      r.Patronymic,
		Phone:           r.Phone,
		Telegram:        r.Telegram,
		Email:           r.Email,
		Gender:          r.Gender,
		Country:         r.Country,
		Region:          r.Region,
		Nationality:     r.Nationality,
		ProfilePicURL:   r.ProfilePicURL,
		Status:          r.Status,
		TariffType:      r.TariffType,
		JobStatus:       r.JobStatus,
		ActivityType:    r.ActivityType,
		Specializations: r.Specializations,
	}
}

type UserResponse struct {
	ID              string            `json:"id"`
	FirstName       string            `json:"first_name"`
	LastName        string            `json:"last_name"`
	Patronymic      string            `json:"patronymic,omitempty"`
	Phone           string            `json:"phone,omitempty"`
	Telegram        string            `json:"telegram,omitempty"`
	Email           string            `json:"email,omitempty"`
	Gender          string            `json:"gender,omitempty"`
	Country         string            `json:"country,omitempty"`
	Region          string            `json:"region,omitempty"`
	Nationality     string            `json:"nationality,omitempty"`
	ProfilePicURL   string            `json:"profile_pic_url,omitempty"`
	Status          string            `json:"status"`
	TariffType      string            `json:"tariff_type"`
	JobStatus       string            `json:"job_status,omitempty"`
	ActivityType    string            `json:"activity_type,omitempty"`
	Specializations []string          `json:"specializations"`
	CreatedAt       string            `json:"created_at"`
	Profile         map[string]string `json:"profile,omitempty"`
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
	return toUserResponseWithProfile(u, nil)
}

func toUserResponseWithProfile(u *domain.User, profile map[string]string) UserResponse {
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
		Profile:         profile,
	}
}
