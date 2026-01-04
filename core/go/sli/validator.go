package sli

import (
	"context"
	"fmt"
	"time"
)

// sloValidator implements SLOValidator
type sloValidator struct {
	tracker     Tracker
	config      *SLOConfig
	serviceName string
}

// NewSLOValidator creates an SLO compliance validator
func NewSLOValidator(tracker Tracker, config *SLOConfig, serviceName string) SLOValidator {
	return &sloValidator{
		tracker:     tracker,
		config:      config,
		serviceName: serviceName,
	}
}

// ValidateAvailability checks if availability SLO is met
func (v *sloValidator) ValidateAvailability(ctx context.Context, window time.Duration) (*Compliance, error) {
	if !v.config.AvailabilitySLO.Enabled {
		return &Compliance{
			SLI:          "availability",
			Target:       0,
			Actual:       0,
			InCompliance: true,
			Margin:       0,
			Window:       window,
		}, nil
	}

	metrics, err := v.tracker.GetMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Calculate actual availability from REAL metrics
	// Availability = SuccessRequests / TotalRequests
	var availability float64
	if metrics.TotalRequests > 0 {
		availability = float64(metrics.SuccessRequests) / float64(metrics.TotalRequests)
	}

	target := v.config.AvailabilitySLO.Target
	inCompliance := availability >= target
	margin := availability - target

	return &Compliance{
		SLI:          "availability",
		Window:       window,
		Target:       target,
		Actual:       availability,
		InCompliance: inCompliance,
		Margin:       margin,
	}, nil
}

// ValidateLatency checks if latency SLO is met
func (v *sloValidator) ValidateLatency(ctx context.Context, window time.Duration) (*Compliance, error) {
	if !v.config.LatencySLO.Enabled {
		return &Compliance{
			SLI:          "latency",
			Target:       0,
			Actual:       0,
			InCompliance: true,
			Margin:       0,
			Window:       window,
		}, nil
	}

	metrics, err := v.tracker.GetMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Check P95 latency target
	p95TargetMs := float64(v.config.LatencySLO.P95Milliseconds)
	p95ActualMs := metrics.LatencyP95

	inCompliance := p95ActualMs <= p95TargetMs
	margin := p95TargetMs - p95ActualMs

	// Also check P99 if configured
	if v.config.LatencySLO.P99Milliseconds > 0 {
		p99TargetMs := float64(v.config.LatencySLO.P99Milliseconds)
		p99ActualMs := metrics.LatencyP99

		if p99ActualMs > p99TargetMs {
			inCompliance = false
		}
	}

	return &Compliance{
		SLI:          "latency_p95",
		Window:       window,
		Target:       p95TargetMs,
		Actual:       p95ActualMs,
		InCompliance: inCompliance,
		Margin:       margin,
	}, nil
}

// ValidateErrorRate checks if error rate SLO is met
func (v *sloValidator) ValidateErrorRate(ctx context.Context, window time.Duration) (*Compliance, error) {
	if !v.config.ErrorRateSLO.Enabled {
		return &Compliance{
			SLI:          "error_rate",
			Target:       0,
			Actual:       0,
			InCompliance: true,
			Margin:       0,
			Window:       window,
		}, nil
	}

	metrics, err := v.tracker.GetMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Calculate actual error rate from REAL metrics
	// ErrorRate = FailedRequests / TotalRequests
	var errorRate float64
	if metrics.TotalRequests > 0 {
		errorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	}

	target := v.config.ErrorRateSLO.MaxErrorRate
	inCompliance := errorRate <= target
	margin := target - errorRate

	return &Compliance{
		SLI:          "error_rate",
		Window:       window,
		Target:       target,
		Actual:       errorRate,
		InCompliance: inCompliance,
		Margin:       margin,
	}, nil
}

// GetCompliance returns overall SLO compliance status
func (v *sloValidator) GetCompliance(ctx context.Context, window time.Duration) (map[string]*Compliance, error) {
	results := make(map[string]*Compliance)

	// Check availability
	availComp, err := v.ValidateAvailability(ctx, window)
	if err != nil {
		return nil, fmt.Errorf("availability check failed: %w", err)
	}
	results["availability"] = availComp

	// Check latency
	latencyComp, err := v.ValidateLatency(ctx, window)
	if err != nil {
		return nil, fmt.Errorf("latency check failed: %w", err)
	}
	results["latency"] = latencyComp

	// Check error rate
	errorComp, err := v.ValidateErrorRate(ctx, window)
	if err != nil {
		return nil, fmt.Errorf("error rate check failed: %w", err)
	}
	results["error_rate"] = errorComp

	return results, nil
}

// nopValidator is a no-op implementation
type nopValidator struct{}

// NewNopValidator creates a no-op validator
func NewNopValidator() SLOValidator {
	return &nopValidator{}
}

func (v *nopValidator) ValidateAvailability(ctx context.Context, window time.Duration) (*Compliance, error) {
	return &Compliance{InCompliance: true}, nil
}
func (v *nopValidator) ValidateLatency(ctx context.Context, window time.Duration) (*Compliance, error) {
	return &Compliance{InCompliance: true}, nil
}
func (v *nopValidator) ValidateErrorRate(ctx context.Context, window time.Duration) (*Compliance, error) {
	return &Compliance{InCompliance: true}, nil
}
func (v *nopValidator) GetCompliance(ctx context.Context, window time.Duration) (map[string]*Compliance, error) {
	return map[string]*Compliance{}, nil
}
