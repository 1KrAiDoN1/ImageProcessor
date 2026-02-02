package imageservice

import (
	"context"
	"imageprocessor/backend/internal/domain/entity"
)

// ImageRepository определяет интерфейс репозитория изображений
type ImageRepositoryInterface interface {
	CreateImage(ctx context.Context, image *entity.Image) error
	GetImageByID(ctx context.Context, imageID string) (*entity.Image, error)
	UpdateImageStatus(ctx context.Context, imageID string, status entity.ImageStatus) error
	DeleteImage(ctx context.Context, imageID string) error
	ListImages(ctx context.Context, limit, offset int) ([]entity.Image, error)

	CreateProcessedImage(ctx context.Context, processed *entity.ProcessedImage) error
	GetProcessedImagesByImageID(ctx context.Context, imageID string) ([]entity.ProcessedImage, error)
	GetProcessedImageByOperation(ctx context.Context, imageID string, operation entity.OperationType) (*entity.ProcessedImage, error)

	CreateProcessingJob(ctx context.Context, job *entity.ProcessingTask) error
	UpdateProcessingJobStatus(ctx context.Context, jobID string, status string, errorMsg string) error
	GetProcessingJobByImageID(ctx context.Context, imageID string) (*entity.ProcessingTask, error)
}

// KafkaProducer определяет интерфейс для отправки сообщений в Kafka
type KafkaProducer interface {
	PublishProcessingTask(ctx context.Context, task *entity.ProcessingTask) error
	PublishBatch(ctx context.Context, tasks []*entity.ProcessingTask) error
	Close() error
}
