package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type ProfileFieldTextService struct {
	repo domain.ProfileFieldTextRepository
}

func NewProfileFieldTextService(repo domain.ProfileFieldTextRepository) *ProfileFieldTextService {
	return &ProfileFieldTextService{repo: repo}
}

func (s *ProfileFieldTextService) CreateProfileFieldText(ctx context.Context, profileFieldID uuid.UUID, lang, content string, isSource bool, modelVersion string) (*domain.ProfileFieldText, error) {
	text := &domain.ProfileFieldText{
		ProfileFieldID: profileFieldID,
		Lang:           lang,
		Content:        content,
		IsSource:       isSource,
		ModelVersion:   modelVersion,
	}
	return s.repo.Create(ctx, text)
}

func (s *ProfileFieldTextService) GetProfileFieldText(ctx context.Context, profileFieldID uuid.UUID, lang string) (*domain.ProfileFieldText, error) {
	return s.repo.Get(ctx, profileFieldID, lang)
}

func (s *ProfileFieldTextService) ListProfileFieldTexts(ctx context.Context, profileFieldID uuid.UUID) ([]domain.ProfileFieldText, error) {
	return s.repo.ListByField(ctx, profileFieldID)
}

func (s *ProfileFieldTextService) UpdateProfileFieldText(ctx context.Context, profileFieldID uuid.UUID, lang, content, modelVersion string) (*domain.ProfileFieldText, error) {
	text := &domain.ProfileFieldText{
		ProfileFieldID: profileFieldID,
		Lang:           lang,
		Content:        content,
		ModelVersion:   modelVersion,
	}
	return s.repo.Update(ctx, text)
}

func (s *ProfileFieldTextService) DeleteProfileFieldText(ctx context.Context, profileFieldID uuid.UUID, lang string) error {
	return s.repo.Delete(ctx, profileFieldID, lang)
}

func (s *ProfileFieldTextService) DeleteProfileFieldTextsByField(ctx context.Context, profileFieldID uuid.UUID) error {
	return s.repo.DeleteByField(ctx, profileFieldID)
}
