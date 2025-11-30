package cache

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"sync"

	// Опять же, этот путь позже изменится на `.../internal/domain`
	"github.com/Evgeniy-Programming/golang/internal/domain"
)

// cacheItem - это то, что мы будем хранить в узлах нашего списка.
// Он содержит и сам заказ, и его ключ, чтобы мы могли удалить его из мапы при вытеснении.
type cacheItem struct {
	key   string
	value domain.Order
}

// OrderProvider - это интерфейс для получения всех заказов из БД.
// Мы оставим его здесь для функции WarmUp, но в "чистой" архитектуре
// он бы тоже переехал в пакет `domain`.
type OrderProvider interface {
	GetAll(ctx context.Context) ([]domain.Order, error)
}

// LRUCache - наша реализация LRU-кэша.
type LRUCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element // Мапа для быстрого доступа
	queue    *list.List               // Двусвязный список для порядка использования
	db       OrderProvider
}

// NewLRUCache создает новый LRU-кэш с заданной вместимостью.
func NewLRUCache(capacity int, db OrderProvider) *LRUCache {
	if capacity <= 0 {
		capacity = 100 // Установим значение по умолчанию, если указана некорректная вместимость
	}
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		queue:    list.New(),
		db:       db,
	}
}

// Set добавляет заказ в кэш.
func (c *LRUCache) Set(order domain.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := order.OrderUID

	// Если элемент уже есть в кэше, обновляем его значение и перемещаем в начало.
	if element, exists := c.items[key]; exists {
		c.queue.MoveToFront(element)
		element.Value.(*cacheItem).value = order
		return
	}

	// Если кэш переполнен, удаляем самый старый элемент (из хвоста списка).
	if c.queue.Len() >= c.capacity {
		c.removeOldest()
	}

	// Создаем новый элемент и добавляем его в начало списка и в мапу.
	item := &cacheItem{key: key, value: order}
	element := c.queue.PushFront(item)
	c.items[key] = element
}

// Get получает заказ из кэша.
func (c *LRUCache) Get(key string) (domain.Order, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Ищем элемент в мапе.
	if element, exists := c.items[key]; exists {
		// Если нашли, перемещаем его в начало списка, так как он стал "недавно использованным".
		c.queue.MoveToFront(element)
		// Возвращаем найденное значение.
		return element.Value.(*cacheItem).value, true
	}

	// Если в кэше нет, возвращаем "не найдено".
	return domain.Order{}, false
}

// removeOldest - это внутренний метод для удаления самого старого элемента.
// Он должен вызываться только внутри блокировки мьютекса.
func (c *LRUCache) removeOldest() {
	// Получаем самый старый элемент (последний в списке).
	element := c.queue.Back()
	if element != nil {
		// Удаляем его из списка.
		item := c.queue.Remove(element).(*cacheItem)
		// Удаляем его из мапы.
		delete(c.items, item.key)
	}
}

// WarmUp заполняет кэш данными из базы данных при старте.
func (c *LRUCache) WarmUp(ctx context.Context) error {
	log.Println("Warming up LRU cache...")

	// Используем метод GetAll, который мы переименовали в репозитории.
	orders, err := c.db.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all orders from db for cache warm up: %w", err)
	}

	// Добавляем заказы в кэш с помощью нашего метода Set.
	// Мы не делаем это напрямую, чтобы не нарушать логику LRU.
	// Если заказов в базе больше, чем вместимость кэша, в кэше останутся только последние.
	for _, order := range orders {
		c.Set(order)
	}

	log.Printf("LRU Cache warmed up. Current size: %d, Capacity: %d", c.queue.Len(), c.capacity)
	return nil
}
