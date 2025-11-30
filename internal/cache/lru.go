package cache

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Evgeniy-Programming/golang/internal/domain"
)

type cacheItem struct {
	key   string
	value domain.Order
}

type OrderProvider interface {
	GetAll(ctx context.Context) ([]domain.Order, error)
}

// реализация lru-cache
type LRUCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element //мапа быстрого доступа
	queue    *list.List
	db       OrderProvider
}

func NewLRUCache(capacity int, db OrderProvider) *LRUCache {
	if capacity <= 0 {
		capacity = 100 //по умолчанию
	}
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		queue:    list.New(),
		db:       db,
	}
}

func (c *LRUCache) Set(order domain.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := order.OrderUID

	if element, exists := c.items[key]; exists {
		c.queue.MoveToFront(element)
		element.Value.(*cacheItem).value = order
		return
	}

	if c.queue.Len() >= c.capacity {
		c.removeOldest()
	}

	item := &cacheItem{key: key, value: order}
	element := c.queue.PushFront(item)
	c.items[key] = element
}

func (c *LRUCache) Get(key string) (domain.Order, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, exists := c.items[key]; exists {
		c.queue.MoveToFront(element)
		return element.Value.(*cacheItem).value, true
	}

	//если в кеше пусто
	return domain.Order{}, false
}

func (c *LRUCache) removeOldest() {
	element := c.queue.Back()
	if element != nil {
		//удаление из списка
		item := c.queue.Remove(element).(*cacheItem)
		//из мапы
		delete(c.items, item.key)
	}
}

func (c *LRUCache) WarmUp(ctx context.Context) error {
	log.Println("Warming up LRU cache...")

	orders, err := c.db.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all orders from db for cache warm up: %w", err)
	}

	for _, order := range orders {
		c.Set(order)
	}

	log.Printf("LRU Cache warmed up. Current size: %d, Capacity: %d", c.queue.Len(), c.capacity)
	return nil
}
