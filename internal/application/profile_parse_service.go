package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
)

type ProfileParseService struct {
	geminiClient    *gemini.Client
	profileFieldSvc *ProfileFieldService
	profileTextSvc  *ProfileFieldTextService
	experienceSvc   *ExperienceItemService
	educationSvc    *EducationItemService
	itemTextSvc     *ItemTextService
	userSvc         *UserService
}

func NewProfileParseService(
	geminiClient *gemini.Client,
	profileFieldSvc *ProfileFieldService,
	profileTextSvc *ProfileFieldTextService,
	experienceSvc *ExperienceItemService,
	educationSvc *EducationItemService,
	itemTextSvc *ItemTextService,
	userSvc *UserService,
) *ProfileParseService {
	return &ProfileParseService{
		geminiClient:    geminiClient,
		profileFieldSvc: profileFieldSvc,
		profileTextSvc:  profileTextSvc,
		experienceSvc:   experienceSvc,
		educationSvc:    educationSvc,
		itemTextSvc:     itemTextSvc,
		userSvc:         userSvc,
	}
}

type ParseResult struct {
	SourceLang string
	Fields     []ParsedFieldResult
	Experience []ParsedExperienceResult
	Education  []ParsedEducationResult
}

type ParsedFieldResult struct {
	Field *domain.ProfileField
	Texts []domain.ProfileFieldText
}

type ParsedExperienceResult struct {
	Item  *domain.ExperienceItem
	Texts []domain.ItemText
}

type ParsedEducationResult struct {
	Item  *domain.EducationItem
	Texts []domain.ItemText
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
	if err := s.deleteExistingData(ctx, userID); err != nil {
		return nil, fmt.Errorf("delete existing data: %w", err)
	}

	langs := []string{"uz", "ru", "en"}
	modelVer := s.geminiClient.ModelVersion()

	result := &ParseResult{
		SourceLang: parsed.SourceLang,
		Fields:     make([]ParsedFieldResult, 0, len(parsed.Fields)),
		Experience: make([]ParsedExperienceResult, 0, len(parsed.Experience)),
		Education:  make([]ParsedEducationResult, 0, len(parsed.Education)),
	}

	// Store text fields
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

	// Store skills as JSON array per language
	if len(parsed.Skills) > 0 {
		field, err := s.profileFieldSvc.CreateProfileField(ctx, &domain.ProfileField{
			UserID:     userID,
			FieldName:  "skills",
			SourceLang: parsed.SourceLang,
		})
		if err != nil {
			return nil, fmt.Errorf("create profile field skills: %w", err)
		}
		fieldResult := ParsedFieldResult{
			Field: field,
			Texts: make([]domain.ProfileFieldText, 0, 3),
		}
		for _, lang := range langs {
			skills, ok := parsed.Skills[lang]
			if !ok || len(skills) == 0 {
				continue
			}
			jsonBytes, _ := json.Marshal(skills)
			text, err := s.profileTextSvc.CreateProfileFieldText(ctx, &domain.ProfileFieldText{
				ProfileFieldID: field.ID,
				Lang:           lang,
				Content:        string(jsonBytes),
				IsSource:       lang == parsed.SourceLang,
				ModelVersion:   modelVer,
			})
			if err != nil {
				return nil, fmt.Errorf("create profile field text skills/%s: %w", lang, err)
			}
			fieldResult.Texts = append(fieldResult.Texts, *text)
		}
		result.Fields = append(result.Fields, fieldResult)
	}

	// Store certifications as JSON array per language
	if len(parsed.Certifications) > 0 {
		field, err := s.profileFieldSvc.CreateProfileField(ctx, &domain.ProfileField{
			UserID:     userID,
			FieldName:  "certifications",
			SourceLang: parsed.SourceLang,
		})
		if err != nil {
			return nil, fmt.Errorf("create profile field certifications: %w", err)
		}
		fieldResult := ParsedFieldResult{
			Field: field,
			Texts: make([]domain.ProfileFieldText, 0, 3),
		}
		for _, lang := range langs {
			certs, ok := parsed.Certifications[lang]
			if !ok || len(certs) == 0 {
				continue
			}
			jsonBytes, _ := json.Marshal(certs)
			text, err := s.profileTextSvc.CreateProfileFieldText(ctx, &domain.ProfileFieldText{
				ProfileFieldID: field.ID,
				Lang:           lang,
				Content:        string(jsonBytes),
				IsSource:       lang == parsed.SourceLang,
				ModelVersion:   modelVer,
			})
			if err != nil {
				return nil, fmt.Errorf("create profile field text certifications/%s: %w", lang, err)
			}
			fieldResult.Texts = append(fieldResult.Texts, *text)
		}
		result.Fields = append(result.Fields, fieldResult)
	}

	// Store languages as JSON array per language
	if len(parsed.Languages) > 0 {
		field, err := s.profileFieldSvc.CreateProfileField(ctx, &domain.ProfileField{
			UserID:     userID,
			FieldName:  "languages",
			SourceLang: parsed.SourceLang,
		})
		if err != nil {
			return nil, fmt.Errorf("create profile field languages: %w", err)
		}
		fieldResult := ParsedFieldResult{
			Field: field,
			Texts: make([]domain.ProfileFieldText, 0, 3),
		}
		for _, lang := range langs {
			var items []map[string]string
			for _, l := range parsed.Languages {
				name, ok := l.Name[lang]
				if !ok || name == "" {
					continue
				}
				items = append(items, map[string]string{
					"name":  name,
					"level": l.Level,
				})
			}
			if len(items) == 0 {
				continue
			}
			jsonBytes, _ := json.Marshal(items)
			text, err := s.profileTextSvc.CreateProfileFieldText(ctx, &domain.ProfileFieldText{
				ProfileFieldID: field.ID,
				Lang:           lang,
				Content:        string(jsonBytes),
				IsSource:       lang == parsed.SourceLang,
				ModelVersion:   modelVer,
			})
			if err != nil {
				return nil, fmt.Errorf("create profile field text languages/%s: %w", lang, err)
			}
			fieldResult.Texts = append(fieldResult.Texts, *text)
		}
		result.Fields = append(result.Fields, fieldResult)
	}

	// Store experience items
	for i, exp := range parsed.Experience {
		// Use source lang for the position value stored in the item row
		position := exp.Position["en"]
		if v, ok := exp.Position[parsed.SourceLang]; ok && v != "" {
			position = v
		}

		projectsJSON := ""
		if len(exp.Projects) > 0 {
			b, _ := json.Marshal(exp.Projects)
			projectsJSON = string(b)
		}

		item, err := s.experienceSvc.CreateExperienceItem(ctx, &domain.ExperienceItem{
			UserID:    userID,
			Company:   exp.Company,
			Position:  position,
			StartDate: exp.StartDate,
			EndDate:   exp.EndDate,
			Projects:  projectsJSON,
			WebSite:   exp.WebSite,
			ItemOrder: int32(i),
		})
		if err != nil {
			return nil, fmt.Errorf("create experience item %d: %w", i, err)
		}

		expResult := ParsedExperienceResult{
			Item:  item,
			Texts: make([]domain.ItemText, 0, 3),
		}

		// Create description texts for each language
		for _, lang := range langs {
			desc, ok := exp.Description[lang]
			if !ok || desc == "" {
				continue
			}

			text, err := s.itemTextSvc.CreateItemText(ctx, &domain.ItemText{
				ItemID:       item.ID,
				ItemType:     "experience",
				Lang:         lang,
				Description:  desc,
				IsSource:     lang == parsed.SourceLang,
				ModelVersion: modelVer,
			})
			if err != nil {
				return nil, fmt.Errorf("create experience text %d/%s: %w", i, lang, err)
			}
			expResult.Texts = append(expResult.Texts, *text)
		}

		result.Experience = append(result.Experience, expResult)
	}

	// Store education items
	for i, edu := range parsed.Education {
		degree := edu.Degree["en"]
		if v, ok := edu.Degree[parsed.SourceLang]; ok && v != "" {
			degree = v
		}
		fieldOfStudy := edu.FieldOfStudy["en"]
		if v, ok := edu.FieldOfStudy[parsed.SourceLang]; ok && v != "" {
			fieldOfStudy = v
		}

		item, err := s.educationSvc.CreateEducationItem(ctx, &domain.EducationItem{
			UserID:       userID,
			Institution:  edu.Institution,
			Degree:       degree,
			FieldOfStudy: fieldOfStudy,
			StartDate:    edu.StartDate,
			EndDate:      edu.EndDate,
			Location:     edu.Location,
			ItemOrder:    int32(i),
		})
		if err != nil {
			return nil, fmt.Errorf("create education item %d: %w", i, err)
		}

		eduResult := ParsedEducationResult{
			Item:  item,
			Texts: make([]domain.ItemText, 0, 3),
		}

		for _, lang := range langs {
			desc, ok := edu.Description[lang]
			if !ok || desc == "" {
				continue
			}

			text, err := s.itemTextSvc.CreateItemText(ctx, &domain.ItemText{
				ItemID:       item.ID,
				ItemType:     "education",
				Lang:         lang,
				Description:  desc,
				IsSource:     lang == parsed.SourceLang,
				ModelVersion: modelVer,
			})
			if err != nil {
				return nil, fmt.Errorf("create education text %d/%s: %w", i, lang, err)
			}
			eduResult.Texts = append(eduResult.Texts, *text)
		}

		result.Education = append(result.Education, eduResult)
	}

	return result, nil
}

func (s *ProfileParseService) deleteExistingData(ctx context.Context, userID uuid.UUID) error {
	// Delete profile field texts, then fields
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

	// Delete experience item texts, then items
	existingExp, err := s.experienceSvc.ListExperienceItemsByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("list existing experience items: %w", err)
	}
	for _, e := range existingExp {
		if err := s.itemTextSvc.DeleteItemTextsByItemID(ctx, e.ID); err != nil {
			return fmt.Errorf("delete texts for experience %s: %w", e.ID, err)
		}
	}
	if err := s.experienceSvc.DeleteExperienceItemsByUser(ctx, userID); err != nil {
		return fmt.Errorf("delete experience items for user: %w", err)
	}

	// Delete education item texts, then items
	existingEdu, err := s.educationSvc.ListEducationItemsByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("list existing education items: %w", err)
	}
	for _, e := range existingEdu {
		if err := s.itemTextSvc.DeleteItemTextsByItemID(ctx, e.ID); err != nil {
			return fmt.Errorf("delete texts for education %s: %w", e.ID, err)
		}
	}
	if err := s.educationSvc.DeleteEducationItemsByUser(ctx, userID); err != nil {
		return fmt.Errorf("delete education items for user: %w", err)
	}

	return nil
}
