package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"imageprocessor/backend/internal/config"
	"imageprocessor/backend/internal/domain/entity"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
	topic  string
}

// NewProducer создает нового Kafka producer
func NewProducer(cfg config.BrokerConfig, logger *zap.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		Compression:  kafka.Snappy,
	}

	logger.Info("Kafka producer initialized",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("topic", cfg.Topic),
	)

	return &Producer{
		writer: writer,
		logger: logger,
		topic:  cfg.Topic,
	}
}

// PublishProcessingTask публикует задачу на обработку изображения в Kafka
func (p *Producer) PublishProcessingTask(ctx context.Context, task *entity.ProcessingTask) error {
	p.logger.Debug("Publishing processing task",
		zap.String("taskId", task.ID),
		zap.String("imageId", task.ImageID),
		zap.Int("operationsCount", len(task.Operations)),
	)

	// Сериализуем задачу в JSON
	taskJSON, err := json.Marshal(task)
	if err != nil {
		p.logger.Error("Failed to marshal task", zap.Error(err), zap.String("taskId", task.ID))
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Создаем сообщение
	message := kafka.Message{
		Key:   []byte(task.ImageID),
		Value: taskJSON,
		Headers: []kafka.Header{
			{Key: "task-id", Value: []byte(task.ID)},
			{Key: "image-id", Value: []byte(task.ImageID)},
		},
	}

	// Отправляем сообщение
	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		p.logger.Error("Failed to publish message",
			zap.Error(err),
			zap.String("taskId", task.ID),
			zap.String("imageId", task.ImageID),
		)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Info("Processing task published successfully",
		zap.String("taskId", task.ID),
		zap.String("imageId", task.ImageID),
		zap.Int("operationsCount", len(task.Operations)),
	)

	return nil
}

// PublishBatch публикует несколько задач одновременно
func (p *Producer) PublishBatch(ctx context.Context, tasks []*entity.ProcessingTask) error {
	if len(tasks) == 0 {
		return nil
	}

	p.logger.Debug("Publishing batch of tasks", zap.Int("count", len(tasks)))

	messages := make([]kafka.Message, 0, len(tasks))

	for _, task := range tasks {
		taskJSON, err := json.Marshal(task)
		if err != nil {
			p.logger.Error("Failed to marshal task", zap.Error(err), zap.String("taskId", task.ID))
			continue
		}

		message := kafka.Message{
			Key:   []byte(task.ImageID),
			Value: taskJSON,
			Headers: []kafka.Header{
				{Key: "task-id", Value: []byte(task.ID)},
				{Key: "image-id", Value: []byte(task.ImageID)},
			},
		}

		messages = append(messages, message)
	}

	err := p.writer.WriteMessages(ctx, messages...)
	if err != nil {
		p.logger.Error("Failed to publish batch", zap.Error(err), zap.Int("count", len(messages)))
		return fmt.Errorf("failed to publish batch: %w", err)
	}

	p.logger.Info("Batch published successfully", zap.Int("count", len(messages)))
	return nil
}

// Close закрывает producer
func (p *Producer) Close() error {
	p.logger.Info("Closing Kafka producer")
	if err := p.writer.Close(); err != nil {
		p.logger.Error("Failed to close Kafka producer", zap.Error(err))
		return err
	}
	return nil
}
