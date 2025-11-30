package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/go-playground/validator/v10"

	"github.com/Evgeniy-Programming/golang/internal/domain"
	"github.com/segmentio/kafka-go"
)

// Consumer - наша новая структура консьюмера.
// Он отвечает за чтение сообщений из Kafka и их передачу в слой бизнес-логики.
type Consumer struct {
	reader   *kafka.Reader
	service  domain.OrderService // <-- ЗАВИСИМОСТЬ ОТ ИНТЕРФЕЙСА СЕРВИСА
	validate *validator.Validate // <-- Экземпляр валидатора
}

// NewConsumer создает новый консьюмер.
// Обрати внимание, он теперь принимает domain.OrderService.
func NewConsumer(brokers []string, topic string, service domain.OrderService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "order-service-group", // Группа консьюмеров для балансировки нагрузки
	})

	return &Consumer{
		reader:   reader,
		service:  service,
		validate: validator.New(), // Создаем новый экземпляр валидатора
	}
}

// Start запускает бесконечный цикл чтения сообщений из Kafka.
func (c *Consumer) Start(ctx context.Context) {
	log.Println("Starting Kafka consumer...")
	for {
		// Используем ReadMessage с контекстом для Graceful Shutdown.
		// Этот вызов блокирующий, он будет ждать нового сообщения.
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			// Если контекст отменен (приложение закрывается), выходим из цикла.
			if ctx.Err() != nil {
				log.Println("Kafka consumer stopping due to context cancellation.")
				break
			}
			log.Printf("ERROR: could not read message from Kafka: %v", err)
			continue
		}

		// Обрабатываем каждое сообщение.
		// Для повышения пропускной способности можно было бы запускать
		// c.processMessage в отдельной горутине, но для простоты оставим так.
		c.processMessage(ctx, msg)
	}
}

// processMessage инкапсулирует логику обработки одного сообщения.
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	// Шаг 1: Десериализация JSON в нашу доменную модель.
	var order domain.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		log.Printf("ERROR: failed to unmarshal order JSON: %v. Raw message: %s", err, string(msg.Value))
		// Это "битое" сообщение, мы его просто пропускаем, т.к. исправить его не можем.
		return
	}

	// Шаг 2: Валидация данных с помощью тегов в структуре.
	if err := c.validate.Struct(order); err != nil {
		log.Printf("ERROR: order data validation failed for order UID %s: %v", order.OrderUID, err)
		// Данные невалидны (например, нет обязательного поля), пропускаем сообщение.
		return
	}

	// Шаг 3: Передача валидного заказа в слой бизнес-логики (сервис).
	// Консьюмер больше не знает, что происходит с заказом дальше (сохранение, кэширование).
	// Он просто выполняет свою единственную задачу — доставку и первичную проверку.
	if err := c.service.CreateOrder(ctx, order); err != nil {
		log.Printf("ERROR: failed to process order %s: %v", order.OrderUID, err)
		// Здесь может быть логика повторной обработки (retry) или отправки
		// в "очередь мертвых писем" (Dead Letter Queue) для последующего анализа.
		return
	}

	log.Printf("Successfully processed and forwarded order to service: %s", order.OrderUID)
}

// Close корректно закрывает соединение с Kafka.
func (c *Consumer) Close() error {
	log.Println("Closing Kafka consumer reader...")
	return c.reader.Close()
}
