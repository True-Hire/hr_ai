package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type HRSavedUsersHandler struct {
	savedUserSvc *application.HRSavedUserService
	userSvc      *application.UserService
	skillSvc     *application.SkillService
}

func NewHRSavedUsersHandler(
	savedUserSvc *application.HRSavedUserService,
	userSvc *application.UserService,
	skillSvc *application.SkillService,
) *HRSavedUsersHandler {
	return &HRSavedUsersHandler{
		savedUserSvc: savedUserSvc,
		userSvc:      userSvc,
		skillSvc:     skillSvc,
	}
}

type SaveUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Note   string `json:"note"`
}

type SavedUserResponse struct {
	UserID    string                   `json:"user_id"`
	Note      string                   `json:"note,omitempty"`
	SavedAt   string                   `json:"saved_at"`
	User      *HRApplicantUserResponse `json:"user,omitempty"`
}

type SavedUserListResponse struct {
	SavedUsers []SavedUserResponse `json:"saved_users"`
	Total      int64               `json:"total"`
	Page       int32               `json:"page"`
	PageSize   int32               `json:"page_size"`
}

// Save godoc
// @Summary Save a user to HR's saved list
// @Tags hr-saved-users
// @Accept json
// @Produce json
// @Param request body SaveUserRequest true "User to save"
// @Success 201 {object} SavedUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security TelegramAuth
// @Router /hr/saved-users [post]
func (h *HRSavedUsersHandler) Save(c *gin.Context) {
	hrID, err := uuid.Parse(c.GetString("hr_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	var req SaveUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user_id is required"})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
		return
	}

	// Verify user exists
	if _, err := h.userSvc.GetUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		return
	}

	saved, err := h.savedUserSvc.Save(c.Request.Context(), hrID, userID, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save user"})
		return
	}

	c.JSON(http.StatusCreated, toSavedUserResponse(saved))
}

// List godoc
// @Summary List saved users with optional filtering
// @Tags hr-saved-users
// @Produce json
// @Param q query string false "Search by name"
// @Param skills query string false "Filter by skills (comma-separated)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} SavedUserListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security TelegramAuth
// @Router /hr/saved-users [get]
func (h *HRSavedUsersHandler) List(c *gin.Context) {
	hrID, err := uuid.Parse(c.GetString("hr_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)
	offset := (page - 1) * pageSize
	nameQuery := c.Query("q")

	var skills []string
	if s := c.Query("skills"); s != "" {
		for _, sk := range strings.Split(s, ",") {
			sk = strings.TrimSpace(sk)
			if sk != "" {
				skills = append(skills, sk)
			}
		}
	}

	ctx := c.Request.Context()

	result, err := h.savedUserSvc.ListFiltered(ctx, hrID, nameQuery, skills, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list saved users"})
		return
	}

	resp := SavedUserListResponse{
		SavedUsers: make([]SavedUserResponse, 0, len(result.SavedUsers)),
		Total:      result.Total,
		Page:       page,
		PageSize:   pageSize,
	}

	for _, su := range result.SavedUsers {
		sr := toSavedUserResponse(&su)
		// Enrich with user data
		if user, err := h.userSvc.GetUser(ctx, su.UserID); err == nil {
			userResp := &HRApplicantUserResponse{
				ID:        user.ID.String(),
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Phone:     user.Phone,
				Telegram:  user.Telegram,
			}
			if userSkills, err := h.skillSvc.ListUserSkills(ctx, su.UserID); err == nil {
				for _, s := range userSkills {
					userResp.Skills = append(userResp.Skills, s.Name)
				}
			}
			sr.User = userResp
		}
		resp.SavedUsers = append(resp.SavedUsers, sr)
	}

	c.JSON(http.StatusOK, resp)
}

// Delete godoc
// @Summary Remove a user from HR's saved list
// @Tags hr-saved-users
// @Param user_id path string true "User ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security TelegramAuth
// @Router /hr/saved-users/{user_id} [delete]
func (h *HRSavedUsersHandler) Delete(c *gin.Context) {
	hrID, err := uuid.Parse(c.GetString("hr_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
		return
	}

	if err := h.savedUserSvc.Unsave(c.Request.Context(), hrID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to unsave user"})
		return
	}

	c.Status(http.StatusNoContent)
}

func toSavedUserResponse(su *domain.HRSavedUser) SavedUserResponse {
	return SavedUserResponse{
		UserID:  su.UserID.String(),
		Note:    su.Note,
		SavedAt: su.CreatedAt.Format(time.RFC3339),
	}
}
