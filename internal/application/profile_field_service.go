package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type ProfileFieldService struct {
	repo domain.ProfileFieldRepository
}

func NewProfileFieldService(repo domain.ProfileFieldRepository) *ProfileFieldService {
	return &ProfileFieldService{repo: repo}
}

func (s *ProfileFieldService) CreateProfileField(ctx context.Context, field *domain.ProfileField) (*domain.ProfileField, error) {
	field.ID = uuid.New()
	created, err := s.repo.Create(ctx, field)
	if err != nil {
		return nil, fmt.Errorf("service create profile field: %w", err)
	}
	return created, nil
}

func (s *ProfileFieldService) GetProfileField(ctx context.Context, id uuid.UUID) (*domain.ProfileField, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProfileFieldService) GetProfileFieldByUserAndName(ctx context.Context, userID uuid.UUID, fieldName string) (*domain.ProfileField, error) {
	return s.repo.GetByUserAndName(ctx, userID, fieldName)
}

func (s *ProfileFieldService) ListProfileFieldsByUser(ctx context.Context, userID uuid.UUID) ([]domain.ProfileField, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *ProfileFieldService) UpdateProfileField(ctx context.Context, field *domain.ProfileField) (*domain.ProfileField, error) {
	return s.repo.Update(ctx, field)
}

func (s *ProfileFieldService) DeleteProfileField(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProfileFieldService) DeleteProfileFieldsByUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteByUser(ctx, userID)
}
