package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type HRVacancyApplicationsHandler struct {
	vacancyAppSvc *application.VacancyApplicationService
	vacancySvc    *application.VacancyService
	userSvc       *application.UserService
	skillSvc      *application.SkillService
}

func NewHRVacancyApplicationsHandler(
	vacancyAppSvc *application.VacancyApplicationService,
	vacancySvc *application.VacancyService,
	userSvc *application.UserService,
	skillSvc *application.SkillService,
) *HRVacancyApplicationsHandler {
	return &HRVacancyApplicationsHandler{
		vacancyAppSvc: vacancyAppSvc,
		vacancySvc:    vacancySvc,
		userSvc:       userSvc,
		skillSvc:      skillSvc,
	}
}

// validateVacancyOwnership checks that the vacancy exists and belongs to the HR.
func (h *HRVacancyApplicationsHandler) validateVacancyOwnership(c *gin.Context) (uuid.UUID, bool) {
	hrID, err := uuid.Parse(c.GetString("hr_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return uuid.Nil, false
	}

	vacancyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid vacancy id"})
		return uuid.Nil, false
	}

	vwd, err := h.vacancySvc.GetVacancy(c.Request.Context(), vacancyID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "vacancy not found"})
		return uuid.Nil, false
	}

	if vwd.Vacancy.HRID != hrID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "vacancy does not belong to this HR"})
		return uuid.Nil, false
	}

	return vacancyID, true
}

// ListApplicants godoc
// @Summary List applicants for a vacancy
// @Tags hr-miniapp
// @Produce json
// @Param id path string true "Vacancy ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} HRApplicationListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security TelegramAuth
// @Router /hr-miniapp/vacancies/{id}/applications [get]
func (h *HRVacancyApplicationsHandler) ListApplicants(c *gin.Context) {
	vacancyID, ok := h.validateVacancyOwnership(c)
	if !ok {
		return
	}

	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 10)
	offset := (page - 1) * pageSize

	ctx := c.Request.Context()

	apps, err := h.vacancyAppSvc.ListByVacancy(ctx, vacancyID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list applications"})
		return
	}

	total, _ := h.vacancyAppSvc.CountByVacancy(ctx, vacancyID)
	unseen, _ := h.vacancyAppSvc.CountUnseenByVacancy(ctx, vacancyID)

	resp := HRApplicationListResponse{
		Applications: make([]HRApplicantResponse, 0, len(apps)),
		Total:        total,
		Seen:         total - unseen,
		Unseen:       unseen,
		Page:         page,
		PageSize:     pageSize,
	}

	for _, app := range apps {
		ar := toHRApplicantResponse(&app)

		// Enrich with user data
		user, err := h.userSvc.GetUser(ctx, app.UserID)
		if err == nil {
			userResp := &HRApplicantUserResponse{
				ID:        user.ID.String(),
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Phone:     user.Phone,
				Telegram:  user.Telegram,
			}
			if skills, err := h.skillSvc.ListUserSkills(ctx, app.UserID); err == nil {
				for _, s := range skills {
					userResp.Skills = append(userResp.Skills, s.Name)
				}
			}
			ar.User = userResp
		}

		resp.Applications = append(resp.Applications, ar)
	}

	c.JSON(http.StatusOK, resp)
}

// GetStats godoc
// @Summary Get application statistics for a vacancy
// @Tags hr-miniapp
// @Produce json
// @Param id path string true "Vacancy ID"
// @Success 200 {object} HRApplicationStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security TelegramAuth
// @Router /hr-miniapp/vacancies/{id}/applications/stats [get]
func (h *HRVacancyApplicationsHandler) GetStats(c *gin.Context) {
	vacancyID, ok := h.validateVacancyOwnership(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()

	total, err := h.vacancyAppSvc.CountByVacancy(ctx, vacancyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to count applications"})
		return
	}

	unseen, err := h.vacancyAppSvc.CountUnseenByVacancy(ctx, vacancyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to count unseen applications"})
		return
	}

	c.JSON(http.StatusOK, HRApplicationStatsResponse{
		Total:  total,
		Seen:   total - unseen,
		Unseen: unseen,
	})
}

// UpdateStatus godoc
// @Summary Update application status (accepted/rejected/pending)
// @Tags hr-miniapp
// @Accept json
// @Produce json
// @Param id path string true "Vacancy ID"
// @Param app_id path string true "Application ID"
// @Param request body HRUpdateApplicationStatusRequest true "New status"
// @Success 200 {object} HRApplicantResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security TelegramAuth
// @Router /hr-miniapp/vacancies/{id}/applications/{app_id}/status [put]
func (h *HRVacancyApplicationsHandler) UpdateStatus(c *gin.Context) {
	_, ok := h.validateVacancyOwnership(c)
	if !ok {
		return
	}

	appID, err := uuid.Parse(c.Param("app_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid application id"})
		return
	}

	var req HRUpdateApplicationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status is required"})
		return
	}

	if req.Status != "accepted" && req.Status != "rejected" && req.Status != "pending" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status must be accepted, rejected, or pending"})
		return
	}

	updated, err := h.vacancyAppSvc.UpdateStatus(c.Request.Context(), appID, req.Status)
	if err != nil {
		if err == domain.ErrVacancyApplicationNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update status"})
		return
	}

	c.JSON(http.StatusOK, toHRApplicantResponse(updated))
}

// MarkSeen godoc
// @Summary Mark application as seen
// @Tags hr-miniapp
// @Produce json
// @Param id path string true "Vacancy ID"
// @Param app_id path string true "Application ID"
// @Success 200 {object} HRApplicantResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security TelegramAuth
// @Router /hr-miniapp/vacancies/{id}/applications/{app_id}/seen [put]
func (h *HRVacancyApplicationsHandler) MarkSeen(c *gin.Context) {
	_, ok := h.validateVacancyOwnership(c)
	if !ok {
		return
	}

	appID, err := uuid.Parse(c.Param("app_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid application id"})
		return
	}

	updated, err := h.vacancyAppSvc.MarkSeen(c.Request.Context(), appID)
	if err != nil {
		if err == domain.ErrVacancyApplicationNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to mark seen"})
		return
	}

	c.JSON(http.StatusOK, toHRApplicantResponse(updated))
}

func toHRApplicantResponse(va *domain.VacancyApplication) HRApplicantResponse {
	r := HRApplicantResponse{
		ID:          va.ID.String(),
		UserID:      va.UserID.String(),
		VacancyID:   va.VacancyID.String(),
		Status:      va.Status,
		CoverLetter: va.CoverLetter,
		CreatedAt:   va.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   va.UpdatedAt.Format(time.RFC3339),
	}
	if va.SeenAt != nil {
		s := va.SeenAt.Format(time.RFC3339)
		r.SeenAt = &s
	}
	return r
}
