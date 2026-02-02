package imageservice

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

type ImageService struct {
	imageRepo     ImageRepositoryInterface
	cloudStorage  cloud.CloudStorageInterface
	kafkaProducer KafkaProducer
	logger        *zap.Logger
	bucket        string
}

func NewImageService(
	imageRepo ImageRepositoryInterface,
	cloudStorage cloud.CloudStorageInterface,
	kafkaProducer KafkaProducer,
	logger *zap.Logger,
	bucket string,
) *ImageService {
	return &ImageService{
		imageRepo:     imageRepo,
		cloudStorage:  cloudStorage,
		kafkaProducer: kafkaProducer,
		logger:        logger,
		bucket:        bucket,
	}
}

// UploadImage загружает изображение, сохраняет в S3 и БД, публикует задачу в Kafka
func (s *ImageService) UploadImage(ctx context.Context, imageData []byte, filename string, mimeType string, operations []entity.OperationParams) (*entity.Image, error) {
	s.logger.Info("Uploading image",
		zap.String("filename", filename),
		zap.Int("size", len(imageData)),
		zap.Int("operationsCount", len(operations)),
	)

	// Генерируем уникальный ID
	imageID := uuid.New().String()

	// Определяем путь для оригинала в S3
	originalPath := fmt.Sprintf("originals/%s/%s", imageID, filename)

	// Загружаем оригинал в S3
	err := s.cloudStorage.UploadFile(ctx, originalPath, bytes.NewReader(imageData), int64(len(imageData)), mimeType)
	if err != nil {
		s.logger.Error("Failed to upload to S3", zap.Error(err), zap.String("imageId", imageID))
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	s.logger.Info("Image uploaded to S3", zap.String("imageId", imageID), zap.String("path", originalPath))

	// Создаем запись в БД
	image := &entity.Image{
		ID:               imageID,
		OriginalFilename: filename,
		OriginalSize:     int64(len(imageData)),
		MimeType:         mimeType,
		Status:           entity.StatusUploaded,
		OriginalPath:     originalPath,
		Bucket:           s.bucket,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = s.imageRepo.CreateImage(ctx, image)
	if err != nil {
		s.logger.Error("Failed to create image in DB", zap.Error(err), zap.String("imageId", imageID))
		// Пытаемся откатить загрузку в S3
		_ = s.cloudStorage.DeleteFile(ctx, originalPath)
		return nil, fmt.Errorf("failed to create image in DB: %w", err)
	}

	s.logger.Info("Image record created in DB", zap.String("imageId", imageID))

	// Если есть операции, создаем задачу на обработку
	if len(operations) > 0 {
		// Определяем формат изображения
		format := s.detectFormat(filename, mimeType)

		// Создаем задачу на обработку
		task := &entity.ProcessingTask{
			ID:           uuid.New().String(),
			ImageID:      imageID,
			OriginalPath: originalPath,
			Bucket:       s.bucket,
			Operations:   operations,
			Format:       format,
		}

		// Создаем запись о задаче в БД
		err = s.imageRepo.CreateProcessingJob(ctx, task)
		if err != nil {
			s.logger.Error("Failed to create processing job", zap.Error(err), zap.String("imageId", imageID))
			return image, fmt.Errorf("failed to create processing job: %w", err)
		}

		// Публикуем задачу в Kafka
		err = s.kafkaProducer.PublishProcessingTask(ctx, task)
		if err != nil {
			s.logger.Error("Failed to publish task to Kafka", zap.Error(err), zap.String("imageId", imageID))
			// Не возвращаем ошибку, так как изображение уже загружено
			// Можно добавить механизм повторной отправки
		} else {
			// Обновляем статус на "processing"
			_ = s.imageRepo.UpdateImageStatus(ctx, imageID, entity.StatusProcessing)
			s.logger.Info("Processing task published to Kafka", zap.String("taskId", task.ID), zap.String("imageId", imageID))
		}
	}

	return image, nil
}

// GetImage получает изображение по ID и операции
func (s *ImageService) GetImage(ctx context.Context, imageID string, operation entity.OperationType) ([]byte, string, error) {
	s.logger.Debug("Getting image", zap.String("imageId", imageID), zap.String("operation", string(operation)))

	// Получаем метаданные из БД
	image, err := s.imageRepo.GetImageByID(ctx, imageID)
	if err != nil {
		s.logger.Error("Image not found", zap.Error(err), zap.String("imageId", imageID))
		return nil, "", fmt.Errorf("image not found: %w", err)
	}

	var objectPath string

	// Если запрашивают оригинал
	if operation == "" || operation == "original" {
		objectPath = image.OriginalPath
	} else {
		// Получаем обработанное изображение
		processed, err := s.imageRepo.GetProcessedImageByOperation(ctx, imageID, operation)
		if err != nil {
			s.logger.Error("Processed image not found",
				zap.Error(err),
				zap.String("imageId", imageID),
				zap.String("operation", string(operation)),
			)
			return nil, "", fmt.Errorf("processed image not found: %w", err)
		}
		objectPath = processed.Path
	}

	// Загружаем файл из S3
	data, err := s.cloudStorage.DownloadFile(ctx, objectPath)
	if err != nil {
		s.logger.Error("Failed to download from S3", zap.Error(err), zap.String("path", objectPath))
		return nil, "", fmt.Errorf("failed to download from S3: %w", err)
	}

	s.logger.Info("Image downloaded successfully", zap.String("imageId", imageID), zap.Int("size", len(data)))

	return data, image.MimeType, nil
}

// GetImagePresignedURL генерирует временную ссылку на изображение
func (s *ImageService) GetImagePresignedURL(ctx context.Context, imageID string, operation entity.OperationType, expiry time.Duration) (string, error) {
	// Получаем метаданные из БД
	image, err := s.imageRepo.GetImageByID(ctx, imageID)
	if err != nil {
		return "", fmt.Errorf("image not found: %w", err)
	}

	var objectPath string

	if operation == "" || operation == "original" {
		objectPath = image.OriginalPath
	} else {
		processed, err := s.imageRepo.GetProcessedImageByOperation(ctx, imageID, operation)
		if err != nil {
			return "", fmt.Errorf("processed image not found: %w", err)
		}
		objectPath = processed.Path
	}

	// Генерируем presigned URL
	url, err := s.cloudStorage.GetPresignedURL(ctx, objectPath, expiry)
	if err != nil {
		s.logger.Error("Failed to generate presigned URL", zap.Error(err))
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

// DeleteImage удаляет изображение и все его версии
func (s *ImageService) DeleteImage(ctx context.Context, imageID string) error {
	s.logger.Info("Deleting image", zap.String("imageId", imageID))

	// Получаем метаданные
	image, err := s.imageRepo.GetImageByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	// Получаем все обработанные версии
	processedImages, err := s.imageRepo.GetProcessedImagesByImageID(ctx, imageID)
	if err != nil {
		s.logger.Warn("Failed to get processed images", zap.Error(err))
	}

	// Собираем все пути для удаления
	pathsToDelete := []string{image.OriginalPath}
	for _, processed := range processedImages {
		pathsToDelete = append(pathsToDelete, processed.Path)
	}

	// Удаляем файлы из S3
	if err := s.cloudStorage.DeleteFiles(ctx, pathsToDelete); err != nil {
		s.logger.Error("Failed to delete files from S3", zap.Error(err))
		// Продолжаем удаление из БД
	}

	// Удаляем запись из БД (каскадно удалятся и обработанные версии)
	err = s.imageRepo.DeleteImage(ctx, imageID)
	if err != nil {
		s.logger.Error("Failed to delete image from DB", zap.Error(err))
		return fmt.Errorf("failed to delete image from DB: %w", err)
	}

	s.logger.Info("Image deleted successfully", zap.String("imageId", imageID))
	return nil
}

// GetImageStatus получает статус обработки изображения
func (s *ImageService) GetImageStatus(ctx context.Context, imageID string) (*ImageStatus, error) {
	// Получаем изображение
	image, err := s.imageRepo.GetImageByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// Получаем обработанные версии
	processedImages, err := s.imageRepo.GetProcessedImagesByImageID(ctx, imageID)
	if err != nil {
		s.logger.Warn("Failed to get processed images", zap.Error(err))
	}

	// Получаем информацию о задаче
	job, err := s.imageRepo.GetProcessingJobByImageID(ctx, imageID)
	if err != nil {
		s.logger.Warn("Failed to get processing job", zap.Error(err))
	}

	status := &ImageStatus{
		ID:                  image.ID,
		Status:              image.Status,
		OriginalFilename:    image.OriginalFilename,
		ProcessedOperations: len(processedImages),
		TotalOperations:     0,
		CreatedAt:           image.CreatedAt,
		UpdatedAt:           image.UpdatedAt,
	}

	if job != nil {
		status.TotalOperations = len(job.Operations)
	}

	// Вычисляем прогресс
	if status.TotalOperations > 0 {
		status.Progress = (status.ProcessedOperations * 100) / status.TotalOperations
	}

	return status, nil
}

// ListImages возвращает список изображений
func (s *ImageService) ListImages(ctx context.Context, limit, offset int) ([]entity.Image, error) {
	return s.imageRepo.ListImages(ctx, limit, offset)
}

// detectFormat определяет формат изображения по имени файла или MIME типу
func (s *ImageService) detectFormat(filename, mimeType string) entity.ImageFormat {
	// Пытаемся определить по MIME типу
	switch mimeType {
	case "image/jpeg":
		return entity.FormatJPEG
	case "image/png":
		return entity.FormatPNG
	case "image/gif":
		return entity.FormatGIF
	case "image/webp":
		return entity.FormatWebP
	}

	// По умолчанию JPEG
	return entity.FormatJPEG
}

// ImageStatus представляет статус обработки изображения
type ImageStatus struct {
	ID                  string             `json:"id"`
	Status              entity.ImageStatus `json:"status"`
	OriginalFilename    string             `json:"original_filename"`
	ProcessedOperations int                `json:"processed_operations"`
	TotalOperations     int                `json:"total_operations"`
	Progress            int                `json:"progress"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}
