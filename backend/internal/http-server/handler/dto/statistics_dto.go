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

// StatisticsFilterRequest представляет фильтры для получения статистики
type StatisticsFilterRequest struct {
	ImageID   string    `json:"image_id,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
}
