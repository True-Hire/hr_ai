package application

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/qdrant"
)

type SearchService struct {
	qdrantClient  *qdrant.Client
	geminiClient  *gemini.Client
	userSvc       *UserService
	vacancySvc    *VacancyService
	experienceSvc *ExperienceItemService
	skillSvc      *SkillService
}

func NewSearchService(
	qdrantClient *qdrant.Client,
	geminiClient *gemini.Client,
	userSvc *UserService,
	vacancySvc *VacancyService,
	experienceSvc *ExperienceItemService,
	skillSvc *SkillService,
) *SearchService {
	return &SearchService{
		qdrantClient:  qdrantClient,
		geminiClient:  geminiClient,
		userSvc:       userSvc,
		vacancySvc:    vacancySvc,
		experienceSvc: experienceSvc,
		skillSvc:      skillSvc,
	}
}

type SearchResult struct {
	UserID uuid.UUID
	Score  float64
}

func (s *SearchService) SearchUsers(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Translate query to English
	englishQuery, err := s.geminiClient.TranslateToEnglish(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("translate query: %w", err)
	}

	// Embed the English query
	vector, err := s.geminiClient.EmbedText(ctx, englishQuery)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	// Search Qdrant — request more results to account for multiple vectors per user
	qdrantLimit := max(limit*4, 20)

	scored, err := s.qdrantClient.Search(ctx, collectionName, vector, qdrantLimit)
	if err != nil {
		return nil, fmt.Errorf("qdrant search: %w", err)
	}

	// Group by user_id and aggregate weighted scores
	userScores := make(map[string]float64)
	for _, sp := range scored {
		userIDStr, _ := sp.Payload["user_id"].(string)
		if userIDStr == "" {
			continue
		}
		weight := 1.0
		if w, ok := sp.Payload["weight"].(float64); ok {
			weight = w
		}
		userScores[userIDStr] += sp.Score * weight
	}

	// Sort by score descending
	type userScore struct {
		id    string
		score float64
	}
	sorted := make([]userScore, 0, len(userScores))
	for id, score := range userScores {
		sorted = append(sorted, userScore{id: id, score: score})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	// Take top N
	if len(sorted) > limit {
		sorted = sorted[:limit]
	}

	results := make([]SearchResult, 0, len(sorted))
	for _, us := range sorted {
		uid, err := uuid.Parse(us.id)
		if err != nil {
			continue
		}
		results = append(results, SearchResult{
			UserID: uid,
			Score:  us.score,
		})
	}

	return results, nil
}

// CandidateMatch represents a user matching a vacancy with match percentage and experience.
type CandidateMatch struct {
	UserID               uuid.UUID
	MatchPercentage      int
	TotalExperienceYears float64
	VectorScore          float64
}

// SearchMatchingCandidates finds users matching a specific vacancy.
// Match percentage is computed from vector similarity (70%) + experience fit (30%).
// Results are sorted by match percentage descending.
func (s *SearchService) SearchMatchingCandidates(ctx context.Context, vacancyID uuid.UUID, limit int) ([]CandidateMatch, error) {
	// Get vacancy details
	vacancy, err := s.vacancySvc.GetVacancy(ctx, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("get vacancy: %w", err)
	}

	// Build search query from vacancy title + skills + requirements
	query := buildVacancySearchQuery(vacancy)
	if query == "" {
		return nil, nil
	}

	// Search users via vector similarity
	results, err := s.SearchUsers(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}

	// Find max score for normalization
	maxScore := 0.0
	for _, r := range results {
		if r.Score > maxScore {
			maxScore = r.Score
		}
	}
	if maxScore == 0 {
		maxScore = 1
	}

	expMin := float64(vacancy.Vacancy.ExperienceMin)
	expMax := float64(vacancy.Vacancy.ExperienceMax)

	// Get vacancy skill names for skill overlap calculation
	vacancySkillNames := make(map[string]bool, len(vacancy.Skills))
	for _, sk := range vacancy.Skills {
		vacancySkillNames[strings.ToLower(sk.Name)] = true
	}

	matches := make([]CandidateMatch, 0, len(results))
	for _, r := range results {
		totalExp := s.calculateTotalExperience(ctx, r.UserID)

		// Vector similarity score normalized to 0-1
		vectorNorm := r.Score / maxScore

		// Experience fit score (0-1)
		expScore := calculateExperienceFit(totalExp, expMin, expMax)

		// Skill overlap score (0-1)
		skillScore := s.calculateSkillOverlap(ctx, r.UserID, vacancySkillNames)

		// Weighted: vector 50% + experience 25% + skills 25%
		matchPct := int(math.Round((vectorNorm*0.50 + expScore*0.25 + skillScore*0.25) * 100))
		if matchPct > 99 {
			matchPct = 99
		}
		if matchPct < 1 {
			matchPct = 1
		}

		matches = append(matches, CandidateMatch{
			UserID:               r.UserID,
			MatchPercentage:      matchPct,
			TotalExperienceYears: math.Round(totalExp*10) / 10,
			VectorScore:          r.Score,
		})
	}

	// Sort by match percentage descending, then by experience descending
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].MatchPercentage != matches[j].MatchPercentage {
			return matches[i].MatchPercentage > matches[j].MatchPercentage
		}
		return matches[i].TotalExperienceYears > matches[j].TotalExperienceYears
	})

	return matches, nil
}

func buildVacancySearchQuery(v *VacancyWithDetails) string {
	var parts []string

	// Add title in English (or any available)
	for _, t := range v.Texts {
		if t.Lang == "en" && t.Title != "" {
			parts = append(parts, t.Title)
			break
		}
	}
	if len(parts) == 0 {
		for _, t := range v.Texts {
			if t.Title != "" {
				parts = append(parts, t.Title)
				break
			}
		}
	}

	// Add requirements in English
	for _, t := range v.Texts {
		if t.Lang == "en" && t.Requirements != "" {
			parts = append(parts, t.Requirements)
			break
		}
	}

	// Add skills
	skillNames := make([]string, 0, len(v.Skills))
	for _, sk := range v.Skills {
		skillNames = append(skillNames, sk.Name)
	}
	if len(skillNames) > 0 {
		parts = append(parts, strings.Join(skillNames, " "))
	}

	return strings.Join(parts, " ")
}

func (s *SearchService) calculateTotalExperience(ctx context.Context, userID uuid.UUID) float64 {
	items, err := s.experienceSvc.ListExperienceItemsByUser(ctx, userID)
	if err != nil || len(items) == 0 {
		return 0
	}

	var total float64
	now := time.Now()

	for _, item := range items {
		start := parseExperienceDate(item.StartDate)
		if start.IsZero() {
			continue
		}
		end := now
		if item.EndDate != "" && !isPresent(item.EndDate) {
			if parsed := parseExperienceDate(item.EndDate); !parsed.IsZero() {
				end = parsed
			}
		}
		years := end.Sub(start).Hours() / (24 * 365.25)
		if years > 0 {
			total += years
		}
	}

	return total
}

func parseExperienceDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}

	// Try common formats
	formats := []string{
		"2006-01",
		"2006-01-02",
		"January 2006",
		"Jan 2006",
		"2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}

	// Try just a year number
	if y, err := strconv.Atoi(s); err == nil && y > 1970 && y < 2100 {
		return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	// Try "Month YYYY" patterns with month names
	months := map[string]time.Month{
		"january": time.January, "february": time.February, "march": time.March,
		"april": time.April, "may": time.May, "june": time.June,
		"july": time.July, "august": time.August, "september": time.September,
		"october": time.October, "november": time.November, "december": time.December,
	}
	lower := strings.ToLower(s)
	for name, month := range months {
		if strings.Contains(lower, name) {
			parts := strings.Fields(s)
			for _, p := range parts {
				if y, err := strconv.Atoi(p); err == nil && y > 1970 {
					return time.Date(y, month, 1, 0, 0, 0, 0, time.UTC)
				}
			}
		}
	}

	return time.Time{}
}

func isPresent(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	return lower == "present" || lower == "current" || lower == "now" ||
		lower == "настоящее время" || lower == "hozirgi" || lower == "hozir"
}

func calculateExperienceFit(totalExp, expMin, expMax float64) float64 {
	// No experience requirements — everyone fits
	if expMin == 0 && expMax == 0 {
		// Reward having some experience
		if totalExp >= 1 {
			return 0.8
		}
		return 0.5
	}

	// Within range = perfect fit
	if totalExp >= expMin && (expMax == 0 || totalExp <= expMax) {
		return 1.0
	}

	// Below minimum
	if totalExp < expMin && expMin > 0 {
		ratio := totalExp / expMin
		if ratio < 0 {
			ratio = 0
		}
		return ratio * 0.8 // max 0.8 if below minimum
	}

	// Above maximum
	if expMax > 0 && totalExp > expMax {
		over := totalExp - expMax
		penalty := over / expMax
		score := 1.0 - penalty*0.3
		if score < 0.4 {
			score = 0.4
		}
		return score
	}

	return 0.5
}

func (s *SearchService) calculateSkillOverlap(ctx context.Context, userID uuid.UUID, vacancySkills map[string]bool) float64 {
	if len(vacancySkills) == 0 {
		return 0.5 // neutral if vacancy has no skills
	}
	userSkills, err := s.skillSvc.ListUserSkills(ctx, userID)
	if err != nil || len(userSkills) == 0 {
		return 0
	}
	matched := 0
	for _, sk := range userSkills {
		if vacancySkills[strings.ToLower(sk.Name)] {
			matched++
		}
	}
	return float64(matched) / float64(len(vacancySkills))
}
