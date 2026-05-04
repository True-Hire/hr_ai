package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
)

type VacancyService struct {
	repo             domain.VacancyRepository
	textRepo         domain.VacancyTextRepository
	workerRepo       domain.VacancyWorkerRepository
	skillSvc         *SkillService
	geminiClient     *gemini.Client
	vectorIndexSvc   *VectorIndexService
	candidateSearchSvc *CandidateSearchService
}

func NewVacancyService(repo domain.VacancyRepository, textRepo domain.VacancyTextRepository, workerRepo domain.VacancyWorkerRepository, skillSvc *SkillService, geminiClient *gemini.Client, vectorIndexSvc *VectorIndexService) *VacancyService {
	return &VacancyService{
		repo:           repo,
		textRepo:       textRepo,
		workerRepo:     workerRepo,
		skillSvc:       skillSvc,
		geminiClient:   geminiClient,
		vectorIndexSvc: vectorIndexSvc,
	}
}

func (s *VacancyService) SetCandidateSearchService(svc *CandidateSearchService) {
	s.candidateSearchSvc = svc
}

type CreateVacancyInput struct {
	HRID             uuid.UUID
	CompanyData      *domain.CompanyData
	CountryID        uuid.UUID
	MainCategoryID   uuid.UUID
	SubCategoryID    uuid.UUID
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
		CompanyData:    input.CompanyData,
		CountryID:      input.CountryID,
		MainCategoryID: input.MainCategoryID,
		SubCategoryID:  input.SubCategoryID,
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
		Status:         domain.VacancyStatusActive,
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

	if s.vectorIndexSvc != nil {
		go func() {
			if err := s.vectorIndexSvc.IndexVacancy(context.Background(), created.ID); err != nil {
				log.Printf("index vacancy %s: %v", created.ID, err)
			}
		}()
	}

	// Trigger automated matching
	go func() {
		if err := s.CalculateAndSaveMatches(context.Background(), created.ID, created.HRID); err != nil {
			log.Printf("failed to calculate matches for vacancy %s: %v", created.ID, err)
		}
	}()

	return &VacancyWithDetails{Vacancy: created, Texts: texts, Skills: skills}, nil
}

func (s *VacancyService) ParseVacancy(ctx context.Context, hrID uuid.UUID, companyData *domain.CompanyData, userInput string) (*VacancyWithDetails, error) {
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
		CompanyData:    companyData,
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
		Status:         domain.VacancyStatusActive,
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
	draftSkills := parsed.Skills["en"]
	if len(draftSkills) == 0 {
		draftSkills = parsed.Skills[parsed.SourceLang]
	}
	if len(draftSkills) > 0 {
		skills, err = s.skillSvc.SetVacancySkills(ctx, created.ID, draftSkills)
		if err != nil {
			return nil, fmt.Errorf("set vacancy skills: %w", err)
		}
	}

	if s.vectorIndexSvc != nil {
		go func() {
			if err := s.vectorIndexSvc.IndexVacancy(context.Background(), created.ID); err != nil {
				log.Printf("index vacancy %s: %v", created.ID, err)
			}
		}()
	}

	// Trigger automated matching
	go func() {
		if err := s.CalculateAndSaveMatches(context.Background(), created.ID, created.HRID); err != nil {
			log.Printf("failed to calculate matches for vacancy %s: %v", created.ID, err)
		}
	}()

	return &VacancyWithDetails{Vacancy: created, Texts: texts, Skills: skills}, nil
}

func (s *VacancyService) CreateVacancyFromParsed(ctx context.Context, hrID uuid.UUID, companyData *domain.CompanyData, parsed *gemini.ParsedVacancyFull) (*VacancyWithDetails, error) {
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
		CompanyData:    companyData,
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
		Status:         domain.VacancyStatusActive,
	}

	created, err := s.repo.Create(ctx, vacancy)
	if err != nil {
		return nil, fmt.Errorf("service create vacancy from parsed: %w", err)
	}

	texts, err := s.storeVacancyTexts(ctx, created.ID, parsed.SourceLang, parsed.Fields)
	if err != nil {
		return nil, err
	}

	var skills []domain.Skill
	draftSkills := parsed.Skills["en"]
	if len(draftSkills) == 0 {
		draftSkills = parsed.Skills[parsed.SourceLang]
	}
	if len(draftSkills) > 0 {
		skills, err = s.skillSvc.SetVacancySkills(ctx, created.ID, draftSkills)
		if err != nil {
			return nil, fmt.Errorf("set vacancy skills: %w", err)
		}
	}

	if s.vectorIndexSvc != nil {
		go func() {
			if err := s.vectorIndexSvc.IndexVacancy(context.Background(), created.ID); err != nil {
				log.Printf("index vacancy %s: %v", created.ID, err)
			}
		}()
	}

	// Trigger automated matching
	go func() {
		if err := s.CalculateAndSaveMatches(context.Background(), created.ID, created.HRID); err != nil {
			log.Printf("failed to calculate matches for vacancy %s: %v", created.ID, err)
		}
	}()

	return &VacancyWithDetails{Vacancy: created, Texts: texts, Skills: skills}, nil
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

	return &VacancyWithDetails{Vacancy: vacancy, Texts: texts, Skills: skills}, nil
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
	for _, v := range vacancies {
		texts, err := s.textRepo.ListByVacancy(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list vacancy texts for %s: %w", v.ID, err)
		}
		skills, err := s.skillSvc.ListVacancySkills(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list vacancy skills for %s: %w", v.ID, err)
		}
		result = append(result, VacancyWithDetails{Vacancy: &v, Texts: texts, Skills: skills})
	}

	return &ListVacanciesResult{Vacancies: result, Total: total}, nil
}

func (s *VacancyService) ListVacanciesByHR(ctx context.Context, hrID uuid.UUID, page, pageSize int32) (*ListVacanciesResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := s.repo.CountByHR(ctx, hrID)
	if err != nil {
		return nil, fmt.Errorf("service count vacancies by hr: %w", err)
	}

	vacancies, err := s.repo.ListByHR(ctx, hrID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("service list vacancies by hr: %w", err)
	}

	result := make([]VacancyWithDetails, 0, len(vacancies))
	for _, v := range vacancies {
		texts, err := s.textRepo.ListByVacancy(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list vacancy texts for %s: %w", v.ID, err)
		}
		skills, err := s.skillSvc.ListVacancySkills(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list vacancy skills for %s: %w", v.ID, err)
		}
		result = append(result, VacancyWithDetails{Vacancy: &v, Texts: texts, Skills: skills})
	}

	return &ListVacanciesResult{Vacancies: result, Total: total}, nil
}

func (s *VacancyService) ListVacanciesFiltered(ctx context.Context, filter domain.VacancyFilter, page, pageSize int32) (*ListVacanciesResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := s.repo.CountFiltered(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("service count filtered vacancies: %w", err)
	}

	vacancies, err := s.repo.ListFiltered(ctx, filter, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("service list filtered vacancies: %w", err)
	}

	result := make([]VacancyWithDetails, 0, len(vacancies))
	for _, v := range vacancies {
		texts, err := s.textRepo.ListByVacancy(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list vacancy texts for %s: %w", v.ID, err)
		}
		skills, err := s.skillSvc.ListVacancySkills(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list vacancy skills for %s: %w", v.ID, err)
		}
		result = append(result, VacancyWithDetails{Vacancy: &v, Texts: texts, Skills: skills})
	}

	return &ListVacanciesResult{Vacancies: result, Total: total}, nil
}

type UpdateVacancyInput struct {
	ID               uuid.UUID
	CountryID        uuid.UUID
	MainCategoryID   uuid.UUID
	SubCategoryID    uuid.UUID
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
		MainCategoryID: input.MainCategoryID,
		SubCategoryID:  input.SubCategoryID,
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

	if s.vectorIndexSvc != nil {
		go func() {
			if err := s.vectorIndexSvc.IndexVacancy(context.Background(), updated.ID); err != nil {
				log.Printf("reindex vacancy %s: %v", updated.ID, err)
			}
		}()
	}

	// Refresh automated matching
	go func() {
		if err := s.CalculateAndSaveMatches(context.Background(), updated.ID, updated.HRID); err != nil {
			log.Printf("failed to refresh matches for vacancy %s: %v", updated.ID, err)
		}
	}()

	return &VacancyWithDetails{Vacancy: updated, Texts: texts, Skills: skills}, nil
}

func (s *VacancyService) UpdateVacancyFromParsed(ctx context.Context, vacancyID uuid.UUID, parsed *gemini.ParsedVacancyFull) (*VacancyWithDetails, error) {
	existing, err := s.repo.GetByID(ctx, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("get vacancy for update: %w", err)
	}

	// Update fields from parsed data
	if parsed.SalaryMin > 0 {
		existing.SalaryMin = parsed.SalaryMin
	}
	if parsed.SalaryMax > 0 {
		existing.SalaryMax = parsed.SalaryMax
	}
	if parsed.SalaryCurrency != "" {
		existing.SalaryCurrency = parsed.SalaryCurrency
	}
	if parsed.ExperienceMin > 0 {
		existing.ExperienceMin = parsed.ExperienceMin
	}
	if parsed.ExperienceMax > 0 {
		existing.ExperienceMax = parsed.ExperienceMax
	}
	if parsed.Format != "" {
		existing.Format = parsed.Format
	}
	if parsed.Schedule != "" {
		existing.Schedule = parsed.Schedule
	}
	if parsed.Phone != "" {
		existing.Phone = parsed.Phone
	}
	if parsed.Telegram != "" {
		existing.Telegram = parsed.Telegram
	}
	if parsed.Email != "" {
		existing.Email = parsed.Email
	}
	if parsed.Address != "" {
		existing.Address = parsed.Address
	}

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("update vacancy: %w", err)
	}

	// Replace texts
	if err := s.textRepo.DeleteByVacancy(ctx, updated.ID); err != nil {
		return nil, fmt.Errorf("delete old vacancy texts: %w", err)
	}
	texts, err := s.storeVacancyTexts(ctx, updated.ID, parsed.SourceLang, parsed.Fields)
	if err != nil {
		return nil, err
	}

	// Update skills
	var skills []domain.Skill
	draftSkills := parsed.Skills["en"]
	if len(draftSkills) == 0 {
		draftSkills = parsed.Skills[parsed.SourceLang]
	}
	if len(draftSkills) > 0 {
		skills, err = s.skillSvc.SetVacancySkills(ctx, updated.ID, draftSkills)
		if err != nil {
			return nil, fmt.Errorf("set vacancy skills: %w", err)
		}
	} else {
		skills, _ = s.skillSvc.ListVacancySkills(ctx, updated.ID)
	}

	if s.vectorIndexSvc != nil {
		go func() {
			if err := s.vectorIndexSvc.IndexVacancy(context.Background(), updated.ID); err != nil {
				log.Printf("reindex vacancy %s after edit: %v", updated.ID, err)
			}
		}()
	}

	// Refresh automated matching
	go func() {
		if err := s.CalculateAndSaveMatches(context.Background(), updated.ID, updated.HRID); err != nil {
			log.Printf("failed to refresh matches for vacancy %s: %v", updated.ID, err)
		}
	}()

	return &VacancyWithDetails{Vacancy: updated, Texts: texts, Skills: skills}, nil
}

func (s *VacancyService) UpdateVacancyStatus(ctx context.Context, id uuid.UUID, status string) error {
	v, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get vacancy: %w", err)
	}
	v.Status = status
	_, err = s.repo.Update(ctx, v)
	return err
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

func (s *VacancyService) ListVacancyMatches(ctx context.Context, vacancyID uuid.UUID) ([]domain.VacancyWorker, error) {
	return s.workerRepo.ListByVacancy(ctx, vacancyID)
}

func (s *VacancyService) CalculateAndSaveMatches(ctx context.Context, vacancyID, hrID uuid.UUID) error {
	if s.candidateSearchSvc == nil {
		return fmt.Errorf("candidate search service not initialized")
	}

	// Find top 20 candidates
	page, err := s.candidateSearchSvc.SearchByVacancy(ctx, vacancyID, hrID, 20)
	if err != nil {
		return fmt.Errorf("search by vacancy: %w", err)
	}

	workers := make([]domain.VacancyWorker, 0, len(page.Items))
	for i, item := range page.Items {
		workers = append(workers, domain.VacancyWorker{
			ID:              uuid.New(),
			VacancyID:       vacancyID,
			UserID:          item.UserID,
			MatchPercentage: item.MatchPercentage,
			MatchScore:      item.FinalScore,
			Rank:            i + 1,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		})
	}

	// Delete old matches if any (e.g. on re-calculation)
	_ = s.workerRepo.DeleteByVacancy(ctx, vacancyID)

	if err := s.workerRepo.BulkCreate(ctx, workers); err != nil {
		return fmt.Errorf("bulk create vacancy workers: %w", err)
	}

	return nil
}
