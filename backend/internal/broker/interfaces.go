package broker

import "context"

type ConsumerMessageBrokerInterface interface {

	// Subscribe подписывается на топик и обрабатывает сообщения
	Subscribe(ctx context.Context, topic string) error

	// Close закрывает соединение с очередью
	Close() error
}

type ProducerMessageBrokerInterface interface {
	// Publish публикует сообщение в очередь
	Publish(ctx context.Context, topic string, message []byte) error

	// Close закрывает соединение с очередью
	Close() error
}
