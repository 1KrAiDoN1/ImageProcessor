package processor

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"imageprocessor/backend/internal/domain/entity"
	imageProcessor "imageprocessor/backend/internal/service/image_processor"
	"imageprocessor/backend/internal/service/image_processor/operations"

	"go.uber.org/zap"
	"golang.org/x/image/webp"
)

type ImageProcessorImpl struct {
	operations map[entity.OperationType]imageProcessor.Operation
	logger     *zap.Logger
}

// NewImageProcessor создает новый процессор изображений
func NewImageProcessor(logger *zap.Logger) *ImageProcessorImpl {
	processor := &ImageProcessorImpl{
		operations: make(map[entity.OperationType]imageProcessor.Operation),
		logger:     logger,
	}

	// Регистрируем все операции
	processor.registerOperation(operations.NewResizeOperation())
	processor.registerOperation(operations.NewThumbnailOperation())
	processor.registerOperation(operations.NewWatermarkOperation())

	logger.Info("Image processor initialized with operations",
		zap.Int("operationCount", len(processor.operations)),
	)

	return processor
}

// registerOperation регистрирует операцию в процессоре
func (p *ImageProcessorImpl) registerOperation(op imageProcessor.Operation) {
	p.operations[op.GetOperationType()] = op
	p.logger.Debug("Registered operation", zap.String("type", string(op.GetOperationType())))
}

// ProcessImage обрабатывает изображение согласно списку операций
func (p *ImageProcessorImpl) ProcessImage(ctx context.Context, imageData []byte, operations []entity.OperationParams) (map[string][]byte, error) {
	p.logger.Info("Processing image with operations",
		zap.Int("dataSize", len(imageData)),
		zap.Int("operationCount", len(operations)),
	)

	// Валидируем изображение
	format, err := p.ValidateImage(imageData)
	if err != nil {
		p.logger.Error("Image validation failed", zap.Error(err))
		return nil, fmt.Errorf("invalid image: %w", err)
	}

	p.logger.Debug("Image validated", zap.String("format", string(format)))

	results := make(map[string][]byte)
	currentData := imageData

	// Выполняем операции последовательно
	for idx, opParams := range operations {
		p.logger.Debug("Executing operation",
			zap.Int("index", idx),
			zap.String("type", string(opParams.Type)),
		)

		operation, exists := p.operations[opParams.Type]
		if !exists {
			p.logger.Error("Unknown operation type", zap.String("type", string(opParams.Type)))
			return nil, fmt.Errorf("unknown operation type: %s", opParams.Type)
		}

		// Валидируем параметры операции
		if err := operation.Validate(opParams.Parameters); err != nil {
			p.logger.Error("Operation validation failed",
				zap.String("type", string(opParams.Type)),
				zap.Error(err),
			)
			return nil, fmt.Errorf("invalid parameters for operation %s: %w", opParams.Type, err)
		}

		// Выполняем операцию
		processedData, err := operation.Execute(currentData, opParams.Parameters)
		if err != nil {
			p.logger.Error("Operation execution failed",
				zap.String("type", string(opParams.Type)),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to execute operation %s: %w", opParams.Type, err)
		}

		// Сохраняем результат
		operationKey := string(opParams.Type)
		results[operationKey] = processedData
		currentData = processedData

		p.logger.Debug("Operation completed",
			zap.String("type", string(opParams.Type)),
			zap.Int("resultSize", len(processedData)),
		)
	}

	p.logger.Info("Image processing completed",
		zap.Int("operationCount", len(operations)),
		zap.Int("resultCount", len(results)),
	)

	return results, nil
}

// ValidateImage проверяет, является ли файл допустимым изображением
func (p *ImageProcessorImpl) ValidateImage(imageData []byte) (entity.ImageFormat, error) {
	if len(imageData) == 0 {
		return "", fmt.Errorf("empty image data")
	}

	// Пытаемся декодировать изображение
	_, format, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		// Пробуем WebP отдельно, так как он не включен в стандартную библиотеку
		_, err := webp.DecodeConfig(bytes.NewReader(imageData))
		if err == nil {
			return entity.FormatWebP, nil
		}
		return "", fmt.Errorf("unsupported image format: %w", err)
	}

	// Преобразуем формат в наш тип
	var imgFormat entity.ImageFormat
	switch format {
	case "jpeg":
		imgFormat = entity.FormatJPEG
	case "png":
		imgFormat = entity.FormatPNG
	case "gif":
		imgFormat = entity.FormatGIF
	case "webp":
		imgFormat = entity.FormatWebP
	default:
		return "", fmt.Errorf("unsupported image format: %s", format)
	}

	return imgFormat, nil
}

// GetImageInfo возвращает информацию об изображении
func (p *ImageProcessorImpl) GetImageInfo(imageData []byte) (*entity.ImageInfo, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	var imgFormat entity.ImageFormat
	switch format {
	case "jpeg":
		imgFormat = entity.FormatJPEG
	case "png":
		imgFormat = entity.FormatPNG
	case "gif":
		imgFormat = entity.FormatGIF
	default:
		imgFormat = entity.ImageFormat(format)
	}

	info := &entity.ImageInfo{
		Width:  config.Width,
		Height: config.Height,
		Format: imgFormat,
		Size:   int64(len(imageData)),
	}

	return info, nil
}
