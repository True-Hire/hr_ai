package http

import (
	"time"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

type ProfileParseTextRequest struct {
	UserInput string `json:"user_input" binding:"required"`
}

type ProfileParseResponse struct {
	SourceLang string                           `json:"source_lang"`
	Fields     []ProfileParseFieldResponse      `json:"fields"`
	Experience []ProfileParseExperienceResponse `json:"experience"`
	Education  []ProfileParseEducationResponse  `json:"education"`
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

type ProfileParseExperienceResponse struct {
	ID        string                          `json:"id"`
	Company   string                          `json:"company"`
	Position  string                          `json:"position"`
	StartDate string                          `json:"start_date,omitempty"`
	EndDate   string                          `json:"end_date,omitempty"`
	Projects  string                          `json:"projects,omitempty"`
	WebSite   string                          `json:"web_site,omitempty"`
	Texts     []ProfileParseItemTextResponse  `json:"texts"`
}

// Note: Projects in parse response is raw JSON string as stored in DB.
// The GET /users/:id endpoint resolves it to structured ProjectResponse per language.

type ProfileParseEducationResponse struct {
	ID           string                          `json:"id"`
	Institution  string                          `json:"institution"`
	Degree       string                          `json:"degree"`
	FieldOfStudy string                          `json:"field_of_study,omitempty"`
	StartDate    string                          `json:"start_date,omitempty"`
	EndDate      string                          `json:"end_date,omitempty"`
	Location     string                          `json:"location,omitempty"`
	Texts        []ProfileParseItemTextResponse  `json:"texts"`
}

type ProfileParseItemTextResponse struct {
	Lang         string `json:"lang"`
	Description  string `json:"description"`
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

	experience := make([]ProfileParseExperienceResponse, 0, len(result.Experience))
	for _, e := range result.Experience {
		texts := make([]ProfileParseItemTextResponse, 0, len(e.Texts))
		for _, t := range e.Texts {
			texts = append(texts, ProfileParseItemTextResponse{
				Lang:         t.Lang,
				Description:  t.Description,
				IsSource:     t.IsSource,
				ModelVersion: t.ModelVersion,
				UpdatedAt:    t.UpdatedAt.Format(time.RFC3339),
			})
		}
		experience = append(experience, ProfileParseExperienceResponse{
			ID:        e.Item.ID.String(),
			Company:   e.Item.Company,
			Position:  e.Item.Position,
			StartDate: e.Item.StartDate,
			EndDate:   e.Item.EndDate,
			Projects:  e.Item.Projects,
			WebSite:   e.Item.WebSite,
			Texts:     texts,
		})
	}

	education := make([]ProfileParseEducationResponse, 0, len(result.Education))
	for _, e := range result.Education {
		texts := make([]ProfileParseItemTextResponse, 0, len(e.Texts))
		for _, t := range e.Texts {
			texts = append(texts, ProfileParseItemTextResponse{
				Lang:         t.Lang,
				Description:  t.Description,
				IsSource:     t.IsSource,
				ModelVersion: t.ModelVersion,
				UpdatedAt:    t.UpdatedAt.Format(time.RFC3339),
			})
		}
		education = append(education, ProfileParseEducationResponse{
			ID:           e.Item.ID.String(),
			Institution:  e.Item.Institution,
			Degree:       e.Item.Degree,
			FieldOfStudy: e.Item.FieldOfStudy,
			StartDate:    e.Item.StartDate,
			EndDate:      e.Item.EndDate,
			Location:     e.Item.Location,
			Texts:        texts,
		})
	}

	return ProfileParseResponse{
		SourceLang: result.SourceLang,
		Fields:     fields,
		Experience: experience,
		Education:  education,
	}
}
