package application

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/qdrant"
)

const collectionName = "user_profile_vectors"
const vacancyCollectionName = "vacancy_vectors"

type sectionDef struct {
	Name   string
	Weight float64
}

var sections = []sectionDef{
	{Name: "title", Weight: 1.5},
	{Name: "skills", Weight: 1.4},
	{Name: "experience", Weight: 1.2},
	{Name: "about", Weight: 1.0},
}

var vacancySections = []sectionDef{
	{Name: "title", Weight: 1.5},
	{Name: "skills", Weight: 1.4},
	{Name: "description", Weight: 1.2},
	{Name: "requirements", Weight: 1.0},
}

type VectorIndexService struct {
	qdrantClient    *qdrant.Client
	geminiClient    *gemini.Client
	profileFieldSvc *ProfileFieldService
	profileTextSvc  *ProfileFieldTextService
	experienceSvc   *ExperienceItemService
	itemTextSvc     *ItemTextService
	skillSvc        *SkillService
	userSvc         *UserService
	vacancyRepo     domain.VacancyRepository
	vacancyTextRepo domain.VacancyTextRepository
}

func NewVectorIndexService(
	qdrantClient *qdrant.Client,
	geminiClient *gemini.Client,
	profileFieldSvc *ProfileFieldService,
	profileTextSvc *ProfileFieldTextService,
	experienceSvc *ExperienceItemService,
	itemTextSvc *ItemTextService,
	skillSvc *SkillService,
	userSvc *UserService,
	vacancyRepo domain.VacancyRepository,
	vacancyTextRepo domain.VacancyTextRepository,
) *VectorIndexService {
	return &VectorIndexService{
		qdrantClient:    qdrantClient,
		geminiClient:    geminiClient,
		profileFieldSvc: profileFieldSvc,
		profileTextSvc:  profileTextSvc,
		experienceSvc:   experienceSvc,
		itemTextSvc:     itemTextSvc,
		skillSvc:        skillSvc,
		userSvc:         userSvc,
		vacancyRepo:     vacancyRepo,
		vacancyTextRepo: vacancyTextRepo,
	}
}

func (s *VectorIndexService) IndexUser(ctx context.Context, userID uuid.UUID) error {
	chunks := s.buildChunks(ctx, userID)

	// Delete existing vectors for this user
	if err := s.qdrantClient.DeletePointsByPayload(ctx, collectionName, "user_id", userID.String()); err != nil {
		return fmt.Errorf("delete existing vectors: %w", err)
	}

	var points []qdrant.Point
	for _, sec := range sections {
		text, ok := chunks[sec.Name]
		if !ok || text == "" {
			continue
		}

		vector, err := s.geminiClient.EmbedText(ctx, text)
		if err != nil {
			log.Printf("embed %s for user %s: %v", sec.Name, userID, err)
			continue
		}

		pointID := deterministicUUID(userID.String(), sec.Name)

		points = append(points, qdrant.Point{
			ID:     pointID,
			Vector: vector,
			Payload: map[string]any{
				"user_id": userID.String(),
				"section": sec.Name,
				"weight":  sec.Weight,
			},
		})
	}

	if len(points) == 0 {
		return nil
	}

	if err := s.qdrantClient.UpsertPoints(ctx, collectionName, points); err != nil {
		return fmt.Errorf("upsert vectors: %w", err)
	}

	return nil
}

func (s *VectorIndexService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.qdrantClient.DeletePointsByPayload(ctx, collectionName, "user_id", userID.String())
}

func (s *VectorIndexService) ReindexAll(ctx context.Context) error {
	// Reindex all users
	page := int32(1)
	pageSize := int32(50)
	for {
		result, err := s.userSvc.ListUsers(ctx, page, pageSize)
		if err != nil {
			return fmt.Errorf("list users page %d: %w", page, err)
		}
		if len(result.Users) == 0 {
			break
		}
		for _, u := range result.Users {
			if err := s.IndexUser(ctx, u.ID); err != nil {
				log.Printf("reindex user %s: %v", u.ID, err)
			}
		}
		if int64(page*pageSize) >= result.Total {
			break
		}
		page++
	}

	// Reindex all vacancies
	if err := s.ReindexAllVacancies(ctx); err != nil {
		return fmt.Errorf("reindex vacancies: %w", err)
	}

	return nil
}

func (s *VectorIndexService) IndexVacancy(ctx context.Context, vacancyID uuid.UUID) error {
	chunks := s.buildVacancyChunks(ctx, vacancyID)

	if err := s.qdrantClient.DeletePointsByPayload(ctx, vacancyCollectionName, "vacancy_id", vacancyID.String()); err != nil {
		return fmt.Errorf("delete existing vacancy vectors: %w", err)
	}

	var points []qdrant.Point
	for _, sec := range vacancySections {
		text, ok := chunks[sec.Name]
		if !ok || text == "" {
			continue
		}

		vector, err := s.geminiClient.EmbedText(ctx, text)
		if err != nil {
			log.Printf("embed %s for vacancy %s: %v", sec.Name, vacancyID, err)
			continue
		}

		pointID := deterministicUUID("vacancy:"+vacancyID.String(), sec.Name)

		points = append(points, qdrant.Point{
			ID:     pointID,
			Vector: vector,
			Payload: map[string]any{
				"vacancy_id": vacancyID.String(),
				"section":    sec.Name,
				"weight":     sec.Weight,
			},
		})
	}

	if len(points) == 0 {
		return nil
	}

	if err := s.qdrantClient.UpsertPoints(ctx, vacancyCollectionName, points); err != nil {
		return fmt.Errorf("upsert vacancy vectors: %w", err)
	}

	return nil
}

func (s *VectorIndexService) DeleteVacancy(ctx context.Context, vacancyID uuid.UUID) error {
	return s.qdrantClient.DeletePointsByPayload(ctx, vacancyCollectionName, "vacancy_id", vacancyID.String())
}

func (s *VectorIndexService) ReindexAllVacancies(ctx context.Context) error {
	page := int32(1)
	pageSize := int32(50)
	for {
		offset := (page - 1) * pageSize
		vacancies, err := s.vacancyRepo.List(ctx, pageSize, offset)
		if err != nil {
			return fmt.Errorf("list vacancies page %d: %w", page, err)
		}
		if len(vacancies) == 0 {
			break
		}
		for _, v := range vacancies {
			if err := s.IndexVacancy(ctx, v.ID); err != nil {
				log.Printf("reindex vacancy %s: %v", v.ID, err)
			}
		}
		if len(vacancies) < int(pageSize) {
			break
		}
		page++
	}
	return nil
}

func (s *VectorIndexService) buildVacancyChunks(ctx context.Context, vacancyID uuid.UUID) map[string]string {
	chunks := make(map[string]string)

	texts, err := s.vacancyTextRepo.ListByVacancy(ctx, vacancyID)
	if err == nil {
		for _, t := range texts {
			if t.Lang == "en" {
				if t.Title != "" {
					chunks["title"] = t.Title
				}
				if t.Description != "" {
					chunks["description"] = t.Description
				}
				if t.Requirements != "" {
					chunks["requirements"] = t.Requirements
				}
				break
			}
		}
	}

	skills, err := s.skillSvc.ListVacancySkills(ctx, vacancyID)
	if err == nil && len(skills) > 0 {
		names := make([]string, 0, len(skills))
		for _, sk := range skills {
			names = append(names, sk.Name)
		}
		chunks["skills"] = strings.Join(names, ", ")
	}

	return chunks
}

func (s *VectorIndexService) buildChunks(ctx context.Context, userID uuid.UUID) map[string]string {
	chunks := make(map[string]string)

	// Get profile fields (title, about) in English
	fields, err := s.profileFieldSvc.ListProfileFieldsByUser(ctx, userID)
	if err == nil {
		for _, f := range fields {
			if f.FieldName == "title" || f.FieldName == "about" {
				text, err := s.profileTextSvc.GetProfileFieldText(ctx, f.ID, "en")
				if err == nil && text.Content != "" {
					chunks[f.FieldName] = text.Content
				}
			}
		}
	}

	// Get skills as comma-separated string
	skills, err := s.skillSvc.ListUserSkills(ctx, userID)
	if err == nil && len(skills) > 0 {
		names := make([]string, 0, len(skills))
		for _, sk := range skills {
			names = append(names, sk.Name)
		}
		chunks["skills"] = strings.Join(names, ", ")
	}

	// Get experience descriptions in English, concatenated
	expItems, err := s.experienceSvc.ListExperienceItemsByUser(ctx, userID)
	if err == nil && len(expItems) > 0 {
		var expTexts []string
		for _, item := range expItems {
			texts, err := s.itemTextSvc.ListItemTextsByItem(ctx, item.ID, "experience")
			if err == nil {
				for _, t := range texts {
					if t.Lang == "en" && t.Description != "" {
						desc := item.Position + " at " + item.Company + ". " + t.Description
						expTexts = append(expTexts, desc)
						break
					}
				}
			}
		}
		if len(expTexts) > 0 {
			chunks["experience"] = strings.Join(expTexts, " | ")
		}
	}

	return chunks
}

func deterministicUUID(userID, section string) string {
	h := sha256.Sum256([]byte(userID + ":" + section))
	// Use first 16 bytes as UUID v4-like
	u := uuid.UUID{}
	copy(u[:], h[:16])
	u[6] = (u[6] & 0x0f) | 0x40 // version 4
	u[8] = (u[8] & 0x3f) | 0x80 // variant 10
	return u.String()
}
