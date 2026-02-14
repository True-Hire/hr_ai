package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
)

type CountryService struct {
	repo         domain.CountryRepository
	textRepo     domain.CountryTextRepository
	geminiClient *gemini.Client
}

func NewCountryService(repo domain.CountryRepository, textRepo domain.CountryTextRepository, geminiClient *gemini.Client) *CountryService {
	return &CountryService{repo: repo, textRepo: textRepo, geminiClient: geminiClient}
}

type CountryWithTexts struct {
	Country *domain.Country
	Texts   []domain.CountryText
}

func (s *CountryService) GetCountry(ctx context.Context, id uuid.UUID) (*CountryWithTexts, error) {
	country, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	texts, err := s.textRepo.ListByCountry(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list country texts: %w", err)
	}

	return &CountryWithTexts{Country: country, Texts: texts}, nil
}

func (s *CountryService) ListCountries(ctx context.Context) ([]CountryWithTexts, error) {
	countries, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list countries: %w", err)
	}

	result := make([]CountryWithTexts, 0, len(countries))
	for _, c := range countries {
		texts, err := s.textRepo.ListByCountry(ctx, c.ID)
		if err != nil {
			return nil, fmt.Errorf("list texts for country %s: %w", c.ID, err)
		}
		result = append(result, CountryWithTexts{Country: &c, Texts: texts})
	}

	return result, nil
}

func (s *CountryService) CreateCountry(ctx context.Context, name, shortCode string) (*CountryWithTexts, error) {
	translated, err := s.geminiClient.TranslateText(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("gemini translate country name: %w", err)
	}

	enName := name
	if en, ok := translated.Translations["en"]; ok && en != "" {
		enName = en
	}

	country, err := s.repo.Create(ctx, &domain.Country{
		ID:        uuid.New(),
		Name:      enName,
		ShortCode: shortCode,
	})
	if err != nil {
		return nil, fmt.Errorf("create country: %w", err)
	}

	langs := []string{"uz", "ru", "en"}
	modelVer := s.geminiClient.ModelVersion()
	texts := make([]domain.CountryText, 0, 3)

	for _, lang := range langs {
		translatedName, ok := translated.Translations[lang]
		if !ok || translatedName == "" {
			continue
		}

		saved, err := s.textRepo.Create(ctx, &domain.CountryText{
			CountryID:    country.ID,
			Lang:         lang,
			Name:         translatedName,
			IsSource:     lang == translated.SourceLang,
			ModelVersion: modelVer,
		})
		if err != nil {
			return nil, fmt.Errorf("create country text %s: %w", lang, err)
		}
		texts = append(texts, *saved)
	}

	return &CountryWithTexts{Country: country, Texts: texts}, nil
}

func (s *CountryService) DeleteCountry(ctx context.Context, id uuid.UUID) error {
	if err := s.textRepo.DeleteByCountry(ctx, id); err != nil {
		return fmt.Errorf("delete country texts: %w", err)
	}
	return s.repo.Delete(ctx, id)
}
