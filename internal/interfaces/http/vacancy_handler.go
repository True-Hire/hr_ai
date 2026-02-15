package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type VacancyHandler struct {
	service *application.VacancyService
}

func NewVacancyHandler(service *application.VacancyService) *VacancyHandler {
	return &VacancyHandler{service: service}
}

// Create godoc
// @Summary Create a new vacancy (text fields translated via Gemini into uz/ru/en)
// @Tags vacancies
// @Accept json
// @Produce json
// @Param request body CreateVacancyRequest true "Vacancy data"
// @Success 201 {object} VacancyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /vacancies [post]
func (h *VacancyHandler) Create(c *gin.Context) {
	hrID, err := uuid.Parse(c.GetString("hr_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid hr_id in token"})
		return
	}

	var req CreateVacancyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid company_id"})
		return
	}

	var countryID uuid.UUID
	if req.CountryID != "" {
		countryID, err = uuid.Parse(req.CountryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid country_id"})
			return
		}
	}

	input := &application.CreateVacancyInput{
		HRID:             hrID,
		CompanyID:        companyID,
		CountryID:        countryID,
		SalaryMin:        req.SalaryMin,
		SalaryMax:        req.SalaryMax,
		SalaryCurrency:   req.SalaryCurrency,
		ExperienceMin:    req.ExperienceMin,
		ExperienceMax:    req.ExperienceMax,
		Format:           req.Format,
		Schedule:         req.Schedule,
		Phone:            req.Phone,
		Telegram:         req.Telegram,
		Email:            req.Email,
		Address:          req.Address,
		Title:            req.Title,
		Description:      req.Description,
		Responsibilities: req.Responsibilities,
		Requirements:     req.Requirements,
		Benefits:         req.Benefits,
		Skills:           req.Skills,
	}

	result, err := h.service.CreateVacancy(c.Request.Context(), input)
	if err != nil {
		log.Printf("create vacancy error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create vacancy"})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	c.JSON(http.StatusCreated, toVacancyResponse(result, lang))
}

// Parse godoc
// @Summary Parse a vacancy from free-form text (Gemini extracts all fields + translates)
// @Tags vacancies
// @Accept json
// @Produce json
// @Param request body VacancyParseRequest true "Free-form vacancy text"
// @Success 201 {object} VacancyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /vacancies/parse [post]
func (h *VacancyHandler) Parse(c *gin.Context) {
	hrID, err := uuid.Parse(c.GetString("hr_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid hr_id in token"})
		return
	}

	var req VacancyParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid company_id"})
		return
	}

	result, err := h.service.ParseVacancy(c.Request.Context(), hrID, companyID, req.UserInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to parse vacancy"})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	c.JSON(http.StatusCreated, toVacancyResponse(result, lang))
}

// GetByID godoc
// @Summary Get vacancy by ID with all translations and skills
// @Tags vacancies
// @Produce json
// @Param id path string true "Vacancy ID (UUID)"
// @Success 200 {object} VacancyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /vacancies/{id} [get]
func (h *VacancyHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid vacancy id"})
		return
	}

	result, err := h.service.GetVacancy(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrVacancyNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "vacancy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get vacancy"})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	c.JSON(http.StatusOK, toVacancyResponse(result, lang))
}

// List godoc
// @Summary List vacancies with pagination (optional company_id filter)
// @Tags vacancies
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param company_id query string false "Filter by company ID"
// @Success 200 {object} PaginatedVacanciesResponse
// @Failure 500 {object} ErrorResponse
// @Router /vacancies [get]
func (h *VacancyHandler) List(c *gin.Context) {
	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)
	companyIDStr := c.Query("company_id")

	var result *application.ListVacanciesResult
	var err error

	if companyIDStr != "" {
		companyID, parseErr := uuid.Parse(companyIDStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid company_id"})
			return
		}
		result, err = h.service.ListVacanciesByCompany(c.Request.Context(), companyID, page, pageSize)
	} else {
		result, err = h.service.ListVacancies(c.Request.Context(), page, pageSize)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list vacancies"})
		return
	}

	lang := c.DefaultQuery("lang", "en")

	resp := PaginatedVacanciesResponse{
		Vacancies: make([]VacancyResponse, 0, len(result.Vacancies)),
		Total:     result.Total,
		Page:      page,
		PageSize:  pageSize,
	}
	for _, vwd := range result.Vacancies {
		resp.Vacancies = append(resp.Vacancies, toVacancyResponse(&vwd, lang))
	}

	c.JSON(http.StatusOK, resp)
}

// Update godoc
// @Summary Update vacancy (non-text fields only)
// @Tags vacancies
// @Accept json
// @Produce json
// @Param id path string true "Vacancy ID (UUID)"
// @Param request body UpdateVacancyRequest true "Updated vacancy data"
// @Success 200 {object} VacancyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /vacancies/{id} [put]
func (h *VacancyHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid vacancy id"})
		return
	}

	var req UpdateVacancyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var countryID uuid.UUID
	if req.CountryID != "" {
		countryID, err = uuid.Parse(req.CountryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid country_id"})
			return
		}
	}

	input := &application.UpdateVacancyInput{
		ID:               id,
		CountryID:        countryID,
		SalaryMin:        req.SalaryMin,
		SalaryMax:        req.SalaryMax,
		SalaryCurrency:   req.SalaryCurrency,
		ExperienceMin:    req.ExperienceMin,
		ExperienceMax:    req.ExperienceMax,
		Format:           req.Format,
		Schedule:         req.Schedule,
		Phone:            req.Phone,
		Telegram:         req.Telegram,
		Email:            req.Email,
		Address:          req.Address,
		Status:           req.Status,
		Title:            req.Title,
		Description:      req.Description,
		Responsibilities: req.Responsibilities,
		Requirements:     req.Requirements,
		Benefits:         req.Benefits,
		Skills:           req.Skills,
	}

	result, err := h.service.UpdateVacancy(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrVacancyNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "vacancy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update vacancy"})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	c.JSON(http.StatusOK, toVacancyResponse(result, lang))
}

// Delete godoc
// @Summary Delete a vacancy and all its translations and skills
// @Tags vacancies
// @Param id path string true "Vacancy ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /vacancies/{id} [delete]
func (h *VacancyHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid vacancy id"})
		return
	}

	if err := h.service.DeleteVacancy(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrVacancyNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "vacancy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete vacancy"})
		return
	}

	c.Status(http.StatusNoContent)
}
