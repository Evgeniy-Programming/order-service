package domain

import "context"

type OrderRepository interface {
	Save(ctx context.Context, order Order) error
	GetAll(ctx context.Context) ([]Order, error)
}

type OrderCache interface {
	Set(order Order)
	Get(uid string) (Order, bool)
	WarmUp(ctx context.Context) error
}

type OrderService interface {
	CreateOrder(ctx context.Context, order Order) error
	GetOrderByUID(ctx context.Context, uid string) (Order, error)
}
