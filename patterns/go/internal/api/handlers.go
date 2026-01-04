package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"github.com/your-github-org/ai-scaffolder/core/go/metrics"
	"github.com/your-github-org/ai-scaffolder/patterns/go/internal/domain/models"
	"github.com/your-github-org/ai-scaffolder/patterns/go/internal/domain/services"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// PatternsHandler handles HTTP requests for the patterns API
type PatternsHandler struct {
	service *services.PatternsService
	logger  *logger.Logger
	metrics *metrics.ServiceMetrics
}

// NewPatternsHandler creates a new patterns handler
func NewPatternsHandler(svc *services.PatternsService, log *logger.Logger, met *metrics.ServiceMetrics) *PatternsHandler {
	return &PatternsHandler{
		service: svc,
		logger:  log,
		metrics: met,
	}
}

// =============================================================================
// Health Endpoints
// =============================================================================

// Health handles GET /health
func (h *PatternsHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health := h.service.HealthCheck(ctx)

	status := http.StatusOK
	for _, v := range health {
		if v != "healthy" {
			status = http.StatusServiceUnavailable
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(health)
}

// LivenessProbe handles GET /health/live
func (h *PatternsHandler) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ReadinessProbe handles GET /health/ready
func (h *PatternsHandler) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health := h.service.HealthCheck(ctx)

	for _, v := range health {
		if v != "healthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// =============================================================================
// Order Endpoints (SQL Server)
// =============================================================================

// CreateOrder handles POST /api/v1/patterns/orders
func (h *PatternsHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Invalid request body", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	order, err := h.service.CreateOrder(ctx, &req)
	if err != nil {
		log.Error("Failed to create order", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, order)
}

// GetOrder handles GET /api/v1/patterns/orders/{id}
func (h *PatternsHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		log.Warn("Invalid order ID", zap.String("id", vars["id"]))
		h.respondError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	order, err := h.service.GetOrder(ctx, id)
	if err != nil {
		log.Error("Failed to get order", zap.Error(err))
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, order)
}

// UpdateOrderStatus handles PATCH /api/v1/patterns/orders/{id}/status
func (h *PatternsHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		log.Warn("Invalid order ID", zap.String("id", vars["id"]))
		h.respondError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	order, err := h.service.UpdateOrderStatus(ctx, id, models.OrderStatus(req.Status))
	if err != nil {
		log.Error("Failed to update order status", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, order)
}

// =============================================================================
// User Endpoints (MongoDB)
// =============================================================================

// CreateUser handles POST /api/v1/patterns/users
func (h *PatternsHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Invalid request body", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.CreateUser(ctx, &req)
	if err != nil {
		log.Error("Failed to create user", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /api/v1/patterns/users/{id}
func (h *PatternsHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		log.Warn("Invalid user ID", zap.String("id", vars["id"]))
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.service.GetUser(ctx, id)
	if err != nil {
		log.Error("Failed to get user", zap.Error(err))
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// UpdateUserPreferences handles PUT /api/v1/patterns/users/{id}/preferences
func (h *PatternsHandler) UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		log.Warn("Invalid user ID", zap.String("id", vars["id"]))
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var prefs models.UserPreferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.UpdateUserPreferences(ctx, id, prefs); err != nil {
		log.Error("Failed to update user preferences", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Preferences updated"})
}

// =============================================================================
// Telemetry Endpoints (ScyllaDB)
// =============================================================================

// RecordTelemetry handles POST /api/v1/patterns/telemetry
func (h *PatternsHandler) RecordTelemetry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	var req models.RecordTelemetryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Invalid request body", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	telemetry, err := h.service.RecordTelemetry(ctx, &req)
	if err != nil {
		log.Error("Failed to record telemetry", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, telemetry)
}

// GetTelemetryHistory handles GET /api/v1/patterns/telemetry/{deviceId}
func (h *PatternsHandler) GetTelemetryHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	vars := mux.Vars(r)
	deviceID := vars["deviceId"]

	// Parse time range from query params (default to last 24 hours)
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if start := r.URL.Query().Get("start"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			startTime = t
		}
	}
	if end := r.URL.Query().Get("end"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			endTime = t
		}
	}

	telemetry, err := h.service.GetTelemetryHistory(ctx, deviceID, startTime, endTime)
	if err != nil {
		log.Error("Failed to get telemetry history", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, telemetry)
}

// =============================================================================
// Leaderboard Endpoints (Redis)
// =============================================================================

// UpdateLeaderboard handles POST /api/v1/patterns/leaderboards/{category}/scores
func (h *PatternsHandler) UpdateLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	vars := mux.Vars(r)
	category := vars["category"]

	var req models.UpdateLeaderboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Invalid request body", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.UpdateLeaderboard(ctx, category, req.UserID, req.Score); err != nil {
		log.Error("Failed to update leaderboard", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Leaderboard updated"})
}

// GetLeaderboard handles GET /api/v1/patterns/leaderboards/{category}
func (h *PatternsHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	vars := mux.Vars(r)
	category := vars["category"]

	// Default to top 10
	top := 10

	entries, err := h.service.GetLeaderboard(ctx, category, top)
	if err != nil {
		log.Error("Failed to get leaderboard", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, entries)
}

// =============================================================================
// Session Endpoints (Redis + Kafka)
// =============================================================================

// CreateSession handles POST /api/v1/patterns/sessions
func (h *PatternsHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	var req models.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Invalid request body", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session, err := h.service.CreateSession(ctx, &req)
	if err != nil {
		log.Error("Failed to create session", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, session)
}

// =============================================================================
// Analytics Endpoints (Cross-Platform)
// =============================================================================

// GetAnalytics handles GET /api/v1/patterns/analytics
func (h *PatternsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	// Parse time range from query params (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.Add(-30 * 24 * time.Hour)

	if start := r.URL.Query().Get("start"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			startDate = t
		}
	}
	if end := r.URL.Query().Get("end"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			endDate = t
		}
	}

	analytics, err := h.service.GetAnalytics(ctx, startDate, endDate)
	if err != nil {
		log.Error("Failed to get analytics", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, analytics)
}

// =============================================================================
// Helper Methods
// =============================================================================

func (h *PatternsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *PatternsHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
