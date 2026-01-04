package middleware

import (
	"context"
	"strings"
	"testing"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
)

func TestNewKafkaCorrelationIDHelper(t *testing.T) {
	helper := NewKafkaCorrelationIDHelper("test-service")
	if helper == nil {
		t.Fatal("NewKafkaCorrelationIDHelper returned nil")
	}
	if helper.serviceName != "test-service" {
		t.Errorf("serviceName = %v, want %v", helper.serviceName, "test-service")
	}
}

func TestExtractFromHeaders(t *testing.T) {
	helper := NewKafkaCorrelationIDHelper("test-service")

	tests := []struct {
		name              string
		headers           []KafkaHeader
		wantCorrelationID string
	}{
		{
			name: "with correlation ID",
			headers: []KafkaHeader{
				{Key: "other-header", Value: []byte("other-value")},
				{Key: CorrelationIDHeader, Value: []byte("kafka-corr-123")},
			},
			wantCorrelationID: "kafka-corr-123",
		},
		{
			name: "without correlation ID",
			headers: []KafkaHeader{
				{Key: "other-header", Value: []byte("other-value")},
			},
			wantCorrelationID: "",
		},
		{
			name:              "empty headers",
			headers:           []KafkaHeader{},
			wantCorrelationID: "",
		},
		{
			name:              "nil headers",
			headers:           nil,
			wantCorrelationID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			correlationID := helper.ExtractFromHeaders(tt.headers)
			if correlationID != tt.wantCorrelationID {
				t.Errorf("ExtractFromHeaders() = %v, want %v", correlationID, tt.wantCorrelationID)
			}
		})
	}
}

func TestAddToHeaders_NewHeader(t *testing.T) {
	helper := NewKafkaCorrelationIDHelper("test-service")
	headers := []KafkaHeader{
		{Key: "other-header", Value: []byte("other-value")},
	}

	updatedHeaders, correlationID := helper.AddToHeaders(headers, "kafka-corr-456")

	if correlationID != "kafka-corr-456" {
		t.Errorf("correlation ID = %v, want %v", correlationID, "kafka-corr-456")
	}

	// Verify header was added
	found := false
	for _, header := range updatedHeaders {
		if header.Key == CorrelationIDHeader && string(header.Value) == "kafka-corr-456" {
			found = true
			break
		}
	}
	if !found {
		t.Error("correlation ID header not found in updated headers")
	}

	// Verify original headers were not modified
	if len(headers) != 1 {
		t.Errorf("original headers length = %v, want %v", len(headers), 1)
	}
}

func TestAddToHeaders_UpdateExisting(t *testing.T) {
	helper := NewKafkaCorrelationIDHelper("test-service")
	headers := []KafkaHeader{
		{Key: CorrelationIDHeader, Value: []byte("old-corr-id")},
		{Key: "other-header", Value: []byte("other-value")},
	}

	updatedHeaders, correlationID := helper.AddToHeaders(headers, "new-corr-id")

	if correlationID != "new-corr-id" {
		t.Errorf("correlation ID = %v, want %v", correlationID, "new-corr-id")
	}

	// Verify header was updated
	count := 0
	for _, header := range updatedHeaders {
		if header.Key == CorrelationIDHeader {
			count++
			if string(header.Value) != "new-corr-id" {
				t.Errorf("header value = %v, want %v", string(header.Value), "new-corr-id")
			}
		}
	}

	// Should only have one correlation ID header
	if count != 1 {
		t.Errorf("found %d correlation ID headers, want 1", count)
	}
}

func TestAddToHeaders_GeneratesID(t *testing.T) {
	helper := NewKafkaCorrelationIDHelper("test-service")
	headers := []KafkaHeader{}

	updatedHeaders, correlationID := helper.AddToHeaders(headers, "")

	// Should generate a correlation ID
	if correlationID == "" {
		t.Error("should generate correlation ID when empty string provided")
	}

	// Should start with service name
	if !strings.HasPrefix(correlationID, "test-service-") {
		t.Errorf("correlation ID should start with service name, got %v", correlationID)
	}

	// Verify header was added
	found := false
	for _, header := range updatedHeaders {
		if header.Key == CorrelationIDHeader {
			found = true
			break
		}
	}
	if !found {
		t.Error("correlation ID header not found")
	}
}

func TestExtractOrGenerateFromHeaders(t *testing.T) {
	helper := NewKafkaCorrelationIDHelper("test-service")

	t.Run("extract existing", func(t *testing.T) {
		headers := []KafkaHeader{
			{Key: CorrelationIDHeader, Value: []byte("existing-corr-789")},
		}

		correlationID, generated := helper.ExtractOrGenerateFromHeaders(headers)

		if correlationID != "existing-corr-789" {
			t.Errorf("correlation ID = %v, want %v", correlationID, "existing-corr-789")
		}
		if generated {
			t.Error("should not be marked as generated")
		}
	})

	t.Run("generate new", func(t *testing.T) {
		headers := []KafkaHeader{}

		correlationID, generated := helper.ExtractOrGenerateFromHeaders(headers)

		if correlationID == "" {
			t.Error("correlation ID should not be empty")
		}
		if !generated {
			t.Error("should be marked as generated")
		}
		if !strings.HasPrefix(correlationID, "test-service-") {
			t.Errorf("correlation ID should start with service name, got %v", correlationID)
		}
	})
}

func TestCreateContextFromHeaders(t *testing.T) {
	helper := NewKafkaCorrelationIDHelper("test-service")

	t.Run("from existing header", func(t *testing.T) {
		headers := []KafkaHeader{
			{Key: CorrelationIDHeader, Value: []byte("kafka-ctx-123")},
		}

		ctx := helper.CreateContextFromHeaders(context.Background(), headers)

		correlationID := ExtractCorrelationID(ctx)
		if correlationID != "kafka-ctx-123" {
			t.Errorf("context correlation ID = %v, want %v", correlationID, "kafka-ctx-123")
		}
	})

	t.Run("generate new for context", func(t *testing.T) {
		headers := []KafkaHeader{}

		ctx := helper.CreateContextFromHeaders(context.Background(), headers)

		correlationID := ExtractCorrelationID(ctx)
		if correlationID == "" {
			t.Error("context should have correlation ID")
		}
		if !strings.HasPrefix(correlationID, "test-service-") {
			t.Errorf("correlation ID should start with service name, got %v", correlationID)
		}
	})
}

func TestPrepareHeadersForPublish(t *testing.T) {
	helper := NewKafkaCorrelationIDHelper("test-service")

	t.Run("from context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), logger.CorrelationIDKey, "ctx-corr-456")
		headers := []KafkaHeader{}

		updatedHeaders := helper.PrepareHeadersForPublish(ctx, headers)

		// Should add correlation ID from context
		found := false
		for _, header := range updatedHeaders {
			if header.Key == CorrelationIDHeader && string(header.Value) == "ctx-corr-456" {
				found = true
				break
			}
		}
		if !found {
			t.Error("correlation ID from context not found in headers")
		}
	})

	t.Run("generate when not in context", func(t *testing.T) {
		ctx := context.Background()
		headers := []KafkaHeader{}

		updatedHeaders := helper.PrepareHeadersForPublish(ctx, headers)

		// Should generate correlation ID
		found := false
		for _, header := range updatedHeaders {
			if header.Key == CorrelationIDHeader {
				found = true
				value := string(header.Value)
				if !strings.HasPrefix(value, "test-service-") {
					t.Errorf("generated correlation ID should start with service name, got %v", value)
				}
				break
			}
		}
		if !found {
			t.Error("correlation ID header not found")
		}
	})

	t.Run("preserve other headers", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), logger.CorrelationIDKey, "ctx-corr-789")
		headers := []KafkaHeader{
			{Key: "other-header-1", Value: []byte("value-1")},
			{Key: "other-header-2", Value: []byte("value-2")},
		}

		updatedHeaders := helper.PrepareHeadersForPublish(ctx, headers)

		// Should have 3 headers total
		if len(updatedHeaders) != 3 {
			t.Errorf("expected 3 headers, got %d", len(updatedHeaders))
		}

		// Check other headers are preserved
		for _, header := range updatedHeaders {
			if header.Key == "other-header-1" && string(header.Value) != "value-1" {
				t.Error("other-header-1 value not preserved")
			}
			if header.Key == "other-header-2" && string(header.Value) != "value-2" {
				t.Error("other-header-2 value not preserved")
			}
		}
	})
}
