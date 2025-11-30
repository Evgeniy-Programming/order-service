package service

import (
	"context"
	"fmt"

	//зависимость только от домена
	"github.com/Evgeniy-Programming/golang/internal/domain"
)

type service struct {
	repo  domain.OrderRepository
	cache domain.OrderCache
}

func NewOrderService(repo domain.OrderRepository, cache domain.OrderCache) domain.OrderService {
	return &service{
		repo:  repo,
		cache: cache,
	}
}

func (s *service) CreateOrder(ctx context.Context, order domain.Order) error {

	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order to repository: %w", err)
	}

	s.cache.Set(order)

	return nil
}

func (s *service) GetOrderByUID(ctx context.Context, uid string) (domain.Order, error) {
	if order, found := s.cache.Get(uid); found {
		return order, nil
	}

	//если в кеше нет - идем в бд

	return domain.Order{}, fmt.Errorf("order with UID %s not found", uid)
}
