package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CompanyHandler struct {
	service *application.CompanyService
}

func NewCompanyHandler(service *application.CompanyService) *CompanyHandler {
	return &CompanyHandler{service: service}
}

// Create godoc
// @Summary Create a new company (data is translated via Gemini into uz/ru/en)
// @Tags companies
// @Accept json
// @Produce json
// @Param request body CreateCompanyRequest true "Company data"
// @Success 201 {object} CompanyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies [post]
func (h *CompanyHandler) Create(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	input := &application.CreateCompanyInput{
		EmployeeCount:   req.EmployeeCount,
		Country:         req.Country,
		Address:         req.Address,
		Phone:           req.Phone,
		Telegram:        req.Telegram,
		TelegramChannel: req.TelegramChannel,
		Email:           req.Email,
		LogoURL:         req.LogoURL,
		WebSite:         req.WebSite,
		Instagram:       req.Instagram,
		Name:            req.Name,
		ActivityType:    req.ActivityType,
		CompanyType:     req.CompanyType,
		About:           req.About,
		Market:          req.Market,
	}

	result, err := h.service.CreateCompany(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create company"})
		return
	}

	c.JSON(http.StatusCreated, toCompanyResponse(result))
}

// GetByID godoc
// @Summary Get company by ID with all translations
// @Tags companies
// @Produce json
// @Param id path string true "Company ID (UUID)"
// @Success 200 {object} CompanyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies/{id} [get]
func (h *CompanyHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid company id"})
		return
	}

	result, err := h.service.GetCompany(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCompanyNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "company not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get company"})
		return
	}

	c.JSON(http.StatusOK, toCompanyResponse(result))
}

// List godoc
// @Summary List companies with pagination
// @Tags companies
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} PaginatedCompaniesResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies [get]
func (h *CompanyHandler) List(c *gin.Context) {
	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)

	result, err := h.service.ListCompanies(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list companies"})
		return
	}

	resp := PaginatedCompaniesResponse{
		Companies: make([]CompanyResponse, 0, len(result.Companies)),
		Total:     result.Total,
		Page:      page,
		PageSize:  pageSize,
	}
	for _, cwt := range result.Companies {
		resp.Companies = append(resp.Companies, toCompanyResponse(&cwt))
	}

	c.JSON(http.StatusOK, resp)
}

// Update godoc
// @Summary Update company (non-text fields only)
// @Tags companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID (UUID)"
// @Param request body UpdateCompanyRequest true "Updated company data"
// @Success 200 {object} CompanyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies/{id} [put]
func (h *CompanyHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid company id"})
		return
	}

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	company := &domain.Company{
		ID:              id,
		EmployeeCount:   req.EmployeeCount,
		Country:         req.Country,
		Address:         req.Address,
		Phone:           req.Phone,
		Telegram:        req.Telegram,
		TelegramChannel: req.TelegramChannel,
		Email:           req.Email,
		LogoURL:         req.LogoURL,
		WebSite:         req.WebSite,
		Instagram:       req.Instagram,
	}

	result, err := h.service.UpdateCompany(c.Request.Context(), company)
	if err != nil {
		if errors.Is(err, domain.ErrCompanyNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "company not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update company"})
		return
	}

	c.JSON(http.StatusOK, toCompanyResponse(result))
}

// Delete godoc
// @Summary Delete a company and all its translations
// @Tags companies
// @Param id path string true "Company ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies/{id} [delete]
func (h *CompanyHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid company id"})
		return
	}

	if err := h.service.DeleteCompany(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrCompanyNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "company not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete company"})
		return
	}

	c.Status(http.StatusNoContent)
}
