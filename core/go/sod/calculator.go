package sod

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// calculator implements the Calculator interface
type calculator struct {
	config           *Config
	runtimeFactors   RuntimeFactors
	mu               sync.RWMutex
	configLoader     ConfigLoader
	metricsCollector MetricsCollector
}

// NewCalculator creates a new SOD calculator with dependency injection
func NewCalculator(loader ConfigLoader, metrics MetricsCollector) (Calculator, error) {
	config, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load SOD config: %w", err)
	}

	calc := &calculator{
		config:           config,
		configLoader:     loader,
		metricsCollector: metrics,
		runtimeFactors: RuntimeFactors{
			LastUpdated: time.Now(),
		},
	}

	return calc, nil
}

// CalculateScore computes the SOD score based on error context and configuration
func (c *calculator) CalculateScore(ctx context.Context, errorCode string, errorContext ErrorContext) (Score, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	errorConfig, ok := c.config.Errors[errorCode]
	if !ok {
		return Score{}, fmt.Errorf("error code %s not found in SOD configuration", errorCode)
	}

	// Calculate base scores
	severity := c.calculateSeverity(errorConfig, errorContext)
	occurrence := c.calculateOccurrence(errorConfig, errorContext)
	detect := c.calculateDetect(errorConfig, errorContext)

	// Calculate total SOD score
	baseScore := severity * occurrence * detect

	// Apply runtime adjustments
	adjustmentFactor := c.calculateAdjustmentFactor(errorContext)
	adjustedScore := int(float64(baseScore) * adjustmentFactor)

	// Clamp to 0-1000 range
	if adjustedScore > 1000 {
		adjustedScore = 1000
	}
	if adjustedScore < 0 {
		adjustedScore = 0
	}

	score := Score{
		Total:            adjustedScore,
		Severity:         severity,
		Occurrence:       occurrence,
		Detect:           detect,
		Normalized:       float64(adjustedScore) / 1000.0,
		BaseScore:        baseScore,
		AdjustedScore:    adjustedScore,
		AdjustmentFactor: adjustmentFactor,
		SeverityReason:   c.getSeverityReason(errorConfig, errorContext),
		OccurrenceReason: c.getOccurrenceReason(errorConfig, errorContext),
		DetectReason:     c.getDetectReason(errorConfig, errorContext),
	}

	// Record metrics
	if c.metricsCollector != nil {
		c.metricsCollector.RecordSODScore(errorCode, score)
		c.metricsCollector.RecordErrorOccurrence(errorCode, c.severityToString(severity))
	}

	return score, nil
}

// calculateSeverity computes severity score with dynamic rules
func (c *calculator) calculateSeverity(cfg ErrorConfig, ctx ErrorContext) int {
	severity := cfg.BaseSeverity

	// Apply severity rules
	for _, rule := range cfg.SeverityRules {
		if c.evaluateCondition(rule.Condition, ctx) {
			if rule.Override != nil {
				severity = *rule.Override
			} else {
				severity = int(float64(severity) * rule.Multiplier)
			}
		}
	}

	// Clamp to 1-10
	if severity > 10 {
		severity = 10
	}
	if severity < 1 {
		severity = 1
	}

	return severity
}

// calculateOccurrence computes occurrence score based on patterns
func (c *calculator) calculateOccurrence(cfg ErrorConfig, ctx ErrorContext) int {
	occurrence := cfg.BaseOccurrence

	// Check occurrence patterns
	for _, pattern := range cfg.OccurrencePatterns {
		switch pattern.Type {
		case "rate":
			if ctx.RecentErrorRate >= pattern.Threshold {
				occurrence = pattern.Score
			}
		case "burst":
			// Burst detection would require time-series data
			// For now, use error rate as proxy
			if ctx.RecentErrorRate > pattern.Threshold*2 {
				occurrence = pattern.Score
			}
		case "trend":
			// Trend analysis would require historical data
			// Placeholder for future implementation
		}
	}

	// Clamp to 1-10
	if occurrence > 10 {
		occurrence = 10
	}
	if occurrence < 1 {
		occurrence = 1
	}

	return occurrence
}

// calculateDetect computes detectability score
func (c *calculator) calculateDetect(cfg ErrorConfig, ctx ErrorContext) int {
	detect := cfg.BaseDetect

	// Adjust based on detection configuration
	if !cfg.DetectionConfig.MonitoringEnabled {
		detect = 10 // Hard to detect without monitoring
	} else if !cfg.DetectionConfig.AlertingEnabled {
		detect = max(detect, 7) // Harder without alerts
	} else if cfg.DetectionConfig.AutoDetect {
		detect = min(detect, 3) // Easy to detect with automation
	}

	// Clamp to 1-10
	if detect > 10 {
		detect = 10
	}
	if detect < 1 {
		detect = 1
	}

	return detect
}

// calculateAdjustmentFactor applies runtime adjustments
func (c *calculator) calculateAdjustmentFactor(ctx ErrorContext) float64 {
	factor := 1.0

	// Environment multiplier
	switch ctx.Environment {
	case "production":
		factor *= 1.5 // Production errors are more severe
	case "staging":
		factor *= 1.0
	case "dev":
		factor *= 0.5 // Dev errors are less critical
	}

	// Business hours multiplier
	if ctx.IsBusinessHours {
		factor *= 1.3 // Errors during business hours have higher impact
	}

	// System load multiplier
	if ctx.SystemLoad > c.config.RuntimeSettings.LoadThresholds.High {
		factor *= 1.4 // High load amplifies error impact
	} else if ctx.SystemLoad > c.config.RuntimeSettings.LoadThresholds.Medium {
		factor *= 1.2
	}

	// Error rate multiplier
	if ctx.RecentErrorRate > 10 {
		factor *= 1.3 // Error storms are more severe
	} else if ctx.RecentErrorRate > 5 {
		factor *= 1.1
	}

	// Runtime factors from global state
	if c.runtimeFactors.SystemDegraded {
		factor *= 1.5 // Errors during degradation compound
	}

	if c.runtimeFactors.MaintenanceMode {
		factor *= 0.7 // Errors during maintenance are expected
	}

	return factor
}

// evaluateCondition checks if a condition is met
func (c *calculator) evaluateCondition(condition string, ctx ErrorContext) bool {
	switch condition {
	case "business_hours":
		return ctx.IsBusinessHours
	case "high_load":
		return ctx.SystemLoad > c.config.RuntimeSettings.LoadThresholds.High
	case "medium_load":
		return ctx.SystemLoad > c.config.RuntimeSettings.LoadThresholds.Medium
	case "production":
		return ctx.Environment == "production"
	case "staging":
		return ctx.Environment == "staging"
	case "dev":
		return ctx.Environment == "dev"
	case "error_storm":
		return ctx.RecentErrorRate > 10
	case "night_hours":
		return ctx.TimeOfDay < 6 || ctx.TimeOfDay > 22
	default:
		return false
	}
}

// GetErrorConfig retrieves the SOD configuration for a specific error code
func (c *calculator) GetErrorConfig(errorCode string) (*ErrorConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cfg, ok := c.config.Errors[errorCode]
	if !ok {
		return nil, fmt.Errorf("error code %s not found", errorCode)
	}

	return &cfg, nil
}

// UpdateRuntimeFactors updates dynamic runtime factors
func (c *calculator) UpdateRuntimeFactors(factors RuntimeFactors) {
	c.mu.Lock()
	defer c.mu.Unlock()

	factors.LastUpdated = time.Now()
	c.runtimeFactors = factors
}

// Helper functions
func (c *calculator) getSeverityReason(cfg ErrorConfig, ctx ErrorContext) string {
	reason := fmt.Sprintf("Base severity: %d", cfg.BaseSeverity)

	for _, rule := range cfg.SeverityRules {
		if c.evaluateCondition(rule.Condition, ctx) {
			if rule.Override != nil {
				reason += fmt.Sprintf("; Overridden to %d (%s)", *rule.Override, rule.Condition)
			} else {
				reason += fmt.Sprintf("; Multiplier %.1fx (%s)", rule.Multiplier, rule.Condition)
			}
		}
	}

	return reason
}

func (c *calculator) getOccurrenceReason(cfg ErrorConfig, ctx ErrorContext) string {
	reason := fmt.Sprintf("Base occurrence: %d", cfg.BaseOccurrence)

	for _, pattern := range cfg.OccurrencePatterns {
		if pattern.Type == "rate" && ctx.RecentErrorRate >= pattern.Threshold {
			reason += fmt.Sprintf("; Rate %.2f/s exceeds %.2f threshold", ctx.RecentErrorRate, pattern.Threshold)
		}
	}

	return reason
}

func (c *calculator) getDetectReason(cfg ErrorConfig, ctx ErrorContext) string {
	reason := fmt.Sprintf("Base detect: %d", cfg.BaseDetect)

	if !cfg.DetectionConfig.MonitoringEnabled {
		reason += "; No monitoring"
	} else if cfg.DetectionConfig.AutoDetect {
		reason += "; Auto-detection enabled"
	}

	return reason
}

func (c *calculator) severityToString(severity int) string {
	if severity >= 9 {
		return "critical"
	} else if severity >= 7 {
		return "high"
	} else if severity >= 4 {
		return "medium"
	}
	return "low"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
