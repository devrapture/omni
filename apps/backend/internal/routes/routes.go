package routes

import (
	"github.com/devrapture/omni/internal/config"
	handlers "github.com/devrapture/omni/internal/handler"
	"github.com/devrapture/omni/internal/middleware"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HandlerDependencies struct {
	AuthHandler         *handlers.AuthHandler
	UserSettingsHandler *handlers.SettingsHandler
	FileUploadHandler   *handlers.FileUploadHandler
	BusinessHandler     *handlers.BusinessHandler
}

func Setup(db *gorm.DB, deps HandlerDependencies, cfg *config.Config, logger *zap.Logger) *gin.Engine {
	utils.RegisterValidators()

	r := gin.New()
	r.Use(middleware.RequestLogger(logger))
	r.Use(gin.Recovery())
	v1 := r.Group("/api/v1")

	{
		v1.GET("/health", handlers.HealthHandler(db))
		// auth
		auth := v1.Group("/auth")

		auth.
			GET("/google", deps.AuthHandler.GoogleLogin).
			GET("/google/callback", deps.AuthHandler.GoogleCallback).
			POST("/google/login", deps.AuthHandler.LoginWithGoogle) // when frontend is using Authjs library

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))

		// user settings
		settings := protected.Group("/settings")

		settings.
			GET("", deps.UserSettingsHandler.GetUserSettings).
			POST("", deps.UserSettingsHandler.UpdateUserSettings).
			DELETE("", deps.UserSettingsHandler.DeleteUserSettings)

		// file upload
		fileUpload := protected.Group("/file-upload")

		fileUpload.
			POST("/presign", deps.FileUploadHandler.CreatePresignedUploadURL).
			POST("/jobs/:jobID/complete", deps.FileUploadHandler.CompleteUpload).
			GET("/jobs/:jobID", deps.FileUploadHandler.GetUploadJob)

		// business
		business := protected.Group("/business")

		business.
			POST("", deps.BusinessHandler.CreateBusiness).
			GET("/knowledge/:businessID", deps.BusinessHandler.ListSources).
			DELETE("/knowledge/:businessID", deps.BusinessHandler.DeleteSource).
			POST("/knowledge/:businessID", deps.BusinessHandler.AddText)

	}

	return r
}
