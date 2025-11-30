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
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("FATAL: could not load config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("Initializing application layers...")

	//слой данных
	dbPool, err := database.ConnectDB(cfg.dbConnStr)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection successful.")

	orderRepo := database.NewOrderRepository(dbPool)

	//слой кеша
	lruCache := cache.NewLRUCache(cfg.cacheCapacity, orderRepo)
	if err := lruCache.WarmUp(ctx); err != nil {
		log.Fatalf("FATAL: Failed to warm up cache: %v", err)
	}
	log.Println("Cache warmed up successfully.")

	orderService := service.NewOrderService(orderRepo, lruCache)
	log.Println("Service layer initialized.")

	kafkaConsumer := kafka.NewConsumer(cfg.kafkaBrokers, "orders", orderService)
	httpServer := server.NewAPIServer(cfg.httpServerAddr, orderService)
	log.Println("Delivery layer initialized.")

	//горутины в фоне
	go kafkaConsumer.Start(ctx)

	go func() {
		if err := httpServer.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("FATAL: HTTP server error: %v", err)
		}
	}()

	log.Println("Application started. Press Ctrl+C to exit.")
	<-ctx.Done()

	log.Println("Shutdown signal received. Shutting down gracefully...")

	if err := kafkaConsumer.Close(); err != nil {
		log.Printf("ERROR: Error closing Kafka consumer: %v", err)
	}

	log.Println("Service stopped.")
}

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
