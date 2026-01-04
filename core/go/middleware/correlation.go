package middleware

import (
	"context"
	"net/http"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
)

const (
	// CorrelationIDHeader is the HTTP header name for correlation ID
	CorrelationIDHeader = "X-Correlation-ID"
)

// CorrelationIDMiddleware is a middleware that extracts or generates correlation IDs
type CorrelationIDMiddleware struct {
	serviceName string
}

// NewCorrelationIDMiddleware creates a new correlation ID middleware
func NewCorrelationIDMiddleware(serviceName string) *CorrelationIDMiddleware {
	return &CorrelationIDMiddleware{
		serviceName: serviceName,
	}
}

// Handler wraps an HTTP handler to add correlation ID support
// It extracts the correlation ID from the X-Correlation-ID header if present,
// or generates a new one if missing. The correlation ID is added to the request
// context and included in the response headers.
func (m *CorrelationIDMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract correlation ID from header or generate new one
		correlationID := r.Header.Get(CorrelationIDHeader)
		if correlationID == "" {
			correlationID = logger.GenerateCorrelationID(m.serviceName)
		}

		// Add correlation ID to request context
		ctx := context.WithValue(r.Context(), logger.CorrelationIDKey, correlationID)

		// Add correlation ID to response header
		w.Header().Set(CorrelationIDHeader, correlationID)

		// Continue with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HandlerFunc is a convenience wrapper for http.HandlerFunc
func (m *CorrelationIDMiddleware) HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Handler(next).ServeHTTP(w, r)
	}
}

// ExtractCorrelationID extracts the correlation ID from a context
// Returns empty string if not found
func ExtractCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if correlationID, ok := ctx.Value(logger.CorrelationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

// AddCorrelationIDToContext adds a correlation ID to a context
// Useful for background jobs, scheduled tasks, or Kafka consumers
func AddCorrelationIDToContext(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, logger.CorrelationIDKey, correlationID)
}

// AddComponentToContext adds a component name to a context
// Useful for tracking which component is processing a request
func AddComponentToContext(ctx context.Context, component string) context.Context {
	return context.WithValue(ctx, logger.ComponentKey, component)
}

// GenerateCorrelationIDForContext creates a new correlation ID and adds it to a context
// Convenience function that combines generation and context addition
func GenerateCorrelationIDForContext(ctx context.Context, serviceName string) (context.Context, string) {
	correlationID := logger.GenerateCorrelationID(serviceName)
	return AddCorrelationIDToContext(ctx, correlationID), correlationID
}
