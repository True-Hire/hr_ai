package http

import (
	"time"

	"github.com/google/uuid"
	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CreateVacancyRequest struct {
	HRID             string   `json:"hr_id" binding:"required"`
	CountryID        string   `json:"country_id"`
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
	HRID      string `json:"hr_id" binding:"required"`
	UserInput string `json:"user_input" binding:"required"`
}

type UpdateVacancyRequest struct {
	CountryID        string   `json:"country_id"`
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
	Status           string   `json:"status"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Responsibilities string   `json:"responsibilities"`
	Requirements     string   `json:"requirements"`
	Benefits         string   `json:"benefits"`
	Skills           []string `json:"skills"`
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

type VacancyCompanyResponse struct {
	Name          string `json:"name"`
	ActivityType  string `json:"activity_type,omitempty"`
	CompanyType   string `json:"company_type,omitempty"`
	About         string `json:"about,omitempty"`
	LogoURL       string `json:"logo_url,omitempty"`
	Country       string `json:"country,omitempty"`
	EmployeeCount int32  `json:"employee_count,omitempty"`
}

type VacancyResponse struct {
	ID             string                  `json:"id"`
	HRID           string                  `json:"hr_id"`
	CountryID      string                  `json:"country_id,omitempty"`
	SalaryMin      int32                   `json:"salary_min,omitempty"`
	SalaryMax      int32                   `json:"salary_max,omitempty"`
	SalaryCurrency string                  `json:"salary_currency"`
	ExperienceMin  int32                   `json:"experience_min,omitempty"`
	ExperienceMax  int32                   `json:"experience_max,omitempty"`
	Format         string                  `json:"format"`
	Schedule       string                  `json:"schedule"`
	Phone          string                  `json:"phone,omitempty"`
	Telegram       string                  `json:"telegram,omitempty"`
	Email          string                  `json:"email,omitempty"`
	Address        string                  `json:"address,omitempty"`
	Status                  string                  `json:"status"`
	SourceLang              string                  `json:"source_lang"`
	CreatedAt               string                  `json:"created_at"`
	Company                 *VacancyCompanyResponse `json:"company,omitempty"`
	Text                    *VacancyTextResponse    `json:"text"`
	Skills                  []SkillResponse         `json:"skills"`
	ApplicationCount        *int64                  `json:"application_count,omitempty"`
	MatchingCandidatesCount *int                    `json:"matching_candidates_count,omitempty"`
}

type PaginatedVacanciesResponse struct {
	Vacancies []VacancyResponse `json:"vacancies"`
	Total     int64             `json:"total"`
	Page      int32             `json:"page"`
	PageSize  int32             `json:"page_size"`
}

func toVacancyResponse(vwd *application.VacancyWithDetails, lang string) VacancyResponse {
	var countryID string
	if vwd.Vacancy.CountryID != uuid.Nil {
		countryID = vwd.Vacancy.CountryID.String()
	}

	resp := VacancyResponse{
		ID:             vwd.Vacancy.ID.String(),
		HRID:           vwd.Vacancy.HRID.String(),
		CountryID:      countryID,
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
		Skills:         make([]SkillResponse, 0, len(vwd.Skills)),
	}

	// Pick the text for the requested language, fallback to English
	for _, t := range vwd.Texts {
		if t.Lang == lang {
			tr := toVacancyTextResponse(&t)
			resp.Text = &tr
			break
		}
	}
	if resp.Text == nil {
		for _, t := range vwd.Texts {
			if t.Lang == "en" {
				tr := toVacancyTextResponse(&t)
				resp.Text = &tr
				break
			}
		}
	}

	// Build company response from embedded CompanyData
	if vwd.Vacancy.CompanyData != nil {
		cd := vwd.Vacancy.CompanyData
		cr := &VacancyCompanyResponse{
			LogoURL:       cd.LogoURL,
			Country:       cd.Country,
			EmployeeCount: cd.EmployeeCount,
		}
		for _, t := range cd.Texts {
			if t.Lang == lang {
				cr.Name = t.Name
				cr.ActivityType = t.ActivityType
				cr.CompanyType = t.CompanyType
				cr.About = t.About
				break
			}
		}
		if cr.Name == "" {
			for _, t := range cd.Texts {
				if t.Lang == "en" {
					cr.Name = t.Name
					cr.ActivityType = t.ActivityType
					cr.CompanyType = t.CompanyType
					cr.About = t.About
					break
				}
			}
		}
		resp.Company = cr
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
