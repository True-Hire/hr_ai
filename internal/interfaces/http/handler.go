package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type UserHandler struct {
	service         *application.UserService
	profileFieldSvc *application.ProfileFieldService
	profileTextSvc  *application.ProfileFieldTextService
	experienceSvc   *application.ExperienceItemService
	educationSvc    *application.EducationItemService
	itemTextSvc     *application.ItemTextService
	skillSvc        *application.SkillService
	authSvc         *application.AuthService
	searchSvc       *application.SearchService
}

func NewUserHandler(
	service *application.UserService,
	profileFieldSvc *application.ProfileFieldService,
	profileTextSvc *application.ProfileFieldTextService,
	experienceSvc *application.ExperienceItemService,
	educationSvc *application.EducationItemService,
	itemTextSvc *application.ItemTextService,
	skillSvc *application.SkillService,
	authSvc *application.AuthService,
	searchSvc *application.SearchService,
) *UserHandler {
	return &UserHandler{
		service:         service,
		profileFieldSvc: profileFieldSvc,
		profileTextSvc:  profileTextSvc,
		experienceSvc:   experienceSvc,
		educationSvc:    educationSvc,
		itemTextSvc:     itemTextSvc,
		skillSvc:        skillSvc,
		authSvc:         authSvc,
		searchSvc:       searchSvc,
	}
}

// Create godoc
// @Summary Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	created, err := h.service.CreateUser(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
		return
	}

	if err := h.authSvc.SetPassword(c.Request.Context(), created.ID, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "user created but failed to set password"})
		return
	}

	c.JSON(http.StatusCreated, toUserResponse(created))
}

// Me godoc
// @Summary Get current authenticated user with full profile
// @Tags users
// @Produce json
// @Param lang query string false "Language code (uz, ru, en)" default(en)
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /users/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get user"})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	profile := h.buildUserProfile(c, user.ID, lang)

	c.JSON(http.StatusOK, toUserResponseWithProfile(user, profile))
}

// GetByID godoc
// @Summary Get user by ID with full profile
// @Tags users
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param lang query string false "Language code (uz, ru, en)" default(en)
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get user"})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	profile := h.buildUserProfile(c, user.ID, lang)

	c.JSON(http.StatusOK, toUserResponseWithProfile(user, profile))
}

// List godoc
// @Summary List users with pagination or semantic search
// @Tags users
// @Produce json
// @Param q query string false "Search query (any language). When provided, returns semantically ranked results"
// @Param page query int false "Page number (pagination mode)" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param lang query string false "Language code (uz, ru, en)" default(en)
// @Success 200 {object} PaginatedUsersResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
	lang := c.DefaultQuery("lang", "en")

	// If search query provided, use semantic search
	if q := c.Query("q"); q != "" {
		h.listBySearch(c, q, lang)
		return
	}

	// Otherwise, regular pagination
	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)

	result, err := h.service.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list users"})
		return
	}

	resp := PaginatedUsersResponse{
		Users:    make([]UserResponse, 0, len(result.Users)),
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, u := range result.Users {
		profile := h.buildUserProfile(c, u.ID, lang)
		resp.Users = append(resp.Users, toUserResponseWithProfile(&u, profile))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) listBySearch(c *gin.Context, query, lang string) {
	pageSize := parseQueryInt32(c, "page_size", 20)

	results, err := h.searchSvc.SearchUsers(c.Request.Context(), query, int(pageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "search failed"})
		return
	}

	users := make([]UserResponse, 0, len(results))
	for _, r := range results {
		user, err := h.service.GetUser(c.Request.Context(), r.UserID)
		if err != nil {
			continue
		}
		profile := h.buildUserProfile(c, user.ID, lang)
		resp := toUserResponseWithProfile(user, profile)
		resp.SearchScore = &r.Score
		users = append(users, resp)
	}

	c.JSON(http.StatusOK, PaginatedUsersResponse{
		Users:    users,
		Total:    int64(len(users)),
		Page:     1,
		PageSize: pageSize,
	})
}

// Update godoc
// @Summary Update a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param request body UpdateUserRequest true "Updated user data"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	updated, err := h.service.UpdateUser(c.Request.Context(), req.ToDomain(id))
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(updated))
}

// Delete godoc
// @Summary Delete a user
// @Tags users
// @Param id path string true "User ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete user"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) buildUserProfile(c *gin.Context, userID uuid.UUID, lang string) *UserProfileResponse {
	ctx := c.Request.Context()

	// Fetch text fields
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

	// Fetch experience items
	expItems, err := h.experienceSvc.ListExperienceItemsByUser(ctx, userID)
	if err != nil {
		expItems = nil
	}

	var experience []ExperienceItemResponse
	for _, item := range expItems {
		// Parse projects from JSON, resolving translated items for requested lang
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

	// Fetch education items
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

	// Fetch skills from skills/user_skills tables
	var skills []string
	if userSkills, err := h.skillSvc.ListUserSkills(ctx, userID); err == nil {
		for _, s := range userSkills {
			skills = append(skills, s.Name)
		}
	}

	// Parse certifications from JSON array
	var certifications []string
	if raw, ok := fieldMap["certifications"]; ok && raw != "" {
		_ = json.Unmarshal([]byte(raw), &certifications)
	}

	// Parse languages from JSON array of objects
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

func parseQueryInt32(c *gin.Context, key string, defaultVal int32) int32 {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return int32(n)
}
