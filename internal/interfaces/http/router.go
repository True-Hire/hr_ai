package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/ruziba3vich/hr-ai/internal/app"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			allowed := []string{
				"https://hr-ai-wb-app.leetcoders.uz",
				"https://hr-ai.compile-me.uz",
				"https://transport-factor-prior.ngrok-free.dev",
				"http://localhost:11911",
				"http://127.0.0.1:11911",
				"http://localhost:5173",
				"http://127.0.0.1:5173",
				"http://localhost:3000",
				"http://127.0.0.1:3000",
			}
			for _, a := range allowed {
				if strings.EqualFold(origin, a) {
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Telegram-Init-Data")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func NewRouter(svc *app.Services) *gin.Engine {
	router := gin.Default()
	router.Use(corsMiddleware())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userHandler := NewUserHandler(svc.User, svc.ProfileField, svc.ProfileFieldText, svc.ExperienceItem, svc.EducationItem, svc.ItemText, svc.Skill, svc.Auth, svc.Search, svc.CandidateSearch)
	profileFieldHandler := NewProfileFieldHandler(svc.ProfileField)
	profileFieldTextHandler := NewProfileFieldTextHandler(svc.ProfileFieldText)
	profileParseHandler := NewProfileParseHandler(svc.ProfileParse)
	skillHandler := NewSkillHandler(svc.Skill)
	authHandler := NewAuthHandler(svc.Auth)
	companyHRHandler := NewCompanyHRHandler(svc.CompanyHR, svc.HRAuth)
	hrAuthHandler := NewHRAuthHandler(svc.HRAuth)
	companyHandler := NewCompanyHandler(svc.Company)
	vacancyHandler := NewVacancyHandler(svc.Vacancy, svc.CompanyHR, svc.VacancySearch, svc.VacancyApplication, svc.Search)
	countryHandler := NewCountryHandler(svc.Country)
	storageHandler := NewStorageHandler(svc.Storage)
	searchHandler := NewSearchHandler(svc.Search, svc.VectorIndex, userHandler)
	accountDeletionHandler := NewAccountDeletionHandler(svc.AccountDeletion)
	miniAppHandler := NewMiniAppHandler(svc.VacancySearch, svc.Vacancy, svc.VacancyApplication,
		svc.User, svc.ProfileField, svc.ProfileFieldText,
		svc.ExperienceItem, svc.EducationItem, svc.ItemText, svc.Skill)
	hrVacancyAppHandler := NewHRVacancyApplicationsHandler(svc.VacancyApplication, svc.Vacancy, svc.User, svc.Skill)
	hrSavedUsersHandler := NewHRSavedUsersHandler(svc.HRSavedUser, svc.User, svc.Skill)
	candidateSearchHandler := NewCandidateSearchHandler(svc.CandidateSearch, userHandler)
	normRuleHandler := NewNormalizationRuleHandler(svc.NormalizationRule)
	hrCombinedAuth := HRCombinedAuthMiddleware(svc.JWTSecret, svc.TelegramHRBotToken, svc.CompanyHR)

	// Serve Mini App HTML
	router.GET("/web/app", func(c *gin.Context) {
		c.File("web/app.html")
	})

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/set-password", authHandler.SetPassword)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", AuthMiddleware(svc.JWTSecret), authHandler.Logout)
		}

		users := v1.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/me", AuthMiddleware(svc.JWTSecret), userHandler.Me)
			users.POST("/me/profile/parse", AuthMiddleware(svc.JWTSecret), profileParseHandler.Parse)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
			users.DELETE("/by-phone/:phone", accountDeletionHandler.DeleteUserByPhone)
			users.POST("/:id/profile-fields", profileFieldHandler.Create)
			users.GET("/:id/profile-fields", profileFieldHandler.ListByUser)
		}

		v1.GET("/skills", skillHandler.Search)

		hrAuth := v1.Group("/hr/auth")
		{
			hrAuth.POST("/set-password", hrAuthHandler.SetPassword)
			hrAuth.POST("/login", hrAuthHandler.Login)
			hrAuth.POST("/refresh", hrAuthHandler.Refresh)
			hrAuth.POST("/logout", HRAuthMiddleware(svc.JWTSecret), hrAuthHandler.Logout)
		}

		hrs := v1.Group("/hrs")
		{
			hrs.POST("", companyHRHandler.Create)
			hrs.GET("", companyHRHandler.List)
			hrs.GET("/me", HRAuthMiddleware(svc.JWTSecret), companyHRHandler.Me)
			hrs.GET("/:id", companyHRHandler.GetByID)
			hrs.PUT("/:id", companyHRHandler.Update)
			hrs.DELETE("/:id", companyHRHandler.Delete)
			hrs.DELETE("/by-phone/:phone", accountDeletionHandler.DeleteHRByPhone)
		}

		companies := v1.Group("/companies")
		{
			companies.POST("", companyHandler.Create)
			companies.GET("", companyHandler.List)
			companies.GET("/:id", companyHandler.GetByID)
			companies.PUT("/:id", companyHandler.Update)
			companies.DELETE("/:id", companyHandler.Delete)
		}

		countries := v1.Group("/countries")
		{
			countries.GET("", countryHandler.List)
			countries.GET("/:id", countryHandler.GetByID)
		}

		vacancies := v1.Group("/vacancies")
		{
			vacancies.POST("", HRAuthMiddleware(svc.JWTSecret), CasbinMiddleware(svc.CasbinEnforcer, "vacancies", "create"), vacancyHandler.Create)
			vacancies.POST("/parse", HRAuthMiddleware(svc.JWTSecret), CasbinMiddleware(svc.CasbinEnforcer, "vacancies", "create"), vacancyHandler.Parse)
			vacancies.GET("", JWTMiddleware(svc.JWTSecret), CasbinMiddleware(svc.CasbinEnforcer, "vacancies", "read"), vacancyHandler.List)
			vacancies.GET("/:id", JWTMiddleware(svc.JWTSecret), CasbinMiddleware(svc.CasbinEnforcer, "vacancies", "read"), vacancyHandler.GetByID)
			vacancies.PUT("/:id", HRAuthMiddleware(svc.JWTSecret), CasbinMiddleware(svc.CasbinEnforcer, "vacancies", "update"), vacancyHandler.Update)
			vacancies.DELETE("/:id", HRAuthMiddleware(svc.JWTSecret), CasbinMiddleware(svc.CasbinEnforcer, "vacancies", "delete"), vacancyHandler.Delete)

			// User JWT: apply & check applications
			vacancies.POST("/:id/apply", AuthMiddleware(svc.JWTSecret), miniAppHandler.Apply)
			vacancies.GET("/:id/application", AuthMiddleware(svc.JWTSecret), miniAppHandler.GetApplicationStatus)

			// HR JWT: manage applications
			vacancies.GET("/:id/applications", HRAuthMiddleware(svc.JWTSecret), hrVacancyAppHandler.ListApplicants)
			vacancies.GET("/:id/applications/stats", HRAuthMiddleware(svc.JWTSecret), hrVacancyAppHandler.GetStats)
			vacancies.PUT("/:id/applications/:app_id/status", HRAuthMiddleware(svc.JWTSecret), hrVacancyAppHandler.UpdateStatus)
			vacancies.PUT("/:id/applications/:app_id/seen", HRAuthMiddleware(svc.JWTSecret), hrVacancyAppHandler.MarkSeen)
		}

		// User JWT: list my applications
		v1.GET("/applications", AuthMiddleware(svc.JWTSecret), miniAppHandler.ListMyApplications)

		search := v1.Group("/search")
		{
			search.GET("/users", searchHandler.SearchUsers)
			search.POST("/reindex", HRAuthMiddleware(svc.JWTSecret), searchHandler.ReindexAll)
		}

		files := v1.Group("/files")
		{
			files.POST("", storageHandler.Upload)
			files.GET("", storageHandler.Get)
			files.DELETE("", storageHandler.Delete)
		}

		profileFields := v1.Group("/profile-fields")
		{
			profileFields.GET("/:id", profileFieldHandler.GetByID)
			profileFields.PUT("/:id", profileFieldHandler.Update)
			profileFields.DELETE("/:id", profileFieldHandler.Delete)
			profileFields.POST("/:id/texts", profileFieldTextHandler.Create)
			profileFields.GET("/:id/texts", profileFieldTextHandler.ListByField)
			profileFields.GET("/:id/texts/:lang", profileFieldTextHandler.Get)
			profileFields.PUT("/:id/texts", profileFieldTextHandler.Update)
			profileFields.DELETE("/:id/texts/:lang", profileFieldTextHandler.Delete)
		}

		miniapp := v1.Group("/miniapp")
		miniapp.Use(TelegramAuthMiddleware(svc.TelegramBotToken, svc.User, svc.CompanyHR))
		{
			miniapp.GET("/me", miniAppHandler.GetProfile)
			miniapp.GET("/vacancies", miniAppHandler.ListForUser)
			miniapp.GET("/vacancies/search", miniAppHandler.Search)
			miniapp.GET("/vacancies/:id", miniAppHandler.GetByID)
			miniapp.POST("/vacancies/:id/apply", miniAppHandler.Apply)
			miniapp.GET("/vacancies/:id/application", miniAppHandler.GetApplicationStatus)
			miniapp.GET("/applications", miniAppHandler.ListMyApplications)
		}

		hrMiniAppHandler := NewHRMiniAppHandler(svc.CompanyHR, svc.GeminiClient)

		hrMiniapp := v1.Group("/hr-miniapp")
		hrMiniapp.Use(TelegramHRAuthMiddleware(svc.TelegramHRBotToken, svc.CompanyHR))
		{
			hrMiniapp.GET("/me", hrMiniAppHandler.GetMe)
			hrMiniapp.PUT("/me", hrMiniAppHandler.UpdateMe)
			hrMiniapp.GET("/hrs", companyHRHandler.List)
			hrMiniapp.GET("/hrs/:id", companyHRHandler.GetByID)
			hrMiniapp.PUT("/hrs/:id", companyHRHandler.Update)
			hrMiniapp.DELETE("/hrs/:id", companyHRHandler.Delete)
			hrMiniapp.GET("/vacancies", vacancyHandler.List)
			hrMiniapp.GET("/vacancies/:id", vacancyHandler.GetByID)
			hrMiniapp.GET("/vacancies/:id/applications", hrVacancyAppHandler.ListApplicants)
			hrMiniapp.GET("/vacancies/:id/applications/stats", hrVacancyAppHandler.GetStats)
			hrMiniapp.PUT("/vacancies/:id/applications/:app_id/status", hrVacancyAppHandler.UpdateStatus)
			hrMiniapp.PUT("/vacancies/:id/applications/:app_id/seen", hrVacancyAppHandler.MarkSeen)
		}

		// Candidate search — combined auth: works with both TG miniapp and JWT
		candidateSearch := v1.Group("/candidate-search")
		candidateSearch.Use(hrCombinedAuth)
		{
			candidateSearch.POST("", candidateSearchHandler.Search)
			candidateSearch.POST("/by-vacancy/:vacancy_id", candidateSearchHandler.SearchByVacancy)
			candidateSearch.GET("/:search_id", candidateSearchHandler.GetPage)
		}

		// Normalization rules CRUD
		normRules := v1.Group("/normalization-rules")
		normRules.Use(HRAuthMiddleware(svc.JWTSecret))
		{
			normRules.POST("", normRuleHandler.Create)
			normRules.GET("", normRuleHandler.List)
			normRules.GET("/:id", normRuleHandler.GetByID)
			normRules.PUT("/:id", normRuleHandler.Update)
			normRules.DELETE("/:id", normRuleHandler.Delete)
		}

		// HR saved users — combined auth: works with both TG miniapp and JWT
		savedUsers := v1.Group("/hr/saved-users")
		savedUsers.Use(hrCombinedAuth)
		{
			savedUsers.POST("", hrSavedUsersHandler.Save)
			savedUsers.GET("", hrSavedUsersHandler.List)
			savedUsers.DELETE("/:user_id", hrSavedUsersHandler.Delete)
		}
	}

	return router
}
