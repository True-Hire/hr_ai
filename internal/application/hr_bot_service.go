package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
)

type HRBotService struct {
	hrSvc         *CompanyHRService
	vacancySvc    *VacancyService
	vacancyAppSvc *VacancyApplicationService
	stateSvc      *BotStateService
	searchSvc     *SearchService
	userSvc       *UserService
	geminiClient  *gemini.Client
}

func NewHRBotService(hrSvc *CompanyHRService, vacancySvc *VacancyService, vacancyAppSvc *VacancyApplicationService, stateSvc *BotStateService, searchSvc *SearchService, userSvc *UserService, geminiClient *gemini.Client) *HRBotService {
	return &HRBotService{
		hrSvc:         hrSvc,
		vacancySvc:    vacancySvc,
		vacancyAppSvc: vacancyAppSvc,
		stateSvc:      stateSvc,
		searchSvc:     searchSvc,
		userSvc:       userSvc,
		geminiClient:  geminiClient,
	}
}

type HRStartResult struct {
	HR    *domain.CompanyHR
	IsNew bool
}

func (s *HRBotService) HandleStart(ctx context.Context, telegramID int64) (*HRStartResult, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	existingHR, err := s.hrSvc.GetByTelegramID(ctx, tgID)
	if err == nil {
		return &HRStartResult{HR: existingHR, IsNew: false}, nil
	}
	if !errors.Is(err, domain.ErrCompanyHRNotFound) {
		return nil, fmt.Errorf("check existing hr: %w", err)
	}

	if err := s.SetState(ctx, telegramID, domain.HRBotStateSharingPhone, nil); err != nil {
		return nil, fmt.Errorf("set sharing phone state: %w", err)
	}

	return &HRStartResult{IsNew: true}, nil
}

func (s *HRBotService) HandlePhoneShared(ctx context.Context, telegramID int64, phone, firstName, lastName, username, language string) (*domain.CompanyHR, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	var tg string
	if username != "" {
		tg = "@" + username
	}

	hr := &domain.CompanyHR{
		FirstName:  firstName,
		LastName:   lastName,
		Phone:      phone,
		Telegram:   tg,
		TelegramID: tgID,
		Language:   language,
	}

	created, err := s.hrSvc.CreateCompanyHR(ctx, hr)
	if err != nil {
		return nil, fmt.Errorf("create company hr: %w", err)
	}

	if err := s.ClearState(ctx, telegramID); err != nil {
		return nil, fmt.Errorf("clear state: %w", err)
	}

	return created, nil
}

func (s *HRBotService) GetHRByTelegramID(ctx context.Context, telegramID string) (*domain.CompanyHR, error) {
	return s.hrSvc.GetByTelegramID(ctx, telegramID)
}

func (s *HRBotService) UpdateLanguage(ctx context.Context, hrID uuid.UUID, language string) (*domain.CompanyHR, error) {
	return s.hrSvc.UpdateLanguage(ctx, hrID, language)
}

func (s *HRBotService) ParseVacancy(ctx context.Context, hrID uuid.UUID, text string) (*VacancyWithDetails, error) {
	return s.vacancySvc.ParseVacancy(ctx, hrID, nil, text)
}

func (s *HRBotService) SearchCandidates(ctx context.Context, query string) ([]SearchResult, error) {
	return s.searchSvc.SearchUsers(ctx, query, 10)
}

func (s *HRBotService) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userSvc.GetUser(ctx, userID)
}

func (s *HRBotService) ListMyVacancies(ctx context.Context, hrID uuid.UUID) ([]VacancyWithDetails, error) {
	result, err := s.vacancySvc.ListVacancies(ctx, 1, 100)
	if err != nil {
		return nil, err
	}
	var filtered []VacancyWithDetails
	for _, v := range result.Vacancies {
		if v.Vacancy.HRID == hrID {
			filtered = append(filtered, v)
		}
	}
	return filtered, nil
}

func (s *HRBotService) ListAllHRs(ctx context.Context) ([]domain.CompanyHR, error) {
	var all []domain.CompanyHR
	var page int32 = 1
	const pageSize int32 = 100
	for {
		result, err := s.hrSvc.ListCompanyHRs(ctx, page, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, result.HRs...)
		if int64(len(all)) >= result.Total {
			break
		}
		page++
	}
	return all, nil
}

func (s *HRBotService) GetVacancyStats(ctx context.Context, vacancyID uuid.UUID) (total, unseen int64, err error) {
	total, err = s.vacancyAppSvc.CountByVacancy(ctx, vacancyID)
	if err != nil {
		return 0, 0, fmt.Errorf("count applications: %w", err)
	}
	unseen, err = s.vacancyAppSvc.CountUnseenByVacancy(ctx, vacancyID)
	if err != nil {
		return 0, 0, fmt.Errorf("count unseen applications: %w", err)
	}
	return total, unseen, nil
}

func (s *HRBotService) CountMatchingCandidates(ctx context.Context, vacancyTitle string, skills []string) int {
	query := vacancyTitle
	if len(skills) > 0 {
		query += " " + strings.Join(skills, " ")
	}
	results, err := s.searchSvc.SearchUsers(ctx, query, 100)
	if err != nil {
		return 0
	}
	// Count results with score above a threshold (rough match)
	count := 0
	for _, r := range results {
		if r.Score > 0.3 {
			count++
		}
	}
	return count
}

// -- Company data parsing (Gemini) --

func (s *HRBotService) ParseCompanyFromText(ctx context.Context, text string) (*gemini.ParsedCompanyFull, error) {
	return s.geminiClient.ParseCompanyFromText(ctx, text)
}

func (s *HRBotService) ParseCompanyFromFile(ctx context.Context, fileData []byte, mimeType string) (*gemini.ParsedCompanyFull, error) {
	return s.geminiClient.ParseCompanyFromFile(ctx, fileData, mimeType)
}

func (s *HRBotService) SaveCompanyData(ctx context.Context, hrID uuid.UUID, parsed *gemini.ParsedCompanyFull) (*domain.CompanyHR, error) {
	hr, err := s.hrSvc.GetCompanyHR(ctx, hrID)
	if err != nil {
		return nil, fmt.Errorf("get hr for company data: %w", err)
	}

	texts := make([]domain.CompanyDataText, 0, 3)
	for _, lang := range []string{"uz", "ru", "en"} {
		t := domain.CompanyDataText{
			Lang:     lang,
			IsSource: lang == parsed.SourceLang,
		}
		if f, ok := parsed.Fields["name"]; ok {
			t.Name = f[lang]
		}
		if f, ok := parsed.Fields["activity_type"]; ok {
			t.ActivityType = f[lang]
		}
		if f, ok := parsed.Fields["company_type"]; ok {
			t.CompanyType = f[lang]
		}
		if f, ok := parsed.Fields["about"]; ok {
			t.About = f[lang]
		}
		if f, ok := parsed.Fields["market"]; ok {
			t.Market = f[lang]
		}
		texts = append(texts, t)
	}

	hr.CompanyData = &domain.CompanyData{
		EmployeeCount:   parsed.EmployeeCount,
		Country:         parsed.Country,
		Address:         parsed.Address,
		Phone:           parsed.Phone,
		Telegram:        parsed.Telegram,
		TelegramChannel: parsed.TelegramChannel,
		Email:           parsed.Email,
		WebSite:         parsed.WebSite,
		Instagram:       parsed.Instagram,
		SourceLang:      parsed.SourceLang,
		Texts:           texts,
	}

	return s.hrSvc.UpdateCompanyHR(ctx, hr)
}

func (s *HRBotService) HasCompanyData(hr *domain.CompanyHR) bool {
	if hr.CompanyData == nil {
		return false
	}
	// Check if at least one text has a company name
	for _, t := range hr.CompanyData.Texts {
		if t.Name != "" {
			return true
		}
	}
	return false
}

// -- Vacancy draft parsing (Gemini) --

func (s *HRBotService) ParseVacancyFromText(ctx context.Context, text string) (*gemini.ParsedVacancyFull, error) {
	return s.geminiClient.ParseVacancyFromText(ctx, text)
}

func (s *HRBotService) ParseVacancyFromFile(ctx context.Context, fileData []byte, mimeType string) (*gemini.ParsedVacancyFull, error) {
	return s.geminiClient.ParseVacancyFromFile(ctx, fileData, mimeType)
}

func (s *HRBotService) MergeVacancy(ctx context.Context, existingJSON, additionalInfo string) (*gemini.ParsedVacancyFull, error) {
	return s.geminiClient.MergeVacancy(ctx, existingJSON, additionalInfo)
}

func (s *HRBotService) EnhanceVacancyDraft(ctx context.Context, draft *gemini.ParsedVacancyFull) (*gemini.ParsedVacancyFull, error) {
	draftJSON, err := json.Marshal(draft)
	if err != nil {
		return nil, fmt.Errorf("marshal draft for enhance: %w", err)
	}
	return s.geminiClient.EnhanceVacancyDescription(ctx, string(draftJSON))
}

func (s *HRBotService) CreateVacancyFromDraft(ctx context.Context, hrID uuid.UUID, companyData *domain.CompanyData, draft *gemini.ParsedVacancyFull) (*VacancyWithDetails, error) {
	return s.vacancySvc.CreateVacancyFromParsed(ctx, hrID, companyData, draft)
}

// -- Vacancy draft storage in Redis --

func (s *HRBotService) SaveVacancyDraft(ctx context.Context, telegramID int64, draft *gemini.ParsedVacancyFull) error {
	key := fmt.Sprintf("hr_bot:vacancy_draft:%d", telegramID)
	return s.stateSvc.redis.Set(ctx, key, draft, s.stateSvc.ttl)
}

func (s *HRBotService) GetVacancyDraft(ctx context.Context, telegramID int64) (*gemini.ParsedVacancyFull, error) {
	key := fmt.Sprintf("hr_bot:vacancy_draft:%d", telegramID)
	var draft gemini.ParsedVacancyFull
	found, err := s.stateSvc.redis.Get(ctx, key, &draft)
	if err != nil {
		return nil, fmt.Errorf("get vacancy draft: %w", err)
	}
	if !found {
		return nil, nil
	}
	return &draft, nil
}

func (s *HRBotService) ClearVacancyDraft(ctx context.Context, telegramID int64) error {
	key := fmt.Sprintf("hr_bot:vacancy_draft:%d", telegramID)
	return s.stateSvc.redis.Delete(ctx, key)
}

// -- State management with hr: prefix --

func (s *HRBotService) GetBotState(ctx context.Context, telegramID int64) (*domain.BotState, error) {
	tgID := "hr:" + strconv.FormatInt(telegramID, 10)
	return s.stateSvc.GetState(ctx, tgID)
}

func (s *HRBotService) SetState(ctx context.Context, telegramID int64, state string, data map[string]string) error {
	tgID := "hr:" + strconv.FormatInt(telegramID, 10)
	if data != nil {
		return s.stateSvc.SetStateWithData(ctx, tgID, state, data)
	}
	return s.stateSvc.SetState(ctx, tgID, state)
}

func (s *HRBotService) ClearState(ctx context.Context, telegramID int64) error {
	tgID := "hr:" + strconv.FormatInt(telegramID, 10)
	return s.stateSvc.ClearState(ctx, tgID)
}

// -- Helpers --

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
