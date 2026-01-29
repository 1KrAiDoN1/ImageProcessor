package kafka

import (
	"context"
	"fmt"
	"imageprocessor/backend/internal/config"
	"log/slog"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func EnsureTopicExists(ctx context.Context, cfg config.KafkaConfig, logger *slog.Logger) error {
	if len(cfg.Brokers) == 0 {
		return fmt.Errorf("no brokers configured")
	}

	conn, err := kafka.DialContext(ctx, "tcp", cfg.Brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("failed to read partitions: %w", err)
	}

	topicExists := false
	for _, p := range partitions {
		if p.Topic == cfg.Topic {
			topicExists = true
			break
		}
	}

	if topicExists {
		if logger != nil {
			logger.Info("Topic already exists", slog.String("topic", cfg.Topic))
		}
		return nil
	}

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             cfg.Topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}
