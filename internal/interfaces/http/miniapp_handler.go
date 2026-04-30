package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type MiniAppHandler struct {
	vacancySearchSvc *application.VacancySearchService
	vacancySvc       *application.VacancyService
	vacancyAppSvc    *application.VacancyApplicationService
	userSvc          *application.UserService
	profileFieldSvc  *application.ProfileFieldService
	profileTextSvc   *application.ProfileFieldTextService
	experienceSvc    *application.ExperienceItemService
	educationSvc     *application.EducationItemService
	itemTextSvc      *application.ItemTextService
	skillSvc         *application.SkillService
}

func NewMiniAppHandler(
	vacancySearchSvc *application.VacancySearchService,
	vacancySvc *application.VacancyService,
	vacancyAppSvc *application.VacancyApplicationService,
	userSvc *application.UserService,
	profileFieldSvc *application.ProfileFieldService,
	profileTextSvc *application.ProfileFieldTextService,
	experienceSvc *application.ExperienceItemService,
	educationSvc *application.EducationItemService,
	itemTextSvc *application.ItemTextService,
	skillSvc *application.SkillService,
) *MiniAppHandler {
	return &MiniAppHandler{
		vacancySearchSvc: vacancySearchSvc,
		vacancySvc:       vacancySvc,
		vacancyAppSvc:    vacancyAppSvc,
		userSvc:          userSvc,
		profileFieldSvc:  profileFieldSvc,
		profileTextSvc:   profileTextSvc,
		experienceSvc:    experienceSvc,
		educationSvc:     educationSvc,
		itemTextSvc:      itemTextSvc,
		skillSvc:         skillSvc,
	}
}

// ListForUser returns vacancies matching the authenticated user's profile.
// @Summary List vacancies for user
// @Description Get a list of vacancies that match the worker's profile skills and preferences
// @Tags miniapp
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param lang query string false "Language code"
// @Success 200 {object} PaginatedVacanciesResponse
// @Failure 401 {object} ErrorResponse
// @Router /miniapp/vacancies [get]
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
// @Summary Search vacancies
// @Description Search for vacancies using text query (vector similarity search)
// @Tags miniapp
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param q query string false "Search query"
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param lang query string false "Language code"
// @Success 200 {object} PaginatedVacanciesResponse
// @Failure 401 {object} ErrorResponse
// @Router /miniapp/vacancies/search [get]
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
// @Summary Get vacancy by ID
// @Description Get detailed information about a specific vacancy
// @Tags miniapp
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param id path string true "Vacancy UUID"
// @Param lang query string false "Language code"
// @Success 200 {object} VacancyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /miniapp/vacancies/{id} [get]
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

// GetProfile returns the authenticated user's profile.
// @Summary Get current user profile
// @Description Get detailed profile for the authenticated worker/user from Telegram Mini App
// @Tags miniapp
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param lang query string false "Language (uz, ru, en)"
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /miniapp/me [get]
func (h *MiniAppHandler) GetProfile(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user"})
		return
	}

	user, err := h.userSvc.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		return
	}

	lang := c.DefaultQuery("lang", user.Language)
	if lang == "" {
		lang = "en"
	}

	profile := h.buildUserProfile(c, userID, lang)
	c.JSON(http.StatusOK, toUserResponseWithProfile(user, profile))
}

func (h *MiniAppHandler) buildUserProfile(c *gin.Context, userID uuid.UUID, lang string) *UserProfileResponse {
	ctx := c.Request.Context()

	fields, err := h.profileFieldSvc.ListProfileFieldsByUser(ctx, userID)
	if err != nil {
		fields = nil
	}

	fieldMap := make(map[string]string, len(fields))
	for _, f := range fields {
		text, err := h.profileTextSvc.GetProfileFieldText(ctx, f.ID, lang)
		if err != nil {
			continue
		}
		fieldMap[f.FieldName] = text.Content
	}

	expItems, err := h.experienceSvc.ListExperienceItemsByUser(ctx, userID)
	if err != nil {
		expItems = nil
	}

	var experience []ExperienceItemResponse
	for _, item := range expItems {
		var projects []ProjectResponse
		if item.Projects != "" {
			var rawProjects []struct {
				Project string              `json:"project"`
				Items   map[string][]string `json:"items"`
			}
			if json.Unmarshal([]byte(item.Projects), &rawProjects) == nil {
				for _, p := range rawProjects {
					items := p.Items[lang]
					if len(items) == 0 {
						items = p.Items["en"]
					}
					projects = append(projects, ProjectResponse{
						Project: p.Project,
						Items:   items,
					})
				}
			}
		}

		resp := ExperienceItemResponse{
			ID:        item.ID.String(),
			Company:   item.Company,
			Position:  item.Position,
			StartDate: item.StartDate,
			EndDate:   item.EndDate,
			Projects:  projects,
			WebSite:   item.WebSite,
		}
		texts, err := h.itemTextSvc.ListItemTextsByItem(ctx, item.ID, "experience")
		if err == nil {
			for _, t := range texts {
				if t.Lang == lang {
					resp.Description = t.Description
					break
				}
			}
		}
		experience = append(experience, resp)
	}

	eduItems, err := h.educationSvc.ListEducationItemsByUser(ctx, userID)
	if err != nil {
		eduItems = nil
	}

	var education []EducationItemResponse
	for _, item := range eduItems {
		resp := EducationItemResponse{
			ID:           item.ID.String(),
			Institution:  item.Institution,
			Degree:       item.Degree,
			FieldOfStudy: item.FieldOfStudy,
			StartDate:    item.StartDate,
			EndDate:      item.EndDate,
			Location:     item.Location,
		}
		texts, err := h.itemTextSvc.ListItemTextsByItem(ctx, item.ID, "education")
		if err == nil {
			for _, t := range texts {
				if t.Lang == lang {
					resp.Description = t.Description
					break
				}
			}
		}
		education = append(education, resp)
	}

	if len(fieldMap) == 0 && len(experience) == 0 && len(education) == 0 {
		return nil
	}

	var skills []string
	if userSkills, err := h.skillSvc.ListUserSkills(ctx, userID); err == nil {
		for _, s := range userSkills {
			skills = append(skills, s.Name)
		}
	}

	var certifications []string
	if raw, ok := fieldMap["certifications"]; ok && raw != "" {
		_ = json.Unmarshal([]byte(raw), &certifications)
	}

	var languages []LanguageItemResponse
	if raw, ok := fieldMap["languages"]; ok && raw != "" {
		_ = json.Unmarshal([]byte(raw), &languages)
	}

	return &UserProfileResponse{
		Title:          fieldMap["title"],
		About:          fieldMap["about"],
		Skills:         skills,
		Certifications: certifications,
		Languages:      languages,
		Achievements:   fieldMap["achievements"],
		Experience:     experience,
		Education:      education,
	}
}

// Apply creates a vacancy application for the authenticated user.
// @Summary Apply for a vacancy
// @Description Submit an application for a specific vacancy
// @Tags miniapp
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param id path string true "Vacancy UUID"
// @Param request body object true "Application details"
// @Success 201 {object} VacancyApplicationResponseDTO
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /miniapp/vacancies/{id}/apply [post]
func (h *MiniAppHandler) Apply(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user"})
		return
	}

	vacancyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid vacancy id"})
		return
	}

	var req struct {
		CoverLetter string `json:"cover_letter"`
	}
	_ = c.ShouldBindJSON(&req)

	va, err := h.vacancyAppSvc.Apply(c.Request.Context(), userID, vacancyID, req.CoverLetter)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyApplied) {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "already applied"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to apply"})
		return
	}

	c.JSON(http.StatusCreated, vacancyApplicationResponse(va))
}

// GetApplicationStatus checks if the authenticated user has applied to a vacancy.
// @Summary Check application status
// @Description Check if the user has already applied for this vacancy and get status
// @Tags miniapp
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param id path string true "Vacancy UUID"
// @Success 200 {object} VacancyApplicationResponseDTO
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /miniapp/vacancies/{id}/application [get]
func (h *MiniAppHandler) GetApplicationStatus(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user"})
		return
	}

	vacancyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid vacancy id"})
		return
	}

	va, err := h.vacancyAppSvc.GetByUserAndVacancy(c.Request.Context(), userID, vacancyID)
	if err != nil {
		if errors.Is(err, domain.ErrVacancyApplicationNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not applied"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check application"})
		return
	}

	c.JSON(http.StatusOK, vacancyApplicationResponse(va))
}

// ListMyApplications returns the authenticated user's vacancy applications.
// @Summary List my applications
// @Description Get a list of all vacancies the current user has applied for
// @Tags miniapp
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} object
// @Failure 401 {object} ErrorResponse
// @Router /miniapp/applications [get]
func (h *MiniAppHandler) ListMyApplications(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user"})
		return
	}

	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)
	offset := (page - 1) * pageSize

	apps, err := h.vacancyAppSvc.ListByUser(c.Request.Context(), userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list applications"})
		return
	}

	total, _ := h.vacancyAppSvc.CountByUser(c.Request.Context(), userID)

	resp := make([]VacancyApplicationResponseDTO, 0, len(apps))
	for _, a := range apps {
		resp = append(resp, VacancyApplicationResponse(&a))
	}

	c.JSON(http.StatusOK, gin.H{
		"applications": resp,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
	})
}

type VacancyApplicationResponseDTO struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	VacancyID   string `json:"vacancy_id"`
	Status      string `json:"status"`
	CoverLetter string `json:"cover_letter,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func VacancyApplicationResponse(va *domain.VacancyApplication) VacancyApplicationResponseDTO {
	return VacancyApplicationResponseDTO{
		ID:          va.ID.String(),
		UserID:      va.UserID.String(),
		VacancyID:   va.VacancyID.String(),
		Status:      va.Status,
		CoverLetter: va.CoverLetter,
		CreatedAt:   va.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   va.UpdatedAt.Format(time.RFC3339),
	}
}
