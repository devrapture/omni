package routes

import (
	handlers "github.com/devrapture/omni/internal/handler"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HandlerDependencies struct {
	AuthHandler *handlers.AuthHandler
}

func Setup(db *gorm.DB, deps HandlerDependencies) *gin.Engine {
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
	}

	return r
}
