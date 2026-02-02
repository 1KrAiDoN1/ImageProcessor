package s3

import (
	"bytes"
	"context"
	"fmt"
	"imageprocessor/backend/internal/config"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type S3CloudStorage struct {
	client *minio.Client
	bucket string
	logger *zap.Logger
	config config.CloudStorageConfig
}

// NewS3CloudStorage создает новый клиент для работы с S3-совместимым хранилищем
func NewS3CloudStorage(ctx context.Context, cfg config.CloudStorageConfig, logger *zap.Logger) (*S3CloudStorage, error) {
	logger.Info("Initializing S3 client",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("bucket", cfg.Bucket),
		zap.Bool("useSSL", cfg.UseSSL),
	)

	// Создаем MinIO клиент
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		logger.Error("Failed to create S3 client", zap.Error(err))
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	s3Storage := &S3CloudStorage{
		client: client,
		bucket: cfg.Bucket,
		logger: logger,
		config: cfg,
	}

	// Проверяем существование bucket и создаем если нужно
	if err := s3Storage.ensureBucketExists(ctx); err != nil {
		logger.Error("Failed to ensure bucket exists", zap.Error(err))
		return nil, err
	}

	logger.Info("S3 client initialized successfully")
	return s3Storage, nil
}

// ensureBucketExists проверяет существование bucket и создает его при необходимости
func (s *S3CloudStorage) ensureBucketExists(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		s.logger.Info("Bucket does not exist, creating", zap.String("bucket", s.bucket))
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{
			Region: s.config.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		s.logger.Info("Bucket created successfully", zap.String("bucket", s.bucket))
	}

	return nil
}

// UploadFile загружает файл в S3
func (s *S3CloudStorage) UploadFile(ctx context.Context, objectKey string, data io.Reader, size int64, contentType string) error {
	s.logger.Debug("Uploading file to S3",
		zap.String("objectKey", objectKey),
		zap.Int64("size", size),
		zap.String("contentType", contentType),
	)

	opts := minio.PutObjectOptions{
		ContentType: contentType,
		PartSize:    uint64(s.config.UploadPartSize),
	}

	info, err := s.client.PutObject(ctx, s.bucket, objectKey, data, size, opts)
	if err != nil {
		s.logger.Error("Failed to upload file",
			zap.String("objectKey", objectKey),
			zap.Error(err),
		)
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	s.logger.Info("File uploaded successfully",
		zap.String("objectKey", objectKey),
		zap.Int64("size", info.Size),
		zap.String("etag", info.ETag),
	)

	return nil
}

// DownloadFile скачивает файл из S3 и возвращает его как байтовый массив
func (s *S3CloudStorage) DownloadFile(ctx context.Context, objectKey string) ([]byte, error) {
	s.logger.Debug("Downloading file from S3", zap.String("objectKey", objectKey))

	object, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		s.logger.Error("Failed to get object",
			zap.String("objectKey", objectKey),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer object.Close()

	// Читаем содержимое объекта
	buf := new(bytes.Buffer)
	n, err := io.Copy(buf, object)
	if err != nil {
		s.logger.Error("Failed to read object",
			zap.String("objectKey", objectKey),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	s.logger.Info("File downloaded successfully",
		zap.String("objectKey", objectKey),
		zap.Int64("size", n),
	)

	return buf.Bytes(), nil
}

// DownloadFileStream скачивает файл как поток
func (s *S3CloudStorage) DownloadFileStream(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	s.logger.Debug("Downloading file stream from S3", zap.String("objectKey", objectKey))

	object, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		s.logger.Error("Failed to get object stream",
			zap.String("objectKey", objectKey),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get object stream from S3: %w", err)
	}

	return object, nil
}

// DeleteFile удаляет файл из S3
func (s *S3CloudStorage) DeleteFile(ctx context.Context, objectKey string) error {
	s.logger.Debug("Deleting file from S3", zap.String("objectKey", objectKey))

	err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		s.logger.Error("Failed to delete file",
			zap.String("objectKey", objectKey),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	s.logger.Info("File deleted successfully", zap.String("objectKey", objectKey))
	return nil
}

// DeleteFiles удаляет множество файлов из S3
func (s *S3CloudStorage) DeleteFiles(ctx context.Context, objectKeys []string) error {
	s.logger.Debug("Deleting multiple files from S3", zap.Int("count", len(objectKeys)))

	objectsCh := make(chan minio.ObjectInfo)

	// Запускаем горутину для отправки имен объектов
	go func() {
		defer close(objectsCh)
		for _, key := range objectKeys {
			objectsCh <- minio.ObjectInfo{Key: key}
		}
	}()

	// Удаляем объекты
	errorCh := s.client.RemoveObjects(ctx, s.bucket, objectsCh, minio.RemoveObjectsOptions{})

	var errors []error
	for err := range errorCh {
		if err.Err != nil {
			s.logger.Error("Failed to delete file",
				zap.String("objectKey", err.ObjectName),
				zap.Error(err.Err),
			)
			errors = append(errors, err.Err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete %d files", len(errors))
	}

	s.logger.Info("Files deleted successfully", zap.Int("count", len(objectKeys)))
	return nil
}

// FileExists проверяет существование файла в S3
func (s *S3CloudStorage) FileExists(ctx context.Context, objectKey string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return true, nil
}

// GetFileSize возвращает размер файла
func (s *S3CloudStorage) GetFileSize(ctx context.Context, objectKey string) (int64, error) {
	info, err := s.client.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.Size, nil
}

// GetPresignedURL генерирует временную ссылку для скачивания
func (s *S3CloudStorage) GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	s.logger.Debug("Generating presigned URL",
		zap.String("objectKey", objectKey),
		zap.Duration("expiry", expiry),
	)

	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, expiry, nil)
	if err != nil {
		s.logger.Error("Failed to generate presigned URL",
			zap.String("objectKey", objectKey),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// ListFiles возвращает список файлов с заданным префиксом
func (s *S3CloudStorage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	s.logger.Debug("Listing files", zap.String("prefix", prefix))

	var files []string
	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			s.logger.Error("Error listing objects", zap.Error(object.Err))
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	s.logger.Info("Files listed successfully",
		zap.String("prefix", prefix),
		zap.Int("count", len(files)),
	)

	return files, nil
}

// CopyFile копирует файл внутри хранилища
func (s *S3CloudStorage) CopyFile(ctx context.Context, sourceKey, destKey string) error {
	s.logger.Debug("Copying file",
		zap.String("source", sourceKey),
		zap.String("destination", destKey),
	)

	src := minio.CopySrcOptions{
		Bucket: s.bucket,
		Object: sourceKey,
	}

	dst := minio.CopyDestOptions{
		Bucket: s.bucket,
		Object: destKey,
	}

	_, err := s.client.CopyObject(ctx, dst, src)
	if err != nil {
		s.logger.Error("Failed to copy file",
			zap.String("source", sourceKey),
			zap.String("destination", destKey),
			zap.Error(err),
		)
		return fmt.Errorf("failed to copy file: %w", err)
	}

	s.logger.Info("File copied successfully",
		zap.String("source", sourceKey),
		zap.String("destination", destKey),
	)

	return nil
}

// GetClient возвращает базовый MinIO клиент (для расширенных операций)
func (s *S3CloudStorage) GetClient() *minio.Client {
	return s.client
}
