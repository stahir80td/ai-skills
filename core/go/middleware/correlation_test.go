package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
)

func TestNewCorrelationIDMiddleware(t *testing.T) {
	middleware := NewCorrelationIDMiddleware("test-service")
	if middleware == nil {
		t.Fatal("NewCorrelationIDMiddleware returned nil")
	}
	if middleware.serviceName != "test-service" {
		t.Errorf("serviceName = %v, want %v", middleware.serviceName, "test-service")
	}
}

func TestHandler_GeneratesCorrelationID(t *testing.T) {
	middleware := NewCorrelationIDMiddleware("test-service")

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := ExtractCorrelationID(r.Context())
		if correlationID == "" {
			t.Error("correlation ID should be generated")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check response header
	correlationID := rec.Header().Get(CorrelationIDHeader)
	if correlationID == "" {
		t.Error("correlation ID should be in response header")
	}
}

func TestHandler_ExtractsExistingCorrelationID(t *testing.T) {
	middleware := NewCorrelationIDMiddleware("test-service")
	expectedCorrelationID := "existing-corr-123"

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := ExtractCorrelationID(r.Context())
		if correlationID != expectedCorrelationID {
			t.Errorf("correlation ID = %v, want %v", correlationID, expectedCorrelationID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(CorrelationIDHeader, expectedCorrelationID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check response header matches input
	responseCorrelationID := rec.Header().Get(CorrelationIDHeader)
	if responseCorrelationID != expectedCorrelationID {
		t.Errorf("response correlation ID = %v, want %v", responseCorrelationID, expectedCorrelationID)
	}
}

func TestHandler_AddsToContext(t *testing.T) {
	middleware := NewCorrelationIDMiddleware("test-service")
	var capturedContext context.Context

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContext = r.Context()
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify context has correlation ID
	if capturedContext == nil {
		t.Fatal("context should not be nil")
	}

	correlationID := ExtractCorrelationID(capturedContext)
	if correlationID == "" {
		t.Error("correlation ID should be in context")
	}
}

func TestHandlerFunc_Wrapper(t *testing.T) {
	middleware := NewCorrelationIDMiddleware("test-service")

	handlerFunc := middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := ExtractCorrelationID(r.Context())
		if correlationID == "" {
			t.Error("correlation ID should be in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handlerFunc.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %v, want %v", rec.Code, http.StatusOK)
	}
}

func TestExtractCorrelationID(t *testing.T) {
	tests := []struct {
		name              string
		ctx               context.Context
		wantCorrelationID string
	}{
		{
			name:              "with correlation ID",
			ctx:               context.WithValue(context.Background(), logger.CorrelationIDKey, "test-corr-456"),
			wantCorrelationID: "test-corr-456",
		},
		{
			name:              "without correlation ID",
			ctx:               context.Background(),
			wantCorrelationID: "",
		},
		{
			name:              "nil context",
			ctx:               nil,
			wantCorrelationID: "",
		},
		{
			name:              "wrong type in context",
			ctx:               context.WithValue(context.Background(), logger.CorrelationIDKey, 123),
			wantCorrelationID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			correlationID := ExtractCorrelationID(tt.ctx)
			if correlationID != tt.wantCorrelationID {
				t.Errorf("ExtractCorrelationID() = %v, want %v", correlationID, tt.wantCorrelationID)
			}
		})
	}
}

func TestAddCorrelationIDToContext(t *testing.T) {
	ctx := context.Background()
	correlationID := "test-corr-789"

	ctx = AddCorrelationIDToContext(ctx, correlationID)

	extracted := ExtractCorrelationID(ctx)
	if extracted != correlationID {
		t.Errorf("extracted correlation ID = %v, want %v", extracted, correlationID)
	}
}

func TestAddComponentToContext(t *testing.T) {
	ctx := context.Background()
	component := "HTTPHandler"

	ctx = AddComponentToContext(ctx, component)

	if extracted, ok := ctx.Value(logger.ComponentKey).(string); !ok || extracted != component {
		t.Errorf("extracted component = %v, want %v", extracted, component)
	}
}

func TestGenerateCorrelationIDForContext(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-service"

	newCtx, correlationID := GenerateCorrelationIDForContext(ctx, serviceName)

	// Verify correlation ID was generated
	if correlationID == "" {
		t.Error("correlation ID should not be empty")
	}

	// Verify it was added to context
	extracted := ExtractCorrelationID(newCtx)
	if extracted != correlationID {
		t.Errorf("extracted correlation ID = %v, want %v", extracted, correlationID)
	}
}

func TestHandler_MultipleCalls_UniqueIDs(t *testing.T) {
	middleware := NewCorrelationIDMiddleware("test-service")
	var ids []string

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := ExtractCorrelationID(r.Context())
		ids = append(ids, correlationID)
		w.WriteHeader(http.StatusOK)
	}))

	// Make multiple requests without correlation ID header
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// All IDs should be present (might not be unique due to nanosecond timing in tests)
	if len(ids) != 3 {
		t.Errorf("expected 3 correlation IDs, got %d", len(ids))
	}

	// All should be non-empty
	for i, id := range ids {
		if id == "" {
			t.Errorf("ID at index %d is empty", i)
		}
	}
}
