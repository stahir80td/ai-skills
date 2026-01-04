# Go Patterns for AI Services

This document defines the standard patterns for Go services including SLI, SOD, SRE, contextual logging, error repository, and Prometheus metrics.

---

## Service Level Indicators (SLI)

Track key performance metrics using the Prometheus client:

```go
package sli

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Availability SLI - percentage of successful requests
    RequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "order_service_requests_total",
            Help: "Total requests",
        },
        []string{"method", "endpoint", "status"},
    )

    // Latency SLI - request duration histogram
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "order_service_request_duration_seconds",
            Help:    "Request duration",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint"},
    )

    // Throughput SLI - orders processed per second
    OrdersProcessed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "order_service_orders_processed_total",
            Help: "Orders processed",
        },
        []string{"status"},
    )
)

func RecordRequest(method, endpoint string, statusCode int, duration float64) {
    status := "success"
    if statusCode >= 400 {
        status = "error"
    }
    RequestsTotal.WithLabelValues(method, endpoint, status).Inc()
    RequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

func RecordOrderProcessed(status string) {
    OrdersProcessed.WithLabelValues(status).Inc()
}
```

### SLI Middleware

```go
package middleware

import (
    "net/http"
    "time"
    
    "github.com/go-chi/chi/v5"
    "yourservice/internal/sli"
)

func SLIMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap response writer to capture status code
        ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        next.ServeHTTP(ww, r)
        
        duration := time.Since(start).Seconds()
        
        // Get route pattern for better labels
        routePattern := chi.RouteContext(r.Context()).RoutePattern()
        if routePattern == "" {
            routePattern = r.URL.Path
        }
        
        sli.RecordRequest(r.Method, routePattern, ww.statusCode, duration)
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
    w.statusCode = code
    w.ResponseWriter.WriteHeader(code)
}
```

---

## Service Oriented Design (SOD)

Structure services with clear separation of concerns:

```
cmd/
└── server/
    └── main.go              # Entry point
internal/
├── api/
│   ├── handlers/            # HTTP handlers
│   ├── middleware/          # Cross-cutting concerns
│   └── routes.go            # Route registration
├── domain/
│   ├── models/              # Domain entities
│   ├── services/            # Business logic
│   └── errors/              # Domain errors
└── infrastructure/
    ├── repository/          # Data access
    ├── kafka/               # Event publishing
    └── redis/               # Caching
```

### Domain Service Pattern

```go
package services

import (
    "context"
    "log/slog"
    
    "yourservice/internal/domain/models"
    "yourservice/internal/domain/errors"
    "yourservice/internal/infrastructure/repository"
    "yourservice/internal/infrastructure/kafka"
)

type OrderService struct {
    repo      repository.OrderRepository
    publisher kafka.EventPublisher
    logger    *slog.Logger
}

func NewOrderService(
    repo repository.OrderRepository,
    publisher kafka.EventPublisher,
    logger *slog.Logger,
) *OrderService {
    return &OrderService{
        repo:      repo,
        publisher: publisher,
        logger:    logger,
    }
}

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*models.Order, error) {
    // Business validation
    if len(req.Items) == 0 {
        return nil, errors.EmptyOrder()
    }
    
    // Domain logic
    order := models.NewOrder(req.CustomerID, req.Items)
    
    // Persist
    if err := s.repo.Save(ctx, order); err != nil {
        return nil, err
    }
    
    // Publish event
    event := kafka.OrderCreatedEvent{
        OrderID:    order.ID,
        CustomerID: order.CustomerID,
        Total:      order.Total,
    }
    if err := s.publisher.Publish(ctx, "orders.order.created", event); err != nil {
        s.logger.Error("failed to publish event", 
            slog.String("event", "OrderCreated"),
            slog.String("orderId", order.ID),
            slog.Any("error", err))
    }
    
    s.logger.Info("order created",
        slog.String("orderId", order.ID),
        slog.String("customerId", order.CustomerID))
    
    return order, nil
}
```

---

## Site Reliability Engineering (SRE)

### Health Checks

```go
package health

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
)

type HealthChecker interface {
    Check(ctx context.Context) error
    Name() string
}

type HealthHandler struct {
    checkers []HealthChecker
}

func NewHealthHandler(checkers ...HealthChecker) *HealthHandler {
    return &HealthHandler{checkers: checkers}
}

// Liveness - just checks if app responds
func (h *HealthHandler) LivenessHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Readiness - checks all dependencies
func (h *HealthHandler) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    results := make(map[string]string)
    allHealthy := true
    
    for _, checker := range h.checkers {
        if err := checker.Check(ctx); err != nil {
            results[checker.Name()] = err.Error()
            allHealthy = false
        } else {
            results[checker.Name()] = "ok"
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    if !allHealthy {
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":  map[bool]string{true: "ok", false: "degraded"}[allHealthy],
        "checks":  results,
    })
}
```

### Circuit Breaker

```go
package resilience

import (
    "context"
    "errors"
    "sync"
    "time"
)

type CircuitBreaker struct {
    mu              sync.Mutex
    failureCount    int
    failureThreshold int
    resetTimeout    time.Duration
    state           string // "closed", "open", "half-open"
    lastFailure     time.Time
}

func NewCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        failureThreshold: threshold,
        resetTimeout:     resetTimeout,
        state:           "closed",
    }
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
    cb.mu.Lock()
    
    // Check if we should transition from open to half-open
    if cb.state == "open" && time.Since(cb.lastFailure) > cb.resetTimeout {
        cb.state = "half-open"
    }
    
    if cb.state == "open" {
        cb.mu.Unlock()
        return ErrCircuitOpen
    }
    
    cb.mu.Unlock()
    
    // Execute the function
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failureCount++
        cb.lastFailure = time.Now()
        
        if cb.failureCount >= cb.failureThreshold {
            cb.state = "open"
        }
        return err
    }
    
    // Success - reset
    cb.failureCount = 0
    cb.state = "closed"
    return nil
}
```

### Retry with Exponential Backoff

```go
package resilience

import (
    "context"
    "math"
    "time"
)

type RetryConfig struct {
    MaxRetries  int
    InitialWait time.Duration
    MaxWait     time.Duration
}

func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
    var lastErr error
    
    for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
        if err := fn(); err == nil {
            return nil
        } else {
            lastErr = err
        }
        
        if attempt < cfg.MaxRetries {
            // Calculate exponential backoff
            wait := time.Duration(math.Pow(2, float64(attempt))) * cfg.InitialWait
            if wait > cfg.MaxWait {
                wait = cfg.MaxWait
            }
            
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(wait):
            }
        }
    }
    
    return lastErr
}
```

---

## Contextual Logging

Always include context in log entries:

```go
package logger

import (
    "context"
    "log/slog"
    "os"
)

type contextKey string

const (
    CorrelationIDKey contextKey = "correlationId"
    UserIDKey        contextKey = "userId"
)

func New(serviceName string, level slog.Level) *slog.Logger {
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: level,
    })
    
    return slog.New(handler).With(
        slog.String("service", serviceName),
    )
}

// WithContext extracts context values and adds them to the logger
func WithContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
    if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
        logger = logger.With(slog.String("correlationId", correlationID))
    }
    if userID, ok := ctx.Value(UserIDKey).(string); ok {
        logger = logger.With(slog.String("userId", userID))
    }
    return logger
}

// Example usage
func (s *OrderService) ProcessOrder(ctx context.Context, orderID string) error {
    log := logger.WithContext(ctx, s.logger)
    
    log.Info("processing order", slog.String("orderId", orderID))
    
    // ... processing logic
    
    log.Info("order processed successfully",
        slog.String("orderId", orderID),
        slog.Float64("processingTime", elapsed.Seconds()))
    
    return nil
}
```

### Correlation ID Middleware

```go
package middleware

import (
    "context"
    "net/http"
    
    "github.com/google/uuid"
    "yourservice/internal/logger"
)

const CorrelationIDHeader = "X-Correlation-ID"

func CorrelationID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        correlationID := r.Header.Get(CorrelationIDHeader)
        if correlationID == "" {
            correlationID = uuid.New().String()
        }
        
        ctx := context.WithValue(r.Context(), logger.CorrelationIDKey, correlationID)
        w.Header().Set(CorrelationIDHeader, correlationID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## Error Repository

Centralized error definitions with codes:

```go
package errors

import (
    "fmt"
    "net/http"
)

type ServiceError struct {
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    StatusCode int                    `json:"-"`
    Details    map[string]interface{} `json:"details,omitempty"`
}

func (e *ServiceError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Error registry - define all errors here
var (
    ErrOrderNotFound = &ServiceError{
        Code:       "ORD-001",
        Message:    "Order not found",
        StatusCode: http.StatusNotFound,
    }
    
    ErrEmptyOrder = &ServiceError{
        Code:       "ORD-002",
        Message:    "Empty order not allowed",
        StatusCode: http.StatusBadRequest,
    }
    
    ErrInvalidStatusTransition = &ServiceError{
        Code:       "ORD-003",
        Message:    "Invalid order status transition",
        StatusCode: http.StatusConflict,
    }
    
    ErrInsufficientInventory = &ServiceError{
        Code:       "ORD-004",
        Message:    "Insufficient inventory",
        StatusCode: http.StatusConflict,
    }
    
    ErrPaymentFailed = &ServiceError{
        Code:       "ORD-005",
        Message:    "Payment failed",
        StatusCode: http.StatusPaymentRequired,
    }
)

// Factory functions with details
func NotFound(orderID string) *ServiceError {
    return &ServiceError{
        Code:       ErrOrderNotFound.Code,
        Message:    ErrOrderNotFound.Message,
        StatusCode: ErrOrderNotFound.StatusCode,
        Details:    map[string]interface{}{"orderId": orderID},
    }
}

func EmptyOrder() *ServiceError {
    return ErrEmptyOrder
}

func InvalidStatusTransition(from, to string) *ServiceError {
    return &ServiceError{
        Code:       ErrInvalidStatusTransition.Code,
        Message:    ErrInvalidStatusTransition.Message,
        StatusCode: ErrInvalidStatusTransition.StatusCode,
        Details:    map[string]interface{}{"from": from, "to": to},
    }
}

func InsufficientInventory(sku string, requested, available int) *ServiceError {
    return &ServiceError{
        Code:       ErrInsufficientInventory.Code,
        Message:    ErrInsufficientInventory.Message,
        StatusCode: ErrInsufficientInventory.StatusCode,
        Details: map[string]interface{}{
            "sku":       sku,
            "requested": requested,
            "available": available,
        },
    }
}
```

### Error Handler Middleware

```go
package middleware

import (
    "encoding/json"
    "errors"
    "log/slog"
    "net/http"
    "time"
    
    appErrors "yourservice/internal/domain/errors"
)

func ErrorHandler(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    logger.Error("panic recovered",
                        slog.Any("panic", rec),
                        slog.String("path", r.URL.Path))
                    
                    writeErrorResponse(w, http.StatusInternalServerError,
                        "SYS-001", "Internal server error", nil)
                }
            }()
            
            next.ServeHTTP(w, r)
        })
    }
}

func writeErrorResponse(w http.ResponseWriter, status int, code, message string, details map[string]interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    response := map[string]interface{}{
        "error": map[string]interface{}{
            "code":      code,
            "message":   message,
            "details":   details,
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        },
    }
    
    json.NewEncoder(w).Encode(response)
}

// Helper for handlers to respond with service errors
func RespondWithError(w http.ResponseWriter, err error) {
    var serviceErr *appErrors.ServiceError
    if errors.As(err, &serviceErr) {
        writeErrorResponse(w, serviceErr.StatusCode, serviceErr.Code, serviceErr.Message, serviceErr.Details)
        return
    }
    
    writeErrorResponse(w, http.StatusInternalServerError, "SYS-001", "Internal server error", nil)
}
```

---

## Prometheus Metrics Endpoint

### Setup in main.go

```go
package main

import (
    "log/slog"
    "net/http"
    "os"
    
    "github.com/go-chi/chi/v5"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    
    "yourservice/internal/api/handlers"
    "yourservice/internal/api/middleware"
    "yourservice/internal/health"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    r := chi.NewRouter()
    
    // Middleware
    r.Use(middleware.CorrelationID)
    r.Use(middleware.SLIMiddleware)
    r.Use(middleware.ErrorHandler(logger))
    
    // Health endpoints
    healthHandler := health.NewHealthHandler(
        // Add your health checkers here
    )
    r.Get("/health/live", healthHandler.LivenessHandler)
    r.Get("/health/ready", healthHandler.ReadinessHandler)
    
    // Prometheus metrics endpoint
    r.Handle("/metrics", promhttp.Handler())
    
    // API routes
    r.Route("/api/v1", func(r chi.Router) {
        r.Mount("/orders", handlers.OrderRoutes(logger))
    })
    
    logger.Info("starting server", slog.String("port", "8080"))
    http.ListenAndServe(":8080", r)
}
```

### Custom Business Metrics

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Counter - things that only go up
    OrdersCreated = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "orders_created_total",
            Help: "Total orders created",
        },
        []string{"source", "customer_type"},
    )
    
    // Gauge - things that go up and down
    PendingOrders = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "orders_pending",
            Help: "Current pending orders count",
        },
    )
    
    // Histogram - distributions (latency, sizes)
    OrderValue = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "order_value_dollars",
            Help:    "Order value distribution",
            Buckets: []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
        },
    )
    
    // Summary - percentiles
    ProcessingTime = promauto.NewSummary(
        prometheus.SummaryOpts{
            Name: "order_processing_seconds",
            Help: "Order processing time",
            Objectives: map[float64]float64{
                0.5:  0.05,
                0.9:  0.01,
                0.99: 0.001,
            },
        },
    )
)

func RecordOrderCreated(source, customerType string, value float64) {
    OrdersCreated.WithLabelValues(source, customerType).Inc()
    OrderValue.Observe(value)
    PendingOrders.Inc()
}

func RecordOrderCompleted(processingSeconds float64) {
    PendingOrders.Dec()
    ProcessingTime.Observe(processingSeconds)
}
```
