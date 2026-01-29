package dto

import "time"

// UploadImageRequest представляет запрос на загрузку изображения
type UploadImageRequest struct {
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

// UploadImageResponse представляет ответ на загрузку изображения
type UploadImageResponse struct {
	ID           string                 `json:"id"`
	OriginalName string                 `json:"original_name"`
	Size         int64                  `json:"size"`
	Width        int                    `json:"width"`
	Height       int                    `json:"height"`
	Format       string                 `json:"format"`
	ContentType  string                 `json:"content_type"`
	UploadedAt   time.Time              `json:"uploaded_at"`
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// GetImageResponse представляет ответ с информацией об изображении
type GetImageResponse struct {
	ID             string                 `json:"id"`
	OriginalName   string                 `json:"original_name"`
	Size           int64                  `json:"size"`
	Width          int                    `json:"width"`
	Height         int                    `json:"height"`
	Format         string                 `json:"format"`
	ContentType    string                 `json:"content_type"`
	UploadedAt     time.Time              `json:"uploaded_at"`
	LastModifiedAt time.Time              `json:"last_modified_at"`
	IsProcessed    bool                   `json:"is_processed"`
	Tags           []string               `json:"tags,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// DeleteImageResponse представляет ответ на удаление изображения
type DeleteImageResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// ImageListResponse представляет список изображений
type ImageListResponse struct {
	Images     []GetImageResponse `json:"images"`
	TotalCount int64              `json:"total_count"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
}
