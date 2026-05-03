package routes

import (
	"github.com/devrapture/omni/handler"
	"github.com/devrapture/omni/utils"
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
