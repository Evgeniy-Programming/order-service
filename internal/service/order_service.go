package service

import (
	"context"
	"fmt"

	// Зависим только от доменного слоя!
	"github.com/Evgeniy-Programming/golang/internal/domain"
)

// service реализует интерфейс domain.OrderService.
type service struct {
	repo  domain.OrderRepository
	cache domain.OrderCache
}

// NewOrderService - конструктор для нашего сервиса.
// Он принимает интерфейсы, а не конкретные реализации.
func NewOrderService(repo domain.OrderRepository, cache domain.OrderCache) domain.OrderService {
	return &service{
		repo:  repo,
		cache: cache,
	}
}

// CreateOrder - основной метод бизнес-логики для создания заказа.
func (s *service) CreateOrder(ctx context.Context, order domain.Order) error {
	// Здесь может быть сложная бизнес-логика: проверки, расчеты и т.д.

	// Шаг 1: Сохраняем в базу данных (используя транзакцию внутри репозитория).
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order to repository: %w", err)
	}

	// Шаг 2: Если сохранение в БД прошло успешно, обновляем кэш.
	s.cache.Set(order)

	return nil
}

// GetOrderByUID - метод для получения заказа.
func (s *service) GetOrderByUID(ctx context.Context, uid string) (domain.Order, error) {
	// Шаг 1: Пытаемся получить из кэша.
	if order, found := s.cache.Get(uid); found {
		return order, nil
	}

	// Шаг 2: Если в кэше нет, идем в базу.
	// (Для этого нам нужно добавить метод GetByID в репозиторий, но пока опустим это
	// для простоты. Сейчас мы сфокусированы на логике сохранения).
	// В реальном приложении здесь был бы вызов s.repo.GetByID(ctx, uid).

	return domain.Order{}, fmt.Errorf("order with UID %s not found", uid)
}
