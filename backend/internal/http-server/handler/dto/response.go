package dto

import (
	"imageprocessor/backend/internal/domain/entity"
	"time"
)

// ErrorResponse представляет ошибку API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// UploadResponse представляет ответ на загрузку изображения
type UploadResponse struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"`
	OriginalURL     string    `json:"original_url,omitempty"`
	Filename        string    `json:"filename"`
	Size            int64     `json:"size"`
	MimeType        string    `json:"mime_type"`
	CreatedAt       time.Time `json:"created_at"`
	EstimatedTime   int       `json:"estimated_time_seconds,omitempty"`
	OperationsCount int       `json:"operations_count"`
}

// ImageStatusResponse представляет статус обработки изображения
type ImageStatusResponse struct {
	ID                  string               `json:"id"`
	Status              string               `json:"status"`
	Progress            int                  `json:"progress"`
	ProcessedOperations int                  `json:"processed_operations"`
	TotalOperations     int                  `json:"total_operations"`
	Results             []ProcessedImageInfo `json:"results,omitempty"`
	ErrorMessage        string               `json:"error_message,omitempty"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
	CompletedAt         *time.Time           `json:"completed_at,omitempty"`
}

// ProcessedImageInfo представляет информацию об обработанном изображении
type ProcessedImageInfo struct {
	Operation string `json:"operation"`
	URL       string `json:"url,omitempty"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Status    string `json:"status"`
}

// ImageResponse представляет информацию об изображении
type ImageResponse struct {
	ID          string               `json:"id"`
	Filename    string               `json:"filename"`
	OriginalURL string               `json:"original_url"`
	Status      string               `json:"status"`
	Size        int64                `json:"size"`
	MimeType    string               `json:"mime_type"`
	Versions    []ProcessedImageInfo `json:"versions,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// GetImageResponse представляет ответ на получение изображения
type GetImageResponse struct {
	URL         string `json:"url,omitempty"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Data        []byte `json:"-"`
}

// DeleteImageResponse представляет ответ на удаление изображения
type DeleteImageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id"`
}

// StatisticsResponse представляет общую статистику
type StatisticsResponse struct {
	TotalImagesUploaded     int64                `json:"total_images_uploaded"`
	TotalImagesProcessed    int64                `json:"total_images_processed"`
	TotalImagesFailed       int64                `json:"total_images_failed"`
	TotalDataProcessedBytes int64                `json:"total_data_processed_bytes"`
	TotalDataProcessedMB    float64              `json:"total_data_processed_mb"`
	AverageProcessingTimeMs float64              `json:"average_processing_time_ms"`
	OperationStatistics     []OperationStatistic `json:"operation_statistics,omitempty"`
	LastUpdated             time.Time            `json:"last_updated"`
}

// OperationStatistic представляет статистику по операции
type OperationStatistic struct {
	OperationType           string  `json:"operation_type"`
	TotalCount              int64   `json:"total_count"`
	SuccessCount            int64   `json:"success_count"`
	FailureCount            int64   `json:"failure_count"`
	AverageProcessingTimeMs float64 `json:"average_processing_time_ms"`
}

// HealthResponse представляет статус здоровья сервиса
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// Вспомогательные функции для конвертации entity в DTO

// FromImageEntity конвертирует entity.Image в ImageResponse
func FromImageEntity(img *entity.Image, versions []entity.ProcessedImage) *ImageResponse {
	resp := &ImageResponse{
		ID:        img.ID,
		Filename:  img.OriginalFilename,
		Status:    string(img.Status),
		Size:      img.OriginalSize,
		MimeType:  img.MimeType,
		CreatedAt: img.CreatedAt,
		UpdatedAt: img.UpdatedAt,
	}

	// Добавляем информацию о версиях
	if len(versions) > 0 {
		resp.Versions = make([]ProcessedImageInfo, 0, len(versions))
		for _, v := range versions {
			resp.Versions = append(resp.Versions, ProcessedImageInfo{
				Operation: string(v.Operation),
				Path:      v.Path,
				Size:      v.Size,
				Status:    v.Status,
			})
		}
	}

	return resp
}

// FromProcessedImages конвертирует слайс ProcessedImage в слайс ProcessedImageInfo
func FromProcessedImages(images []entity.ProcessedImage) []ProcessedImageInfo {
	result := make([]ProcessedImageInfo, 0, len(images))
	for _, img := range images {
		result = append(result, ProcessedImageInfo{
			Operation: string(img.Operation),
			Path:      img.Path,
			Size:      img.Size,
			Status:    img.Status,
		})
	}
	return result
}

// FromStatisticsEntity конвертирует entity.ProcessingStatistics в StatisticsResponse
func FromStatisticsEntity(stats *entity.ProcessingStatistics, opStats []OperationStatistic) *StatisticsResponse {
	return &StatisticsResponse{
		TotalImagesUploaded:     stats.TotalImagesUploaded,
		TotalImagesProcessed:    stats.TotalImagesProcessed,
		TotalImagesFailed:       stats.FailedProcessingAttempts,
		TotalDataProcessedBytes: stats.TotalDataProcessedBytes,
		TotalDataProcessedMB:    float64(stats.TotalDataProcessedBytes) / (1024 * 1024),
		AverageProcessingTimeMs: stats.AverageProcessingTimeMs,
		OperationStatistics:     opStats,
		LastUpdated:             stats.LastStatisticsUpdatedAt,
	}
}
