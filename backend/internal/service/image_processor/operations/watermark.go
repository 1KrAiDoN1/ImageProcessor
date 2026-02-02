package operations

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"imageprocessor/backend/internal/domain/entity"

	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

type WatermarkOperation struct {
	font *truetype.Font
}

func NewWatermarkOperation() *WatermarkOperation {
	// Загружаем встроенный шрифт
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		// Если не удалось загрузить шрифт, используем nil
		// В Execute будем проверять
		font = nil
	}

	return &WatermarkOperation{
		font: font,
	}
}

func (o *WatermarkOperation) GetOperationType() entity.OperationType {
	return entity.OpWatermark
}

func (o *WatermarkOperation) Validate(params map[string]interface{}) error {
	if opacity, exists := params[entity.ParamOpacity]; exists {
		op, ok := opacity.(float64)
		if !ok {
			return fmt.Errorf("opacity must be a number")
		}
		if op < 0 || op > 1 {
			return fmt.Errorf("opacity must be between 0 and 1")
		}
	}

	if position, exists := params[entity.ParamPosition]; exists {
		pos, ok := position.(string)
		if !ok {
			return fmt.Errorf("position must be a string")
		}
		validPositions := map[string]bool{
			"top-left":      true,
			"top-right":     true,
			"top-center":    true,
			"bottom-left":   true,
			"bottom-right":  true,
			"bottom-center": true,
			"center":        true,
		}
		if !validPositions[pos] {
			return fmt.Errorf("invalid position: %s", pos)
		}
	}

	return nil
}

func (o *WatermarkOperation) Execute(imageData []byte, params map[string]interface{}) ([]byte, error) {
	// Декодируем изображение
	img, format, err := DecodeImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Получаем параметры
	text := getStringParam(params, entity.ParamText, entity.DefaultWatermarkText)
	opacity := getFloat64Param(params, entity.ParamOpacity, entity.DefaultWatermarkOpacity)
	position := getStringParam(params, entity.ParamPosition, string(entity.WatermarkBottomRight))
	fontSize := getIntParam(params, entity.ParamFontSize, 24)

	// Добавляем водяной знак
	watermarked, err := o.addTextWatermark(img, text, position, opacity, fontSize)
	if err != nil {
		return nil, fmt.Errorf("failed to add watermark: %w", err)
	}

	// Кодируем обратно в байты
	quality := getIntParam(params, "quality", entity.DefaultJPEGQuality)
	return EncodeImage(watermarked, format, quality)
}

func (o *WatermarkOperation) addTextWatermark(img image.Image, text, position string, opacity float64, fontSize int) (image.Image, error) {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	if o.font == nil {
		return nil, fmt.Errorf("font not available")
	}

	// Создаем контекст для рисования текста
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(o.font)
	c.SetFontSize(float64(fontSize))
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)

	// Цвет текста с прозрачностью
	alpha := uint8(opacity * 255)
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: alpha}
	c.SetSrc(image.NewUniform(textColor))

	// Вычисляем размер текста
	textWidth := len(text) * fontSize / 2 // Приблизительная ширина
	textHeight := fontSize

	// Вычисляем позицию
	var pt fixed.Point26_6
	switch position {
	case "top-left":
		pt = freetype.Pt(10, 10+textHeight)
	case "top-right":
		pt = freetype.Pt(bounds.Dx()-textWidth-10, 10+textHeight)
	case "top-center":
		pt = freetype.Pt((bounds.Dx()-textWidth)/2, 10+textHeight)
	case "bottom-left":
		pt = freetype.Pt(10, bounds.Dy()-10)
	case "bottom-right":
		pt = freetype.Pt(bounds.Dx()-textWidth-10, bounds.Dy()-10)
	case "bottom-center":
		pt = freetype.Pt((bounds.Dx()-textWidth)/2, bounds.Dy()-10)
	case "center":
		pt = freetype.Pt((bounds.Dx()-textWidth)/2, (bounds.Dy()+textHeight)/2)
	default:
		pt = freetype.Pt(bounds.Dx()-textWidth-10, bounds.Dy()-10)
	}

	// Рисуем текст
	_, err := c.DrawString(text, pt)
	if err != nil {
		return nil, fmt.Errorf("failed to draw text: %w", err)
	}

	return rgba, nil
}

// Альтернативная простая реализация без freetype (на случай проблем)
func (o *WatermarkOperation) addSimpleWatermark(img image.Image, opacity float64) image.Image {
	// Создаем копию изображения с наложением полупрозрачного слоя
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Добавляем полупрозрачный прямоугольник в правом нижнем углу
	watermarkBounds := image.Rect(
		bounds.Dx()-200, bounds.Dy()-50,
		bounds.Dx()-10, bounds.Dy()-10,
	)

	alpha := uint8(opacity * 255)
	watermarkColor := color.RGBA{R: 0, G: 0, B: 0, A: alpha}
	watermark := image.NewUniform(watermarkColor)

	draw.Draw(rgba, watermarkBounds, watermark, image.Point{}, draw.Over)

	return imaging.Clone(rgba)
}
