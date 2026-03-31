package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application/scoring"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CandidateIndexingService struct {
	searchProfileRepo domain.CandidateSearchProfileRepository
	companyRefRepo    domain.CompanyReferenceRepository
	universityRefRepo domain.UniversityReferenceRepository
	userSvc           *UserService
	profileFieldSvc   *ProfileFieldService
	profileTextSvc    *ProfileFieldTextService
	experienceSvc     *ExperienceItemService
	educationSvc      *EducationItemService
	itemTextSvc       *ItemTextService
	skillSvc          *SkillService
	normSvc           *NormalizationService
}

func NewCandidateIndexingService(
	searchProfileRepo domain.CandidateSearchProfileRepository,
	companyRefRepo domain.CompanyReferenceRepository,
	universityRefRepo domain.UniversityReferenceRepository,
	userSvc *UserService,
	profileFieldSvc *ProfileFieldService,
	profileTextSvc *ProfileFieldTextService,
	experienceSvc *ExperienceItemService,
	educationSvc *EducationItemService,
	itemTextSvc *ItemTextService,
	skillSvc *SkillService,
) *CandidateIndexingService {
	return &CandidateIndexingService{
		searchProfileRepo: searchProfileRepo,
		companyRefRepo:    companyRefRepo,
		universityRefRepo: universityRefRepo,
		userSvc:           userSvc,
		profileFieldSvc:   profileFieldSvc,
		profileTextSvc:    profileTextSvc,
		experienceSvc:     experienceSvc,
		educationSvc:      educationSvc,
		itemTextSvc:       itemTextSvc,
		skillSvc:          skillSvc,
	}
}

// SetNormalizationService enables auto-discovery of new normalization rules.
func (s *CandidateIndexingService) SetNormalizationService(normSvc *NormalizationService) {
	s.normSvc = normSvc
}

// IndexCandidate computes all signal scores for a user and upserts the candidate search profile.
func (s *CandidateIndexingService) IndexCandidate(ctx context.Context, userID uuid.UUID) error {
	// 1. Load user
	user, err := s.userSvc.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("indexing: get user: %w", err)
	}

	// 2. Load skills
	skills, err := s.skillSvc.ListUserSkills(ctx, userID)
	if err != nil {
		return fmt.Errorf("indexing: list skills: %w", err)
	}

	// 3. Load experience items
	experienceItems, err := s.experienceSvc.ListExperienceItemsByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("indexing: list experience items: %w", err)
	}

	// 4. Load education items
	educationItems, err := s.educationSvc.ListEducationItemsByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("indexing: list education items: %w", err)
	}

	// 5. Load profile fields and their texts
	profileFields, err := s.profileFieldSvc.ListProfileFieldsByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("indexing: list profile fields: %w", err)
	}

	fieldTexts := make(map[string]string) // fieldName -> English content
	for _, pf := range profileFields {
		texts, err := s.profileTextSvc.ListProfileFieldTexts(ctx, pf.ID)
		if err != nil {
			continue
		}
		fieldTexts[pf.FieldName] = preferEnglish(texts)
	}

	// 6. Load experience descriptions (English preferred)
	var experienceTexts []string
	for _, item := range experienceItems {
		texts, err := s.itemTextSvc.ListItemTextsByItem(ctx, item.ID, "experience")
		if err != nil {
			continue
		}
		if content := preferEnglishItemText(texts); content != "" {
			experienceTexts = append(experienceTexts, content)
		}
	}

	// 7. Load education descriptions (English preferred)
	var educationTexts []string
	for _, item := range educationItems {
		texts, err := s.itemTextSvc.ListItemTextsByItem(ctx, item.ID, "education")
		if err != nil {
			continue
		}
		if content := preferEnglishItemText(texts); content != "" {
			educationTexts = append(educationTexts, content)
		}
	}

	// 8. Normalize

	// Determine title
	title := fieldTexts["title"]
	if title == "" && len(user.Specializations) > 0 {
		title = user.Specializations[0]
	}
	if title == "" && len(experienceItems) > 0 {
		title = experienceItems[0].Position
	}

	role := scoring.NormalizeRole(title)
	roleFamily := scoring.RoleFamily(role)

	// Auto-discover: persist role mapping if new
	if s.normSvc != nil && title != "" {
		s.normSvc.EnsureNormalized(ctx, "role", strings.ToLower(title), role)
		s.normSvc.EnsureNormalized(ctx, "role_family", strings.ToLower(role), roleFamily)
	}

	// Normalize skills
	normalizedSkills := make([]string, 0, len(skills))
	seen := make(map[string]bool)
	for _, sk := range skills {
		ns := scoring.NormalizeSkill(sk.Name)
		if !seen[ns] {
			normalizedSkills = append(normalizedSkills, ns)
			seen[ns] = true
		}
		// Auto-discover: persist skill mapping if new
		if s.normSvc != nil {
			s.normSvc.EnsureNormalized(ctx, "skill", strings.ToLower(sk.Name), ns)
		}
	}

	// Calculate total experience months
	totalExpMonths := calculateTotalExperienceMonths(experienceItems)

	// Determine leadership evidence from title
	hasLeadership := scoring.CalcLeadershipScore(title, experienceTexts) > 0.3
	seniority := scoring.DetermineSeniority(totalExpMonths, hasLeadership)

	// Extract domains from experience + about texts
	aboutText := fieldTexts["about"]
	domainInputTexts := append(experienceTexts, aboutText)
	domains := scoring.ExtractDomains(domainInputTexts...)

	// Normalize company names
	var normalizedCompanies []string
	companySet := make(map[string]bool)
	for _, item := range experienceItems {
		if item.Company == "" {
			continue
		}
		nc := scoring.NormalizeCompany(item.Company)
		if !companySet[nc] {
			normalizedCompanies = append(normalizedCompanies, nc)
			companySet[nc] = true
		}
		// Auto-discover: persist company mapping if new
		if s.normSvc != nil {
			s.normSvc.EnsureNormalized(ctx, "company", strings.ToLower(item.Company), nc)
		}
	}

	// University names and education fields
	var universityNames []string
	var educationFields []string
	uniSet := make(map[string]bool)
	fieldSet := make(map[string]bool)
	for _, item := range educationItems {
		if item.Institution != "" && !uniSet[item.Institution] {
			universityNames = append(universityNames, item.Institution)
			uniSet[item.Institution] = true
		}
		if item.FieldOfStudy != "" && !fieldSet[item.FieldOfStudy] {
			educationFields = append(educationFields, item.FieldOfStudy)
			fieldSet[item.FieldOfStudy] = true
		}
	}

	// 9. Look up company references
	now := time.Now()
	var companyPrestigeInputs []scoring.CompanyPrestigeInput
	var companyEngineeringScores []float64
	var companyPrestigeScores []float64
	var companyCategories []string
	var internships []scoring.CompanyPrestigeInput

	for i, item := range experienceItems {
		nc := scoring.NormalizeCompany(item.Company)
		ref, err := s.companyRefRepo.GetByNormalizedName(ctx, nc)
		if err != nil || ref == nil {
			continue
		}

		start := indexParseExperienceDate(item.StartDate)
		end := now
		if item.EndDate != "" && !indexIsPresent(item.EndDate) {
			if parsed := indexParseExperienceDate(item.EndDate); !parsed.IsZero() {
				end = parsed
			}
		}
		durationMonths := int(math.Round(end.Sub(start).Hours() / (24 * 30.44)))
		if durationMonths < 0 {
			durationMonths = 0
		}

		isInternship := strings.Contains(strings.ToLower(item.Position), "intern")
		isCurrent := item.EndDate == "" || indexIsPresent(item.EndDate)

		input := scoring.CompanyPrestigeInput{
			PrestigeScore:    ref.PrestigeScore,
			EngineeringScore: ref.EngineeringScore,
			HiringBarScore:   ref.HiringBarScore,
			ScaleScore:       ref.ScaleScore,
			DurationMonths:   durationMonths,
			IsRecent:         i == 0,
			IsCurrent:        isCurrent,
			IsInternship:     isInternship,
		}
		companyPrestigeInputs = append(companyPrestigeInputs, input)
		companyEngineeringScores = append(companyEngineeringScores, ref.EngineeringScore)
		companyPrestigeScores = append(companyPrestigeScores, ref.PrestigeScore)
		if ref.Category != "" {
			companyCategories = append(companyCategories, ref.Category)
		}

		if isInternship {
			internships = append(internships, input)
		}
	}

	// 10. Look up university references
	var institutionScores []float64
	for _, uniName := range universityNames {
		normalized := strings.ToLower(strings.TrimSpace(uniName))
		ref, err := s.universityRefRepo.GetByNormalizedName(ctx, normalized)
		if err != nil || ref == nil {
			continue
		}
		institutionScores = append(institutionScores, ref.EducationScore)
	}

	// 11. Compute signal scores

	// Positions list for growth trajectory
	var positions []string
	for _, item := range experienceItems {
		if item.Position != "" {
			positions = append(positions, item.Position)
		}
	}

	achievements := fieldTexts["achievements"]

	// Role scores
	backendScore := scoring.CalcBackendScore(title, normalizedSkills, experienceTexts)
	frontendScore := scoring.CalcFrontendScore(title, normalizedSkills, experienceTexts)
	mobileScore := scoring.CalcMobileScore(title, normalizedSkills, experienceTexts)
	dataScore := scoring.CalcDataScore(title, normalizedSkills, experienceTexts)
	qaScore := scoring.CalcQAScore(title, normalizedSkills, experienceTexts)
	pmScore := scoring.CalcPMScore(title, normalizedSkills, experienceTexts)
	devOpsScore := scoring.CalcDevOpsRoleScore(title, normalizedSkills, experienceTexts)
	designScore := scoring.CalcDesignScore(title, normalizedSkills, experienceTexts)

	// Capability scores
	devOpsSupportScore := scoring.CalcDevOpsSupportScore(normalizedSkills, experienceTexts)
	clientCommunicationScore := scoring.CalcClientCommunicationScore(experienceTexts)
	projectManagementScore := scoring.CalcProjectManagementScore(experienceTexts)
	ownershipScore := scoring.CalcOwnershipScore(experienceTexts)
	leadershipScore := scoring.CalcLeadershipScore(title, experienceTexts)
	mentoringScore := scoring.CalcMentoringScore(experienceTexts)
	startupAdaptabilityScore := scoring.CalcStartupAdaptabilityScore(experienceTexts, companyCategories)

	// Market strength scores
	companyPrestigeScore := scoring.CalcCompanyPrestigeScore(companyPrestigeInputs)
	engineeringEnvironmentScore := scoring.CalcEngineeringEnvironmentScore(experienceTexts, companyEngineeringScores)
	internshipQualityScore := scoring.CalcInternshipQualityScore(internships, experienceTexts)
	educationQualityScore := scoring.CalcEducationQualityScore(institutionScores, educationFields, achievements)
	competitionScore := scoring.CalcCompetitionScore(achievements, "")
	openSourceScore := scoring.CalcOpenSourceScore(experienceTexts, achievements)
	growthTrajectoryScore := scoring.CalcGrowthTrajectoryScore(positions, companyPrestigeScores)
	projectComplexityScore := scoring.CalcProjectComplexityScore(experienceTexts)

	// 12. Aggregated strength scores
	overallStrength := scoring.CalcOverallStrength(
		companyPrestigeScore, engineeringEnvironmentScore, projectComplexityScore,
		ownershipScore, leadershipScore, educationQualityScore,
		internshipQualityScore, competitionScore, openSourceScore, growthTrajectoryScore,
	)
	backendStrength := scoring.CalcBackendStrength(
		backendScore, engineeringEnvironmentScore, projectComplexityScore,
		ownershipScore, devOpsSupportScore, companyPrestigeScore,
		internshipQualityScore, clientCommunicationScore, projectManagementScore,
	)
	frontendStrength := scoring.CalcFrontendStrength(
		frontendScore, projectComplexityScore, ownershipScore,
		clientCommunicationScore, designScore, companyPrestigeScore,
	)
	dataStrength := scoring.CalcDataStrength(
		dataScore, projectComplexityScore, engineeringEnvironmentScore,
		ownershipScore, companyPrestigeScore,
	)

	// 13. Build search text
	searchText := scoring.BuildSearchText(
		role, roleFamily, seniority, normalizedSkills, domains,
		normalizedCompanies, universityNames, educationFields,
		user.Region, totalExpMonths,
	)

	// Determine highest education level
	highestEducation := ""
	degreeRank := map[string]int{
		"phd": 4, "doctorate": 4,
		"master": 3, "магистр": 3,
		"bachelor": 2, "бакалавр": 2,
		"associate": 1,
	}
	bestRank := 0
	for _, item := range educationItems {
		lower := strings.ToLower(item.Degree)
		for kw, rank := range degreeRank {
			if strings.Contains(lower, kw) && rank > bestRank {
				bestRank = rank
				highestEducation = item.Degree
			}
		}
	}

	// Known languages: from user.Language + profile field "languages"
	var knownLanguages []string
	langSet := make(map[string]bool)
	if user.Language != "" {
		knownLanguages = append(knownLanguages, user.Language)
		langSet[user.Language] = true
	}
	if langJSON := fieldTexts["languages"]; langJSON != "" {
		// Languages are stored as JSON array: [{"name":"English","level":"B2"}, ...]
		type langEntry struct {
			Name  string `json:"name"`
			Level string `json:"level"`
		}
		var langs []langEntry
		if err := json.Unmarshal([]byte(langJSON), &langs); err == nil {
			for _, l := range langs {
				name := strings.ToLower(strings.TrimSpace(l.Name))
				if name != "" && !langSet[name] {
					knownLanguages = append(knownLanguages, l.Name)
					langSet[name] = true
				}
			}
		}
	}

	// 14. Location
	locationCity := user.Region
	locationCountry := user.Country

	// 15. Upsert
	profile := &domain.CandidateSearchProfile{
		UserID:                userID,
		PrimaryRole:           role,
		RoleFamily:            roleFamily,
		Seniority:             seniority,
		TotalExperienceMonths: totalExpMonths,
		HighestEducationLevel: highestEducation,

		Skills:          normalizedSkills,
		Industries:      domains,
		ProjectDomains:  domains,
		CompanyNames:    normalizedCompanies,
		KnownLanguages:  knownLanguages,
		EducationFields: educationFields,
		Universities:    universityNames,

		LocationCity:    locationCity,
		LocationCountry: locationCountry,

		BackendScore:  backendScore,
		FrontendScore: frontendScore,
		MobileScore:   mobileScore,
		DataScore:     dataScore,
		QAScore:       qaScore,
		PMScore:       pmScore,
		DevOpsScore:   devOpsScore,
		DesignScore:   designScore,

		DevOpsSupportScore:       devOpsSupportScore,
		ClientCommunicationScore: clientCommunicationScore,
		ProjectManagementScore:   projectManagementScore,
		OwnershipScore:           ownershipScore,
		LeadershipScore:          leadershipScore,
		MentoringScore:           mentoringScore,
		StartupAdaptabilityScore: startupAdaptabilityScore,

		CompanyPrestigeScore:        companyPrestigeScore,
		EngineeringEnvironmentScore: engineeringEnvironmentScore,
		InternshipQualityScore:      internshipQualityScore,
		EducationQualityScore:       educationQualityScore,
		CompetitionScore:            competitionScore,
		OpenSourceScore:             openSourceScore,
		GrowthTrajectoryScore:       growthTrajectoryScore,
		ProjectComplexityScore:      projectComplexityScore,

		OverallStrengthScore:  overallStrength,
		BackendStrengthScore:  backendStrength,
		FrontendStrengthScore: frontendStrength,
		DataStrengthScore:     dataStrength,

		SearchText: searchText,

		ScoringFactors: map[string]interface{}{
			"title":            title,
			"normalized_role":  role,
			"role_family":      roleFamily,
			"seniority":        seniority,
			"total_exp_months": totalExpMonths,
			"skill_count":      len(normalizedSkills),
			"company_count":    len(normalizedCompanies),
			"education_count":  len(educationItems),
		},
		ParsedEntities: map[string]interface{}{
			"skills":      normalizedSkills,
			"companies":   normalizedCompanies,
			"domains":     domains,
			"universities": universityNames,
		},

		UpdatedAt: time.Now(),
	}

	if err := s.searchProfileRepo.Upsert(ctx, profile); err != nil {
		return fmt.Errorf("indexing: upsert search profile: %w", err)
	}

	return nil
}

// ReindexAll lists all users page by page and re-indexes each candidate.
func (s *CandidateIndexingService) ReindexAll(ctx context.Context) error {
	const pageSize int32 = 100
	var page int32 = 1

	for {
		result, err := s.userSvc.ListUsers(ctx, page, pageSize)
		if err != nil {
			return fmt.Errorf("indexing: list users page %d: %w", page, err)
		}

		for _, user := range result.Users {
			if err := s.IndexCandidate(ctx, user.ID); err != nil {
				log.Printf("indexing: failed to index user %s: %v", user.ID, err)
				continue
			}
		}

		if int32(len(result.Users)) < pageSize {
			break
		}
		page++
	}

	return nil
}

// preferEnglish picks English content from profile field texts, falling back to any available.
func preferEnglish(texts []domain.ProfileFieldText) string {
	var fallback string
	for _, t := range texts {
		if t.Lang == "en" {
			return t.Content
		}
		if fallback == "" {
			fallback = t.Content
		}
	}
	return fallback
}

// preferEnglishItemText picks English description from item texts, falling back to any available.
func preferEnglishItemText(texts []domain.ItemText) string {
	var fallback string
	for _, t := range texts {
		if t.Lang == "en" {
			return t.Description
		}
		if fallback == "" {
			fallback = t.Description
		}
	}
	return fallback
}

// calculateTotalExperienceMonths sums duration in months across all experience items.
func calculateTotalExperienceMonths(items []domain.ExperienceItem) int {
	var totalMonths int
	now := time.Now()

	for _, item := range items {
		start := indexParseExperienceDate(item.StartDate)
		if start.IsZero() {
			continue
		}
		end := now
		if item.EndDate != "" && !indexIsPresent(item.EndDate) {
			if parsed := indexParseExperienceDate(item.EndDate); !parsed.IsZero() {
				end = parsed
			}
		}
		months := int(end.Sub(start).Hours() / (24 * 30.44))
		if months > 0 {
			totalMonths += months
		}
	}

	return totalMonths
}

// indexParseExperienceDate parses date strings in various common formats.
func indexParseExperienceDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}

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

	if y, err := strconv.Atoi(s); err == nil && y > 1970 && y < 2100 {
		return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
	}

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

// indexIsPresent checks if the date string represents "present" / "current".
func indexIsPresent(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	return lower == "present" || lower == "current" || lower == "now" ||
		lower == "настоящее время" || lower == "hozirgi" || lower == "hozir"
}
