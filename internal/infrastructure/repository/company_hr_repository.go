package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	companyhrsdb "github.com/ruziba3vich/hr-ai/db/sqlc/company_hrs"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CompanyHRRepository struct {
	q *companyhrsdb.Queries
}

func NewCompanyHRRepository(pool *pgxpool.Pool) *CompanyHRRepository {
	return &CompanyHRRepository{
		q: companyhrsdb.New(pool),
	}
}

func (r *CompanyHRRepository) Create(ctx context.Context, hr *domain.CompanyHR) (*domain.CompanyHR, error) {
	row, err := r.q.CreateCompanyHR(ctx, companyhrsdb.CreateCompanyHRParams{
		ID:            uuidToPgtype(hr.ID),
		FirstName:     hr.FirstName,
		LastName:      hr.LastName,
		Patronymic:    textToPgtype(hr.Patronymic),
		Phone:         textToPgtype(hr.Phone),
		Telegram:      textToPgtype(hr.Telegram),
		TelegramID:    textToPgtype(hr.TelegramID),
		Email:         textToPgtype(hr.Email),
		Position:      textToPgtype(hr.Position),
		Status:        hr.Status,
		CompanyName:   textToPgtype(hr.CompanyName),
		ActivityType:  textToPgtype(hr.ActivityType),
		CompanyType:   textToPgtype(hr.CompanyType),
		EmployeeCount: int4ToPgtype(hr.EmployeeCount),
		Country:       textToPgtype(hr.Country),
		Market:        textToPgtype(hr.Market),
		WebSite:       textToPgtype(hr.WebSite),
		About:         textToPgtype(hr.About),
		LogoUrl:       textToPgtype(hr.LogoURL),
		Instagram:     textToPgtype(hr.Instagram),
	})
	if err != nil {
		return nil, fmt.Errorf("create company hr: %w", err)
	}
	return companyHRToDomain(row), nil
}

func (r *CompanyHRRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CompanyHR, error) {
	row, err := r.q.GetCompanyHRByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyHRNotFound
		}
		return nil, fmt.Errorf("get company hr by id: %w", err)
	}
	return companyHRToDomain(row), nil
}

func (r *CompanyHRRepository) List(ctx context.Context, limit, offset int32) ([]domain.CompanyHR, error) {
	rows, err := r.q.ListCompanyHRs(ctx, companyhrsdb.ListCompanyHRsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list company hrs: %w", err)
	}
	hrs := make([]domain.CompanyHR, 0, len(rows))
	for _, row := range rows {
		hrs = append(hrs, *companyHRToDomain(row))
	}
	return hrs, nil
}

func (r *CompanyHRRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.q.CountCompanyHRs(ctx)
	if err != nil {
		return 0, fmt.Errorf("count company hrs: %w", err)
	}
	return count, nil
}

func (r *CompanyHRRepository) Update(ctx context.Context, hr *domain.CompanyHR) (*domain.CompanyHR, error) {
	row, err := r.q.UpdateCompanyHR(ctx, companyhrsdb.UpdateCompanyHRParams{
		ID:            uuidToPgtype(hr.ID),
		FirstName:     hr.FirstName,
		LastName:      hr.LastName,
		Patronymic:    hr.Patronymic,
		Phone:         hr.Phone,
		Telegram:      hr.Telegram,
		TelegramID:    hr.TelegramID,
		Email:         hr.Email,
		Position:      hr.Position,
		Status:        hr.Status,
		CompanyName:   hr.CompanyName,
		ActivityType:  hr.ActivityType,
		CompanyType:   hr.CompanyType,
		EmployeeCount: hr.EmployeeCount,
		Country:       hr.Country,
		Market:        hr.Market,
		WebSite:       hr.WebSite,
		About:         hr.About,
		LogoUrl:       hr.LogoURL,
		Instagram:     hr.Instagram,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyHRNotFound
		}
		return nil, fmt.Errorf("update company hr: %w", err)
	}
	return companyHRToDomain(row), nil
}

func (r *CompanyHRRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteCompanyHR(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete company hr: %w", err)
	}
	return nil
}

func (r *CompanyHRRepository) GetByPhone(ctx context.Context, phone string) (*domain.CompanyHR, error) {
	row, err := r.q.GetCompanyHRByPhone(ctx, textToPgtype(phone))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyHRNotFound
		}
		return nil, fmt.Errorf("get company hr by phone: %w", err)
	}
	return companyHRToDomain(row), nil
}

func (r *CompanyHRRepository) GetByEmail(ctx context.Context, email string) (*domain.CompanyHR, error) {
	row, err := r.q.GetCompanyHRByEmail(ctx, textToPgtype(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyHRNotFound
		}
		return nil, fmt.Errorf("get company hr by email: %w", err)
	}
	return companyHRToDomain(row), nil
}

func (r *CompanyHRRepository) SetPassword(ctx context.Context, id uuid.UUID, hash string) error {
	err := r.q.SetCompanyHRPassword(ctx, companyhrsdb.SetCompanyHRPasswordParams{
		ID:           uuidToPgtype(id),
		PasswordHash: textToPgtype(hash),
	})
	if err != nil {
		return fmt.Errorf("set company hr password: %w", err)
	}
	return nil
}

func companyHRToDomain(row companyhrsdb.CompanyHr) *domain.CompanyHR {
	return &domain.CompanyHR{
		ID:            pgtypeToUUID(row.ID),
		FirstName:     row.FirstName,
		LastName:      row.LastName,
		Patronymic:    pgtypeToString(row.Patronymic),
		Phone:         pgtypeToString(row.Phone),
		Telegram:      pgtypeToString(row.Telegram),
		TelegramID:    pgtypeToString(row.TelegramID),
		Email:         pgtypeToString(row.Email),
		Position:      pgtypeToString(row.Position),
		Status:        row.Status,
		PasswordHash:  pgtypeToString(row.PasswordHash),
		CompanyName:   pgtypeToString(row.CompanyName),
		ActivityType:  pgtypeToString(row.ActivityType),
		CompanyType:   pgtypeToString(row.CompanyType),
		EmployeeCount: pgtypeToInt32(row.EmployeeCount),
		Country:       pgtypeToString(row.Country),
		Market:        pgtypeToString(row.Market),
		WebSite:       pgtypeToString(row.WebSite),
		About:         pgtypeToString(row.About),
		LogoURL:       pgtypeToString(row.LogoUrl),
		Instagram:     pgtypeToString(row.Instagram),
		CreatedAt:     row.CreatedAt.Time,
	}
}
