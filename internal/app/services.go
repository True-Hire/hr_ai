package app

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
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
	JWTSecret        string
}

func NewServices(pool *pgxpool.Pool, geminiAPIKey, jwtSecret string) *Services {
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	userSvc := application.NewUserService(userRepo)
	pfSvc := application.NewProfileFieldService(repository.NewProfileFieldRepository(pool))
	pftSvc := application.NewProfileFieldTextService(repository.NewProfileFieldTextRepository(pool))
	expSvc := application.NewExperienceItemService(repository.NewExperienceItemRepository(pool))
	eduSvc := application.NewEducationItemService(repository.NewEducationItemRepository(pool))
	itSvc := application.NewItemTextService(repository.NewItemTextRepository(pool))
	skillSvc := application.NewSkillService(repository.NewSkillRepository(pool))

	companyHRRepo := repository.NewCompanyHRRepository(pool)
	hrSessionRepo := repository.NewHRSessionRepository(pool)

	geminiClient := gemini.NewClient(geminiAPIKey)

	return &Services{
		User:             userSvc,
		ProfileField:     pfSvc,
		ProfileFieldText: pftSvc,
		ExperienceItem:   expSvc,
		EducationItem:    eduSvc,
		ItemText:         itSvc,
		Skill:            skillSvc,
		ProfileParse:     application.NewProfileParseService(geminiClient, pfSvc, pftSvc, expSvc, eduSvc, itSvc, skillSvc, userSvc),
		Auth:             application.NewAuthService(userRepo, sessionRepo, jwtSecret),
		CompanyHR:        application.NewCompanyHRService(companyHRRepo),
		HRAuth:           application.NewHRAuthService(companyHRRepo, hrSessionRepo, jwtSecret),
		JWTSecret:        jwtSecret,
	}
}
