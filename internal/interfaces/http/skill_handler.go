package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

type SkillHandler struct {
	service *application.SkillService
}

func NewSkillHandler(service *application.SkillService) *SkillHandler {
	return &SkillHandler{service: service}
}

type SkillResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Search godoc
// @Summary Search or list all skills
// @Description Returns skills matching the query, or all skills if no query is provided. Used for autocomplete/suggestions.
// @Tags skills
// @Produce json
// @Param q query string false "Search query"
// @Success 200 {array} SkillResponse
// @Failure 500 {object} ErrorResponse
// @Router /skills [get]
func (h *SkillHandler) Search(c *gin.Context) {
	query := c.Query("q")

	var skills []SkillResponse
	var err error

	if query == "" {
		result, e := h.service.ListAllSkills(c.Request.Context())
		err = e
		for _, s := range result {
			skills = append(skills, SkillResponse{ID: s.ID.String(), Name: s.Name})
		}
	} else {
		result, e := h.service.SearchSkills(c.Request.Context(), query)
		err = e
		for _, s := range result {
			skills = append(skills, SkillResponse{ID: s.ID.String(), Name: s.Name})
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to search skills"})
		return
	}

	if skills == nil {
		skills = []SkillResponse{}
	}
	c.JSON(http.StatusOK, skills)
}
