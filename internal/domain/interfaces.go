package domain

import "context"

// OrderRepository определяет контракт для хранилища заказов.
// Любая структура, которая хочет быть репозиторием, должна реализовать эти методы.
type OrderRepository interface {
	Save(ctx context.Context, order Order) error
	GetAll(ctx context.Context) ([]Order, error)
	// В будущем здесь могут быть GetByID, Update и т.д.
}

// OrderCache определяет контракт для кэша заказов.
type OrderCache interface {
	Set(order Order)
	Get(uid string) (Order, bool)
	// WarmUp может быть частью интерфейса, если мы хотим, чтобы все кэши его поддерживали.
	WarmUp(ctx context.Context) error
}

// OrderService определяет контракт для нашей бизнес-логики.
type OrderService interface {
	CreateOrder(ctx context.Context, order Order) error
	GetOrderByUID(ctx context.Context, uid string) (Order, error)
}
