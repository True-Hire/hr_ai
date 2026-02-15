package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

type MiniAppHandler struct {
	vacancySearchSvc *application.VacancySearchService
	vacancySvc       *application.VacancyService
}

func NewMiniAppHandler(vacancySearchSvc *application.VacancySearchService, vacancySvc *application.VacancyService) *MiniAppHandler {
	return &MiniAppHandler{vacancySearchSvc: vacancySearchSvc, vacancySvc: vacancySvc}
}

// ListForUser returns vacancies matching the authenticated user's profile.
func (h *MiniAppHandler) ListForUser(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user"})
		return
	}

	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)
	lang := c.DefaultQuery("lang", "en")

	result, err := h.vacancySearchSvc.MatchVacanciesForUser(c.Request.Context(), userID, lang, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to search vacancies"})
		return
	}

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

// Search searches vacancies by a user-typed query using vector similarity.
func (h *MiniAppHandler) Search(c *gin.Context) {
	query := c.Query("q")
	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)
	lang := c.DefaultQuery("lang", "en")

	var result *application.ListVacanciesResult
	var err error

	if query == "" {
		// No query — return user-matched vacancies
		userIDStr := c.GetString("user_id")
		userID, parseErr := uuid.Parse(userIDStr)
		if parseErr != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user"})
			return
		}
		result, err = h.vacancySearchSvc.MatchVacanciesForUser(c.Request.Context(), userID, lang, page, pageSize)
	} else {
		result, err = h.vacancySearchSvc.SearchVacancies(c.Request.Context(), query, page, pageSize)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to search vacancies"})
		return
	}

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

// GetByID returns a single vacancy by ID.
func (h *MiniAppHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid vacancy id"})
		return
	}

	result, err := h.vacancySvc.GetVacancy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "vacancy not found"})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	c.JSON(http.StatusOK, toVacancyResponse(result, lang))
}
