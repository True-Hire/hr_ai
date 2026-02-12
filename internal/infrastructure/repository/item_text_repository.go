package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	itdb "github.com/ruziba3vich/hr-ai/db/sqlc/item_texts"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type ItemTextRepository struct {
	q *itdb.Queries
}

func NewItemTextRepository(pool *pgxpool.Pool) *ItemTextRepository {
	return &ItemTextRepository{
		q: itdb.New(pool),
	}
}

func (r *ItemTextRepository) Create(ctx context.Context, text *domain.ItemText) (*domain.ItemText, error) {
	row, err := r.q.CreateItemText(ctx, itdb.CreateItemTextParams{
		ItemID:       uuidToPgtype(text.ItemID),
		ItemType:     text.ItemType,
		Lang:         text.Lang,
		Description:  text.Description,
		IsSource:     text.IsSource,
		ModelVersion: textToPgtype(text.ModelVersion),
	})
	if err != nil {
		return nil, fmt.Errorf("create item text: %w", err)
	}
	return itemTextToDomain(row), nil
}

func (r *ItemTextRepository) Get(ctx context.Context, itemID uuid.UUID, itemType string, lang string) (*domain.ItemText, error) {
	row, err := r.q.GetItemText(ctx, itdb.GetItemTextParams{
		ItemID:   uuidToPgtype(itemID),
		ItemType: itemType,
		Lang:     lang,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrItemTextNotFound
		}
		return nil, fmt.Errorf("get item text: %w", err)
	}
	return itemTextToDomain(row), nil
}

func (r *ItemTextRepository) ListByItem(ctx context.Context, itemID uuid.UUID, itemType string) ([]domain.ItemText, error) {
	rows, err := r.q.ListItemTextsByItem(ctx, itdb.ListItemTextsByItemParams{
		ItemID:   uuidToPgtype(itemID),
		ItemType: itemType,
	})
	if err != nil {
		return nil, fmt.Errorf("list item texts by item: %w", err)
	}
	texts := make([]domain.ItemText, 0, len(rows))
	for _, row := range rows {
		texts = append(texts, *itemTextToDomain(row))
	}
	return texts, nil
}

func (r *ItemTextRepository) Update(ctx context.Context, text *domain.ItemText) (*domain.ItemText, error) {
	row, err := r.q.UpdateItemText(ctx, itdb.UpdateItemTextParams{
		Description:  text.Description,
		ModelVersion: textToPgtype(text.ModelVersion),
		ItemID:       uuidToPgtype(text.ItemID),
		ItemType:     text.ItemType,
		Lang:         text.Lang,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrItemTextNotFound
		}
		return nil, fmt.Errorf("update item text: %w", err)
	}
	return itemTextToDomain(row), nil
}

func (r *ItemTextRepository) Delete(ctx context.Context, itemID uuid.UUID, itemType string, lang string) error {
	err := r.q.DeleteItemText(ctx, itdb.DeleteItemTextParams{
		ItemID:   uuidToPgtype(itemID),
		ItemType: itemType,
		Lang:     lang,
	})
	if err != nil {
		return fmt.Errorf("delete item text: %w", err)
	}
	return nil
}

func (r *ItemTextRepository) DeleteByItem(ctx context.Context, itemID uuid.UUID, itemType string) error {
	err := r.q.DeleteItemTextsByItem(ctx, itdb.DeleteItemTextsByItemParams{
		ItemID:   uuidToPgtype(itemID),
		ItemType: itemType,
	})
	if err != nil {
		return fmt.Errorf("delete item texts by item: %w", err)
	}
	return nil
}

func (r *ItemTextRepository) DeleteByItemID(ctx context.Context, itemID uuid.UUID) error {
	err := r.q.DeleteItemTextsByItemID(ctx, uuidToPgtype(itemID))
	if err != nil {
		return fmt.Errorf("delete item texts by item id: %w", err)
	}
	return nil
}

func itemTextToDomain(row itdb.ItemText) *domain.ItemText {
	return &domain.ItemText{
		ItemID:       pgtypeToUUID(row.ItemID),
		ItemType:     row.ItemType,
		Lang:         row.Lang,
		Description:  row.Description,
		IsSource:     row.IsSource,
		ModelVersion: pgtypeToString(row.ModelVersion),
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
