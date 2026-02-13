package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CompanyTextService struct {
	repo domain.CompanyTextRepository
}

func NewCompanyTextService(repo domain.CompanyTextRepository) *CompanyTextService {
	return &CompanyTextService{repo: repo}
}

func (s *CompanyTextService) GetCompanyText(ctx context.Context, companyID uuid.UUID, lang string) (*domain.CompanyText, error) {
	return s.repo.Get(ctx, companyID, lang)
}

func (s *CompanyTextService) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.CompanyText, error) {
	return s.repo.ListByCompany(ctx, companyID)
}

func (s *CompanyTextService) UpdateCompanyText(ctx context.Context, ct *domain.CompanyText) (*domain.CompanyText, error) {
	return s.repo.Update(ctx, ct)
}
