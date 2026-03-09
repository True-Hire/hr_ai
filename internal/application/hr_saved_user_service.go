package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/repository"
)

type HRSavedUserService struct {
	repo *repository.HRSavedUserRepository
}

func NewHRSavedUserService(repo *repository.HRSavedUserRepository) *HRSavedUserService {
	return &HRSavedUserService{repo: repo}
}

func (s *HRSavedUserService) Save(ctx context.Context, hrID, userID uuid.UUID, note string) (*domain.HRSavedUser, error) {
	return s.repo.Save(ctx, hrID, userID, note)
}

func (s *HRSavedUserService) Unsave(ctx context.Context, hrID, userID uuid.UUID) error {
	return s.repo.Unsave(ctx, hrID, userID)
}

func (s *HRSavedUserService) IsSaved(ctx context.Context, hrID, userID uuid.UUID) (bool, error) {
	return s.repo.IsSaved(ctx, hrID, userID)
}

type SavedUserListResult struct {
	SavedUsers []domain.HRSavedUser
	Total      int64
}

func (s *HRSavedUserService) ListFiltered(ctx context.Context, hrID uuid.UUID, nameQuery string, skills []string, limit, offset int32) (*SavedUserListResult, error) {
	users, total, err := s.repo.ListByHRFiltered(ctx, hrID, nameQuery, skills, limit, offset)
	if err != nil {
		return nil, err
	}
	return &SavedUserListResult{SavedUsers: users, Total: total}, nil
}
