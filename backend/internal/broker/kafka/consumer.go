package kafka

import (
	"context"
	"imageprocessor/backend/internal/config"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaConsumer struct {
	Conn   *kafka.Conn
	reader *kafka.Reader
	log    *zap.Logger
}

func NewKafkaConsumer(log *zap.Logger, cfg config.KafkaConfig) *KafkaConsumer {
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
			log.Info("Успешно подключено к Kafka")
			conn.Close()
			break
		}

		if attempt < maxRetries {
			log.Warn("Не удалось подключиться к Kafka, повторная попытка...", zap.Int("attempt", attempt), zap.Error(err))
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

	return &KafkaConsumer{
		Conn: conn,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:          cfg.Brokers,
			Topic:            cfg.Topic,
			MinBytes:         1,
			MaxBytes:         10e4,
			CommitInterval:   time.Second,
			StartOffset:      kafka.LastOffset,
			ReadBatchTimeout: 100 * time.Millisecond,
		}),
		log: log,
	}
}

func (kc *KafkaConsumer) ReadMessage(ctx context.Context) (*kafka.Message, error) {
	kafkaMessage, err := kc.reader.ReadMessage(ctx)
	if err != nil {
		kc.log.Error("Failed to read message", zap.String("status", "error"), zap.Error(err))
		return nil, err
	}
	return &kafkaMessage, nil
}

func (kc *KafkaConsumer) CommitMessage(ctx context.Context, msg *kafka.Message) error {
	return kc.reader.CommitMessages(ctx, *msg)
}

func (kc *KafkaConsumer) Close() error {
	return kc.reader.Close()
}
