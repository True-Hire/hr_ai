package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	companytextsdb "github.com/ruziba3vich/hr-ai/db/sqlc/company_texts"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CompanyTextRepository struct {
	q *companytextsdb.Queries
}

func NewCompanyTextRepository(pool *pgxpool.Pool) *CompanyTextRepository {
	return &CompanyTextRepository{
		q: companytextsdb.New(pool),
	}
}

func (r *CompanyTextRepository) Create(ctx context.Context, ct *domain.CompanyText) (*domain.CompanyText, error) {
	row, err := r.q.CreateCompanyText(ctx, companytextsdb.CreateCompanyTextParams{
		CompanyID:    uuidToPgtype(ct.CompanyID),
		Lang:         ct.Lang,
		Name:         ct.Name,
		ActivityType: ct.ActivityType,
		CompanyType:  ct.CompanyType,
		About:        ct.About,
		Market:       ct.Market,
		IsSource:     ct.IsSource,
		ModelVersion: textToPgtype(ct.ModelVersion),
	})
	if err != nil {
		return nil, fmt.Errorf("create company text: %w", err)
	}
	return companyTextToDomain(row), nil
}

func (r *CompanyTextRepository) Get(ctx context.Context, companyID uuid.UUID, lang string) (*domain.CompanyText, error) {
	row, err := r.q.GetCompanyText(ctx, companytextsdb.GetCompanyTextParams{
		CompanyID: uuidToPgtype(companyID),
		Lang:      lang,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyNotFound
		}
		return nil, fmt.Errorf("get company text: %w", err)
	}
	return companyTextToDomain(row), nil
}

func (r *CompanyTextRepository) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.CompanyText, error) {
	rows, err := r.q.ListCompanyTextsByCompany(ctx, uuidToPgtype(companyID))
	if err != nil {
		return nil, fmt.Errorf("list company texts: %w", err)
	}
	texts := make([]domain.CompanyText, 0, len(rows))
	for _, row := range rows {
		texts = append(texts, *companyTextToDomain(row))
	}
	return texts, nil
}

func (r *CompanyTextRepository) Update(ctx context.Context, ct *domain.CompanyText) (*domain.CompanyText, error) {
	row, err := r.q.UpdateCompanyText(ctx, companytextsdb.UpdateCompanyTextParams{
		CompanyID:    uuidToPgtype(ct.CompanyID),
		Lang:         ct.Lang,
		Name:         ct.Name,
		ActivityType: ct.ActivityType,
		CompanyType:  ct.CompanyType,
		About:        ct.About,
		Market:       ct.Market,
		IsSource:     ct.IsSource,
		ModelVersion: textToPgtype(ct.ModelVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyNotFound
		}
		return nil, fmt.Errorf("update company text: %w", err)
	}
	return companyTextToDomain(row), nil
}

func (r *CompanyTextRepository) Delete(ctx context.Context, companyID uuid.UUID, lang string) error {
	return r.q.DeleteCompanyText(ctx, companytextsdb.DeleteCompanyTextParams{
		CompanyID: uuidToPgtype(companyID),
		Lang:      lang,
	})
}

func (r *CompanyTextRepository) DeleteByCompany(ctx context.Context, companyID uuid.UUID) error {
	return r.q.DeleteCompanyTextsByCompany(ctx, uuidToPgtype(companyID))
}

func companyTextToDomain(row companytextsdb.CompanyText) *domain.CompanyText {
	return &domain.CompanyText{
		CompanyID:    pgtypeToUUID(row.CompanyID),
		Lang:         row.Lang,
		Name:         row.Name,
		ActivityType: row.ActivityType,
		CompanyType:  row.CompanyType,
		About:        row.About,
		Market:       row.Market,
		IsSource:     row.IsSource,
		ModelVersion: pgtypeToString(row.ModelVersion),
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
