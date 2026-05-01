package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type VacancyAIRepository struct {
	pool *pgxpool.Pool
}

func NewVacancyAIRepository(pool *pgxpool.Pool) *VacancyAIRepository {
	return &VacancyAIRepository{pool: pool}
}

func (r *VacancyAIRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Vacancy, error) {
	row := r.pool.QueryRow(ctx, "SELECT id, hr_id, company_data, format, salary_currency, status FROM vacancies WHERE id = $1", id)
	var v domain.Vacancy
	var cdBytes []byte
	err := row.Scan(&v.ID, &v.HRID, &cdBytes, &v.Format, &v.SalaryCurrency, &v.Status)
	if err != nil {
		return nil, err
	}
	if len(cdBytes) > 0 {
		json.Unmarshal(cdBytes, &v.CompanyData)
	}
	return &v, nil
}

// GetAllReferencesForAI reads all taxonomy data from the DB to be injected into the Claude Prompt Cache.
func (r *VacancyAIRepository) GetAllReferencesForAI(ctx context.Context) (string, error) {
	var out strings.Builder

	out.WriteString("Main Categories:\n")
	rows, err := r.pool.Query(ctx, "SELECT id, name FROM main_category")
	if err == nil {
		for rows.Next() {
			var id uuid.UUID
			var name string
			if err := rows.Scan(&id, &name); err == nil {
				out.WriteString(fmt.Sprintf("- ID: %s | Name: %s\n", id, name))
			}
		}
		rows.Close()
	}

	out.WriteString("\nSub Categories:\n")
	rows, err = r.pool.Query(ctx, "SELECT id, name FROM sub_category")
	if err == nil {
		for rows.Next() {
			var id uuid.UUID
			var name string
			if err := rows.Scan(&id, &name); err == nil {
				out.WriteString(fmt.Sprintf("- ID: %s | Name: %s\n", id, name))
			}
		}
		rows.Close()
	}

	out.WriteString("\nTechnologies (Bazadagi mavjud texnologiyalar):\n")
	rows, err = r.pool.Query(ctx, "SELECT id, name FROM technologies")
	if err == nil {
		for rows.Next() {
			var id uuid.UUID
			var name string
			if err := rows.Scan(&id, &name); err == nil {
				out.WriteString(fmt.Sprintf("- ID: %s | Name: %s | TYPE: Technology\n", id, name))
			}
		}
		rows.Close()
	}

	out.WriteString("\nSkills (Bazadagi mavjud ko'nikmalar):\n")
	rows, err = r.pool.Query(ctx, "SELECT id, name FROM skills")
	if err == nil {
		forbiddenKeywords := []string{
			"swift", "kotlin", "java", "flutter", "dart", "figma", "photoshop", "illustrator", "adobe",
			"go", "golang", "python", "javascript", "react", "vue", "angular", "docker", "kubernetes", "git",
			"postgresql", "mysql", "mongodb", "redis", "firebase", "rest", "api", "ci/cd", "ci", "cd",
			"deployment", "publishing", "store", "ios", "android", "aws", "azure", "gcp", "cloud", "linux", "nginx",
		}

		for rows.Next() {
			var id uuid.UUID
			var name string
			if err := rows.Scan(&id, &name); err == nil {
				lowerName := strings.ToLower(name)
				isForbidden := false
				for _, kw := range forbiddenKeywords {
					if strings.Contains(lowerName, kw) {
						isForbidden = true
						break
					}
				}
				if isForbidden {
					continue
				}
				out.WriteString(fmt.Sprintf("- ID: %s | Name: %s | TYPE: Skill\n", id, name))
			}
		}
		rows.Close()
	}
	return out.String(), nil
}

// SaveParsedVacancy handles the complex transaction of inserting newly discovered technologies/skills
// and creating the vacancy record with all junction tables correctly mapped.
func (r *VacancyAIRepository) SaveParsedVacancy(ctx context.Context, v *domain.Vacancy, parsed domain.AIParsedVacancy) error {
	// 1. Prepare Taxonomy Guard Keywords
	techKeywords := []string{
		// IT & Development
		"swift", "kotlin", "java", "flutter", "dart", "go", "golang", "python", "javascript", "react", "vue", "angular", "docker", "kubernetes", "git",
		"postgresql", "mysql", "mongodb", "redis", "firebase", "rest", "api", "ci/cd", "ci", "cd", "aws", "azure", "gcp", "cloud", "linux", "nginx",
		// Design & UI/UX
		"figma", "photoshop", "illustrator", "adobe", "sketch", "zeplin", "invision", "canva", "indesign", "xd", "coreldraw",
		// Video & Audio Editing
		"premiere", "after effects", "davinci", "final cut", "audition", "obs", "vray", "blender", "3ds max", "maya", "cinema 4d",
		// Hardware & Equipment
		"camera", "lens", "microphone", "drone", "lighting", "rig", "stabilizer",
		// General Platforms
		"ios", "android", "deployment", "publishing", "store",
	}
	isTech := func(name string) bool {
		lower := strings.ToLower(name)
		for _, kw := range techKeywords {
			if strings.Contains(lower, kw) {
				return true
			}
		}
		return false
	}

	// 2. Re-route misclassified new skills to technologies
	finalNewSkills := []string{}
	for _, s := range parsed.NewSkills {
		if isTech(s) {
			parsed.NewTechnologies = append(parsed.NewTechnologies, s)
		} else {
			finalNewSkills = append(finalNewSkills, s)
		}
	}
	parsed.NewSkills = finalNewSkills

	// 3. Re-route misclassified matched skills to technologies (Check names by ID)
	finalMatchedSkillIDs := []string{}
	for _, idStr := range parsed.MatchedSkillIDs {
		var name string
		err := r.pool.QueryRow(ctx, "SELECT name FROM skills WHERE id = $1", idStr).Scan(&name)
		if err == nil && isTech(name) {
			// It's actually a tech! Let's see if we can find it in technologies table or add as new
			var techID uuid.UUID
			err = r.pool.QueryRow(ctx, "SELECT id FROM technologies WHERE LOWER(name) = LOWER($1)", name).Scan(&techID)
			if err == nil {
				parsed.MatchedTechIDs = append(parsed.MatchedTechIDs, techID.String())
			} else {
				parsed.NewTechnologies = append(parsed.NewTechnologies, name)
			}
		} else {
			finalMatchedSkillIDs = append(finalMatchedSkillIDs, idStr)
		}
	}
	parsed.MatchedSkillIDs = finalMatchedSkillIDs

	// --- END OF TAXONOMY GUARD ---

	// 4. Start Transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Process Main Category and Sub Category Dynamic Creation
	var mainCatID uuid.UUID
	var subCatID uuid.UUID
	
	if parsed.MatchedMainCatID != "" {
		if id, err := uuid.Parse(parsed.MatchedMainCatID); err == nil {
			mainCatID = id
		}
	}
	if mainCatID == uuid.Nil && parsed.NewMainCategory != "" {
		mainCatID = uuid.New()
		_, err := tx.Exec(ctx, `
			INSERT INTO main_category (id, name, created_at, updated_at) 
			VALUES ($1, $2, NOW(), NOW())
		`, mainCatID, parsed.NewMainCategory)
		if err != nil {
			return fmt.Errorf("insert new main category %s: %w", parsed.NewMainCategory, err)
		}
	}

	if parsed.MatchedSubCatID != "" {
		if id, err := uuid.Parse(parsed.MatchedSubCatID); err == nil {
			subCatID = id
		}
	}
	if subCatID == uuid.Nil && parsed.NewSubCategory != "" {
		subCatID = uuid.New()
		if mainCatID != uuid.Nil {
			_, err := tx.Exec(ctx, `
				INSERT INTO sub_category (id, name, main_category_id, created_at, updated_at) 
				VALUES ($1, $2, $3, NOW(), NOW())
			`, subCatID, parsed.NewSubCategory, mainCatID)
			if err != nil {
				return fmt.Errorf("insert new sub category %s: %w", parsed.NewSubCategory, err)
			}
		}
	}

	// 2. Insert new technologies and collect all IDs
	allTechIDs := make([]uuid.UUID, 0, len(parsed.MatchedTechIDs)+len(parsed.NewTechnologies))
	for _, oldTech := range parsed.MatchedTechIDs {
		if id, err := uuid.Parse(oldTech); err == nil {
			allTechIDs = append(allTechIDs, id)
		}
	}
	for _, newTechName := range parsed.NewTechnologies {
		var finalID uuid.UUID
		if subCatID != uuid.Nil {
			err = tx.QueryRow(ctx, `
				INSERT INTO technologies (id, name, sub_category_ids, created_at, updated_at) 
				VALUES ($1, $2, ARRAY[$3]::uuid[], NOW(), NOW())
				ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
				RETURNING id
			`, uuid.New(), newTechName, subCatID).Scan(&finalID)
		} else {
			err = tx.QueryRow(ctx, `
				INSERT INTO technologies (id, name, sub_category_ids, created_at, updated_at) 
				VALUES ($1, $2, ARRAY[]::uuid[], NOW(), NOW())
				ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
				RETURNING id
			`, uuid.New(), newTechName).Scan(&finalID)
		}
		if err != nil {
			return fmt.Errorf("upsert tech %s: %w", newTechName, err)
		}
		allTechIDs = append(allTechIDs, finalID)
	}

	// 3. Insert new skills and collect all IDs
	allSkillIDs := make([]uuid.UUID, 0, len(parsed.MatchedSkillIDs)+len(parsed.NewSkills))
	for _, oldSkill := range parsed.MatchedSkillIDs {
		if id, err := uuid.Parse(oldSkill); err == nil {
			allSkillIDs = append(allSkillIDs, id)
		}
	}
	for _, newSkillName := range parsed.NewSkills {
		var finalID uuid.UUID
		if subCatID != uuid.Nil {
			err = tx.QueryRow(ctx, `
				INSERT INTO skills (id, name, sub_category_ids, created_at, updated_at) 
				VALUES ($1, $2, ARRAY[$3]::uuid[], NOW(), NOW())
				ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
				RETURNING id
			`, uuid.New(), newSkillName, subCatID).Scan(&finalID)
		} else {
			err = tx.QueryRow(ctx, `
				INSERT INTO skills (id, name, sub_category_ids, created_at, updated_at) 
				VALUES ($1, $2, ARRAY[]::uuid[], NOW(), NOW())
				ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
				RETURNING id
			`, uuid.New(), newSkillName).Scan(&finalID)
		}
		if err != nil {
			return fmt.Errorf("upsert skill %s: %w", newSkillName, err)
		}
		allSkillIDs = append(allSkillIDs, finalID)
	}

	// 4. Save Vacancy Record (Update main categories)
	cdBytes, _ := json.Marshal(v.CompanyData)
	var mID, sID *uuid.UUID
	if mainCatID != uuid.Nil {
		mID = &mainCatID
	}
	if subCatID != uuid.Nil {
		sID = &subCatID
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO vacancies (id, hr_id, company_data, format, salary_currency, status, main_category_id, sub_category_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET 
			main_category_id = EXCLUDED.main_category_id,
			sub_category_id = EXCLUDED.sub_category_id,
			format = EXCLUDED.format,
			salary_currency = EXCLUDED.salary_currency
	`, v.ID, v.HRID, cdBytes, v.Format, v.SalaryCurrency, v.Status, mID, sID)
	if err != nil {
		return fmt.Errorf("upsert vacancy: %w", err)
	}

	// 5. Connect Technologies (Junction Table)
	for _, techID := range allTechIDs {
		var exists bool
		tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM technologies WHERE id=$1)", techID).Scan(&exists)
		if !exists {
			continue
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO vacancy_technologies (vacancy_id, technology_id, created_at, updated_at) 
			VALUES ($1, $2, NOW(), NOW()) ON CONFLICT DO NOTHING
		`, v.ID, techID)
	}

	// 6. Connect Skills (Junction Table)
	for _, skillID := range allSkillIDs {
		var exists bool
		tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM skills WHERE id=$1)", skillID).Scan(&exists)
		if !exists {
			continue
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO vacancy_skills (vacancy_id, skill_id) 
			VALUES ($1, $2) ON CONFLICT DO NOTHING
		`, v.ID, skillID)
	}

	return tx.Commit(ctx)
}
