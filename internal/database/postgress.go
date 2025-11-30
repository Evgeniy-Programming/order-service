package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Evgeniy-Programming/golang/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderRepository реализует логику взаимодействия с хранилищем заказов в PostgreSQL.
type OrderRepository struct {
	db *pgxpool.Pool
}

// NewOrderRepository создает новый экземпляр репозитория.
func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// ConnectDB по-прежнему отвечает за подключение к БД. Здесь изменений нет.
func ConnectDB(connStr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return pool, nil
}

// Save сохраняет заказ в базу данных, используя транзакцию.
// Этот метод теперь соответствует нашему будущему интерфейсу.
func (r *OrderRepository) Save(ctx context.Context, order domain.Order) error {
	// 1. Начинаем транзакцию.
	// Все последующие операции будут выполняться в рамках этой транзакции.
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// 2. defer tx.Rollback() - это гарантия безопасности.
	// Если функция завершится с ошибкой в любом месте (например, после первого Exec),
	// эта строка автоматически откатит все изменения.
	// Если же мы успешно вызовем tx.Commit(), то Rollback уже ничего не сделает.
	defer tx.Rollback(ctx)

	// 3. Подготавливаем данные для вставки, теперь с обработкой ошибок.
	deliveryJSON, err := json.Marshal(order.Delivery)
	if err != nil {
		return fmt.Errorf("failed to marshal delivery data: %w", err)
	}
	paymentJSON, err := json.Marshal(order.Payment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment data: %w", err)
	}
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal items data: %w", err)
	}

	// 4. Выполняем SQL-запрос, но теперь используя объект транзакции `tx`.
	query := `
        INSERT INTO orders (order_uid, track_number, entry, delivery, payment, items, locale, customer_id, date_created)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (order_uid) DO NOTHING;`

	// В будущем, если появятся другие таблицы (например, order_items),
	// мы добавим сюда новые вызовы tx.Exec(...) для вставки в них.

	if _, err := tx.Exec(ctx, query,
		order.OrderUID, order.TrackNumber, order.Entry, deliveryJSON, paymentJSON, itemsJSON,
		order.Locale, order.CustomerID, order.DateCreated,
	); err != nil {
		return fmt.Errorf("failed to execute insert query: %w", err)
	}

	// 5. Если все запросы выше прошли без ошибок, мы "фиксируем" транзакцию.
	// Только после этого вызова изменения станут видимы для других.
	return tx.Commit(ctx)
}

// GetAll получает все заказы из базы данных для "прогрева" кэша.
// Этот метод тоже переименован для соответствия интерфейсу.
func (r *OrderRepository) GetAll(ctx context.Context) ([]domain.Order, error) {
	query := `SELECT order_uid, track_number, entry, delivery, payment, items, locale, customer_id, date_created FROM orders;`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all orders: %w", err)
	}
	defer rows.Close()

	// Используем make для предварительного выделения памяти, если ожидаем много заказов.
	// Это небольшая оптимизация.
	orders := make([]domain.Order, 0)

	for rows.Next() {
		var o domain.Order
		var deliveryJSON, paymentJSON, itemsJSON []byte

		if err := rows.Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &deliveryJSON, &paymentJSON, &itemsJSON, &o.Locale, &o.CustomerID, &o.DateCreated); err != nil {
			// Если произошла ошибка сканирования строки, логируем и продолжаем,
			// чтобы одна плохая запись не сломала весь процесс.
			log.Printf("ERROR: failed to scan order row: %v", err)
			continue
		}

		// Теперь Unmarshal с обработкой ошибок.
		if err := json.Unmarshal(deliveryJSON, &o.Delivery); err != nil {
			log.Printf("ERROR: failed to unmarshal delivery for order %s: %v", o.OrderUID, err)
			continue
		}
		if err := json.Unmarshal(paymentJSON, &o.Payment); err != nil {
			log.Printf("ERROR: failed to unmarshal payment for order %s: %v", o.OrderUID, err)
			continue
		}
		if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
			log.Printf("ERROR: failed to unmarshal items for order %s: %v", o.OrderUID, err)
			continue
		}

		orders = append(orders, o)
	}

	// Проверяем, не было ли ошибок во время итерации по строкам.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return orders, nil
}
