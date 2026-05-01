package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/config"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/repository"
)

func main() {
	// 1. Load config
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	// 2. Connect to DB
	var pool *pgxpool.Pool
	for i := 0; i < 5; i++ {
		pool, err = pgxpool.New(ctx, cfg.DatabaseURL)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				break
			}
		}
		log.Printf("Waiting for database... retry %d/5", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// 3. Initialize components
	// Get API keys from env
	geminiKey := os.Getenv("GEMINI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	
	aiClient := gemini.NewClient(geminiKey, anthropicKey)
	vacancyRepo := repository.NewVacancyRepository(pool)
	aiService := application.NewVacancyAIService(aiClient, vacancyRepo)

	// 4. Test text
	testText := `
Assalomu alaykum! Bizga zudlik bilan Backend dev kerak. 
Asosan Go tilida yozamiz, shuning uchun Golang ni yirtadigan bo'lishi shart. 
Baza sifatida PostgreSQL ishlatamiz, Redis va Kafka bilan ishlashni bilsa zo'r bo'lardi. 
Mikroservis arxitekturasi tushunchasi bo'lishi muhim. 
Oylik 1500-2000 dollar atrofida. Biz ofisdamiz, Toshkent markazida. 
Muhimi jamoada ishlay olishi (Teamwork) va muammolarni tez hal qilishi (Problem Solving) kerak.
`

	fmt.Println("--- TESTING AI PARSING ---")
	fmt.Printf("Input Text: %s\n", testText)
	fmt.Println("Waiting for AI response...")

	// 5. Run Parse
	parsed, err := aiService.ParseVacancyText(ctx, testText)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	// 6. Print Results
	fmt.Println("\n--- PARSED RESULT ---")
	prettyJSON, _ := json.MarshalIndent(parsed, "", "  ")
	fmt.Println(string(prettyJSON))

	fmt.Println("\n--- VERIFICATION ---")
	fmt.Printf("Main Category: %s (New: %s)\n", parsed.MatchedMainCatID, parsed.NewMainCategory)
	fmt.Printf("Sub Category:  %s (New: %s)\n", parsed.MatchedSubCatID, parsed.NewSubCategory)
	fmt.Printf("Technologies:  Matched: %v, New: %v\n", parsed.MatchedTechIDs, parsed.NewTechnologies)
	fmt.Printf("Skills:        Matched: %v, New: %v\n", parsed.MatchedSkillIDs, parsed.NewSkills)
}
