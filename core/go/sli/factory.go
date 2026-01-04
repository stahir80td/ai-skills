package sli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Factory creates SLI/SLO components
type Factory struct {
	serviceName string
	environment string
}

// NewFactory creates a new SLI/SLO factory
func NewFactory(serviceName, environment string) *Factory {
	return &Factory{
		serviceName: serviceName,
		environment: environment,
	}
}

// CreateTracker creates a Prometheus-based tracker
func (f *Factory) CreateTracker() Tracker {
	return NewPrometheusTracker(f.serviceName)
}

// CreateConfigLoader creates a config loader based on environment
func (f *Factory) CreateConfigLoader() ConfigLoader {
	// Try file-based config first
	configPath := filepath.Join("config", "slo.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return NewFileConfigLoader(configPath)
	}

	// Fall back to environment variables
	if os.Getenv("SLO_AVAILABILITY_TARGET") != "" {
		return NewEnvConfigLoader(f.serviceName)
	}

	// Use defaults
	return NewDefaultConfigLoader(f.serviceName)
}

// CreateBudgetTracker creates an error budget tracker
func (f *Factory) CreateBudgetTracker(tracker Tracker, config *SLOConfig) BudgetTracker {
	return NewBudgetTracker(tracker, config, f.serviceName)
}

// CreateValidator creates an SLO validator
func (f *Factory) CreateValidator(tracker Tracker, config *SLOConfig) SLOValidator {
	return NewSLOValidator(tracker, config, f.serviceName)
}

// CreateAll creates all SLI/SLO components with auto-configuration
func (f *Factory) CreateAll() (*SLIManager, error) {
	// Create tracker
	tracker := f.CreateTracker()

	// Load configuration
	configLoader := f.CreateConfigLoader()
	config, err := configLoader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load SLO config: %w", err)
	}

	// Create budget tracker and validator
	budgetTracker := f.CreateBudgetTracker(tracker, config)
	validator := f.CreateValidator(tracker, config)

	return &SLIManager{
		Tracker:       tracker,
		BudgetTracker: budgetTracker,
		Validator:     validator,
		Config:        config,
	}, nil
}

// SLIManager bundles all SLI/SLO components
type SLIManager struct {
	Tracker       Tracker
	BudgetTracker BudgetTracker
	Validator     SLOValidator
	Config        *SLOConfig
}

// Helper provides simplified SLI tracking
type Helper struct {
	manager *SLIManager
}

// NewHelper creates a simplified SLI helper
func NewHelper(manager *SLIManager) *Helper {
	return &Helper{manager: manager}
}

// TrackRequest is a convenience method for tracking requests
func (h *Helper) TrackRequest(outcome RequestOutcome) {
	h.manager.Tracker.RecordRequest(context.TODO(), outcome)
}

// CheckBudget is a convenience method for checking error budget
func (h *Helper) CheckBudget() (*Budget, error) {
	return h.manager.BudgetTracker.CalculateBudget(context.TODO(), h.manager.Config.ErrorBudget.Window)
}

// CheckCompliance is a convenience method for checking SLO compliance
func (h *Helper) CheckCompliance() (map[string]*Compliance, error) {
	return h.manager.Validator.GetCompliance(context.TODO(), h.manager.Config.ErrorBudget.Window)
}

// ShouldAlert checks if error budget alerts should fire
func (h *Helper) ShouldAlert() ([]Alert, error) {
	return h.manager.BudgetTracker.ShouldAlert(context.TODO())
}

// Utility functions

// IsCriticalBurnRate returns true if burn rate is critical (fast burn)
func IsCriticalBurnRate(alerts []Alert) bool {
	for _, alert := range alerts {
		if alert.Severity == "critical" {
			return true
		}
	}
	return false
}

// IsWarningBurnRate returns true if burn rate is warning (slow burn)
func IsWarningBurnRate(alerts []Alert) bool {
	for _, alert := range alerts {
		if alert.Severity == "warning" {
			return true
		}
	}
	return false
}

// GetBudgetStatus returns a human-readable budget status
func GetBudgetStatus(budget *Budget) string {
	if budget.BudgetPercent >= 50 {
		return "healthy"
	} else if budget.BudgetPercent >= 25 {
		return "warning"
	} else if budget.BudgetPercent >= 10 {
		return "critical"
	}
	return "exhausted"
}

// FormatBudget returns a formatted budget summary
func FormatBudget(budget *Budget) string {
	return fmt.Sprintf(
		"Error Budget: %.1f%% remaining (%d/%d errors). Burn rate: %.1f errors/hr. Time to exhaustion: %s",
		budget.BudgetPercent,
		budget.RemainingBudget,
		budget.AllowedErrors,
		budget.CurrentBurnRate,
		budget.TimeToExhaustion,
	)
}
