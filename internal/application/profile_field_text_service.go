package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
)

type ProfileFieldTextService struct {
	repo         domain.ProfileFieldTextRepository
	geminiClient *gemini.Client
}

func NewProfileFieldTextService(repo domain.ProfileFieldTextRepository, geminiClient *gemini.Client) *ProfileFieldTextService {
	return &ProfileFieldTextService{repo: repo, geminiClient: geminiClient}
}

func (s *ProfileFieldTextService) CreateProfileFieldText(ctx context.Context, text *domain.ProfileFieldText) (*domain.ProfileFieldText, error) {
	return s.repo.Create(ctx, text)
}

func (s *ProfileFieldTextService) GetProfileFieldText(ctx context.Context, profileFieldID uuid.UUID, lang string) (*domain.ProfileFieldText, error) {
	return s.repo.Get(ctx, profileFieldID, lang)
}

func (s *ProfileFieldTextService) ListProfileFieldTexts(ctx context.Context, profileFieldID uuid.UUID) ([]domain.ProfileFieldText, error) {
	return s.repo.ListByField(ctx, profileFieldID)
}

func (s *ProfileFieldTextService) UpdateProfileFieldText(ctx context.Context, profileFieldID uuid.UUID, content string) ([]domain.ProfileFieldText, error) {
	translated, err := s.geminiClient.TranslateText(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("gemini translate text: %w", err)
	}

	if err := s.repo.DeleteByField(ctx, profileFieldID); err != nil {
		return nil, fmt.Errorf("delete old profile field texts: %w", err)
	}

	langs := []string{"uz", "ru", "en"}
	modelVer := s.geminiClient.ModelVersion()
	texts := make([]domain.ProfileFieldText, 0, 3)

	for _, lang := range langs {
		translatedContent, ok := translated.Translations[lang]
		if !ok || translatedContent == "" {
			continue
		}

		saved, err := s.repo.Create(ctx, &domain.ProfileFieldText{
			ProfileFieldID: profileFieldID,
			Lang:           lang,
			Content:        translatedContent,
			IsSource:       lang == translated.SourceLang,
			ModelVersion:   modelVer,
		})
		if err != nil {
			return nil, fmt.Errorf("create profile field text %s: %w", lang, err)
		}
		texts = append(texts, *saved)
	}

	return texts, nil
}

func (s *ProfileFieldTextService) DeleteProfileFieldText(ctx context.Context, profileFieldID uuid.UUID, lang string) error {
	return s.repo.Delete(ctx, profileFieldID, lang)
}

func (s *ProfileFieldTextService) DeleteProfileFieldTextsByField(ctx context.Context, profileFieldID uuid.UUID) error {
	return s.repo.DeleteByField(ctx, profileFieldID)
}
