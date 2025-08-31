package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Evgeniy-Programming/golang/internal/cache"
	"github.com/Evgeniy-Programming/golang/internal/database"
	"github.com/Evgeniy-Programming/golang/internal/kafka"
	"github.com/Evgeniy-Programming/golang/internal/models"
	"github.com/Evgeniy-Programming/golang/internal/server"
	"github.com/joho/godotenv"
)

// SaverWithCache "декорирует" репозиторий, добавляя логику обновления кэша.
// type SaverWithCache struct {
// 	DB    kafka.OrderSaver
// 	Cache interface {
// 		Set(order models.Order)
// 	}
// }

// // SaveOrder реализует интерфейс kafka.OrderSaver
// func (s *SaverWithCache) SaveOrder(ctx context.Context, order models.Order) error {
// 	if err := s.DB.SaveOrder(ctx, order); err != nil {
// 		return err
// 	}
// 	s.Cache.Set(order)
// 	return nil
// }

// func main() {
// 	dbConnStr := "postgres://order_user:order_pass@localhost:5432/orders_db?sslmode=disable"
// 	kafkaBrokers := []string{"localhost:9092"}
// 	kafkaTopic := "orders"
// 	httpServerAddr := ":8081"

// 	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
// 	defer cancel()

// 	dbPool, err := database.ConnectDB(dbConnStr)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}
// 	defer dbPool.Close()
// 	log.Println("Database connection successful.")

// 	orderRepo := database.NewOrderRepository(dbPool)
// 	memoryCache := cache.NewMemoryCache(orderRepo)
// 	if err := memoryCache.WarmUp(ctx); err != nil {
// 		log.Fatalf("Failed to warm up cache: %v", err)
// 	}

// 	saverWithCache := &SaverWithCache{
// 		DB:    orderRepo,
// 		Cache: memoryCache,
// 	}

// 	consumer := kafka.NewConsumer(kafkaBrokers, kafkaTopic, saverWithCache)
// 	go consumer.Start(ctx)

// 	apiServer := server.NewAPIServer(httpServerAddr, memoryCache)
// 	go func() {
// 		if err := apiServer.Start(); err != nil && err.Error() != "http: Server closed" {
// 			log.Fatalf("HTTP server error: %v", err)
// 		}
// 	}()

// 	<-ctx.Done()
// 	log.Println("Shutdown signal received. Shutting down gracefully...")

// 	if err := consumer.Close(); err != nil {
// 		log.Printf("Error closing Kafka consumer: %v", err)
// 	}
// 	log.Println("Service stopped.")
// }

// SaverWithCache "декорирует" репозиторий, добавляя логику обновления кэша.
// Этот тип реализует интерфейс kafka.OrderSaver.
type SaverWithCache struct {
	DB    kafka.OrderSaver
	Cache interface {
		Set(order models.Order)
	}
}

// SaveOrder сначала сохраняет заказ в БД, а затем, в случае успеха, обновляет кэш.
func (s *SaverWithCache) SaveOrder(ctx context.Context, order models.Order) error {
	if err := s.DB.SaveOrder(ctx, order); err != nil {
		return err // Если не удалось сохранить в БД, возвращаем ошибку
	}
	// Если сохранение в БД прошло успешно, обновляем запись в кэше
	s.Cache.Set(order)
	return nil
}

func main() {
	// --- 1. Загрузка конфигурации ---
	// Загружаем переменные из .env файла в окружение.
	// Эта функция ищет файл .env в текущей директории и загружает его содержимое.
	// Если файл не найден, ошибки не будет - это позволяет приложению работать
	// в окружениях (например, в Docker или Kubernetes), где переменные задаются напрямую.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Читаем переменные из окружения. os.Getenv вернет пустую строку, если переменная не найдена.
	dbConnStr := os.Getenv("DB_CONN_STR")
	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	httpServerAddr := os.Getenv("HTTP_SERVER_ADDR")

	// Устанавливаем значения по умолчанию. Это делает приложение "самодостаточным"
	// и позволяет запустить его "из коробки" без предварительной настройки .env файла.
	if dbConnStr == "" {
		dbConnStr = "postgres://order_user:order_pass@localhost:5432/orders_db?sslmode=disable"
	}
	if kafkaBrokersStr == "" {
		kafkaBrokersStr = "127.0.0.1:9092"
	}
	if httpServerAddr == "" {
		httpServerAddr = ":8081"
	}

	// Kafka-клиент ожидает слайс строк (для подключения к нескольким брокерам),
	// а переменная окружения - это одна строка. Разбиваем ее по запятой.
	kafkaBrokers := strings.Split(kafkaBrokersStr, ",")

	// --- 2. Настройка Graceful Shutdown ---
	// Создаем контекст, который будет отменен при получении сигнала SIGINT (Ctrl+C) или SIGTERM.
	// Это позволит нашему приложению корректно завершить работу.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// --- 3. Инициализация зависимостей ---

	// Подключение к базе данных
	dbPool, err := database.ConnectDB(dbConnStr)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	defer dbPool.Close() // defer гарантирует, что соединение закроется при выходе из main
	log.Println("Database connection successful.")

	// Создание репозитория и кэша
	orderRepo := database.NewOrderRepository(dbPool)
	memoryCache := cache.NewMemoryCache(orderRepo)

	// "Прогрев" кэша: загружаем все существующие заказы из БД в память при старте
	if err := memoryCache.WarmUp(ctx); err != nil {
		log.Fatalf("FATAL: Failed to warm up cache: %v", err)
	}

	// Создаем "декоратор", который будет и сохранять в БД, и обновлять кэш
	saverWithCache := &SaverWithCache{
		DB:    orderRepo,
		Cache: memoryCache,
	}

	// --- 4. Запуск основных компонентов в горутинах ---

	// Запуск Kafka-консьюмера в отдельной горутине
	consumer := kafka.NewConsumer(kafkaBrokers, "orders", saverWithCache)
	go consumer.Start(ctx)

	// Запуск HTTP-сервера в отдельной горутине
	apiServer := server.NewAPIServer(httpServerAddr, memoryCache)
	go func() {
		if err := apiServer.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("FATAL: HTTP server error: %v", err)
		}
	}()

	// --- 5. Ожидание сигнала завершения ---
	// Этот код заблокирует выполнение до тех пор, пока контекст `ctx` не будет отменен (т.е. до Ctrl+C)
	<-ctx.Done()

	log.Println("Shutdown signal received. Shutting down gracefully...")

	// Корректно закрываем соединение с Kafka
	if err := consumer.Close(); err != nil {
		log.Printf("ERROR: Error closing Kafka consumer: %v", err)
	}

	log.Println("Service stopped.")
}
