package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/application/scoring"
	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/repository"
)

type SampleUser struct {
	ID                      string   `json:"id"`
	FirstName               string   `json:"first_name"`
	LastName                string   `json:"last_name"`
	Patronymic              string   `json:"patronymic"`
	Phone                   string   `json:"phone"`
	Telegram                string   `json:"telegram"`
	TelegramID              string   `json:"telegram_id"`
	Email                   string   `json:"email"`
	Gender                  string   `json:"gender"`
	Country                 string   `json:"country"`
	Region                  string   `json:"region"`
	Nationality             string   `json:"nationality"`
	Status                  string   `json:"status"`
	TariffType              string   `json:"tariff_type"`
	JobStatus               string   `json:"job_status"`
	ActivityType            string   `json:"activity_type"`
	Specializations         []string `json:"specializations"`
	ProfileScore            int32    `json:"profile_score"`
	EstimatedSalaryMin      int32    `json:"estimated_salary_min"`
	EstimatedSalaryMax      int32    `json:"estimated_salary_max"`
	EstimatedSalaryCurrency string   `json:"estimated_salary_currency"`
	Profile                 SampleProfile `json:"profile"`
}

type SampleProfile struct {
	Title          string                `json:"title"`
	About          string                `json:"about"`
	Skills         []string              `json:"skills"`
	Languages      []SampleLanguage      `json:"languages"`
	Certifications []string              `json:"certifications"`
	Achievements   string                `json:"achievements"`
	Experience     []SampleExperience    `json:"experience"`
	Education      []SampleEducation     `json:"education"`
}

type SampleLanguage struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

type SampleExperience struct {
	ID          string          `json:"id"`
	Company     string          `json:"company"`
	Position    string          `json:"position"`
	StartDate   string          `json:"start_date"`
	EndDate     string          `json:"end_date"`
	Projects    []SampleProject `json:"projects"`
	WebSite     string          `json:"web_site"`
	Description string          `json:"description"`
}

type SampleProject struct {
	Project string   `json:"project"`
	Items   []string `json:"items"`
}

type SampleEducation struct {
	ID           string `json:"id"`
	Institution  string `json:"institution"`
	Degree       string `json:"degree"`
	FieldOfStudy string `json:"field_of_study"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Location     string `json:"location"`
	Description  string `json:"description"`
}

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect to DB: %v", err)
	}
	defer pool.Close()

	// Load sample data
	filename := "sample_users.json"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("read sample_users.json: %v", err)
	}

	var samples []SampleUser
	if err := json.Unmarshal(data, &samples); err != nil {
		log.Fatalf("parse sample_users.json: %v", err)
	}

	// Init repos
	userRepo := repository.NewUserRepository(pool)
	pfRepo := repository.NewProfileFieldRepository(pool)
	pftRepo := repository.NewProfileFieldTextRepository(pool)
	expRepo := repository.NewExperienceItemRepository(pool)
	eduRepo := repository.NewEducationItemRepository(pool)
	itRepo := repository.NewItemTextRepository(pool)
	skillRepo := repository.NewSkillRepository(pool)
	cspRepo := repository.NewCandidateSearchProfileRepository(pool)
	companyRefRepo := repository.NewCompanyReferenceRepository(pool)
	universityRefRepo := repository.NewUniversityReferenceRepository(pool)

	// Init services
	userSvc := application.NewUserService(userRepo)
	pfSvc := application.NewProfileFieldService(pfRepo)
	pftSvc := application.NewProfileFieldTextService(pftRepo, nil) // no gemini needed
	expSvc := application.NewExperienceItemService(expRepo)
	eduSvc := application.NewEducationItemService(eduRepo)
	itSvc := application.NewItemTextService(itRepo, nil) // no gemini needed
	skillSvc := application.NewSkillService(skillRepo)

	indexingSvc := application.NewCandidateIndexingService(
		cspRepo, companyRefRepo, universityRefRepo,
		userSvc, pfSvc, pftSvc, expSvc, eduSvc, itSvc, skillSvc,
	)

	created := 0
	skipped := 0

	for i, sample := range samples {
		log.Printf("[%d/%d] Processing %s %s (%s)...", i+1, len(samples), sample.FirstName, sample.LastName, sample.Email)

		// Check if user already exists by email
		if sample.Email != "" {
			if existing, _ := userRepo.GetByEmail(ctx, sample.Email); existing != nil {
				log.Printf("  -> skipped (email exists)")
				skipped++
				// Still try to index
				_ = indexingSvc.IndexCandidate(ctx, existing.ID)
				continue
			}
		}

		// Create user
		user := &domain.User{
			FirstName:               sample.FirstName,
			LastName:                sample.LastName,
			Patronymic:             sample.Patronymic,
			Phone:                  sample.Phone,
			Telegram:               sample.Telegram,
			TelegramID:             sample.TelegramID,
			Email:                  sample.Email,
			Gender:                 sample.Gender,
			Country:                sample.Country,
			Region:                 sample.Region,
			Nationality:            sample.Nationality,
			Status:                 sample.Status,
			TariffType:             sample.TariffType,
			JobStatus:              sample.JobStatus,
			ActivityType:           sample.ActivityType,
			Specializations:        sample.Specializations,
			ProfileScore:           sample.ProfileScore,
			EstimatedSalaryMin:     sample.EstimatedSalaryMin,
			EstimatedSalaryMax:     sample.EstimatedSalaryMax,
			EstimatedSalaryCurrency: sample.EstimatedSalaryCurrency,
			Language:               "ru",
		}

		createdUser, err := userSvc.CreateUser(ctx, user)
		if err != nil {
			log.Printf("  -> ERROR creating user: %v", err)
			continue
		}

		userID := createdUser.ID

		// Set password
		if err := userRepo.SetPassword(ctx, userID, "$2a$10$dummyhashforseeding000000000000000000000000000"); err != nil {
			log.Printf("  -> warning: set password failed: %v", err)
		}

		// Set profile score and salary
		_ = userRepo.SetProfileScore(ctx, userID, sample.ProfileScore)
		_ = userRepo.SetEstimatedSalary(ctx, userID, sample.EstimatedSalaryMin, sample.EstimatedSalaryMax, sample.EstimatedSalaryCurrency)

		// Create profile fields
		profile := sample.Profile

		// Title
		if profile.Title != "" {
			createProfileField(ctx, pfSvc, pftSvc, userID, "title", profile.Title)
		}

		// About
		if profile.About != "" {
			createProfileField(ctx, pfSvc, pftSvc, userID, "about", profile.About)
		}

		// Achievements
		if profile.Achievements != "" {
			createProfileField(ctx, pfSvc, pftSvc, userID, "achievements", profile.Achievements)
		}

		// Certifications as JSON
		if len(profile.Certifications) > 0 {
			certsJSON, _ := json.Marshal(profile.Certifications)
			createProfileField(ctx, pfSvc, pftSvc, userID, "certifications", string(certsJSON))
		}

		// Languages as JSON
		if len(profile.Languages) > 0 {
			var langMaps []map[string]string
			for _, l := range profile.Languages {
				langMaps = append(langMaps, map[string]string{"name": l.Name, "level": l.Level})
			}
			langsJSON, _ := json.Marshal(langMaps)
			createProfileField(ctx, pfSvc, pftSvc, userID, "languages", string(langsJSON))
		}

		// Skills
		if len(profile.Skills) > 0 {
			if _, err := skillSvc.SetUserSkills(ctx, userID, profile.Skills); err != nil {
				log.Printf("  -> warning: set skills failed: %v", err)
			}
		}

		// Experience items
		for j, exp := range profile.Experience {
			projectsJSON := ""
			if len(exp.Projects) > 0 {
				b, _ := json.Marshal(exp.Projects)
				projectsJSON = string(b)
			}

			item, err := expSvc.CreateExperienceItem(ctx, &domain.ExperienceItem{
				UserID:    userID,
				Company:   exp.Company,
				Position:  exp.Position,
				StartDate: exp.StartDate,
				EndDate:   exp.EndDate,
				Projects:  projectsJSON,
				WebSite:   exp.WebSite,
				ItemOrder: int32(j),
			})
			if err != nil {
				log.Printf("  -> warning: create experience %d failed: %v", j, err)
				continue
			}

			if exp.Description != "" {
				_, _ = itSvc.CreateItemText(ctx, &domain.ItemText{
					ItemID:      item.ID,
					ItemType:    "experience",
					Lang:        "en",
					Description: exp.Description,
					IsSource:    true,
				})
			}
		}

		// Education items
		for j, edu := range profile.Education {
			item, err := eduSvc.CreateEducationItem(ctx, &domain.EducationItem{
				UserID:       userID,
				Institution:  edu.Institution,
				Degree:       edu.Degree,
				FieldOfStudy: edu.FieldOfStudy,
				StartDate:    edu.StartDate,
				EndDate:      edu.EndDate,
				Location:     edu.Location,
				ItemOrder:    int32(j),
			})
			if err != nil {
				log.Printf("  -> warning: create education %d failed: %v", j, err)
				continue
			}

			if edu.Description != "" {
				_, _ = itSvc.CreateItemText(ctx, &domain.ItemText{
					ItemID:      item.ID,
					ItemType:    "education",
					Lang:        "en",
					Description: edu.Description,
					IsSource:    true,
				})
			}
		}

		// Index candidate search profile
		if err := indexingSvc.IndexCandidate(ctx, userID); err != nil {
			log.Printf("  -> warning: index failed: %v", err)
		}

		created++
		log.Printf("  -> created + indexed (%s)", userID)
	}

	log.Printf("\nDone! Created: %d, Skipped: %d, Total processed: %d", created, skipped, len(samples))

	// Verify
	var count int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM candidate_search_profiles").Scan(&count)
	log.Printf("candidate_search_profiles rows: %d", count)
}

func createProfileField(ctx context.Context, pfSvc *application.ProfileFieldService, pftSvc *application.ProfileFieldTextService, userID uuid.UUID, fieldName, content string) {
	field, err := pfSvc.CreateProfileField(ctx, &domain.ProfileField{
		UserID:     userID,
		FieldName:  fieldName,
		SourceLang: "en",
	})
	if err != nil {
		return
	}
	_, _ = pftSvc.CreateProfileFieldText(ctx, &domain.ProfileFieldText{
		ProfileFieldID: field.ID,
		Lang:           "en",
		Content:        content,
		IsSource:       true,
	})
}

// Unused imports guard
var _ = scoring.NormalizeRole
var _ = strings.TrimSpace
var _ = time.Now
var _ = fmt.Sprintf
