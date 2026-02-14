package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type BotService struct {
	userSvc      *UserService
	profileParse *ProfileParseService
	storage      *StorageService
}

func NewBotService(userSvc *UserService, profileParse *ProfileParseService, storage *StorageService) *BotService {
	return &BotService{userSvc: userSvc, profileParse: profileParse, storage: storage}
}

func (s *BotService) HandleStart(ctx context.Context, telegramID int64, firstName, lastName, username string, photoData []byte) (*domain.User, bool, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	existing, err := s.userSvc.GetByTelegramID(ctx, tgID)
	if err == nil {
		return existing, false, nil
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, false, fmt.Errorf("check existing user: %w", err)
	}

	var telegram string
	if username != "" {
		telegram = "@" + username
	}

	var photoURL string
	if len(photoData) > 0 {
		url, err := s.storage.UploadProfilePhoto(ctx, photoData, "image/jpeg")
		if err == nil {
			photoURL = url
		}
	}

	user := &domain.User{
		FirstName:     firstName,
		LastName:      lastName,
		Telegram:      telegram,
		TelegramID:    tgID,
		ProfilePicURL: photoURL,
	}

	created, err := s.userSvc.CreateUser(ctx, user)
	if err != nil {
		return nil, false, fmt.Errorf("create user from telegram: %w", err)
	}

	return created, true, nil
}

func (s *BotService) HandleResumeText(ctx context.Context, userID uuid.UUID, text string) (*ParseResult, error) {
	return s.profileParse.ParseFromText(ctx, userID, text)
}

func (s *BotService) HandleResumeFile(ctx context.Context, userID uuid.UUID, fileData []byte, mimeType string) (*ParseResult, error) {
	return s.profileParse.ParseFromFile(ctx, userID, fileData, mimeType)
}
