package entity

type ProcessingTask struct {
	ID           string
	ImageID      string
	OriginalPath string
	Bucket       string
	Operations   []OperationParams
	Format       ImageFormat
}

type OperationParams struct {
	Type       OperationType
	Parameters map[string]interface{}
}

type ProcessingResult struct {
	ID             string
	ImageID        string
	Status         ImageStatus
	ProcessedPaths map[string]string
	Error          string
}

type WatermarkPosition string

const (
	WatermarkTopLeft      WatermarkPosition = "top-left"
	WatermarkTopRight     WatermarkPosition = "top-right"
	WatermarkTopCenter    WatermarkPosition = "top-center"
	WatermarkBottomLeft   WatermarkPosition = "bottom-left"
	WatermarkBottomRight  WatermarkPosition = "bottom-right"
	WatermarkBottomCenter WatermarkPosition = "bottom-center"
	WatermarkCenter       WatermarkPosition = "center"
)

const (
	KafkaTopicProcessing = "image-processing"
	KafkaTopicResults    = "image-processed"
	KafkaGroupID         = "image-processor-group"
)

const (
	BucketOriginal  = "original"
	BucketProcessed = "processed"
)

const (
	PathPrefixOriginal  = "original/"
	PathPrefixProcessed = "processed/"
	PathPrefixThumbnail = "thumbnails/"
)

const (
	DefaultMaxUploadSize    = 32 << 20
	DefaultThumbnailSize    = 200
	DefaultJPEGQuality      = 85
	DefaultWatermarkText    = "© ImageProcessor"
	DefaultWatermarkOpacity = 0.5
)

const (
	ParamWidth      = "width"
	ParamHeight     = "height"
	ParamSize       = "size"
	ParamText       = "text"
	ParamPosition   = "position"
	ParamOpacity    = "opacity"
	ParamFontSize   = "font_size"
	ParamFontColor  = "font_color"
	ParamKeepAspect = "keep_aspect"
	ParamCropToFit  = "crop_to_fit"
	ParamAngle      = "angle"
)
