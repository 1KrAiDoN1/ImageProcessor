package operations

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"imageprocessor/backend/internal/domain/entity"

	"github.com/disintegration/imaging"
)

type ResizeOperation struct{}

func NewResizeOperation() *ResizeOperation {
	return &ResizeOperation{}
}

func (o *ResizeOperation) GetOperationType() entity.OperationType {
	return entity.OpResize
}

func (o *ResizeOperation) Validate(params map[string]interface{}) error {
	width, hasWidth := params[entity.ParamWidth]
	height, hasHeight := params[entity.ParamHeight]

	if !hasWidth && !hasHeight {
		return fmt.Errorf("width or height parameter is required")
	}

	if hasWidth {
		w, ok := width.(float64)
		if !ok {
			wInt, ok := width.(int)
			if !ok {
				return fmt.Errorf("width must be a number")
			}
			w = float64(wInt)
		}
		if w <= 0 {
			return fmt.Errorf("width must be positive")
		}
	}

	if hasHeight {
		h, ok := height.(float64)
		if !ok {
			hInt, ok := height.(int)
			if !ok {
				return fmt.Errorf("height must be a number")
			}
			h = float64(hInt)
		}
		if h <= 0 {
			return fmt.Errorf("height must be positive")
		}
	}

	return nil
}

func (o *ResizeOperation) Execute(imageData []byte, params map[string]interface{}) ([]byte, error) {
	// Декодируем изображение
	img, format, err := DecodeImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Получаем параметры
	width := getIntParam(params, entity.ParamWidth, 0)
	height := getIntParam(params, entity.ParamHeight, 0)
	keepAspect := getBoolParam(params, entity.ParamKeepAspect, true)

	var resized *image.NRGBA

	if width == 0 && height == 0 {
		return nil, fmt.Errorf("width or height must be specified")
	}

	if keepAspect {
		// Сохраняем пропорции
		if width > 0 && height > 0 {
			// Изменяем размер, вписывая в заданные границы
			resized = imaging.Fit(img, width, height, imaging.Lanczos)
		} else if width > 0 {
			// Изменяем только ширину, высота пропорциональна
			resized = imaging.Resize(img, width, 0, imaging.Lanczos)
		} else {
			// Изменяем только высоту, ширина пропорциональна
			resized = imaging.Resize(img, 0, height, imaging.Lanczos)
		}
	} else {
		// Не сохраняем пропорции, растягиваем до указанных размеров
		if width == 0 {
			width = img.Bounds().Dx()
		}
		if height == 0 {
			height = img.Bounds().Dy()
		}
		resized = imaging.Resize(img, width, height, imaging.Lanczos)
	}

	// Кодируем обратно в байты
	quality := getIntParam(params, "quality", entity.DefaultJPEGQuality)
	return EncodeImage(resized, format, quality)
}

// Вспомогательные функции для извлечения параметров
func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if val, exists := params[key]; exists {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case int64:
			return int(v)
		}
	}
	return defaultValue
}

func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, exists := params[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getStringParam(params map[string]interface{}, key string, defaultValue string) string {
	if val, exists := params[key]; exists {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getFloat64Param(params map[string]interface{}, key string, defaultValue float64) float64 {
	if val, exists := params[key]; exists {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return defaultValue
}

// DecodeImage декодирует изображение из байтов
func DecodeImage(data []byte) (image.Image, entity.ImageFormat, error) {
	reader := bytes.NewReader(data)

	// Сначала определяем формат
	_, formatStr, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image config: %w", err)
	}

	// Сбрасываем reader
	_, err = reader.Seek(0, 0)
	if err != nil {
		return nil, "", fmt.Errorf("failed to reset reader: %w", err)
	}

	// Декодируем изображение
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	var format entity.ImageFormat
	switch formatStr {
	case "jpeg":
		format = entity.FormatJPEG
	case "png":
		format = entity.FormatPNG
	case "gif":
		format = entity.FormatGIF
	case "webp":
		format = entity.FormatWebP
	default:
		format = entity.ImageFormat(formatStr)
	}

	return img, format, nil
}

// EncodeImage кодирует изображение в байты
func EncodeImage(img image.Image, format entity.ImageFormat, quality int) ([]byte, error) {
	var buf bytes.Buffer

	switch format {
	case entity.FormatJPEG, entity.FormatJPG:
		if quality <= 0 || quality > 100 {
			quality = entity.DefaultJPEGQuality
		}
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, fmt.Errorf("failed to encode JPEG: %w", err)
		}
	case entity.FormatPNG:
		encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
		err := encoder.Encode(&buf, img)
		if err != nil {
			return nil, fmt.Errorf("failed to encode PNG: %w", err)
		}
	case entity.FormatGIF:
		err := gif.Encode(&buf, img, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to encode GIF: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}

	return buf.Bytes(), nil
}
