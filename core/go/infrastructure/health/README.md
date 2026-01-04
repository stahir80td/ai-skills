# Health Infrastructure Package

## Overview

The `health` package provides a centralized health check framework for all infrastructure services in the core module. It enables standardized health checking across Kafka, MongoDB, Redis, SQL Server, and ScyllaDB.

## Concepts

### Health Checker

The central component that:
- Registers health check functions for various services
- Executes checks concurrently with timeout protection
- Aggregates results with component-level status
- Provides structured logging and error context

### Check Status

Each health check returns one of three states:

- `StatusHealthy` - Service operational and responsive
- `StatusUnhealthy` - Service unavailable or critical error
- `StatusDegraded` - Service available but performing slowly

### Check Result

Contains detailed information about each check:

```go
type CheckResult struct {
    Name     string         // Check name (e.g., "kafka-producer")
    Status   CheckStatus    // Current status
    Duration time.Duration  // Execution time
    Message  string         // Human-readable status message
    Error    string         // Error details if failed
    LastRun  time.Time      // Last execution timestamp
}
```

## Usage

### Creating a Health Checker

```go
logger, _ := zap.NewProduction()
defer logger.Sync()

checker := health.NewChecker(logger)

// Optional: Set global timeout for all checks
checker.SetTimeout(5 * time.Second)
```

### Registering Health Checks

```go
// Kafka producer check
checker.Register("kafka-producer", func(ctx context.Context) error {
    return kafkaProducer.Health(ctx)
})

// MongoDB client check
checker.Register("mongodb", func(ctx context.Context) error {
    return mongoClient.Health(ctx)
})

// Redis cache check
checker.Register("redis-cache", func(ctx context.Context) error {
    return redisClient.Health(ctx)
})

// Custom check with internal logic
checker.Register("custom-service", func(ctx context.Context) error {
    // Perform custom health validation
    err := customService.Ping(ctx)
    if err != nil {
        return fmt.Errorf("custom_service_unavailable: %w", err)
    }
    return nil
})
```

### Running Health Checks

```go
// Execute all registered checks concurrently
results := checker.Check(context.Background())

// Check overall health status
isHealthy := checker.IsHealthy(context.Background())

// Log results
for name, result := range results {
    if result.Status == health.StatusUnhealthy {
        logger.Error("health_check_failed",
            zap.String("check_name", name),
            zap.String("status", result.Status),
            zap.String("message", result.Message),
            zap.Duration("duration", result.Duration),
        )
    }
}
```

## Features

- **Concurrent Execution**: All checks run in parallel for faster results
- **Timeout Protection**: Individual and global timeout enforcement
- **Structured Logging**: Every check operation logged with context
- **Error Context**: Detailed error information for troubleshooting
- **Duration Tracking**: Performance metrics for each check
- **Status Aggregation**: Overall health status from component checks

## Integration Patterns

### HTTP Health Endpoint

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    results := checker.Check(r.Context())
    
    status := http.StatusOK
    if !checker.IsHealthy(r.Context()) {
        status = http.StatusServiceUnavailable
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "healthy": checker.IsHealthy(r.Context()),
        "checks":  results,
    })
}

router.HandleFunc("/health", healthHandler)
```

### Startup Validation

```go
func startup(ctx context.Context) error {
    // Initialize all services
    kafkaProducer, _ := kafka.NewProducer(kafkaCfg)
    mongoClient, _ := mongodb.NewClient(mongoCfg)
    
    // Register health checks
    checker.Register("kafka", func(ctx context.Context) error {
        return kafkaProducer.Health(ctx)
    })
    checker.Register("mongodb", func(ctx context.Context) error {
        return mongoClient.Health(ctx)
    })
    
    // Verify all services are healthy
    if !checker.IsHealthy(ctx) {
        results := checker.Check(ctx)
        for name, result := range results {
            if result.Status != health.StatusHealthy {
                return fmt.Errorf("startup_failed: %s unhealthy: %s", name, result.Message)
            }
        }
    }
    
    return nil
}
```

### Periodic Health Monitoring

```go
func startHealthMonitoring(checker *health.Checker, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for range ticker.C {
        results := checker.Check(context.Background())
        
        for name, result := range results {
            if result.Status == health.StatusUnhealthy {
                logger.Warn("service_unhealthy",
                    zap.String("service", name),
                    zap.String("error", result.Error),
                )
                // Trigger alerting, rebalancing, etc.
            }
        }
    }
}

// Start monitoring
go startHealthMonitoring(checker, 10*time.Second)
```

## Logging Output

Health checks produce structured logs for monitoring:

```json
{
  "level": "info",
  "ts": 1705180234.123,
  "msg": "health_checks_start",
  "check_count": 5,
  "component": "health"
}
```

```json
{
  "level": "debug",
  "ts": 1705180234.145,
  "msg": "health_check_passed",
  "name": "kafka-producer",
  "duration": 0.015234,
  "component": "health"
}
```

```json
{
  "level": "warn",
  "ts": 1705180234.234,
  "msg": "health_check_failed",
  "name": "redis-cache",
  "error": "connection refused",
  "duration": 0.012456,
  "component": "health"
}
```

## Configuration Recommendations

| Setting | Recommendation | Rationale |
|---------|---|-----------|
| Global Timeout | 5-30 seconds | Balance between comprehensive checks and fast responses |
| Per-Check Timeout | 2-10 seconds | Allow individual slow services without blocking others |
| HTTP Endpoint Timeout | 30 seconds | Account for multiple concurrent checks |
| Monitoring Interval | 10-60 seconds | Catch failures quickly without overwhelming logs |

## Testing

```bash
go test ./infrastructure/health -v
```

Test cases include:
- Register checks and verify execution
- Timeout handling for slow checks
- Overall health status aggregation
- Concurrent check execution

## Dependencies

- **core/logger** - Structured logging with zap
- **Standard library** - `context`, `sync`, `time`

## Best Practices

1. **Lightweight Checks**: Keep individual checks fast (< 1 second)
2. **Meaningful Names**: Use descriptive check names (e.g., "kafka-producer" not "check1")
3. **Specific Errors**: Return errors with context explaining what failed
4. **Timeout Tuning**: Set timeouts appropriate to each service's response time
5. **Graceful Degradation**: Continue checking even if one check fails
6. **Error Recovery**: Log health check failures for incident investigation

## Future Enhancements

- Weighted health scores based on service criticality
- Automatic remediation triggers (e.g., restart unhealthy service)
- Health trend analysis over time
- Custom metrics export for Prometheus
- Circuit breaker integration with health status
