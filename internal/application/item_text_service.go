package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type ItemTextService struct {
	repo domain.ItemTextRepository
}

func NewItemTextService(repo domain.ItemTextRepository) *ItemTextService {
	return &ItemTextService{repo: repo}
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

func (s *ItemTextService) UpdateItemText(ctx context.Context, text *domain.ItemText) (*domain.ItemText, error) {
	return s.repo.Update(ctx, text)
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
