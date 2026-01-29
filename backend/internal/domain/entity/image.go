package entity

import (
	"time"
)

// Image представляет изображение в системе
type Image struct {
	ID              string
	OriginalName    string
	StoragePath     string
	ContentType     string
	Size            int64
	Width           int
	Height          int
	Format          string
	UploadedAt      time.Time
	LastModifiedAt  time.Time
	IsProcessed     bool
	ProcessingError string
	Tags            []string
}

// ProcessedImage представляет обработанное изображение
type ProcessedImage struct {
	ID              string
	OriginalImageID string
	OperationType   string
	StoragePath     string
	ContentType     string
	Size            int64
	Width           int
	Height          int
	ProcessedAt     time.Time
}
