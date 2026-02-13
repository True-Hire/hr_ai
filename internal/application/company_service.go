package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
)

type CompanyService struct {
	repo         domain.CompanyRepository
	textRepo     domain.CompanyTextRepository
	geminiClient *gemini.Client
}

func NewCompanyService(repo domain.CompanyRepository, textRepo domain.CompanyTextRepository, geminiClient *gemini.Client) *CompanyService {
	return &CompanyService{repo: repo, textRepo: textRepo, geminiClient: geminiClient}
}

type CreateCompanyInput struct {
	EmployeeCount   int32
	Country         string
	Address         string
	Phone           string
	Telegram        string
	TelegramChannel string
	Email           string
	LogoURL         string
	WebSite         string
	Instagram       string
	Name            string
	ActivityType    string
	CompanyType     string
	About           string
	Market          string
}

type CompanyWithTexts struct {
	Company *domain.Company
	Texts   []domain.CompanyText
}

func (s *CompanyService) CreateCompany(ctx context.Context, input *CreateCompanyInput) (*CompanyWithTexts, error) {
	// Build text input for Gemini
	textData := map[string]string{}
	if input.Name != "" {
		textData["name"] = input.Name
	}
	if input.ActivityType != "" {
		textData["activity_type"] = input.ActivityType
	}
	if input.CompanyType != "" {
		textData["company_type"] = input.CompanyType
	}
	if input.About != "" {
		textData["about"] = input.About
	}
	if input.Market != "" {
		textData["market"] = input.Market
	}

	jsonBytes, _ := json.Marshal(textData)

	parsed, err := s.geminiClient.TranslateCompany(ctx, string(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("gemini translate company: %w", err)
	}

	company := &domain.Company{
		ID:              uuid.New(),
		EmployeeCount:   input.EmployeeCount,
		Country:         input.Country,
		Address:         input.Address,
		Phone:           input.Phone,
		Telegram:        input.Telegram,
		TelegramChannel: input.TelegramChannel,
		Email:           input.Email,
		LogoURL:         input.LogoURL,
		WebSite:         input.WebSite,
		Instagram:       input.Instagram,
		SourceLang:      parsed.SourceLang,
	}

	created, err := s.repo.Create(ctx, company)
	if err != nil {
		return nil, fmt.Errorf("service create company: %w", err)
	}

	langs := []string{"uz", "ru", "en"}
	modelVer := s.geminiClient.ModelVersion()
	texts := make([]domain.CompanyText, 0, 3)

	for _, lang := range langs {
		ct := &domain.CompanyText{
			CompanyID:    created.ID,
			Lang:         lang,
			IsSource:     lang == parsed.SourceLang,
			ModelVersion: modelVer,
		}

		if translations, ok := parsed.Fields["name"]; ok {
			ct.Name = translations[lang]
		}
		if translations, ok := parsed.Fields["activity_type"]; ok {
			ct.ActivityType = translations[lang]
		}
		if translations, ok := parsed.Fields["company_type"]; ok {
			ct.CompanyType = translations[lang]
		}
		if translations, ok := parsed.Fields["about"]; ok {
			ct.About = translations[lang]
		}
		if translations, ok := parsed.Fields["market"]; ok {
			ct.Market = translations[lang]
		}

		savedText, err := s.textRepo.Create(ctx, ct)
		if err != nil {
			return nil, fmt.Errorf("create company text %s: %w", lang, err)
		}
		texts = append(texts, *savedText)
	}

	return &CompanyWithTexts{Company: created, Texts: texts}, nil
}

func (s *CompanyService) GetCompany(ctx context.Context, id uuid.UUID) (*CompanyWithTexts, error) {
	company, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	texts, err := s.textRepo.ListByCompany(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list company texts: %w", err)
	}

	return &CompanyWithTexts{Company: company, Texts: texts}, nil
}

type ListCompaniesResult struct {
	Companies []CompanyWithTexts
	Total     int64
}

func (s *CompanyService) ListCompanies(ctx context.Context, page, pageSize int32) (*ListCompaniesResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("service count companies: %w", err)
	}

	companies, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("service list companies: %w", err)
	}

	result := make([]CompanyWithTexts, 0, len(companies))
	for _, c := range companies {
		texts, err := s.textRepo.ListByCompany(ctx, c.ID)
		if err != nil {
			return nil, fmt.Errorf("list company texts for %s: %w", c.ID, err)
		}
		result = append(result, CompanyWithTexts{Company: &c, Texts: texts})
	}

	return &ListCompaniesResult{Companies: result, Total: total}, nil
}

func (s *CompanyService) UpdateCompany(ctx context.Context, company *domain.Company) (*CompanyWithTexts, error) {
	updated, err := s.repo.Update(ctx, company)
	if err != nil {
		return nil, err
	}

	texts, err := s.textRepo.ListByCompany(ctx, updated.ID)
	if err != nil {
		return nil, fmt.Errorf("list company texts: %w", err)
	}

	return &CompanyWithTexts{Company: updated, Texts: texts}, nil
}

func (s *CompanyService) DeleteCompany(ctx context.Context, id uuid.UUID) error {
	if err := s.textRepo.DeleteByCompany(ctx, id); err != nil {
		return fmt.Errorf("delete company texts: %w", err)
	}
	return s.repo.Delete(ctx, id)
}
