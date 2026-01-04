package sli

import (
	"context"
	"time"
)

// Tracker defines the interface for tracking Service Level Indicators
type Tracker interface {
	// RecordRequest tracks a request with its outcome
	RecordRequest(ctx context.Context, outcome RequestOutcome)

	// RecordLatency tracks request latency
	RecordLatency(ctx context.Context, duration time.Duration, operation string)

	// RecordThroughput tracks throughput events (messages, requests)
	RecordThroughput(ctx context.Context, count int, operation string)

	// GetMetrics returns current SLI metrics snapshot
	GetMetrics(ctx context.Context) (*Metrics, error)
}

// BudgetTracker defines the interface for error budget tracking
type BudgetTracker interface {
	// CalculateBudget computes current error budget state
	CalculateBudget(ctx context.Context, window time.Duration) (*Budget, error)

	// GetBurnRate calculates current burn rate
	GetBurnRate(ctx context.Context, window time.Duration) (float64, error)

	// ShouldAlert determines if error budget alerts should fire
	ShouldAlert(ctx context.Context) ([]Alert, error)
}

// SLOValidator validates if SLOs are being met
type SLOValidator interface {
	// ValidateAvailability checks if availability SLO is met
	ValidateAvailability(ctx context.Context, window time.Duration) (*Compliance, error)

	// ValidateLatency checks if latency SLO is met
	ValidateLatency(ctx context.Context, window time.Duration) (*Compliance, error)

	// ValidateErrorRate checks if error rate SLO is met
	ValidateErrorRate(ctx context.Context, window time.Duration) (*Compliance, error)

	// GetCompliance returns overall SLO compliance
	GetCompliance(ctx context.Context, window time.Duration) (map[string]*Compliance, error)
}

// RequestOutcome represents the result of a request
type RequestOutcome struct {
	Success       bool
	ErrorCode     string
	ErrorSeverity string
	Latency       time.Duration
	Operation     string
	Timestamp     time.Time
	UserID        string
	DeviceID      string
}

// Metrics represents current SLI metrics
type Metrics struct {
	// Availability metrics
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	Availability    float64 // Percentage (0-100)

	// Latency metrics (milliseconds as float64)
	LatencyP50 float64
	LatencyP95 float64
	LatencyP99 float64
	LatencyAvg float64

	// Error rate metrics
	ErrorRate        float64 // Errors per second
	ErrorRatePercent float64 // Percentage (0-100)

	// Throughput metrics
	RequestsPerSecond float64
	MessagesPerSecond float64

	// Time window
	WindowStart time.Time
	WindowEnd   time.Time
}

// Budget represents error budget state
type Budget struct {
	// SLO target (e.g., 99.9% = 0.999)
	SLOTarget float64

	// Total budget for the period
	TotalRequests int64
	AllowedErrors int64 // Based on SLO target

	// Current consumption
	ActualErrors    int64
	ConsumedBudget  int64
	RemainingBudget int64
	BudgetPercent   float64 // Percentage remaining (0-100)

	// Burn rate analysis
	CurrentBurnRate   float64       // Errors per hour
	ProjectedBurnRate float64       // Projected over full period
	TimeToExhaustion  time.Duration // When budget will be exhausted

	// Period
	WindowStart time.Time
	WindowEnd   time.Time
	WindowSize  time.Duration
}

// Compliance represents SLO compliance status
type Compliance struct {
	SLI          string  // Name of the SLI (availability, latency_p95, etc.)
	Target       float64 // Target value
	Actual       float64 // Actual measured value
	InCompliance bool    // Whether target is met
	Margin       float64 // How much margin (positive = good, negative = violation)
	Window       time.Duration
}

// OverallCompliance represents overall SLO compliance across all SLIs
type OverallCompliance struct {
	ServiceName       string
	AllSLOsMet        bool
	Compliances       []Compliance
	ErrorBudget       *Budget
	OverallHealth     string // "HEALTHY", "AT_RISK", "CRITICAL"
	RecommendedAction string
	Timestamp         time.Time
}

// Alert represents an error budget alert
type Alert struct {
	Severity    string // "WARNING", "CRITICAL", "PAGE"
	Title       string
	Description string
	BurnRate    float64
	Window      time.Duration
	Threshold   float64
	Action      string // Recommended action
	Timestamp   time.Time
}

// SLOConfig represents SLO configuration for a service
type SLOConfig struct {
	ServiceName string
	Environment string

	// Availability SLO
	AvailabilitySLO AvailabilitySLO

	// Latency SLOs
	LatencySLO LatencySLO

	// Error Rate SLO
	ErrorRateSLO ErrorRateSLO

	// Error Budget Policy
	ErrorBudget ErrorBudgetConfig

	// Alert configuration
	Alerting AlertingConfig
}

// AvailabilitySLO defines availability targets
type AvailabilitySLO struct {
	Enabled bool
	Target  float64 // e.g., 0.999 for 99.9%
}

// LatencySLO defines latency targets
type LatencySLO struct {
	Enabled         bool
	P95Milliseconds int64 // P95 latency in milliseconds
	P99Milliseconds int64 // P99 latency in milliseconds
}

// ErrorRateSLO defines error rate targets
type ErrorRateSLO struct {
	Enabled      bool
	MaxErrorRate float64 // Maximum error rate (0-1)
}

// ErrorBudgetConfig defines error budget policy
type ErrorBudgetConfig struct {
	Window          time.Duration // Calculation window (e.g., 30 days)
	FastBurnWindow  time.Duration // Fast burn detection (e.g., 1 hour)
	FastBurnPercent float64       // % budget consumed in fast window to alert
	SlowBurnWindow  time.Duration // Slow burn detection (e.g., 6 hours)
	SlowBurnPercent float64       // % budget consumed in slow window to alert
}

// AlertingConfig defines alerting settings
type AlertingConfig struct {
	Enabled               bool
	PageOnBudgetCritical  bool
	NotifyOnBudgetWarning bool
}
