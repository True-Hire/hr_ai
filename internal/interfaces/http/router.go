package http

import (
	"github.com/gin-gonic/gin"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

func NewRouter(userService *application.UserService, profileFieldService *application.ProfileFieldService) *gin.Engine {
	router := gin.Default()

	userHandler := NewUserHandler(userService)
	profileFieldHandler := NewProfileFieldHandler(profileFieldService)

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
		}
	}

	return router
}
