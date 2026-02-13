package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

const (
	accessTokenExpiry  = 15 * time.Minute
	refreshTokenExpiry = 30 * 24 * time.Hour
)

type AuthService struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	jwtSecret   string
}

func NewAuthService(userRepo domain.UserRepository, sessionRepo domain.SessionRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   jwtSecret,
	}
}

type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	DeviceID     string
	ExpiresAt    time.Time
}

func (s *AuthService) SetPassword(ctx context.Context, userID uuid.UUID, password string) error {
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return fmt.Errorf("verify user: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.userRepo.SetPassword(ctx, userID, string(hash)); err != nil {
		return fmt.Errorf("set password: %w", err)
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, login, password, fcmToken, ipAddress string) (*AuthResponse, error) {
	user, err := s.findUserByLogin(ctx, login)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.PasswordHash == "" {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	deviceID := uuid.New().String()

	accessToken, err := s.generateToken(user.ID, deviceID, "access", accessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(user.ID, deviceID, "refresh", refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	_, err = s.sessionRepo.Create(ctx, &domain.UserSession{
		ID:               uuid.New(),
		UserID:           user.ID,
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

func (s *AuthService) Refresh(ctx context.Context, refreshToken, deviceID string) (*AuthResponse, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if claims["type"] != "refresh" {
		return nil, domain.ErrInvalidCredentials
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, domain.ErrInvalidCredentials
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	session, err := s.sessionRepo.GetByDeviceID(ctx, userID, deviceID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if session.RefreshTokenHash != hashToken(refreshToken) {
		return nil, domain.ErrInvalidCredentials
	}

	newAccessToken, err := s.generateToken(userID, deviceID, "access", accessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := s.generateToken(userID, deviceID, "refresh", refreshTokenExpiry)
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

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, deviceID string) error {
	session, err := s.sessionRepo.GetByDeviceID(ctx, userID, deviceID)
	if err != nil {
		return fmt.Errorf("find session: %w", err)
	}
	return s.sessionRepo.SoftDelete(ctx, session.ID)
}

func (s *AuthService) findUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	if strings.Contains(login, "@") {
		return s.userRepo.GetByEmail(ctx, login)
	}
	return s.userRepo.GetByPhone(ctx, login)
}

func (s *AuthService) generateToken(userID uuid.UUID, deviceID, tokenType string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":       userID.String(),
		"device_id": deviceID,
		"type":      tokenType,
		"exp":       time.Now().Add(expiry).Unix(),
		"iat":       time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) parseToken(tokenString string) (jwt.MapClaims, error) {
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

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
