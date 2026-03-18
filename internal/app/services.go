package app

import (
	"context"
	"fmt"

	casbinlib "github.com/casbin/casbin/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ruziba3vich/hr-ai/internal/application"
	casbininfra "github.com/ruziba3vich/hr-ai/internal/infrastructure/casbin"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
	minioclient "github.com/ruziba3vich/hr-ai/internal/infrastructure/minio"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/qdrant"
	redisclient "github.com/ruziba3vich/hr-ai/internal/infrastructure/redis"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/repository"
)

type Services struct {
	User             *application.UserService
	ProfileField     *application.ProfileFieldService
	ProfileFieldText *application.ProfileFieldTextService
	ExperienceItem   *application.ExperienceItemService
	EducationItem    *application.EducationItemService
	ItemText         *application.ItemTextService
	Skill            *application.SkillService
	ProfileParse     *application.ProfileParseService
	Auth             *application.AuthService
	CompanyHR        *application.CompanyHRService
	HRAuth           *application.HRAuthService
	Company          *application.CompanyService
	CompanyText      *application.CompanyTextService
	Country          *application.CountryService
	Vacancy          *application.VacancyService
	VacancyText      *application.VacancyTextService
	VectorIndex      *application.VectorIndexService
	Search           *application.SearchService
	VacancySearch       *application.VacancySearchService
	VacancyApplication  *application.VacancyApplicationService
	Bot                 *application.BotService
	HRBot               *application.HRBotService
	AccountDeletion     *application.AccountDeletionService
	HRSavedUser         *application.HRSavedUserService
	CandidateIndexing   *application.CandidateIndexingService
	CandidateSearch     *application.CandidateSearchService
	Storage          *application.StorageService
	GeminiClient     *gemini.Client
	CasbinEnforcer   *casbinlib.Enforcer
	RedisClient      *redisclient.Client
	JWTSecret        string
	TelegramBotToken   string
	TelegramHRBotToken string
}

func NewServices(pool *pgxpool.Pool, geminiAPIKey, jwtSecret, databaseURL, qdrantURL, qdrantAPIKey, redisURL, minioEndpoint, minioAccessKey, minioSecretKey, minioBucket string, minioUseSSL bool, minioPublicURL, telegramBotToken, telegramHRBotToken string) (*Services, error) {
	rc, err := redisclient.NewClient(redisURL)
	if err != nil {
		return nil, fmt.Errorf("init redis: %w", err)
	}
	cacheSvc := application.NewCacheService(rc)

	mc, err := minioclient.NewClient(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, minioUseSSL, minioPublicURL)
	if err != nil {
		rc.Close()
		return nil, fmt.Errorf("init minio: %w", err)
	}
	storageSvc := application.NewStorageService(mc)

	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	userSvc := application.NewUserService(userRepo)
	geminiClient := gemini.NewClient(geminiAPIKey)

	pfRepo := repository.NewProfileFieldRepository(pool)
	pftRepo := repository.NewProfileFieldTextRepository(pool)
	expRepo := repository.NewExperienceItemRepository(pool)
	eduRepo := repository.NewEducationItemRepository(pool)
	itRepo := repository.NewItemTextRepository(pool)
	skillRepo := repository.NewSkillRepository(pool)

	pfSvc := application.NewProfileFieldService(pfRepo)
	pftSvc := application.NewProfileFieldTextService(pftRepo, geminiClient)
	expSvc := application.NewExperienceItemService(expRepo)
	eduSvc := application.NewEducationItemService(eduRepo)
	itSvc := application.NewItemTextService(itRepo, geminiClient)
	skillSvc := application.NewSkillService(skillRepo)

	companyHRRepo := repository.NewCompanyHRRepository(pool)
	hrSessionRepo := repository.NewHRSessionRepository(pool)

	companyRepo := repository.NewCompanyRepository(pool)
	companyTextRepo := repository.NewCompanyTextRepository(pool)

	countryRepo := repository.NewCountryRepository(pool)
	countryTextRepo := repository.NewCountryTextRepository(pool)

	vacancyRepo := repository.NewVacancyRepository(pool)
	vacancyTextRepo := repository.NewVacancyTextRepository(pool)
	vacancyAppRepo := repository.NewVacancyApplicationRepository(pool)

	qdrantClient := qdrant.NewClient(qdrantURL, qdrantAPIKey)
	if err := qdrantClient.EnsureCollection(context.Background(), "user_profile_vectors", 768); err != nil {
		return nil, fmt.Errorf("ensure qdrant collection: %w", err)
	}
	if err := qdrantClient.EnsureCollection(context.Background(), "vacancy_vectors", 768); err != nil {
		return nil, fmt.Errorf("ensure vacancy qdrant collection: %w", err)
	}

	vectorIndexSvc := application.NewVectorIndexService(qdrantClient, geminiClient, pfSvc, pftSvc, expSvc, itSvc, skillSvc, userSvc, vacancyRepo, vacancyTextRepo)

	companySvc := application.NewCompanyService(companyRepo, companyTextRepo, geminiClient, cacheSvc)

	enforcer, err := casbininfra.NewEnforcer(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("init casbin enforcer: %w", err)
	}

	profileParseSvc := application.NewProfileParseService(geminiClient, pfSvc, pftSvc, expSvc, eduSvc, itSvc, skillSvc, userSvc, vectorIndexSvc)

	vacancySvc := application.NewVacancyService(vacancyRepo, vacancyTextRepo, skillSvc, geminiClient, vectorIndexSvc)
	vacancySearchSvc := application.NewVacancySearchService(qdrantClient, geminiClient, vacancySvc, pfSvc, pftSvc, skillSvc)
	searchSvc := application.NewSearchService(qdrantClient, geminiClient, userSvc, vacancySvc, expSvc, skillSvc)
	vacancyAppSvc := application.NewVacancyApplicationService(vacancyAppRepo)

	companyHRSvc := application.NewCompanyHRService(companyHRRepo)
	accountDeletionSvc := application.NewAccountDeletionService(
		userRepo, companyHRRepo, sessionRepo, hrSessionRepo,
		vacancyRepo, vacancyAppRepo,
		pfRepo, pftRepo, expRepo, eduRepo, itRepo, skillRepo,
		vectorIndexSvc,
	)
	botStateSvc := application.NewBotStateService(rc)

	hrSavedUserRepo := repository.NewHRSavedUserRepository(pool)
	hrSavedUserSvc := application.NewHRSavedUserService(hrSavedUserRepo)

	// Candidate search & scoring
	cspRepo := repository.NewCandidateSearchProfileRepository(pool)
	companyRefRepo := repository.NewCompanyReferenceRepository(pool)
	universityRefRepo := repository.NewUniversityReferenceRepository(pool)
	searchSessionRepo := repository.NewSearchSessionRepository(pool)

	candidateIndexingSvc := application.NewCandidateIndexingService(
		cspRepo, companyRefRepo, universityRefRepo,
		userSvc, pfSvc, pftSvc, expSvc, eduSvc, itSvc, skillSvc,
	)
	candidateSearchSvc := application.NewCandidateSearchService(
		cspRepo, searchSessionRepo, userSvc, skillSvc, vacancySvc, expSvc,
	)

	// Hook indexing into profile parse
	profileParseSvc.SetCandidateIndexingService(candidateIndexingSvc)

	hrBotSvc := application.NewHRBotService(companyHRSvc, vacancySvc, vacancyAppSvc, botStateSvc, searchSvc, userSvc, geminiClient)

	return &Services{
		User:             userSvc,
		ProfileField:     pfSvc,
		ProfileFieldText: pftSvc,
		ExperienceItem:   expSvc,
		EducationItem:    eduSvc,
		ItemText:         itSvc,
		Skill:            skillSvc,
		ProfileParse:     profileParseSvc,
		Auth:             application.NewAuthService(userRepo, sessionRepo, jwtSecret),
		CompanyHR:        companyHRSvc,
		HRAuth:           application.NewHRAuthService(companyHRRepo, hrSessionRepo, jwtSecret),
		Country:          application.NewCountryService(countryRepo, countryTextRepo, geminiClient, cacheSvc),
		Company:          companySvc,
		CompanyText:      application.NewCompanyTextService(companyTextRepo),
		Vacancy:          vacancySvc,
		VacancyText:      application.NewVacancyTextService(vacancyTextRepo),
		VectorIndex:      vectorIndexSvc,
		Search:           searchSvc,
		VacancySearch:      vacancySearchSvc,
		VacancyApplication: vacancyAppSvc,
		Bot:                application.NewBotService(userSvc, companyHRSvc, profileParseSvc, storageSvc, botStateSvc, geminiClient),
		HRBot:              hrBotSvc,
		AccountDeletion:    accountDeletionSvc,
		HRSavedUser:        hrSavedUserSvc,
		CandidateIndexing:  candidateIndexingSvc,
		CandidateSearch:    candidateSearchSvc,
		Storage:          storageSvc,
		GeminiClient:     geminiClient,
		CasbinEnforcer:   enforcer,
		RedisClient:      rc,
		JWTSecret:        jwtSecret,
		TelegramBotToken:   telegramBotToken,
		TelegramHRBotToken: telegramHRBotToken,
	}, nil
}
