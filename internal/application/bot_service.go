package application

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

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
	user.Country = countryFromPhone(phone)
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


// HandleGoalSelection transitions based on the selected goal.
// "salary" → collecting_resume state; "job" → clears state.
func (s *BotService) HandleGoalSelection(ctx context.Context, telegramID int64, goal string) (string, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	state, err := s.stateSvc.GetState(ctx, tgID)
	if err != nil || state == nil {
		return "", fmt.Errorf("no active state for goal selection")
	}

	language := state.Data["language"]
	if language == "" {
		language = "en"
	}

	if goal == "salary" {
		data := map[string]string{"language": language}
		if err := s.stateSvc.SetStateWithData(ctx, tgID, domain.BotStateCollectingResume, data); err != nil {
			return language, fmt.Errorf("set collecting resume state: %w", err)
		}
		return language, nil
	}

	if err := s.stateSvc.ClearState(ctx, tgID); err != nil {
		return language, fmt.Errorf("clear state: %w", err)
	}

	return language, nil
}

// AddResumeText stores a text entry for later processing.
func (s *BotService) AddResumeText(ctx context.Context, telegramID int64, text string) error {
	tgID := strconv.FormatInt(telegramID, 10)
	return s.stateSvc.AddResumeEntry(ctx, tgID, &ResumeEntry{
		Type: "text",
		Text: text,
	})
}

// AddResumeFile stores a file entry (base64-encoded) for later processing.
func (s *BotService) AddResumeFile(ctx context.Context, telegramID int64, data []byte, mimeType string) error {
	tgID := strconv.FormatInt(telegramID, 10)
	encoded := base64.StdEncoding.EncodeToString(data)
	return s.stateSvc.AddResumeEntry(ctx, tgID, &ResumeEntry{
		Type:     "file",
		Data:     encoded,
		MimeType: mimeType,
	})
}

// ProcessCollectedResume sends all collected resume data to Gemini and stores the result.
func (s *BotService) ProcessCollectedResume(ctx context.Context, telegramID int64) (*ParseResult, error) {
	tgID := strconv.FormatInt(telegramID, 10)

	user, err := s.userSvc.GetByTelegramID(ctx, tgID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	entries, err := s.stateSvc.GetResumeEntries(ctx, tgID)
	if err != nil || len(entries) == 0 {
		return nil, fmt.Errorf("no resume data collected")
	}

	// Separate text and file entries
	var texts []string
	var fileData []byte
	var fileMime string
	for _, e := range entries {
		switch e.Type {
		case "text":
			texts = append(texts, e.Text)
		case "file":
			if fileData == nil {
				decoded, decErr := base64.StdEncoding.DecodeString(e.Data)
				if decErr == nil {
					fileData = decoded
					fileMime = e.MimeType
				}
			}
		}
	}

	var result *ParseResult

	if fileData != nil && len(texts) > 0 {
		// Combine: prepend text context, then parse the file
		combinedText := strings.Join(texts, "\n\n") + "\n\n[Additional context from user messages above. The file below is the main resume.]"
		result, err = s.profileParse.ParseFromText(ctx, user.ID, combinedText)
		if err != nil {
			// Fallback: try just the file
			result, err = s.profileParse.ParseFromFile(ctx, user.ID, fileData, fileMime)
		}
	} else if fileData != nil {
		result, err = s.profileParse.ParseFromFile(ctx, user.ID, fileData, fileMime)
	} else if len(texts) > 0 {
		combinedText := strings.Join(texts, "\n\n")
		result, err = s.profileParse.ParseFromText(ctx, user.ID, combinedText)
	} else {
		return nil, fmt.Errorf("no resume data to process")
	}

	if err != nil {
		return nil, err
	}

	// Clean up
	_ = s.stateSvc.ClearResumeEntries(ctx, tgID)
	_ = s.stateSvc.ClearState(ctx, tgID)

	return result, nil
}

func (s *BotService) HandleResumeText(ctx context.Context, userID uuid.UUID, text string) (*ParseResult, error) {
	return s.profileParse.ParseFromText(ctx, userID, text)
}

func (s *BotService) HandleResumeFile(ctx context.Context, userID uuid.UUID, fileData []byte, mimeType string) (*ParseResult, error) {
	return s.profileParse.ParseFromFile(ctx, userID, fileData, mimeType)
}

// countryFromPhone returns an ISO 3166-1 alpha-2 country code based on phone number prefix.
func countryFromPhone(phone string) string {
	p := strings.TrimLeft(phone, "+")

	// Longest prefixes first to avoid ambiguity
	prefixes := []struct {
		prefix, code string
	}{
		// 4-digit
		{"9989", "UZ"}, // Uzbekistan mobile
		{"9987", "UZ"},
		{"9971", "UZ"},
		// 3-digit
		{"998", "UZ"},  // Uzbekistan
		{"996", "KG"},  // Kyrgyzstan
		{"995", "GE"},  // Georgia
		{"994", "AZ"},  // Azerbaijan
		{"993", "TM"},  // Turkmenistan
		{"992", "TJ"},  // Tajikistan
		{"971", "AE"},  // UAE
		{"966", "SA"},  // Saudi Arabia
		{"420", "CZ"},  // Czech Republic
		{"380", "UA"},  // Ukraine
		{"375", "BY"},  // Belarus
		{"374", "AM"},  // Armenia
		{"373", "MD"},  // Moldova
		{"372", "EE"},  // Estonia
		{"371", "LV"},  // Latvia
		{"370", "LT"},  // Lithuania
		// 2-digit
		{"77", "KZ"},  // Kazakhstan
		{"79", "RU"},  // Russia mobile
		{"78", "RU"},  // Russia
		{"74", "RU"},  // Russia
		{"73", "RU"},  // Russia
		{"91", "IN"},  // India
		{"90", "TR"},  // Turkey
		{"86", "CN"},  // China
		{"82", "KR"},  // South Korea
		{"81", "JP"},  // Japan
		{"49", "DE"},  // Germany
		{"48", "PL"},  // Poland
		{"44", "GB"},  // UK
		{"33", "FR"},  // France
		// 1-digit
		{"7", "RU"},   // Russia/KZ fallback
		{"1", "US"},   // USA/Canada
	}

	for _, entry := range prefixes {
		if strings.HasPrefix(p, entry.prefix) {
			return entry.code
		}
	}

	return ""
}
