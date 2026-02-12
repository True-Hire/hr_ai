package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
)

type ProfileParseService struct {
	geminiClient    *gemini.Client
	profileFieldSvc *ProfileFieldService
	profileTextSvc  *ProfileFieldTextService
	userSvc         *UserService
}

func NewProfileParseService(
	geminiClient *gemini.Client,
	profileFieldSvc *ProfileFieldService,
	profileTextSvc *ProfileFieldTextService,
	userSvc *UserService,
) *ProfileParseService {
	return &ProfileParseService{
		geminiClient:    geminiClient,
		profileFieldSvc: profileFieldSvc,
		profileTextSvc:  profileTextSvc,
		userSvc:         userSvc,
	}
}

type ParseResult struct {
	SourceLang string
	Fields     []ParsedFieldResult
}

type ParsedFieldResult struct {
	Field *domain.ProfileField
	Texts []domain.ProfileFieldText
}

func (s *ProfileParseService) ParseFromText(ctx context.Context, userID uuid.UUID, userInput string) (*ParseResult, error) {
	if _, err := s.userSvc.GetUser(ctx, userID); err != nil {
		return nil, fmt.Errorf("verify user: %w", err)
	}

	parsed, err := s.geminiClient.ParseProfileFromText(ctx, userInput)
	if err != nil {
		return nil, fmt.Errorf("gemini parse text: %w", err)
	}

	return s.storeResults(ctx, userID, parsed)
}

func (s *ProfileParseService) ParseFromFile(ctx context.Context, userID uuid.UUID, fileData []byte, mimeType string) (*ParseResult, error) {
	if _, err := s.userSvc.GetUser(ctx, userID); err != nil {
		return nil, fmt.Errorf("verify user: %w", err)
	}

	parsed, err := s.geminiClient.ParseProfileFromFile(ctx, fileData, mimeType)
	if err != nil {
		return nil, fmt.Errorf("gemini parse file: %w", err)
	}

	return s.storeResults(ctx, userID, parsed)
}

func (s *ProfileParseService) storeResults(ctx context.Context, userID uuid.UUID, parsed *gemini.ParsedProfile) (*ParseResult, error) {
	if err := s.deleteExistingFields(ctx, userID); err != nil {
		return nil, fmt.Errorf("delete existing fields: %w", err)
	}

	result := &ParseResult{
		SourceLang: parsed.SourceLang,
		Fields:     make([]ParsedFieldResult, 0, len(parsed.Fields)),
	}

	langs := []string{"uz", "ru", "en"}
	modelVer := s.geminiClient.ModelVersion()

	for fieldName, translations := range parsed.Fields {
		field, err := s.profileFieldSvc.CreateProfileField(ctx, &domain.ProfileField{
			UserID:     userID,
			FieldName:  fieldName,
			SourceLang: parsed.SourceLang,
		})
		if err != nil {
			return nil, fmt.Errorf("create profile field %q: %w", fieldName, err)
		}

		fieldResult := ParsedFieldResult{
			Field: field,
			Texts: make([]domain.ProfileFieldText, 0, 3),
		}

		for _, lang := range langs {
			content, ok := translations[lang]
			if !ok || content == "" {
				continue
			}

			text, err := s.profileTextSvc.CreateProfileFieldText(ctx, &domain.ProfileFieldText{
				ProfileFieldID: field.ID,
				Lang:           lang,
				Content:        content,
				IsSource:       lang == parsed.SourceLang,
				ModelVersion:   modelVer,
			})
			if err != nil {
				return nil, fmt.Errorf("create profile field text %q/%s: %w", fieldName, lang, err)
			}
			fieldResult.Texts = append(fieldResult.Texts, *text)
		}

		result.Fields = append(result.Fields, fieldResult)
	}

	return result, nil
}

func (s *ProfileParseService) deleteExistingFields(ctx context.Context, userID uuid.UUID) error {
	existingFields, err := s.profileFieldSvc.ListProfileFieldsByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("list existing fields: %w", err)
	}

	for _, f := range existingFields {
		if err := s.profileTextSvc.DeleteProfileFieldTextsByField(ctx, f.ID); err != nil {
			return fmt.Errorf("delete texts for field %s: %w", f.ID, err)
		}
	}

	if err := s.profileFieldSvc.DeleteProfileFieldsByUser(ctx, userID); err != nil {
		return fmt.Errorf("delete fields for user: %w", err)
	}

	return nil
}
