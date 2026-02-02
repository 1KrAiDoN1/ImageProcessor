package stats_service

import (
	"context"
	"fmt"
	"imageprocessor/backend/internal/domain/entity"
	"sync"

	"go.uber.org/zap"
)

type StatsService struct {
	statsRepo StatsRepositoryInterface
	logger    *zap.Logger
	mu        sync.RWMutex
}

func NewStatsService(statsRepo StatsRepositoryInterface, logger *zap.Logger) *StatsService {
	return &StatsService{
		statsRepo: statsRepo,
		logger:    logger,
	}
}

// GetStatistics возвращает общую статистику системы
func (s *StatsService) GetStatistics(ctx context.Context) (*entity.ProcessingStatistics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.logger.Debug("Getting statistics")

	stats, err := s.statsRepo.GetStatistics(ctx)
	if err != nil {
		s.logger.Error("Failed to get statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// Получаем наиболее используемую операцию
	mostUsed, err := s.statsRepo.GetMostUsedOperation(ctx)
	if err != nil {
		s.logger.Warn("Failed to get most used operation", zap.Error(err))
		mostUsed = "unknown"
	}
	stats.MostUsedOperationType = mostUsed

	return stats, nil
}

// GetOperationStatistics возвращает статистику по операциям
func (s *StatsService) GetOperationStatistics(ctx context.Context) ([]entity.OperationStat, error) {
	s.logger.Debug("Getting operation statistics")

	opStats, err := s.statsRepo.GetOperationStatistics(ctx)
	if err != nil {
		s.logger.Error("Failed to get operation statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get operation statistics: %w", err)
	}

	// Конвертируем в DTO
	result := make([]entity.OperationStat, 0, len(opStats))
	for _, stat := range opStats {
		result = append(result, entity.OperationStat{
			OperationType:           stat.OperationType,
			TotalCount:              stat.TotalCount,
			SuccessCount:            stat.SuccessCount,
			FailureCount:            stat.FailureCount,
			AverageProcessingTimeMs: stat.AverageProcessingTimeMs,
		})
	}

	return result, nil
}

// RecordImageUploaded записывает факт загрузки изображения
func (s *StatsService) RecordImageUploaded(ctx context.Context, size int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug("Recording image uploaded", zap.Int64("size", size))

	err := s.statsRepo.IncrementImageUploaded(ctx, size)
	if err != nil {
		s.logger.Error("Failed to record image uploaded", zap.Error(err))
		return fmt.Errorf("failed to record image uploaded: %w", err)
	}

	return nil
}

// RecordImageProcessed записывает факт успешной обработки изображения
func (s *StatsService) RecordImageProcessed(ctx context.Context, operation entity.OperationType, processingTimeMs float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug("Recording image processed",
		zap.String("operation", string(operation)),
		zap.Float64("processingTimeMs", processingTimeMs),
	)

	// Обновляем общую статистику
	err := s.statsRepo.IncrementImageProcessed(ctx, processingTimeMs)
	if err != nil {
		s.logger.Error("Failed to record image processed", zap.Error(err))
		return fmt.Errorf("failed to record image processed: %w", err)
	}

	// Обновляем статистику по операции
	err = s.statsRepo.UpdateOperationStatistics(ctx, operation, true, processingTimeMs)
	if err != nil {
		s.logger.Error("Failed to update operation statistics",
			zap.Error(err),
			zap.String("operation", string(operation)),
		)
		return fmt.Errorf("failed to update operation statistics: %w", err)
	}

	return nil
}

// RecordImageFailed записывает факт неудачной обработки изображения
func (s *StatsService) RecordImageFailed(ctx context.Context, operation entity.OperationType, processingTimeMs float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug("Recording image processing failed",
		zap.String("operation", string(operation)),
		zap.Float64("processingTimeMs", processingTimeMs),
	)

	// Обновляем общую статистику
	err := s.statsRepo.IncrementImageFailed(ctx)
	if err != nil {
		s.logger.Error("Failed to record image failed", zap.Error(err))
		return fmt.Errorf("failed to record image failed: %w", err)
	}

	// Обновляем статистику по операции
	err = s.statsRepo.UpdateOperationStatistics(ctx, operation, false, processingTimeMs)
	if err != nil {
		s.logger.Error("Failed to update operation statistics",
			zap.Error(err),
			zap.String("operation", string(operation)),
		)
		return fmt.Errorf("failed to update operation statistics: %w", err)
	}

	return nil
}

// RecordOperationsProcessed записывает факт обработки нескольких операций
func (s *StatsService) RecordOperationsProcessed(ctx context.Context, operations []entity.OperationType, processingTimes map[entity.OperationType]float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug("Recording operations processed", zap.Int("count", len(operations)))

	for _, op := range operations {
		processingTime, exists := processingTimes[op]
		if !exists {
			processingTime = 0
		}

		err := s.statsRepo.UpdateOperationStatistics(ctx, op, true, processingTime)
		if err != nil {
			s.logger.Error("Failed to update operation statistics",
				zap.Error(err),
				zap.String("operation", string(op)),
			)
			continue
		}
	}

	// Вычисляем среднее время обработки
	var totalTime float64
	for _, time := range processingTimes {
		totalTime += time
	}
	avgTime := totalTime / float64(len(processingTimes))

	err := s.statsRepo.IncrementImageProcessed(ctx, avgTime)
	if err != nil {
		s.logger.Error("Failed to increment image processed", zap.Error(err))
		return fmt.Errorf("failed to increment image processed: %w", err)
	}

	return nil
}

// GetDetailedStatistics возвращает детальную статистику с разбивкой по операциям
func (s *StatsService) GetDetailedStatistics(ctx context.Context) (*entity.DetailedStatistics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Получаем общую статистику
	generalStats, err := s.GetStatistics(ctx)
	if err != nil {
		return nil, err
	}

	// Получаем статистику по операциям
	opStats, err := s.GetOperationStatistics(ctx)
	if err != nil {
		return nil, err
	}

	detailed := &entity.DetailedStatistics{
		GeneralStatistics:   generalStats,
		OperationStatistics: opStats,
	}

	// Вычисляем дополнительные метрики
	if generalStats.TotalImagesProcessed > 0 {
		detailed.SuccessRate = float64(generalStats.TotalImagesProcessed-generalStats.FailedProcessingAttempts) / float64(generalStats.TotalImagesProcessed) * 100
	}

	return detailed, nil
}
