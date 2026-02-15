package http

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

type MiniAppHandler struct {
	vacancySearchSvc *application.VacancySearchService
	vacancySvc       *application.VacancyService
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

// GetProfile returns the authenticated user's profile.
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
