package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"imageprocessor/backend/internal/config"
	"imageprocessor/backend/internal/domain/entity"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	reader *kafka.Reader
	logger *zap.Logger
}

// NewConsumer создает нового Kafka consumer
func NewConsumer(cfg config.BrokerConfig, logger *zap.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           cfg.Brokers,
		Topic:             cfg.Topic,
		GroupID:           cfg.ConsumerGroup,
		MinBytes:          cfg.FetchMinBytes,
		MaxBytes:          cfg.FetchMaxBytes,
		CommitInterval:    cfg.CommitInterval,
		SessionTimeout:    cfg.SessionTimeout,
		HeartbeatInterval: cfg.HeartbeatInterval,
		RebalanceTimeout:  cfg.RebalanceTimeout,
		StartOffset:       kafka.FirstOffset, // Читать с начала для новой consumer group
		MaxWait:           500 * time.Millisecond,
	})

	logger.Info("Kafka consumer initialized",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("topic", cfg.Topic),
		zap.String("groupId", cfg.ConsumerGroup),
	)

	return &Consumer{
		reader: reader,
		logger: logger,
	}
}

// Start начинает чтение сообщений из Kafka
func (c *Consumer) Start(ctx context.Context, handler func(ctx context.Context, task *entity.ProcessingTask) error) error {
	c.logger.Info("Starting Kafka consumer")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer stopped by context")
			return ctx.Err()
		default:
			message, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					c.logger.Info("Consumer stopped")
					return nil
				}
				c.logger.Error("Failed to read message",
					zap.Error(err),
					zap.String("error_type", fmt.Sprintf("%T", err)),
				)
				continue
			}

			c.logger.Info("Message received",
				zap.String("topic", message.Topic),
				zap.Int("partition", message.Partition),
				zap.Int64("offset", message.Offset),
				zap.Int("value_size", len(message.Value)),
			)

			// Обрабатываем сообщение
			if err := c.processMessage(ctx, message, handler); err != nil {
				c.logger.Error("Failed to process message",
					zap.Error(err),
					zap.Int64("offset", message.Offset),
				)
				continue
			}
			// Коммитим только после успешной обработки
			if err := c.reader.CommitMessages(ctx, message); err != nil {
				c.logger.Error("Failed to commit message",
					zap.Error(err),
					zap.Int64("offset", message.Offset),
				)
			} else {
				c.logger.Info("Message processed and committed",
					zap.Int64("offset", message.Offset),
				)
			}
		}
	}
}

// processMessage обрабатывает одно сообщение
func (c *Consumer) processMessage(ctx context.Context, message kafka.Message, handler func(ctx context.Context, task *entity.ProcessingTask) error) error {
	// Десериализуем задачу
	var task entity.ProcessingTask
	if err := json.Unmarshal(message.Value, &task); err != nil {
		c.logger.Error("Failed to unmarshal task", zap.Error(err))
		return fmt.Errorf("failed to unmarshal task: %w", err)
	}

	c.logger.Info("Processing task",
		zap.String("taskId", task.ID),
		zap.String("imageId", task.ImageID),
		zap.Int("operationsCount", len(task.Operations)),
	)

	// Вызываем обработчик
	startTime := time.Now()
	err := handler(ctx, &task)
	duration := time.Since(startTime)

	if err != nil {
		c.logger.Error("Task processing failed",
			zap.Error(err),
			zap.String("taskId", task.ID),
			zap.Duration("duration", duration),
		)
		return err
	}

	c.logger.Info("Task processed successfully",
		zap.String("taskId", task.ID),
		zap.String("imageId", task.ImageID),
		zap.Duration("duration", duration),
	)

	return nil
}

// ReadBatch читает пакет сообщений
func (c *Consumer) ReadBatch(ctx context.Context, maxMessages int) ([]*entity.ProcessingTask, error) {
	tasks := make([]*entity.ProcessingTask, 0, maxMessages)

	for i := 0; i < maxMessages; i++ {
		// Используем контекст с таймаутом для неблокирующего чтения
		readCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		message, err := c.reader.FetchMessage(readCtx)
		cancel()

		if err != nil {
			if err == context.DeadlineExceeded {
				// Нет больше сообщений, возвращаем что есть
				break
			}
			c.logger.Error("Failed to fetch message", zap.Error(err))
			return tasks, err
		}

		// Десериализуем задачу
		var task entity.ProcessingTask
		if err := json.Unmarshal(message.Value, &task); err != nil {
			c.logger.Error("Failed to unmarshal task", zap.Error(err))
			continue
		}

		tasks = append(tasks, &task)

		// Коммитим сообщение
		if err := c.reader.CommitMessages(ctx, message); err != nil {
			c.logger.Error("Failed to commit message", zap.Error(err))
		}
	}

	return tasks, nil
}

// Close закрывает consumer
func (c *Consumer) Close() error {
	c.logger.Info("Closing Kafka consumer")
	if err := c.reader.Close(); err != nil {
		c.logger.Error("Failed to close Kafka consumer", zap.Error(err))
		return err
	}
	return nil
}

// Stats возвращает статистику consumer
func (c *Consumer) Stats() (int64, int64) {
	msg := c.reader.Stats().Messages
	byt := c.reader.Stats().FetchBytes.Sum
	return msg, byt
}
