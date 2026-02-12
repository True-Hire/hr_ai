package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	eddb "github.com/ruziba3vich/hr-ai/db/sqlc/education_items"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type EducationItemRepository struct {
	q *eddb.Queries
}

func NewEducationItemRepository(pool *pgxpool.Pool) *EducationItemRepository {
	return &EducationItemRepository{
		q: eddb.New(pool),
	}
}

func (r *EducationItemRepository) Create(ctx context.Context, item *domain.EducationItem) (*domain.EducationItem, error) {
	row, err := r.q.CreateEducationItem(ctx, eddb.CreateEducationItemParams{
		ID:           uuidToPgtype(item.ID),
		UserID:       uuidToPgtype(item.UserID),
		Institution:  item.Institution,
		Degree:       item.Degree,
		FieldOfStudy: textToPgtype(item.FieldOfStudy),
		StartDate:    textToPgtype(item.StartDate),
		EndDate:      textToPgtype(item.EndDate),
		Location:     textToPgtype(item.Location),
		ItemOrder:    item.ItemOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("create education item: %w", err)
	}
	return educationItemToDomain(row), nil
}

func (r *EducationItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.EducationItem, error) {
	row, err := r.q.GetEducationItemByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEducationItemNotFound
		}
		return nil, fmt.Errorf("get education item by id: %w", err)
	}
	return educationItemToDomain(row), nil
}

func (r *EducationItemRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.EducationItem, error) {
	rows, err := r.q.ListEducationItemsByUser(ctx, uuidToPgtype(userID))
	if err != nil {
		return nil, fmt.Errorf("list education items by user: %w", err)
	}
	items := make([]domain.EducationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, *educationItemToDomain(row))
	}
	return items, nil
}

func (r *EducationItemRepository) Update(ctx context.Context, item *domain.EducationItem) (*domain.EducationItem, error) {
	row, err := r.q.UpdateEducationItem(ctx, eddb.UpdateEducationItemParams{
		Institution:  item.Institution,
		Degree:       item.Degree,
		FieldOfStudy: item.FieldOfStudy,
		StartDate:    item.StartDate,
		EndDate:      item.EndDate,
		Location:     item.Location,
		ItemOrder:    item.ItemOrder,
		ID:           uuidToPgtype(item.ID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEducationItemNotFound
		}
		return nil, fmt.Errorf("update education item: %w", err)
	}
	return educationItemToDomain(row), nil
}

func (r *EducationItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteEducationItem(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete education item: %w", err)
	}
	return nil
}

func (r *EducationItemRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	err := r.q.DeleteEducationItemsByUser(ctx, uuidToPgtype(userID))
	if err != nil {
		return fmt.Errorf("delete education items by user: %w", err)
	}
	return nil
}

func educationItemToDomain(row eddb.EducationItem) *domain.EducationItem {
	return &domain.EducationItem{
		ID:           pgtypeToUUID(row.ID),
		UserID:       pgtypeToUUID(row.UserID),
		Institution:  row.Institution,
		Degree:       row.Degree,
		FieldOfStudy: pgtypeToString(row.FieldOfStudy),
		StartDate:    pgtypeToString(row.StartDate),
		EndDate:      pgtypeToString(row.EndDate),
		Location:     pgtypeToString(row.Location),
		ItemOrder:    row.ItemOrder,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
