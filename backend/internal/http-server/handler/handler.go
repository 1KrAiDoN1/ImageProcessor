package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"imageprocessor/backend/internal/domain/entity"
	"imageprocessor/backend/internal/http-server/handler/dto"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	maxFileSize = 32 << 20 // 32 MB
)

type Handler struct {
	logger            *zap.Logger
	imageService      ImageServiceInterface
	statisticsService StatisticsServiceInterface
}

func NewHandler(log *zap.Logger, imageService ImageServiceInterface, statisticsService StatisticsServiceInterface) *Handler {
	return &Handler{
		logger:            log,
		imageService:      imageService,
		statisticsService: statisticsService,
	}
}

// UploadImage обрабатывает загрузку изображения с операциями
func (h *Handler) UploadImage(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	// Ограничиваем размер загружаемого файла
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	// Парсим multipart форму (максимум 32MB)
	err := c.Request.ParseMultipartForm(maxFileSize)
	if err != nil {
		h.logger.Error("Failed to parse multipart form", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse form: " + err.Error(),
		})
		return
	}

	// Получаем файл из формы
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		h.logger.Error("Failed to get file from form", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to get image file: " + err.Error(),
		})
		return
	}
	h.logger.Info("Content-Type of file", zap.String("content_type", c.Request.Header.Get("Content-Type")))
	defer func() {
		err = file.Close()
		if err != nil {
			h.logger.Error("Failed to close file", zap.Error(err))
		}
	}()

	// Читаем содержимое файла
	imageData, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("Failed to read file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "read_error",
			Message: "Failed to read file: " + err.Error(),
		})
		return
	}

	// Получаем список операций из формы
	operationsJSON := c.PostForm("operations")
	var operations []dto.OperationRequest

	if operationsJSON != "" {
		err = json.Unmarshal([]byte(operationsJSON), &operations)
		if err != nil {
			h.logger.Error("Failed to parse operations", zap.Error(err))
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_operations",
				Message: "Failed to parse operations: " + err.Error(),
			})
			return
		}
	} else {
		// Если операции не указаны, используем дефолтные
		operations = []dto.OperationRequest{
			{
				Type: string(entity.OpThumbnail),
				Parameters: map[string]interface{}{
					"size": 200,
				},
			},
		}
	}

	// Валидируем операции
	entityOperations := make([]entity.OperationParams, 0, len(operations))
	for i, op := range operations {
		if err := op.Validate(); err != nil {
			h.logger.Error("Invalid operation", zap.Error(err), zap.Int("index", i))
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_operation",
				Message: fmt.Sprintf("Invalid operation at index %d: %s", i, err.Error()),
			})
			return
		}
		entityOperations = append(entityOperations, op.ToEntity())
	}

	// Получаем MIME тип
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Вызываем сервис для загрузки изображения
	image, err := h.imageService.UploadImage(ctx, imageData, header.Filename, mimeType, entityOperations)
	if err != nil {
		h.logger.Error("Failed to upload image", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "upload_failed",
			Message: "Failed to upload image: " + err.Error(),
		})
		return
	}

	// Записываем статистику загрузки
	err = h.statisticsService.RecordImageUploaded(ctx, image.OriginalSize)
	if err != nil {
		h.logger.Warn("Failed to record image upload statistics", zap.Error(err))
		// Не возвращаем ошибку, так как изображение уже загружено
	}

	h.logger.Info("Image uploaded successfully", zap.String("imageId", image.ID))

	// Формируем ответ
	response := dto.UploadResponse{
		ID:              image.ID,
		Status:          string(image.Status),
		Filename:        header.Filename,
		Size:            image.OriginalSize,
		MimeType:        image.MimeType,
		CreatedAt:       image.CreatedAt,
		OperationsCount: len(entityOperations),
		EstimatedTime:   len(entityOperations) * 2, // Примерная оценка в секундах
	}

	c.JSON(http.StatusCreated, response)
}

// GetImage возвращает изображение по ID и операции
func (h *Handler) GetImage(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "missing_id",
			Message: "Image ID is required",
		})
		return
	}

	// Получаем тип операции из query параметров
	operationStr := c.DefaultQuery("operation", "original")
	operation := entity.OperationType(operationStr)

	h.logger.Debug("Get image request",
		zap.String("imageId", imageID),
		zap.String("operation", operationStr),
	)

	// Получаем изображение из сервиса
	imageData, mimeType, err := h.imageService.GetImage(ctx, imageID, operation)
	if err != nil {
		h.logger.Error("Failed to get image", zap.Error(err), zap.String("imageId", imageID))
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Image not found: " + err.Error(),
		})
		return
	}

	// Устанавливаем заголовки
	c.Header("Content-Type", mimeType)
	c.Header("Content-Length", strconv.Itoa(len(imageData)))
	c.Header("Cache-Control", "public, max-age=31536000")
	c.Header("ETag", fmt.Sprintf("%s-%s", imageID, operationStr))

	// Возвращаем бинарные данные
	c.Data(http.StatusOK, mimeType, imageData)
}

// GetImageStatus возвращает статус обработки изображения
func (h *Handler) GetImageStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "missing_id",
			Message: "Image ID is required",
		})
		return
	}

	h.logger.Debug("Get image status request", zap.String("imageId", imageID))

	// Получаем статус из сервиса
	status, err := h.imageService.GetImageStatus(ctx, imageID)
	if err != nil {
		h.logger.Error("Failed to get image status", zap.Error(err), zap.String("imageId", imageID))
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Image not found: " + err.Error(),
		})
		return
	}

	// Формируем ответ
	response := dto.ImageStatusResponse{
		ID:                  status.ID,
		Status:              string(status.Status),
		Progress:            status.Progress,
		ProcessedOperations: status.ProcessedOperations,
		TotalOperations:     status.TotalOperations,
		CreatedAt:           status.CreatedAt,
		UpdatedAt:           status.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteImage удаляет изображение и все его версии
func (h *Handler) DeleteImage(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "missing_id",
			Message: "Image ID is required",
		})
		return
	}

	h.logger.Info("Delete image request", zap.String("imageId", imageID))

	// Удаляем изображение через сервис
	err := h.imageService.DeleteImage(ctx, imageID)
	if err != nil {
		h.logger.Error("Failed to delete image", zap.Error(err), zap.String("imageId", imageID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete image: " + err.Error(),
		})
		return
	}

	h.logger.Info("Image deleted successfully", zap.String("imageId", imageID))

	c.JSON(http.StatusOK, dto.DeleteImageResponse{
		Success: true,
		Message: "Image deleted successfully",
		ID:      imageID,
	})
}

// GetStatistics возвращает общую статистику системы
func (h *Handler) GetStatistics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	h.logger.Debug("Get statistics request")

	// Получаем статистику из сервиса
	stats, err := h.statisticsService.GetDetailedStatistics(ctx)
	if err != nil {
		h.logger.Error("Failed to get statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "stats_failed",
			Message: "Failed to get statistics: " + err.Error(),
		})
		return
	}

	// Конвертируем статистику по операциям
	opStats := make([]dto.OperationStatistic, 0, len(stats.OperationStatistics))
	for _, stat := range stats.OperationStatistics {
		opStats = append(opStats, dto.OperationStatistic{
			OperationType:           stat.OperationType,
			TotalCount:              stat.TotalCount,
			SuccessCount:            stat.SuccessCount,
			FailureCount:            stat.FailureCount,
			AverageProcessingTimeMs: stat.AverageProcessingTimeMs,
		})
	}

	// Формируем ответ
	response := dto.StatisticsResponse{
		TotalImagesUploaded:     stats.GeneralStatistics.TotalImagesUploaded,
		TotalImagesProcessed:    stats.GeneralStatistics.TotalImagesProcessed,
		TotalImagesFailed:       stats.GeneralStatistics.FailedProcessingAttempts,
		TotalDataProcessedBytes: stats.GeneralStatistics.TotalDataProcessedBytes,
		TotalDataProcessedMB:    float64(stats.GeneralStatistics.TotalDataProcessedBytes) / (1024 * 1024),
		AverageProcessingTimeMs: stats.GeneralStatistics.AverageProcessingTimeMs,
		OperationStatistics:     opStats,
		LastUpdated:             stats.GeneralStatistics.LastStatisticsUpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetImagePresignedURL генерирует временную ссылку для скачивания изображения
func (h *Handler) GetImagePresignedURL(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "missing_id",
			Message: "Image ID is required",
		})
		return
	}

	operationStr := c.DefaultQuery("operation", "original")
	operation := entity.OperationType(operationStr)

	expiryStr := c.DefaultQuery("expiry", "3600")
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil || expiry <= 0 {
		expiry = 3600
	}

	h.logger.Debug("Get presigned URL request",
		zap.String("imageId", imageID),
		zap.String("operation", operationStr),
		zap.Int("expiry", expiry),
	)

	// Генерируем URL
	url, err := h.imageService.GetImagePresignedURL(
		ctx,
		imageID,
		operation,
		time.Duration(expiry)*time.Second,
	)
	if err != nil {
		h.logger.Error("Failed to generate presigned URL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "url_generation_failed",
			Message: "Failed to generate presigned URL: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":        url,
		"expires_in": expiry,
	})
}

// ListImages возвращает список изображений с пагинацией
func (h *Handler) ListImages(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	h.logger.Debug("List images request",
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	images, err := h.imageService.ListImages(ctx, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list images", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "list_failed",
			Message: "Failed to list images: " + err.Error(),
		})
		return
	}

	// Формируем ответ
	response := make([]dto.ImageResponse, 0, len(images))
	for _, img := range images {
		response = append(response, dto.ImageResponse{
			ID:        img.ID,
			Filename:  img.OriginalFilename,
			Status:    string(img.Status),
			Size:      img.OriginalSize,
			MimeType:  img.MimeType,
			CreatedAt: img.CreatedAt,
			UpdatedAt: img.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"images": response,
		"limit":  limit,
		"offset": offset,
		"count":  len(response),
	})
}

// HealthCheck проверяет здоровье сервиса
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services: map[string]string{
			"api":      "ok",
			"database": "ok",
			"storage":  "ok",
		},
	})
}
