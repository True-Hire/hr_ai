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
	TelegramID      string   `json:"telegram_id"`
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
	Password        string   `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	Patronymic      string   `json:"patronymic"`
	Phone           string   `json:"phone"`
	Telegram        string   `json:"telegram"`
	TelegramID      string   `json:"telegram_id"`
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
		TelegramID:      r.TelegramID,
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
		TelegramID:      r.TelegramID,
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
	ID              string               `json:"id"`
	FirstName       string               `json:"first_name"`
	LastName        string               `json:"last_name"`
	Patronymic      string               `json:"patronymic,omitempty"`
	Phone           string               `json:"phone,omitempty"`
	Telegram        string               `json:"telegram,omitempty"`
	TelegramID      string               `json:"telegram_id,omitempty"`
	Email           string               `json:"email,omitempty"`
	Gender          string               `json:"gender,omitempty"`
	Country         string               `json:"country,omitempty"`
	Region          string               `json:"region,omitempty"`
	Nationality     string               `json:"nationality,omitempty"`
	ProfilePicURL   string               `json:"profile_pic_url,omitempty"`
	Status          string               `json:"status"`
	TariffType      string               `json:"tariff_type"`
	JobStatus       string               `json:"job_status,omitempty"`
	ActivityType    string               `json:"activity_type,omitempty"`
	Specializations []string             `json:"specializations"`
	ProfileScore    int32                `json:"profile_score"`
	CreatedAt       string               `json:"created_at"`
	Profile         *UserProfileResponse `json:"profile,omitempty"`
	SearchScore     *float64             `json:"search_score,omitempty"`
}

type UserProfileResponse struct {
	Title          string                   `json:"title,omitempty"`
	About          string                   `json:"about,omitempty"`
	Skills         []string                 `json:"skills,omitempty"`
	Languages      []LanguageItemResponse   `json:"languages,omitempty"`
	Certifications []string                 `json:"certifications,omitempty"`
	Achievements   string                   `json:"achievements,omitempty"`
	Experience     []ExperienceItemResponse `json:"experience,omitempty"`
	Education      []EducationItemResponse  `json:"education,omitempty"`
}

type LanguageItemResponse struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

type ProjectResponse struct {
	Project string   `json:"project"`
	Items   []string `json:"items"`
}

type ExperienceItemResponse struct {
	ID          string            `json:"id"`
	Company     string            `json:"company"`
	Position    string            `json:"position"`
	StartDate   string            `json:"start_date,omitempty"`
	EndDate     string            `json:"end_date,omitempty"`
	Projects    []ProjectResponse `json:"projects,omitempty"`
	WebSite     string            `json:"web_site,omitempty"`
	Description string            `json:"description,omitempty"`
}

type EducationItemResponse struct {
	ID           string `json:"id"`
	Institution  string `json:"institution"`
	Degree       string `json:"degree"`
	FieldOfStudy string `json:"field_of_study,omitempty"`
	StartDate    string `json:"start_date,omitempty"`
	EndDate      string `json:"end_date,omitempty"`
	Location     string `json:"location,omitempty"`
	Description  string `json:"description,omitempty"`
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

func toUserResponseWithProfile(u *domain.User, profile *UserProfileResponse) UserResponse {
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
		TelegramID:      u.TelegramID,
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
		ProfileScore:    u.ProfileScore,
		CreatedAt:       u.CreatedAt.Format(time.RFC3339),
		Profile:         profile,
	}
}
