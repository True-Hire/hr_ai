package http

import (
	"time"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CreateProfileFieldRequest struct {
	FieldName  string `json:"field_name" binding:"required"`
	SourceLang string `json:"source_lang" binding:"required"`
}

type UpdateProfileFieldRequest struct {
	FieldName  string `json:"field_name"`
	SourceLang string `json:"source_lang"`
}

type ProfileFieldResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	FieldName  string `json:"field_name"`
	SourceLang string `json:"source_lang"`
	UpdatedAt  string `json:"updated_at"`
}

func toProfileFieldResponse(f *domain.ProfileField) ProfileFieldResponse {
	return ProfileFieldResponse{
		ID:         f.ID.String(),
		UserID:     f.UserID.String(),
		FieldName:  f.FieldName,
		SourceLang: f.SourceLang,
		UpdatedAt:  f.UpdatedAt.Format(time.RFC3339),
	}
}
