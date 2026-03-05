package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
)

type HRMiniAppHandler struct {
	hrService    *application.CompanyHRService
	geminiClient *gemini.Client
}

func NewHRMiniAppHandler(hrService *application.CompanyHRService, geminiClient *gemini.Client) *HRMiniAppHandler {
	return &HRMiniAppHandler{
		hrService:    hrService,
		geminiClient: geminiClient,
	}
}

// GetMe godoc
// @Summary Get current HR profile with company data
// @Tags hr-miniapp
// @Produce json
// @Success 200 {object} HRMiniAppMeResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security TelegramAuth
// @Router /hr-miniapp/me [get]
func (h *HRMiniAppHandler) GetMe(c *gin.Context) {
	hrID, err := uuid.Parse(c.GetString("hr_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	hr, err := h.hrService.GetCompanyHR(c.Request.Context(), hrID)
	if err != nil {
		if errors.Is(err, domain.ErrCompanyHRNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "company hr not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get company hr"})
		return
	}

	resp := HRMiniAppMeResponse{
		CompanyHRResponse: toCompanyHRResponse(hr),
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateMe godoc
// @Summary Update current HR profile and optionally create/update company data
// @Tags hr-miniapp
// @Accept json
// @Produce json
// @Param request body HRMiniAppUpdateRequest true "HR and optional company data"
// @Success 200 {object} HRMiniAppMeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security TelegramAuth
// @Router /hr-miniapp/me [put]
func (h *HRMiniAppHandler) UpdateMe(c *gin.Context) {
	hrID, err := uuid.Parse(c.GetString("hr_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	var req HRMiniAppUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	hr, err := h.hrService.GetCompanyHR(c.Request.Context(), hrID)
	if err != nil {
		if errors.Is(err, domain.ErrCompanyHRNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "company hr not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get company hr"})
		return
	}

	// Merge non-empty request fields into existing HR
	if req.FirstName != "" {
		hr.FirstName = req.FirstName
	}
	if req.LastName != "" {
		hr.LastName = req.LastName
	}
	if req.Patronymic != "" {
		hr.Patronymic = req.Patronymic
	}
	if req.Phone != "" {
		hr.Phone = req.Phone
	}
	if req.Email != "" {
		hr.Email = req.Email
	}
	if req.Position != "" {
		hr.Position = req.Position
	}

	// Handle company data — translate text fields via Gemini, embed as JSONB
	if req.Company != nil {
		cd := hr.CompanyData
		if cd == nil {
			cd = &domain.CompanyData{}
		}

		// Update non-text fields
		if req.Company.EmployeeCount > 0 {
			cd.EmployeeCount = req.Company.EmployeeCount
		}
		if req.Company.Country != "" {
			cd.Country = req.Company.Country
		}
		if req.Company.Address != "" {
			cd.Address = req.Company.Address
		}
		if req.Company.Phone != "" {
			cd.Phone = req.Company.Phone
		}
		if req.Company.Telegram != "" {
			cd.Telegram = req.Company.Telegram
		}
		if req.Company.TelegramChannel != "" {
			cd.TelegramChannel = req.Company.TelegramChannel
		}
		if req.Company.Email != "" {
			cd.Email = req.Company.Email
		}
		if req.Company.LogoURL != "" {
			cd.LogoURL = req.Company.LogoURL
		}
		if req.Company.WebSite != "" {
			cd.WebSite = req.Company.WebSite
		}
		if req.Company.Instagram != "" {
			cd.Instagram = req.Company.Instagram
		}

		// Build text fields for Gemini translation
		textData := map[string]string{}
		if req.Company.Name != "" {
			textData["name"] = req.Company.Name
		}
		if req.Company.ActivityType != "" {
			textData["activity_type"] = req.Company.ActivityType
		}
		if req.Company.CompanyType != "" {
			textData["company_type"] = req.Company.CompanyType
		}
		if req.Company.About != "" {
			textData["about"] = req.Company.About
		}
		if req.Company.Market != "" {
			textData["market"] = req.Company.Market
		}

		if len(textData) > 0 {
			jsonBytes, _ := json.Marshal(textData)
			parsed, err := h.geminiClient.TranslateCompany(c.Request.Context(), string(jsonBytes))
			if err != nil {
				log.Printf("hr-miniapp: failed to translate company: %v", err)
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("failed to translate company texts: %v", err)})
				return
			}
			cd.SourceLang = parsed.SourceLang

			texts := make([]domain.CompanyDataText, 0, 3)
			for _, lang := range []string{"uz", "ru", "en"} {
				t := domain.CompanyDataText{
					Lang:     lang,
					IsSource: lang == parsed.SourceLang,
				}
				if translations, ok := parsed.Fields["name"]; ok {
					t.Name = translations[lang]
				}
				if translations, ok := parsed.Fields["activity_type"]; ok {
					t.ActivityType = translations[lang]
				}
				if translations, ok := parsed.Fields["company_type"]; ok {
					t.CompanyType = translations[lang]
				}
				if translations, ok := parsed.Fields["about"]; ok {
					t.About = translations[lang]
				}
				if translations, ok := parsed.Fields["market"]; ok {
					t.Market = translations[lang]
				}
				texts = append(texts, t)
			}
			cd.Texts = texts
		}

		hr.CompanyData = cd
	}

	// Save HR
	updated, err := h.hrService.UpdateCompanyHR(c.Request.Context(), hr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update company hr"})
		return
	}

	resp := HRMiniAppMeResponse{
		CompanyHRResponse: toCompanyHRResponse(updated),
	}

	c.JSON(http.StatusOK, resp)
}
