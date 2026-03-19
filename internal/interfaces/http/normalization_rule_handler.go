package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type NormalizationRuleHandler struct {
	service *application.NormalizationService
}

func NewNormalizationRuleHandler(service *application.NormalizationService) *NormalizationRuleHandler {
	return &NormalizationRuleHandler{service: service}
}

// --- Request / Response DTOs ---

type CreateNormalizationRuleRequest struct {
	Category        string         `json:"category" binding:"required"`
	SourceValue     string         `json:"source_value" binding:"required"`
	NormalizedValue string         `json:"normalized_value" binding:"required"`
	Metadata        map[string]any `json:"metadata"`
}

type UpdateNormalizationRuleRequest struct {
	Category        string         `json:"category"`
	SourceValue     string         `json:"source_value"`
	NormalizedValue string         `json:"normalized_value"`
	Metadata        map[string]any `json:"metadata"`
}

type NormalizationRuleResponse struct {
	ID              string         `json:"id"`
	Category        string         `json:"category"`
	SourceValue     string         `json:"source_value"`
	NormalizedValue string         `json:"normalized_value"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	CreatedAt       string         `json:"created_at"`
}

type PaginatedNormalizationRulesResponse struct {
	Rules    []NormalizationRuleResponse `json:"rules"`
	Total    int64                       `json:"total"`
	Page     int32                       `json:"page"`
	PageSize int32                       `json:"page_size"`
}

// --- Handlers ---

// Create handles POST /normalization-rules
func (h *NormalizationRuleHandler) Create(c *gin.Context) {
	var req CreateNormalizationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	rule := &domain.NormalizationRule{
		Category:        req.Category,
		SourceValue:     req.SourceValue,
		NormalizedValue: req.NormalizedValue,
		Metadata:        req.Metadata,
	}

	created, err := h.service.CreateRule(c.Request.Context(), rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toNormalizationRuleResponse(created))
}

// List handles GET /normalization-rules
func (h *NormalizationRuleHandler) List(c *gin.Context) {
	category := c.Query("category")
	query := c.Query("q")
	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 50)

	result, err := h.service.ListRules(c.Request.Context(), category, query, int(page), int(pageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list rules: " + err.Error()})
		return
	}

	rules := make([]NormalizationRuleResponse, 0, len(result.Rules))
	for i := range result.Rules {
		rules = append(rules, toNormalizationRuleResponse(&result.Rules[i]))
	}

	c.JSON(http.StatusOK, PaginatedNormalizationRulesResponse{
		Rules:    rules,
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetByID handles GET /normalization-rules/:id
func (h *NormalizationRuleHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	rule, err := h.service.GetRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "rule not found"})
		return
	}

	c.JSON(http.StatusOK, toNormalizationRuleResponse(rule))
}

// Update handles PUT /normalization-rules/:id
func (h *NormalizationRuleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	var req UpdateNormalizationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	rule := &domain.NormalizationRule{
		ID:              id,
		Category:        req.Category,
		SourceValue:     req.SourceValue,
		NormalizedValue: req.NormalizedValue,
		Metadata:        req.Metadata,
	}

	updated, err := h.service.UpdateRule(c.Request.Context(), rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, toNormalizationRuleResponse(updated))
}

// Delete handles DELETE /normalization-rules/:id
func (h *NormalizationRuleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	if err := h.service.DeleteRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete rule: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// --- Helper ---

func toNormalizationRuleResponse(r *domain.NormalizationRule) NormalizationRuleResponse {
	return NormalizationRuleResponse{
		ID:              r.ID.String(),
		Category:        r.Category,
		SourceValue:     r.SourceValue,
		NormalizedValue: r.NormalizedValue,
		Metadata:        r.Metadata,
		CreatedAt:       r.CreatedAt.Format(time.RFC3339),
	}
}
