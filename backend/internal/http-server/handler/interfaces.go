package handler

import (
	"context"
	"imageprocessor/backend/internal/domain/entity"
	"io"
)

type imageServiceInterface interface {
	UploadImage(ctx context.Context, file io.Reader, image *entity.Image) (entity.Image, error)
	GetImage(ctx context.Context, id, operation string) (entity.Image, io.ReadCloser, error)
	DeleteImage(ctx context.Context, id string) error
}

type statisticServiceInterface interface {
	GetImageStatistics(ctx context.Context, imageID string) (entity.ProcessingStatistics, error)
}
