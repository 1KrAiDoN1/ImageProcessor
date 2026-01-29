package imageprocessor

import "context"

type ImageProcessorInterface interface {
	// Resize изменяет размер изображения
	Resize(ctx context.Context, data []byte, width, height int) ([]byte, error)

	// CreateThumbnail создает миниатюру
	CreateThumbnail(ctx context.Context, data []byte, size int) ([]byte, error)

	// AddWatermark добавляет водяной знак
	AddWatermark(ctx context.Context, data []byte, watermarkPath string, opacity float64) ([]byte, error)

	// ValidateFormat проверяет формат изображения
	ValidateFormat(data []byte) (ImageFormat, error)
}
