package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type HRMiniAppHandler struct {
	hrService      *application.CompanyHRService
	companyService *application.CompanyService
}

func NewHRMiniAppHandler(hrService *application.CompanyHRService, companyService *application.CompanyService) *HRMiniAppHandler {
	return &HRMiniAppHandler{
		hrService:      hrService,
		companyService: companyService,
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

	if hr.CompanyID != uuid.Nil {
		cwt, err := h.companyService.GetCompany(c.Request.Context(), hr.CompanyID)
		if err == nil {
			cr := toCompanyResponse(cwt)
			resp.Company = &cr
		}
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateMe godoc
// @Summary Update current HR profile and optionally create/update company
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

	// Handle company data
	if req.Company != nil {
		if hr.CompanyID == uuid.Nil {
			// Create new company
			input := &application.CreateCompanyInput{
				Name:            req.Company.Name,
				ActivityType:    req.Company.ActivityType,
				CompanyType:     req.Company.CompanyType,
				About:           req.Company.About,
				Market:          req.Company.Market,
				EmployeeCount:   req.Company.EmployeeCount,
				Country:         req.Company.Country,
				Address:         req.Company.Address,
				Phone:           req.Company.Phone,
				Telegram:        req.Company.Telegram,
				TelegramChannel: req.Company.TelegramChannel,
				Email:           req.Company.Email,
				LogoURL:         req.Company.LogoURL,
				WebSite:         req.Company.WebSite,
				Instagram:       req.Company.Instagram,
			}
			cwt, err := h.companyService.CreateCompany(c.Request.Context(), input)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create company"})
				return
			}
			hr.CompanyID = cwt.Company.ID
		} else {
			// Update existing company (non-text fields only)
			company := &domain.Company{
				ID:              hr.CompanyID,
				EmployeeCount:   req.Company.EmployeeCount,
				Country:         req.Company.Country,
				Address:         req.Company.Address,
				Phone:           req.Company.Phone,
				Telegram:        req.Company.Telegram,
				TelegramChannel: req.Company.TelegramChannel,
				Email:           req.Company.Email,
				LogoURL:         req.Company.LogoURL,
				WebSite:         req.Company.WebSite,
				Instagram:       req.Company.Instagram,
			}
			if _, err := h.companyService.UpdateCompany(c.Request.Context(), company); err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update company"})
				return
			}
		}
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

	if updated.CompanyID != uuid.Nil {
		cwt, err := h.companyService.GetCompany(c.Request.Context(), updated.CompanyID)
		if err == nil {
			cr := toCompanyResponse(cwt)
			resp.Company = &cr
		}
	}

	c.JSON(http.StatusOK, resp)
}
