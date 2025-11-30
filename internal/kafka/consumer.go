package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/go-playground/validator/v10"

	"github.com/Evgeniy-Programming/golang/internal/domain"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader   *kafka.Reader
	service  domain.OrderService
	validate *validator.Validate
}

func NewConsumer(brokers []string, topic string, service domain.OrderService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "order-service-group",
	})

	return &Consumer{
		reader:   reader,
		service:  service,
		validate: validator.New(),
	}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Println("Starting Kafka consumer...")
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			//если отменен - выходим
			if ctx.Err() != nil {
				log.Println("Kafka consumer stopping due to context cancellation.")
				break
			}
			log.Printf("ERROR: could not read message from Kafka: %v", err)
			continue
		}
		c.processMessage(ctx, msg) //отдельно в каждой горутине
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	var order domain.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		log.Printf("ERROR: failed to unmarshal order JSON: %v. Raw message: %s", err, string(msg.Value))
		return
	}

	if err := c.validate.Struct(order); err != nil {
		//невалидные данные
		log.Printf("ERROR: order data validation failed for order UID %s: %v", order.OrderUID, err)
		return
	}
	if err := c.service.CreateOrder(ctx, order); err != nil {
		log.Printf("ERROR: failed to process order %s: %v", order.OrderUID, err)
		return
	}

	log.Printf("Successfully processed and forwarded order to service: %s", order.OrderUID)
}

// закрывает соединение с кафкой
func (c *Consumer) Close() error {
	log.Println("Closing Kafka consumer reader...")
	return c.reader.Close()
}
