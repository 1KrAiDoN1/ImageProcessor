package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"imageprocessor/backend/internal/domain/entity"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ImageRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewImageRepository(db *pgxpool.Pool, logger *zap.Logger) *ImageRepository {
	return &ImageRepository{
		db:     db,
		logger: logger,
	}
}

// CreateImage создает запись об изображении в БД
func (r *ImageRepository) CreateImage(ctx context.Context, image *entity.Image) error {
	query := `
		INSERT INTO images (id, original_filename, original_size, mime_type, status, original_path, bucket, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		image.ID,
		image.OriginalFilename,
		image.OriginalSize,
		image.MimeType,
		image.Status,
		image.OriginalPath,
		image.Bucket,
		image.CreatedAt,
		image.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create image", zap.Error(err), zap.String("imageId", image.ID))
		return fmt.Errorf("failed to create image: %w", err)
	}

	r.logger.Info("Image created successfully", zap.String("imageId", image.ID))
	return nil
}

// GetImageByID получает изображение по ID
func (r *ImageRepository) GetImageByID(ctx context.Context, imageID string) (*entity.Image, error) {
	query := `
		SELECT id, original_filename, original_size, mime_type, status, original_path, bucket, created_at, updated_at
		FROM images
		WHERE id = $1
	`

	var image entity.Image
	err := r.db.QueryRow(ctx, query, imageID).Scan(
		&image.ID,
		&image.OriginalFilename,
		&image.OriginalSize,
		&image.MimeType,
		&image.Status,
		&image.OriginalPath,
		&image.Bucket,
		&image.CreatedAt,
		&image.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("image not found: %s", imageID)
		}
		r.logger.Error("Failed to get image", zap.Error(err), zap.String("imageId", imageID))
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return &image, nil
}

// UpdateImageStatus обновляет статус изображения
func (r *ImageRepository) UpdateImageStatus(ctx context.Context, imageID string, status entity.ImageStatus) error {
	query := `
		UPDATE images
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(ctx, query, status, time.Now(), imageID)
	if err != nil {
		r.logger.Error("Failed to update image status", zap.Error(err), zap.String("imageId", imageID))
		return fmt.Errorf("failed to update image status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("image not found: %s", imageID)
	}

	r.logger.Info("Image status updated", zap.String("imageId", imageID), zap.String("status", string(status)))
	return nil
}

// DeleteImage удаляет изображение из БД
func (r *ImageRepository) DeleteImage(ctx context.Context, imageID string) error {
	query := `DELETE FROM images WHERE id = $1`

	result, err := r.db.Exec(ctx, query, imageID)
	if err != nil {
		r.logger.Error("Failed to delete image", zap.Error(err), zap.String("imageId", imageID))
		return fmt.Errorf("failed to delete image: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("image not found: %s", imageID)
	}

	r.logger.Info("Image deleted successfully", zap.String("imageId", imageID))
	return nil
}

// ListImages возвращает список изображений с пагинацией
func (r *ImageRepository) ListImages(ctx context.Context, limit, offset int) ([]entity.Image, error) {
	query := `
		SELECT id, original_filename, original_size, mime_type, status, original_path, bucket, created_at, updated_at
		FROM images
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to list images", zap.Error(err))
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	defer rows.Close()

	var images []entity.Image
	for rows.Next() {
		var image entity.Image
		err := rows.Scan(
			&image.ID,
			&image.OriginalFilename,
			&image.OriginalSize,
			&image.MimeType,
			&image.Status,
			&image.OriginalPath,
			&image.Bucket,
			&image.CreatedAt,
			&image.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan image row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, image)
	}

	return images, nil
}

// CreateProcessedImage создает запись об обработанном изображении
func (r *ImageRepository) CreateProcessedImage(ctx context.Context, processed *entity.ProcessedImage) error {
	paramsJSON, err := json.Marshal(processed.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	query := `
		INSERT INTO processed_images (id, image_id, operation, parameters, path, size, mime_type, format, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Exec(ctx, query,
		processed.ID,
		processed.ImageID,
		processed.Operation,
		paramsJSON,
		processed.Path,
		processed.Size,
		processed.MimeType,
		processed.Format,
		processed.Status,
		processed.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create processed image",
			zap.Error(err),
			zap.String("imageId", processed.ImageID),
			zap.String("operation", string(processed.Operation)),
		)
		return fmt.Errorf("failed to create processed image: %w", err)
	}

	r.logger.Info("Processed image created",
		zap.String("imageId", processed.ImageID),
		zap.String("operation", string(processed.Operation)),
	)
	return nil
}

// GetProcessedImagesByImageID получает все обработанные версии изображения
func (r *ImageRepository) GetProcessedImagesByImageID(ctx context.Context, imageID string) ([]entity.ProcessedImage, error) {
	query := `
		SELECT id, image_id, operation, parameters, path, size, mime_type, format, status, created_at
		FROM processed_images
		WHERE image_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, imageID)
	if err != nil {
		r.logger.Error("Failed to get processed images", zap.Error(err), zap.String("imageId", imageID))
		return nil, fmt.Errorf("failed to get processed images: %w", err)
	}
	defer rows.Close()

	var processedImages []entity.ProcessedImage
	for rows.Next() {
		var processed entity.ProcessedImage
		var paramsJSON []byte

		err := rows.Scan(
			&processed.ID,
			&processed.ImageID,
			&processed.Operation,
			&paramsJSON,
			&processed.Path,
			&processed.Size,
			&processed.MimeType,
			&processed.Format,
			&processed.Status,
			&processed.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan processed image", zap.Error(err))
			return nil, fmt.Errorf("failed to scan processed image: %w", err)
		}

		if err := json.Unmarshal(paramsJSON, &processed.Parameters); err != nil {
			r.logger.Warn("Failed to unmarshal parameters", zap.Error(err))
		}

		processedImages = append(processedImages, processed)
	}

	return processedImages, nil
}

// GetProcessedImageByOperation получает обработанное изображение по типу операции
func (r *ImageRepository) GetProcessedImageByOperation(ctx context.Context, imageID string, operation entity.OperationType) (*entity.ProcessedImage, error) {
	query := `
		SELECT id, image_id, operation, parameters, path, size, mime_type, format, status, created_at
		FROM processed_images
		WHERE image_id = $1 AND operation = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var processed entity.ProcessedImage
	var paramsJSON []byte

	err := r.db.QueryRow(ctx, query, imageID, operation).Scan(
		&processed.ID,
		&processed.ImageID,
		&processed.Operation,
		&paramsJSON,
		&processed.Path,
		&processed.Size,
		&processed.MimeType,
		&processed.Format,
		&processed.Status,
		&processed.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("processed image not found")
		}
		r.logger.Error("Failed to get processed image", zap.Error(err))
		return nil, fmt.Errorf("failed to get processed image: %w", err)
	}

	if err := json.Unmarshal(paramsJSON, &processed.Parameters); err != nil {
		r.logger.Warn("Failed to unmarshal parameters", zap.Error(err))
	}

	return &processed, nil
}

// CreateProcessingJob создает задачу на обработку
func (r *ImageRepository) CreateProcessingJob(ctx context.Context, job *entity.ProcessingTask) error {
	operationsJSON, err := json.Marshal(job.Operations)
	if err != nil {
		return fmt.Errorf("failed to marshal operations: %w", err)
	}

	query := `
		INSERT INTO processing_jobs (id, image_id, operations, status, attempts, max_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', 0, 3, $4, $5)
	`

	now := time.Now()
	_, err = r.db.Exec(ctx, query, job.ID, job.ImageID, operationsJSON, now, now)
	if err != nil {
		r.logger.Error("Failed to create processing job", zap.Error(err), zap.String("jobId", job.ID))
		return fmt.Errorf("failed to create processing job: %w", err)
	}

	r.logger.Info("Processing job created", zap.String("jobId", job.ID), zap.String("imageId", job.ImageID))
	return nil
}

// UpdateProcessingJobStatus обновляет статус задачи
func (r *ImageRepository) UpdateProcessingJobStatus(ctx context.Context, jobID string, status string, errorMsg string) error {
	query := `
		UPDATE processing_jobs
		SET status = $1, error_message = $2, updated_at = $3,
		    completed_at = CASE WHEN $1 IN ('completed', 'failed') THEN $3 ELSE completed_at END
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, status, errorMsg, time.Now(), jobID)
	if err != nil {
		r.logger.Error("Failed to update processing job status", zap.Error(err), zap.String("jobId", jobID))
		return fmt.Errorf("failed to update processing job status: %w", err)
	}

	return nil
}

// GetProcessingJobByImageID получает задачу по ID изображения
func (r *ImageRepository) GetProcessingJobByImageID(ctx context.Context, imageID string) (*entity.ProcessingTask, error) {
	query := `
		SELECT id, image_id, operations, status, attempts, max_attempts
		FROM processing_jobs
		WHERE image_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var job entity.ProcessingTask
	var operationsJSON []byte
	var status string
	var attempts, maxAttempts int

	err := r.db.QueryRow(ctx, query, imageID).Scan(
		&job.ID,
		&job.ImageID,
		&operationsJSON,
		&status,
		&attempts,
		&maxAttempts,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("processing job not found")
		}
		return nil, fmt.Errorf("failed to get processing job: %w", err)
	}

	if err := json.Unmarshal(operationsJSON, &job.Operations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal operations: %w", err)
	}

	return &job, nil
}
