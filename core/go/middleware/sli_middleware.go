package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/sli"
)

// SLIMiddleware creates HTTP middleware that automatically records SLI metrics for all requests
// It measures latency, success/failure rates, and error types for dashboard observability
func SLIMiddleware(tracker sli.Tracker, serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip metrics and health check endpoints (they would pollute metrics)
			if r.URL.Path == "/metrics" || r.URL.Path == "/health" || r.URL.Path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip WebSocket upgrade requests - they require direct access to underlying connection via Hijack()
			// Wrapping the response writer breaks the WebSocket handshake
			if r.Header.Get("Upgrade") == "websocket" && r.Header.Get("Connection") == "Upgrade" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			// Operation name for metrics: METHOD + PATH (e.g., "GET /api/users")
			operation := r.Method + " " + r.URL.Path

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler in chain
			next.ServeHTTP(wrapped, r)

			// Record SLI metrics after request completes
			duration := time.Since(start)
			statusCode := wrapped.statusCode

			// Determine success based on status code
			isSuccess := statusCode >= 200 && statusCode < 400

			// Record request outcome
			outcome := sli.RequestOutcome{
				Success:   isSuccess,
				ErrorCode: http.StatusText(statusCode),
				Latency:   duration,
				Operation: operation,
				Timestamp: start,
			}

			ctx := context.Background()
			tracker.RecordRequest(ctx, outcome)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and additional context
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader implements http.ResponseWriter.WriteHeader to capture status code
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write implements http.ResponseWriter.Write to ensure WriteHeader is called
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
