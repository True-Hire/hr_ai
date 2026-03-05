package http

import (
	"time"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CreateCompanyHRRequest struct {
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Patronymic string `json:"patronymic"`
	Phone      string `json:"phone"`
	Telegram   string `json:"telegram"`
	TelegramID string `json:"telegram_id"`
	Email      string `json:"email" binding:"omitempty,email"`
	Position   string `json:"position"`
	Password   string `json:"password" binding:"required,min=6"`
}

type UpdateCompanyHRRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
	Phone      string `json:"phone"`
	Telegram   string `json:"telegram"`
	TelegramID string `json:"telegram_id"`
	Email      string `json:"email" binding:"omitempty,email"`
	Position   string `json:"position"`
}

func (r *CreateCompanyHRRequest) ToDomain() *domain.CompanyHR {
	return &domain.CompanyHR{
		FirstName:  r.FirstName,
		LastName:   r.LastName,
		Patronymic: r.Patronymic,
		Phone:      r.Phone,
		Telegram:   r.Telegram,
		TelegramID: r.TelegramID,
		Email:      r.Email,
		Position:   r.Position,
	}
}

func (r *UpdateCompanyHRRequest) ToDomain(id uuid.UUID) *domain.CompanyHR {
	return &domain.CompanyHR{
		ID:         id,
		FirstName:  r.FirstName,
		LastName:   r.LastName,
		Patronymic: r.Patronymic,
		Phone:      r.Phone,
		Telegram:   r.Telegram,
		TelegramID: r.TelegramID,
		Email:      r.Email,
		Position:   r.Position,
	}
}

type CompanyHRResponse struct {
	ID          string              `json:"id"`
	FirstName   string              `json:"first_name"`
	LastName    string              `json:"last_name"`
	Patronymic  string              `json:"patronymic,omitempty"`
	Phone       string              `json:"phone,omitempty"`
	Telegram    string              `json:"telegram,omitempty"`
	TelegramID  string              `json:"telegram_id,omitempty"`
	Email       string              `json:"email,omitempty"`
	Position    string              `json:"position,omitempty"`
	Status      string              `json:"status"`
	CompanyData *domain.CompanyData `json:"company_data,omitempty"`
	CreatedAt   string              `json:"created_at"`
}

type PaginatedCompanyHRsResponse struct {
	HRs      []CompanyHRResponse `json:"hrs"`
	Total    int64               `json:"total"`
	Page     int32               `json:"page"`
	PageSize int32               `json:"page_size"`
}

type HRMiniAppUpdateRequest struct {
	FirstName  string                `json:"first_name"`
	LastName   string                `json:"last_name"`
	Patronymic string                `json:"patronymic"`
	Phone      string                `json:"phone"`
	Email      string                `json:"email" binding:"omitempty,email"`
	Position   string                `json:"position"`
	Company    *HRMiniAppCompanyData `json:"company"`
}

type HRMiniAppCompanyData struct {
	Name            string `json:"name"`
	ActivityType    string `json:"activity_type"`
	CompanyType     string `json:"company_type"`
	About           string `json:"about"`
	Market          string `json:"market"`
	EmployeeCount   int32  `json:"employee_count"`
	Country         string `json:"country"`
	Address         string `json:"address"`
	Phone           string `json:"phone"`
	Telegram        string `json:"telegram"`
	TelegramChannel string `json:"telegram_channel"`
	Email           string `json:"email"`
	LogoURL         string `json:"logo_url"`
	WebSite         string `json:"web_site"`
	Instagram       string `json:"instagram"`
}

type HRMiniAppMeResponse struct {
	CompanyHRResponse
}

func toCompanyHRResponse(hr *domain.CompanyHR) CompanyHRResponse {
	return CompanyHRResponse{
		ID:          hr.ID.String(),
		FirstName:   hr.FirstName,
		LastName:    hr.LastName,
		Patronymic:  hr.Patronymic,
		Phone:       hr.Phone,
		Telegram:    hr.Telegram,
		TelegramID:  hr.TelegramID,
		Email:       hr.Email,
		Position:    hr.Position,
		Status:      hr.Status,
		CompanyData: hr.CompanyData,
		CreatedAt:   hr.CreatedAt.Format(time.RFC3339),
	}
}
