package http

import (
	"time"

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
