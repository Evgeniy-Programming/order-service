package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Evgeniy-Programming/golang/internal/models"
)

type OrderGetter interface {
	Get(uid string) (models.Order, bool)
}

type APIServer struct {
	addr  string
	cache OrderGetter
}

func NewAPIServer(addr string, cache OrderGetter) *APIServer {
	return &APIServer{addr: addr, cache: cache}
}

func (s *APIServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/order/", s.getOrderHandler)
	mux.Handle("/", http.FileServer(http.Dir("./web/static")))

	log.Printf("Starting HTTP server on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *APIServer) getOrderHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Path[len("/order/"):]
	if uid == "" {
		http.Error(w, "Order UID is required", http.StatusBadRequest)
		return
	}

	order, found := s.cache.Get(uid)
	if !found {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Для простоты разработки
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
