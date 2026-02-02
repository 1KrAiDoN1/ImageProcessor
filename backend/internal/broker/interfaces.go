package broker

import (
	"context"
	"imageprocessor/backend/internal/domain/entity"
)

type ConsumerMessageBrokerInterface interface {
	Start(ctx context.Context, handler func(ctx context.Context, task *entity.ProcessingTask) error) error
	ReadBatch(ctx context.Context, maxMessages int) ([]*entity.ProcessingTask, error)
	Stats() (int64, int64)
	Close() error
}
