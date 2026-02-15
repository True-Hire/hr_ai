package application

import (
	"context"
	"log"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/qdrant"
)

type VacancySearchService struct {
	qdrantClient    *qdrant.Client
	geminiClient    *gemini.Client
	vacancySvc      *VacancyService
	profileFieldSvc *ProfileFieldService
	profileTextSvc  *ProfileFieldTextService
	skillSvc        *SkillService
}

func NewVacancySearchService(
	qdrantClient *qdrant.Client,
	geminiClient *gemini.Client,
	vacancySvc *VacancyService,
	profileFieldSvc *ProfileFieldService,
	profileTextSvc *ProfileFieldTextService,
	skillSvc *SkillService,
) *VacancySearchService {
	return &VacancySearchService{
		qdrantClient:    qdrantClient,
		geminiClient:    geminiClient,
		vacancySvc:      vacancySvc,
		profileFieldSvc: profileFieldSvc,
		profileTextSvc:  profileTextSvc,
		skillSvc:        skillSvc,
	}
}

// MatchVacanciesForUser finds vacancies matching a user's profile via vector similarity.
func (s *VacancySearchService) MatchVacanciesForUser(ctx context.Context, userID uuid.UUID, lang string, page, pageSize int32) (*ListVacanciesResult, error) {
	profileText := s.buildUserProfileText(ctx, userID)
	if profileText == "" {
		return s.vacancySvc.ListVacancies(ctx, page, pageSize)
	}

	return s.searchByText(ctx, profileText, page, pageSize)
}

// SearchVacancies searches vacancies by a user-typed query using vector similarity.
func (s *VacancySearchService) SearchVacancies(ctx context.Context, query string, page, pageSize int32) (*ListVacanciesResult, error) {
	if query == "" {
		return s.vacancySvc.ListVacancies(ctx, page, pageSize)
	}

	return s.searchByText(ctx, query, page, pageSize)
}

// searchByText translates text to English, embeds, and searches vacancy_vectors.
func (s *VacancySearchService) searchByText(ctx context.Context, text string, page, pageSize int32) (*ListVacanciesResult, error) {
	englishText, err := s.geminiClient.TranslateToEnglish(ctx, text)
	if err != nil {
		log.Printf("translate search text: %v", err)
		return s.vacancySvc.ListVacancies(ctx, page, pageSize)
	}

	vector, err := s.geminiClient.EmbedText(ctx, englishText)
	if err != nil {
		log.Printf("embed search text: %v", err)
		return s.vacancySvc.ListVacancies(ctx, page, pageSize)
	}

	qdrantLimit := max(int(pageSize)*4, 40)
	scored, err := s.qdrantClient.Search(ctx, vacancyCollectionName, vector, qdrantLimit)
	if err != nil {
		log.Printf("qdrant vacancy search: %v", err)
		return s.vacancySvc.ListVacancies(ctx, page, pageSize)
	}

	vacancyScores := make(map[string]float64)
	for _, sp := range scored {
		vacancyIDStr, _ := sp.Payload["vacancy_id"].(string)
		if vacancyIDStr == "" {
			continue
		}
		weight := 1.0
		if w, ok := sp.Payload["weight"].(float64); ok {
			weight = w
		}
		vacancyScores[vacancyIDStr] += sp.Score * weight
	}

	type vacancyScore struct {
		id    string
		score float64
	}
	sorted := make([]vacancyScore, 0, len(vacancyScores))
	for id, score := range vacancyScores {
		sorted = append(sorted, vacancyScore{id: id, score: score})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	total := int64(len(sorted))
	start := int((page - 1) * pageSize)
	if start >= len(sorted) {
		return &ListVacanciesResult{Vacancies: []VacancyWithDetails{}, Total: total}, nil
	}
	end := min(start+int(pageSize), len(sorted))
	pageItems := sorted[start:end]

	vacancies := make([]VacancyWithDetails, 0, len(pageItems))
	for _, vs := range pageItems {
		vid, err := uuid.Parse(vs.id)
		if err != nil {
			continue
		}
		vwd, err := s.vacancySvc.GetVacancy(ctx, vid)
		if err != nil {
			continue
		}
		vacancies = append(vacancies, *vwd)
	}

	return &ListVacanciesResult{Vacancies: vacancies, Total: total}, nil
}

func (s *VacancySearchService) buildUserProfileText(ctx context.Context, userID uuid.UUID) string {
	var parts []string

	fields, err := s.profileFieldSvc.ListProfileFieldsByUser(ctx, userID)
	if err == nil {
		for _, f := range fields {
			if f.FieldName == "title" || f.FieldName == "about" {
				text, err := s.profileTextSvc.GetProfileFieldText(ctx, f.ID, "en")
				if err == nil && text.Content != "" {
					parts = append(parts, text.Content)
				}
			}
		}
	}

	skills, err := s.skillSvc.ListUserSkills(ctx, userID)
	if err == nil && len(skills) > 0 {
		names := make([]string, 0, len(skills))
		for _, sk := range skills {
			names = append(names, sk.Name)
		}
		parts = append(parts, "Skills: "+strings.Join(names, ", "))
	}

	return strings.Join(parts, ". ")
}
