package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type EducationItemService struct {
	repo domain.EducationItemRepository
}

func NewEducationItemService(repo domain.EducationItemRepository) *EducationItemService {
	return &EducationItemService{repo: repo}
}

func (s *EducationItemService) CreateEducationItem(ctx context.Context, item *domain.EducationItem) (*domain.EducationItem, error) {
	item.ID = uuid.New()
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("service create education item: %w", err)
	}
	return created, nil
}

func (s *EducationItemService) GetEducationItem(ctx context.Context, id uuid.UUID) (*domain.EducationItem, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EducationItemService) ListEducationItemsByUser(ctx context.Context, userID uuid.UUID) ([]domain.EducationItem, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *EducationItemService) UpdateEducationItem(ctx context.Context, item *domain.EducationItem) (*domain.EducationItem, error) {
	return s.repo.Update(ctx, item)
}

func (s *EducationItemService) DeleteEducationItem(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *EducationItemService) DeleteEducationItemsByUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteByUser(ctx, userID)
}
