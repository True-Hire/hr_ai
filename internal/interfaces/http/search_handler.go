package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

type SearchHandler struct {
	searchSvc      *application.SearchService
	vectorIndexSvc *application.VectorIndexService
	userHandler    *UserHandler
}

func NewSearchHandler(
	searchSvc *application.SearchService,
	vectorIndexSvc *application.VectorIndexService,
	userHandler *UserHandler,
) *SearchHandler {
	return &SearchHandler{
		searchSvc:      searchSvc,
		vectorIndexSvc: vectorIndexSvc,
		userHandler:    userHandler,
	}
}

func (h *SearchHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "query parameter 'q' is required"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	lang := c.DefaultQuery("lang", "en")

	results, err := h.searchSvc.SearchUsers(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "search failed"})
		return
	}

	items := make([]SearchResultItem, 0, len(results))
	for _, r := range results {
		user, err := h.userHandler.service.GetUser(c.Request.Context(), r.UserID)
		if err != nil {
			continue
		}
		profile := h.userHandler.buildUserProfile(c, user.ID, lang)
		items = append(items, SearchResultItem{
			User:  toUserResponseWithProfile(user, profile),
			Score: r.Score,
		})
	}

	c.JSON(http.StatusOK, SearchUsersResponse{
		Users: items,
		Query: query,
		Total: len(items),
	})
}

func (h *SearchHandler) ReindexAll(c *gin.Context) {
	go func() {
		_ = h.vectorIndexSvc.ReindexAll(context.Background())
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": "reindexing started"})
}
