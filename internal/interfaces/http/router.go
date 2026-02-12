package http

import (
	"github.com/gin-gonic/gin"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

func NewRouter(
	userService *application.UserService,
	profileFieldService *application.ProfileFieldService,
	profileFieldTextService *application.ProfileFieldTextService,
) *gin.Engine {
	router := gin.Default()

	userHandler := NewUserHandler(userService)
	profileFieldHandler := NewProfileFieldHandler(profileFieldService)
	profileFieldTextHandler := NewProfileFieldTextHandler(profileFieldTextService)

	v1 := router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)

			users.POST("/:user_id/profile-fields", profileFieldHandler.Create)
			users.GET("/:user_id/profile-fields", profileFieldHandler.ListByUser)
		}

		profileFields := v1.Group("/profile-fields")
		{
			profileFields.GET("/:id", profileFieldHandler.GetByID)
			profileFields.PUT("/:id", profileFieldHandler.Update)
			profileFields.DELETE("/:id", profileFieldHandler.Delete)

			profileFields.POST("/:field_id/texts", profileFieldTextHandler.Create)
			profileFields.GET("/:field_id/texts", profileFieldTextHandler.ListByField)
			profileFields.GET("/:field_id/texts/:lang", profileFieldTextHandler.Get)
			profileFields.PUT("/:field_id/texts/:lang", profileFieldTextHandler.Update)
			profileFields.DELETE("/:field_id/texts/:lang", profileFieldTextHandler.Delete)
		}
	}

	return router
}
