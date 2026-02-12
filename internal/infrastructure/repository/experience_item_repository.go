package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	eidb "github.com/ruziba3vich/hr-ai/db/sqlc/experience_items"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type ExperienceItemRepository struct {
	q *eidb.Queries
}

func NewExperienceItemRepository(pool *pgxpool.Pool) *ExperienceItemRepository {
	return &ExperienceItemRepository{
		q: eidb.New(pool),
	}
}

func (r *ExperienceItemRepository) Create(ctx context.Context, item *domain.ExperienceItem) (*domain.ExperienceItem, error) {
	row, err := r.q.CreateExperienceItem(ctx, eidb.CreateExperienceItemParams{
		ID:        uuidToPgtype(item.ID),
		UserID:    uuidToPgtype(item.UserID),
		Company:   item.Company,
		Position:  item.Position,
		StartDate: textToPgtype(item.StartDate),
		EndDate:   textToPgtype(item.EndDate),
		Projects:  textToPgtype(item.Projects),
		WebSite:   textToPgtype(item.WebSite),
		ItemOrder: item.ItemOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("create experience item: %w", err)
	}
	return experienceItemToDomain(row), nil
}

func (r *ExperienceItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ExperienceItem, error) {
	row, err := r.q.GetExperienceItemByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrExperienceItemNotFound
		}
		return nil, fmt.Errorf("get experience item by id: %w", err)
	}
	return experienceItemToDomain(row), nil
}

func (r *ExperienceItemRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.ExperienceItem, error) {
	rows, err := r.q.ListExperienceItemsByUser(ctx, uuidToPgtype(userID))
	if err != nil {
		return nil, fmt.Errorf("list experience items by user: %w", err)
	}
	items := make([]domain.ExperienceItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, *experienceItemToDomain(row))
	}
	return items, nil
}

func (r *ExperienceItemRepository) Update(ctx context.Context, item *domain.ExperienceItem) (*domain.ExperienceItem, error) {
	row, err := r.q.UpdateExperienceItem(ctx, eidb.UpdateExperienceItemParams{
		Company:   item.Company,
		Position:  item.Position,
		StartDate: item.StartDate,
		EndDate:   item.EndDate,
		Projects:  item.Projects,
		WebSite:   item.WebSite,
		ItemOrder: item.ItemOrder,
		ID:        uuidToPgtype(item.ID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrExperienceItemNotFound
		}
		return nil, fmt.Errorf("update experience item: %w", err)
	}
	return experienceItemToDomain(row), nil
}

func (r *ExperienceItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteExperienceItem(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete experience item: %w", err)
	}
	return nil
}

func (r *ExperienceItemRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	err := r.q.DeleteExperienceItemsByUser(ctx, uuidToPgtype(userID))
	if err != nil {
		return fmt.Errorf("delete experience items by user: %w", err)
	}
	return nil
}

func experienceItemToDomain(row eidb.ExperienceItem) *domain.ExperienceItem {
	return &domain.ExperienceItem{
		ID:        pgtypeToUUID(row.ID),
		UserID:    pgtypeToUUID(row.UserID),
		Company:   row.Company,
		Position:  row.Position,
		StartDate: pgtypeToString(row.StartDate),
		EndDate:   pgtypeToString(row.EndDate),
		Projects:  pgtypeToString(row.Projects),
		WebSite:   pgtypeToString(row.WebSite),
		ItemOrder: row.ItemOrder,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
