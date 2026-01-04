package api

import (
	"net/http"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"github.com/your-github-org/ai-scaffolder/core/go/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes configures all routes for the patterns API
func SetupRoutes(
	handler *PatternsHandler,
	log *logger.Logger,
	met *metrics.ServiceMetrics,
) *mux.Router {
	router := mux.NewRouter()

	// Apply global middleware using Core packages
	router.Use(
		CORSMiddleware(),
		CorrelationMiddleware(log),    // Core.Logger correlation
		RequestLoggingMiddleware(log), // Core.Logger request logging
		MetricsMiddleware(met),        // Core.Metrics
		RecoveryMiddleware(log, met),  // Core.Logger + Core.Metrics
		TimeoutMiddleware(30*time.Second),
	)

	// ========================================================================
	// Health & Monitoring Routes
	// ========================================================================
	router.HandleFunc("/health", handler.Health).Methods("GET")
	router.HandleFunc("/health/live", handler.LivenessProbe).Methods("GET")
	router.HandleFunc("/health/ready", handler.ReadinessProbe).Methods("GET")

	// Prometheus metrics endpoint (Core.Metrics)
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// ========================================================================
	// API v1 Routes - Demonstrating Core Infrastructure Usage
	// ========================================================================
	apiV1 := router.PathPrefix("/api/v1/patterns").Subrouter()

	// Health check within API
	apiV1.HandleFunc("/health", handler.Health).Methods("GET")

	// SQL Server Patterns - Orders (Core.Infrastructure.SqlServer)
	apiV1.HandleFunc("/orders", handler.CreateOrder).Methods("POST")
	apiV1.HandleFunc("/orders/{id}", handler.GetOrder).Methods("GET")
	apiV1.HandleFunc("/orders/{id}/status", handler.UpdateOrderStatus).Methods("PATCH")

	// MongoDB Patterns - Users (Core.Infrastructure.MongoDB)
	apiV1.HandleFunc("/users", handler.CreateUser).Methods("POST")
	apiV1.HandleFunc("/users/{id}", handler.GetUser).Methods("GET")
	apiV1.HandleFunc("/users/{id}/preferences", handler.UpdateUserPreferences).Methods("PUT")

	// ScyllaDB Patterns - Telemetry (Core.Infrastructure.ScyllaDB)
	apiV1.HandleFunc("/telemetry", handler.RecordTelemetry).Methods("POST")
	apiV1.HandleFunc("/telemetry/{deviceId}", handler.GetTelemetryHistory).Methods("GET")

	// Redis Patterns - Leaderboards (Core.Infrastructure.Redis)
	apiV1.HandleFunc("/leaderboards/{category}/scores", handler.UpdateLeaderboard).Methods("POST")
	apiV1.HandleFunc("/leaderboards/{category}", handler.GetLeaderboard).Methods("GET")

	// Redis + Kafka Patterns - Sessions
	apiV1.HandleFunc("/sessions", handler.CreateSession).Methods("POST")

	// Cross-Platform Analytics (All Core Infrastructure)
	apiV1.HandleFunc("/analytics", handler.GetAnalytics).Methods("GET")

	return router
}

// NewServer creates a new HTTP server with the configured router
func NewServer(
	port string,
	handler *PatternsHandler,
	log *logger.Logger,
	met *metrics.ServiceMetrics,
) *http.Server {
	router := SetupRoutes(handler, log, met)

	return &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
