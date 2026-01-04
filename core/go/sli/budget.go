package sli

import (
	"context"
	"fmt"
	"time"
)

// budgetTracker implements BudgetTracker
type budgetTracker struct {
	tracker     Tracker
	config      *SLOConfig
	serviceName string
}

// NewBudgetTracker creates an error budget tracker
func NewBudgetTracker(tracker Tracker, config *SLOConfig, serviceName string) BudgetTracker {
	return &budgetTracker{
		tracker:     tracker,
		config:      config,
		serviceName: serviceName,
	}
}

// CalculateBudget computes error budget from REAL metrics
func (b *budgetTracker) CalculateBudget(ctx context.Context, window time.Duration) (*Budget, error) {
	// Get actual metrics from Prometheus
	metrics, err := b.tracker.GetMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Calculate error budget based on SLO target
	sloTarget := b.config.AvailabilitySLO.Target
	totalRequests := metrics.TotalRequests

	// Error Budget Formula (REAL DATA):
	// AllowedErrors = TotalRequests × (1 - SLO)
	// Example: 1,000,000 requests × (1 - 0.999) = 1,000 errors allowed
	allowedErrors := int64(float64(totalRequests) * (1.0 - sloTarget))
	actualErrors := metrics.FailedRequests
	remainingBudget := allowedErrors - actualErrors

	// Calculate burn rate (errors per hour)
	windowHours := window.Hours()
	if windowHours == 0 {
		windowHours = 1 // prevent division by zero
	}
	currentBurnRate := float64(actualErrors) / windowHours

	// Calculate time to exhaustion
	var timeToExhaustion time.Duration
	if currentBurnRate > 0 && remainingBudget > 0 {
		hoursRemaining := float64(remainingBudget) / currentBurnRate
		timeToExhaustion = time.Duration(hoursRemaining * float64(time.Hour))
	} else if remainingBudget <= 0 {
		timeToExhaustion = 0 // already exhausted
	} else {
		timeToExhaustion = time.Duration(999999 * time.Hour) // effectively infinite
	}

	// Calculate budget consumption percentage
	budgetPercent := 100.0
	if allowedErrors > 0 {
		budgetPercent = (float64(remainingBudget) / float64(allowedErrors)) * 100.0
	}

	return &Budget{
		SLOTarget:        sloTarget,
		WindowSize:       window,
		TotalRequests:    totalRequests,
		AllowedErrors:    allowedErrors,
		ActualErrors:     actualErrors,
		RemainingBudget:  remainingBudget,
		BudgetPercent:    budgetPercent,
		CurrentBurnRate:  currentBurnRate,
		TimeToExhaustion: timeToExhaustion,
	}, nil
}

// GetBurnRate calculates the current error budget burn rate
func (b *budgetTracker) GetBurnRate(ctx context.Context, window time.Duration) (float64, error) {
	budget, err := b.CalculateBudget(ctx, window)
	if err != nil {
		return 0, err
	}

	return budget.CurrentBurnRate, nil
}

// ShouldAlert checks if burn rate alerts should fire
func (b *budgetTracker) ShouldAlert(ctx context.Context) ([]Alert, error) {
	var alerts []Alert

	// Fast burn window (1 hour) - critical if 2% budget consumed
	fastBudget, err := b.CalculateBudget(ctx, b.config.ErrorBudget.FastBurnWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fast burn: %w", err)
	}

	fastBurnPercent := 100.0 - fastBudget.BudgetPercent
	if fastBurnPercent >= b.config.ErrorBudget.FastBurnPercent {
		alerts = append(alerts, Alert{
			Severity: "CRITICAL",
			Title:    "Fast Error Budget Burn Detected",
			Description: fmt.Sprintf(
				"CRITICAL: Fast burn detected! %.1f%% of error budget consumed in %s. "+
					"At current rate (%.1f errors/hour), budget will exhaust in %s.",
				fastBurnPercent,
				b.config.ErrorBudget.FastBurnWindow,
				fastBudget.CurrentBurnRate,
				fastBudget.TimeToExhaustion,
			),
			BurnRate:  fastBudget.CurrentBurnRate,
			Window:    b.config.ErrorBudget.FastBurnWindow,
			Threshold: b.config.ErrorBudget.FastBurnPercent,
			Action:    "Page on-call. Stop deployments immediately. Investigate error spike.",
			Timestamp: time.Now(),
		})
	}

	// Slow burn window (6 hours) - warning if 5% budget consumed
	slowBudget, err := b.CalculateBudget(ctx, b.config.ErrorBudget.SlowBurnWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate slow burn: %w", err)
	}

	slowBurnPercent := 100.0 - slowBudget.BudgetPercent
	if slowBurnPercent >= b.config.ErrorBudget.SlowBurnPercent {
		alerts = append(alerts, Alert{
			Severity: "WARNING",
			Title:    "Slow Error Budget Burn Detected",
			Description: fmt.Sprintf(
				"WARNING: Slow burn detected. %.1f%% of error budget consumed in %s. "+
					"At current rate (%.1f errors/hour), budget will exhaust in %s.",
				slowBurnPercent,
				b.config.ErrorBudget.SlowBurnWindow,
				slowBudget.CurrentBurnRate,
				slowBudget.TimeToExhaustion,
			),
			BurnRate:  slowBudget.CurrentBurnRate,
			Window:    b.config.ErrorBudget.SlowBurnWindow,
			Threshold: b.config.ErrorBudget.SlowBurnPercent,
			Action:    "Alert team. Review recent deployments. Monitor closely.",
			Timestamp: time.Now(),
		})
	}

	// Overall budget check - alert if less than 10% remaining
	overallBudget, err := b.CalculateBudget(ctx, b.config.ErrorBudget.Window)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate overall budget: %w", err)
	}

	if overallBudget.BudgetPercent < 10.0 {
		alerts = append(alerts, Alert{
			Severity: "WARNING",
			Title:    "Error Budget Low",
			Description: fmt.Sprintf(
				"WARNING: Error budget low. Only %.1f%% remaining in %s window. "+
					"Consider freezing non-critical deployments.",
				overallBudget.BudgetPercent,
				b.config.ErrorBudget.Window,
			),
			BurnRate:  overallBudget.CurrentBurnRate,
			Window:    b.config.ErrorBudget.Window,
			Threshold: 10.0,
			Action:    "Freeze non-critical deployments. Focus on reliability improvements.",
			Timestamp: time.Now(),
		})
	}

	return alerts, nil
}

// nopBudgetTracker is a no-op implementation
type nopBudgetTracker struct{}

// NewNopBudgetTracker creates a no-op budget tracker
func NewNopBudgetTracker() BudgetTracker {
	return &nopBudgetTracker{}
}

func (t *nopBudgetTracker) CalculateBudget(ctx context.Context, window time.Duration) (*Budget, error) {
	return &Budget{}, nil
}
func (t *nopBudgetTracker) GetBurnRate(ctx context.Context, window time.Duration) (float64, error) {
	return 0, nil
}
func (t *nopBudgetTracker) ShouldAlert(ctx context.Context) ([]Alert, error) {
	return nil, nil
}
