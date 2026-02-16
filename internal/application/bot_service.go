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

// HandleLanguageSelection creates the user and transitions to sharing_phone state.
func (s *BotService) HandleLanguageSelection(ctx context.Context, telegramID int64, language, firstName, lastName, username string, photoData []byte) (string, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	if language != "en" && language != "ru" && language != "uz" {
		return "", fmt.Errorf("invalid language: %s", language)
	}

	var tg string
	if username != "" {
		tg = "@" + username
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
		Telegram:      tg,
		TelegramID:    tgID,
		ProfilePicURL: photoURL,
		Language:      language,
	}
	if _, err := s.userSvc.CreateUser(ctx, user); err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	// Transition to sharing_phone state
	data := map[string]string{"language": language}
	if err := s.stateSvc.SetStateWithData(ctx, tgID, domain.BotStateSharingPhone, data); err != nil {
		return "", fmt.Errorf("set sharing phone state: %w", err)
	}

	return language, nil
}

// HandlePhoneShared updates the phone number on the user record and transitions to choosing_role state.
func (s *BotService) HandlePhoneShared(ctx context.Context, telegramID int64, phone string) (string, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	state, err := s.stateSvc.GetState(ctx, tgID)
	if err != nil || state == nil {
		return "", fmt.Errorf("no active state for phone sharing")
	}

	language := state.Data["language"]
	if language == "" {
		language = "en"
	}

	user, err := s.userSvc.GetByTelegramID(ctx, tgID)
	if err != nil {
		return language, fmt.Errorf("get user for phone update: %w", err)
	}
	user.Phone = phone
	if _, err := s.userSvc.UpdateUser(ctx, user); err != nil {
		return language, fmt.Errorf("update user phone: %w", err)
	}

	// Transition to choosing_role (goal selection)
	data := map[string]string{"language": language}
	if err := s.stateSvc.SetStateWithData(ctx, tgID, domain.BotStateChoosingRole, data); err != nil {
		return language, fmt.Errorf("set choosing role state: %w", err)
	}

	return language, nil
}

func (s *BotService) GetBotState(ctx context.Context, telegramID int64) (*domain.BotState, error) {
	tgID := strconv.FormatInt(telegramID, 10)
	return s.stateSvc.GetState(ctx, tgID)
}


// HandleGoalSelection clears the bot state after the user picks a goal.
func (s *BotService) HandleGoalSelection(ctx context.Context, telegramID int64) (string, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	state, err := s.stateSvc.GetState(ctx, tgID)
	if err != nil || state == nil {
		return "", fmt.Errorf("no active state for goal selection")
	}

	language := state.Data["language"]
	if language == "" {
		language = "en"
	}

	if err := s.stateSvc.ClearState(ctx, tgID); err != nil {
		return language, fmt.Errorf("clear state: %w", err)
	}

	return language, nil
}

func (s *BotService) HandleResumeText(ctx context.Context, userID uuid.UUID, text string) (*ParseResult, error) {
	return s.profileParse.ParseFromText(ctx, userID, text)
}

func (s *BotService) HandleResumeFile(ctx context.Context, userID uuid.UUID, fileData []byte, mimeType string) (*ParseResult, error) {
	return s.profileParse.ParseFromFile(ctx, userID, fileData, mimeType)
}
