package workerservice

import (
	"bytes"
	"context"
	"fmt"
	"imageprocessor/backend/internal/domain/entity"
	"imageprocessor/backend/internal/repository/cloud"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WorkerService struct {
	processor    ImageProcessorInterface
	cloudStorage cloud.CloudStorageInterface
	imageRepo    ImageRepositoryInterface
	statsService StatsServiceInterface
	logger       *zap.Logger
	bucket       string
}

func NewWorkerService(
	processor ImageProcessorInterface,
	cloudStorage cloud.CloudStorageInterface,
	imageRepo ImageRepositoryInterface,
	statsService StatsServiceInterface,
	logger *zap.Logger,
	bucket string,
) *WorkerService {
	return &WorkerService{
		processor:    processor,
		cloudStorage: cloudStorage,
		imageRepo:    imageRepo,
		statsService: statsService,
		logger:       logger,
		bucket:       bucket,
	}
}

// ProcessTask обрабатывает задачу из Kafka
func (w *WorkerService) ProcessTask(ctx context.Context, task *entity.ProcessingTask) error {
	w.logger.Info("Processing task",
		zap.String("taskId", task.ID),
		zap.String("imageId", task.ImageID),
		zap.Int("operationsCount", len(task.Operations)),
	)

	startTime := time.Now()

	// Скачиваем оригинальное изображение из S3
	imageData, err := w.cloudStorage.DownloadFile(ctx, task.OriginalPath)
	if err != nil {
		w.logger.Error("Failed to download original image",
			zap.Error(err),
			zap.String("taskId", task.ID),
			zap.String("path", task.OriginalPath),
		)
		err = w.updateJobStatus(ctx, task.ID, "failed", err.Error())
		if err != nil {
			w.logger.Error("Failed to update job status", zap.Error(err))
		}
		err = w.updateImageStatus(ctx, task.ImageID, entity.StatusFailed)
		if err != nil {
			w.logger.Error("Failed to update image status", zap.Error(err))
		}
		for _, op := range task.Operations {
			_ = w.statsService.RecordImageFailed(ctx, op.Type, 0)
		}
		return fmt.Errorf("failed to download original image: %w", err)
	}

	w.logger.Debug("Original image downloaded",
		zap.String("taskId", task.ID),
		zap.Int("size", len(imageData)),
	)

	// Обрабатываем изображение
	processedImages, err := w.processor.ProcessImage(ctx, imageData, task.Operations)
	if err != nil {
		w.logger.Error("Failed to process image",
			zap.Error(err),
			zap.String("taskId", task.ID),
		)
		err = w.updateJobStatus(ctx, task.ID, "failed", err.Error())
		if err != nil {
			w.logger.Error("Failed to update job status", zap.Error(err))
		}
		err = w.updateImageStatus(ctx, task.ImageID, entity.StatusFailed)
		if err != nil {
			w.logger.Error("Failed to update image status", zap.Error(err))
		}
		// Записываем в статистику неудачную обработку
		for _, op := range task.Operations {
			_ = w.statsService.RecordImageFailed(ctx, op.Type, 0)
		}

		return fmt.Errorf("failed to process image: %w", err)
	}

	w.logger.Info("Image processed successfully",
		zap.String("taskId", task.ID),
		zap.Int("resultCount", len(processedImages)),
	)

	// Сохраняем обработанные изображения в S3 и БД
	processingTimes := make(map[entity.OperationType]float64)

	for operationType, processedData := range processedImages {
		opStartTime := time.Now()

		// Формируем путь для обработанного изображения
		processedPath := fmt.Sprintf("processed/%s/%s/%s.jpg",
			task.ImageID,
			operationType,
			uuid.New().String(),
		)

		// Загружаем в S3
		err = w.cloudStorage.UploadFile(ctx, processedPath,
			bytes.NewReader(processedData),
			int64(len(processedData)),
			"image/jpeg")
		if err != nil {
			w.logger.Error("Failed to upload processed image",
				zap.Error(err),
				zap.String("operation", operationType),
			)
			continue
		}

		w.logger.Debug("Processed image uploaded",
			zap.String("taskId", task.ID),
			zap.String("operation", operationType),
			zap.String("path", processedPath),
		)

		// Создаем запись в БД
		processedImage := &entity.ProcessedImage{
			ID:         uuid.New().String(),
			ImageID:    task.ImageID,
			Operation:  entity.OperationType(operationType),
			Parameters: getOperationParams(task.Operations, entity.OperationType(operationType)),
			Path:       processedPath,
			Size:       int64(len(processedData)),
			MimeType:   "image/jpeg",
			Format:     task.Format,
			Status:     "completed",
			CreatedAt:  time.Now(),
		}

		err = w.imageRepo.CreateProcessedImage(ctx, processedImage)
		if err != nil {
			w.logger.Error("Failed to create processed image record",
				zap.Error(err),
				zap.String("operation", operationType),
			)
			continue
		}

		// Записываем время обработки
		opDuration := time.Since(opStartTime)
		processingTimes[entity.OperationType(operationType)] = float64(opDuration.Milliseconds())

		w.logger.Info("Processed image saved",
			zap.String("operation", operationType),
			zap.Duration("processingTime", opDuration),
		)
	}

	// Обновляем статус задачи
	err = w.updateJobStatus(ctx, task.ID, "completed", "")
	if err != nil {
		w.logger.Error("Failed to update job status", zap.Error(err))
	}

	// Обновляем статус изображения
	err = w.updateImageStatus(ctx, task.ImageID, entity.StatusCompleted)
	if err != nil {
		w.logger.Error("Failed to update image status", zap.Error(err))
	}

	// Записываем статистику
	totalDuration := time.Since(startTime)

	// Записываем статистику по каждой операции
	for opType, procTime := range processingTimes {
		err = w.statsService.RecordImageProcessed(ctx, opType, procTime)
		if err != nil {
			w.logger.Error("Failed to record stats",
				zap.Error(err),
				zap.String("operation", string(opType)),
			)
		}
	}

	w.logger.Info("Task completed successfully",
		zap.String("taskId", task.ID),
		zap.String("imageId", task.ImageID),
		zap.Duration("totalDuration", totalDuration),
		zap.Int("operationsProcessed", len(processedImages)),
	)

	return nil
}

// ProcessTaskWithRetry обрабатывает задачу с повторными попытками
func (w *WorkerService) ProcessTaskWithRetry(ctx context.Context, task *entity.ProcessingTask, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		w.logger.Info("Processing task attempt",
			zap.String("taskId", task.ID),
			zap.Int("attempt", attempt),
			zap.Int("maxRetries", maxRetries),
		)

		err := w.ProcessTask(ctx, task)
		if err == nil {
			return nil
		}

		lastErr = err
		w.logger.Warn("Task processing failed, will retry",
			zap.Error(err),
			zap.String("taskId", task.ID),
			zap.Int("attempt", attempt),
		)

		// Экспоненциальная задержка перед повтором
		if attempt < maxRetries {
			backoff := time.Duration(attempt*2) * time.Second
			w.logger.Info("Waiting before retry",
				zap.Duration("backoff", backoff),
			)
			time.Sleep(backoff)
		}
	}

	// Все попытки исчерпаны
	w.logger.Error("Task processing failed after all retries",
		zap.Error(lastErr),
		zap.String("taskId", task.ID),
		zap.Int("maxRetries", maxRetries),
	)

	return fmt.Errorf("task processing failed after %d retries: %w", maxRetries, lastErr)
}

// updateJobStatus обновляет статус задачи в БД
func (w *WorkerService) updateJobStatus(ctx context.Context, jobID, status, errorMsg string) error {
	return w.imageRepo.UpdateProcessingJobStatus(ctx, jobID, status, errorMsg)
}

// updateImageStatus обновляет статус изображения в БД
func (w *WorkerService) updateImageStatus(ctx context.Context, imageID string, status entity.ImageStatus) error {
	return w.imageRepo.UpdateImageStatus(ctx, imageID, status)
}

// getOperationParams извлекает параметры для конкретной операции
func getOperationParams(operations []entity.OperationParams, opType entity.OperationType) string {
	for _, op := range operations {
		if op.Type == opType {
			// Конвертируем параметры в строку (можно использовать JSON)
			return fmt.Sprintf("%v", op.Parameters)
		}
	}
	return ""
}
