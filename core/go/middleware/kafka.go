package middleware

import (
	"context"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
)

// KafkaHeader represents a Kafka message header
type KafkaHeader struct {
	Key   string
	Value []byte
}

// KafkaCorrelationIDHelper provides utilities for working with correlation IDs in Kafka messages
type KafkaCorrelationIDHelper struct {
	serviceName string
}

// NewKafkaCorrelationIDHelper creates a new Kafka correlation ID helper
func NewKafkaCorrelationIDHelper(serviceName string) *KafkaCorrelationIDHelper {
	return &KafkaCorrelationIDHelper{
		serviceName: serviceName,
	}
}

// ExtractFromHeaders extracts correlation ID from Kafka message headers
// Returns empty string if not found
func (h *KafkaCorrelationIDHelper) ExtractFromHeaders(headers []KafkaHeader) string {
	for _, header := range headers {
		if header.Key == CorrelationIDHeader {
			return string(header.Value)
		}
	}
	return ""
}

// AddToHeaders adds correlation ID to Kafka message headers
// If correlation ID is empty, generates a new one
// Returns the updated headers slice and the correlation ID used
func (h *KafkaCorrelationIDHelper) AddToHeaders(headers []KafkaHeader, correlationID string) ([]KafkaHeader, string) {
	// Generate correlation ID if not provided
	if correlationID == "" {
		correlationID = logger.GenerateCorrelationID(h.serviceName)
	}

	// Check if correlation ID already exists in headers
	for i, header := range headers {
		if header.Key == CorrelationIDHeader {
			headers[i].Value = []byte(correlationID)
			return headers, correlationID
		}
	}

	// Add new header
	headers = append(headers, KafkaHeader{
		Key:   CorrelationIDHeader,
		Value: []byte(correlationID),
	})

	return headers, correlationID
}

// ExtractOrGenerateFromHeaders extracts correlation ID from headers or generates a new one
// Returns the correlation ID and whether it was newly generated
func (h *KafkaCorrelationIDHelper) ExtractOrGenerateFromHeaders(headers []KafkaHeader) (string, bool) {
	correlationID := h.ExtractFromHeaders(headers)
	if correlationID != "" {
		return correlationID, false
	}

	return logger.GenerateCorrelationID(h.serviceName), true
}

// CreateContextFromHeaders creates a context with correlation ID from Kafka headers
// Extracts correlation ID from headers or generates a new one, then adds to context
func (h *KafkaCorrelationIDHelper) CreateContextFromHeaders(ctx context.Context, headers []KafkaHeader) context.Context {
	correlationID, _ := h.ExtractOrGenerateFromHeaders(headers)
	return AddCorrelationIDToContext(ctx, correlationID)
}

// PrepareHeadersForPublish prepares Kafka headers with correlation ID from context
// Extracts correlation ID from context if available, or generates a new one
// Returns the updated headers slice
func (h *KafkaCorrelationIDHelper) PrepareHeadersForPublish(ctx context.Context, headers []KafkaHeader) []KafkaHeader {
	correlationID := ExtractCorrelationID(ctx)
	updatedHeaders, _ := h.AddToHeaders(headers, correlationID)
	return updatedHeaders
}
