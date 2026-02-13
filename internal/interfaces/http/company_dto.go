package http

import (
	"time"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CreateCompanyRequest struct {
	Name            string `json:"name" binding:"required"`
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
	Email           string `json:"email" binding:"omitempty,email"`
	LogoURL         string `json:"logo_url"`
	WebSite         string `json:"web_site"`
	Instagram       string `json:"instagram"`
}

type UpdateCompanyRequest struct {
	EmployeeCount   int32  `json:"employee_count"`
	Country         string `json:"country"`
	Address         string `json:"address"`
	Phone           string `json:"phone"`
	Telegram        string `json:"telegram"`
	TelegramChannel string `json:"telegram_channel"`
	Email           string `json:"email" binding:"omitempty,email"`
	LogoURL         string `json:"logo_url"`
	WebSite         string `json:"web_site"`
	Instagram       string `json:"instagram"`
}

type CompanyTextResponse struct {
	Lang         string `json:"lang"`
	Name         string `json:"name"`
	ActivityType string `json:"activity_type,omitempty"`
	CompanyType  string `json:"company_type,omitempty"`
	About        string `json:"about,omitempty"`
	Market       string `json:"market,omitempty"`
	IsSource     bool   `json:"is_source"`
	ModelVersion string `json:"model_version,omitempty"`
	UpdatedAt    string `json:"updated_at"`
}

type CompanyResponse struct {
	ID              string                `json:"id"`
	EmployeeCount   int32                 `json:"employee_count,omitempty"`
	Country         string                `json:"country,omitempty"`
	Address         string                `json:"address,omitempty"`
	Phone           string                `json:"phone,omitempty"`
	Telegram        string                `json:"telegram,omitempty"`
	TelegramChannel string                `json:"telegram_channel,omitempty"`
	Email           string                `json:"email,omitempty"`
	LogoURL         string                `json:"logo_url,omitempty"`
	WebSite         string                `json:"web_site,omitempty"`
	Instagram       string                `json:"instagram,omitempty"`
	SourceLang      string                `json:"source_lang"`
	CreatedAt       string                `json:"created_at"`
	Texts           []CompanyTextResponse `json:"texts"`
}

type PaginatedCompaniesResponse struct {
	Companies []CompanyResponse `json:"companies"`
	Total     int64             `json:"total"`
	Page      int32             `json:"page"`
	PageSize  int32             `json:"page_size"`
}

func toCompanyResponse(cwt *application.CompanyWithTexts) CompanyResponse {
	resp := CompanyResponse{
		ID:              cwt.Company.ID.String(),
		EmployeeCount:   cwt.Company.EmployeeCount,
		Country:         cwt.Company.Country,
		Address:         cwt.Company.Address,
		Phone:           cwt.Company.Phone,
		Telegram:        cwt.Company.Telegram,
		TelegramChannel: cwt.Company.TelegramChannel,
		Email:           cwt.Company.Email,
		LogoURL:         cwt.Company.LogoURL,
		WebSite:         cwt.Company.WebSite,
		Instagram:       cwt.Company.Instagram,
		SourceLang:      cwt.Company.SourceLang,
		CreatedAt:       cwt.Company.CreatedAt.Format(time.RFC3339),
		Texts:           make([]CompanyTextResponse, 0, len(cwt.Texts)),
	}
	for _, t := range cwt.Texts {
		resp.Texts = append(resp.Texts, toCompanyTextResponse(&t))
	}
	return resp
}

func toCompanyTextResponse(ct *domain.CompanyText) CompanyTextResponse {
	return CompanyTextResponse{
		Lang:         ct.Lang,
		Name:         ct.Name,
		ActivityType: ct.ActivityType,
		CompanyType:  ct.CompanyType,
		About:        ct.About,
		Market:       ct.Market,
		IsSource:     ct.IsSource,
		ModelVersion: ct.ModelVersion,
		UpdatedAt:    ct.UpdatedAt.Format(time.RFC3339),
	}
}
