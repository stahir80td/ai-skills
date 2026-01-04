package sod

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Factory creates and configures SOD components with dependency injection
type Factory struct {
	serviceName string
	environment string
}

// NewFactory creates a new SOD factory
func NewFactory(serviceName, environment string) *Factory {
	return &Factory{
		serviceName: serviceName,
		environment: environment,
	}
}

// CreateCalculator creates a fully configured SOD calculator
func (f *Factory) CreateCalculator(configPath string) (Calculator, error) {
	// Create config loader
	loader := NewFileConfigLoader(configPath)

	// Create metrics collector
	metrics := NewPrometheusMetrics(f.serviceName)

	// Create calculator with dependencies
	calc, err := NewCalculator(loader, metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOD calculator: %w", err)
	}

	return calc, nil
}

// CreateCalculatorWithAutoConfig creates calculator with automatic config discovery
func (f *Factory) CreateCalculatorWithAutoConfig() (Calculator, error) {
	configPath := f.discoverConfigPath()
	return f.CreateCalculator(configPath)
}

// CreateCalculatorWithEnv creates calculator using environment variables
func (f *Factory) CreateCalculatorWithEnv() (Calculator, error) {
	loader := NewEnvConfigLoader(f.serviceName, f.environment)
	metrics := NewPrometheusMetrics(f.serviceName)

	calc, err := NewCalculator(loader, metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOD calculator: %w", err)
	}

	return calc, nil
}

// discoverConfigPath attempts to find SOD config file
func (f *Factory) discoverConfigPath() string {
	// Try environment variable first
	if path := os.Getenv("SOD_CONFIG_PATH"); path != "" {
		return path
	}

	// Try common locations
	candidates := []string{
		filepath.Join("config", fmt.Sprintf("sod_%s.yaml", f.serviceName)),
		filepath.Join("config", "sod.yaml"),
		fmt.Sprintf("sod_%s.yaml", f.serviceName),
		"sod.yaml",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Fallback to default
	return filepath.Join("config", "sod.yaml")
}

// Helper creates a helper for easier SOD score calculation
type Helper struct {
	calculator  Calculator
	serviceName string
	environment string
}

// NewHelper creates a SOD helper
func NewHelper(calculator Calculator, serviceName, environment string) *Helper {
	return &Helper{
		calculator:  calculator,
		serviceName: serviceName,
		environment: environment,
	}
}

// CalculateForError calculates SOD score for an error with minimal context
func (h *Helper) CalculateForError(ctx context.Context, errorCode string) (Score, error) {
	errorContext := h.buildErrorContext(ctx)
	return h.calculator.CalculateScore(ctx, errorCode, errorContext)
}

// CalculateWithContext calculates SOD score with full context
func (h *Helper) CalculateWithContext(ctx context.Context, errorCode string, errorContext ErrorContext) (Score, error) {
	// Merge with auto-detected context
	autoContext := h.buildErrorContext(ctx)

	// Override with provided values
	if errorContext.ServiceName == "" {
		errorContext.ServiceName = autoContext.ServiceName
	}
	if errorContext.Environment == "" {
		errorContext.Environment = autoContext.Environment
	}
	if errorContext.Timestamp.IsZero() {
		errorContext.Timestamp = autoContext.Timestamp
	}
	errorContext.SystemLoad = autoContext.SystemLoad
	errorContext.TimeOfDay = autoContext.TimeOfDay
	errorContext.IsBusinessHours = autoContext.IsBusinessHours

	return h.calculator.CalculateScore(ctx, errorCode, errorContext)
}

// buildErrorContext builds error context from available information
func (h *Helper) buildErrorContext(ctx context.Context) ErrorContext {
	now := time.Now()

	errorContext := ErrorContext{
		Timestamp:   now,
		ServiceName: h.serviceName,
		Environment: h.environment,
		TimeOfDay:   now.Hour(),
		CustomTags:  make(map[string]string),
	}

	// Extract from context if available
	if ctx != nil {
		if userID, ok := ctx.Value("user_id").(string); ok {
			errorContext.UserID = userID
		}
		if deviceID, ok := ctx.Value("device_id").(string); ok {
			errorContext.DeviceID = deviceID
		}
		if requestPath, ok := ctx.Value("request_path").(string); ok {
			errorContext.RequestPath = requestPath
		}
	}

	// Detect business hours (Mon-Fri, 9am-5pm)
	weekday := now.Weekday()
	hour := now.Hour()
	errorContext.IsBusinessHours = weekday >= time.Monday &&
		weekday <= time.Friday &&
		hour >= 9 && hour < 17

	// Get system load (basic implementation)
	errorContext.SystemLoad = h.getSystemLoad()

	return errorContext
}

// getSystemLoad returns current system load (simplified)
func (h *Helper) getSystemLoad() float64 {
	// In production, use actual system metrics
	// For now, return a placeholder
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Simple heuristic: if using > 75% of allocated memory, consider high load
	usage := float64(m.Alloc) / float64(m.Sys)
	if usage > 1.0 {
		usage = 1.0
	}

	return usage
}

// GetSeverityLevel returns human-readable severity level
func GetSeverityLevel(score Score) string {
	if score.Total >= 700 {
		return "CRITICAL"
	} else if score.Total >= 500 {
		return "HIGH"
	} else if score.Total >= 300 {
		return "MEDIUM"
	}
	return "LOW"
}

// ShouldAlert determines if an error should trigger an alert
func ShouldAlert(score Score, thresholds Thresholds) bool {
	return score.Total >= thresholds.High
}

// ShouldPage determines if an error should page on-call
func ShouldPage(score Score, thresholds Thresholds) bool {
	return score.Total >= thresholds.Critical
}

// FormatScoreDetails formats SOD score details for logging
func FormatScoreDetails(score Score) string {
	return fmt.Sprintf(
		"SOD=%d (S=%d, O=%d, D=%d) [%s] - Severity: %s | Occurrence: %s | Detect: %s",
		score.Total,
		score.Severity,
		score.Occurrence,
		score.Detect,
		GetSeverityLevel(score),
		score.SeverityReason,
		score.OccurrenceReason,
		score.DetectReason,
	)
}
