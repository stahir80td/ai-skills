package logger

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds logger configuration options
type Config struct {
	ServiceName string
	Environment string
	Version     string
	LogLevel    string // Optional: "debug", "info", "warn", "error"

	// Advanced options
	EnableCaller     bool // Include caller information (file:line)
	EnableStacktrace bool // Include stacktrace for errors
}

// Logger wraps zap.Logger with additional SRE functionality
type Logger struct {
	*zap.Logger
	serviceName string
}

// ContextLogger provides correlation-aware logging
type ContextLogger struct {
	*zap.Logger
	correlationID string
	component     string
}

// ContextKey type for context values
type contextKey string

const (
	// CorrelationIDKey is the context key for correlation IDs
	CorrelationIDKey contextKey = "correlation_id"

	// ComponentKey is the context key for component names
	ComponentKey contextKey = "component"
)

// New creates a new structured logger with consistent configuration
func New(cfg Config) (*Logger, error) {
	var config zap.Config

	if cfg.Environment == "development" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	// Consistent encoder configuration for SRE
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	// Set log level if specified
	if cfg.LogLevel != "" {
		var level zapcore.Level
		if err := level.UnmarshalText([]byte(cfg.LogLevel)); err == nil {
			config.Level = zap.NewAtomicLevelAt(level)
		}
	}

	// Configure caller and stacktrace
	options := []zap.Option{}
	if cfg.EnableCaller {
		options = append(options, zap.AddCaller())
	}
	if cfg.EnableStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Build logger with service metadata
	fields := []zap.Field{
		zap.String("service", cfg.ServiceName),
		zap.String("environment", cfg.Environment),
	}

	if cfg.Version != "" {
		fields = append(fields, zap.String("version", cfg.Version))
	}

	options = append(options, zap.Fields(fields...))

	zapLogger, err := config.Build(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{
		Logger:      zapLogger,
		serviceName: cfg.ServiceName,
	}, nil
}

// NewProduction creates a production logger with standard settings
// Enables caller info and stacktraces for errors
func NewProduction(serviceName, version string) (*Logger, error) {
	return New(Config{
		ServiceName:      serviceName,
		Environment:      "production",
		Version:          version,
		LogLevel:         "info",
		EnableCaller:     true,
		EnableStacktrace: true,
	})
}

// NewDevelopment creates a development logger with colorized output
func NewDevelopment(serviceName, version string) (*Logger, error) {
	return New(Config{
		ServiceName:      serviceName,
		Environment:      "development",
		Version:          version,
		LogLevel:         "debug",
		EnableCaller:     true,
		EnableStacktrace: true,
	})
}

// WithContext creates a context-aware logger from the context
// Extracts correlation_id and component from context if available
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	correlationID := ""
	component := ""

	if ctx != nil {
		if corrID, ok := ctx.Value(CorrelationIDKey).(string); ok {
			correlationID = corrID
		}
		if comp, ok := ctx.Value(ComponentKey).(string); ok {
			component = comp
		}
	}

	fields := []zap.Field{}
	if correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}
	if component != "" {
		fields = append(fields, zap.String("component", component))
	}

	return &ContextLogger{
		Logger:        l.Logger.With(fields...),
		correlationID: correlationID,
		component:     component,
	}
}

// WithCorrelation creates a logger with correlation ID
func (l *Logger) WithCorrelation(correlationID string) *ContextLogger {
	return &ContextLogger{
		Logger:        l.Logger.With(zap.String("correlation_id", correlationID)),
		correlationID: correlationID,
	}
}

// WithComponent creates a logger with component name
func (l *Logger) WithComponent(component string) *ContextLogger {
	return &ContextLogger{
		Logger:    l.Logger.With(zap.String("component", component)),
		component: component,
	}
}

// WithError creates a logger with error code and severity on base Logger
// This is the SRE-compliant way to log errors
func (l *Logger) WithError(errorCode, severity string) *ContextLogger {
	return &ContextLogger{
		Logger: l.Logger.With(
			zap.String("error_code", errorCode),
			zap.String("severity", severity),
		),
	}
}

// WithError creates a logger with error code and severity
// This is the SRE-compliant way to log errors
func (l *ContextLogger) WithError(errorCode, severity string) *ContextLogger {
	return &ContextLogger{
		Logger: l.Logger.With(
			zap.String("error_code", errorCode),
			zap.String("severity", severity),
		),
		correlationID: l.correlationID,
		component:     l.component,
	}
}

// WithComponent adds component to existing context logger
func (cl *ContextLogger) WithComponent(component string) *ContextLogger {
	return &ContextLogger{
		Logger:        cl.Logger.With(zap.String("component", component)),
		correlationID: cl.correlationID,
		component:     component,
	}
}

// WithCorrelation adds correlation ID to existing context logger
func (cl *ContextLogger) WithCorrelation(correlationID string) *ContextLogger {
	return &ContextLogger{
		Logger:        cl.Logger.With(zap.String("correlation_id", correlationID)),
		correlationID: correlationID,
		component:     cl.component,
	}
}

// GetCorrelationID returns the correlation ID from the context logger
func (cl *ContextLogger) GetCorrelationID() string {
	return cl.correlationID
}

// GetComponent returns the component name from the context logger
func (cl *ContextLogger) GetComponent() string {
	return cl.component
}

// GenerateCorrelationID creates a new correlation ID
// Format: {service}-{timestamp_ns}
func GenerateCorrelationID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, time.Now().UnixNano())
}
