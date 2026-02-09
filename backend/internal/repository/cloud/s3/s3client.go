package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"imageprocessor/backend/internal/config"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3CloudStorage struct {
	client *minio.Client
	bucket string
	config config.CloudStorageConfig
}

// NewS3CloudStorage создает новый клиент для работы с S3-совместимым хранилищем
func NewS3CloudStorage(ctx context.Context, cfg config.CloudStorageConfig) (*S3CloudStorage, error) {

	// Создаем MinIO клиент
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	s3Storage := &S3CloudStorage{
		client: client,
		bucket: cfg.Bucket,
		config: cfg,
	}

	// Проверяем существование bucket и создаем если нужно
	if err := s3Storage.ensureBucketExists(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return s3Storage, nil
}

// ensureBucketExists проверяет существование bucket и создает его при необходимости
func (s *S3CloudStorage) ensureBucketExists(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{
			Region: s.config.Region,
		})
		if errors.Is(err, fmt.Errorf("your previous request to create the named bucket succeeded and you already own it")) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// UploadFile загружает файл в S3
func (s *S3CloudStorage) UploadFile(ctx context.Context, objectKey string, data io.Reader, size int64, contentType string) error {

	opts := minio.PutObjectOptions{
		ContentType: contentType,
		PartSize:    uint64(s.config.UploadPartSize),
	}

	// Retry logic with exponential backoff
	maxRetries := 5
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// For retries, we need to seek back if data is a seeker
		if attempt > 1 {
			if seeker, ok := data.(io.Seeker); ok {
				_, err := seeker.Seek(0, 0)
				if err != nil {
					return fmt.Errorf("failed to seek data for retry: %w", err)
				}
			}

			backoff := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoff)
		}

		_, err := s.client.PutObject(ctx, s.bucket, objectKey, data, size, opts)
		if err != nil {

			if attempt == maxRetries {
				return fmt.Errorf("failed to upload file to S3 after %d attempts: %w", maxRetries, err)
			}
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to upload file to S3 after %d attempts", maxRetries)
}

// DownloadFile скачивает файл из S3 и возвращает его как байтовый массив
func (s *S3CloudStorage) DownloadFile(ctx context.Context, objectKey string) ([]byte, error) {

	object, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer func() {
		_ = object.Close()
	}()

	// Читаем содержимое объекта
	buf := new(bytes.Buffer)
	n, err := io.Copy(buf, object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	if n == 0 {
		return nil, fmt.Errorf("object is empty")
	}
	return buf.Bytes(), nil
}

// DownloadFileStream скачивает файл как поток
func (s *S3CloudStorage) DownloadFileStream(ctx context.Context, objectKey string) (io.ReadCloser, error) {

	object, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object stream from S3: %w", err)
	}

	return object, nil
}

// DeleteFile удаляет файл из S3
func (s *S3CloudStorage) DeleteFile(ctx context.Context, objectKey string) error {

	err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

// DeleteFiles удаляет множество файлов из S3
func (s *S3CloudStorage) DeleteFiles(ctx context.Context, objectKeys []string) error {

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
			errors = append(errors, err.Err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete %d files", len(errors))
	}

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

	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// ListFiles возвращает список файлов с заданным префиксом
func (s *S3CloudStorage) ListFiles(ctx context.Context, prefix string) ([]string, error) {

	var files []string
	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}

// CopyFile копирует файл внутри хранилища
func (s *S3CloudStorage) CopyFile(ctx context.Context, sourceKey, destKey string) error {

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
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// GetClient возвращает базовый MinIO клиент (для расширенных операций)
func (s *S3CloudStorage) GetClient() *minio.Client {
	return s.client
}
