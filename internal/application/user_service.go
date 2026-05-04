package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	user.ID = uuid.New()
	if user.Status == "" {
		user.Status = "active"
	}
	if user.TariffType == "" {
		user.TariffType = "free"
	}
	created, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("service create user: %w", err)
	}
	return created, nil
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

type ListUsersResult struct {
	Users []domain.User
	Total int64
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int32) (*ListUsersResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("service count users: %w", err)
	}

	users, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("service list users: %w", err)
	}

	return &ListUsersResult{Users: users, Total: total}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return s.repo.Update(ctx, user)
}

func (s *UserService) GetByTelegramID(ctx context.Context, telegramID string) (*domain.User, error) {
	return s.repo.GetByTelegramID(ctx, telegramID)
}

func (s *UserService) UpdateLanguage(ctx context.Context, id uuid.UUID, language string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Language = language
	return s.repo.Update(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) CountMatchingUsers(ctx context.Context, mainCatID, subCatID uuid.UUID) (int64, error) {
	return s.repo.CountMatchingUsers(ctx, mainCatID, subCatID)
}

func (s *UserService) SetProfileScore(ctx context.Context, id uuid.UUID, score int32) error {
	return s.repo.SetProfileScore(ctx, id, score)
}

func (s *UserService) SetEstimatedSalary(ctx context.Context, id uuid.UUID, min, max int32, currency string) error {
	return s.repo.SetEstimatedSalary(ctx, id, min, max, currency)
}
