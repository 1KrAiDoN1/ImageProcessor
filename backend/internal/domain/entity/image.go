package entity

import (
	"time"
)

type Image struct {
	ID               string
	OriginalFilename string
	OriginalSize     int64
	MimeType         string
	Status           ImageStatus
	OriginalPath     string
	Bucket           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ProcessedImage struct {
	ID         string
	ImageID    string
	Operation  OperationType
	Parameters string
	Path       string
	Size       int64
	MimeType   string
	Format     ImageFormat
	Status     string
	CreatedAt  time.Time
}

type ImageStatus string

const (
	StatusUploaded   ImageStatus = "uploaded"
	StatusProcessing ImageStatus = "processing"
	StatusCompleted  ImageStatus = "completed"
	StatusFailed     ImageStatus = "failed"
	StatusDeleted    ImageStatus = "deleted"
)

type OperationType string

const (
	OpResize    OperationType = "resize"
	OpThumbnail OperationType = "thumbnail"
	OpWatermark OperationType = "watermark"
	OpCrop      OperationType = "crop"
	OpRotate    OperationType = "rotate"
	OpFlip      OperationType = "flip"
	OpGrayscale OperationType = "grayscale"
)

type ImageFormat string

const (
	FormatJPEG ImageFormat = "jpeg"
	FormatJPG  ImageFormat = "jpg"
	FormatPNG  ImageFormat = "png"
	FormatGIF  ImageFormat = "gif"
	FormatWebP ImageFormat = "webp"
	FormatBMP  ImageFormat = "bmp"
	FormatTIFF ImageFormat = "tiff"
)

type ImageInfo struct {
	Width      int
	Height     int
	Format     ImageFormat
	Size       int64
	ColorSpace string
	HasAlpha   bool
}
