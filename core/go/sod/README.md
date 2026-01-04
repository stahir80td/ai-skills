# SOD Framework

A reusable **Severity × Occurrence × Detectability** framework for dynamic error prioritization in distributed systems.

## Overview

The SOD Framework provides a systematic approach to prioritizing errors based on three dimensions:
- **Severity (S)**: Impact of the error (1-10)
- **Occurrence (O)**: How often the error happens (1-10)  
- **Detectability (D)**: How hard it is to detect (1-10)

**SOD Score** = S × O × D (0-1000)

Higher scores indicate errors that need immediate attention.

## Architecture

### Core Components

1. **Calculator** - Computes SOD scores with dynamic runtime adjustments
2. **ConfigLoader** - Loads SOD configuration from YAML files or environment
3. **MetricsCollector** - Records SOD scores and metrics to Prometheus
4. **Factory** - Creates and wires components with dependency injection
5. **Helper** - Simplifies SOD score calculation in applications

### Interfaces (Dependency Injection)

```go
type Calculator interface {
    CalculateScore(ctx context.Context, errorCode string, errorContext ErrorContext) (Score, error)
    GetErrorConfig(errorCode string) (*ErrorConfig, error)
    UpdateRuntimeFactors(factors RuntimeFactors)
}

type ConfigLoader interface {
    Load() (*Config, error)
    Reload() error
    Watch(ctx context.Context, callback func(*Config)) error
}

type MetricsCollector interface {
    RecordSODScore(errorCode string, score Score)
    RecordErrorOccurrence(errorCode string, severity string)
    RecordMTTD(errorCode string, duration time.Duration)
    RecordMTTR(errorCode string, duration time.Duration)
}
```

## Quick Start

### 1. Create SOD Configuration

Create `config/sod.yaml`:

```yaml
service_name: "my-service"
environment: "production"

global_thresholds:
  critical: 700
  high: 500
  medium: 300
  low: 100

errors:
  ERR-001:
    code: "ERR-001"
    description: "Database connection failed"
    base_severity: 10
    base_occurrence: 3
    base_detect: 2
    severity_rules:
      - condition: "production"
        multiplier: 1.5
      - condition: "business_hours"
        multiplier: 1.3
    occurrence_patterns:
      - type: "rate"
        threshold: 1.0
        score: 10
    detection_config:
      monitoring_enabled: true
      alerting_enabled: true
      auto_detect: true
      mttd_target: 10s
```

### 2. Initialize SOD Framework

```go
package main

import (
    "context"
    "log"
    
    "github.com/your-org/shared/sod"
)

func main() {
    // Create factory
    factory := sod.NewFactory("my-service", "production")
    
    // Create calculator with auto-config discovery
    calculator, err := factory.CreateCalculatorWithAutoConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create helper for easier usage
    helper := sod.NewHelper(calculator, "my-service", "production")
    
    // Use in your application
    score, err := helper.CalculateForError(context.Background(), "ERR-001")
    if err != nil {
        log.Printf("Failed to calculate SOD: %v", err)
        return
    }
    
    log.Printf("SOD Score: %s", sod.FormatScoreDetails(score))
    
    // Check if should alert
    config, _ := calculator.GetErrorConfig("ERR-001")
    if sod.ShouldPage(score, config.Thresholds) {
        // Page on-call engineer
        log.Printf("CRITICAL: Paging on-call")
    }
}
```

### 3. Integrate with Error Handling

```go
// In your error handler
func handleError(ctx context.Context, err error, errorCode string) {
    // Calculate SOD score
    score, sodErr := sodHelper.CalculateForError(ctx, errorCode)
    if sodErr != nil {
        logger.Error("Failed to calculate SOD", zap.Error(sodErr))
        return
    }
    
    // Log with SOD context
    logger.WithError(errorCode, sod.GetSeverityLevel(score)).
        Error("Error occurred",
            zap.Int("sod_score", score.Total),
            zap.Int("severity", score.Severity),
            zap.Int("occurrence", score.Occurrence),
            zap.Int("detect", score.Detect),
            zap.Error(err))
    
    // Alert based on SOD score
    if score.Total >= 700 {
        alertingService.PageOnCall(errorCode, score)
    } else if score.Total >= 500 {
        alertingService.SendAlert(errorCode, score)
    }
}
```

## Configuration Reference

### Severity Rules

Adjust severity based on runtime conditions:

```yaml
severity_rules:
  - condition: "production"      # If in production
    multiplier: 1.5              # Increase severity by 50%
  
  - condition: "business_hours"  # If during business hours
    override: 10                 # Set severity to 10 (max)
  
  - condition: "high_load"       # If system load > 75%
    multiplier: 1.4
```

**Available Conditions:**
- `production`, `staging`, `dev` - Environment
- `business_hours`, `night_hours` - Time of day
- `high_load`, `medium_load` - System load
- `error_storm` - High error rate (>10/sec)

### Occurrence Patterns

Define how occurrence score changes based on error patterns:

```yaml
occurrence_patterns:
  - type: "rate"           # Error rate pattern
    threshold: 10.0        # 10 errors/second
    score: 9               # Set occurrence to 9
  
  - type: "burst"          # Burst pattern
    threshold: 5.0         # 5 errors/second
    score: 8               # Set occurrence to 8
```

### Detection Configuration

Configure error detection and monitoring:

```yaml
detection_config:
  monitoring_enabled: true    # Has monitoring
  alerting_enabled: true      # Has alerts configured
  auto_detect: true           # Automatically detected
  mttd_target: 30s            # Target: detect within 30s
  log_level: "ERROR"          # Log level for this error
```

## Dynamic SOD Calculation

SOD scores are adjusted based on runtime context:

### Runtime Factors

```go
// Update runtime factors periodically
calculator.UpdateRuntimeFactors(sod.RuntimeFactors{
    CurrentLoad:     0.85,     // 85% system load
    ErrorRate:       5.2,      // 5.2 errors/second
    ActiveUsers:     1500,
    SystemDegraded:  false,
    MaintenanceMode: false,
})
```

### Error Context

Provide rich context for accurate scoring:

```go
score, err := calculator.CalculateScore(ctx, "ERR-001", sod.ErrorContext{
    Timestamp:       time.Now(),
    ServiceName:     "my-service",
    Environment:     "production",
    UserID:          "user-123",
    DeviceID:        "device-456",
    SystemLoad:      0.75,
    RecentErrorRate: 3.5,
    IsBusinessHours: true,
})
```

## Prometheus Metrics

The framework exports the following metrics:

```promql
# Current SOD score
sod_score_total{service="my-service", error_code="ERR-001", severity_level="critical"}

# SOD score distribution
sod_score_distribution_bucket{service="my-service", error_code="ERR-001"}

# Component scores
sod_severity_score{service="my-service", error_code="ERR-001"}
sod_occurrence_score{service="my-service", error_code="ERR-001"}
sod_detect_score{service="my-service", error_code="ERR-001"}

# Error tracking
sod_error_occurrences_total{service="my-service", error_code="ERR-001", severity="critical"}

# MTTD/MTTR
sod_mean_time_to_detect_seconds{service="my-service", error_code="ERR-001"}
sod_mean_time_to_resolve_seconds{service="my-service", error_code="ERR-001"}
```

## Best Practices

### 1. Start with Static Scores
Begin with static base scores, then add dynamic rules as you learn error patterns.

### 2. Review SOD Scores Regularly
Use metrics to validate that SOD scores match real-world impact.

### 3. Use Consistent Error Codes
Follow a naming convention: `SERVICE-NNN` (e.g., `AUTH-001`, `DB-002`)

### 4. Set Realistic Thresholds
Tune thresholds based on your team's capacity and SLOs.

### 5. Monitor MTTD/MTTR
Track detection and resolution times to improve `D` scores.

## Example: Integration with Existing Error Registry

```go
// Enhance existing error registry with SOD
type ServiceError struct {
    Code     string
    Severity errors.Severity
    Message  string
    
    // Add SOD score
    SODScore sod.Score
}

func wrapErrorWithSOD(ctx context.Context, err error, code string) *ServiceError {
    // Get existing error from registry
    serviceErr := ErrorRegistry.WrapError(err, code)
    
    // Calculate SOD score
    sodScore, _ := sodHelper.CalculateForError(ctx, code)
    
    serviceErr.SODScore = sodScore
    
    return serviceErr
}
```

## Testing

```go
// Use NopMetrics for testing
loader := sod.NewFileConfigLoader("testdata/sod_test.yaml")
metrics := sod.NewNopMetrics()
calculator, _ := sod.NewCalculator(loader, metrics)

score, err := calculator.CalculateScore(ctx, "TEST-001", errorContext)
assert.NoError(t, err)
assert.Equal(t, 600, score.Total)
```

## Advanced Usage

### Hot Reload Configuration

```go
// Watch for config changes
ctx := context.Background()
loader.Watch(ctx, func(newConfig *sod.Config) {
    log.Printf("SOD config reloaded: %d errors configured", len(newConfig.Errors))
})
```

### Custom Metrics Collector

```go
type CustomMetrics struct {
    // Your custom metrics implementation
}

func (m *CustomMetrics) RecordSODScore(errorCode string, score sod.Score) {
    // Send to your custom monitoring system
}

// Use custom collector
calculator, _ := sod.NewCalculator(loader, &CustomMetrics{})
```

## Contributing

When adding new features:
1. Maintain interface-based design
2. Add tests for new components
3. Update configuration schema
4. Document in README

## License

MIT License - Internal use at your-org IoT Platform
