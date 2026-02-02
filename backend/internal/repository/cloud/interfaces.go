package cloud

import (
	"context"
	"io"
	"time"
)

// CloudStorageInterface определяет методы для работы с облачным хранилищем
type CloudStorageInterface interface {
	// UploadFile загружает файл в хранилище
	UploadFile(ctx context.Context, objectKey string, data io.Reader, size int64, contentType string) error

	// DownloadFile скачивает файл из хранилища
	DownloadFile(ctx context.Context, objectKey string) ([]byte, error)

	// DownloadFileStream скачивает файл как поток
	DownloadFileStream(ctx context.Context, objectKey string) (io.ReadCloser, error)

	// DeleteFile удаляет файл из хранилища
	DeleteFile(ctx context.Context, objectKey string) error

	// DeleteFiles удаляет множество файлов
	DeleteFiles(ctx context.Context, objectKeys []string) error

	// FileExists проверяет существование файла
	FileExists(ctx context.Context, objectKey string) (bool, error)

	// GetFileSize возвращает размер файла
	GetFileSize(ctx context.Context, objectKey string) (int64, error)

	// GetPresignedURL генерирует временную ссылку для скачивания
	GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)

	// ListFiles возвращает список файлов с префиксом
	ListFiles(ctx context.Context, prefix string) ([]string, error)

	// CopyFile копирует файл внутри хранилища
	CopyFile(ctx context.Context, sourceKey, destKey string) error
}
