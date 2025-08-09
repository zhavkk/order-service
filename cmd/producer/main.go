// Супер примитивный Kafka producer на Go
package main

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/internal/models"
)

func main() {
	brokers := []string{"localhost:9092"}
	topic := "orders"

	logger.Init("local")

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		logger.Log.Error("Failed to create Kafka producer", "error", err)
		return
	}
	defer producer.Close()

	logger.Log.Info("Kafka producer started", "brokers", brokers, "topic", topic)
	for i := 0; i < 10; i++ {
		order := generateRandomOrder()

		orderJSON, err := json.Marshal(order)
		if err != nil {
			logger.Log.Error("Failed to marshal order", "error", err)
			continue
		}
		message := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(orderJSON),
		}

		partition, offset, err := producer.SendMessage(message)
		if err != nil {
			logger.Log.Error("Failed to send message", "error", err)
		} else {
			logger.Log.Info("Message sent", "partition", partition, "offset", offset, "message", message.Value)
		}
		time.Sleep(1 * time.Second)

	}

	logger.Log.Info("Kafka producer finished sending messages")
}

func generateRandomOrder() models.Order {
	return models.Order{
		OrderUID:    uuid.NewString(),
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    "Test Delivery",
			Phone:   "+1234567890",
			Zip:     "123456",
			City:    "Test City",
			Address: "Test Address",
			Region:  "Test Region",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  uuid.NewString(),
			Currency:     "RUB",
			Provider:     "Test Provider",
			Amount:       1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "Test Bank",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      123456,
				TrackNumber: "WBILMTESTTRACK",
				Price:       1000,
				Rid:         uuid.NewString(),
				Name:        "Test Item",
				Sale:        0,
				Size:        "M",
				TotalPrice:  1000,
				NmId:        2389212,
				Brand:       "Test Brand",
				Status:      1,
			},
		},
		Locale:            "ru",
		InternalSignature: "Test Signature",
		CustomerID:        "test_customer",
		DeliveryService:   "Test Delivery Service",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}
