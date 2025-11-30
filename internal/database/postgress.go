package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Evgeniy-Programming/golang/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func ConnectDB(connStr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return pool, nil
}

func (r *OrderRepository) Save(ctx context.Context, order domain.Order) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

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

	//запрос в бд, но уже с транзакцией
	query := `
        INSERT INTO orders (order_uid, track_number, entry, delivery, payment, items, locale, customer_id, date_created)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (order_uid) DO NOTHING;`

	if _, err := tx.Exec(ctx, query,
		order.OrderUID, order.TrackNumber, order.Entry, deliveryJSON, paymentJSON, itemsJSON,
		order.Locale, order.CustomerID, order.DateCreated,
	); err != nil {
		return fmt.Errorf("failed to execute insert query: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) GetAll(ctx context.Context) ([]domain.Order, error) {
	query := `SELECT order_uid, track_number, entry, delivery, payment, items, locale, customer_id, date_created FROM orders;`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all orders: %w", err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)

	for rows.Next() {
		var o domain.Order
		var deliveryJSON, paymentJSON, itemsJSON []byte

		if err := rows.Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &deliveryJSON, &paymentJSON, &itemsJSON, &o.Locale, &o.CustomerID, &o.DateCreated); err != nil {
			log.Printf("ERROR: failed to scan order row: %v", err)
			continue
		}

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

	//проверка на неточности
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return orders, nil
}
