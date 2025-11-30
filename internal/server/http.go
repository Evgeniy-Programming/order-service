package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Evgeniy-Programming/golang/internal/domain"
)

type APIServer struct {
	addr    string
	service domain.OrderService //зависит от интерфейса
	router  *chi.Mux
}

func NewAPIServer(addr string, service domain.OrderService) *APIServer {
	router := chi.NewRouter()
	server := &APIServer{
		addr:    addr,
		service: service,
		router:  router,
	}

	server.setupRoutes()

	return server
}

func (s *APIServer) setupRoutes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	//роут получения заказа
	s.router.Get("/order/{orderUID}", s.getOrderHandler)
	s.router.Handle("/*", http.FileServer(http.Dir("./web/static")))
}

func (s *APIServer) Start() error {
	log.Printf("Starting HTTP server on %s", s.addr)
	return http.ListenAndServe(s.addr, s.router)
}

func (s *APIServer) getOrderHandler(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "orderUID")
	if uid == "" {
		s.errorResponse(w, http.StatusBadRequest, "Order UID is required")
		return
	}

	order, err := s.service.GetOrderByUID(r.Context(), uid)
	if err != nil {
		s.errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	s.jsonResponse(w, http.StatusOK, order)
}

func (s *APIServer) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("ERROR: failed to encode json response: %v", err)
	}
}

func (s *APIServer) errorResponse(w http.ResponseWriter, status int, message string) {
	s.jsonResponse(w, status, map[string]string{"error": message})
}
