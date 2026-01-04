# Enterprise Coding Standards

> **Unified patterns for Go, Python, and .NET using Core Packages**

This document defines the coding standards for enterprise microservices using the Core package library. All three language implementations (Go, Python, .NET) share identical patterns to ensure consistency across polyglot services.

---

## Table of Contents

1. [Core Package Overview](#core-package-overview)
2. [Dependency Injection & Testability](#dependency-injection--testability)
3. [Logging Standards](#logging-standards)
4. [Error Handling](#error-handling)
5. [Metrics & Observability](#metrics--observability)
6. [SLI/SLO Tracking](#slislo-tracking)
7. [SOD Scoring](#sod-scoring)
8. [Reliability Patterns](#reliability-patterns)
9. [Configuration Management](#configuration-management)
10. [Infrastructure Health](#infrastructure-health)
11. [Code Structure](#code-structure)
12. [Testing Requirements](#testing-requirements)

---

## Core Package Overview

The Core package provides **consistent patterns** across all three languages:

| Package | Go | Python | .NET | Purpose |
|---------|-----|--------|------|---------|
| **Logger** | `core/logger` | `core.logger` | `Core.Logger` | Structured JSON logging with correlation IDs |
| **Errors** | `core/errors` | `core.errors` | `Core.Errors` | ServiceError with codes, severity, context |
| **Metrics** | `core/metrics` | `core.metrics` | `Core.Metrics` | Four Golden Signals (Prometheus) |
| **Config** | `core/config` | `core.config` | `Core.Config` | Validated configuration with timeouts |
| **SLI** | `core/sli` | `core.sli` | `Core.Sli` | SLI/SLO tracking with error budgets |
| **SOD** | `core/sod` | `core.sod` | `Core.Sod` | Severity × Occurrence × Detectability scoring |
| **Reliability** | `core/reliability` | `core.reliability` | `Core.Reliability` | Circuit breaker, retry, rate limiting |
| **Infrastructure** | `core/infrastructure` | `core.infrastructure` | `Core.Infrastructure` | Redis, health checks, liveness/readiness |

---

## Dependency Injection & Testability

### Principles

1. **Constructor injection** - All dependencies passed via constructor
2. **Interface-based** - Depend on abstractions, not implementations
3. **No hidden dependencies** - No service locators or static singletons
4. **Mockable boundaries** - External systems behind interfaces

### Go DI Pattern

```go
package service

import (
    "context"
    "core/logger"
    "core/errors"
    "core/metrics"
)

// Define interfaces for dependencies
type DeviceRepository interface {
    Get(ctx context.Context, id string) (*Device, error)
    Save(ctx context.Context, device *Device) error
}

type EventPublisher interface {
    Publish(topic string, event interface{}) error
}

// Service with injected dependencies
type DeviceService struct {
    repo      DeviceRepository
    publisher EventPublisher
    logger    *logger.Logger
    metrics   *metrics.ServiceMetrics
}

// Constructor receives all dependencies
func NewDeviceService(
    repo DeviceRepository,
    publisher EventPublisher,
    log *logger.Logger,
    metrics *metrics.ServiceMetrics,
) *DeviceService {
    return &DeviceService{
        repo:      repo,
        publisher: publisher,
        logger:    log,
        metrics:   metrics,
    }
}
```

### Python DI Pattern

```python
from core.logger import Logger, ContextLogger
from core.errors import ServiceError, ErrorRegistry
from core.metrics import ServiceMetrics
from abc import ABC, abstractmethod

# Define protocol/interface
class DeviceRepository(ABC):
    @abstractmethod
    def get(self, device_id: str) -> Device:
        pass
    
    @abstractmethod
    def save(self, device: Device) -> None:
        pass

class EventPublisher(ABC):
    @abstractmethod
    def publish(self, topic: str, event: dict) -> None:
        pass

# Service with injected dependencies
class DeviceService:
    def __init__(
        self,
        repository: DeviceRepository,
        publisher: EventPublisher,
        logger: Logger,
        metrics: ServiceMetrics,
    ):
        self._repository = repository
        self._publisher = publisher
        self._logger = logger
        self._metrics = metrics
```

### .NET DI Pattern

```csharp
using Core.Logger;
using Core.Errors;
using Core.Metrics;

// Define interfaces
public interface IDeviceRepository
{
    Task<Device> GetAsync(string id, CancellationToken ct = default);
    Task SaveAsync(Device device, CancellationToken ct = default);
}

public interface IEventPublisher
{
    Task PublishAsync(string topic, object @event);
}

// Service with injected dependencies
public class DeviceService
{
    private readonly IDeviceRepository _repository;
    private readonly IEventPublisher _publisher;
    private readonly ServiceLogger _logger;
    private readonly ServiceMetrics _metrics;

    public DeviceService(
        IDeviceRepository repository,
        IEventPublisher publisher,
        ServiceLogger logger,
        ServiceMetrics metrics)
    {
        _repository = repository;
        _publisher = publisher;
        _logger = logger;
        _metrics = metrics;
    }
}

// DI registration in Program.cs
builder.Services.AddSingleton(ServiceLogger.NewProduction("device-service", "1.0.0"));
builder.Services.AddSingleton<ServiceMetrics>();
builder.Services.AddScoped<IDeviceRepository, RedisDeviceRepository>();
builder.Services.AddScoped<IEventPublisher, KafkaEventPublisher>();
builder.Services.AddScoped<DeviceService>();
```

---

## Logging Standards

### Standard Log Format

All services output **JSON structured logs** with consistent fields:

```json
{
    "timestamp": "2024-01-15T10:30:45.123Z",
    "level": "INFO",
    "service": "device-service",
    "version": "1.2.3",
    "environment": "production",
    "correlation_id": "req-550e8400-e29b-41d4-a716-446655440000",
    "component": "DeviceRepository",
    "operation": "Save",
    "message": "Device saved successfully",
    "device_id": "dev-123",
    "error_code": "OK",
    "severity": "LOW",
    "duration_ms": 45
}
```

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | ISO8601 | UTC timestamp |
| `level` | string | DEBUG, INFO, WARN, ERROR, FATAL |
| `service` | string | Service name |
| `correlation_id` | UUID | Request tracing ID |
| `component` | string | Class/module name |
| `operation` | string | Method/function name |
| `message` | string | Human-readable message |
| `error_code` | string | Error code (if applicable) |
| `severity` | string | LOW, MEDIUM, HIGH, CRITICAL |

### Go Logging with Core Package

```go
import "core/logger"

func main() {
    // Create logger
    log, err := logger.New(logger.Config{
        ServiceName: "device-service",
        Environment: "production",
        Version:     "1.2.3",
        LogLevel:    "info",
    })
    if err != nil {
        panic(err)
    }
    defer log.Sync()
    
    // Basic logging
    log.Info("Service started", zap.String("port", "8080"))
    
    // Context-aware logging with correlation ID
    ctx := context.WithValue(context.Background(), 
        logger.CorrelationIDKey, "req-550e8400")
    
    ctxLog := log.WithContext(ctx, "DeviceService")
    ctxLog.Info("Processing request",
        zap.String("operation", "UpdateDevice"),
        zap.String("device_id", "dev-123"),
    )
    
    // Error logging
    ctxLog.Error("Failed to save device",
        zap.String("operation", "UpdateDevice"),
        zap.String("device_id", "dev-123"),
        zap.String("error_code", "DEV-002"),
        zap.String("severity", "HIGH"),
        zap.Error(err),
    )
}
```

### Python Logging with Core Package

```python
from core.logger import Logger

# Create logger
logger = Logger(
    service_name="device-service",
    version="1.2.3",
    environment="production",
    log_level="INFO",
)

# Basic logging
logger.info("Service started", port=8080)

# Context-aware logging
ctx_logger = logger.with_correlation("req-550e8400", component="DeviceService")
ctx_logger.info(
    "Processing request",
    operation="update_device",
    device_id="dev-123",
)

# Error logging
ctx_logger.error(
    "Failed to save device",
    operation="update_device",
    device_id="dev-123",
    error_code="DEV-002",
    severity="HIGH",
    error=str(ex),
    exc_info=True,
)
```

### .NET Logging with Core Package

```csharp
using Core.Logger;

// Create logger
var logger = ServiceLogger.NewProduction("device-service", "1.2.3");

// Basic logging
logger.Info("Service started", new { port = 8080 });

// Context-aware logging
using var ctx = logger.WithCorrelation("req-550e8400", "DeviceService");
ctx.Info("Processing request", new
{
    operation = "UpdateDevice",
    device_id = "dev-123"
});

// Error logging
ctx.Error("Failed to save device", ex, new
{
    operation = "UpdateDevice",
    device_id = "dev-123",
    error_code = "DEV-002",
    severity = "HIGH"
});

// Using scope for correlation
using (CorrelationScope.Push("req-550e8400"))
{
    // All logs within this scope include correlation ID
    logger.Info("Inside correlation scope");
}
```

---

## Error Handling

### Principles

1. **Never swallow exceptions** - Always log before re-throwing
2. **Use error codes** - Every error path has a unique code
3. **Include severity** - Tag errors for SRE escalation
4. **Add context** - Include relevant identifiers
5. **Use ServiceError** - Consistent error structure across languages

### Severity Levels

| Severity | Response Time | Auto-Page | Description |
|----------|---------------|-----------|-------------|
| `CRITICAL` | 15 min | Yes (24/7) | Service down, data loss, security breach |
| `HIGH` | 1 hour | Yes (business) | Feature unavailable, blocking users |
| `MEDIUM` | 4 hours | No | Degraded experience, workaround available |
| `LOW` | 24 hours | No | Minor issue, minimal impact |
| `INFO` | N/A | No | Informational only |

### Go Error Handling

```go
import "core/errors"

// Create a ServiceError
func (s *DeviceService) GetDevice(ctx context.Context, id string) (*Device, error) {
    device, err := s.repo.Get(ctx, id)
    if err != nil {
        // Wrap with ServiceError
        return nil, errors.Wrap(err, "DEV-001", errors.SeverityHigh,
            "Failed to retrieve device from database").
            WithContext("device_id", id).
            WithContext("correlation_id", ctx.Value("correlation_id"))
    }
    return device, nil
}

// Using ErrorRegistry for pre-defined errors
var errorRegistry = errors.NewErrorRegistry()

func init() {
    errorRegistry.Register(&errors.ErrorDefinition{
        Code:        "DEV-001",
        Severity:    errors.SeverityHigh,
        Description: "Failed to retrieve device: %s",
        SODScore:    150,
        Mitigation:  "Check database connectivity",
    })
}

func (s *DeviceService) SaveDevice(ctx context.Context, device *Device) error {
    if err := s.repo.Save(ctx, device); err != nil {
        return errorRegistry.WrapError(err, "DEV-002", device.ID)
    }
    return nil
}
```

### Python Error Handling

```python
from core.errors import ServiceError, ErrorRegistry, Severity

# Create a ServiceError
class DeviceService:
    def get_device(self, device_id: str, correlation_id: str) -> Device:
        try:
            return self._repository.get(device_id)
        except Exception as ex:
            raise ServiceError(
                code="DEV-001",
                severity=Severity.HIGH,
                message="Failed to retrieve device from database",
                original_error=ex,
            ).with_context("device_id", device_id)\
             .with_context("correlation_id", correlation_id)

# Using ErrorRegistry
error_registry = ErrorRegistry()
error_registry.register(ErrorDefinition(
    code="DEV-001",
    severity=Severity.HIGH,
    description="Failed to retrieve device: %s",
    sod_score=150,
    mitigation="Check database connectivity",
))

def save_device(self, device: Device) -> None:
    try:
        self._repository.save(device)
    except Exception as ex:
        raise error_registry.create_error("DEV-002", device.id) from ex
```

### .NET Error Handling

```csharp
using Core.Errors;

// Create a ServiceError
public class DeviceService
{
    public async Task<Device> GetDeviceAsync(string id, CancellationToken ct)
    {
        try
        {
            return await _repository.GetAsync(id, ct);
        }
        catch (Exception ex)
        {
            throw new ServiceError("DEV-001", Severity.High, 
                "Failed to retrieve device from database", ex)
                .WithContext("device_id", id)
                .WithContext("correlation_id", CorrelationScope.Current);
        }
    }
}

// Using ErrorRegistry
public static class DeviceErrors
{
    private static readonly ErrorRegistry Registry = new();

    static DeviceErrors()
    {
        Registry.Register(new ErrorDefinition
        {
            Code = "DEV-001",
            Severity = Severity.High,
            Description = "Failed to retrieve device: {0}",
            SodScore = 150,
            Mitigation = "Check database connectivity"
        });
    }

    public static ServiceError NotFound(string deviceId) =>
        Registry.CreateError("DEV-001", deviceId);
}

// Usage
throw DeviceErrors.NotFound(deviceId);
```

### Error Code Format

```
[SERVICE]-[NUMBER]

Examples:
- DEV-001: Device service error #1
- AUTH-002: Auth service error #2  
- INGEST-003: Ingest service error #3
```

### Service Prefixes

| Service | Prefix | Range |
|---------|--------|-------|
| api-gateway | `GW` | GW-001 to GW-099 |
| device-service | `DEV` | DEV-001 to DEV-099 |
| device-ingest | `INGEST` | INGEST-001 to INGEST-099 |
| auth-service | `AUTH` | AUTH-001 to AUTH-099 |
| notification-service | `NOTIF` | NOTIF-001 to NOTIF-099 |
| analytics-* | `ANALYTICS` | ANALYTICS-001 to ANALYTICS-099 |
| model-* | `ML` | ML-001 to ML-099 |
| frontend | `UI` | UI-001 to UI-099 |

---

## Metrics & Observability

### Four Golden Signals

All services expose Prometheus metrics for the **Four Golden Signals**:

1. **Latency** - Request duration histograms
2. **Traffic** - Request count by endpoint/status
3. **Errors** - Error count by code/severity
4. **Saturation** - Resource utilization (CPU, memory, connections)

### Go Metrics

```go
import "core/metrics"

// Create metrics
m := metrics.New(metrics.Config{
    ServiceName: "device-service",
    Namespace:   "iot",
})

// Record request duration
start := time.Now()
// ... process request ...
m.RecordLatency("UpdateDevice", time.Since(start))

// Increment request count
m.IncrementRequests("UpdateDevice", "200")

// Track active connections
m.SetGauge("active_connections", 42.0, "redis")

// Record error
m.IncrementErrors("DEV-001", errors.SeverityHigh)

// Expose /metrics endpoint
http.Handle("/metrics", m.Handler())
```

### Python Metrics

```python
from core.metrics import ServiceMetrics

# Create metrics
metrics = ServiceMetrics(
    service_name="device-service",
    namespace="iot",
)

# Record request duration
with metrics.track_latency("update_device"):
    # ... process request ...
    pass

# Or manual tracking
start = time.time()
# ... process request ...
metrics.record_latency("update_device", time.time() - start)

# Increment request count
metrics.increment_requests("update_device", status_code="200")

# Track active connections
metrics.set_gauge("active_connections", 42, labels={"type": "redis"})

# Record error
metrics.increment_errors("DEV-001", severity="HIGH")
```

### .NET Metrics

```csharp
using Core.Metrics;

// Create metrics
var metrics = new ServiceMetrics(new MetricsConfig
{
    ServiceName = "device-service",
    Namespace = "iot"
});

// Record request duration
var stopwatch = Stopwatch.StartNew();
// ... process request ...
stopwatch.Stop();
metrics.RecordLatency("UpdateDevice", stopwatch.Elapsed);

// Or using the tracking method
using (metrics.TrackLatency("UpdateDevice"))
{
    // ... process request ...
}

// Increment request count
metrics.IncrementRequests("UpdateDevice", "200");

// Track active connections
metrics.SetGauge("active_connections", 42, new[] { "redis" });

// Record error
metrics.IncrementErrors("DEV-001", Severity.High);

// Expose /metrics endpoint (ASP.NET Core)
app.UseMetricServer();
```

---

## SLI/SLO Tracking

### Service Level Indicators

Track SLIs using the Core SLI package:

```go
// Go
import "core/sli"

tracker := sli.NewPrometheusSliTracker("device-service")

// Record request outcome
tracker.RecordRequest("UpdateDevice", sli.RequestOutcome{
    Success:     true,
    LatencyMs:   45,
    StatusCode:  200,
    ErrorCode:   "",
})

// Check error budget
budget := tracker.GetErrorBudget("UpdateDevice", 0.999) // 99.9% target
if budget.Remaining < 0.1 {
    // Less than 10% budget remaining - slow down deployments
}
```

```python
# Python
from core.sli import SliTracker, RequestOutcome

tracker = SliTracker("device-service")

# Record request outcome
tracker.record_request("update_device", RequestOutcome(
    success=True,
    latency_ms=45,
    status_code=200,
))

# Check error budget
budget = tracker.get_error_budget("update_device", target=0.999)
if budget.remaining < 0.1:
    # Less than 10% budget remaining
    pass
```

```csharp
// .NET
using Core.Sli;

var tracker = new PrometheusSliTracker("device-service");

// Record request outcome
tracker.RecordRequest("UpdateDevice", new RequestOutcome
{
    Success = true,
    LatencyMs = 45,
    StatusCode = 200
});

// Check error budget
var budget = tracker.GetErrorBudget("UpdateDevice", 0.999);
if (budget.Remaining < 0.1)
{
    // Less than 10% budget remaining
}
```

---

## SOD Scoring

### Severity × Occurrence × Detectability

Use SOD scoring for error prioritization:

| Factor | Score | Description |
|--------|-------|-------------|
| **Severity (S)** | 1-10 | Impact on users/system |
| **Occurrence (O)** | 1-10 | How often it happens |
| **Detectability (D)** | 1-10 | How hard to detect (10 = hardest) |

**SOD Score = S × O × D** (Range: 1-1000)

### Go SOD Calculator

```go
import "core/sod"

calculator := sod.NewSodCalculator()

// Calculate SOD score
score := calculator.Calculate(sod.ErrorContext{
    ErrorCode:   "DEV-001",
    Severity:    8,  // High impact
    Occurrence:  3,  // Rare
    Detectability: 2,  // Easy to detect
})
// SOD Score: 48 (8 × 3 × 2)

// Classify priority
priority := calculator.ClassifyPriority(score)
// Returns: "MEDIUM" (score < 100)
```

### Python SOD Calculator

```python
from core.sod import SodCalculator, ErrorContext

calculator = SodCalculator()

# Calculate SOD score
score = calculator.calculate(ErrorContext(
    error_code="DEV-001",
    severity=8,      # High impact
    occurrence=3,    # Rare
    detectability=2, # Easy to detect
))
# SOD Score: 48

# Classify priority
priority = calculator.classify_priority(score)
# Returns: "MEDIUM"
```

### .NET SOD Calculator

```csharp
using Core.Sod;

var calculator = new SodCalculator();

// Calculate SOD score
var score = calculator.Calculate(new ErrorContext
{
    ErrorCode = "DEV-001",
    Severity = 8,       // High impact
    Occurrence = 3,     // Rare
    Detectability = 2   // Easy to detect
});
// SOD Score: 48

// Classify priority
var priority = calculator.ClassifyPriority(score);
// Returns: SodPriority.Medium
```

### Priority Classification

| SOD Score | Priority | Action |
|-----------|----------|--------|
| 1-50 | LOW | Monitor, fix when convenient |
| 51-100 | MEDIUM | Schedule for next sprint |
| 101-200 | HIGH | Fix within 48 hours |
| 201-500 | CRITICAL | Fix immediately |
| 501-1000 | EMERGENCY | All hands on deck |

---

## Reliability Patterns

### Circuit Breaker

Protect against cascading failures:

```go
// Go
import "core/reliability"

cb := reliability.NewCircuitBreaker(reliability.CircuitBreakerConfig{
    Name:                "database",
    MaxRequests:         5,
    Interval:            10 * time.Second,
    Timeout:             30 * time.Second,
    FailureThreshold:    5,
    SuccessThreshold:    2,
})

result, err := cb.Execute(func() (interface{}, error) {
    return db.Query("SELECT * FROM devices")
})
```

```python
# Python
from core.reliability import CircuitBreaker, CircuitBreakerConfig

cb = CircuitBreaker(CircuitBreakerConfig(
    name="database",
    failure_threshold=5,
    success_threshold=2,
    timeout_seconds=30,
))

@cb.protect
def query_database():
    return db.query("SELECT * FROM devices")
```

```csharp
// .NET (using Polly)
using Core.Reliability;

var cb = new CircuitBreaker(new CircuitBreakerConfig
{
    Name = "database",
    FailureThreshold = 5,
    BreakDuration = TimeSpan.FromSeconds(30)
});

var result = await cb.ExecuteAsync(async () =>
{
    return await db.QueryAsync("SELECT * FROM devices");
});
```

### Retry Policy

Handle transient failures:

```go
// Go
retry := reliability.NewRetryPolicy(reliability.RetryConfig{
    MaxRetries:      3,
    InitialInterval: 100 * time.Millisecond,
    MaxInterval:     2 * time.Second,
    Multiplier:      2.0,
})

result, err := retry.Execute(func() (interface{}, error) {
    return httpClient.Post(url, body)
})
```

```python
# Python
from core.reliability import RetryPolicy, RetryConfig

retry = RetryPolicy(RetryConfig(
    max_retries=3,
    initial_interval_ms=100,
    max_interval_ms=2000,
    multiplier=2.0,
))

@retry.with_retry
def call_external_api():
    return requests.post(url, json=body)
```

```csharp
// .NET
using Core.Reliability;

var retry = new RetryPolicy(new RetryConfig
{
    MaxRetries = 3,
    InitialDelay = TimeSpan.FromMilliseconds(100),
    MaxDelay = TimeSpan.FromSeconds(2),
    BackoffMultiplier = 2.0
});

var result = await retry.ExecuteAsync(async () =>
{
    return await httpClient.PostAsync(url, content);
});
```

### Rate Limiter

Prevent overload:

```go
// Go
limiter := reliability.NewRateLimiter(reliability.RateLimiterConfig{
    RequestsPerSecond: 100,
    BurstSize:         10,
})

if limiter.Allow() {
    // Process request
} else {
    // Return 429 Too Many Requests
}
```

```csharp
// .NET
var limiter = new RateLimiter(new RateLimiterConfig
{
    PermitsPerSecond = 100,
    BurstSize = 10
});

if (await limiter.TryAcquireAsync())
{
    // Process request
}
else
{
    return StatusCode(429);
}
```

### Bulkhead

Isolate failures:

```csharp
// .NET
var bulkhead = new Bulkhead(new BulkheadConfig
{
    MaxConcurrency = 10,
    MaxQueueSize = 100
});

var result = await bulkhead.ExecuteAsync(async () =>
{
    return await ProcessRequest();
});
```

---

## Configuration Management

### Timeout Standards

All external calls must have timeouts. **Minimum timeout is 60 seconds**:

```go
// Go
import "core/config"

cfg := config.NewTimeoutConfig(config.TimeoutDefaults{
    Database:     90 * time.Second,
    HTTP:         60 * time.Second,
    Redis:        60 * time.Second,
    MessageQueue: 120 * time.Second,
})

// Validation ensures minimums
if err := cfg.Validate(); err != nil {
    // Error if any timeout < 60s
}
```

```csharp
// .NET
using Core.Config;

var timeouts = new TimeoutConfig
{
    Database = TimeSpan.FromSeconds(90),
    Http = TimeSpan.FromSeconds(60),
    Redis = TimeSpan.FromSeconds(60),
    MessageQueue = TimeSpan.FromSeconds(120)
};

timeouts.Validate(); // Throws if any < 60s
```

### Service Configuration

```csharp
// .NET
var config = new ServiceConfig
{
    ServiceName = "device-service",
    Environment = "production",
    Version = "1.2.3",
    LogLevel = "Information",
    MetricsPort = 9090,
    HealthPort = 8081,
    Timeouts = new TimeoutConfig
    {
        Database = TimeSpan.FromSeconds(90),
        Http = TimeSpan.FromSeconds(60)
    }
};

config.Validate(); // Validates all settings
```

---

## Infrastructure Health

### Health Checks

Implement liveness and readiness probes:

```go
// Go
import "core/infrastructure"

checker := infrastructure.NewHealthChecker()
checker.AddCheck("redis", redisHealthCheck)
checker.AddCheck("database", dbHealthCheck)

// Liveness - is the service alive?
http.HandleFunc("/health/live", checker.LivenessHandler)

// Readiness - can it serve traffic?
http.HandleFunc("/health/ready", checker.ReadinessHandler)
```

```csharp
// .NET
using Core.Infrastructure;

var checker = new HealthChecker();
checker.AddCheck("redis", new RedisHealthCheck(redisClient));
checker.AddCheck("database", new DatabaseHealthCheck(dbConnection));

// ASP.NET Core
app.MapGet("/health/live", checker.CheckLiveness);
app.MapGet("/health/ready", checker.CheckReadiness);
```

### Redis Client

```csharp
// .NET
using Core.Infrastructure;

var redis = new RedisClient(new RedisConfig
{
    ConnectionString = "localhost:6379",
    Database = 0,
    ConnectTimeout = TimeSpan.FromSeconds(60),
    DefaultExpiry = TimeSpan.FromMinutes(30)
});

// Basic operations
await redis.SetAsync("key", "value");
var value = await redis.GetAsync("key");

// Health check
var isHealthy = await redis.IsHealthyAsync();
```

---

## Code Structure

### Go Service Structure

```
service-name/
├── main.go                 # Entry point
├── Dockerfile
├── go.mod
├── internal/
│   ├── domain/            # Business entities
│   ├── repository/        # Data access (interfaces + implementations)
│   ├── service/           # Business logic
│   ├── handler/           # HTTP/gRPC handlers
│   └── config/            # Configuration
├── pkg/                   # Reusable packages
└── tests/
    ├── unit/
    └── integration/
```

### Python Service Structure

```
service-name/
├── main.py                # Entry point
├── Dockerfile
├── requirements.txt
├── pyproject.toml
├── app/
│   ├── __init__.py
│   ├── domain/            # Business entities
│   ├── repository/        # Data access
│   ├── service/           # Business logic
│   ├── api/               # HTTP routes
│   └── config/            # Configuration
└── tests/
    ├── unit/
    └── integration/
```

### .NET Service Structure

```
ServiceName/
├── Program.cs             # Entry point
├── Dockerfile
├── ServiceName.csproj
├── Domain/                # Business entities
├── Repository/            # Data access
├── Services/              # Business logic
├── Controllers/           # HTTP endpoints
├── Configuration/         # Settings classes
└── Tests/
    └── ServiceName.Tests/
```

---

## Testing Requirements

### Coverage Requirements

| Test Type | Minimum Coverage | Scope |
|-----------|-----------------|-------|
| Unit Tests | 80% | All business logic |
| Integration Tests | All endpoints | API surface |
| E2E Tests | Critical paths | User journeys |

### Unit Test Pattern

```go
// Go
func TestDeviceService_UpdateDevice(t *testing.T) {
    // Arrange
    mockRepo := &MockDeviceRepository{
        GetFunc: func(id string) (*Device, error) {
            return &Device{ID: id, Status: "offline"}, nil
        },
    }
    logger, _ := logger.NewDevelopment()
    service := NewDeviceService(mockRepo, nil, logger, nil)

    // Act
    err := service.UpdateDevice(ctx, "dev-123", map[string]interface{}{"status": "online"})

    // Assert
    assert.NoError(t, err)
}
```

```csharp
// .NET
[Fact]
public async Task UpdateDevice_WithValidData_Succeeds()
{
    // Arrange
    var mockRepo = new Mock<IDeviceRepository>();
    mockRepo.Setup(r => r.GetAsync("dev-123", It.IsAny<CancellationToken>()))
        .ReturnsAsync(new Device { Id = "dev-123", Status = "offline" });

    var logger = ServiceLogger.NewDevelopment("test", "1.0.0");
    var service = new DeviceService(mockRepo.Object, null, logger, null);

    // Act
    await service.UpdateDeviceAsync("dev-123", new { status = "online" });

    // Assert
    mockRepo.Verify(r => r.SaveAsync(It.IsAny<Device>(), It.IsAny<CancellationToken>()));
}
```

---

## Summary

By using the **Core packages consistently** across Go, Python, and .NET:

✅ **Unified Patterns** - Same interfaces across all languages  
✅ **Testability** - DI enables isolated unit tests  
✅ **Observability** - Structured logging + Prometheus metrics  
✅ **Debuggability** - Error codes with SOD scoring  
✅ **Reliability** - Circuit breakers, retries, rate limiting  
✅ **Operational Excellence** - SLI tracking + health checks  

---

## Related Documentation

- [Core-Package.md](./Core-Package.md) - Core package design
- [SLI-Framework.md](./SLI-Framework.md) - SLI/SLO details
- [Error-Repository.md](./Error-Repository.md) - Error code registry
- [Reliability-Engineering.md](./Reliability-Engineering.md) - Resilience patterns
