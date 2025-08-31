# Микросервис данных о заказах


## Краткое описание

Этот проект представляет собой демонстрационный микросервис на **Go**, который реализует полный жизненный цикл обработки данных о заказах. Сервис слушает топик в **Apache Kafka**, получает данные о заказах, сохраняет их в базу данных **PostgreSQL** и кэширует в памяти для мгновенной отдачи через **REST API**. Проект также включает в себя простейший веб-интерфейс для отображения данных заказа по его ID.

## Демонстрация работы

Google Disk:
https://drive.google.com/file/d/1VTngmC31q95rKGjbFTM7cesfgGEon93v/view?usp=sharing


## Стек технологий

*   **Бэкенд:** Go
*   **База данных:** PostgreSQL
*   **Брокер сообщений:** Apache Kafka
*   **Контейнеризация:** Docker, Docker Compose
*   **Ключевые Go-библиотеки:**
    *   `pgx/v5` - драйвер для PostgreSQL
    *   `segmentio/kafka-go` - клиент для Kafka
    *   `joho/godotenv` - управление конфигурацией через `.env` файлы


## Инструкция по установке и запуску

### 1. Клонирование репозитория

```bash
git clone https://github.com/your-username/order-service.git
cd order-service
```

### 2. Настройка конфигурации
Переменные окружения для конфигурации. Создание .env 

```bash
cp .env.example .env
```

### 3. Запуск инфраструктуры
Запуск PostgreSQL и Kafka в Docker контейнерах.

```bash
docker-compose up -d
```

### 4. Создание таблиц в базе данных
SQL-скрипт для создания таблицы.

```bash
# Для macOS / Linux
cat schema.sql | docker exec -i order_db psql -U order_user -d orders_db
```

### 5. Установка зависимостей Go
Go-модули.

```bash
go mod tidy
```

### 6. Запуск основного сервиса
В первом окне терминала основной Go-сервис. Он начнет слушать Kafka и обслуживать HTTP-запросы.

```bash
go run ./cmd/service/main.go
```

### 7. Отправка тестовых данных
В новом (втором) скрипт-публикатор, чтобы отправить тестовый заказ в Kafka.

```bash
go run ./tools/publisher/main.go
```

Получить данные заказа по ID : 
```bash
curl http://localhost:8081/order/b563feb7b2b84b6test
```