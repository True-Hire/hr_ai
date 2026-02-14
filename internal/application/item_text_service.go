package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
)

type ItemTextService struct {
	repo         domain.ItemTextRepository
	geminiClient *gemini.Client
}

func NewItemTextService(repo domain.ItemTextRepository, geminiClient *gemini.Client) *ItemTextService {
	return &ItemTextService{repo: repo, geminiClient: geminiClient}
}

func (s *ItemTextService) CreateItemText(ctx context.Context, text *domain.ItemText) (*domain.ItemText, error) {
	return s.repo.Create(ctx, text)
}

func (s *ItemTextService) GetItemText(ctx context.Context, itemID uuid.UUID, itemType string, lang string) (*domain.ItemText, error) {
	return s.repo.Get(ctx, itemID, itemType, lang)
}

func (s *ItemTextService) ListItemTextsByItem(ctx context.Context, itemID uuid.UUID, itemType string) ([]domain.ItemText, error) {
	return s.repo.ListByItem(ctx, itemID, itemType)
}

func (s *ItemTextService) UpdateItemText(ctx context.Context, itemID uuid.UUID, itemType string, description string) ([]domain.ItemText, error) {
	translated, err := s.geminiClient.TranslateText(ctx, description)
	if err != nil {
		return nil, fmt.Errorf("gemini translate text: %w", err)
	}

	if err := s.repo.DeleteByItem(ctx, itemID, itemType); err != nil {
		return nil, fmt.Errorf("delete old item texts: %w", err)
	}

	langs := []string{"uz", "ru", "en"}
	modelVer := s.geminiClient.ModelVersion()
	texts := make([]domain.ItemText, 0, 3)

	for _, lang := range langs {
		translatedDesc, ok := translated.Translations[lang]
		if !ok || translatedDesc == "" {
			continue
		}

		saved, err := s.repo.Create(ctx, &domain.ItemText{
			ItemID:       itemID,
			ItemType:     itemType,
			Lang:         lang,
			Description:  translatedDesc,
			IsSource:     lang == translated.SourceLang,
			ModelVersion: modelVer,
		})
		if err != nil {
			return nil, fmt.Errorf("create item text %s: %w", lang, err)
		}
		texts = append(texts, *saved)
	}

	return texts, nil
}

func (s *ItemTextService) DeleteItemText(ctx context.Context, itemID uuid.UUID, itemType string, lang string) error {
	return s.repo.Delete(ctx, itemID, itemType, lang)
}

func (s *ItemTextService) DeleteItemTextsByItem(ctx context.Context, itemID uuid.UUID, itemType string) error {
	return s.repo.DeleteByItem(ctx, itemID, itemType)
}

func (s *ItemTextService) DeleteItemTextsByItemID(ctx context.Context, itemID uuid.UUID) error {
	return s.repo.DeleteByItemID(ctx, itemID)
}
