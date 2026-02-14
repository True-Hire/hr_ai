package app

import (
	"context"
	"fmt"

	casbinlib "github.com/casbin/casbin/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ruziba3vich/hr-ai/internal/application"
	casbininfra "github.com/ruziba3vich/hr-ai/internal/infrastructure/casbin"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/qdrant"
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
	CasbinEnforcer   *casbinlib.Enforcer
	JWTSecret        string
}

func NewServices(pool *pgxpool.Pool, geminiAPIKey, jwtSecret, databaseURL, qdrantURL, qdrantAPIKey string) (*Services, error) {
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	userSvc := application.NewUserService(userRepo)
	geminiClient := gemini.NewClient(geminiAPIKey)

	pfSvc := application.NewProfileFieldService(repository.NewProfileFieldRepository(pool))
	pftSvc := application.NewProfileFieldTextService(repository.NewProfileFieldTextRepository(pool), geminiClient)
	expSvc := application.NewExperienceItemService(repository.NewExperienceItemRepository(pool))
	eduSvc := application.NewEducationItemService(repository.NewEducationItemRepository(pool))
	itSvc := application.NewItemTextService(repository.NewItemTextRepository(pool), geminiClient)
	skillSvc := application.NewSkillService(repository.NewSkillRepository(pool))

	companyHRRepo := repository.NewCompanyHRRepository(pool)
	hrSessionRepo := repository.NewHRSessionRepository(pool)

	companyRepo := repository.NewCompanyRepository(pool)
	companyTextRepo := repository.NewCompanyTextRepository(pool)

	countryRepo := repository.NewCountryRepository(pool)
	countryTextRepo := repository.NewCountryTextRepository(pool)

	vacancyRepo := repository.NewVacancyRepository(pool)
	vacancyTextRepo := repository.NewVacancyTextRepository(pool)

	qdrantClient := qdrant.NewClient(qdrantURL, qdrantAPIKey)
	if err := qdrantClient.EnsureCollection(context.Background(), "user_profile_vectors", 768); err != nil {
		return nil, fmt.Errorf("ensure qdrant collection: %w", err)
	}

	vectorIndexSvc := application.NewVectorIndexService(qdrantClient, geminiClient, pfSvc, pftSvc, expSvc, itSvc, skillSvc, userSvc)
	searchSvc := application.NewSearchService(qdrantClient, geminiClient, userSvc)

	companySvc := application.NewCompanyService(companyRepo, companyTextRepo, geminiClient)

	enforcer, err := casbininfra.NewEnforcer(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("init casbin enforcer: %w", err)
	}

	return &Services{
		User:             userSvc,
		ProfileField:     pfSvc,
		ProfileFieldText: pftSvc,
		ExperienceItem:   expSvc,
		EducationItem:    eduSvc,
		ItemText:         itSvc,
		Skill:            skillSvc,
		ProfileParse:     application.NewProfileParseService(geminiClient, pfSvc, pftSvc, expSvc, eduSvc, itSvc, skillSvc, userSvc, vectorIndexSvc),
		Auth:             application.NewAuthService(userRepo, sessionRepo, jwtSecret),
		CompanyHR:        application.NewCompanyHRService(companyHRRepo),
		HRAuth:           application.NewHRAuthService(companyHRRepo, hrSessionRepo, jwtSecret),
		Country:          application.NewCountryService(countryRepo, countryTextRepo, geminiClient),
		Company:          companySvc,
		CompanyText:      application.NewCompanyTextService(companyTextRepo),
		Vacancy:          application.NewVacancyService(vacancyRepo, vacancyTextRepo, skillSvc, companySvc, geminiClient),
		VacancyText:      application.NewVacancyTextService(vacancyTextRepo),
		VectorIndex:      vectorIndexSvc,
		Search:           searchSvc,
		CasbinEnforcer:   enforcer,
		JWTSecret:        jwtSecret,
	}, nil
}
