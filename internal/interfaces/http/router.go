package http

import (
	"github.com/gin-gonic/gin"

	"github.com/ruziba3vich/hr-ai/internal/app"
)

func NewRouter(svc *app.Services) *gin.Engine {
	router := gin.Default()

	userHandler := NewUserHandler(svc.User)
	profileFieldHandler := NewProfileFieldHandler(svc.ProfileField)
	profileFieldTextHandler := NewProfileFieldTextHandler(svc.ProfileFieldText)

	v1 := router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
			users.POST("/:id/profile-fields", profileFieldHandler.Create)
			users.GET("/:id/profile-fields", profileFieldHandler.ListByUser)
		}

		profileFields := v1.Group("/profile-fields")
		{
			profileFields.GET("/:id", profileFieldHandler.GetByID)
			profileFields.PUT("/:id", profileFieldHandler.Update)
			profileFields.DELETE("/:id", profileFieldHandler.Delete)
			profileFields.POST("/:id/texts", profileFieldTextHandler.Create)
			profileFields.GET("/:id/texts", profileFieldTextHandler.ListByField)
			profileFields.GET("/:id/texts/:lang", profileFieldTextHandler.Get)
			profileFields.PUT("/:id/texts/:lang", profileFieldTextHandler.Update)
			profileFields.DELETE("/:id/texts/:lang", profileFieldTextHandler.Delete)
		}
	}

	return router
}
