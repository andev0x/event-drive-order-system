package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andev0x/analytics-service/internal/service"
)

// AnalyticsHandler handles HTTP requests for analytics
type AnalyticsHandler struct {
	service *service.AnalyticsService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(service *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
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
