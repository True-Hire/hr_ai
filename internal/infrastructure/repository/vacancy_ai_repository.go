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
	row := r.pool.QueryRow(ctx, "SELECT id, hr_id, company_data, format, salary_currency, status, main_category_id, sub_category_id FROM vacancies WHERE id = $1", id)
	var v domain.Vacancy
	var cdBytes []byte
	var mID, sID *uuid.UUID
	err := row.Scan(&v.ID, &v.HRID, &cdBytes, &v.Format, &v.SalaryCurrency, &v.Status, &mID, &sID)
	if err != nil {
		return nil, err
	}
	if len(cdBytes) > 0 {
		json.Unmarshal(cdBytes, &v.CompanyData)
	}
	if mID != nil {
		v.MainCategoryID = *mID
	}
	if sID != nil {
		v.SubCategoryID = *sID
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
			// IT & Development
			"swift", "kotlin", "java", "flutter", "dart", "go", "golang", "python", "javascript", "typescript", "react", "vue", "angular", "node.js", "nodejs", "express", "django", "flask", "laravel", "php", "ruby", "rails", "c#", ".net", "dotnet", "c++", "rust", "elixir", "scala", "clojure", "objective-c", "unity", "unreal",
			// Web & Frontend
			"html", "css", "scss", "sass", "less", "tailwind", "bootstrap", "material ui", "nextjs", "nuxtjs", "webpack", "vite", "babel", "jquery", "wordpress", "elementor",
			// Databases & Storage
			"postgresql", "mysql", "mongodb", "redis", "firebase", "sqlite", "mariadb", "cassandra", "elasticsearch", "oracle", "sql server", "dynamodb", "supabase", "pocketbase",
			// Infrastructure & DevOps
			"docker", "kubernetes", "git", "github", "gitlab", "bitbucket", "aws", "azure", "gcp", "cloud", "linux", "nginx", "apache", "terraform", "ansible", "jenkins", "ci/cd", "ci", "cd", "ubuntu", "centos", "debian", "grafana", "prometheus",
			// APIs & Protocols
			"rest", "api", "graphql", "grpc", "soap", "websocket", "http", "https", "tcp/ip", "udp",
			// Design & UI/UX & Media
			"figma", "photoshop", "illustrator", "adobe", "sketch", "zeplin", "invision", "canva", "indesign", "xd", "coreldraw", "affinity", "procreate", "framer", "principle", "blender", "maya", "3ds max", "cinema 4d", "vray", "lumion", "premiere", "after effects", "davinci", "final cut", "audition", "obs", "handbrake", "vlc", "ableton", "logic pro", "fl studio",
			// Hardware & General Equipment
			"camera", "lens", "microphone", "drone", "lighting", "rig", "stabilizer", "router", "switch", "server", "workstation", "monitor", "laptop", "printer", "scanner", "pos",
			// Construction & Craftsmanship
			"perforator", "drill", "shurupovert", "bolgarka", "welder", "welding", "laser level", "leveler", "saw", "hammer", "screwdriver", "pliers", "wrench", "measuring tape", "mixer", "generator", "compressor", "jackhammer", "tile cutter", "kafel", "gipsokarton", "malyarka",
			// Medical & Healthcare
			"stethoscope", "ultrasound", "uzi", "mri", "mrt", "x-ray", "rentgen", "ecg", "eeg", "scalpel", "forceps", "thermometer", "tonometer", "ventilator", "syringe", "microscope", "dental drill", "implants",
			// Platforms & OS & E-sports
			"ios", "android", "windows", "macos", "unix", "ubuntu", "redhat", "store", "app store", "google play", "steam", "playstation", "xbox", "nintendo", "esports", "gaming pc",
			// Office & CRM & Business & Law
			"excel", "word", "powerpoint", "outlook", "google workspace", "sheets", "docs", "bitrix24", "amocrm", "salesforce", "jira", "trello", "asana", "slack", "zoom", "teams", "confluence", "notion", "clickup", "monday.com", "electronic signature", "eds", "e-imzo", "crm", "erp",
			// General Technical Tools
			"bot", "plugin", "extension", "addon",
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
		"swift", "kotlin", "java", "flutter", "dart", "go", "golang", "python", "javascript", "typescript", "react", "vue", "angular", "node.js", "nodejs", "express", "django", "flask", "laravel", "php", "ruby", "rails", "c#", ".net", "dotnet", "c++", "rust", "elixir", "scala", "clojure", "objective-c", "unity", "unreal",
		// Web & Frontend
		"html", "css", "scss", "sass", "less", "tailwind", "bootstrap", "material ui", "nextjs", "nuxtjs", "webpack", "vite", "babel", "jquery", "wordpress", "elementor",
		// Databases & Storage
		"postgresql", "mysql", "mongodb", "redis", "firebase", "sqlite", "mariadb", "cassandra", "elasticsearch", "oracle", "sql server", "dynamodb", "supabase", "pocketbase",
		// Infrastructure & DevOps
		"docker", "kubernetes", "git", "github", "gitlab", "bitbucket", "aws", "azure", "gcp", "cloud", "linux", "nginx", "apache", "terraform", "ansible", "jenkins", "ci/cd", "ci", "cd", "ubuntu", "centos", "debian", "grafana", "prometheus",
		// APIs & Protocols
		"rest", "api", "graphql", "grpc", "soap", "websocket", "http", "https", "tcp/ip", "udp",
		// Design & UI/UX & Media
		"figma", "photoshop", "illustrator", "adobe", "sketch", "zeplin", "invision", "canva", "indesign", "xd", "coreldraw", "affinity", "procreate", "framer", "principle", "blender", "maya", "3ds max", "cinema 4d", "vray", "lumion", "premiere", "after effects", "davinci", "final cut", "audition", "obs", "handbrake", "vlc", "ableton", "logic pro", "fl studio",
		// Hardware & General Equipment
		"camera", "lens", "microphone", "drone", "lighting", "rig", "stabilizer", "router", "switch", "server", "workstation", "monitor", "laptop", "printer", "scanner", "pos",
		// Construction & Craftsmanship (Ustachilik)
		"perforator", "drill", "shurupovert", "bolgarka", "welder", "welding", "laser level", "leveler", "saw", "hammer", "screwdriver", "pliers", "wrench", "measuring tape", "mixer", "generator", "compressor", "jackhammer", "tile cutter", "kafel", "gipsokarton", "malyarka",
		// Medical & Healthcare (Tibbiyot)
		"stethoscope", "ultrasound", "uzi", "mri", "mrt", "x-ray", "rentgen", "ecg", "eeg", "scalpel", "forceps", "thermometer", "tonometer", "ventilator", "syringe", "microscope", "dental drill", "implants",
		// Platforms & OS & E-sports
		"ios", "android", "windows", "macos", "unix", "ubuntu", "redhat", "store", "app store", "google play", "steam", "playstation", "xbox", "nintendo", "esports", "gaming pc",
		// Office & CRM & Business & Law
		"excel", "word", "powerpoint", "outlook", "google workspace", "sheets", "docs", "bitrix24", "amocrm", "salesforce", "jira", "trello", "asana", "slack", "zoom", "teams", "confluence", "notion", "clickup", "monday.com", "electronic signature", "eds", "e-imzo", "crm", "erp",
		// General Technical Tools
		"bot", "plugin", "extension", "addon",
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
			// Update existing tech to link with current sub-category
			if subCatID != uuid.Nil {
				_, _ = tx.Exec(ctx, `
					UPDATE technologies 
					SET sub_category_ids = CASE 
						WHEN $1 = ANY(sub_category_ids) THEN sub_category_ids 
						ELSE array_append(COALESCE(sub_category_ids, ARRAY[]::uuid[]), $1) 
					END,
					updated_at = NOW()
					WHERE id = $2
				`, subCatID, id)
			}
		}
	}
	for _, newTechName := range parsed.NewTechnologies {
		var finalID uuid.UUID
		if subCatID != uuid.Nil {
			err = tx.QueryRow(ctx, `
				INSERT INTO technologies (id, name, sub_category_ids, created_at, updated_at) 
				VALUES ($1, $2, ARRAY[$3]::uuid[], NOW(), NOW())
				ON CONFLICT (name) 
				DO UPDATE SET 
					sub_category_ids = CASE 
						WHEN $3 = ANY(technologies.sub_category_ids) THEN technologies.sub_category_ids 
						ELSE array_append(COALESCE(technologies.sub_category_ids, ARRAY[]::uuid[]), $3) 
					END,
					updated_at = NOW()
				RETURNING id
			`, uuid.New(), newTechName, subCatID).Scan(&finalID)
		} else {
			err = tx.QueryRow(ctx, `
				INSERT INTO technologies (id, name, sub_category_ids, created_at, updated_at) 
				VALUES ($1, $2, ARRAY[]::uuid[], NOW(), NOW())
				ON CONFLICT (name) 
				DO UPDATE SET 
					sub_category_ids = COALESCE(technologies.sub_category_ids, ARRAY[]::uuid[]),
					updated_at = NOW()
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
			// Update existing skill to link with current sub-category
			if subCatID != uuid.Nil {
				_, _ = tx.Exec(ctx, `
					UPDATE skills 
					SET sub_category_ids = CASE 
						WHEN $1 = ANY(sub_category_ids) THEN sub_category_ids 
						ELSE array_append(COALESCE(sub_category_ids, ARRAY[]::uuid[]), $1) 
					END,
					updated_at = NOW()
					WHERE id = $2
				`, subCatID, id)
			}
		}
	}
	for _, newSkillName := range parsed.NewSkills {
		var finalID uuid.UUID
		if subCatID != uuid.Nil {
			err = tx.QueryRow(ctx, `
				INSERT INTO skills (id, name, sub_category_ids, created_at, updated_at) 
				VALUES ($1, $2, ARRAY[$3]::uuid[], NOW(), NOW())
				ON CONFLICT (name) 
				DO UPDATE SET 
					sub_category_ids = CASE 
						WHEN $3 = ANY(skills.sub_category_ids) THEN skills.sub_category_ids 
						ELSE array_append(COALESCE(skills.sub_category_ids, ARRAY[]::uuid[]), $3) 
					END,
					updated_at = NOW()
				RETURNING id
			`, uuid.New(), newSkillName, subCatID).Scan(&finalID)
		} else {
			err = tx.QueryRow(ctx, `
				INSERT INTO skills (id, name, sub_category_ids, created_at, updated_at) 
				VALUES ($1, $2, ARRAY[]::uuid[], NOW(), NOW())
				ON CONFLICT (name) 
				DO UPDATE SET 
					sub_category_ids = COALESCE(skills.sub_category_ids, ARRAY[]::uuid[]),
					updated_at = NOW()
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

func (r *VacancyAIRepository) SaveParsedProfile(ctx context.Context, userID uuid.UUID, parsed *domain.AIParsedVacancy) error {
	// 1. Taxonomy Guard (Server-side)
	techKeywords := []string{
		"swift", "kotlin", "java", "flutter", "dart", "go", "golang", "python", "javascript", "typescript", "react", "vue", "angular", "node.js", "nodejs", "express", "django", "flask", "laravel", "php", "ruby", "rails", "c#", ".net", "dotnet", "c++", "rust", "elixir", "scala", "clojure", "objective-c", "unity", "unreal",
		"html", "css", "scss", "sass", "less", "tailwind", "bootstrap", "material ui", "nextjs", "nuxtjs", "webpack", "vite", "babel", "jquery", "wordpress", "elementor",
		"postgresql", "mysql", "mongodb", "redis", "firebase", "sqlite", "mariadb", "cassandra", "elasticsearch", "oracle", "sql server", "dynamodb", "supabase", "pocketbase",
		"docker", "kubernetes", "git", "github", "gitlab", "bitbucket", "aws", "azure", "gcp", "cloud", "linux", "nginx", "apache", "terraform", "ansible", "jenkins", "ci/cd", "ci", "cd", "ubuntu", "centos", "debian", "grafana", "prometheus",
		"rest", "api", "graphql", "grpc", "soap", "websocket", "http", "https", "tcp/ip", "udp",
		"figma", "photoshop", "illustrator", "adobe", "sketch", "zeplin", "invision", "canva", "indesign", "xd", "coreldraw", "affinity", "procreate", "framer", "principle", "blender", "maya", "3ds max", "cinema 4d", "vray", "lumion", "premiere", "after effects", "davinci", "final cut", "audition", "obs", "handbrake", "vlc", "ableton", "logic pro", "fl studio",
		"camera", "lens", "microphone", "drone", "lighting", "rig", "stabilizer", "router", "switch", "server", "workstation", "monitor", "laptop", "printer", "scanner", "pos",
		"perforator", "drill", "shurupovert", "bolgarka", "welder", "welding", "laser level", "leveler", "saw", "hammer", "screwdriver", "pliers", "wrench", "measuring tape", "mixer", "generator", "compressor", "jackhammer", "tile cutter", "kafel", "gipsokarton", "malyarka",
		"stethoscope", "ultrasound", "uzi", "mri", "mrt", "x-ray", "rentgen", "ecg", "eeg", "scalpel", "forceps", "thermometer", "tonometer", "ventilator", "syringe", "microscope", "dental drill", "implants",
		"ios", "android", "windows", "macos", "unix", "ubuntu", "redhat", "store", "app store", "google play", "steam", "playstation", "xbox", "nintendo", "esports", "gaming pc",
		"excel", "word", "powerpoint", "outlook", "google workspace", "sheets", "docs", "bitrix24", "amocrm", "salesforce", "jira", "trello", "asana", "slack", "zoom", "teams", "confluence", "notion", "clickup", "monday.com", "electronic signature", "eds", "e-imzo", "crm", "erp",
		"bot", "plugin", "extension", "addon",
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

	finalMatchedSkillIDs := make([]string, 0)
	for _, idStr := range parsed.MatchedSkillIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		var name string
		err = r.pool.QueryRow(ctx, "SELECT name FROM skills WHERE id = $1", id).Scan(&name)
		if err == nil && isTech(name) {
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

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var mainCatID uuid.UUID
	var subCatID uuid.UUID

	if parsed.MatchedMainCatID != "" {
		mainCatID, _ = uuid.Parse(parsed.MatchedMainCatID)
	}
	if mainCatID == uuid.Nil && parsed.NewMainCategory != "" {
		mainCatID = uuid.New()
		_, _ = tx.Exec(ctx, "INSERT INTO main_category (id, name, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) ON CONFLICT (name) DO UPDATE SET updated_at = NOW() RETURNING id", mainCatID, parsed.NewMainCategory)
		_ = tx.QueryRow(ctx, "SELECT id FROM main_category WHERE LOWER(name) = LOWER($1)", parsed.NewMainCategory).Scan(&mainCatID)
	}

	if parsed.MatchedSubCatID != "" {
		subCatID, _ = uuid.Parse(parsed.MatchedSubCatID)
	}
	if subCatID == uuid.Nil && parsed.NewSubCategory != "" {
		subCatID = uuid.New()
		_, _ = tx.Exec(ctx, "INSERT INTO sub_category (id, name, main_category_id, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) ON CONFLICT (name) DO UPDATE SET updated_at = NOW() RETURNING id", subCatID, parsed.NewSubCategory, mainCatID)
		_ = tx.QueryRow(ctx, "SELECT id FROM sub_category WHERE LOWER(name) = LOWER($1)", parsed.NewSubCategory).Scan(&subCatID)
	}

	// Update user categories
	_, err = tx.Exec(ctx, "UPDATE users SET main_category_id = $1, sub_category_id = $2 WHERE id = $3", mainCatID, subCatID, userID)
	if err != nil {
		return fmt.Errorf("update user categories: %w", err)
	}

	allTechIDs := make([]uuid.UUID, 0)
	for _, oldTech := range parsed.MatchedTechIDs {
		if id, err := uuid.Parse(oldTech); err == nil {
			allTechIDs = append(allTechIDs, id)
			if subCatID != uuid.Nil {
				_, _ = tx.Exec(ctx, `UPDATE technologies SET sub_category_ids = CASE WHEN $1 = ANY(sub_category_ids) THEN sub_category_ids ELSE array_append(COALESCE(sub_category_ids, ARRAY[]::uuid[]), $1) END, updated_at = NOW() WHERE id = $2`, subCatID, id)
			}
		}
	}
	for _, newTechName := range parsed.NewTechnologies {
		var finalID uuid.UUID
		if subCatID != uuid.Nil {
			err = tx.QueryRow(ctx, `INSERT INTO technologies (id, name, sub_category_ids, created_at, updated_at) VALUES ($1, $2, ARRAY[$3]::uuid[], NOW(), NOW()) ON CONFLICT (name) DO UPDATE SET sub_category_ids = CASE WHEN $3 = ANY(technologies.sub_category_ids) THEN technologies.sub_category_ids ELSE array_append(COALESCE(technologies.sub_category_ids, ARRAY[]::uuid[]), $3) END, updated_at = NOW() RETURNING id`, uuid.New(), newTechName, subCatID).Scan(&finalID)
		} else {
			err = tx.QueryRow(ctx, `INSERT INTO technologies (id, name, sub_category_ids, created_at, updated_at) VALUES ($1, $2, ARRAY[]::uuid[], NOW(), NOW()) ON CONFLICT (name) DO UPDATE SET sub_category_ids = COALESCE(technologies.sub_category_ids, ARRAY[]::uuid[]), updated_at = NOW() RETURNING id`, uuid.New(), newTechName).Scan(&finalID)
		}
		if err == nil {
			allTechIDs = append(allTechIDs, finalID)
		}
	}

	allSkillIDs := make([]uuid.UUID, 0)
	for _, oldSkill := range parsed.MatchedSkillIDs {
		if id, err := uuid.Parse(oldSkill); err == nil {
			allSkillIDs = append(allSkillIDs, id)
			if subCatID != uuid.Nil {
				_, _ = tx.Exec(ctx, `UPDATE skills SET sub_category_ids = CASE WHEN $1 = ANY(sub_category_ids) THEN sub_category_ids ELSE array_append(COALESCE(sub_category_ids, ARRAY[]::uuid[]), $1) END, updated_at = NOW() WHERE id = $2`, subCatID, id)
			}
		}
	}
	for _, newSkillName := range parsed.NewSkills {
		var finalID uuid.UUID
		if subCatID != uuid.Nil {
			err = tx.QueryRow(ctx, `INSERT INTO skills (id, name, sub_category_ids, created_at, updated_at) VALUES ($1, $2, ARRAY[$3]::uuid[], NOW(), NOW()) ON CONFLICT (name) DO UPDATE SET sub_category_ids = CASE WHEN $3 = ANY(skills.sub_category_ids) THEN skills.sub_category_ids ELSE array_append(COALESCE(skills.sub_category_ids, ARRAY[]::uuid[]), $3) END, updated_at = NOW() RETURNING id`, uuid.New(), newSkillName, subCatID).Scan(&finalID)
		} else {
			err = tx.QueryRow(ctx, `INSERT INTO skills (id, name, sub_category_ids, created_at, updated_at) VALUES ($1, $2, ARRAY[]::uuid[], NOW(), NOW()) ON CONFLICT (name) DO UPDATE SET sub_category_ids = COALESCE(skills.sub_category_ids, ARRAY[]::uuid[]), updated_at = NOW() RETURNING id`, uuid.New(), newSkillName).Scan(&finalID)
		}
		if err == nil {
			allSkillIDs = append(allSkillIDs, finalID)
		}
	}

	// Connect Junctions
	for _, techID := range allTechIDs {
		_, _ = tx.Exec(ctx, `INSERT INTO user_technologies (user_id, technology_id, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) ON CONFLICT DO NOTHING`, userID, techID)
	}
	for _, skillID := range allSkillIDs {
		_, _ = tx.Exec(ctx, `INSERT INTO user_skills (user_id, skill_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, skillID)
	}

	return tx.Commit(ctx)
}
