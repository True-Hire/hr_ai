package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Skill struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

type SkillRepository interface {
	Upsert(ctx context.Context, skill *Skill) (*Skill, error)
	GetByName(ctx context.Context, name string) (*Skill, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Skill, error)
	ListAll(ctx context.Context) ([]Skill, error)
	Search(ctx context.Context, query string) ([]Skill, error)
	AddUserSkill(ctx context.Context, userID uuid.UUID, skillID uuid.UUID) error
	RemoveUserSkills(ctx context.Context, userID uuid.UUID) error
	ListUserSkills(ctx context.Context, userID uuid.UUID) ([]Skill, error)
	AddVacancySkill(ctx context.Context, vacancyID uuid.UUID, skillID uuid.UUID) error
	RemoveVacancySkills(ctx context.Context, vacancyID uuid.UUID) error
	ListVacancySkills(ctx context.Context, vacancyID uuid.UUID) ([]Skill, error)
}
