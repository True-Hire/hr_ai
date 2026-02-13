package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type HRAuthHandler struct {
	service *application.HRAuthService
}

func NewHRAuthHandler(service *application.HRAuthService) *HRAuthHandler {
	return &HRAuthHandler{service: service}
}

// SetPassword godoc
// @Summary Set password for a company HR
// @Tags hr-auth
// @Accept json
// @Produce json
// @Param request body SetPasswordRequest true "HR ID and password"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hr/auth/set-password [post]
func (h *HRAuthHandler) SetPassword(c *gin.Context) {
	var req SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid hr id"})
		return
	}

	if err := h.service.SetPassword(c.Request.Context(), userID, req.Password); err != nil {
		if errors.Is(err, domain.ErrCompanyHRNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "company hr not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to set password"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "password set successfully"})
}

// Login godoc
// @Summary Login as company HR
// @Tags hr-auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /hr/auth/login [post]
func (h *HRAuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req.Login, req.Password, req.FcmToken, c.ClientIP())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "login failed"})
		return
	}

	c.JSON(http.StatusOK, toAuthTokenResponse(resp))
}

// Refresh godoc
// @Summary Refresh HR access token
// @Tags hr-auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token and device ID"
// @Success 200 {object} AuthTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /hr/auth/refresh [post]
func (h *HRAuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.service.Refresh(c.Request.Context(), req.RefreshToken, req.DeviceID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "refresh failed"})
		return
	}

	c.JSON(http.StatusOK, toAuthTokenResponse(resp))
}

// Logout godoc
// @Summary Logout HR (invalidate session)
// @Tags hr-auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Device ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /hr/auth/logout [post]
func (h *HRAuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	hrID, _ := c.Get("hr_id")
	id, err := uuid.Parse(hrID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	if err := h.service.Logout(c.Request.Context(), id, req.DeviceID); err != nil {
		if errors.Is(err, domain.ErrHRSessionNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "logout failed"})
		return
	}

	c.Status(http.StatusNoContent)
}
