package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CandidateSearchProfileRepository struct {
	pool *pgxpool.Pool
}

func NewCandidateSearchProfileRepository(pool *pgxpool.Pool) *CandidateSearchProfileRepository {
	return &CandidateSearchProfileRepository{pool: pool}
}

func (r *CandidateSearchProfileRepository) Upsert(ctx context.Context, profile *domain.CandidateSearchProfile) error {
	scoringJSON, err := json.Marshal(profile.ScoringFactors)
	if err != nil {
		return fmt.Errorf("marshal scoring_factors: %w", err)
	}
	parsedJSON, err := json.Marshal(profile.ParsedEntities)
	if err != nil {
		return fmt.Errorf("marshal parsed_entities: %w", err)
	}

	query := `
		INSERT INTO candidate_search_profiles (
			user_id, primary_role, role_family, seniority, total_experience_months,
			highest_education_level, skills, industries, project_domains, company_names,
			known_languages, education_fields, universities, location_city, location_country,
			willing_to_relocate,
			backend_score, frontend_score, mobile_score, data_score, qa_score, pm_score,
			devops_score, design_score,
			devops_support_score, client_communication_score, project_management_score,
			ownership_score, leadership_score, mentoring_score, startup_adaptability_score,
			company_prestige_score, engineering_environment_score, internship_quality_score,
			education_quality_score, competition_score, open_source_score,
			growth_trajectory_score, project_complexity_score,
			overall_strength_score, backend_strength_score, frontend_strength_score,
			data_strength_score,
			search_text, search_tsv,
			scoring_factors, parsed_entities,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16,
			$17, $18, $19, $20, $21, $22,
			$23, $24,
			$25, $26, $27,
			$28, $29, $30, $31,
			$32, $33, $34,
			$35, $36, $37,
			$38, $39,
			$40, $41, $42,
			$43,
			$44, to_tsvector('english', $44),
			$45, $46,
			now()
		)
		ON CONFLICT (user_id) DO UPDATE SET
			primary_role = EXCLUDED.primary_role,
			role_family = EXCLUDED.role_family,
			seniority = EXCLUDED.seniority,
			total_experience_months = EXCLUDED.total_experience_months,
			highest_education_level = EXCLUDED.highest_education_level,
			skills = EXCLUDED.skills,
			industries = EXCLUDED.industries,
			project_domains = EXCLUDED.project_domains,
			company_names = EXCLUDED.company_names,
			known_languages = EXCLUDED.known_languages,
			education_fields = EXCLUDED.education_fields,
			universities = EXCLUDED.universities,
			location_city = EXCLUDED.location_city,
			location_country = EXCLUDED.location_country,
			willing_to_relocate = EXCLUDED.willing_to_relocate,
			backend_score = EXCLUDED.backend_score,
			frontend_score = EXCLUDED.frontend_score,
			mobile_score = EXCLUDED.mobile_score,
			data_score = EXCLUDED.data_score,
			qa_score = EXCLUDED.qa_score,
			pm_score = EXCLUDED.pm_score,
			devops_score = EXCLUDED.devops_score,
			design_score = EXCLUDED.design_score,
			devops_support_score = EXCLUDED.devops_support_score,
			client_communication_score = EXCLUDED.client_communication_score,
			project_management_score = EXCLUDED.project_management_score,
			ownership_score = EXCLUDED.ownership_score,
			leadership_score = EXCLUDED.leadership_score,
			mentoring_score = EXCLUDED.mentoring_score,
			startup_adaptability_score = EXCLUDED.startup_adaptability_score,
			company_prestige_score = EXCLUDED.company_prestige_score,
			engineering_environment_score = EXCLUDED.engineering_environment_score,
			internship_quality_score = EXCLUDED.internship_quality_score,
			education_quality_score = EXCLUDED.education_quality_score,
			competition_score = EXCLUDED.competition_score,
			open_source_score = EXCLUDED.open_source_score,
			growth_trajectory_score = EXCLUDED.growth_trajectory_score,
			project_complexity_score = EXCLUDED.project_complexity_score,
			overall_strength_score = EXCLUDED.overall_strength_score,
			backend_strength_score = EXCLUDED.backend_strength_score,
			frontend_strength_score = EXCLUDED.frontend_strength_score,
			data_strength_score = EXCLUDED.data_strength_score,
			search_text = EXCLUDED.search_text,
			search_tsv = EXCLUDED.search_tsv,
			scoring_factors = EXCLUDED.scoring_factors,
			parsed_entities = EXCLUDED.parsed_entities,
			updated_at = now()
	`

	_, err = r.pool.Exec(ctx, query,
		profile.UserID,
		profile.PrimaryRole,
		profile.RoleFamily,
		profile.Seniority,
		profile.TotalExperienceMonths,
		profile.HighestEducationLevel,
		profile.Skills,
		profile.Industries,
		profile.ProjectDomains,
		profile.CompanyNames,
		profile.KnownLanguages,
		profile.EducationFields,
		profile.Universities,
		profile.LocationCity,
		profile.LocationCountry,
		profile.WillingToRelocate,
		profile.BackendScore,
		profile.FrontendScore,
		profile.MobileScore,
		profile.DataScore,
		profile.QAScore,
		profile.PMScore,
		profile.DevOpsScore,
		profile.DesignScore,
		profile.DevOpsSupportScore,
		profile.ClientCommunicationScore,
		profile.ProjectManagementScore,
		profile.OwnershipScore,
		profile.LeadershipScore,
		profile.MentoringScore,
		profile.StartupAdaptabilityScore,
		profile.CompanyPrestigeScore,
		profile.EngineeringEnvironmentScore,
		profile.InternshipQualityScore,
		profile.EducationQualityScore,
		profile.CompetitionScore,
		profile.OpenSourceScore,
		profile.GrowthTrajectoryScore,
		profile.ProjectComplexityScore,
		profile.OverallStrengthScore,
		profile.BackendStrengthScore,
		profile.FrontendStrengthScore,
		profile.DataStrengthScore,
		profile.SearchText,
		scoringJSON,
		parsedJSON,
	)
	if err != nil {
		return fmt.Errorf("upsert candidate search profile: %w", err)
	}
	return nil
}

func (r *CandidateSearchProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.CandidateSearchProfile, error) {
	query := `
		SELECT user_id, primary_role, role_family, seniority, total_experience_months,
			highest_education_level, skills, industries, project_domains, company_names,
			known_languages, education_fields, universities, location_city, location_country,
			willing_to_relocate,
			backend_score, frontend_score, mobile_score, data_score, qa_score, pm_score,
			devops_score, design_score,
			devops_support_score, client_communication_score, project_management_score,
			ownership_score, leadership_score, mentoring_score, startup_adaptability_score,
			company_prestige_score, engineering_environment_score, internship_quality_score,
			education_quality_score, competition_score, open_source_score,
			growth_trajectory_score, project_complexity_score,
			overall_strength_score, backend_strength_score, frontend_strength_score,
			data_strength_score,
			search_text, scoring_factors, parsed_entities, updated_at
		FROM candidate_search_profiles
		WHERE user_id = $1
	`

	row := r.pool.QueryRow(ctx, query, userID)
	return scanCandidateSearchProfile(row)
}

func (r *CandidateSearchProfileRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM candidate_search_profiles WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("delete candidate search profile: %w", err)
	}
	return nil
}

func (r *CandidateSearchProfileRepository) SearchPool(
	ctx context.Context,
	query string,
	roleFamily string,
	seniority string,
	locationCity string,
	locationCountry string,
	poolSize int,
) ([]domain.CandidateSearchProfile, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	// Use OR-based tsquery for broader matching (each word is optional)
	conditions = append(conditions, fmt.Sprintf("search_tsv @@ to_tsquery('english', $%d)", argIdx))
	orQuery := buildOrTsQuery(query)
	args = append(args, orQuery)
	argIdx++

	if roleFamily != "" {
		conditions = append(conditions, fmt.Sprintf("role_family = $%d", argIdx))
		args = append(args, roleFamily)
		argIdx++
	}
	if seniority != "" {
		conditions = append(conditions, fmt.Sprintf("seniority = $%d", argIdx))
		args = append(args, seniority)
		argIdx++
	}
	if locationCity != "" {
		conditions = append(conditions, fmt.Sprintf("(location_city = $%d OR willing_to_relocate = true)", argIdx))
		args = append(args, locationCity)
		argIdx++
	}
	if locationCountry != "" {
		conditions = append(conditions, fmt.Sprintf("(location_country = $%d OR willing_to_relocate = true)", argIdx))
		args = append(args, locationCountry)
		argIdx++
	}

	sql := fmt.Sprintf(`
		SELECT user_id, primary_role, role_family, seniority, total_experience_months,
			highest_education_level, skills, industries, project_domains, company_names,
			known_languages, education_fields, universities, location_city, location_country,
			willing_to_relocate,
			backend_score, frontend_score, mobile_score, data_score, qa_score, pm_score,
			devops_score, design_score,
			devops_support_score, client_communication_score, project_management_score,
			ownership_score, leadership_score, mentoring_score, startup_adaptability_score,
			company_prestige_score, engineering_environment_score, internship_quality_score,
			education_quality_score, competition_score, open_source_score,
			growth_trajectory_score, project_complexity_score,
			overall_strength_score, backend_strength_score, frontend_strength_score,
			data_strength_score,
			search_text, scoring_factors, parsed_entities, updated_at
		FROM candidate_search_profiles
		WHERE %s
		ORDER BY ts_rank(search_tsv, to_tsquery('english', $1)) DESC
		LIMIT $%d
	`, strings.Join(conditions, " AND "), argIdx)
	args = append(args, poolSize)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("search pool: %w", err)
	}
	defer rows.Close()

	return collectCandidateSearchProfiles(rows)
}

func (r *CandidateSearchProfileRepository) SearchPoolBySkills(
	ctx context.Context,
	skills []string,
	roleFamily string,
	locationCity string,
	poolSize int,
) ([]domain.CandidateSearchProfile, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("skills && $%d", argIdx))
	args = append(args, skills)
	argIdx++

	if roleFamily != "" {
		conditions = append(conditions, fmt.Sprintf("role_family = $%d", argIdx))
		args = append(args, roleFamily)
		argIdx++
	}
	if locationCity != "" {
		conditions = append(conditions, fmt.Sprintf("(location_city = $%d OR willing_to_relocate = true)", argIdx))
		args = append(args, locationCity)
		argIdx++
	}

	sql := fmt.Sprintf(`
		SELECT user_id, primary_role, role_family, seniority, total_experience_months,
			highest_education_level, skills, industries, project_domains, company_names,
			known_languages, education_fields, universities, location_city, location_country,
			willing_to_relocate,
			backend_score, frontend_score, mobile_score, data_score, qa_score, pm_score,
			devops_score, design_score,
			devops_support_score, client_communication_score, project_management_score,
			ownership_score, leadership_score, mentoring_score, startup_adaptability_score,
			company_prestige_score, engineering_environment_score, internship_quality_score,
			education_quality_score, competition_score, open_source_score,
			growth_trajectory_score, project_complexity_score,
			overall_strength_score, backend_strength_score, frontend_strength_score,
			data_strength_score,
			search_text, scoring_factors, parsed_entities, updated_at
		FROM candidate_search_profiles
		WHERE %s
		ORDER BY overall_strength_score DESC
		LIMIT $%d
	`, strings.Join(conditions, " AND "), argIdx)
	args = append(args, poolSize)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("search pool by skills: %w", err)
	}
	defer rows.Close()

	return collectCandidateSearchProfiles(rows)
}

func scanCandidateSearchProfile(row pgx.Row) (*domain.CandidateSearchProfile, error) {
	var p domain.CandidateSearchProfile
	var scoringJSON, parsedJSON []byte

	err := row.Scan(
		&p.UserID,
		&p.PrimaryRole,
		&p.RoleFamily,
		&p.Seniority,
		&p.TotalExperienceMonths,
		&p.HighestEducationLevel,
		&p.Skills,
		&p.Industries,
		&p.ProjectDomains,
		&p.CompanyNames,
		&p.KnownLanguages,
		&p.EducationFields,
		&p.Universities,
		&p.LocationCity,
		&p.LocationCountry,
		&p.WillingToRelocate,
		&p.BackendScore,
		&p.FrontendScore,
		&p.MobileScore,
		&p.DataScore,
		&p.QAScore,
		&p.PMScore,
		&p.DevOpsScore,
		&p.DesignScore,
		&p.DevOpsSupportScore,
		&p.ClientCommunicationScore,
		&p.ProjectManagementScore,
		&p.OwnershipScore,
		&p.LeadershipScore,
		&p.MentoringScore,
		&p.StartupAdaptabilityScore,
		&p.CompanyPrestigeScore,
		&p.EngineeringEnvironmentScore,
		&p.InternshipQualityScore,
		&p.EducationQualityScore,
		&p.CompetitionScore,
		&p.OpenSourceScore,
		&p.GrowthTrajectoryScore,
		&p.ProjectComplexityScore,
		&p.OverallStrengthScore,
		&p.BackendStrengthScore,
		&p.FrontendStrengthScore,
		&p.DataStrengthScore,
		&p.SearchText,
		&scoringJSON,
		&parsedJSON,
		&p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("candidate search profile not found")
		}
		return nil, fmt.Errorf("scan candidate search profile: %w", err)
	}

	if len(scoringJSON) > 0 {
		if err := json.Unmarshal(scoringJSON, &p.ScoringFactors); err != nil {
			return nil, fmt.Errorf("unmarshal scoring_factors: %w", err)
		}
	}
	if len(parsedJSON) > 0 {
		if err := json.Unmarshal(parsedJSON, &p.ParsedEntities); err != nil {
			return nil, fmt.Errorf("unmarshal parsed_entities: %w", err)
		}
	}

	return &p, nil
}

func collectCandidateSearchProfiles(rows pgx.Rows) ([]domain.CandidateSearchProfile, error) {
	var profiles []domain.CandidateSearchProfile
	for rows.Next() {
		var p domain.CandidateSearchProfile
		var scoringJSON, parsedJSON []byte

		err := rows.Scan(
			&p.UserID,
			&p.PrimaryRole,
			&p.RoleFamily,
			&p.Seniority,
			&p.TotalExperienceMonths,
			&p.HighestEducationLevel,
			&p.Skills,
			&p.Industries,
			&p.ProjectDomains,
			&p.CompanyNames,
			&p.KnownLanguages,
			&p.EducationFields,
			&p.Universities,
			&p.LocationCity,
			&p.LocationCountry,
			&p.WillingToRelocate,
			&p.BackendScore,
			&p.FrontendScore,
			&p.MobileScore,
			&p.DataScore,
			&p.QAScore,
			&p.PMScore,
			&p.DevOpsScore,
			&p.DesignScore,
			&p.DevOpsSupportScore,
			&p.ClientCommunicationScore,
			&p.ProjectManagementScore,
			&p.OwnershipScore,
			&p.LeadershipScore,
			&p.MentoringScore,
			&p.StartupAdaptabilityScore,
			&p.CompanyPrestigeScore,
			&p.EngineeringEnvironmentScore,
			&p.InternshipQualityScore,
			&p.EducationQualityScore,
			&p.CompetitionScore,
			&p.OpenSourceScore,
			&p.GrowthTrajectoryScore,
			&p.ProjectComplexityScore,
			&p.OverallStrengthScore,
			&p.BackendStrengthScore,
			&p.FrontendStrengthScore,
			&p.DataStrengthScore,
			&p.SearchText,
			&scoringJSON,
			&parsedJSON,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan candidate search profile row: %w", err)
		}

		if len(scoringJSON) > 0 {
			if err := json.Unmarshal(scoringJSON, &p.ScoringFactors); err != nil {
				return nil, fmt.Errorf("unmarshal scoring_factors: %w", err)
			}
		}
		if len(parsedJSON) > 0 {
			if err := json.Unmarshal(parsedJSON, &p.ParsedEntities); err != nil {
				return nil, fmt.Errorf("unmarshal parsed_entities: %w", err)
			}
		}

		profiles = append(profiles, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate candidate search profiles: %w", err)
	}
	return profiles, nil
}

// buildOrTsQuery converts a space-separated query string into OR-based tsquery.
// "backend developer go postgresql" → "backend | developer | go | postgresql"
func buildOrTsQuery(query string) string {
	words := strings.Fields(strings.ToLower(query))
	var clean []string
	for _, w := range words {
		// Remove non-alphanumeric chars that break tsquery syntax
		w = strings.TrimFunc(w, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_')
		})
		if w != "" {
			clean = append(clean, w)
		}
	}
	if len(clean) == 0 {
		return ""
	}
	return strings.Join(clean, " | ")
}
