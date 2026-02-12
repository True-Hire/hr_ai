package app

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/repository"
)

type Services struct {
	User             *application.UserService
	ProfileField     *application.ProfileFieldService
	ProfileFieldText *application.ProfileFieldTextService
}

func NewServices(pool *pgxpool.Pool) *Services {
	return &Services{
		User:             application.NewUserService(repository.NewUserRepository(pool)),
		ProfileField:     application.NewProfileFieldService(repository.NewProfileFieldRepository(pool)),
		ProfileFieldText: application.NewProfileFieldTextService(repository.NewProfileFieldTextRepository(pool)),
	}
}
