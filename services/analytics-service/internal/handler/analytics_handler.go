package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andev0x/analytics-service/internal/service"
)

// AnalyticsHandler handles HTTP requests for analytics
type AnalyticsHandler struct {
	service     *service.AnalyticsService
	healthCheck *HealthChecker
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	DBHealthFunc    func() error
	CacheHealthFunc func() error
	MQHealthFunc    func() error
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(service *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		service:     service,
		healthCheck: nil, // Set via SetHealthChecker
	}
}

// SetHealthChecker sets the health checker
func (h *AnalyticsHandler) SetHealthChecker(hc *HealthChecker) {
	h.healthCheck = hc
}

// GetSummary handles GET /analytics/summary
func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.GetSummary(r.Context())
	if err != nil {
		log.Printf("Error getting summary: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get analytics summary")
		return
	}

	respondWithJSON(w, http.StatusOK, summary)
}

// HealthCheck handles GET /health
func (h *AnalyticsHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "analytics-service",
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
		w.Write([]byte("Internal server error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
