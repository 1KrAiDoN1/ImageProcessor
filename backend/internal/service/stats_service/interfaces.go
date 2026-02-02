package stats_service

import (
	"context"
	"imageprocessor/backend/internal/domain/entity"
)

// StatsRepository определяет интерфейс репозитория статистики
type StatsRepositoryInterface interface {
	GetStatistics(ctx context.Context) (*entity.ProcessingStatistics, error)
	IncrementImageUploaded(ctx context.Context, size int64) error
	IncrementImageProcessed(ctx context.Context, processingTimeMs float64) error
	IncrementImageFailed(ctx context.Context) error
	GetOperationStatistics(ctx context.Context) ([]entity.OperationStat, error)
	UpdateOperationStatistics(ctx context.Context, operation entity.OperationType, success bool, processingTimeMs float64) error
	GetMostUsedOperation(ctx context.Context) (string, error)
}
