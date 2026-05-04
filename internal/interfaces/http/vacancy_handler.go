package http

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type VacancyHandler struct {
	service          *application.VacancyService
	companyHRSvc     *application.CompanyHRService
	vacancySearchSvc *application.VacancySearchService
	vacancyAppSvc    *application.VacancyApplicationService
	candidateSearch  *application.CandidateSearchService
}

func NewVacancyHandler(service *application.VacancyService, companyHRSvc *application.CompanyHRService, vacancySearchSvc *application.VacancySearchService, vacancyAppSvc *application.VacancyApplicationService, candidateSearch *application.CandidateSearchService) *VacancyHandler {
	return &VacancyHandler{service: service, companyHRSvc: companyHRSvc, vacancySearchSvc: vacancySearchSvc, vacancyAppSvc: vacancyAppSvc, candidateSearch: candidateSearch}
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

	// Fetch HR to get company data
	hr, err := h.companyHRSvc.GetCompanyHR(c.Request.Context(), hrID)
	if err != nil {
		log.Printf("create vacancy: failed to get hr: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get hr"})
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
		CompanyData:      hr.CompanyData,
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

	// Fetch HR to get company data
	hr, err := h.companyHRSvc.GetCompanyHR(c.Request.Context(), hrID)
	if err != nil {
		log.Printf("parse vacancy: failed to get hr: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get hr"})
		return
	}

	result, err := h.service.ParseVacancy(c.Request.Context(), hrID, hr.CompanyData, req.UserInput)
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
// @Summary List vacancies with pagination and optional filters
// @Tags vacancies
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param q query string false "Search query"
// @Param lang query string false "Language code" default(en)
// @Param status query string false "Filter by status (active, draft, closed)"
// @Param format query string false "Filter by format (office, remote, hybrid)"
// @Param schedule query string false "Filter by schedule (full-time, part-time)"
// @Param salary_currency query string false "Filter by salary currency"
// @Param salary_min query int false "Min salary (returns vacancies that pay at least this)"
// @Param salary_max query int false "Max salary (returns vacancies that pay at most this)"
// @Param experience_min query int false "Min experience (returns vacancies accepting this level)"
// @Param experience_max query int false "Max experience (returns vacancies accepting this level)"
// @Param country_id query string false "Filter by country ID (UUID)"
// @Param hr_id query string false "Filter by HR ID (UUID)"
// @Success 200 {object} PaginatedVacanciesResponse
// @Failure 500 {object} ErrorResponse
// @Router /vacancies [get]
func (h *VacancyHandler) List(c *gin.Context) {
	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)
	query := c.Query("q")
	lang := c.DefaultQuery("lang", "en")

	// Build filter from query params
	filter := h.buildVacancyFilter(c)

	var result *application.ListVacanciesResult
	var err error

	// If semantic search query is provided, use vector search
	if query != "" && h.vacancySearchSvc != nil {
		result, err = h.vacancySearchSvc.SearchVacancies(c.Request.Context(), query, page, pageSize)
	} else if hasVacancyFilters(filter) {
		// Use filtered query when any filter is set
		result, err = h.service.ListVacanciesFiltered(c.Request.Context(), filter, page, pageSize)
	} else {
		result, err = h.service.ListVacancies(c.Request.Context(), page, pageSize)
	}

	if err != nil {
		log.Printf("list vacancies error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list vacancies"})
		return
	}

	resp := PaginatedVacanciesResponse{
		Vacancies: make([]VacancyResponse, len(result.Vacancies)),
		Total:     result.Total,
		Page:      page,
		PageSize:  pageSize,
	}

	// Build base responses
	for i, vwd := range result.Vacancies {
		resp.Vacancies[i] = toVacancyResponse(&vwd, lang)
	}

	// Enrich with application counts and matching candidates concurrently
	ctx := c.Request.Context()
	var wg sync.WaitGroup
	for i, vwd := range result.Vacancies {
		wg.Add(1)
		go func(idx int, v application.VacancyWithDetails) {
			defer wg.Done()

			// Application count (cheap DB query)
			if h.vacancyAppSvc != nil {
				if count, err := h.vacancyAppSvc.CountByVacancy(ctx, v.Vacancy.ID); err == nil {
					resp.Vacancies[idx].ApplicationCount = &count
				}
			}

			// Matching candidates count
			if h.candidateSearch != nil {
				count := h.candidateSearch.CountMatchingByVacancy(ctx, v.Vacancy.ID)
				resp.Vacancies[idx].MatchingCandidatesCount = &count
			}
		}(i, vwd)
	}
	wg.Wait()

	c.JSON(http.StatusOK, resp)
}

func (h *VacancyHandler) buildVacancyFilter(c *gin.Context) domain.VacancyFilter {
	var filter domain.VacancyFilter

	// HR calling via middleware → auto-filter by their ID
	if hrID, exists := c.Get("hr_id"); exists {
		if parsed, err := uuid.Parse(hrID.(string)); err == nil {
			filter.HRID = parsed
		}
	}

	// Explicit hr_id query param overrides (for admin use)
	if hrIDParam := c.Query("hr_id"); hrIDParam != "" {
		if parsed, err := uuid.Parse(hrIDParam); err == nil {
			filter.HRID = parsed
		}
	}

	filter.Status = c.Query("status")
	filter.Format = c.Query("format")
	filter.Schedule = c.Query("schedule")
	filter.SalaryCurrency = c.Query("salary_currency")
	filter.SalaryMin = parseQueryInt32(c, "salary_min", 0)
	filter.SalaryMax = parseQueryInt32(c, "salary_max", 0)
	filter.ExperienceMin = parseQueryInt32(c, "experience_min", 0)
	filter.ExperienceMax = parseQueryInt32(c, "experience_max", 0)

	if countryID := c.Query("country_id"); countryID != "" {
		if parsed, err := uuid.Parse(countryID); err == nil {
			filter.CountryID = parsed
		}
	}

	return filter
}

func hasVacancyFilters(f domain.VacancyFilter) bool {
	return f.HRID != uuid.Nil || f.Status != "" || f.Format != "" ||
		f.Schedule != "" || f.SalaryCurrency != "" ||
		f.SalaryMin > 0 || f.SalaryMax > 0 ||
		f.ExperienceMin > 0 || f.ExperienceMax > 0 ||
		f.CountryID != uuid.Nil
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

	var mainCatID uuid.UUID
	if req.MainCategoryID != "" {
		mainCatID, _ = uuid.Parse(req.MainCategoryID)
	}
	var subCatID uuid.UUID
	if req.SubCategoryID != "" {
		subCatID, _ = uuid.Parse(req.SubCategoryID)
	}

	input := &application.UpdateVacancyInput{
		ID:               id,
		CountryID:        countryID,
		MainCategoryID:   mainCatID,
		SubCategoryID:    subCatID,
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
