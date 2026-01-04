# Core Analytics Package

Production-grade analytics and ML capabilities for IoT your-org platform.

## Overview

The Core Analytics Package extends the existing Core Package with data platform capabilities needed by analytics and ML services. All packages follow strict SRE standards with dependency injection, circuit breakers, comprehensive logging, and Prometheus metrics.

## Packages

### 📊 `analytics/aggregation`

Time-series aggregation utilities for computing hourly/daily rollups from raw events.

**Key Features:**
- Hourly and daily statistics computation
- Event bucketing and merging
- Feature extraction from metadata
- Full test coverage

**Usage Example:**
```go
import (
    "github.com/your-org/core/analytics/aggregation"
    "github.com/your-org/core/logger"
)

log, _ := logger.New(logger.Config{ServiceName: "my-service", Environment: "prod"})
agg := aggregation.NewAggregator(aggregation.Config{Logger: log})

// Compute hourly stats
features := agg.ExtractFeatures(event)
hourlyStats := agg.ComputeHourlyStats(ctx, event, features)

// Merge multiple hourly stats
merged := agg.MergeHourlyStats(ctx, hourlyStatsList)

// Compute daily stats from hourly
dailyStats := agg.ComputeDailyStats(ctx, hourlyStatsList)
```

### 🎯 `analytics/features`

Feature engineering utilities for ML model inputs.

**Key Features:**
- Rolling statistics (avg, sum, min, max, stddev)
- Percentile calculations (P50, P95, P99)
- Rate of change and delta computation
- Time-based features (hour_of_day, day_of_week, is_weekend)
- Exponential moving averages
- Outlier detection using IQR method
- Full test coverage

**Usage Example:**
```go
import "github.com/your-org/core/analytics/features"

calc := features.NewCalculator(features.Config{Logger: log})

// Rolling statistics
rollingAvg := calc.ComputeRollingAverage(ctx, dataPoints, 10)
stats := calc.ComputeRollingStats(ctx, dataPoints, 10)

// Percentiles
percentiles := calc.ComputePercentiles(ctx, dataPoints) // P50, P95, P99

// Time features for ML
timeFeatures := calc.ExtractTimeFeatures(ctx, timestamp)

// Outlier detection
outliers := calc.DetectOutliers(ctx, dataPoints, 1.5)
```

### 🤖 `analytics/mlflow`

MLflow client wrapper with circuit breaker and core logger integration.

**Key Features:**
- Model registry operations (get, register, create version)
- Run tracking and metrics logging
- Circuit breaker for fault tolerance
- Structured logging with correlation IDs
- Service error integration

**Usage Example:**
```go
import "github.com/your-org/core/analytics/mlflow"

client := mlflow.NewClient(mlflow.Config{
    BaseURL: "http://mlflow:5000",
    Timeout: 30 * time.Second,
    Logger:  log,
})

// Get model
model, err := client.GetModel(ctx, "anomaly-detection", "v1.2.0")

// Log metrics
err = client.LogMetric(ctx, runID, "accuracy", 0.95, timestamp)

// Health check
err = client.HealthCheck(ctx)
```

### ✅ `analytics/validation`

Data quality validation utilities.

**Key Features:**
- Schema validation with field types
- Range checks for numeric values
- Null value detection
- Outlier detection
- Timestamp validation
- Service error integration

**Usage Example:**
```go
import "github.com/your-org/core/analytics/validation"

validator := validation.NewValidator(validation.Config{Logger: log})

// Schema validation
schema := &validation.Schema{
    Fields: []validation.SchemaField{
        {Name: "device_id", Type: "string", Required: true},
        {Name: "temperature", Type: "float", Required: true, MinValue: ptrFloat64(-50), MaxValue: ptrFloat64(100)},
    },
}
result := validator.ValidateSchema(ctx, data, schema)

// Range validation
result := validator.ValidateRange(ctx, value, 0, 100, "temperature")

// Timestamp validation
result := validator.ValidateTimestamp(ctx, timestamp, 1*time.Hour)
```

### ⏱️ `analytics/timeseries`

Time-series processing utilities.

**Key Features:**
- Tumbling windows (non-overlapping)
- Sliding windows (overlapping)
- Session windows (gap-based)
- Time bucketing
- Missing value interpolation (linear, nearest, forward, backward)
- Resampling with aggregation
- Downsampling for visualization

**Usage Example:**
```go
import "github.com/your-org/core/analytics/timeseries"

processor := timeseries.NewProcessor(timeseries.Config{Logger: log})

// Create tumbling windows
windows := processor.CreateTumblingWindows(ctx, dataPoints, 1*time.Hour)

// Interpolate missing values
interpolated := processor.InterpolateMissing(ctx, dataPoints, 5*time.Minute, timeseries.InterpolationLinear)

// Resample to different frequency
resampled := processor.Resample(ctx, dataPoints, 15*time.Minute, "mean")

// Calculate window statistics
stats := processor.CalculateWindowStatistics(ctx, windows)
```

### 📈 `analytics/metrics`

Analytics-specific Prometheus metrics following core patterns.

**Key Metrics:**
- Model prediction latency and counts
- Model drift and accuracy scores
- Data quality metrics
- Feature store performance
- Aggregation performance
- MLflow API metrics
- Time-series processing metrics
- Anomaly detection metrics
- Query performance metrics

**Usage Example:**
```go
import "github.com/your-org/core/analytics/metrics"

// Record model prediction
metrics.RecordModelPrediction("anomaly-detection", "v1.2.0", "success", 0.125)

// Record data quality
metrics.RecordDataQuality("device_events", "completeness", 0.98)

// Record feature computation
metrics.RecordFeatureCompute("rolling_avg_temp", 0.005)

// Record anomaly detection
metrics.RecordAnomalyDetection("device-123", "temperature_spike", "high", 0.85, 0.050, "v1.0")
```

## Design Principles

### ✅ Dependency Injection

All packages use DI for easy testing and flexibility:

```go
type Config struct {
    Logger *logger.Logger
}

func NewComponent(cfg Config) *Component {
    return &Component{logger: cfg.Logger}
}
```

### ✅ Context Propagation

All methods accept `context.Context` for cancellation and tracing:

```go
func (c *Component) Process(ctx context.Context, data Data) error {
    // Implementation
}
```

### ✅ Core Logger Usage

**NEVER use fmt.Println or other loggers!** Always use the core logger:

```go
c.logger.Debug("Processing data", zap.String("id", id))
c.logger.Info("Completed", zap.Int("count", count))
c.logger.Error("Failed", zap.Error(err))
```

### ✅ Error Handling

Use core error types with codes and severity:

```go
return &errors.ServiceError{
    Code:       "VALIDATION-001",
    Message:    "Schema validation failed",
    Severity:   errors.SeverityMedium,
    Underlying: err,
}
```

### ✅ Circuit Breakers

External service calls use circuit breakers:

```go
cb := reliability.NewCircuitBreaker("external-api", 5, 60*time.Second)
err := cb.Execute(func() error {
    return externalCall()
})
```

### ✅ Prometheus Metrics

All operations emit metrics for observability:

```go
metrics.RecordOperation("operation_type", duration, success)
```

## Testing

All packages have comprehensive unit tests:

```bash
cd core/go
go test ./analytics/...
```

## Integration

Services using analytics packages:

1. **Analytics Ingestion** - aggregation, validation, timeseries
2. **Feature Store** - features, validation, timeseries
3. **Query Service** - timeseries, aggregation
4. **Operational Intelligence** - features, aggregation, validation
5. **Anomaly Detection** - features, mlflow, validation
6. **Predictive Analytics** - features, mlflow, timeseries
7. **Business Analytics** - aggregation, timeseries

## Dependencies

```go
module github.com/your-org/core

require (
    github.com/prometheus/client_golang v1.23.2
    go.uber.org/zap v1.27.1
    // ... other core dependencies
)
```

## Standards Compliance

- ✅ Uses core logger (no fmt.Println)
- ✅ Dependency injection for all components
- ✅ Context propagation throughout
- ✅ Circuit breakers for external calls
- ✅ Comprehensive error handling
- ✅ Prometheus metrics everywhere
- ✅ Full test coverage
- ✅ Structured logging with correlation IDs

## Next Steps

With the core analytics package complete, you can now:

1. Build the 4 Go analytics services
2. Build the 3 Python ML services
3. Deploy using existing Helm charts
4. Monitor via Grafana dashboards

---

**Created:** December 20, 2025  
**Status:** Production Ready ✅
