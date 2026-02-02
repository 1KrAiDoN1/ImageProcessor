package image_processor

import (
	"imageprocessor/backend/internal/domain/entity"
)

// Operation определяет интерфейс для операции обработки изображения
type Operation interface {
	// Execute выполняет операцию над изображением
	Execute(imageData []byte, params map[string]interface{}) ([]byte, error)

	// GetOperationType возвращает тип операции
	GetOperationType() entity.OperationType

	// Validate проверяет параметры операции
	Validate(params map[string]interface{}) error
}
