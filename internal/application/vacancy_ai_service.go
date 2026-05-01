package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/repository"
)

type VacancyAIService struct {
	aiClient    *gemini.Client
	vacancyRepo *repository.VacancyAIRepository
}

func NewVacancyAIService(aiClient *gemini.Client, repo *repository.VacancyAIRepository) *VacancyAIService {
	return &VacancyAIService{
		aiClient:    aiClient,
		vacancyRepo: repo,
	}
}

func (s *VacancyAIService) ParseVacancyText(ctx context.Context, userInput string) (*domain.AIParsedVacancy, error) {
	// 1. Get all taxonomy from DB to cache in AI
	dbReferences, err := s.vacancyRepo.GetAllReferencesForAI(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch references for AI cache: %w", err)
	}
	log.Printf("[VacancyAIService] DB References sent to AI:\n%s\n", dbReferences)

	// 2. Build the exact system instruction
	systemPrompt := fmt.Sprintf(`Sen professional HR-Tech tahlilchisan. Sening vazifang vakansiya matnidan quyidagi 4 ta toifani MUTLAQO ANIQ va ALOHIDA ajratib olish:

### QAT'IY QOIDALAR (STRICT RULES):
1. **UNIVERSAL TAXONOMY**: Bu qoidalar barcha sohalarga tegishli (IT, Dizayn, Video, Montaj, va h.k.).
2. **TECHNOLOGY (TEXNOLOGIYA)**: Bu ish bajarish uchun asbob, dastur yoki uskuna.
   - *IT*: Swift, Go, Firebase, Docker.
   - *Design/Video*: Figma, Premiere Pro, After Effects, DaVinci, Camera, Drone.
3. **SKILL (KO'NIKMA)**: Bu bilim, metodologiya yoki insoniy qobiliyat.
   - *IT*: OOP, Clean Code, Agile.
   - *Design/Video*: Color Grading, Typography, Storytelling, Prototyping.
4. **LANGUAGE**: Ismlarni har doim INGLIZ TILIDA (English) ajratib ol!

**MUHIM**: Agar tushuncha (masalan: Figma) bazada "Skill" bo'limida bo'lsa ham, uni "matched_skill_ids"ga QO'SHMA! Uni faqat Texnologiya deb tahlil qil va "new_technologies"ga yoz.

--- MA'LUMOTLAR BAZASI (MAVJUD ID-LAR) ---
%s
--- BAZA TUGADI ---

Faqat va faqat JSON qaytar!
JSON formati:
{
  "title": "Vakansiya nomi",
  "description": "Vakansiya haqida qisqa ma'lumot",
  "matched_main_category_id": "UUID yoki empty",
  "matched_sub_category_id": "UUID yoki empty",
  "new_main_category": "IT (faqat bazada bo'lmasa)",
  "new_sub_category": "Backend (faqat bazada bo'lmasa)",
  "matched_technology_ids": ["UUIDs"],
  "matched_skill_ids": ["UUIDs"],
  "new_technologies": ["Figma", "Photoshop"],
  "new_skills": ["User Research", "Teamwork"]
}`, dbReferences)

	// 3. Call Claude API with System Cache
	log.Printf("[VacancyAIService] Input Text: %s\n", userInput)
	log.Printf("[VacancyAIService] Sending request to AI. Prompt length: %d\n", len(systemPrompt))
	jsonRespStr, err := s.aiClient.GenerateJSONWithSystemCache(ctx, systemPrompt, userInput)
	if err != nil {
		return nil, fmt.Errorf("ai generation failed: %w", err)
	}
	log.Printf("[VacancyAIService] Raw AI Response: %s\n", jsonRespStr)

	// 4. Parse JSON
	var parsed domain.AIParsedVacancy
	if err := json.Unmarshal([]byte(jsonRespStr), &parsed); err != nil {
		log.Printf("AI JSON Parse Error: %v\nRaw AI Response: %s", err, jsonRespStr)
		return nil, fmt.Errorf("failed to parse ai json: %w", err)
	}

	// 5. Post-process to fix AI categorization errors
	s.fixCategorization(&parsed)

	return &parsed, nil
}

func (s *VacancyAIService) fixCategorization(p *domain.AIParsedVacancy) {
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

	finalSkills := []string{}
	for _, skill := range p.NewSkills {
		lowerSkill := strings.ToLower(skill)
		isTech := false
		for _, kw := range techKeywords {
			if strings.Contains(lowerSkill, kw) {
				isTech = true
				break
			}
		}

		if isTech {
			p.NewTechnologies = append(p.NewTechnologies, skill)
		} else {
			finalSkills = append(finalSkills, skill)
		}
	}
	p.NewSkills = finalSkills
}
