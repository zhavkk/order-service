package consumer

import "context"

type KafkaConsumer interface {
	Consume(ctx context.Context, topic string, handler func(message []byte) error) error
	Close() error
}
