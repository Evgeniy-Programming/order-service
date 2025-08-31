package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Evgeniy-Programming/golang/internal/models"
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

func (r *OrderRepository) SaveOrder(ctx context.Context, order models.Order) error {
	query := `
        INSERT INTO orders (order_uid, track_number, entry, delivery, payment, items, locale, customer_id, date_created)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (order_uid) DO NOTHING;`

	deliveryJSON, _ := json.Marshal(order.Delivery)
	paymentJSON, _ := json.Marshal(order.Payment)
	itemsJSON, _ := json.Marshal(order.Items)

	_, err := r.db.Exec(ctx, query,
		order.OrderUID, order.TrackNumber, order.Entry, deliveryJSON, paymentJSON, itemsJSON,
		order.Locale, order.CustomerID, order.DateCreated,
	)
	return err
}

func (r *OrderRepository) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	query := `SELECT order_uid, track_number, entry, delivery, payment, items, locale, customer_id, date_created FROM orders;`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		var deliveryJSON, paymentJSON, itemsJSON []byte
		if err := rows.Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &deliveryJSON, &paymentJSON, &itemsJSON, &o.Locale, &o.CustomerID, &o.DateCreated); err != nil {
			return nil, err
		}
		json.Unmarshal(deliveryJSON, &o.Delivery)
		json.Unmarshal(paymentJSON, &o.Payment)
		json.Unmarshal(itemsJSON, &o.Items)
		orders = append(orders, o)
	}
	return orders, nil
}
