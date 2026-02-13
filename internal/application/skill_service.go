package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type SkillService struct {
	repo domain.SkillRepository
}

func NewSkillService(repo domain.SkillRepository) *SkillService {
	return &SkillService{repo: repo}
}

func (s *SkillService) SetUserSkills(ctx context.Context, userID uuid.UUID, names []string) ([]domain.Skill, error) {
	if err := s.repo.RemoveUserSkills(ctx, userID); err != nil {
		return nil, fmt.Errorf("remove existing user skills: %w", err)
	}

	var result []domain.Skill
	seen := make(map[string]bool)
	for _, name := range names {
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true

		skill, err := s.repo.Upsert(ctx, &domain.Skill{
			ID:   uuid.New(),
			Name: name,
		})
		if err != nil {
			return nil, fmt.Errorf("upsert skill %q: %w", name, err)
		}

		if err := s.repo.AddUserSkill(ctx, userID, skill.ID); err != nil {
			return nil, fmt.Errorf("add user skill %q: %w", name, err)
		}
		result = append(result, *skill)
	}
	return result, nil
}

func (s *SkillService) ListUserSkills(ctx context.Context, userID uuid.UUID) ([]domain.Skill, error) {
	return s.repo.ListUserSkills(ctx, userID)
}

func (s *SkillService) SetVacancySkills(ctx context.Context, vacancyID uuid.UUID, names []string) ([]domain.Skill, error) {
	if err := s.repo.RemoveVacancySkills(ctx, vacancyID); err != nil {
		return nil, fmt.Errorf("remove existing vacancy skills: %w", err)
	}

	var result []domain.Skill
	seen := make(map[string]bool)
	for _, name := range names {
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true

		skill, err := s.repo.Upsert(ctx, &domain.Skill{
			ID:   uuid.New(),
			Name: name,
		})
		if err != nil {
			return nil, fmt.Errorf("upsert skill %q: %w", name, err)
		}

		if err := s.repo.AddVacancySkill(ctx, vacancyID, skill.ID); err != nil {
			return nil, fmt.Errorf("add vacancy skill %q: %w", name, err)
		}
		result = append(result, *skill)
	}
	return result, nil
}

func (s *SkillService) ListVacancySkills(ctx context.Context, vacancyID uuid.UUID) ([]domain.Skill, error) {
	return s.repo.ListVacancySkills(ctx, vacancyID)
}

func (s *SkillService) RemoveVacancySkills(ctx context.Context, vacancyID uuid.UUID) error {
	return s.repo.RemoveVacancySkills(ctx, vacancyID)
}

func (s *SkillService) SearchSkills(ctx context.Context, query string) ([]domain.Skill, error) {
	return s.repo.Search(ctx, strings.ToLower(strings.TrimSpace(query)))
}

func (s *SkillService) ListAllSkills(ctx context.Context) ([]domain.Skill, error) {
	return s.repo.ListAll(ctx)
}
