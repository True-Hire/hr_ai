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
	vectorSearchSvc   *SearchService // optional: Qdrant vector search for pool enrichment
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

// SetVectorSearchService enables hybrid retrieval (tsvector + Qdrant vectors).
func (s *CandidateSearchService) SetVectorSearchService(svc *SearchService) {
	s.vectorSearchSvc = svc
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
	// Build query text from parsedQuery using normalized terms.
	// Keep concise to avoid overwhelming websearch_to_tsquery.
	var queryParts []string
	if parsedQuery.PrimaryRole != "" {
		normalized := scoring.NormalizeRole(parsedQuery.PrimaryRole)
		queryParts = append(queryParts, normalized)
	}
	// Add up to 5 top skills (most relevant for tsvector matching)
	skillLimit := 5
	if len(parsedQuery.Skills) < skillLimit {
		skillLimit = len(parsedQuery.Skills)
	}
	for _, sk := range parsedQuery.Skills[:skillLimit] {
		queryParts = append(queryParts, scoring.NormalizeSkill(sk))
	}
	for _, d := range parsedQuery.MustDomains {
		queryParts = append(queryParts, d)
	}
	queryText := strings.Join(queryParts, " ")

	// Determine role family and seniority from query or filters
	roleFamily := parsedQuery.RoleFamily
	if filters.RoleFamily != "" {
		roleFamily = filters.RoleFamily
	}
	const poolSize = 500

	// Retrieve candidate pool — use broad retrieval, then score/rank.
	// Location and experience are scoring penalties, NOT hard filters.
	var pool []domain.CandidateSearchProfile
	var err error

	if queryText != "" {
		// Only use roleFamily for pool retrieval, not location/seniority (those affect scoring)
		pool, err = s.searchProfileRepo.SearchPool(ctx, queryText, roleFamily, "", "", "", poolSize)
	} else if len(filters.Skills) > 0 {
		pool, err = s.searchProfileRepo.SearchPoolBySkills(ctx, filters.Skills, roleFamily, "", poolSize)
	} else if len(parsedQuery.Skills) > 0 {
		pool, err = s.searchProfileRepo.SearchPoolBySkills(ctx, parsedQuery.Skills, roleFamily, "", poolSize)
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

	// Hybrid retrieval: enrich pool with Qdrant vector search results
	if s.vectorSearchSvc != nil && queryText != "" {
		pool = s.enrichPoolWithVectorSearch(ctx, pool, queryText, poolSize)
	}

	if len(pool) == 0 {
		return &SearchSessionPage{
			SearchID:   uuid.Nil,
			Items:      nil,
			NextRank:   0,
			TotalCount: 0,
		}, nil
	}

	// Determine target location for scoring penalty
	targetLocation := parsedQuery.LocationCity
	if filters.LocationCity != "" {
		targetLocation = filters.LocationCity
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

		// Apply location penalty (soft — not a hard filter)
		locationPenalty := 1.0
		if targetLocation != "" && !strings.EqualFold(c.LocationCity, targetLocation) {
			if c.WillingToRelocate {
				locationPenalty = 0.92
			} else {
				locationPenalty = 0.85
			}
		}
		finalScore *= locationPenalty

		// Apply experience fit penalty (soft — not a hard filter)
		if filters.MinExperience > 0 || filters.MaxExperience > 0 {
			expMonths := c.TotalExperienceMonths
			minExp := filters.MinExperience
			maxExp := filters.MaxExperience
			if minExp > 0 && expMonths < minExp {
				// Below minimum: scale penalty by how far below
				ratio := float64(expMonths) / float64(minExp)
				if ratio < 0 {
					ratio = 0
				}
				finalScore *= 0.70 + 0.30*ratio // at 0 exp → 0.70x, at min → 1.0x
			} else if maxExp > 0 && expMonths > maxExp {
				// Above maximum: mild penalty
				over := float64(expMonths-maxExp) / float64(maxExp)
				penalty := 1.0 - over*0.15
				if penalty < 0.75 {
					penalty = 0.75
				}
				finalScore *= penalty
			}
		}

		isLocationMatch := targetLocation == "" || strings.EqualFold(c.LocationCity, targetLocation)
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
			"location_match":  isLocationMatch,
			"location_city":   c.LocationCity,
			"final_score":     finalScore,
		}

		scored = append(scored, scoredEntry{
			profile:    c,
			finalScore: finalScore,
			breakdown:  breakdown,
		})
	}

	// Sort: location matches first, then by finalScore DESC, then UserID ASC
	sort.Slice(scored, func(i, j int) bool {
		iLocal := targetLocation != "" && strings.EqualFold(scored[i].profile.LocationCity, targetLocation)
		jLocal := targetLocation != "" && strings.EqualFold(scored[j].profile.LocationCity, targetLocation)
		if iLocal != jLocal {
			return iLocal // local candidates first
		}
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

// enrichPoolWithVectorSearch fetches candidates from Qdrant vector search
// and merges them into the tsvector pool, deduplicating by user ID.
// This gives us hybrid retrieval: keyword matching + semantic similarity.
func (s *CandidateSearchService) enrichPoolWithVectorSearch(ctx context.Context, pool []domain.CandidateSearchProfile, queryText string, poolSize int) []domain.CandidateSearchProfile {
	// Collect existing user IDs for dedup
	existing := make(map[uuid.UUID]bool, len(pool))
	for _, p := range pool {
		existing[p.UserID] = true
	}

	// Qdrant vector search — use the old SearchService which handles translate + embed + search
	vectorLimit := poolSize / 2
	if vectorLimit < 50 {
		vectorLimit = 50
	}
	vectorResults, err := s.vectorSearchSvc.SearchUsers(ctx, queryText, vectorLimit)
	if err != nil {
		// Vector search failed — just use tsvector pool, don't block
		return pool
	}

	// For each vector result not already in pool, load their search profile
	var newUserIDs []uuid.UUID
	for _, vr := range vectorResults {
		if !existing[vr.UserID] {
			newUserIDs = append(newUserIDs, vr.UserID)
			existing[vr.UserID] = true
		}
	}

	// Load search profiles for new candidates
	for _, uid := range newUserIDs {
		profile, err := s.searchProfileRepo.GetByUserID(ctx, uid)
		if err != nil || profile == nil {
			continue
		}
		pool = append(pool, *profile)
	}

	return pool
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

	// Determine seniority from experience range
	avgExpMonths := int((vacancy.Vacancy.ExperienceMin + vacancy.Vacancy.ExperienceMax) / 2 * 12)
	if avgExpMonths > 0 {
		parsedQuery.Seniority = scoring.DetermineSeniority(avgExpMonths, false)
	}

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

	// Extract role from English title and normalize
	var rawTitle string
	for _, t := range v.Texts {
		if t.Lang == "en" && t.Title != "" {
			rawTitle = t.Title
			break
		}
	}
	if rawTitle == "" {
		for _, t := range v.Texts {
			if t.Title != "" {
				rawTitle = t.Title
				break
			}
		}
	}
	pq.PrimaryRole = scoring.NormalizeRole(rawTitle)
	pq.RoleFamily = scoring.RoleFamily(pq.PrimaryRole)

	// Extract skills from vacancy skills
	skills := make([]string, 0, len(v.Skills))
	for _, sk := range v.Skills {
		skills = append(skills, sk.Name)
	}
	pq.Skills = skills

	// Extract domains from English requirements/description using domain keyword detection
	for _, t := range v.Texts {
		if t.Lang == "en" {
			var domainTexts []string
			if t.Requirements != "" {
				domainTexts = append(domainTexts, t.Requirements)
			}
			if t.Description != "" {
				domainTexts = append(domainTexts, t.Description)
			}
			if len(domainTexts) > 0 {
				pq.PreferredDomains = scoring.ExtractDomains(domainTexts...)
			}
			break
		}
	}

	return pq
}
