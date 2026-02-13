package http

import (
	"time"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CreateCompanyHRRequest struct {
	FirstName     string `json:"first_name" binding:"required"`
	LastName      string `json:"last_name" binding:"required"`
	Patronymic    string `json:"patronymic"`
	Phone         string `json:"phone"`
	Telegram      string `json:"telegram"`
	TelegramID    string `json:"telegram_id"`
	Email         string `json:"email" binding:"omitempty,email"`
	Position      string `json:"position"`
	CompanyName   string `json:"company_name"`
	ActivityType  string `json:"activity_type"`
	CompanyType   string `json:"company_type"`
	EmployeeCount int32  `json:"employee_count"`
	Country       string `json:"country"`
	Market        string `json:"market"`
	WebSite       string `json:"web_site"`
	About         string `json:"about"`
	LogoURL       string `json:"logo_url"`
	Instagram     string `json:"instagram"`
	Password      string `json:"password" binding:"required,min=6"`
}

type UpdateCompanyHRRequest struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Patronymic    string `json:"patronymic"`
	Phone         string `json:"phone"`
	Telegram      string `json:"telegram"`
	TelegramID    string `json:"telegram_id"`
	Email         string `json:"email" binding:"omitempty,email"`
	Position      string `json:"position"`
	CompanyName   string `json:"company_name"`
	ActivityType  string `json:"activity_type"`
	CompanyType   string `json:"company_type"`
	EmployeeCount int32  `json:"employee_count"`
	Country       string `json:"country"`
	Market        string `json:"market"`
	WebSite       string `json:"web_site"`
	About         string `json:"about"`
	LogoURL       string `json:"logo_url"`
	Instagram     string `json:"instagram"`
}

func (r *CreateCompanyHRRequest) ToDomain() *domain.CompanyHR {
	return &domain.CompanyHR{
		FirstName:     r.FirstName,
		LastName:      r.LastName,
		Patronymic:    r.Patronymic,
		Phone:         r.Phone,
		Telegram:      r.Telegram,
		TelegramID:    r.TelegramID,
		Email:         r.Email,
		Position:      r.Position,
		CompanyName:   r.CompanyName,
		ActivityType:  r.ActivityType,
		CompanyType:   r.CompanyType,
		EmployeeCount: r.EmployeeCount,
		Country:       r.Country,
		Market:        r.Market,
		WebSite:       r.WebSite,
		About:         r.About,
		LogoURL:       r.LogoURL,
		Instagram:     r.Instagram,
	}
}

func (r *UpdateCompanyHRRequest) ToDomain(id uuid.UUID) *domain.CompanyHR {
	return &domain.CompanyHR{
		ID:            id,
		FirstName:     r.FirstName,
		LastName:      r.LastName,
		Patronymic:    r.Patronymic,
		Phone:         r.Phone,
		Telegram:      r.Telegram,
		TelegramID:    r.TelegramID,
		Email:         r.Email,
		Position:      r.Position,
		CompanyName:   r.CompanyName,
		ActivityType:  r.ActivityType,
		CompanyType:   r.CompanyType,
		EmployeeCount: r.EmployeeCount,
		Country:       r.Country,
		Market:        r.Market,
		WebSite:       r.WebSite,
		About:         r.About,
		LogoURL:       r.LogoURL,
		Instagram:     r.Instagram,
	}
}

type CompanyHRResponse struct {
	ID            string `json:"id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Patronymic    string `json:"patronymic,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Telegram      string `json:"telegram,omitempty"`
	TelegramID    string `json:"telegram_id,omitempty"`
	Email         string `json:"email,omitempty"`
	Position      string `json:"position,omitempty"`
	Status        string `json:"status"`
	CompanyName   string `json:"company_name,omitempty"`
	ActivityType  string `json:"activity_type,omitempty"`
	CompanyType   string `json:"company_type,omitempty"`
	EmployeeCount int32  `json:"employee_count,omitempty"`
	Country       string `json:"country,omitempty"`
	Market        string `json:"market,omitempty"`
	WebSite       string `json:"web_site,omitempty"`
	About         string `json:"about,omitempty"`
	LogoURL       string `json:"logo_url,omitempty"`
	Instagram     string `json:"instagram,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type PaginatedCompanyHRsResponse struct {
	HRs      []CompanyHRResponse `json:"hrs"`
	Total    int64               `json:"total"`
	Page     int32               `json:"page"`
	PageSize int32               `json:"page_size"`
}

func toCompanyHRResponse(hr *domain.CompanyHR) CompanyHRResponse {
	return CompanyHRResponse{
		ID:            hr.ID.String(),
		FirstName:     hr.FirstName,
		LastName:      hr.LastName,
		Patronymic:    hr.Patronymic,
		Phone:         hr.Phone,
		Telegram:      hr.Telegram,
		TelegramID:    hr.TelegramID,
		Email:         hr.Email,
		Position:      hr.Position,
		Status:        hr.Status,
		CompanyName:   hr.CompanyName,
		ActivityType:  hr.ActivityType,
		CompanyType:   hr.CompanyType,
		EmployeeCount: hr.EmployeeCount,
		Country:       hr.Country,
		Market:        hr.Market,
		WebSite:       hr.WebSite,
		About:         hr.About,
		LogoURL:       hr.LogoURL,
		Instagram:     hr.Instagram,
		CreatedAt:     hr.CreatedAt.Format(time.RFC3339),
	}
}
