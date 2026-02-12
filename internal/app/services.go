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
	ProfileParse     *application.ProfileParseService
}

func NewServices(pool *pgxpool.Pool, geminiAPIKey string) *Services {
	userSvc := application.NewUserService(repository.NewUserRepository(pool))
	pfSvc := application.NewProfileFieldService(repository.NewProfileFieldRepository(pool))
	pftSvc := application.NewProfileFieldTextService(repository.NewProfileFieldTextRepository(pool))
	expSvc := application.NewExperienceItemService(repository.NewExperienceItemRepository(pool))
	eduSvc := application.NewEducationItemService(repository.NewEducationItemRepository(pool))
	itSvc := application.NewItemTextService(repository.NewItemTextRepository(pool))

	geminiClient := gemini.NewClient(geminiAPIKey)

	return &Services{
		User:             userSvc,
		ProfileField:     pfSvc,
		ProfileFieldText: pftSvc,
		ExperienceItem:   expSvc,
		EducationItem:    eduSvc,
		ItemText:         itSvc,
		ProfileParse:     application.NewProfileParseService(geminiClient, pfSvc, pftSvc, expSvc, eduSvc, itSvc, userSvc),
	}
}
