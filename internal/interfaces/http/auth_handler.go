package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type AuthHandler struct {
	service *application.AuthService
}

func NewAuthHandler(service *application.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// SetPassword godoc
// @Summary Set password for an existing user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SetPasswordRequest true "User ID and password"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/set-password [post]
func (h *AuthHandler) SetPassword(c *gin.Context) {
	var req SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	if err := h.service.SetPassword(c.Request.Context(), userID, req.Password); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to set password"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "password set"})
}

// Login godoc
// @Summary Login with phone or email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.Login(c.Request.Context(), req.Login, req.Password, req.FcmToken, c.ClientIP())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to login"})
		return
	}

	c.JSON(http.StatusOK, toAuthTokenResponse(result))
}

// Refresh godoc
// @Summary Get new access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token and device ID"
// @Success 200 {object} AuthTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.Refresh(c.Request.Context(), req.RefreshToken, req.DeviceID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired refresh token"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, toAuthTokenResponse(result))
}

// Logout godoc
// @Summary Logout and invalidate session
// @Tags auth
// @Accept json
// @Param Authorization header string true "Bearer access token"
// @Param request body LogoutRequest true "Device ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id in token"})
		return
	}

	if err := h.service.Logout(c.Request.Context(), userID, req.DeviceID); err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to logout"})
		return
	}

	c.Status(http.StatusNoContent)
}
