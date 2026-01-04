package sod

import (
	"context"
	"time"
)

// Calculator defines the interface for SOD score calculation
type Calculator interface {
	// CalculateScore computes the SOD score based on error context and configuration
	CalculateScore(ctx context.Context, errorCode string, errorContext ErrorContext) (Score, error)

	// GetErrorConfig retrieves the SOD configuration for a specific error code
	GetErrorConfig(errorCode string) (*ErrorConfig, error)

	// UpdateRuntimeFactors updates dynamic runtime factors (load, error rate, etc.)
	UpdateRuntimeFactors(factors RuntimeFactors)
}

// ConfigLoader defines the interface for loading SOD configuration
type ConfigLoader interface {
	// Load reads SOD configuration from a source (file, remote config, etc.)
	Load() (*Config, error)

	// Reload refreshes the configuration (for hot-reload scenarios)
	Reload() error

	// Watch monitors configuration changes and triggers reload
	Watch(ctx context.Context, callback func(*Config)) error
}

// MetricsCollector defines the interface for SOD metrics
type MetricsCollector interface {
	// RecordSODScore records an SOD score for an error
	RecordSODScore(errorCode string, score Score)

	// RecordErrorOccurrence tracks error occurrence for rate calculation
	RecordErrorOccurrence(errorCode string, severity string)

	// RecordMTTD tracks Mean Time To Detect
	RecordMTTD(errorCode string, duration time.Duration)

	// RecordMTTR tracks Mean Time To Resolve
	RecordMTTR(errorCode string, duration time.Duration)
}

// ErrorContext contains runtime context for SOD calculation
type ErrorContext struct {
	Timestamp       time.Time
	ServiceName     string
	Environment     string // dev, staging, production
	UserID          string
	DeviceID        string
	RequestPath     string
	SystemLoad      float64 // 0-1 (CPU/Memory average)
	RecentErrorRate float64 // errors per second in last 5 minutes
	TimeOfDay       int     // hour 0-23
	IsBusinessHours bool
	CustomTags      map[string]string
}

// RuntimeFactors contains dynamic factors affecting SOD scores
type RuntimeFactors struct {
	CurrentLoad     float64 // 0-1
	ErrorRate       float64 // errors/second
	ActiveUsers     int
	SystemDegraded  bool
	MaintenanceMode bool
	LastUpdated     time.Time
}

// Score represents a calculated SOD score with breakdown
type Score struct {
	Total      int     // 0-1000 (Severity * Occurrence * Detectability)
	Severity   int     // 1-10
	Occurrence int     // 1-10
	Detect     int     // 1-10
	Normalized float64 // 0-1 normalized score

	// Breakdown for transparency
	SeverityReason   string
	OccurrenceReason string
	DetectReason     string

	// Runtime adjustments
	BaseScore        int
	AdjustedScore    int
	AdjustmentFactor float64
}

// ErrorConfig defines SOD configuration for a specific error
type ErrorConfig struct {
	Code        string
	Description string

	// Base scores (static configuration)
	BaseSeverity   int
	BaseOccurrence int
	BaseDetect     int

	// Dynamic severity rules (adjust based on context)
	SeverityRules []SeverityRule

	// Occurrence patterns
	OccurrencePatterns []OccurrencePattern

	// Detection configuration
	DetectionConfig DetectionConfig

	// Thresholds
	Thresholds Thresholds
}

// SeverityRule adjusts severity based on conditions
type SeverityRule struct {
	Condition  string  // "business_hours", "high_load", "production", etc.
	Multiplier float64 // severity multiplier
	Override   *int    // optional: override severity completely
}

// OccurrencePattern defines how to calculate occurrence score
type OccurrencePattern struct {
	Type      string  // "rate", "burst", "trend"
	Threshold float64 // threshold for this pattern
	Score     int     // occurrence score if threshold exceeded
}

// DetectionConfig defines how quickly errors are detected
type DetectionConfig struct {
	MonitoringEnabled bool
	AlertingEnabled   bool
	AutoDetect        bool
	MTTDTarget        time.Duration // Mean Time To Detect target
	LogLevel          string        // ERROR, WARN, INFO
}

// Thresholds defines SOD score thresholds for alerting
type Thresholds struct {
	Critical int // SOD score threshold for critical alerts
	High     int // SOD score threshold for high priority
	Medium   int // SOD score threshold for medium priority
	Low      int // SOD score threshold for low priority
}

// Config represents the complete SOD configuration
type Config struct {
	ServiceName string
	Environment string
	Version     string

	// Global settings
	GlobalThresholds Thresholds

	// Per-error configurations
	Errors map[string]ErrorConfig

	// Runtime factor settings
	RuntimeSettings RuntimeSettings
}

// RuntimeSettings configures runtime factor calculations
type RuntimeSettings struct {
	LoadThresholds struct {
		Low    float64 // < 0.5
		Medium float64 // 0.5-0.75
		High   float64 // > 0.75
	}
	ErrorRateWindow time.Duration // window for error rate calculation
	BusinessHours   BusinessHours
}

// BusinessHours defines business hours for severity adjustments
type BusinessHours struct {
	Enabled   bool
	StartHour int // 0-23
	EndHour   int // 0-23
	Weekdays  []time.Weekday
	Timezone  string
}
