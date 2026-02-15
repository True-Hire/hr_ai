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

// HandleLanguageSelection transitions from choosing_language to choosing_role.
// Stores the chosen language in state data for use when creating the account.
func (s *BotService) HandleLanguageSelection(ctx context.Context, telegramID int64, language string) (string, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	if language != "en" && language != "ru" && language != "uz" {
		return "", fmt.Errorf("invalid language: %s", language)
	}

	data := map[string]string{"language": language}
	if err := s.stateSvc.SetStateWithData(ctx, tgID, domain.BotStateChoosingRole, data); err != nil {
		return "", fmt.Errorf("set role selection state: %w", err)
	}

	return language, nil
}

// HandleRoleSelection creates either a User or CompanyHR based on the chosen role.
// Returns the language, whether the user chose HR, and any error.
func (s *BotService) HandleRoleSelection(ctx context.Context, telegramID int64, role, firstName, lastName, username string, photoData []byte) (string, bool, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	state, err := s.stateSvc.GetState(ctx, tgID)
	if err != nil || state == nil {
		return "", false, fmt.Errorf("no active state for role selection")
	}

	language := state.Data["language"]
	if language == "" {
		language = "en"
	}

	if err := s.stateSvc.ClearState(ctx, tgID); err != nil {
		return "", false, fmt.Errorf("clear state: %w", err)
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

	if role == "hr" {
		hr := &domain.CompanyHR{
			FirstName:  firstName,
			LastName:   lastName,
			Telegram:   tg,
			TelegramID: tgID,
			Language:   language,
		}
		if _, err := s.hrSvc.CreateCompanyHR(ctx, hr); err != nil {
			return "", false, fmt.Errorf("create hr: %w", err)
		}
		return language, true, nil
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
		return "", false, fmt.Errorf("create user: %w", err)
	}
	return language, false, nil
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
