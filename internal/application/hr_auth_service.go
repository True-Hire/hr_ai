package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type HRAuthService struct {
	hrRepo      domain.CompanyHRRepository
	sessionRepo domain.HRSessionRepository
	jwtSecret   string
}

func NewHRAuthService(hrRepo domain.CompanyHRRepository, sessionRepo domain.HRSessionRepository, jwtSecret string) *HRAuthService {
	return &HRAuthService{
		hrRepo:      hrRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   jwtSecret,
	}
}

func (s *HRAuthService) SetPassword(ctx context.Context, hrID uuid.UUID, password string) error {
	if _, err := s.hrRepo.GetByID(ctx, hrID); err != nil {
		return fmt.Errorf("verify hr: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.hrRepo.SetPassword(ctx, hrID, string(hash)); err != nil {
		return fmt.Errorf("set password: %w", err)
	}
	return nil
}

func (s *HRAuthService) Login(ctx context.Context, login, password, fcmToken, ipAddress string) (*AuthResponse, error) {
	hr, err := s.findHRByLogin(ctx, login)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if hr.PasswordHash == "" {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hr.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	deviceID := uuid.New().String()

	accessToken, err := s.generateToken(hr.ID, deviceID, "access", accessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(hr.ID, deviceID, "refresh", refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	_, err = s.sessionRepo.Create(ctx, &domain.HRSession{
		ID:               uuid.New(),
		HRID:             hr.ID,
		DeviceID:         deviceID,
		RefreshTokenHash: hashToken(refreshToken),
		FcmToken:         fcmToken,
		IPAddress:        ipAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		DeviceID:     deviceID,
		ExpiresAt:    time.Now().Add(accessTokenExpiry),
	}, nil
}

func (s *HRAuthService) Refresh(ctx context.Context, refreshToken, deviceID string) (*AuthResponse, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if claims["type"] != "refresh" {
		return nil, domain.ErrInvalidCredentials
	}

	hrIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, domain.ErrInvalidCredentials
	}

	hrID, err := uuid.Parse(hrIDStr)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	session, err := s.sessionRepo.GetByDeviceID(ctx, hrID, deviceID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if session.RefreshTokenHash != hashToken(refreshToken) {
		return nil, domain.ErrInvalidCredentials
	}

	newAccessToken, err := s.generateToken(hrID, deviceID, "access", accessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := s.generateToken(hrID, deviceID, "refresh", refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.sessionRepo.UpdateRefreshToken(ctx, session.ID, hashToken(newRefreshToken)); err != nil {
		return nil, fmt.Errorf("update refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		DeviceID:     deviceID,
		ExpiresAt:    time.Now().Add(accessTokenExpiry),
	}, nil
}

func (s *HRAuthService) Logout(ctx context.Context, hrID uuid.UUID, deviceID string) error {
	session, err := s.sessionRepo.GetByDeviceID(ctx, hrID, deviceID)
	if err != nil {
		return fmt.Errorf("find session: %w", err)
	}
	return s.sessionRepo.SoftDelete(ctx, session.ID)
}

func (s *HRAuthService) findHRByLogin(ctx context.Context, login string) (*domain.CompanyHR, error) {
	if strings.Contains(login, "@") {
		return s.hrRepo.GetByEmail(ctx, login)
	}
	return s.hrRepo.GetByPhone(ctx, login)
}

func (s *HRAuthService) generateToken(hrID uuid.UUID, deviceID, tokenType string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":       hrID.String(),
		"device_id": deviceID,
		"type":      tokenType,
		"role":      "hr",
		"exp":       time.Now().Add(expiry).Unix(),
		"iat":       time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *HRAuthService) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
