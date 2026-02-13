package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CompanyHRHandler struct {
	service    *application.CompanyHRService
	hrAuthSvc  *application.HRAuthService
}

func NewCompanyHRHandler(service *application.CompanyHRService, hrAuthSvc *application.HRAuthService) *CompanyHRHandler {
	return &CompanyHRHandler{
		service:   service,
		hrAuthSvc: hrAuthSvc,
	}
}

// Create godoc
// @Summary Create a new company HR
// @Tags hrs
// @Accept json
// @Produce json
// @Param request body CreateCompanyHRRequest true "HR data"
// @Success 201 {object} CompanyHRResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hrs [post]
func (h *CompanyHRHandler) Create(c *gin.Context) {
	var req CreateCompanyHRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	created, err := h.service.CreateCompanyHR(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create company hr"})
		return
	}

	if err := h.hrAuthSvc.SetPassword(c.Request.Context(), created.ID, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "hr created but failed to set password"})
		return
	}

	c.JSON(http.StatusCreated, toCompanyHRResponse(created))
}

// Me godoc
// @Summary Get current authenticated HR with full details
// @Tags hrs
// @Produce json
// @Success 200 {object} CompanyHRResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /hrs/me [get]
func (h *CompanyHRHandler) Me(c *gin.Context) {
	rawID, _ := c.Get("hr_id")
	id, err := uuid.Parse(rawID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	hr, err := h.service.GetCompanyHR(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCompanyHRNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "company hr not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get company hr"})
		return
	}

	c.JSON(http.StatusOK, toCompanyHRResponse(hr))
}

// GetByID godoc
// @Summary Get company HR by ID
// @Tags hrs
// @Produce json
// @Param id path string true "HR ID (UUID)"
// @Success 200 {object} CompanyHRResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hrs/{id} [get]
func (h *CompanyHRHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid hr id"})
		return
	}

	hr, err := h.service.GetCompanyHR(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCompanyHRNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "company hr not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get company hr"})
		return
	}

	c.JSON(http.StatusOK, toCompanyHRResponse(hr))
}

// List godoc
// @Summary List company HRs with pagination
// @Tags hrs
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} PaginatedCompanyHRsResponse
// @Failure 500 {object} ErrorResponse
// @Router /hrs [get]
func (h *CompanyHRHandler) List(c *gin.Context) {
	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)

	result, err := h.service.ListCompanyHRs(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list company hrs"})
		return
	}

	resp := PaginatedCompanyHRsResponse{
		HRs:      make([]CompanyHRResponse, 0, len(result.HRs)),
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, hr := range result.HRs {
		resp.HRs = append(resp.HRs, toCompanyHRResponse(&hr))
	}

	c.JSON(http.StatusOK, resp)
}

// Update godoc
// @Summary Update a company HR
// @Tags hrs
// @Accept json
// @Produce json
// @Param id path string true "HR ID (UUID)"
// @Param request body UpdateCompanyHRRequest true "Updated HR data"
// @Success 200 {object} CompanyHRResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hrs/{id} [put]
func (h *CompanyHRHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid hr id"})
		return
	}

	var req UpdateCompanyHRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	updated, err := h.service.UpdateCompanyHR(c.Request.Context(), req.ToDomain(id))
	if err != nil {
		if errors.Is(err, domain.ErrCompanyHRNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "company hr not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update company hr"})
		return
	}

	c.JSON(http.StatusOK, toCompanyHRResponse(updated))
}

// Delete godoc
// @Summary Delete a company HR
// @Tags hrs
// @Param id path string true "HR ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hrs/{id} [delete]
func (h *CompanyHRHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid hr id"})
		return
	}

	if err := h.service.DeleteCompanyHR(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrCompanyHRNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "company hr not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete company hr"})
		return
	}

	c.Status(http.StatusNoContent)
}
