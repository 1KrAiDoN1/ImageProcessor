package entity

import (
	"time"
)

// ProcessingStatus представляет статус обработки
type ProcessingStatus string

const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusCompleted  ProcessingStatus = "completed"
	StatusFailed     ProcessingStatus = "failed"
	StatusCancelled  ProcessingStatus = "cancelled"
)

// ProcessingJob представляет задачу обработки изображения
type ProcessingJob struct {
	ID            string
	ImageID       string
	OperationType string
	Status        ProcessingStatus
	Parameters    map[string]interface{}
	ResultPath    string
	StartedAt     time.Time
	CompletedAt   time.Time
	ErrorMessage  string
	RetryCount    int
	MaxRetries    int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
