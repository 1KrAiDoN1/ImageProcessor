package routes

import (
	"imageprocessor/backend/internal/http-server/handler"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRoutes(
	router *gin.RouterGroup,
	logger *zap.Logger,
	handlers *handler.Handler,
) {

	image := router.Group("/image")
	{
		image.GET("/:id", handlers.GetImage)
		image.POST("/upload", handlers.UploadImage)
		image.DELETE("/:id", handlers.DeleteImage)
	}

	statistics := router.Group("/statistics")
	{
		statistics.GET("/", handlers.GetStatistics)
	}

}
