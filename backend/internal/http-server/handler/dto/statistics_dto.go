package dto

import "time"

// ImageStatisticsResponse представляет статистику по одному изображению
type ImageStatisticsResponse struct {
	ImageID             string    `json:"image_id"`
	TotalProcessingTime int64     `json:"total_processing_time_ms"`
	TimesAccessed       int64     `json:"times_accessed"`
	TimesProcessed      int64     `json:"times_processed"`
	LastAccessedAt      time.Time `json:"last_accessed_at,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// StatisticsResponse представляет общую статистику системы
type StatisticsResponse struct {
	TotalImagesProcessed     int64     `json:"total_images_processed"`
	TotalImagesUploaded      int64     `json:"total_images_uploaded"`
	AverageProcessingTimeMs  float64   `json:"average_processing_time_ms"`
	TotalDataProcessedBytes  int64     `json:"total_data_processed_bytes"`
	FailedProcessingAttempts int64     `json:"failed_processing_attempts"`
	SuccessRate              float64   `json:"success_rate"`
	MostUsedOperationType    string    `json:"most_used_operation_type"`
	LastStatisticsUpdatedAt  time.Time `json:"last_statistics_updated_at"`
}

// StatisticsFilterRequest представляет фильтры для получения статистики
type StatisticsFilterRequest struct {
	ImageID   string    `json:"image_id,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
}
