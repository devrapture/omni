package routes

import (
	"github.com/devrapture/omni/internal/config"
	handlers "github.com/devrapture/omni/internal/handler"
	"github.com/devrapture/omni/internal/middleware"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HandlerDependencies struct {
	AuthHandler         *handlers.AuthHandler
	UserSettingsHandler *handlers.SettingsHandler
}

func Setup(db *gorm.DB, deps HandlerDependencies, cfg *config.Config) *gin.Engine {
	utils.RegisterValidators()

	r := gin.Default()

	v1 := r.Group("/api/v1")

	{
		v1.GET("/health", handlers.HealthHandler(db))
		// auth
		auth := v1.Group("/auth")

		auth.
			GET("/google", deps.AuthHandler.GoogleLogin).
			GET("/google/callback", deps.AuthHandler.GoogleCallback).
			POST("/google/login", deps.AuthHandler.LoginWithGoogle) // when frontend is using Authjs library

		// user settings
		settings := v1.Group("/settings")
		settings.Use(middleware.AuthMiddleware(cfg))

		settings.
			GET("", deps.UserSettingsHandler.GetUserSettings).
			POST("", deps.UserSettingsHandler.UpdateUserSettings).
			DELETE("", deps.UserSettingsHandler.DeleteUserSettings)

	}

	return r
}
