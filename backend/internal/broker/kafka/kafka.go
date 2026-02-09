package kafka

import (
	"context"
	"fmt"
	"imageprocessor/backend/internal/config"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// EnsureTopicExists проверяет существование топика и создает его при необходимости
func EnsureTopicExists(cfg config.BrokerConfig, logger *zap.Logger) error {
	logger.Info("Checking if Kafka topic exists", zap.String("topic", cfg.Topic))

	conn, err := kafka.Dial("tcp", cfg.Brokers[0])
	if err != nil {
		logger.Error("Failed to connect to Kafka", zap.Error(err))
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			logger.Error("Failed to close Kafka connection", zap.Error(err))
		}
	}()
	controller, err := conn.Controller()
	if err != nil {
		logger.Error("Failed to get controller", zap.Error(err))
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		logger.Error("Failed to connect to controller", zap.Error(err))
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer func() {
		err = controllerConn.Close()
		if err != nil {
			logger.Error("Failed to close Kafka controller connection", zap.Error(err))
		}
	}()

	// Получаем список топиков
	partitions, err := controllerConn.ReadPartitions()
	if err != nil {
		logger.Error("Failed to read partitions", zap.Error(err))
		return fmt.Errorf("failed to read partitions: %w", err)
	}

	// Проверяем, существует ли топик
	topicExists := false
	for _, partition := range partitions {
		if partition.Topic == cfg.Topic {
			topicExists = true
			break
		}
	}

	if topicExists {
		logger.Info("Kafka topic already exists", zap.String("topic", cfg.Topic))
		return nil
	}

	// Создаем топик
	logger.Info("Creating Kafka topic", zap.String("topic", cfg.Topic))

	topicConfig := kafka.TopicConfig{
		Topic:             cfg.Topic,
		NumPartitions:     3,
		ReplicationFactor: 1,
		ConfigEntries: []kafka.ConfigEntry{
			{
				ConfigName:  "retention.ms",
				ConfigValue: "604800000", // 7 дней
			},
			{
				ConfigName:  "compression.type",
				ConfigValue: "snappy",
			},
		},
	}

	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		logger.Error("Failed to create topic", zap.Error(err), zap.String("topic", cfg.Topic))
		return fmt.Errorf("failed to create topic: %w", err)
	}

	logger.Info("Kafka topic created successfully", zap.String("topic", cfg.Topic))
	return nil
}

// CheckConnection проверяет подключение к Kafka
func CheckConnection(brokers []string, logger *zap.Logger) error {
	logger.Info("Checking Kafka connection", zap.Strings("brokers", brokers))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, broker := range brokers {
		conn, err := kafka.DialContext(ctx, "tcp", broker)
		if err != nil {
			logger.Error("Failed to connect to broker",
				zap.Error(err),
				zap.String("broker", broker),
			)
			return fmt.Errorf("failed to connect to broker %s: %w", broker, err)
		}
		defer func() {
			err = conn.Close()
			if err != nil {
				logger.Error("Failed to close Kafka connection", zap.Error(err))
			}
		}()

		// Пытаемся получить метаданные
		_, err = conn.ApiVersions()
		if err != nil {
			logger.Error("Failed to get API versions",
				zap.Error(err),
				zap.String("broker", broker),
			)
			return fmt.Errorf("failed to get API versions from %s: %w", broker, err)
		}

		logger.Info("Successfully connected to broker", zap.String("broker", broker))
	}

	logger.Info("Kafka connection check passed")
	return nil
}

// GetTopicMetadata возвращает метаданные топика
func GetTopicMetadata(brokers []string, topic string, logger *zap.Logger) (*kafka.Topic, error) {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			logger.Error("Failed to close Kafka connection", zap.Error(err))
		}
	}()

	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	if len(partitions) == 0 {
		return nil, fmt.Errorf("topic %s has no partitions", topic)
	}

	topicMetadata := &kafka.Topic{
		Name:       topic,
		Partitions: make([]kafka.Partition, len(partitions)),
	}

	for i, partition := range partitions {
		topicMetadata.Partitions[i] = kafka.Partition{
			Topic:    partition.Topic,
			ID:       partition.ID,
			Leader:   partition.Leader,
			Replicas: partition.Replicas,
			Isr:      partition.Isr,
		}
	}

	return topicMetadata, nil
}
