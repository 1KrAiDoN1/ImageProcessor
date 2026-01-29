package dto

// ErrorResponse представляет стандартный ответ об ошибке
type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// SuccessResponse представляет стандартный успешный ответ
type SuccessResponse struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// HealthCheckResponse представляет ответ проверки здоровья сервиса
type HealthCheckResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Database  string `json:"database"`
	Timestamp int64  `json:"timestamp"`
}

// PaginationParams представляет параметры пагинации
type PaginationParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Sort     string `json:"sort,omitempty"`
	Order    string `json:"order,omitempty"` // asc или desc
}
