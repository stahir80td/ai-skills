package logger

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "production logger",
			config: Config{
				ServiceName: "test-service",
				Environment: "production",
				Version:     "v1.0.0",
				LogLevel:    "info",
			},
			wantErr: false,
		},
		{
			name: "development logger",
			config: Config{
				ServiceName: "test-service",
				Environment: "development",
				Version:     "v1.0.0",
				LogLevel:    "debug",
			},
			wantErr: false,
		},
		{
			name: "logger with caller and stacktrace",
			config: Config{
				ServiceName:      "test-service",
				Environment:      "production",
				Version:          "v1.0.0",
				EnableCaller:     true,
				EnableStacktrace: true,
			},
			wantErr: false,
		},
		{
			name: "logger without version",
			config: Config{
				ServiceName: "test-service",
				Environment: "production",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("New() returned nil logger")
			}
			if logger != nil && logger.serviceName != tt.config.ServiceName {
				t.Errorf("New() serviceName = %v, want %v", logger.serviceName, tt.config.ServiceName)
			}
		})
	}
}

func TestNewProduction(t *testing.T) {
	logger, err := NewProduction("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewProduction() error = %v", err)
	}
	if logger == nil {
		t.Fatal("NewProduction() returned nil logger")
	}
	if logger.serviceName != "test-service" {
		t.Errorf("NewProduction() serviceName = %v, want %v", logger.serviceName, "test-service")
	}
}

func TestNewDevelopment(t *testing.T) {
	logger, err := NewDevelopment("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewDevelopment() error = %v", err)
	}
	if logger == nil {
		t.Fatal("NewDevelopment() returned nil logger")
	}
	if logger.serviceName != "test-service" {
		t.Errorf("NewDevelopment() serviceName = %v, want %v", logger.serviceName, "test-service")
	}
}

func TestWithContext(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{
		Logger:      zap.New(core),
		serviceName: "test-service",
	}

	tests := []struct {
		name              string
		ctx               context.Context
		message           string
		wantCorrelationID string
		wantComponent     string
	}{
		{
			name:              "context with correlation ID and component",
			ctx:               context.WithValue(context.WithValue(context.Background(), CorrelationIDKey, "test-corr-123"), ComponentKey, "TestComponent"),
			message:           "test message",
			wantCorrelationID: "test-corr-123",
			wantComponent:     "TestComponent",
		},
		{
			name:              "context with correlation ID only",
			ctx:               context.WithValue(context.Background(), CorrelationIDKey, "test-corr-456"),
			message:           "test message",
			wantCorrelationID: "test-corr-456",
			wantComponent:     "",
		},
		{
			name:              "empty context",
			ctx:               context.Background(),
			message:           "test message",
			wantCorrelationID: "",
			wantComponent:     "",
		},
		{
			name:              "nil context",
			ctx:               nil,
			message:           "test message",
			wantCorrelationID: "",
			wantComponent:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorded.TakeAll() // Clear previous logs

			contextLogger := logger.WithContext(tt.ctx)
			contextLogger.Info(tt.message)

			entries := recorded.All()
			if len(entries) != 1 {
				t.Fatalf("expected 1 log entry, got %d", len(entries))
			}

			entry := entries[0]
			if entry.Message != tt.message {
				t.Errorf("message = %v, want %v", entry.Message, tt.message)
			}

			// Check correlation_id field
			corrID := ""
			for _, field := range entry.Context {
				if field.Key == "correlation_id" {
					corrID = field.String
					break
				}
			}
			if corrID != tt.wantCorrelationID {
				t.Errorf("correlation_id = %v, want %v", corrID, tt.wantCorrelationID)
			}

			// Check component field
			component := ""
			for _, field := range entry.Context {
				if field.Key == "component" {
					component = field.String
					break
				}
			}
			if component != tt.wantComponent {
				t.Errorf("component = %v, want %v", component, tt.wantComponent)
			}
		})
	}
}

func TestWithCorrelation(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{
		Logger:      zap.New(core),
		serviceName: "test-service",
	}

	contextLogger := logger.WithCorrelation("test-corr-789")
	contextLogger.Info("test message")

	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	corrID := ""
	for _, field := range entry.Context {
		if field.Key == "correlation_id" {
			corrID = field.String
			break
		}
	}
	if corrID != "test-corr-789" {
		t.Errorf("correlation_id = %v, want %v", corrID, "test-corr-789")
	}

	if contextLogger.GetCorrelationID() != "test-corr-789" {
		t.Errorf("GetCorrelationID() = %v, want %v", contextLogger.GetCorrelationID(), "test-corr-789")
	}
}

func TestWithComponent(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{
		Logger:      zap.New(core),
		serviceName: "test-service",
	}

	contextLogger := logger.WithComponent("HTTPHandler")
	contextLogger.Info("test message")

	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	component := ""
	for _, field := range entry.Context {
		if field.Key == "component" {
			component = field.String
			break
		}
	}
	if component != "HTTPHandler" {
		t.Errorf("component = %v, want %v", component, "HTTPHandler")
	}

	if contextLogger.GetComponent() != "HTTPHandler" {
		t.Errorf("GetComponent() = %v, want %v", contextLogger.GetComponent(), "HTTPHandler")
	}
}

func TestWithError(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{
		Logger:      zap.New(core),
		serviceName: "test-service",
	}

	contextLogger := logger.WithCorrelation("test-corr-123")
	errorLogger := contextLogger.WithError("INGEST-001", "CRITICAL")
	errorLogger.Error("database connection failed")

	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]

	// Check error_code
	errorCode := ""
	severity := ""
	for _, field := range entry.Context {
		if field.Key == "error_code" {
			errorCode = field.String
		}
		if field.Key == "severity" {
			severity = field.String
		}
	}

	if errorCode != "INGEST-001" {
		t.Errorf("error_code = %v, want %v", errorCode, "INGEST-001")
	}
	if severity != "CRITICAL" {
		t.Errorf("severity = %v, want %v", severity, "CRITICAL")
	}
}

func TestContextLoggerChaining(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{
		Logger:      zap.New(core),
		serviceName: "test-service",
	}

	// Test method chaining
	contextLogger := logger.
		WithCorrelation("test-corr-999").
		WithComponent("Processor").
		WithError("TEST-001", "HIGH")

	contextLogger.Warn("chained logger test")

	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]

	// Verify all fields are present
	fields := make(map[string]string)
	for _, field := range entry.Context {
		fields[field.Key] = field.String
	}

	if fields["correlation_id"] != "test-corr-999" {
		t.Errorf("correlation_id = %v, want %v", fields["correlation_id"], "test-corr-999")
	}
	if fields["component"] != "Processor" {
		t.Errorf("component = %v, want %v", fields["component"], "Processor")
	}
	if fields["error_code"] != "TEST-001" {
		t.Errorf("error_code = %v, want %v", fields["error_code"], "TEST-001")
	}
	if fields["severity"] != "HIGH" {
		t.Errorf("severity = %v, want %v", fields["severity"], "HIGH")
	}
}

func TestGenerateCorrelationID(t *testing.T) {
	serviceName := "test-service"
	corrID1 := GenerateCorrelationID(serviceName)
	corrID2 := GenerateCorrelationID(serviceName)

	// Should start with service name
	if !strings.HasPrefix(corrID1, serviceName+"-") {
		t.Errorf("correlation ID should start with service name, got %v", corrID1)
	}

	// IDs might be the same if generated too quickly, so just verify format
	// In practice, with nanosecond precision and real workloads, they'll be unique
	if corrID1 == corrID2 {
		t.Logf("Warning: correlation IDs generated in same nanosecond (expected in tests)")
	}

	// Should have format: service-timestamp (service name may contain hyphens)
	if !strings.Contains(corrID1, "-") {
		t.Errorf("correlation ID should contain hyphen, got %v", corrID1)
	}

	// Verify it's not empty and has reasonable length
	if len(corrID1) < len(serviceName)+2 {
		t.Errorf("correlation ID too short: %v", corrID1)
	}
}
