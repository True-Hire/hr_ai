package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	sdb "github.com/ruziba3vich/hr-ai/db/sqlc/skills"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type SkillRepository struct {
	q *sdb.Queries
}

func NewSkillRepository(pool *pgxpool.Pool) *SkillRepository {
	return &SkillRepository{
		q: sdb.New(pool),
	}
}

func (r *SkillRepository) Upsert(ctx context.Context, skill *domain.Skill) (*domain.Skill, error) {
	row, err := r.q.UpsertSkill(ctx, sdb.UpsertSkillParams{
		ID:   uuidToPgtype(skill.ID),
		Name: skill.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert skill: %w", err)
	}
	return skillToDomain(row), nil
}

func (r *SkillRepository) GetByName(ctx context.Context, name string) (*domain.Skill, error) {
	row, err := r.q.GetSkillByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get skill by name: %w", err)
	}
	return skillToDomain(row), nil
}

func (r *SkillRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Skill, error) {
	row, err := r.q.GetSkillByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get skill by id: %w", err)
	}
	return skillToDomain(row), nil
}

func (r *SkillRepository) ListAll(ctx context.Context) ([]domain.Skill, error) {
	rows, err := r.q.ListSkills(ctx)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	result := make([]domain.Skill, 0, len(rows))
	for _, row := range rows {
		result = append(result, *skillToDomain(row))
	}
	return result, nil
}

func (r *SkillRepository) Search(ctx context.Context, query string) ([]domain.Skill, error) {
	rows, err := r.q.SearchSkills(ctx, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("search skills: %w", err)
	}
	result := make([]domain.Skill, 0, len(rows))
	for _, row := range rows {
		result = append(result, *skillToDomain(row))
	}
	return result, nil
}

func (r *SkillRepository) AddUserSkill(ctx context.Context, userID uuid.UUID, skillID uuid.UUID) error {
	err := r.q.AddUserSkill(ctx, sdb.AddUserSkillParams{
		UserID:  uuidToPgtype(userID),
		SkillID: uuidToPgtype(skillID),
	})
	if err != nil {
		return fmt.Errorf("add user skill: %w", err)
	}
	return nil
}

func (r *SkillRepository) RemoveUserSkills(ctx context.Context, userID uuid.UUID) error {
	err := r.q.RemoveUserSkills(ctx, uuidToPgtype(userID))
	if err != nil {
		return fmt.Errorf("remove user skills: %w", err)
	}
	return nil
}

func (r *SkillRepository) ListUserSkills(ctx context.Context, userID uuid.UUID) ([]domain.Skill, error) {
	rows, err := r.q.ListUserSkills(ctx, uuidToPgtype(userID))
	if err != nil {
		return nil, fmt.Errorf("list user skills: %w", err)
	}
	result := make([]domain.Skill, 0, len(rows))
	for _, row := range rows {
		result = append(result, *skillToDomain(row))
	}
	return result, nil
}

func skillToDomain(row sdb.Skill) *domain.Skill {
	return &domain.Skill{
		ID:        pgtypeToUUID(row.ID),
		Name:      row.Name,
		CreatedAt: row.CreatedAt.Time,
	}
}
