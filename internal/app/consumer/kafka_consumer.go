package consumer

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/pkg/utils"
)

type Consumer interface {
	Consume(ctx context.Context, topic string, handler func(message []byte) error) error
	Close() error
}

type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	handler       func(message []byte) error
	retryCount    int
	backoff       time.Duration
}

func NewKafkaConsumer(
	brokers []string,
	topic string,
	handler func(message []byte) error,
	cfg *sarama.Config,
	groupID string,
	retryCount int,
	backoff time.Duration,
) (*KafkaConsumer, error) {
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumerGroup: consumerGroup,
		topic:         topic,
		handler:       handler,
		retryCount:    retryCount,
		backoff:       backoff,
	}, nil
}

func (kc *KafkaConsumer) Consume(ctx context.Context) error {
	for {
		if err := kc.consumerGroup.Consume(ctx, []string{kc.topic}, kc); err != nil {
			logger.Log.Error("Error consuming messages", "error", err)
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

	}
}

func (kc *KafkaConsumer) Close() error {
	logger.Log.Info("Closing Kafka consumer")
	return kc.consumerGroup.Close()
}

func (kc *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (kc *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (kc *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		err := utils.RetryWithBackoff(func() error {
			return kc.handler(message.Value)
		}, kc.retryCount, kc.backoff)

		if err != nil {
			logger.Log.Error("Failed to handle message after retries", "error", err, "message", string(message.Value))
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
}
