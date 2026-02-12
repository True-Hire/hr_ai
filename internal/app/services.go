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
	ProfileParse     *application.ProfileParseService
}

func NewServices(pool *pgxpool.Pool, geminiAPIKey string) *Services {
	userSvc := application.NewUserService(repository.NewUserRepository(pool))
	pfSvc := application.NewProfileFieldService(repository.NewProfileFieldRepository(pool))
	pftSvc := application.NewProfileFieldTextService(repository.NewProfileFieldTextRepository(pool))

	geminiClient := gemini.NewClient(geminiAPIKey)

	return &Services{
		User:             userSvc,
		ProfileField:     pfSvc,
		ProfileFieldText: pftSvc,
		ProfileParse:     application.NewProfileParseService(geminiClient, pfSvc, pftSvc, userSvc),
	}
}
