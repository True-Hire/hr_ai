package http

import (
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
}

func NewUserHandler(
	service *application.UserService,
	profileFieldSvc *application.ProfileFieldService,
	profileTextSvc *application.ProfileFieldTextService,
	experienceSvc *application.ExperienceItemService,
	educationSvc *application.EducationItemService,
	itemTextSvc *application.ItemTextService,
) *UserHandler {
	return &UserHandler{
		service:         service,
		profileFieldSvc: profileFieldSvc,
		profileTextSvc:  profileTextSvc,
		experienceSvc:   experienceSvc,
		educationSvc:    educationSvc,
		itemTextSvc:     itemTextSvc,
	}
}

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

	c.JSON(http.StatusCreated, toUserResponse(created))
}

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

func (h *UserHandler) List(c *gin.Context) {
	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)

	result, err := h.service.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list users"})
		return
	}

	lang := c.DefaultQuery("lang", "en")

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
		resp := ExperienceItemResponse{
			ID:        item.ID.String(),
			Company:   item.Company,
			Position:  item.Position,
			StartDate: item.StartDate,
			EndDate:   item.EndDate,
			Projects:  item.Projects,
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

	return &UserProfileResponse{
		Title:          fieldMap["title"],
		About:          fieldMap["about"],
		Skills:         fieldMap["skills"],
		Languages:      fieldMap["languages"],
		Certifications: fieldMap["certifications"],
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
