package dto

import (
	"fmt"
	"imageprocessor/backend/internal/domain/entity"
	"mime/multipart"
	"strings"
)

// UploadImageRequest представляет запрос на загрузку изображения
type UploadImageRequest struct {
	Image      *multipart.FileHeader `form:"image" binding:"required"`
	Operations []OperationRequest    `form:"operations"`
}

// OperationRequest представляет операцию обработки изображения
type OperationRequest struct {
	Type       string                 `json:"type" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
}

// GetImageRequest представляет запрос на получение изображения
type GetImageRequest struct {
	ID        string `uri:"id" binding:"required"`
	Version   string `form:"version"`
	Operation string `form:"operation"`
}

// DeleteImageRequest представляет запрос на удаление изображения
type DeleteImageRequest struct {
	ID string `uri:"id" binding:"required"`
}

// GetImageStatusRequest представляет запрос на получение статуса изображения
type GetImageStatusRequest struct {
	ID string `uri:"id" binding:"required"`
}

// Валидация запроса на загрузку изображения
func (r *UploadImageRequest) Validate() error {
	if r.Image == nil {
		return fmt.Errorf("image file is required")
	}

	// Проверка размера файла (32MB максимум)
	if r.Image.Size > entity.DefaultMaxUploadSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", entity.DefaultMaxUploadSize)
	}

	// Проверка MIME типа
	contentType := r.Image.Header.Get("Content-Type")
	if !isValidImageContentType(contentType) {
		return fmt.Errorf("invalid content type: %s. Supported types: image/jpeg, image/png, image/gif, image/webp", contentType)
	}

	// Проверка расширения файла
	if !isValidImageExtension(r.Image.Filename) {
		return fmt.Errorf("invalid file extension. Supported: .jpg, .jpeg, .png, .gif, .webp")
	}

	// Валидация операций
	if len(r.Operations) > 0 {
		for i, op := range r.Operations {
			if err := op.Validate(); err != nil {
				return fmt.Errorf("invalid operation at index %d: %w", i, err)
			}
		}
	}

	return nil
}

// Валидация операции
func (o *OperationRequest) Validate() error {
	if o.Type == "" {
		return fmt.Errorf("operation type is required")
	}

	// Проверяем, что тип операции валидный
	validTypes := map[string]bool{
		string(entity.OpResize):    true,
		string(entity.OpThumbnail): true,
		string(entity.OpWatermark): true,
		string(entity.OpCrop):      true,
		string(entity.OpRotate):    true,
		string(entity.OpFlip):      true,
		string(entity.OpGrayscale): true,
	}

	if !validTypes[o.Type] {
		return fmt.Errorf("invalid operation type: %s", o.Type)
	}

	// Валидация параметров в зависимости от типа операции
	switch entity.OperationType(o.Type) {
	case entity.OpResize:
		return o.validateResizeParams()
	case entity.OpThumbnail:
		return o.validateThumbnailParams()
	case entity.OpWatermark:
		return o.validateWatermarkParams()
	}

	return nil
}

func (o *OperationRequest) validateResizeParams() error {
	if o.Parameters == nil {
		return fmt.Errorf("resize parameters are required")
	}

	width, hasWidth := o.Parameters[entity.ParamWidth]
	height, hasHeight := o.Parameters[entity.ParamHeight]

	if !hasWidth && !hasHeight {
		return fmt.Errorf("at least one of width or height is required for resize")
	}

	if hasWidth {
		w := getFloat64(width)
		if w <= 0 || w > 4096 {
			return fmt.Errorf("width must be between 1 and 4096")
		}
	}

	if hasHeight {
		h := getFloat64(height)
		if h <= 0 || h > 4096 {
			return fmt.Errorf("height must be between 1 and 4096")
		}
	}

	return nil
}

func (o *OperationRequest) validateThumbnailParams() error {
	if o.Parameters == nil {
		o.Parameters = make(map[string]interface{})
		o.Parameters[entity.ParamSize] = entity.DefaultThumbnailSize
		return nil
	}

	if size, hasSize := o.Parameters[entity.ParamSize]; hasSize {
		s := getFloat64(size)
		if s <= 0 || s > 1000 {
			return fmt.Errorf("thumbnail size must be between 1 and 1000")
		}
	} else {
		o.Parameters[entity.ParamSize] = entity.DefaultThumbnailSize
	}

	return nil
}

func (o *OperationRequest) validateWatermarkParams() error {
	if o.Parameters == nil {
		o.Parameters = make(map[string]interface{})
	}

	// Устанавливаем значения по умолчанию
	if _, hasText := o.Parameters[entity.ParamText]; !hasText {
		o.Parameters[entity.ParamText] = entity.DefaultWatermarkText
	}

	if _, hasOpacity := o.Parameters[entity.ParamOpacity]; !hasOpacity {
		o.Parameters[entity.ParamOpacity] = entity.DefaultWatermarkOpacity
	}

	if opacity, hasOpacity := o.Parameters[entity.ParamOpacity]; hasOpacity {
		op := getFloat64(opacity)
		if op < 0 || op > 1 {
			return fmt.Errorf("watermark opacity must be between 0 and 1")
		}
	}

	if position, hasPosition := o.Parameters[entity.ParamPosition]; hasPosition {
		pos, ok := position.(string)
		if !ok {
			return fmt.Errorf("watermark position must be a string")
		}

		validPositions := map[string]bool{
			string(entity.WatermarkTopLeft):      true,
			string(entity.WatermarkTopRight):     true,
			string(entity.WatermarkTopCenter):    true,
			string(entity.WatermarkBottomLeft):   true,
			string(entity.WatermarkBottomRight):  true,
			string(entity.WatermarkBottomCenter): true,
			string(entity.WatermarkCenter):       true,
		}

		if !validPositions[pos] {
			return fmt.Errorf("invalid watermark position: %s", pos)
		}
	}

	return nil
}

// ToEntity конвертирует DTO в entity
func (o *OperationRequest) ToEntity() entity.OperationParams {
	return entity.OperationParams{
		Type:       entity.OperationType(o.Type),
		Parameters: o.Parameters,
	}
}

// Вспомогательные функции

func isValidImageContentType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[contentType]
}

func isValidImageExtension(filename string) bool {
	lower := strings.ToLower(filename)
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	for _, ext := range validExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func getFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}
