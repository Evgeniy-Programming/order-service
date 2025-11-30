package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Evgeniy-Programming/golang/internal/domain"
	"github.com/Evgeniy-Programming/golang/internal/domain/mocks"

	"go.uber.org/mock/gomock"
)

// наш тестовый сценарий
func TestOrderService_CreateOrder(t *testing.T) {

	//контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrderRepository(ctrl)
	mockCache := mocks.NewMockOrderCache(ctrl)
	orderService := NewOrderService(mockRepo, mockCache)

	//тестовый заказ из кафки
	testOrder := domain.Order{
		OrderUID:    "test-uid-123",
		TrackNumber: "TEST-TRACK",
		DateCreated: time.Now(),
	}
	ctx := context.Background()

	//ожидаем, что метод save будет вызван 1 раз
	mockRepo.EXPECT().Save(gomock.Any(), testOrder).Times(1).Return(nil)
	mockCache.EXPECT().Set(testOrder).Times(1)

	err := orderService.CreateOrder(ctx, testOrder)

	if err != nil {
		t.Errorf("CreateOrder() вернул неожиданную ошибку: %v", err)
	}
}

// сценарий, если репозиторий вернется с ошибкой
func TestOrderService_CreateOrder_RepoError(t *testing.T) {
	// --- Arrange ---
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrderRepository(ctrl)
	mockCache := mocks.NewMockOrderCache(ctrl)
	orderService := NewOrderService(mockRepo, mockCache)
	testOrder := domain.Order{OrderUID: "test-uid-456"}
	ctx := context.Background()

	// --- Expectations ---

	repoError := fmt.Errorf("database is down")
	mockRepo.EXPECT().Save(gomock.Any(), testOrder).Times(1).Return(repoError)
	mockCache.EXPECT().Set(gomock.Any()).Times(0)

	// --- Act ---
	err := orderService.CreateOrder(ctx, testOrder)

	// --- Assert ---
	//проверяем, что наш сервис "пробросил" ошибку наверх
	if err == nil {
		t.Errorf("CreateOrder() должен был вернуть ошибку, но вернул nil")
	}
	if err != nil && err.Error() != "failed to save order to repository: database is down" {
		t.Errorf("CreateOrder() вернул не ту ошибку: получили %v", err)
	}
}
