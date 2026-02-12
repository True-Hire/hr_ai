package http

import (
	"time"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CreateProfileFieldTextRequest struct {
	Lang         string `json:"lang" binding:"required"`
	Content      string `json:"content" binding:"required"`
	IsSource     bool   `json:"is_source"`
	ModelVersion string `json:"model_version"`
}

type UpdateProfileFieldTextRequest struct {
	Content      string `json:"content" binding:"required"`
	ModelVersion string `json:"model_version"`
}

func (r *CreateProfileFieldTextRequest) ToDomain(profileFieldID uuid.UUID) *domain.ProfileFieldText {
	return &domain.ProfileFieldText{
		ProfileFieldID: profileFieldID,
		Lang:           r.Lang,
		Content:        r.Content,
		IsSource:       r.IsSource,
		ModelVersion:   r.ModelVersion,
	}
}

func (r *UpdateProfileFieldTextRequest) ToDomain(profileFieldID uuid.UUID, lang string) *domain.ProfileFieldText {
	return &domain.ProfileFieldText{
		ProfileFieldID: profileFieldID,
		Lang:           lang,
		Content:        r.Content,
		ModelVersion:   r.ModelVersion,
	}
}

type ProfileFieldTextResponse struct {
	ProfileFieldID string `json:"profile_field_id"`
	Lang           string `json:"lang"`
	Content        string `json:"content"`
	IsSource       bool   `json:"is_source"`
	ModelVersion   string `json:"model_version,omitempty"`
	UpdatedAt      string `json:"updated_at"`
}

func toProfileFieldTextResponse(t *domain.ProfileFieldText) ProfileFieldTextResponse {
	return ProfileFieldTextResponse{
		ProfileFieldID: t.ProfileFieldID.String(),
		Lang:           t.Lang,
		Content:        t.Content,
		IsSource:       t.IsSource,
		ModelVersion:   t.ModelVersion,
		UpdatedAt:      t.UpdatedAt.Format(time.RFC3339),
	}
}
