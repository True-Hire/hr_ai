package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CompanyHRService struct {
	repo domain.CompanyHRRepository
}

func NewCompanyHRService(repo domain.CompanyHRRepository) *CompanyHRService {
	return &CompanyHRService{repo: repo}
}

func (s *CompanyHRService) CreateCompanyHR(ctx context.Context, hr *domain.CompanyHR) (*domain.CompanyHR, error) {
	hr.ID = uuid.New()
	if hr.Status == "" {
		hr.Status = "active"
	}
	created, err := s.repo.Create(ctx, hr)
	if err != nil {
		return nil, fmt.Errorf("service create company hr: %w", err)
	}
	return created, nil
}

func (s *CompanyHRService) GetCompanyHR(ctx context.Context, id uuid.UUID) (*domain.CompanyHR, error) {
	return s.repo.GetByID(ctx, id)
}

type ListCompanyHRsResult struct {
	HRs   []domain.CompanyHR
	Total int64
}

func (s *CompanyHRService) ListCompanyHRs(ctx context.Context, page, pageSize int32) (*ListCompanyHRsResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("service count company hrs: %w", err)
	}

	hrs, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("service list company hrs: %w", err)
	}

	return &ListCompanyHRsResult{HRs: hrs, Total: total}, nil
}

func (s *CompanyHRService) UpdateCompanyHR(ctx context.Context, hr *domain.CompanyHR) (*domain.CompanyHR, error) {
	return s.repo.Update(ctx, hr)
}

func (s *CompanyHRService) DeleteCompanyHR(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
