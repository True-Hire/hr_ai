package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type VacancyTextService struct {
	repo domain.VacancyTextRepository
}

func NewVacancyTextService(repo domain.VacancyTextRepository) *VacancyTextService {
	return &VacancyTextService{repo: repo}
}

func (s *VacancyTextService) GetVacancyText(ctx context.Context, vacancyID uuid.UUID, lang string) (*domain.VacancyText, error) {
	return s.repo.Get(ctx, vacancyID, lang)
}

func (s *VacancyTextService) ListByVacancy(ctx context.Context, vacancyID uuid.UUID) ([]domain.VacancyText, error) {
	return s.repo.ListByVacancy(ctx, vacancyID)
}

func (s *VacancyTextService) UpdateVacancyText(ctx context.Context, vt *domain.VacancyText) (*domain.VacancyText, error) {
	return s.repo.Update(ctx, vt)
}
