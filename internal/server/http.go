package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Evgeniy-Programming/golang/internal/domain"
)

// APIServer - наша структура HTTP-сервера.
type APIServer struct {
	addr    string
	service domain.OrderService // <-- ЗАВИСИМОСТЬ ОТ ИНТЕРФЕЙСА СЕРВИСА
	router  *chi.Mux
}

// NewAPIServer создает новый экземпляр сервера.
// Теперь он принимает domain.OrderService вместо кэша.
func NewAPIServer(addr string, service domain.OrderService) *APIServer {
	router := chi.NewRouter()
	server := &APIServer{
		addr:    addr,
		service: service,
		router:  router,
	}

	// Настраиваем роуты
	server.setupRoutes()

	return server
}

// setupRoutes инкапсулирует всю логику настройки роутинга.
func (s *APIServer) setupRoutes() {
	// Используем стандартные middleware от chi для логирования запросов,
	// восстановления после паник и т.д.
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Роут для получения заказа. {orderUID} - это параметр пути.
	s.router.Get("/order/{orderUID}", s.getOrderHandler)

	// Роут для статических файлов нашего веб-интерфейса.
	s.router.Handle("/*", http.FileServer(http.Dir("./web/static")))
}

// Start запускает HTTP-сервер.
func (s *APIServer) Start() error {
	log.Printf("Starting HTTP server on %s", s.addr)
	return http.ListenAndServe(s.addr, s.router)
}

// getOrderHandler - обработчик запроса на получение заказа.
func (s *APIServer) getOrderHandler(w http.ResponseWriter, r *http.Request) {
	// chi позволяет легко получить параметр из URL.
	uid := chi.URLParam(r, "orderUID")
	if uid == "" {
		s.errorResponse(w, http.StatusBadRequest, "Order UID is required")
		return
	}

	// Вызываем метод сервиса, а не кэша.
	// Сервер не знает, откуда придут данные - из кэша или из БД. Это инкапсулировано в сервисе.
	order, err := s.service.GetOrderByUID(r.Context(), uid)
	if err != nil {
		// Если сервис вернул ошибку, значит, заказ не найден.
		s.errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	// Отправляем успешный JSON-ответ.
	s.jsonResponse(w, http.StatusOK, order)
}

// jsonResponse - вспомогательная функция для отправки JSON-ответов.
func (s *APIServer) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Для разработки
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("ERROR: failed to encode json response: %v", err)
	}
}

// errorResponse - вспомогательная функция для отправки JSON-ответов с ошибками.
func (s *APIServer) errorResponse(w http.ResponseWriter, status int, message string) {
	s.jsonResponse(w, status, map[string]string{"error": message})
}
