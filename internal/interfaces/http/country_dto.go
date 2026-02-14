package http

import (
	"time"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

type CountryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortCode string `json:"short_code"`
	CreatedAt string `json:"created_at"`
}

func toCountryResponse(cwt *application.CountryWithTexts, lang string) CountryResponse {
	resp := CountryResponse{
		ID:        cwt.Country.ID.String(),
		Name:      cwt.Country.Name,
		ShortCode: cwt.Country.ShortCode,
		CreatedAt: cwt.Country.CreatedAt.Format(time.RFC3339),
	}

	for _, t := range cwt.Texts {
		if t.Lang == lang {
			resp.Name = t.Name
			break
		}
	}

	return resp
}
