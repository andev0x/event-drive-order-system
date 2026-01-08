// Package handler provides HTTP request handlers for the order service.
package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/andev0x/order-service/internal/model"
	"github.com/andev0x/order-service/internal/service"
	"github.com/gorilla/mux"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	service     *service.OrderService
	healthCheck *HealthChecker
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	DBHealthFunc    func() error
	CacheHealthFunc func() error
	MQHealthFunc    func() error
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{
		service:     service,
		healthCheck: nil, // Set via SetHealthChecker
	}
}

// SetHealthChecker sets the health checker
func (h *OrderHandler) SetHealthChecker(hc *HealthChecker) {
	h.healthCheck = hc
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
func (h *OrderHandler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "order-service",
	}

	// Check dependencies if health checker is configured
	if h.healthCheck != nil {
		checks := map[string]string{
			"database": "healthy",
			"cache":    "healthy",
			"mq":       "healthy",
		}

		overallHealthy := true

		// Check database
		if h.healthCheck.DBHealthFunc != nil {
			if err := h.healthCheck.DBHealthFunc(); err != nil {
				checks["database"] = "unhealthy: " + err.Error()
				overallHealthy = false
			}
		}

		// Check cache
		if h.healthCheck.CacheHealthFunc != nil {
			if err := h.healthCheck.CacheHealthFunc(); err != nil {
				checks["cache"] = "unhealthy: " + err.Error()
				overallHealthy = false
			}
		}

		// Check message queue
		if h.healthCheck.MQHealthFunc != nil {
			if err := h.healthCheck.MQHealthFunc(); err != nil {
				checks["mq"] = "unhealthy: " + err.Error()
				overallHealthy = false
			}
		}

		response["checks"] = checks
		if !overallHealthy {
			response["status"] = "degraded"
			respondWithJSON(w, http.StatusServiceUnavailable, response)
			return
		}
	}

	respondWithJSON(w, http.StatusOK, response)
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
		if _, writeErr := w.Write([]byte("Internal server error")); writeErr != nil {
			log.Printf("Error writing error response: %v", writeErr)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write(response); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
