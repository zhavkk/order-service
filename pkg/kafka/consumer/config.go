package kafkapkg

import (
	"github.com/IBM/sarama"
	"github.com/zhavkk/order-service/internal/config"
)

func NewSaramaConfig(conf *config.Config) (*sarama.Config, error) {
	cfg := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion(conf.Kafka.Version)
	if err != nil {
		return nil, err
	}
	cfg.Version = version

	cfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Offsets.AutoCommit.Enable = true
	cfg.Consumer.Offsets.AutoCommit.Interval = conf.Kafka.AutoCommitInterval

	return cfg, nil
}
