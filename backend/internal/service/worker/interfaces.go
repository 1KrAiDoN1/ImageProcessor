package workerservice

import (
	"context"
	"imageprocessor/backend/internal/domain/entity"
)

// ImageRepository определяет интерфейс репозитория для worker
type ImageRepositoryInterface interface {
	GetImageByID(ctx context.Context, imageID string) (*entity.Image, error)
	UpdateImageStatus(ctx context.Context, imageID string, status entity.ImageStatus) error
	CreateProcessedImage(ctx context.Context, processed *entity.ProcessedImage) error
	UpdateProcessingJobStatus(ctx context.Context, jobID string, status string, errorMsg string) error
}

// StatsService определяет интерфейс сервиса статистики
type StatsServiceInterface interface {
	RecordImageProcessed(ctx context.Context, operation entity.OperationType, processingTimeMs float64) error
	RecordImageFailed(ctx context.Context, operation entity.OperationType, processingTimeMs float64) error
	RecordOperationsProcessed(ctx context.Context, operations []entity.OperationType, processingTimes map[entity.OperationType]float64) error
}

// ImageProcessor определяет интерфейс для обработки изображений
type ImageProcessorInterface interface {
	// ProcessImage обрабатывает изображение согласно списку операций
	ProcessImage(ctx context.Context, imageData []byte, operations []entity.OperationParams) (map[string][]byte, error)

	// ValidateImage проверяет, является ли файл допустимым изображением
	ValidateImage(imageData []byte) (entity.ImageFormat, error)

	// GetImageInfo возвращает информацию об изображении (размер, формат)
	GetImageInfo(imageData []byte) (*entity.ImageInfo, error)
}
