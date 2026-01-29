package kafka

import (
	"context"
	"fmt"
	"imageprocessor/backend/internal/config"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaProducer struct {
	Conn   *kafka.Conn
	writer *kafka.Writer
	config config.KafkaConfig
	logger *zap.Logger
}

func NewKafkaProducer(cfg config.KafkaConfig, log *zap.Logger) *KafkaProducer {
	var conn *kafka.Conn
	var err error

	maxRetries := 5
	initialBackoff := 2 * time.Second
	maxBackoff := 30 * time.Second

	currentBackoff := initialBackoff

	for attempt := 1; attempt <= maxRetries; attempt++ {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, err = kafka.DialLeader(ctx, "tcp", cfg.Brokers[0], cfg.Topic, 0)
		if err == nil {
			// Успех!
			log.Info("Успешно подключено к Kafka")
			conn.Close()
			break
		}

		log.Error("Не удалось подключиться к Kafka", zap.Int("attempt", attempt), zap.Error(err))
		if attempt < maxRetries {
			time.Sleep(currentBackoff)
			currentBackoff *= 2
			if currentBackoff > maxBackoff {
				currentBackoff = maxBackoff
			}
		}
	}

	if err != nil {
		log.Error("Не удалось подключиться к Kafka после нескольких попыток", zap.Error(err))
		return nil
	}

	return &KafkaProducer{
		Conn: conn,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        cfg.Topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Compression:  kafka.Snappy,
			BatchTimeout: 10 * time.Millisecond, // Быстрее отправка батчей
			WriteTimeout: 5 * time.Second,
		},
		config: cfg,
		logger: log,
	}
}

func (kp *KafkaProducer) SendMessage(ctx context.Context, msg kafka.Message) error {
	kp.logger.Info("Sending message to Kafka", zap.ByteString("key", msg.Key), zap.Int("value_size", len(msg.Value)))
	err := kp.writer.WriteMessages(ctx, msg)
	if err != nil {
		kp.logger.Error("Failed to send message", zap.Error(err))
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}
