package application

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application/scoring"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CandidateSearchService struct {
	searchProfileRepo domain.CandidateSearchProfileRepository
	searchSessionRepo domain.SearchSessionRepository
	userSvc           *UserService
	skillSvc          *SkillService
	vacancySvc        *VacancyService
	experienceSvc     *ExperienceItemService
}

func NewCandidateSearchService(
	searchProfileRepo domain.CandidateSearchProfileRepository,
	searchSessionRepo domain.SearchSessionRepository,
	userSvc *UserService,
	skillSvc *SkillService,
	vacancySvc *VacancyService,
	experienceSvc *ExperienceItemService,
) *CandidateSearchService {
	return &CandidateSearchService{
		searchProfileRepo: searchProfileRepo,
		searchSessionRepo: searchSessionRepo,
		userSvc:           userSvc,
		skillSvc:          skillSvc,
		vacancySvc:        vacancySvc,
		experienceSvc:     experienceSvc,
	}
}

type SearchFilters struct {
	LocationCity    string
	LocationCountry string
	Seniority       string
	RoleFamily      string
	Skills          []string
	MinExperience   int // months
	MaxExperience   int // months
}

type SearchSessionPage struct {
	SearchID   uuid.UUID
	Items      []ScoredCandidate
	NextRank   int
	TotalCount int
}

type ScoredCandidate struct {
	UserID               uuid.UUID
	FinalScore           float64
	MatchPercentage      int
	TotalExperienceYears float64
	ScoreBreakdown       map[string]interface{}
}

// Search performs candidate search using parsed query and filters, returning paginated results.
func (s *CandidateSearchService) Search(ctx context.Context, hrID uuid.UUID, parsedQuery scoring.ParsedQuery, filters SearchFilters, pageSize int) (*SearchSessionPage, error) {
	// Build query text from parsedQuery
	var queryParts []string
	if parsedQuery.PrimaryRole != "" {
		queryParts = append(queryParts, parsedQuery.PrimaryRole)
	}
	for _, sk := range parsedQuery.Skills {
		queryParts = append(queryParts, sk)
	}
	for _, d := range parsedQuery.MustDomains {
		queryParts = append(queryParts, d)
	}
	for _, d := range parsedQuery.PreferredDomains {
		queryParts = append(queryParts, d)
	}
	queryText := strings.Join(queryParts, " ")

	// Determine role family and seniority from query or filters
	roleFamily := parsedQuery.RoleFamily
	if filters.RoleFamily != "" {
		roleFamily = filters.RoleFamily
	}
	seniority := parsedQuery.Seniority
	if filters.Seniority != "" {
		seniority = filters.Seniority
	}

	locationCity := parsedQuery.LocationCity
	if filters.LocationCity != "" {
		locationCity = filters.LocationCity
	}

	const poolSize = 500

	// Retrieve candidate pool
	var pool []domain.CandidateSearchProfile
	var err error

	if queryText != "" {
		pool, err = s.searchProfileRepo.SearchPool(ctx, queryText, roleFamily, seniority, locationCity, filters.LocationCountry, poolSize)
	} else if len(filters.Skills) > 0 {
		pool, err = s.searchProfileRepo.SearchPoolBySkills(ctx, filters.Skills, roleFamily, locationCity, poolSize)
	} else if len(parsedQuery.Skills) > 0 {
		pool, err = s.searchProfileRepo.SearchPoolBySkills(ctx, parsedQuery.Skills, roleFamily, locationCity, poolSize)
	} else {
		return &SearchSessionPage{
			SearchID:   uuid.Nil,
			Items:      nil,
			NextRank:   0,
			TotalCount: 0,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("search pool: %w", err)
	}
	if len(pool) == 0 {
		return &SearchSessionPage{
			SearchID:   uuid.Nil,
			Items:      nil,
			NextRank:   0,
			TotalCount: 0,
		}, nil
	}

	// Filter by experience range if specified
	if filters.MinExperience > 0 || filters.MaxExperience > 0 {
		filtered := make([]domain.CandidateSearchProfile, 0, len(pool))
		for _, c := range pool {
			if filters.MinExperience > 0 && c.TotalExperienceMonths < filters.MinExperience {
				continue
			}
			if filters.MaxExperience > 0 && c.TotalExperienceMonths > filters.MaxExperience {
				continue
			}
			filtered = append(filtered, c)
		}
		pool = filtered
	}

	// Score each candidate
	type scoredEntry struct {
		profile    domain.CandidateSearchProfile
		finalScore float64
		breakdown  map[string]interface{}
	}

	scored := make([]scoredEntry, 0, len(pool))
	for _, c := range pool {
		roleMatch := scoring.CalcRoleMatchScore(parsedQuery.PrimaryRole, parsedQuery.RoleFamily, c.PrimaryRole, c.RoleFamily)
		domainMatch := scoring.CalcDomainMatchScore(parsedQuery.MustDomains, parsedQuery.PreferredDomains, c.ProjectDomains)
		skillMatch := scoring.CalcSkillMatchScore(parsedQuery.Skills, c.Skills)
		educationMatch := scoring.CalcEducationMatchScore(parsedQuery.PreferredEducationFields, c.EducationFields)
		seniorityMatch := scoring.CalcSeniorityMatchScore(parsedQuery.Seniority, c.Seniority)
		textRank := 0.5 // base score — already filtered by tsvector
		queryRelevance := scoring.CalcQueryRelevanceScore(roleMatch, domainMatch, skillMatch, educationMatch, seniorityMatch, textRank)

		roleBonus := scoring.CalcRoleSpecificBonusScore(
			parsedQuery.RoleFamily,
			c.DevOpsSupportScore,
			c.OwnershipScore,
			c.EngineeringEnvironmentScore,
			c.ProjectManagementScore,
			c.ClientCommunicationScore,
			c.LeadershipScore,
		)

		marketStrength := scoring.CalcMarketStrengthScore(
			c.CompanyPrestigeScore,
			c.ProjectComplexityScore,
			c.InternshipQualityScore,
			c.EducationQualityScore,
			c.GrowthTrajectoryScore,
			c.CompetitionScore,
		)

		finalScore := scoring.CalcFinalScore(queryRelevance, roleBonus, marketStrength)

		// Apply location penalty
		if parsedQuery.LocationCity != "" && !strings.EqualFold(c.LocationCity, parsedQuery.LocationCity) {
			if c.WillingToRelocate {
				finalScore *= 0.90
			} else {
				finalScore *= 0.80
			}
		}

		breakdown := map[string]interface{}{
			"role_match":      roleMatch,
			"domain_match":    domainMatch,
			"skill_match":     skillMatch,
			"education_match": educationMatch,
			"seniority_match": seniorityMatch,
			"text_rank":       textRank,
			"query_relevance": queryRelevance,
			"role_bonus":      roleBonus,
			"market_strength": marketStrength,
			"final_score":     finalScore,
		}

		scored = append(scored, scoredEntry{
			profile:    c,
			finalScore: finalScore,
			breakdown:  breakdown,
		})
	}

	// Sort by finalScore DESC, then UserID ASC for deterministic ordering
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].finalScore != scored[j].finalScore {
			return scored[i].finalScore > scored[j].finalScore
		}
		return scored[i].profile.UserID.String() < scored[j].profile.UserID.String()
	})

	// Create search session
	session, err := s.searchSessionRepo.Create(ctx, &domain.SearchSession{
		ID:        uuid.New(),
		HRID:      hrID,
		QueryText: queryText,
		ParsedQuery: map[string]interface{}{
			"role_family":     parsedQuery.RoleFamily,
			"primary_role":    parsedQuery.PrimaryRole,
			"skills":          parsedQuery.Skills,
			"must_domains":    parsedQuery.MustDomains,
			"seniority":       parsedQuery.Seniority,
			"location_city":   parsedQuery.LocationCity,
		},
		Filters: map[string]interface{}{
			"location_city":    filters.LocationCity,
			"location_country": filters.LocationCountry,
			"seniority":        filters.Seniority,
			"role_family":      filters.RoleFamily,
			"skills":           filters.Skills,
			"min_experience":   filters.MinExperience,
			"max_experience":   filters.MaxExperience,
		},
		TotalResults: len(scored),
		Status:       "active",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		return nil, fmt.Errorf("create search session: %w", err)
	}

	// Build SearchSessionResult list
	results := make([]domain.SearchSessionResult, 0, len(scored))
	for i, entry := range scored {
		results = append(results, domain.SearchSessionResult{
			SearchID:       session.ID,
			Rank:           i + 1,
			UserID:         entry.profile.UserID,
			FinalScore:     entry.finalScore,
			ScoreBreakdown: entry.breakdown,
		})
	}

	if err := s.searchSessionRepo.InsertResults(ctx, session.ID, results); err != nil {
		return nil, fmt.Errorf("insert search results: %w", err)
	}

	// Return first page
	end := pageSize
	if end > len(scored) {
		end = len(scored)
	}

	items := make([]ScoredCandidate, 0, end)
	for _, entry := range scored[:end] {
		totalExp := float64(entry.profile.TotalExperienceMonths) / 12.0
		matchPct := int(math.Round(entry.finalScore * 100))
		if matchPct > 99 {
			matchPct = 99
		}
		if matchPct < 1 && entry.finalScore > 0 {
			matchPct = 1
		}

		items = append(items, ScoredCandidate{
			UserID:               entry.profile.UserID,
			FinalScore:           entry.finalScore,
			MatchPercentage:      matchPct,
			TotalExperienceYears: math.Round(totalExp*10) / 10,
			ScoreBreakdown:       entry.breakdown,
		})
	}

	nextRank := 0
	if end < len(scored) {
		nextRank = end + 1
	}

	return &SearchSessionPage{
		SearchID:   session.ID,
		Items:      items,
		NextRank:   nextRank,
		TotalCount: len(scored),
	}, nil
}

// GetPage returns a page of results from an existing search session.
func (s *CandidateSearchService) GetPage(ctx context.Context, searchID uuid.UUID, afterRank int, pageSize int) (*SearchSessionPage, error) {
	session, err := s.searchSessionRepo.GetByID(ctx, searchID)
	if err != nil {
		return nil, fmt.Errorf("get search session: %w", err)
	}

	results, err := s.searchSessionRepo.GetResultsPage(ctx, searchID, afterRank, pageSize)
	if err != nil {
		return nil, fmt.Errorf("get results page: %w", err)
	}

	items := make([]ScoredCandidate, 0, len(results))
	for _, r := range results {
		matchPct := int(math.Round(r.FinalScore * 100))
		if matchPct > 99 {
			matchPct = 99
		}
		if matchPct < 1 && r.FinalScore > 0 {
			matchPct = 1
		}

		// Retrieve total experience from search profile
		var totalExpYears float64
		profile, err := s.searchProfileRepo.GetByUserID(ctx, r.UserID)
		if err == nil && profile != nil {
			totalExpYears = math.Round(float64(profile.TotalExperienceMonths)/12.0*10) / 10
		}

		items = append(items, ScoredCandidate{
			UserID:               r.UserID,
			FinalScore:           r.FinalScore,
			MatchPercentage:      matchPct,
			TotalExperienceYears: totalExpYears,
			ScoreBreakdown:       r.ScoreBreakdown,
		})
	}

	nextRank := 0
	if len(results) == pageSize {
		lastRank := afterRank + len(results)
		if lastRank < session.TotalResults {
			nextRank = lastRank + 1
		}
	}

	return &SearchSessionPage{
		SearchID:   searchID,
		Items:      items,
		NextRank:   nextRank,
		TotalCount: session.TotalResults,
	}, nil
}

// SearchByVacancy finds candidates matching a specific vacancy.
func (s *CandidateSearchService) SearchByVacancy(ctx context.Context, vacancyID uuid.UUID, hrID uuid.UUID, pageSize int) (*SearchSessionPage, error) {
	vacancy, err := s.vacancySvc.GetVacancy(ctx, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("get vacancy: %w", err)
	}

	// Build ParsedQuery from vacancy data
	parsedQuery := buildParsedQueryFromVacancy(vacancy)

	// Build SearchFilters from vacancy data
	filters := SearchFilters{
		MinExperience: int(vacancy.Vacancy.ExperienceMin) * 12, // years to months
		MaxExperience: int(vacancy.Vacancy.ExperienceMax) * 12,
	}

	// Extract location from vacancy address
	if vacancy.Vacancy.Address != "" {
		parsedQuery.LocationCity = vacancy.Vacancy.Address
		filters.LocationCity = vacancy.Vacancy.Address
	}

	return s.Search(ctx, hrID, parsedQuery, filters, pageSize)
}

// CountMatchingByVacancy returns a quick count of candidates matching a vacancy.
func (s *CandidateSearchService) CountMatchingByVacancy(ctx context.Context, vacancyID uuid.UUID) int {
	vacancy, err := s.vacancySvc.GetVacancy(ctx, vacancyID)
	if err != nil {
		return 0
	}

	// Build search text from title + skills
	var parts []string
	for _, t := range vacancy.Texts {
		if t.Lang == "en" && t.Title != "" {
			parts = append(parts, t.Title)
			break
		}
	}
	if len(parts) == 0 {
		for _, t := range vacancy.Texts {
			if t.Title != "" {
				parts = append(parts, t.Title)
				break
			}
		}
	}

	skillNames := make([]string, 0, len(vacancy.Skills))
	for _, sk := range vacancy.Skills {
		skillNames = append(skillNames, sk.Name)
	}
	if len(skillNames) > 0 {
		parts = append(parts, strings.Join(skillNames, " "))
	}

	searchText := strings.Join(parts, " ")
	if searchText == "" && len(skillNames) == 0 {
		return 0
	}

	const poolSize = 100
	var pool []domain.CandidateSearchProfile

	if searchText != "" {
		pool, err = s.searchProfileRepo.SearchPool(ctx, searchText, "", "", "", "", poolSize)
	} else {
		pool, err = s.searchProfileRepo.SearchPoolBySkills(ctx, skillNames, "", "", poolSize)
	}
	if err != nil {
		return 0
	}

	// Count candidates with a reasonable score
	parsedQuery := buildParsedQueryFromVacancy(vacancy)
	count := 0
	for _, c := range pool {
		roleMatch := scoring.CalcRoleMatchScore(parsedQuery.PrimaryRole, parsedQuery.RoleFamily, c.PrimaryRole, c.RoleFamily)
		skillMatch := scoring.CalcSkillMatchScore(parsedQuery.Skills, c.Skills)
		avgScore := (roleMatch + skillMatch) / 2.0
		if avgScore > 0.3 {
			count++
		}
	}

	return count
}

// buildParsedQueryFromVacancy constructs a ParsedQuery from vacancy details.
func buildParsedQueryFromVacancy(v *VacancyWithDetails) scoring.ParsedQuery {
	var pq scoring.ParsedQuery

	// Extract role from English title
	for _, t := range v.Texts {
		if t.Lang == "en" && t.Title != "" {
			pq.PrimaryRole = t.Title
			break
		}
	}
	if pq.PrimaryRole == "" {
		for _, t := range v.Texts {
			if t.Title != "" {
				pq.PrimaryRole = t.Title
				break
			}
		}
	}

	// Extract skills from vacancy skills
	skills := make([]string, 0, len(v.Skills))
	for _, sk := range v.Skills {
		skills = append(skills, sk.Name)
	}
	pq.Skills = skills

	// Extract domains from English requirements/description
	for _, t := range v.Texts {
		if t.Lang == "en" {
			var domains []string
			if t.Requirements != "" {
				domains = append(domains, t.Requirements)
			}
			if t.Description != "" {
				domains = append(domains, t.Description)
			}
			if len(domains) > 0 {
				pq.PreferredDomains = domains
			}
			break
		}
	}

	return pq
}
