package sli

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigLoader handles SLO configuration loading
type ConfigLoader interface {
	Load() (*SLOConfig, error)
}

// fileConfigLoader loads SLO config from YAML file
type fileConfigLoader struct {
	filePath string
}

// NewFileConfigLoader creates a file-based config loader
func NewFileConfigLoader(filePath string) ConfigLoader {
	return &fileConfigLoader{filePath: filePath}
}

// Load reads and parses the SLO configuration file
func (l *fileConfigLoader) Load() (*SLOConfig, error) {
	data, err := os.ReadFile(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", l.filePath, err)
	}

	var config SLOConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig ensures SLO configuration is valid
func validateConfig(config *SLOConfig) error {
	if config.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}

	// Validate availability SLO
	if config.AvailabilitySLO.Enabled {
		if config.AvailabilitySLO.Target <= 0 || config.AvailabilitySLO.Target > 1 {
			return fmt.Errorf("availability_slo.target must be between 0 and 1 (e.g., 0.999 for 99.9%%)")
		}
	}

	// Validate latency SLO
	if config.LatencySLO.Enabled {
		if config.LatencySLO.P95Milliseconds <= 0 {
			return fmt.Errorf("latency_slo.p95_milliseconds must be positive")
		}
		if config.LatencySLO.P99Milliseconds < 0 {
			return fmt.Errorf("latency_slo.p99_milliseconds cannot be negative")
		}
	}

	// Validate error rate SLO
	if config.ErrorRateSLO.Enabled {
		if config.ErrorRateSLO.MaxErrorRate < 0 || config.ErrorRateSLO.MaxErrorRate > 1 {
			return fmt.Errorf("error_rate_slo.max_error_rate must be between 0 and 1")
		}
	}

	// Validate error budget
	if config.ErrorBudget.Window <= 0 {
		return fmt.Errorf("error_budget.window must be positive")
	}
	if config.ErrorBudget.FastBurnWindow <= 0 {
		return fmt.Errorf("error_budget.fast_burn_window must be positive")
	}
	if config.ErrorBudget.SlowBurnWindow <= 0 {
		return fmt.Errorf("error_budget.slow_burn_window must be positive")
	}
	if config.ErrorBudget.FastBurnPercent <= 0 {
		return fmt.Errorf("error_budget.fast_burn_percent must be positive")
	}
	if config.ErrorBudget.SlowBurnPercent <= 0 {
		return fmt.Errorf("error_budget.slow_burn_percent must be positive")
	}

	return nil
}

// envConfigLoader loads SLO config from environment variables
type envConfigLoader struct {
	serviceName string
}

// NewEnvConfigLoader creates an environment-based config loader
func NewEnvConfigLoader(serviceName string) ConfigLoader {
	return &envConfigLoader{serviceName: serviceName}
}

// Load reads SLO configuration from environment variables
func (l *envConfigLoader) Load() (*SLOConfig, error) {
	config := &SLOConfig{
		ServiceName: l.serviceName,
	}

	// Availability SLO
	if target := os.Getenv("SLO_AVAILABILITY_TARGET"); target != "" {
		var val float64
		if _, err := fmt.Sscanf(target, "%f", &val); err != nil {
			return nil, fmt.Errorf("invalid SLO_AVAILABILITY_TARGET: %w", err)
		}
		config.AvailabilitySLO.Enabled = true
		config.AvailabilitySLO.Target = val
	}

	// Latency SLO
	if p95 := os.Getenv("SLO_LATENCY_P95_MS"); p95 != "" {
		var val int64
		if _, err := fmt.Sscanf(p95, "%d", &val); err != nil {
			return nil, fmt.Errorf("invalid SLO_LATENCY_P95_MS: %w", err)
		}
		config.LatencySLO.Enabled = true
		config.LatencySLO.P95Milliseconds = val
	}

	if p99 := os.Getenv("SLO_LATENCY_P99_MS"); p99 != "" {
		var val int64
		if _, err := fmt.Sscanf(p99, "%d", &val); err != nil {
			return nil, fmt.Errorf("invalid SLO_LATENCY_P99_MS: %w", err)
		}
		config.LatencySLO.P99Milliseconds = val
	}

	// Error rate SLO
	if maxErr := os.Getenv("SLO_ERROR_RATE_MAX"); maxErr != "" {
		var val float64
		if _, err := fmt.Sscanf(maxErr, "%f", &val); err != nil {
			return nil, fmt.Errorf("invalid SLO_ERROR_RATE_MAX: %w", err)
		}
		config.ErrorRateSLO.Enabled = true
		config.ErrorRateSLO.MaxErrorRate = val
	}

	// Error budget (with defaults)
	config.ErrorBudget = ErrorBudgetConfig{
		Window:          30 * 24 * time.Hour, // 30 days default
		FastBurnWindow:  1 * time.Hour,
		SlowBurnWindow:  6 * time.Hour,
		FastBurnPercent: 2.0,
		SlowBurnPercent: 5.0,
	}

	if window := os.Getenv("SLO_ERROR_BUDGET_WINDOW_HOURS"); window != "" {
		var hours int
		if _, err := fmt.Sscanf(window, "%d", &hours); err != nil {
			return nil, fmt.Errorf("invalid SLO_ERROR_BUDGET_WINDOW_HOURS: %w", err)
		}
		config.ErrorBudget.Window = time.Duration(hours) * time.Hour
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// defaultConfigLoader provides sensible defaults
type defaultConfigLoader struct {
	serviceName string
}

// NewDefaultConfigLoader creates a config loader with default values
func NewDefaultConfigLoader(serviceName string) ConfigLoader {
	return &defaultConfigLoader{serviceName: serviceName}
}

// Load returns default SLO configuration
func (l *defaultConfigLoader) Load() (*SLOConfig, error) {
	return &SLOConfig{
		ServiceName: l.serviceName,
		AvailabilitySLO: AvailabilitySLO{
			Enabled: true,
			Target:  0.999, // 99.9%
		},
		LatencySLO: LatencySLO{
			Enabled:         true,
			P95Milliseconds: 200,
			P99Milliseconds: 500,
		},
		ErrorRateSLO: ErrorRateSLO{
			Enabled:      true,
			MaxErrorRate: 0.001, // 0.1%
		},
		ErrorBudget: ErrorBudgetConfig{
			Window:          30 * 24 * time.Hour, // 30 days
			FastBurnWindow:  1 * time.Hour,
			FastBurnPercent: 2.0,
			SlowBurnWindow:  6 * time.Hour,
			SlowBurnPercent: 5.0,
		},
	}, nil
}
