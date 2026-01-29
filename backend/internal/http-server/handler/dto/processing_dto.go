package dto

import "time"

// ProcessImageRequest представляет запрос на обработку изображения
type ProcessImageRequest struct {
	Operations []ProcessingOperationDTO `json:"operations"`
}

// ProcessingOperationDTO представляет операцию обработки в DTO
type ProcessingOperationDTO struct {
	Type       string                 `json:"type"` // resize, thumbnail, watermark, etc.
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority,omitempty"`
}

// ProcessImageResponse представляет ответ на запрос обработки
type ProcessImageResponse struct {
	JobID     string    `json:"job_id"`
	ImageID   string    `json:"image_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ProcessingJobResponse представляет информацию о задаче обработки
type ProcessingJobResponse struct {
	ID            string                 `json:"id"`
	ImageID       string                 `json:"image_id"`
	OperationType string                 `json:"operation_type"`
	Status        string                 `json:"status"`
	ResultPath    string                 `json:"result_path,omitempty"`
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
	StartedAt     time.Time              `json:"started_at,omitempty"`
	CompletedAt   time.Time              `json:"completed_at,omitempty"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	RetryCount    int                    `json:"retry_count"`
	MaxRetries    int                    `json:"max_retries"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ProcessingJobListResponse представляет список задач обработки
type ProcessingJobListResponse struct {
	Jobs       []ProcessingJobResponse `json:"jobs"`
	TotalCount int64                   `json:"total_count"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
}

// CancelProcessingJobRequest представляет запрос на отмену обработки
type CancelProcessingJobRequest struct {
	Reason string `json:"reason,omitempty"`
}

// CancelProcessingJobResponse представляет ответ на отмену обработки
type CancelProcessingJobResponse struct {
	JobID       string    `json:"job_id"`
	Status      string    `json:"status"`
	CancelledAt time.Time `json:"cancelled_at"`
}
