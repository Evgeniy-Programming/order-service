package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/Evgeniy-Programming/golang/internal/cache"
	"github.com/Evgeniy-Programming/golang/internal/database"
	"github.com/Evgeniy-Programming/golang/internal/kafka"
	"github.com/Evgeniy-Programming/golang/internal/server"
	"github.com/Evgeniy-Programming/golang/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	// --- 1. Загрузка конфигурации ---
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("FATAL: could not load config: %v", err)
	}

	// --- 2. Настройка Graceful Shutdown ---
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// --- 3. Инициализация слоев (Dependency Injection) ---
	// Этот блок теперь называется "Композиция приложения" (Application Composition)
	log.Println("Initializing application layers...")

	// Слой данных (Data Layer)
	dbPool, err := database.ConnectDB(cfg.dbConnStr)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection successful.")

	orderRepo := database.NewOrderRepository(dbPool)

	// Слой кэша (Cache Layer)
	lruCache := cache.NewLRUCache(cfg.cacheCapacity, orderRepo)
	if err := lruCache.WarmUp(ctx); err != nil {
		log.Fatalf("FATAL: Failed to warm up cache: %v", err)
	}
	log.Println("Cache warmed up successfully.")

	// Слой бизнес-логики (Service Layer)
	// Создаем сервис, "внедряя" в него зависимости в виде интерфейсов.
	// orderRepo реализует domain.OrderRepository, а lruCache - domain.OrderCache.
	orderService := service.NewOrderService(orderRepo, lruCache)
	log.Println("Service layer initialized.")

	// Слой доставки (Delivery Layer)
	// Создаем компоненты, которые будут взаимодействовать с внешним миром.
	// Они зависят только от слоя сервиса.
	kafkaConsumer := kafka.NewConsumer(cfg.kafkaBrokers, "orders", orderService)
	httpServer := server.NewAPIServer(cfg.httpServerAddr, orderService)
	log.Println("Delivery layer initialized.")

	// --- 4. Запуск фоновых процессов ---
	go kafkaConsumer.Start(ctx)

	go func() {
		if err := httpServer.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("FATAL: HTTP server error: %v", err)
		}
	}()

	// --- 5. Ожидание сигнала завершения ---
	log.Println("Application started. Press Ctrl+C to exit.")
	<-ctx.Done()

	// --- 6. Корректное завершение работы ---
	log.Println("Shutdown signal received. Shutting down gracefully...")

	if err := kafkaConsumer.Close(); err != nil {
		log.Printf("ERROR: Error closing Kafka consumer: %v", err)
	}

	// Здесь можно добавить graceful shutdown и для HTTP-сервера, но для простоты опустим.

	log.Println("Service stopped.")
}

// --- Вспомогательная структура и функция для конфигурации ---

type config struct {
	dbConnStr      string
	kafkaBrokers   []string
	httpServerAddr string
	cacheCapacity  int
}

func loadConfig() (*config, error) {
	dbConnStr := os.Getenv("DB_CONN_STR")
	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	httpServerAddr := os.Getenv("HTTP_SERVER_ADDR")
	cacheCapacityStr := os.Getenv("CACHE_CAPACITY")

	if dbConnStr == "" {
		dbConnStr = "postgres://order_user:order_pass@localhost:5432/orders_db?sslmode=disable"
	}
	if kafkaBrokersStr == "" {
		kafkaBrokersStr = "127.0.0.1:9092"
	}
	if httpServerAddr == "" {
		httpServerAddr = ":8081"
	}

	cacheCapacity, err := strconv.Atoi(cacheCapacityStr)
	if err != nil || cacheCapacity <= 0 {
		cacheCapacity = 1000
	}

	return &config{
		dbConnStr:      dbConnStr,
		kafkaBrokers:   strings.Split(kafkaBrokersStr, ","),
		httpServerAddr: httpServerAddr,
		cacheCapacity:  cacheCapacity,
	}, nil
}
