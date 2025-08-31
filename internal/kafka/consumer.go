package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Evgeniy-Programming/golang/internal/models"
	"github.com/segmentio/kafka-go"
)

type OrderSaver interface {
	SaveOrder(ctx context.Context, order models.Order) error
}

type Consumer struct {
	reader *kafka.Reader
	saver  OrderSaver
}

func NewConsumer(brokers []string, topic string, saver OrderSaver) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "order-service-group",
	})
	return &Consumer{reader: reader, saver: saver}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Println("Starting Kafka consumer...")
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("Error reading message from Kafka: %v", err)
			continue
		}
		var order models.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("Failed to unmarshal order: %v. Raw message: %s", err, string(msg.Value))
			continue
		}
		if err := c.saver.SaveOrder(ctx, order); err != nil {
			log.Printf("Failed to save order %s to DB: %v", order.OrderUID, err)
			continue
		}
		log.Printf("Successfully processed and saved order: %s", order.OrderUID)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
