package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
)

type VacancyService struct {
	repo         domain.VacancyRepository
	textRepo     domain.VacancyTextRepository
	skillSvc     *SkillService
	companySvc   *CompanyService
	geminiClient *gemini.Client
}

func NewVacancyService(repo domain.VacancyRepository, textRepo domain.VacancyTextRepository, skillSvc *SkillService, companySvc *CompanyService, geminiClient *gemini.Client) *VacancyService {
	return &VacancyService{repo: repo, textRepo: textRepo, skillSvc: skillSvc, companySvc: companySvc, geminiClient: geminiClient}
}

type CreateVacancyInput struct {
	HRID             uuid.UUID
	CompanyID        uuid.UUID
	CountryID        uuid.UUID
	SalaryMin        int32
	SalaryMax        int32
	SalaryCurrency   string
	ExperienceMin    int32
	ExperienceMax    int32
	Format           string
	Schedule         string
	Phone            string
	Telegram         string
	Email            string
	Address          string
	Title            string
	Description      string
	Responsibilities string
	Requirements     string
	Benefits         string
	Skills           []string
}

type VacancyWithDetails struct {
	Vacancy *domain.Vacancy
	Texts   []domain.VacancyText
	Skills  []domain.Skill
	Company *CompanyWithTexts
}

func (s *VacancyService) CreateVacancy(ctx context.Context, input *CreateVacancyInput) (*VacancyWithDetails, error) {
	textData := map[string]string{}
	if input.Title != "" {
		textData["title"] = input.Title
	}
	if input.Description != "" {
		textData["description"] = input.Description
	}
	if input.Responsibilities != "" {
		textData["responsibilities"] = input.Responsibilities
	}
	if input.Requirements != "" {
		textData["requirements"] = input.Requirements
	}
	if input.Benefits != "" {
		textData["benefits"] = input.Benefits
	}

	jsonBytes, _ := json.Marshal(textData)

	parsed, err := s.geminiClient.TranslateVacancy(ctx, string(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("gemini translate vacancy: %w", err)
	}

	vacancy := &domain.Vacancy{
		ID:             uuid.New(),
		HRID:           input.HRID,
		CompanyID:      input.CompanyID,
		CountryID:      input.CountryID,
		SalaryMin:      input.SalaryMin,
		SalaryMax:      input.SalaryMax,
		SalaryCurrency: input.SalaryCurrency,
		ExperienceMin:  input.ExperienceMin,
		ExperienceMax:  input.ExperienceMax,
		Format:         input.Format,
		Schedule:       input.Schedule,
		Phone:          input.Phone,
		Telegram:       input.Telegram,
		Email:          input.Email,
		Address:        input.Address,
		SourceLang:     parsed.SourceLang,
		Status:         "draft",
	}

	if vacancy.SalaryCurrency == "" {
		vacancy.SalaryCurrency = "USD"
	}
	if vacancy.Format == "" {
		vacancy.Format = "office"
	}
	if vacancy.Schedule == "" {
		vacancy.Schedule = "full-time"
	}

	created, err := s.repo.Create(ctx, vacancy)
	if err != nil {
		return nil, fmt.Errorf("service create vacancy: %w", err)
	}

	texts, err := s.storeVacancyTexts(ctx, created.ID, parsed.SourceLang, parsed.Fields)
	if err != nil {
		return nil, err
	}

	var skills []domain.Skill
	if len(input.Skills) > 0 {
		skills, err = s.skillSvc.SetVacancySkills(ctx, created.ID, input.Skills)
		if err != nil {
			return nil, fmt.Errorf("set vacancy skills: %w", err)
		}
	}

	company, _ := s.companySvc.GetCompany(ctx, created.CompanyID)

	return &VacancyWithDetails{Vacancy: created, Texts: texts, Skills: skills, Company: company}, nil
}

func (s *VacancyService) ParseVacancy(ctx context.Context, hrID, companyID uuid.UUID, userInput string) (*VacancyWithDetails, error) {
	parsed, err := s.geminiClient.ParseVacancyFromText(ctx, userInput)
	if err != nil {
		return nil, fmt.Errorf("gemini parse vacancy: %w", err)
	}

	salaryCurrency := parsed.SalaryCurrency
	if salaryCurrency == "" {
		salaryCurrency = "USD"
	}
	format := parsed.Format
	if format == "" {
		format = "office"
	}
	schedule := parsed.Schedule
	if schedule == "" {
		schedule = "full-time"
	}

	vacancy := &domain.Vacancy{
		ID:             uuid.New(),
		HRID:           hrID,
		CompanyID:      companyID,
		SalaryMin:      parsed.SalaryMin,
		SalaryMax:      parsed.SalaryMax,
		SalaryCurrency: salaryCurrency,
		ExperienceMin:  parsed.ExperienceMin,
		ExperienceMax:  parsed.ExperienceMax,
		Format:         format,
		Schedule:       schedule,
		Phone:          parsed.Phone,
		Telegram:       parsed.Telegram,
		Email:          parsed.Email,
		Address:        parsed.Address,
		SourceLang:     parsed.SourceLang,
		Status:         "draft",
	}

	created, err := s.repo.Create(ctx, vacancy)
	if err != nil {
		return nil, fmt.Errorf("service create vacancy: %w", err)
	}

	texts, err := s.storeVacancyTexts(ctx, created.ID, parsed.SourceLang, parsed.Fields)
	if err != nil {
		return nil, err
	}

	var skills []domain.Skill
	if len(parsed.Skills) > 0 {
		skills, err = s.skillSvc.SetVacancySkills(ctx, created.ID, parsed.Skills)
		if err != nil {
			return nil, fmt.Errorf("set vacancy skills: %w", err)
		}
	}

	company, _ := s.companySvc.GetCompany(ctx, created.CompanyID)

	return &VacancyWithDetails{Vacancy: created, Texts: texts, Skills: skills, Company: company}, nil
}

func (s *VacancyService) storeVacancyTexts(ctx context.Context, vacancyID uuid.UUID, sourceLang string, fields map[string]map[string]string) ([]domain.VacancyText, error) {
	langs := []string{"uz", "ru", "en"}
	modelVer := s.geminiClient.ModelVersion()
	texts := make([]domain.VacancyText, 0, 3)

	for _, lang := range langs {
		vt := &domain.VacancyText{
			VacancyID:    vacancyID,
			Lang:         lang,
			IsSource:     lang == sourceLang,
			ModelVersion: modelVer,
		}

		if translations, ok := fields["title"]; ok {
			vt.Title = translations[lang]
		}
		if translations, ok := fields["description"]; ok {
			vt.Description = translations[lang]
		}
		if translations, ok := fields["responsibilities"]; ok {
			vt.Responsibilities = translations[lang]
		}
		if translations, ok := fields["requirements"]; ok {
			vt.Requirements = translations[lang]
		}
		if translations, ok := fields["benefits"]; ok {
			vt.Benefits = translations[lang]
		}

		savedText, err := s.textRepo.Create(ctx, vt)
		if err != nil {
			return nil, fmt.Errorf("create vacancy text %s: %w", lang, err)
		}
		texts = append(texts, *savedText)
	}

	return texts, nil
}

func (s *VacancyService) GetVacancy(ctx context.Context, id uuid.UUID) (*VacancyWithDetails, error) {
	vacancy, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	texts, err := s.textRepo.ListByVacancy(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list vacancy texts: %w", err)
	}

	skills, err := s.skillSvc.ListVacancySkills(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list vacancy skills: %w", err)
	}

	company, _ := s.companySvc.GetCompany(ctx, vacancy.CompanyID)

	return &VacancyWithDetails{Vacancy: vacancy, Texts: texts, Skills: skills, Company: company}, nil
}

type ListVacanciesResult struct {
	Vacancies []VacancyWithDetails
	Total     int64
}

func (s *VacancyService) ListVacancies(ctx context.Context, page, pageSize int32) (*ListVacanciesResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("service count vacancies: %w", err)
	}

	vacancies, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("service list vacancies: %w", err)
	}

	result := make([]VacancyWithDetails, 0, len(vacancies))
	companyCache := map[uuid.UUID]*CompanyWithTexts{}
	for _, v := range vacancies {
		texts, err := s.textRepo.ListByVacancy(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list vacancy texts for %s: %w", v.ID, err)
		}
		skills, err := s.skillSvc.ListVacancySkills(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list vacancy skills for %s: %w", v.ID, err)
		}
		company, ok := companyCache[v.CompanyID]
		if !ok {
			company, _ = s.companySvc.GetCompany(ctx, v.CompanyID)
			companyCache[v.CompanyID] = company
		}
		result = append(result, VacancyWithDetails{Vacancy: &v, Texts: texts, Skills: skills, Company: company})
	}

	return &ListVacanciesResult{Vacancies: result, Total: total}, nil
}

func (s *VacancyService) ListVacanciesByCompany(ctx context.Context, companyID uuid.UUID, page, pageSize int32) (*ListVacanciesResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := s.repo.CountByCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("service count vacancies by company: %w", err)
	}

	vacancies, err := s.repo.ListByCompany(ctx, companyID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("service list vacancies by company: %w", err)
	}

	result := make([]VacancyWithDetails, 0, len(vacancies))
	var company *CompanyWithTexts
	if len(vacancies) > 0 {
		company, _ = s.companySvc.GetCompany(ctx, companyID)
	}
	for _, v := range vacancies {
		texts, err := s.textRepo.ListByVacancy(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list vacancy texts for %s: %w", v.ID, err)
		}
		skills, err := s.skillSvc.ListVacancySkills(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list vacancy skills for %s: %w", v.ID, err)
		}
		result = append(result, VacancyWithDetails{Vacancy: &v, Texts: texts, Skills: skills, Company: company})
	}

	return &ListVacanciesResult{Vacancies: result, Total: total}, nil
}

type UpdateVacancyInput struct {
	ID               uuid.UUID
	CountryID        uuid.UUID
	SalaryMin        int32
	SalaryMax        int32
	SalaryCurrency   string
	ExperienceMin    int32
	ExperienceMax    int32
	Format           string
	Schedule         string
	Phone            string
	Telegram         string
	Email            string
	Address          string
	Status           string
	Title            string
	Description      string
	Responsibilities string
	Requirements     string
	Benefits         string
	Skills           []string
}

func (s *VacancyService) UpdateVacancy(ctx context.Context, input *UpdateVacancyInput) (*VacancyWithDetails, error) {
	vacancy := &domain.Vacancy{
		ID:             input.ID,
		CountryID:      input.CountryID,
		SalaryMin:      input.SalaryMin,
		SalaryMax:      input.SalaryMax,
		SalaryCurrency: input.SalaryCurrency,
		ExperienceMin:  input.ExperienceMin,
		ExperienceMax:  input.ExperienceMax,
		Format:         input.Format,
		Schedule:       input.Schedule,
		Phone:          input.Phone,
		Telegram:       input.Telegram,
		Email:          input.Email,
		Address:        input.Address,
		Status:         input.Status,
	}

	updated, err := s.repo.Update(ctx, vacancy)
	if err != nil {
		return nil, err
	}

	// If any text field provided, retranslate via Gemini and replace all texts
	if input.Title != "" || input.Description != "" || input.Responsibilities != "" || input.Requirements != "" || input.Benefits != "" {
		textData := map[string]string{}
		if input.Title != "" {
			textData["title"] = input.Title
		}
		if input.Description != "" {
			textData["description"] = input.Description
		}
		if input.Responsibilities != "" {
			textData["responsibilities"] = input.Responsibilities
		}
		if input.Requirements != "" {
			textData["requirements"] = input.Requirements
		}
		if input.Benefits != "" {
			textData["benefits"] = input.Benefits
		}

		jsonBytes, _ := json.Marshal(textData)
		parsed, err := s.geminiClient.TranslateVacancy(ctx, string(jsonBytes))
		if err != nil {
			return nil, fmt.Errorf("gemini translate vacancy: %w", err)
		}

		if err := s.textRepo.DeleteByVacancy(ctx, updated.ID); err != nil {
			return nil, fmt.Errorf("delete old vacancy texts: %w", err)
		}

		if _, err := s.storeVacancyTexts(ctx, updated.ID, parsed.SourceLang, parsed.Fields); err != nil {
			return nil, err
		}
	}

	// If skills provided, update them
	if input.Skills != nil {
		if _, err := s.skillSvc.SetVacancySkills(ctx, updated.ID, input.Skills); err != nil {
			return nil, fmt.Errorf("set vacancy skills: %w", err)
		}
	}

	texts, err := s.textRepo.ListByVacancy(ctx, updated.ID)
	if err != nil {
		return nil, fmt.Errorf("list vacancy texts: %w", err)
	}

	skills, err := s.skillSvc.ListVacancySkills(ctx, updated.ID)
	if err != nil {
		return nil, fmt.Errorf("list vacancy skills: %w", err)
	}

	company, _ := s.companySvc.GetCompany(ctx, updated.CompanyID)

	return &VacancyWithDetails{Vacancy: updated, Texts: texts, Skills: skills, Company: company}, nil
}

func (s *VacancyService) DeleteVacancy(ctx context.Context, id uuid.UUID) error {
	if err := s.skillSvc.RemoveVacancySkills(ctx, id); err != nil {
		return fmt.Errorf("remove vacancy skills: %w", err)
	}
	if err := s.textRepo.DeleteByVacancy(ctx, id); err != nil {
		return fmt.Errorf("delete vacancy texts: %w", err)
	}
	return s.repo.Delete(ctx, id)
}
