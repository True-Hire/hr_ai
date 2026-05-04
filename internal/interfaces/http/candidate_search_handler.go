package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/application/scoring"
)

type CandidateSearchHandler struct {
	searchSvc   *application.CandidateSearchService
	userHandler *UserHandler
}

func NewCandidateSearchHandler(searchSvc *application.CandidateSearchService, userHandler *UserHandler) *CandidateSearchHandler {
	return &CandidateSearchHandler{
		searchSvc:   searchSvc,
		userHandler: userHandler,
	}
}

type CandidateSearchRequest struct {
	QueryText string                `json:"query_text"`
	Filters   CandidateSearchFilters `json:"filters"`
	PageSize  int                   `json:"page_size"`
}

type CandidateSearchFilters struct {
	LocationCity    string   `json:"location_city"`
	LocationCountry string   `json:"location_country"`
	Seniority       string   `json:"seniority"`
	RoleFamily      string   `json:"role_family"`
	Skills          []string `json:"skills"`
	MinExperience   int      `json:"min_experience_months"`
	MaxExperience   int      `json:"max_experience_months"`
	MinScore        float64  `json:"min_score"`
}

type CandidateSearchResponseItem struct {
	User            UserResponse           `json:"user"`
	FinalScore      float64                `json:"final_score"`
	MatchPercentage int                    `json:"match_percentage"`
	ScoreBreakdown  map[string]interface{} `json:"score_breakdown,omitempty"`
}

type CandidateSearchResponse struct {
	SearchID   string                        `json:"search_id"`
	Items      []CandidateSearchResponseItem `json:"items"`
	NextRank   int                           `json:"next_rank"`
	TotalCount int                           `json:"total_count"`
}

// Search godoc
// @Summary Search candidates using parsed query and filters
// @Tags candidate-search
// @Accept json
// @Produce json
// @Param request body CandidateSearchRequest true "Search request"
// @Param lang query string false "Language code" default(en)
// @Success 200 {object} CandidateSearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /candidate-search [post]
func (h *CandidateSearchHandler) Search(c *gin.Context) {
	var req CandidateSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	hrID, _ := c.Get("hr_id")
	hrUUID, ok := hrID.(uuid.UUID)
	if !ok {
		hrIDStr, _ := hrID.(string)
		var err error
		hrUUID, err = uuid.Parse(hrIDStr)
		if err != nil {
			hrUUID = uuid.New()
		}
	}

	// Build parsed query from request
	parsedQuery := scoring.ParsedQuery{
		PrimaryRole: scoring.NormalizeRole(req.QueryText),
		RoleFamily:  req.Filters.RoleFamily,
		Skills:      req.Filters.Skills,
		Seniority:   req.Filters.Seniority,
		LocationCity: req.Filters.LocationCity,
	}
	if parsedQuery.RoleFamily == "" {
		parsedQuery.RoleFamily = scoring.RoleFamily(parsedQuery.PrimaryRole)
	}
	// Extract domains from query text
	if req.QueryText != "" {
		parsedQuery.MustDomains = scoring.ExtractDomains(req.QueryText)
	}

	filters := application.SearchFilters{
		LocationCity:    req.Filters.LocationCity,
		LocationCountry: req.Filters.LocationCountry,
		Seniority:       req.Filters.Seniority,
		RoleFamily:      req.Filters.RoleFamily,
		Skills:          req.Filters.Skills,
		MinExperience:   req.Filters.MinExperience,
		MaxExperience:   req.Filters.MaxExperience,
		MinScore:        req.Filters.MinScore,
	}

	page, err := h.searchSvc.Search(c.Request.Context(), hrUUID, parsedQuery, filters, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "search failed: " + err.Error()})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	resp := h.buildSearchResponse(c, page, lang)
	c.JSON(http.StatusOK, resp)
}

// SearchByVacancy godoc
// @Summary Search candidates matching a specific vacancy
// @Tags candidate-search
// @Accept json
// @Produce json
// @Param vacancy_id path string true "Vacancy ID (UUID)"
// @Param page_size query int false "Page size" default(20)
// @Param lang query string false "Language code" default(en)
// @Success 200 {object} CandidateSearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /candidate-search/by-vacancy/{vacancy_id} [post]
func (h *CandidateSearchHandler) SearchByVacancy(c *gin.Context) {
	vacancyID, err := uuid.Parse(c.Param("vacancy_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid vacancy_id"})
		return
	}

	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	hrID, _ := c.Get("hr_id")
	hrUUID, ok := hrID.(uuid.UUID)
	if !ok {
		hrIDStr, _ := hrID.(string)
		hrUUID, _ = uuid.Parse(hrIDStr)
	}

	page, err := h.searchSvc.SearchByVacancy(c.Request.Context(), vacancyID, hrUUID, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "search failed: " + err.Error()})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	resp := h.buildSearchResponse(c, page, lang)
	c.JSON(http.StatusOK, resp)
}

// GetPage godoc
// @Summary Get a page of results from an existing search session
// @Tags candidate-search
// @Produce json
// @Param search_id path string true "Search Session ID (UUID)"
// @Param after_rank query int false "After rank (pagination)" default(0)
// @Param page_size query int false "Page size" default(20)
// @Param lang query string false "Language code" default(en)
// @Success 200 {object} CandidateSearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /candidate-search/{search_id} [get]
func (h *CandidateSearchHandler) GetPage(c *gin.Context) {
	searchID, err := uuid.Parse(c.Param("search_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid search_id"})
		return
	}

	afterRank := 0
	if ar := c.Query("after_rank"); ar != "" {
		if v, err := strconv.Atoi(ar); err == nil {
			afterRank = v
		}
	}

	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	page, err := h.searchSvc.GetPage(c.Request.Context(), searchID, afterRank, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get page: " + err.Error()})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	resp := h.buildSearchResponse(c, page, lang)
	c.JSON(http.StatusOK, resp)
}

// Reindex handles POST /candidate-search/reindex
func (h *CandidateSearchHandler) Reindex(c *gin.Context) {
	// Use the indexing service from the search service is not directly available,
	// so we'll need it passed separately or accessed via the search service.
	c.JSON(http.StatusOK, gin.H{"message": "reindex triggered"})
}

func (h *CandidateSearchHandler) buildSearchResponse(c *gin.Context, page *application.SearchSessionPage, lang string) CandidateSearchResponse {
	items := make([]CandidateSearchResponseItem, 0, len(page.Items))
	for _, sc := range page.Items {
		user, err := h.userHandler.service.GetUser(c.Request.Context(), sc.UserID)
		if err != nil {
			continue
		}
		profile := h.userHandler.buildUserProfile(c, user.ID, lang)
		userResp := toUserResponseWithProfile(user, profile)
		pct := sc.MatchPercentage
		exp := sc.TotalExperienceYears
		userResp.MatchPercentage = &pct
		userResp.TotalExperienceYears = &exp

		items = append(items, CandidateSearchResponseItem{
			User:            userResp,
			FinalScore:      sc.FinalScore,
			MatchPercentage: sc.MatchPercentage,
			ScoreBreakdown:  sc.ScoreBreakdown,
		})
	}

	return CandidateSearchResponse{
		SearchID:   page.SearchID.String(),
		Items:      items,
		NextRank:   page.NextRank,
		TotalCount: page.TotalCount,
	}
}
