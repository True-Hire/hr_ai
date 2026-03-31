package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type VacancyApplicationService struct {
	repo domain.VacancyApplicationRepository
}

func NewVacancyApplicationService(repo domain.VacancyApplicationRepository) *VacancyApplicationService {
	return &VacancyApplicationService{repo: repo}
}

func (s *VacancyApplicationService) Apply(ctx context.Context, userID, vacancyID uuid.UUID, coverLetter string) (*domain.VacancyApplication, error) {
	existing, err := s.repo.GetByUserAndVacancy(ctx, userID, vacancyID)
	if err == nil && existing != nil {
		return nil, domain.ErrAlreadyApplied
	}
	if err != nil && !errors.Is(err, domain.ErrVacancyApplicationNotFound) {
		return nil, fmt.Errorf("check existing application: %w", err)
	}

	va := &domain.VacancyApplication{
		ID:          uuid.New(),
		UserID:      userID,
		VacancyID:   vacancyID,
		Status:      "pending",
		CoverLetter: coverLetter,
	}

	created, err := s.repo.Create(ctx, va)
	if err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	return created, nil
}

func (s *VacancyApplicationService) GetByUserAndVacancy(ctx context.Context, userID, vacancyID uuid.UUID) (*domain.VacancyApplication, error) {
	return s.repo.GetByUserAndVacancy(ctx, userID, vacancyID)
}

func (s *VacancyApplicationService) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]domain.VacancyApplication, error) {
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *VacancyApplicationService) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.CountByUser(ctx, userID)
}

func (s *VacancyApplicationService) ListByVacancy(ctx context.Context, vacancyID uuid.UUID, limit, offset int32) ([]domain.VacancyApplication, error) {
	return s.repo.ListByVacancy(ctx, vacancyID, limit, offset)
}

func (s *VacancyApplicationService) CountByVacancy(ctx context.Context, vacancyID uuid.UUID) (int64, error) {
	return s.repo.CountByVacancy(ctx, vacancyID)
}

func (s *VacancyApplicationService) CountUnseenByVacancy(ctx context.Context, vacancyID uuid.UUID) (int64, error) {
	return s.repo.CountUnseenByVacancy(ctx, vacancyID)
}

func (s *VacancyApplicationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.VacancyApplication, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *VacancyApplicationService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*domain.VacancyApplication, error) {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *VacancyApplicationService) MarkSeen(ctx context.Context, id uuid.UUID) (*domain.VacancyApplication, error) {
	return s.repo.MarkSeen(ctx, id)
}
