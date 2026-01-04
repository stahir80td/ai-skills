# Core .NET Package

Enterprise-grade .NET 8 core packages for building microservices with SRE best practices.

## Packages

| Package | Description |
|---------|-------------|
| **Core.Logger** | Structured logging with Serilog, correlation IDs, and component scoping |
| **Core.Errors** | Error repository pattern with error codes, severity, and SOD scoring |
| **Core.Metrics** | Prometheus metrics following Four Golden Signals |
| **Core.Config** | Configuration management with timeout validation |
| **Core.Sli** | Service Level Indicator tracking with Prometheus |
| **Core.Sod** | Symptom-Oriented Diagnosis framework |
| **Core.Reliability** | Circuit breaker, retry, rate limiting, bulkhead with Polly |
| **Core.Infrastructure** | Redis client, health checks |

## Quick Start

### Logger

```csharp
using Core.Logger;

// Production logger (JSON output)
using var logger = ServiceLogger.NewProduction("my-service", "1.0.0");

// Development logger (colorized console)
using var logger = ServiceLogger.NewDevelopment("my-service", "1.0.0");

// Log with correlation ID
var contextLogger = logger.WithContext(correlationId: "req-123", component: "OrderService");
contextLogger.Information("Processing order {OrderId}", orderId);
```

### Error Handling

```csharp
using Core.Errors;

// Create error registry
var registry = new ErrorRegistry();
registry.Register(new ErrorDefinition
{
    Code = "ORDER-001",
    Severity = Severity.High,
    Description = "Order {0} not found",
    SeverityScore = 7,
    OccurrenceScore = 3,
    DetectabilityScore = 2
});

// Create errors
var error = registry.CreateError("ORDER-001", orderId);
throw error.WithContext("user_id", userId);
```

### Metrics

```csharp
using Core.Metrics;

var metrics = new ServiceMetrics(new MetricsConfig
{
    ServiceName = "order-service",
    Namespace = "AI"
});

// Record request
using (metrics.TimeRequest("POST", "/api/orders", () => statusCode))
{
    // Process request
}

// Record error
metrics.RecordError("ORDER-001", Severity.High, "OrderProcessor");
```

### SLI Tracking

```csharp
using Core.Sli;

var tracker = new PrometheusSliTracker("order-service");

tracker.RecordRequest(new RequestOutcome
{
    Operation = "CreateOrder",
    Success = true,
    Latency = TimeSpan.FromMilliseconds(150)
});

// Track error budget
var budget = new SliBudget("order-service", targetAvailability: 99.9, TimeSpan.FromDays(30));
var remaining = budget.RemainingBudget(downtimeUsed);
```

### Reliability Patterns

```csharp
using Core.Reliability;

// Circuit Breaker
var breaker = new CircuitBreaker(new CircuitBreakerConfig
{
    Name = "external-api",
    FailureThreshold = 5,
    BreakDuration = TimeSpan.FromSeconds(30)
});

await breaker.ExecuteAsync(async () => await CallExternalApi());

// Retry with exponential backoff
var retry = new RetryPolicy(new RetryConfig
{
    Name = "database",
    MaxAttempts = 3,
    InitialDelay = TimeSpan.FromMilliseconds(100)
});

await retry.ExecuteAsync(async () => await SaveToDatabase());

// Rate Limiting
var limiter = new RateLimiter(new RateLimiterConfig
{
    Name = "api",
    RequestsPerSecond = 100,
    Burst = 10
});

if (limiter.TryAcquire())
{
    // Process request
}

// Bulkhead
var bulkhead = new Bulkhead("database", maxConcurrency: 10, timeout: TimeSpan.FromSeconds(5));
await bulkhead.ExecuteAsync(async () => await QueryDatabase());
```

### Infrastructure

```csharp
using Core.Infrastructure;

// Redis
var redis = await RedisClient.CreateAsync(new RedisConfig
{
    Host = "localhost",
    Port = 6379,
    ConnectTimeout = TimeSpan.FromSeconds(60)
});

await redis.SetAsync("key", "value");
var value = await redis.GetAsync("key");

// Health Checks
var checker = new HealthChecker(timeout: TimeSpan.FromSeconds(60));
checker.Register("redis", async ct => await redis.HealthCheckAsync());
checker.Register("liveness", new LivenessCheck());

var response = await checker.CheckAsync();
// response.Status: Healthy, Degraded, or Unhealthy
```

## Building

```bash
cd core/dotnet
dotnet build
dotnet test
```

## Dependencies

- .NET 8.0
- Serilog (logging)
- Prometheus-net (metrics)
- Polly (resilience)
- StackExchange.Redis (Redis client)
- Microsoft.Extensions.Diagnostics.HealthChecks

## Architecture Alignment

These packages mirror the Go core packages in `core/go/`:

| .NET Package | Go Package |
|--------------|------------|
| Core.Logger | logger |
| Core.Errors | errors |
| Core.Metrics | metrics |
| Core.Config | config |
| Core.Sli | sli |
| Core.Sod | sod |
| Core.Reliability | reliability |
| Core.Infrastructure | infrastructure |
