package routes

import (
	"imageprocessor/backend/internal/http-server/handler"
	"imageprocessor/backend/internal/http-server/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRoutes настраивает все маршруты API
func SetupRoutes(router *gin.RouterGroup, h *handler.Handler, logger *zap.Logger) {
	// Middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", h.HealthCheck)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Images endpoints
		images := v1.Group("/images")
		{
			images.POST("", h.UploadImage)                 // Загрузка изображения с операциями
			images.GET("", h.ListImages)                   // Список изображений
			images.GET("/:id", h.GetImage)                 // Получение изображения
			images.GET("/:id/status", h.GetImageStatus)    // Статус обработки
			images.GET("/:id/url", h.GetImagePresignedURL) // Генерация presigned URL
			images.DELETE("/:id", h.DeleteImage)           // Удаление изображения
		}

		// Statistics endpoints
		statistics := v1.Group("/statistics")
		{
			statistics.GET("", h.GetStatistics) // Общая статистика
		}
	}

	// Версия API
	router.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "1.0.0",
			"service": "image-processor-api",
		})
	})
}
