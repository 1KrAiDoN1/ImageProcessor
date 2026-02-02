package operations

import (
	"fmt"
	"image"
	"imageprocessor/backend/internal/domain/entity"

	"github.com/disintegration/imaging"
)

type ThumbnailOperation struct{}

func NewThumbnailOperation() *ThumbnailOperation {
	return &ThumbnailOperation{}
}

func (o *ThumbnailOperation) GetOperationType() entity.OperationType {
	return entity.OpThumbnail
}

func (o *ThumbnailOperation) Validate(params map[string]interface{}) error {
	size, hasSize := params[entity.ParamSize]
	if hasSize {
		s, ok := size.(float64)
		if !ok {
			sInt, ok := size.(int)
			if !ok {
				return fmt.Errorf("size must be a number")
			}
			s = float64(sInt)
		}
		if s <= 0 {
			return fmt.Errorf("size must be positive")
		}
		if s > 1000 {
			return fmt.Errorf("size must not exceed 1000 pixels")
		}
	}
	return nil
}

func (o *ThumbnailOperation) Execute(imageData []byte, params map[string]interface{}) ([]byte, error) {
	// Декодируем изображение
	img, format, err := DecodeImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Получаем параметры
	size := getIntParam(params, entity.ParamSize, entity.DefaultThumbnailSize)
	cropToFit := getBoolParam(params, entity.ParamCropToFit, false)

	var thumbnail *image.NRGBA

	if cropToFit {
		// Обрезаем и изменяем размер до квадрата
		thumbnail = imaging.Fill(img, size, size, imaging.Center, imaging.Lanczos)
	} else {
		// Вписываем в квадрат с сохранением пропорций
		thumbnail = imaging.Fit(img, size, size, imaging.Lanczos)
	}

	// Кодируем обратно в байты с высоким качеством
	quality := getIntParam(params, "quality", entity.DefaultJPEGQuality)
	return EncodeImage(thumbnail, format, quality)
}
