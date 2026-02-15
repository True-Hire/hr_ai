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
	hrSvc        *CompanyHRService
	profileParse *ProfileParseService
	storage      *StorageService
	stateSvc     *BotStateService
}

func NewBotService(userSvc *UserService, hrSvc *CompanyHRService, profileParse *ProfileParseService, storage *StorageService, stateSvc *BotStateService) *BotService {
	return &BotService{
		userSvc:      userSvc,
		hrSvc:        hrSvc,
		profileParse: profileParse,
		storage:      storage,
		stateSvc:     stateSvc,
	}
}

type StartResult struct {
	User  *domain.User
	HR    *domain.CompanyHR
	IsNew bool
	IsHR  bool
}

func (s *BotService) HandleStart(ctx context.Context, telegramID int64) (*StartResult, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	existingUser, err := s.userSvc.GetByTelegramID(ctx, tgID)
	if err == nil {
		return &StartResult{User: existingUser, IsNew: false, IsHR: false}, nil
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}

	existingHR, err := s.hrSvc.GetByTelegramID(ctx, tgID)
	if err == nil {
		return &StartResult{HR: existingHR, IsNew: false, IsHR: true}, nil
	}
	if !errors.Is(err, domain.ErrCompanyHRNotFound) {
		return nil, fmt.Errorf("check existing hr: %w", err)
	}

	if err := s.stateSvc.SetState(ctx, tgID, domain.BotStateChoosingLanguage); err != nil {
		return nil, fmt.Errorf("set language selection state: %w", err)
	}

	return &StartResult{IsNew: true}, nil
}

func (s *BotService) HandleLanguageSelection(ctx context.Context, telegramID int64, language, firstName, lastName, username string, photoData []byte) (*domain.User, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	if language != "en" && language != "ru" && language != "uz" {
		return nil, fmt.Errorf("invalid language: %s", language)
	}

	if err := s.stateSvc.ClearState(ctx, tgID); err != nil {
		return nil, fmt.Errorf("clear state: %w", err)
	}

	var telegram string
	if username != "" {
		telegram = "@" + username
	}

	var photoURL string
	if len(photoData) > 0 {
		result, err := s.storage.UploadProfilePhoto(ctx, photoData, "image/jpeg")
		if err == nil {
			photoURL = result.URL
		}
	}

	user := &domain.User{
		FirstName:     firstName,
		LastName:      lastName,
		Telegram:      telegram,
		TelegramID:    tgID,
		ProfilePicURL: photoURL,
		Language:      language,
	}

	created, err := s.userSvc.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}

func (s *BotService) GetBotState(ctx context.Context, telegramID int64) (*domain.BotState, error) {
	tgID := strconv.FormatInt(telegramID, 10)
	return s.stateSvc.GetState(ctx, tgID)
}

func (s *BotService) HandleResumeText(ctx context.Context, userID uuid.UUID, text string) (*ParseResult, error) {
	return s.profileParse.ParseFromText(ctx, userID, text)
}

func (s *BotService) HandleResumeFile(ctx context.Context, userID uuid.UUID, fileData []byte, mimeType string) (*ParseResult, error) {
	return s.profileParse.ParseFromFile(ctx, userID, fileData, mimeType)
}
