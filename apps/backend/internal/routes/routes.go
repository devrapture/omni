package routes

import (
	"github.com/devrapture/omni/internal/handler"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB) *gin.Engine {
	utils.RegisterValidators()

	r := gin.Default()

	v1 := r.Group("/api/v1")

	{
		v1.GET("/health", handler.HealthHandler(db))

	}

	return r
}
