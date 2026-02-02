package handler

import (
	"context"
	"imageprocessor/backend/internal/domain/entity"
	imageservice "imageprocessor/backend/internal/service/image_service"
	"time"
)

// ImageService определяет интерфейс сервиса изображений для хэндлеров
type ImageServiceInterface interface {
	UploadImage(ctx context.Context, imageData []byte, filename string, mimeType string, operations []entity.OperationParams) (*entity.Image, error)
	GetImage(ctx context.Context, imageID string, operation entity.OperationType) ([]byte, string, error)
	GetImagePresignedURL(ctx context.Context, imageID string, operation entity.OperationType, expiry time.Duration) (string, error)
	DeleteImage(ctx context.Context, imageID string) error
	GetImageStatus(ctx context.Context, imageID string) (*imageservice.ImageStatus, error)
	ListImages(ctx context.Context, limit, offset int) ([]entity.Image, error)
}

// StatisticsService определяет интерфейс сервиса статистики для хэндлеров
type StatisticsServiceInterface interface {
	GetStatistics(ctx context.Context) (*entity.ProcessingStatistics, error)
	GetOperationStatistics(ctx context.Context) ([]entity.OperationStat, error)
	GetDetailedStatistics(ctx context.Context) (*entity.DetailedStatistics, error)
	RecordImageUploaded(ctx context.Context, size int64) error
	RecordImageProcessed(ctx context.Context, operation entity.OperationType, processingTimeMs float64) error
	RecordImageFailed(ctx context.Context, operation entity.OperationType, processingTimeMs float64) error
}
