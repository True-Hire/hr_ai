package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type ExperienceItemService struct {
	repo domain.ExperienceItemRepository
}

func NewExperienceItemService(repo domain.ExperienceItemRepository) *ExperienceItemService {
	return &ExperienceItemService{repo: repo}
}

func (s *ExperienceItemService) CreateExperienceItem(ctx context.Context, item *domain.ExperienceItem) (*domain.ExperienceItem, error) {
	item.ID = uuid.New()
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("service create experience item: %w", err)
	}
	return created, nil
}

func (s *ExperienceItemService) GetExperienceItem(ctx context.Context, id uuid.UUID) (*domain.ExperienceItem, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ExperienceItemService) ListExperienceItemsByUser(ctx context.Context, userID uuid.UUID) ([]domain.ExperienceItem, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *ExperienceItemService) UpdateExperienceItem(ctx context.Context, item *domain.ExperienceItem) (*domain.ExperienceItem, error) {
	return s.repo.Update(ctx, item)
}

func (s *ExperienceItemService) DeleteExperienceItem(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *ExperienceItemService) DeleteExperienceItemsByUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteByUser(ctx, userID)
}
