package api

import (
	"context"
	"net/http"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"github.com/your-github-org/ai-scaffolder/core/go/metrics"
	"go.uber.org/zap"
)

// Key type for context values
type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
	requestStartKey  contextKey = "request_start"
)

// CorrelationMiddleware adds correlation ID to requests using Core.Logger
func CorrelationMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get or generate correlation ID
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = logger.GenerateCorrelationID("patterns")
			}

			// Add to response headers
			w.Header().Set("X-Correlation-ID", correlationID)

			// Add to context using Core.Logger's context support
			ctx := context.WithValue(r.Context(), correlationIDKey, correlationID)
			ctx = context.WithValue(ctx, logger.CorrelationIDKey, correlationID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestLoggingMiddleware logs requests using Core.Logger
func RequestLoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create wrapped response writer to capture status
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Add start time to context
			ctx := context.WithValue(r.Context(), requestStartKey, start)

			// Get contextual logger
			reqLog := log.WithContext(ctx)

			reqLog.Info("Request started",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()))

			next.ServeHTTP(wrapped, r.WithContext(ctx))

			duration := time.Since(start)
			reqLog.Info("Request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration))
		})
	}
}

// MetricsMiddleware records metrics using Core.Metrics
func MetricsMiddleware(met *metrics.ServiceMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create wrapped response writer to capture status
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			// Record metrics using Core.Metrics
			duration := time.Since(start)
			status := http.StatusText(wrapped.statusCode)
			met.RecordRequest(r.Method, r.URL.Path, status, duration)
		})
	}
}

// RecoveryMiddleware recovers from panics using Core.Logger
func RecoveryMiddleware(log *logger.Logger, met *metrics.ServiceMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					reqLog := log.WithContext(r.Context())
					reqLog.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method))

					// Record error metric using Core.Metrics (with component)
					met.RecordError("PANIC", "critical", r.URL.Path)

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware adds request timeout
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-ID")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
