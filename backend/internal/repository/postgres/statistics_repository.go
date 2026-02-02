package postgres

import (
	"context"
	"fmt"
	"imageprocessor/backend/internal/domain/entity"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type StatisticsRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewStatisticsRepository(db *pgxpool.Pool, logger *zap.Logger) *StatisticsRepository {
	return &StatisticsRepository{
		db:     db,
		logger: logger,
	}
}

// GetStatistics получает общую статистику
func (r *StatisticsRepository) GetStatistics(ctx context.Context) (*entity.ProcessingStatistics, error) {
	query := `
		SELECT id, total_images_uploaded, total_images_processed, total_images_failed,
		       total_data_processed_bytes, average_processing_time_ms, updated_at
		FROM statistics
		WHERE id = 'default'
	`

	var stats entity.ProcessingStatistics
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.ID,
		&stats.TotalImagesUploaded,
		&stats.TotalImagesProcessed,
		&stats.TotalDataProcessedBytes,
		&stats.FailedProcessingAttempts,
		&stats.TotalDataProcessedBytes,
		&stats.AverageProcessingTimeMs,
		&stats.LastStatisticsUpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to get statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return &stats, nil
}

// IncrementImageUploaded увеличивает счетчик загруженных изображений
func (r *StatisticsRepository) IncrementImageUploaded(ctx context.Context, size int64) error {
	query := `
		UPDATE statistics
		SET total_images_uploaded = total_images_uploaded + 1,
		    total_data_processed_bytes = total_data_processed_bytes + $1
		WHERE id = 'default'
	`

	_, err := r.db.Exec(ctx, query, size)
	if err != nil {
		r.logger.Error("Failed to increment image uploaded", zap.Error(err))
		return fmt.Errorf("failed to increment image uploaded: %w", err)
	}

	return nil
}

// IncrementImageProcessed увеличивает счетчик обработанных изображений
func (r *StatisticsRepository) IncrementImageProcessed(ctx context.Context, processingTimeMs float64) error {
	query := `
		UPDATE statistics
		SET total_images_processed = total_images_processed + 1,
		    average_processing_time_ms = (average_processing_time_ms * total_images_processed + $1) / (total_images_processed + 1)
		WHERE id = 'default'
	`

	_, err := r.db.Exec(ctx, query, processingTimeMs)
	if err != nil {
		r.logger.Error("Failed to increment image processed", zap.Error(err))
		return fmt.Errorf("failed to increment image processed: %w", err)
	}

	return nil
}

// IncrementImageFailed увеличивает счетчик неудачных обработок
func (r *StatisticsRepository) IncrementImageFailed(ctx context.Context) error {
	query := `
		UPDATE statistics
		SET total_images_failed = total_images_failed + 1
		WHERE id = 'default'
	`

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		r.logger.Error("Failed to increment image failed", zap.Error(err))
		return fmt.Errorf("failed to increment image failed: %w", err)
	}

	return nil
}

// GetOperationStatistics получает статистику по операциям
func (r *StatisticsRepository) GetOperationStatistics(ctx context.Context) ([]entity.OperationStat, error) {
	query := `
		SELECT operation_type, total_count, success_count, failure_count, average_processing_time_ms
		FROM operation_statistics
		ORDER BY total_count DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get operation statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get operation statistics: %w", err)
	}
	defer rows.Close()

	var stats []entity.OperationStat
	for rows.Next() {
		var stat entity.OperationStat
		err := rows.Scan(
			&stat.OperationType,
			&stat.TotalCount,
			&stat.SuccessCount,
			&stat.FailureCount,
			&stat.AverageProcessingTimeMs,
		)
		if err != nil {
			r.logger.Error("Failed to scan operation stat", zap.Error(err))
			return nil, fmt.Errorf("failed to scan operation stat: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// UpdateOperationStatistics обновляет статистику по операции
func (r *StatisticsRepository) UpdateOperationStatistics(ctx context.Context, operation entity.OperationType, success bool, processingTimeMs float64) error {
	query := `
		INSERT INTO operation_statistics (operation_type, total_count, success_count, failure_count, average_processing_time_ms, total_processing_time_ms)
		VALUES ($1, 1, $2, $3, $4, $5)
		ON CONFLICT (operation_type) DO UPDATE
		SET total_count = operation_statistics.total_count + 1,
		    success_count = operation_statistics.success_count + EXCLUDED.success_count,
		    failure_count = operation_statistics.failure_count + EXCLUDED.failure_count,
		    total_processing_time_ms = operation_statistics.total_processing_time_ms + $5,
		    average_processing_time_ms = (operation_statistics.total_processing_time_ms + $5) / (operation_statistics.total_count + 1)
	`

	successCount := 0
	failureCount := 0
	if success {
		successCount = 1
	} else {
		failureCount = 1
	}

	_, err := r.db.Exec(ctx, query, operation, successCount, failureCount, processingTimeMs, int64(processingTimeMs))
	if err != nil {
		r.logger.Error("Failed to update operation statistics",
			zap.Error(err),
			zap.String("operation", string(operation)),
		)
		return fmt.Errorf("failed to update operation statistics: %w", err)
	}

	return nil
}

// GetMostUsedOperation возвращает наиболее используемую операцию
func (r *StatisticsRepository) GetMostUsedOperation(ctx context.Context) (string, error) {
	query := `
		SELECT operation_type
		FROM operation_statistics
		ORDER BY total_count DESC
		LIMIT 1
	`

	var operationType string
	err := r.db.QueryRow(ctx, query).Scan(&operationType)
	if err != nil {
		r.logger.Error("Failed to get most used operation", zap.Error(err))
		return "", fmt.Errorf("failed to get most used operation: %w", err)
	}

	return operationType, nil
}
