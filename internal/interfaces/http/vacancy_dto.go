package http

import (
	"time"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CreateVacancyRequest struct {
	CompanyID        string   `json:"company_id" binding:"required"`
	Title            string   `json:"title" binding:"required"`
	Description      string   `json:"description"`
	Responsibilities string   `json:"responsibilities"`
	Requirements     string   `json:"requirements"`
	Benefits         string   `json:"benefits"`
	SalaryMin        int32    `json:"salary_min"`
	SalaryMax        int32    `json:"salary_max"`
	SalaryCurrency   string   `json:"salary_currency"`
	ExperienceMin    int32    `json:"experience_min"`
	ExperienceMax    int32    `json:"experience_max"`
	Format           string   `json:"format"`
	Schedule         string   `json:"schedule"`
	Phone            string   `json:"phone"`
	Telegram         string   `json:"telegram"`
	Email            string   `json:"email" binding:"omitempty,email"`
	Address          string   `json:"address"`
	Skills           []string `json:"skills"`
}

type VacancyParseRequest struct {
	CompanyID string `json:"company_id" binding:"required"`
	UserInput string `json:"user_input" binding:"required"`
}

type UpdateVacancyRequest struct {
	SalaryMin      int32  `json:"salary_min"`
	SalaryMax      int32  `json:"salary_max"`
	SalaryCurrency string `json:"salary_currency"`
	ExperienceMin  int32  `json:"experience_min"`
	ExperienceMax  int32  `json:"experience_max"`
	Format         string `json:"format"`
	Schedule       string `json:"schedule"`
	Phone          string `json:"phone"`
	Telegram       string `json:"telegram"`
	Email          string `json:"email" binding:"omitempty,email"`
	Address        string `json:"address"`
	Status         string `json:"status"`
}

type VacancyTextResponse struct {
	Lang             string `json:"lang"`
	Title            string `json:"title"`
	Description      string `json:"description,omitempty"`
	Responsibilities string `json:"responsibilities,omitempty"`
	Requirements     string `json:"requirements,omitempty"`
	Benefits         string `json:"benefits,omitempty"`
	IsSource         bool   `json:"is_source"`
	ModelVersion     string `json:"model_version,omitempty"`
	UpdatedAt        string `json:"updated_at"`
}

type VacancyResponse struct {
	ID             string                `json:"id"`
	HRID           string                `json:"hr_id"`
	CompanyID      string                `json:"company_id"`
	SalaryMin      int32                 `json:"salary_min,omitempty"`
	SalaryMax      int32                 `json:"salary_max,omitempty"`
	SalaryCurrency string                `json:"salary_currency"`
	ExperienceMin  int32                 `json:"experience_min,omitempty"`
	ExperienceMax  int32                 `json:"experience_max,omitempty"`
	Format         string                `json:"format"`
	Schedule       string                `json:"schedule"`
	Phone          string                `json:"phone,omitempty"`
	Telegram       string                `json:"telegram,omitempty"`
	Email          string                `json:"email,omitempty"`
	Address        string                `json:"address,omitempty"`
	Status         string                `json:"status"`
	SourceLang     string                `json:"source_lang"`
	CreatedAt      string                `json:"created_at"`
	Texts          []VacancyTextResponse `json:"texts"`
	Skills         []SkillResponse       `json:"skills"`
}

type PaginatedVacanciesResponse struct {
	Vacancies []VacancyResponse `json:"vacancies"`
	Total     int64             `json:"total"`
	Page      int32             `json:"page"`
	PageSize  int32             `json:"page_size"`
}

func toVacancyResponse(vwd *application.VacancyWithDetails) VacancyResponse {
	resp := VacancyResponse{
		ID:             vwd.Vacancy.ID.String(),
		HRID:           vwd.Vacancy.HRID.String(),
		CompanyID:      vwd.Vacancy.CompanyID.String(),
		SalaryMin:      vwd.Vacancy.SalaryMin,
		SalaryMax:      vwd.Vacancy.SalaryMax,
		SalaryCurrency: vwd.Vacancy.SalaryCurrency,
		ExperienceMin:  vwd.Vacancy.ExperienceMin,
		ExperienceMax:  vwd.Vacancy.ExperienceMax,
		Format:         vwd.Vacancy.Format,
		Schedule:       vwd.Vacancy.Schedule,
		Phone:          vwd.Vacancy.Phone,
		Telegram:       vwd.Vacancy.Telegram,
		Email:          vwd.Vacancy.Email,
		Address:        vwd.Vacancy.Address,
		Status:         vwd.Vacancy.Status,
		SourceLang:     vwd.Vacancy.SourceLang,
		CreatedAt:      vwd.Vacancy.CreatedAt.Format(time.RFC3339),
		Texts:          make([]VacancyTextResponse, 0, len(vwd.Texts)),
		Skills:         make([]SkillResponse, 0, len(vwd.Skills)),
	}
	for _, t := range vwd.Texts {
		resp.Texts = append(resp.Texts, toVacancyTextResponse(&t))
	}
	for _, s := range vwd.Skills {
		resp.Skills = append(resp.Skills, SkillResponse{
			ID:   s.ID.String(),
			Name: s.Name,
		})
	}
	return resp
}

func toVacancyTextResponse(vt *domain.VacancyText) VacancyTextResponse {
	return VacancyTextResponse{
		Lang:             vt.Lang,
		Title:            vt.Title,
		Description:      vt.Description,
		Responsibilities: vt.Responsibilities,
		Requirements:     vt.Requirements,
		Benefits:         vt.Benefits,
		IsSource:         vt.IsSource,
		ModelVersion:     vt.ModelVersion,
		UpdatedAt:        vt.UpdatedAt.Format(time.RFC3339),
	}
}
