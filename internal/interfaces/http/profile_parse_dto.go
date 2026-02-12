package http

import (
	"time"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

type ProfileParseTextRequest struct {
	UserInput string `json:"user_input" binding:"required"`
}

type ProfileParseResponse struct {
	SourceLang string                      `json:"source_lang"`
	Fields     []ProfileParseFieldResponse `json:"fields"`
}

type ProfileParseFieldResponse struct {
	ID        string                          `json:"id"`
	FieldName string                          `json:"field_name"`
	Texts     []ProfileParseFieldTextResponse `json:"texts"`
}

type ProfileParseFieldTextResponse struct {
	Lang         string `json:"lang"`
	Content      string `json:"content"`
	IsSource     bool   `json:"is_source"`
	ModelVersion string `json:"model_version,omitempty"`
	UpdatedAt    string `json:"updated_at"`
}

func toProfileParseResponse(result *application.ParseResult) ProfileParseResponse {
	fields := make([]ProfileParseFieldResponse, 0, len(result.Fields))
	for _, f := range result.Fields {
		texts := make([]ProfileParseFieldTextResponse, 0, len(f.Texts))
		for _, t := range f.Texts {
			texts = append(texts, ProfileParseFieldTextResponse{
				Lang:         t.Lang,
				Content:      t.Content,
				IsSource:     t.IsSource,
				ModelVersion: t.ModelVersion,
				UpdatedAt:    t.UpdatedAt.Format(time.RFC3339),
			})
		}
		fields = append(fields, ProfileParseFieldResponse{
			ID:        f.Field.ID.String(),
			FieldName: f.Field.FieldName,
			Texts:     texts,
		})
	}
	return ProfileParseResponse{
		SourceLang: result.SourceLang,
		Fields:     fields,
	}
}
