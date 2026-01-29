package entity

import (
	"time"
)

// ProcessingStatistics содержит статистику по обработке
type ProcessingStatistics struct {
	ID                       string
	TotalImagesProcessed     int64
	TotalImagesUploaded      int64
	AverageProcessingTimeMs  float64
	TotalDataProcessedBytes  int64
	FailedProcessingAttempts int64
	MostUsedOperationType    string
	LastStatisticsUpdatedAt  time.Time
}
