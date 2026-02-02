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

// OperationStat представляет статистику по операции
type OperationStat struct {
	OperationType           string
	TotalCount              int64
	SuccessCount            int64
	FailureCount            int64
	AverageProcessingTimeMs float64
}

// DetailedStatistics представляет детальную статистику
type DetailedStatistics struct {
	GeneralStatistics   *ProcessingStatistics `json:"general_statistics"`
	OperationStatistics []OperationStat       `json:"operation_statistics"`
	SuccessRate         float64               `json:"success_rate"`
}
