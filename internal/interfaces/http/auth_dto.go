package http

import (
	"time"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

type SetPasswordRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	FcmToken string `json:"fcm_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	DeviceID     string `json:"device_id" binding:"required"`
}

type LogoutRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
	ExpiresAt    string `json:"expires_at"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func toAuthTokenResponse(r *application.AuthResponse) AuthTokenResponse {
	return AuthTokenResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		DeviceID:     r.DeviceID,
		ExpiresAt:    r.ExpiresAt.Format(time.RFC3339),
	}
}
