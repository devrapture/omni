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
	FileUploadHandler   *handlers.FileUploadHandler
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

		// file upload
		fileUpload := v1.Group("/file-upload")
		fileUpload.Use(middleware.AuthMiddleware(cfg))

		fileUpload.
			// POST("", deps.FileUploadHandler.HandleFileUpload).
			POST("/presign", deps.FileUploadHandler.CreatePresignedUploadURL).
			POST("/jobs/:jobID/complete", deps.FileUploadHandler.CompleteUpload).
			GET("/jobs/:jobID", deps.FileUploadHandler.GetUploadJob)
	}

	return r
}
