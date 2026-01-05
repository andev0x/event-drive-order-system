package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/anvndev/order-service/internal/model"
	"github.com/anvndev/order-service/internal/service"
	"github.com/gorilla/mux"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	service *service.OrderService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	order, err := h.service.CreateOrder(r.Context(), &req)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, order)
}

// GetOrder handles GET /orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), id)
	if err != nil {
		log.Printf("Error getting order: %v", err)
		respondWithError(w, http.StatusNotFound, "Order not found")
		return
	}

	respondWithJSON(w, http.StatusOK, order)
}

// ListOrders handles GET /orders
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	orders, err := h.service.ListOrders(r.Context(), limit, offset)
	if err != nil {
		log.Printf("Error listing orders: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to list orders")
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}

// HealthCheck handles GET /health
func (h *OrderHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
