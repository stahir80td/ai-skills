---
name: ai-sli-middleware
description: >
  Service Level Indicator (SLI) patterns using Core.Sli for .NET and core/go/sli for Go.
  Use when adding metrics, SLI tracking, or implementing the /api/v1/sli endpoint.
  Ensures all services track availability, latency, and throughput on every request.
  Required for SRE observability and dashboards.
---

# AI SLI (Service Level Indicators) Patterns

## What SLI Tracks

| Metric | Description | Target |
|--------|-------------|--------|
| **Availability** | % of successful requests (non-5xx) | 99.9% |
| **Latency** | p50, p95, p99 response times | p99 < 500ms |
| **Throughput** | Requests per second | Varies by service |
| **Error Rate** | % of failed requests | < 0.1% |

---

## .NET SLI Implementation

### Program.cs Registration

```csharp
using Core.Sli;

// Register SLI Tracker
builder.Services.AddSingleton<SliTracker>();
```

### SLI Middleware (REQUIRED)

```csharp
using Core.Sli;
using System.Diagnostics;

namespace MyService.Api.Middleware;

public class SliMiddleware
{
    private readonly RequestDelegate _next;
    private readonly SliTracker _sliTracker;

    public SliMiddleware(RequestDelegate next, SliTracker sliTracker)
    {
        _next = next;
        _sliTracker = sliTracker;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        var stopwatch = Stopwatch.StartNew();
        var endpoint = $"{context.Request.Method} {context.Request.Path}";
        var success = true;

        try
        {
            await _next(context);
            
            // 5xx = failure for SLI purposes
            if (context.Response.StatusCode >= 500)
            {
                success = false;
            }
        }
        catch
        {
            success = false;
            throw;
        }
        finally
        {
            stopwatch.Stop();
            _sliTracker.RecordRequest(
                endpoint,
                stopwatch.ElapsedMilliseconds,
                success);
        }
    }
}
```

### Middleware Pipeline (ORDER MATTERS!)

```csharp
var app = builder.Build();

// Middleware order:
app.UseMiddleware<CorrelationIdMiddleware>();  // 1. First - captures correlation ID
app.UseMiddleware<ErrorHandlerMiddleware>();   // 2. Second - catches errors
app.UseMiddleware<SliMiddleware>();            // 3. Third - records metrics

app.MapControllers();
```

### SLI Controller (REQUIRED)

```csharp
using Core.Sli;
using Microsoft.AspNetCore.Mvc;

namespace MyService.Api.Controllers;

[ApiController]
[Route("api/v1/sli")]
public class SliController : ControllerBase
{
    private readonly SliTracker _sliTracker;

    public SliController(SliTracker sliTracker)
    {
        _sliTracker = sliTracker;
    }

    /// <summary>
    /// Get current SLI metrics for the service
    /// </summary>
    [HttpGet]
    public IActionResult GetSli()
    {
        var metrics = _sliTracker.GetMetrics();
        
        return Ok(new
        {
            service = "my-service",
            timestamp = DateTime.UtcNow,
            sli = new
            {
                availability = new
                {
                    current = metrics.Availability,
                    target = 99.9
                },
                latency = new
                {
                    p50_ms = metrics.LatencyP50,
                    p95_ms = metrics.LatencyP95,
                    p99_ms = metrics.LatencyP99,
                    target_p99_ms = 500
                },
                throughput = new
                {
                    requests_per_second = metrics.RequestsPerSecond,
                    total_requests = metrics.TotalRequests
                },
                errors = new
                {
                    error_rate = metrics.ErrorRate,
                    total_errors = metrics.TotalErrors
                }
            }
        });
    }

    /// <summary>
    /// Get SLI metrics by endpoint
    /// </summary>
    [HttpGet("endpoints")]
    public IActionResult GetEndpointSli()
    {
        var endpointMetrics = _sliTracker.GetEndpointMetrics();
        
        return Ok(new
        {
            service = "my-service",
            timestamp = DateTime.UtcNow,
            endpoints = endpointMetrics.Select(e => new
            {
                endpoint = e.Endpoint,
                availability = e.Availability,
                latency_p99_ms = e.LatencyP99,
                requests_per_second = e.RequestsPerSecond
            })
        });
    }
}
```

### Custom Business SLI

```csharp
public class OrderService
{
    private readonly SliTracker _sliTracker;

    public async Task<Order> CreateOrder(CreateOrderRequest request)
    {
        var stopwatch = Stopwatch.StartNew();
        var success = true;

        try
        {
            var order = await _repository.CreateAsync(request);
            return order;
        }
        catch
        {
            success = false;
            throw;
        }
        finally
        {
            stopwatch.Stop();
            
            // Track business operation SLI
            _sliTracker.RecordBusinessOperation(
                "order.create",
                stopwatch.ElapsedMilliseconds,
                success);
        }
    }
}
```

---

## Go SLI Implementation

### Setup in main.go

```go
import (
    "github.com/your-github-org/ai-scaffolder/core/go/sli"
    "github.com/your-github-org/ai-scaffolder/core/go/metrics"
)

func main() {
    // Create SLI tracker
    sliTracker := sli.NewPrometheusTracker("order-service")
    
    // Create service metrics
    serviceMetrics := metrics.NewServiceMetrics(metrics.Config{
        ServiceName: "order-service",
        Namespace:   "AI",
    })
    
    // Pass to server
    server := api.NewServer(api.ServerConfig{
        SliTracker: sliTracker,
        Metrics:    serviceMetrics,
    })
}
```

### SLI Middleware (REQUIRED)

```go
package middleware

import (
    "net/http"
    "time"

    "github.com/your-github-org/ai-scaffolder/core/go/sli"
)

func SLI(tracker *sli.PrometheusTracker) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Wrap response writer to capture status code
            rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            
            next.ServeHTTP(rw, r)
            
            duration := time.Since(start)
            endpoint := r.Method + " " + r.URL.Path
            success := rw.statusCode < 500
            
            tracker.RecordRequest(endpoint, duration.Milliseconds(), success)
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

### Middleware Chain (ORDER MATTERS!)

```go
func (s *Server) routes() http.Handler {
    r := mux.NewRouter()
    
    // Apply middleware in order:
    r.Use(middleware.CorrelationID)     // 1. First
    r.Use(middleware.ErrorHandler(s.logger))  // 2. Second
    r.Use(middleware.SLI(s.sliTracker)) // 3. Third
    r.Use(middleware.Logging(s.logger))
    
    // Routes...
    return r
}
```

### SLI Handler (REQUIRED)

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/your-github-org/ai-scaffolder/core/go/sli"
)

type SLIHandler struct {
    tracker *sli.PrometheusTracker
}

func NewSLIHandler(tracker *sli.PrometheusTracker) *SLIHandler {
    return &SLIHandler{tracker: tracker}
}

type SLIResponse struct {
    Service   string    `json:"service"`
    Timestamp time.Time `json:"timestamp"`
    SLI       SLIMetrics `json:"sli"`
}

type SLIMetrics struct {
    Availability AvailabilityMetrics `json:"availability"`
    Latency      LatencyMetrics      `json:"latency"`
    Throughput   ThroughputMetrics   `json:"throughput"`
    Errors       ErrorMetrics        `json:"errors"`
}

type AvailabilityMetrics struct {
    Current float64 `json:"current"`
    Target  float64 `json:"target"`
}

type LatencyMetrics struct {
    P50Ms      float64 `json:"p50_ms"`
    P95Ms      float64 `json:"p95_ms"`
    P99Ms      float64 `json:"p99_ms"`
    TargetP99  float64 `json:"target_p99_ms"`
}

type ThroughputMetrics struct {
    RequestsPerSecond float64 `json:"requests_per_second"`
    TotalRequests     int64   `json:"total_requests"`
}

type ErrorMetrics struct {
    ErrorRate   float64 `json:"error_rate"`
    TotalErrors int64   `json:"total_errors"`
}

func (h *SLIHandler) GetSLI(w http.ResponseWriter, r *http.Request) {
    metrics := h.tracker.GetMetrics()
    
    response := SLIResponse{
        Service:   "order-service",
        Timestamp: time.Now().UTC(),
        SLI: SLIMetrics{
            Availability: AvailabilityMetrics{
                Current: metrics.Availability,
                Target:  99.9,
            },
            Latency: LatencyMetrics{
                P50Ms:     metrics.LatencyP50,
                P95Ms:     metrics.LatencyP95,
                P99Ms:     metrics.LatencyP99,
                TargetP99: 500,
            },
            Throughput: ThroughputMetrics{
                RequestsPerSecond: metrics.RequestsPerSecond,
                TotalRequests:     metrics.TotalRequests,
            },
            Errors: ErrorMetrics{
                ErrorRate:   metrics.ErrorRate,
                TotalErrors: metrics.TotalErrors,
            },
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### Custom Business SLI

```go
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    start := time.Now()
    success := true
    
    defer func() {
        duration := time.Since(start)
        s.sliTracker.RecordBusinessOperation(
            "order.create",
            duration.Milliseconds(),
            success)
    }()
    
    order, err := s.repo.Create(ctx, req)
    if err != nil {
        success = false
        return nil, err
    }
    
    return order, nil
}
```

---

## Required Endpoints Summary

Every AI service MUST expose:

| Endpoint | Purpose | Response |
|----------|---------|----------|
| `/health/live` | Kubernetes liveness probe | `{ "status": "healthy" }` |
| `/health/ready` | Kubernetes readiness probe | `{ "status": "ready" }` |
| `/metrics` | Prometheus scraping | Prometheus format |
| `/api/v1/sli` | SLI dashboard data | JSON with availability, latency, throughput |

---

## SLI Response Example

```json
{
  "service": "order-service",
  "timestamp": "2024-01-15T10:30:45Z",
  "sli": {
    "availability": {
      "current": 99.95,
      "target": 99.9
    },
    "latency": {
      "p50_ms": 45,
      "p95_ms": 120,
      "p99_ms": 280,
      "target_p99_ms": 500
    },
    "throughput": {
      "requests_per_second": 150.5,
      "total_requests": 1250000
    },
    "errors": {
      "error_rate": 0.05,
      "total_errors": 625
    }
  }
}
```
