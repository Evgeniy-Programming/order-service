package cache

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Evgeniy-Programming/golang/internal/models"
)

type OrderProvider interface {
	GetAllOrders(ctx context.Context) ([]models.Order, error)
}

type MemoryCache struct {
	sync.RWMutex
	orders map[string]models.Order
	db     OrderProvider
}

func NewMemoryCache(db OrderProvider) *MemoryCache {
	return &MemoryCache{
		orders: make(map[string]models.Order),
		db:     db,
	}
}

func (c *MemoryCache) WarmUp(ctx context.Context) error {
	log.Println("Warming up cache...")
	orders, err := c.db.GetAllOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all orders from db for cache warm up: %w", err)
	}
	c.Lock()
	defer c.Unlock()
	for _, order := range orders {
		c.orders[order.OrderUID] = order
	}
	log.Printf("%d orders loaded into cache.", len(orders))
	return nil
}

func (c *MemoryCache) Get(uid string) (models.Order, bool) {
	c.RLock()
	defer c.RUnlock()
	order, found := c.orders[uid]
	return order, found
}

func (c *MemoryCache) Set(order models.Order) {
	c.Lock()
	defer c.Unlock()
	c.orders[order.OrderUID] = order
}
