# Order Data Streaming Service

## Краткое описание

Этот проект представляет собой демонстрационный микросервис на **Go**, который реализует полный жизненный цикл обработки данных о заказах. Сервис слушает топик в **Apache Kafka**, получает данные о заказах, сохраняет их в базу данных **PostgreSQL** и кэширует в памяти для мгновенной отдачи через **REST API**. Проект также включает в себя простейший веб-интерфейс для отображения данных заказа по его ID.

## Демонстрация работы

https://drive.google.com/file/d/1VTngmC31q95rKGjbFTM7cesfgGEon93v/view?usp=sharing

*(**Примечание для тебя:** Запиши короткое видео работы, конвертируй его в GIF (например, через ezgif.com) и загрузи в репозиторий. Затем замени ссылку выше на реальную.)*

## Стек технологий

*   **Бэкенд:** Go
*   **База данных:** PostgreSQL
*   **Брокер сообщений:** Apache Kafka
*   **Контейнеризация:** Docker, Docker Compose
*   **Ключевые Go-библиотеки:**
    *   `pgx/v5` - драйвер для PostgreSQL
    *   `segmentio/kafka-go` - клиент для Kafka
    *   `joho/godotenv` - управление конфигурацией через `.env` файлы

## Требования

Перед началом работы убедитесь, что на вашей машине установлены:

*   [Go](https://go.dev/doc/install) (версия 1.20+)
*   [Docker](https://www.docker.com/products/docker-desktop/)
*   [Docker Compose](httpsf//docs.docker.com/compose/install/) (обычно идет в комплекте с Docker Desktop)

## Инструкция по установке и запуску

### 1. Клонирование репозитория

```bash
git clone https://github.com/your-username/order-service.git
cd order-service
```

### 2. Настройка конфигурации
Проект использует переменные окружения для конфигурации. Необходимо создать .env файл на основе примера.

```bash
cp .env.example .env
```
Файл .env уже содержит все необходимые настройки по умолчанию для запуска через Docker. Вносить изменения не требуется.

### 3. Запуск инфраструктуры
Запустите PostgreSQL и Kafka в Docker-контейнерах.

```bash
docker-compose up -d
```
Эта команда скачает необходимые образы и запустит сервисы в фоновом режиме.

### 4. Создание таблиц в базе данных
Выполните SQL-скрипт для создания необходимой таблицы в базе данных.

```bash
# Для macOS / Linux
cat schema.sql | docker exec -i order_db psql -U order_user -d orders_db
```

### 5. Установка зависимостей Go
Скачайте все необходимые Go-модули.

```bash
go mod tidy
```
### 6. Запуск основного сервиса
В первом окне терминала запустите основной Go-сервис. Он начнет слушать Kafka и обслуживать HTTP-запросы.

```bash
go run ./cmd/service/main.go
```

Вы должны увидеть логи об успешном подключении к БД, прогреве кэша и запуске сервера. Оставьте этот терминал работать.

### 7. Отправка тестовых данных
В новом (втором) окне терминала запустите скрипт-публикатор, чтобы отправить тестовый заказ в Kafka.

```bash
go run ./tools/publisher/main.go
```

В логах основного сервиса (в первом терминале) должно появиться сообщение об успешной обработке заказа.

### 8. Проверка через веб-интерфейс
Откройте браузер и перейдите по адресу http://localhost:8081.
На странице уже будет введен тестовый order_uid. Нажмите кнопку "Find", и вы увидите данные заказа в формате JSON.

Получить данные заказа по ID : 
```bash
curl http://localhost:8081/order/b563feb7b2b84b6test
```

## Структура Проекта

├── cmd/service/            # Точка входа в основное приложение (main.go)
├── internal/               # Внутренняя логика приложения, не предназначенная для импорта извне
│   ├── cache/              # Реализация кэша в памяти
│   ├── database/           # Логика работы с базой данных (репозиторий)
│   ├── kafka/              # Kafka-консьюмер
│   ├── models/             # Структуры данных (модель заказа)
│   └── server/             # HTTP-сервер и обработчики API
├── tools/publisher/        # Вспомогательный скрипт для отправки сообщений в Kafka
├── web/static/             # Файлы фронтенда (HTML, JS, CSS)
├── .env.example            # Шаблон файла конфигурации
├── .gitignore              # Файлы, игнорируемые Git
├── docker-compose.yml      # Конфигурация для запуска инфраструктуры
├── go.mod / go.sum         # Управление зависимостями Go
└── schema.sql              # SQL-схема для создания таблиц в БД

### Старый cmd/service/main.go: 


```go
SaverWithCache "декорирует" репозиторий, добавляя логику обновления кэша.
type SaverWithCache struct {
	DB    kafka.OrderSaver
	Cache interface {
		Set(order models.Order)
	}
}

// SaveOrder реализует интерфейс kafka.OrderSaver
func (s *SaverWithCache) SaveOrder(ctx context.Context, order models.Order) error {
	if err := s.DB.SaveOrder(ctx, order); err != nil {
		return err
	}
	s.Cache.Set(order)
	return nil
}

func main() {
	dbConnStr := "postgres://order_user:order_pass@localhost:5432/orders_db?sslmode=disable"
	kafkaBrokers := []string{"localhost:9092"}
	kafkaTopic := "orders"
	httpServerAddr := ":8081"

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbPool, err := database.ConnectDB(dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection successful.")

	orderRepo := database.NewOrderRepository(dbPool)
	memoryCache := cache.NewMemoryCache(orderRepo)
	if err := memoryCache.WarmUp(ctx); err != nil {
		log.Fatalf("Failed to warm up cache: %v", err)
	}

	saverWithCache := &SaverWithCache{
		DB:    orderRepo,
		Cache: memoryCache,
	}

	consumer := kafka.NewConsumer(kafkaBrokers, kafkaTopic, saverWithCache)
	go consumer.Start(ctx)

	apiServer := server.NewAPIServer(httpServerAddr, memoryCache)
	go func() {
		if err := apiServer.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received. Shutting down gracefully...")

	if err := consumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}
	log.Println("Service stopped.")
}
```
