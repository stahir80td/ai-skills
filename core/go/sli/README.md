# SLI/SLO Framework

Production-ready Service Level Indicator (SLI) and Service Level Objective (SLO) tracking framework with **REAL** error budget calculation using Prometheus metrics.

## Overview

This framework provides:
- **Real-time SLI tracking** using actual Prometheus metrics
- **Error budget calculation** from real request/error counts
- **Multi-window burn rate alerts** (fast/slow burn detection)
- **SLO compliance validation** across availability, latency, and error rate
- **Prometheus integration** with standardized metrics

## Key Features

### 1. Real Error Budget Tracking

Error budgets are calculated from **ACTUAL** Prometheus data:

```
Total Requests: FROM prometheus (http_requests_total)
Actual Errors: FROM prometheus (http_requests_failed_total)
Allowed Errors = Total Requests × (1 - SLO Target)
Remaining Budget = Allowed Errors - Actual Errors
```

**Example:**
- Service: api-gateway, 30-day window
- Total Requests: 10,000,000 (from Prometheus)
- SLO Target: 99.9%
- Allowed Errors: 10,000,000 × 0.001 = 10,000
- Actual Errors: 2,500 (from Prometheus)
- Remaining Budget: 7,500 (75% remaining)
- Burn Rate: 83 errors/day
- Time to Exhaustion: 90 days

### 2. Multi-Window Burn Rate Alerts

Alerts fire when error budget is consumed too quickly:

**Fast Burn (Critical)**
- Window: 1 hour
- Threshold: 2% of budget consumed
- Action: **PAGE ON-CALL** immediately

**Slow Burn (Warning)**
- Window: 6 hours
- Threshold: 5% of budget consumed
- Action: Alert team, review deployments

### 3. SLI Tracking

Tracks three core SLIs using real metrics:

**Availability**
- Formula: `SuccessRequests / TotalRequests`
- Target: 99.9% (configurable)

**Latency**
- Metrics: P95, P99 from Prometheus histograms
- Target: P95 < 200ms, P99 < 500ms (configurable)

**Error Rate**
- Formula: `FailedRequests / TotalRequests`
- Target: < 0.1% (configurable)

## Installation

```bash
go get github.com/your-org/shared/sli
```

## Quick Start

### 1. Create SLO Configuration

Create `config/slo.yaml`:

```yaml
service_name: "api-gateway"
environment: "production"

# Availability SLO
availability_slo:
  enabled: true
  target: 0.999  # 99.9%

# Latency SLO
latency_slo:
  enabled: true
  p95_milliseconds: 200
  p99_milliseconds: 500

# Error Rate SLO
error_rate_slo:
  enabled: true
  max_error_rate: 0.001  # 0.1%

# Error Budget
error_budget:
  window: 720h  # 30 days
  fast_burn_window: 1h
  fast_burn_percent: 2.0
  slow_burn_window: 6h
  slow_burn_percent: 5.0
```

### 2. Initialize in Service

```go
package main

import (
    "github.com/your-org/shared/sli"
)

func main() {
    // Create factory
    factory := sli.NewFactory("api-gateway", "production")
    
    // Create all components (auto-loads config)
    manager, err := factory.CreateAll()
    if err != nil {
        log.Fatalf("Failed to create SLI manager: %v", err)
    }
    
    // Create helper for simplified usage
    helper := sli.NewHelper(manager)
    
    // Use in your service...
}
```

### 3. Track Requests

```go
import (
    "time"
    "github.com/your-org/shared/sli"
)

// In HTTP handler
func handleRequest(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // Process request
    err := processRequest(r)
    
    // Record outcome
    helper.TrackRequest(sli.RequestOutcome{
        Success:   err == nil,
        Latency:   time.Since(start),
        Operation: r.Method + " " + r.URL.Path,
        ErrorCode: getErrorCode(err),
        ErrorSeverity: getErrorSeverity(err),
    })
    
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    w.WriteHeader(200)
}
```

### 4. Check Error Budget

```go
import (
    "context"
    "log"
)

// Periodic budget check (e.g., every 5 minutes)
func monitorErrorBudget(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Check for burn rate alerts
            alerts, err := helper.ShouldAlert()
            if err != nil {
                log.Printf("Error checking alerts: %v", err)
                continue
            }
            
            // Handle alerts
            for _, alert := range alerts {
                if alert.Severity == "critical" {
                    // PAGE ON-CALL
                    pageOnCall(alert)
                } else {
                    // Send team alert
                    alertTeam(alert)
                }
                
                log.Printf("[%s] %s - %s", 
                    alert.Severity, alert.Message, alert.RecommendedAction)
            }
            
            // Log budget status
            budget, _ := helper.CheckBudget()
            log.Printf("Error Budget: %.1f%% remaining (%s)", 
                budget.BudgetPercent, sli.GetBudgetStatus(budget))
                
        case <-ctx.Done():
            return
        }
    }
}
```

### 5. Validate SLO Compliance

```go
import "log"

// Periodic SLO validation (e.g., hourly)
func validateSLOCompliance(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            compliance, err := helper.CheckCompliance()
            if err != nil {
                log.Printf("Error checking compliance: %v", err)
                continue
            }
            
            for sliName, comp := range compliance {
                if !comp.InCompliance {
                    log.Printf("❌ SLO VIOLATION: %s - %s", sliName, comp.Message)
                } else {
                    log.Printf("✓ SLO OK: %s - %s", sliName, comp.Message)
                }
            }
            
        case <-ctx.Done():
            return
        }
    }
}
```

## Prometheus Metrics

The framework exports standardized Prometheus metrics:

### Request Metrics
```
# Total requests
sli_requests_total{service="api-gateway", operation="GET /devices"}

# Success requests
sli_requests_success_total{service="api-gateway", operation="GET /devices"}

# Failed requests
sli_requests_failed_total{service="api-gateway", operation="GET /devices", error_code="AUTH-001", severity="high"}
```

### Latency Metrics
```
# Request duration histogram
sli_request_duration_seconds{service="api-gateway", operation="GET /devices"}
```

### SLI Gauges
```
# Current availability percentage
sli_availability_percent{service="api-gateway"} 99.95

# P95 latency in milliseconds
sli_latency_p95_milliseconds{service="api-gateway", operation="GET /devices"} 185

# P99 latency in milliseconds
sli_latency_p99_milliseconds{service="api-gateway", operation="GET /devices"} 420

# Current error rate percentage
sli_error_rate_percent{service="api-gateway"} 0.05
```

## PromQL Queries

Use these queries in Grafana or Alertmanager:

### Availability
```promql
# Current availability (5m window)
(sum(rate(sli_requests_success_total{service="api-gateway"}[5m])) / 
 sum(rate(sli_requests_total{service="api-gateway"}[5m]))) * 100

# 30-day availability
(sum(rate(sli_requests_success_total{service="api-gateway"}[30d])) / 
 sum(rate(sli_requests_total{service="api-gateway"}[30d]))) * 100
```

### Error Budget
```promql
# Remaining error budget (30-day window)
(sum(rate(sli_requests_total{service="api-gateway"}[30d])) * (1 - 0.999)) - 
sum(rate(sli_requests_failed_total{service="api-gateway"}[30d]))

# Burn rate (errors per hour)
sum(rate(sli_requests_failed_total{service="api-gateway"}[1h])) * 3600
```

### Latency
```promql
# P95 latency
histogram_quantile(0.95, 
  rate(sli_request_duration_seconds_bucket{service="api-gateway"}[5m]))

# P99 latency
histogram_quantile(0.99, 
  rate(sli_request_duration_seconds_bucket{service="api-gateway"}[5m]))
```

### Error Rate
```promql
# Current error rate
(sum(rate(sli_requests_failed_total{service="api-gateway"}[5m])) / 
 sum(rate(sli_requests_total{service="api-gateway"}[5m]))) * 100
```

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────┐
│                      Factory                            │
│  Creates all components with dependency injection       │
└─────────────────────────────────────────────────────────┘
                            │
           ┌────────────────┼────────────────┐
           ▼                ▼                ▼
    ┌──────────┐    ┌──────────────┐  ┌────────────┐
    │ Tracker  │    │ BudgetTracker│  │ Validator  │
    └──────────┘    └──────────────┘  └────────────┘
           │                │                │
           └────────────────┼────────────────┘
                            ▼
                    ┌──────────────┐
                    │  Prometheus  │
                    │   Metrics    │
                    └──────────────┘
```

### Data Flow

```
HTTP Request → RecordRequest() → Prometheus Counters/Histograms
                                         ↓
                                  GetMetrics()
                                         ↓
                         ┌───────────────┴──────────────┐
                         ▼                              ▼
                  CalculateBudget()            ValidateCompliance()
                         │                              │
                         ▼                              ▼
                  Burn Rate Alerts              SLO Status Reports
```

## Error Budget Decision Making

Use error budget to guide deployment decisions:

### Budget Status → Actions

| Remaining Budget | Status      | Action                                    |
|------------------|-------------|-------------------------------------------|
| > 50%            | Healthy     | Deploy freely, iterate quickly            |
| 25-50%           | Warning     | Deploy cautiously, increase monitoring    |
| 10-25%           | Critical    | Deploy only critical fixes                |
| < 10%            | Exhausted   | **FREEZE** all non-critical deployments   |

### Burn Rate → Alerts

| Window  | Threshold | Severity  | Action                                  |
|---------|-----------|-----------|----------------------------------------|
| 1 hour  | 2% budget | Critical  | **PAGE** on-call, stop all deployments |
| 6 hours | 5% budget | Warning   | Alert team, review recent changes      |

## Configuration Options

### File-based (config/slo.yaml)
```yaml
service_name: "api-gateway"
availability_slo:
  enabled: true
  target: 0.999
latency_slo:
  enabled: true
  p95_milliseconds: 200
  p99_milliseconds: 500
error_rate_slo:
  enabled: true
  max_error_rate: 0.001
error_budget:
  window: 720h
  fast_burn_window: 1h
  fast_burn_percent: 2.0
  slow_burn_window: 6h
  slow_burn_percent: 5.0
```

### Environment Variables
```bash
SLO_AVAILABILITY_TARGET=0.999
SLO_LATENCY_P95_MS=200
SLO_LATENCY_P99_MS=500
SLO_ERROR_RATE_MAX=0.001
SLO_ERROR_BUDGET_WINDOW_HOURS=720
```

### Programmatic
```go
config := &sli.SLOConfig{
    ServiceName: "api-gateway",
    AvailabilitySLO: sli.AvailabilitySLO{
        Enabled: true,
        Target:  0.999,
    },
    // ... other settings
}
```

## Best Practices

1. **Set Realistic SLOs**
   - Start with current performance baseline
   - Gradually tighten targets over time
   - Don't set 100% - it's impossible and wasteful

2. **Monitor Burn Rate**
   - Run burn rate checks every 5 minutes
   - Page on fast burn (1h window)
   - Alert on slow burn (6h window)

3. **Use Error Budget for Decisions**
   - Deploy freely when budget is healthy (>50%)
   - Freeze deployments when exhausted (<10%)
   - Use budget consumption to prioritize reliability work

4. **Track by Operation**
   - Record operation name with each request
   - Separate SLIs for critical vs non-critical endpoints
   - Use operation-level latency tracking

5. **Integrate with CI/CD**
   - Check error budget before deploying
   - Auto-rollback if budget drops rapidly
   - Require SRE approval when budget is low

## Example: Full Integration

```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"
    
    "github.com/your-org/shared/sli"
)

var sliHelper *sli.Helper

func main() {
    // Initialize SLI framework
    factory := sli.NewFactory("api-gateway", "production")
    manager, err := factory.CreateAll()
    if err != nil {
        log.Fatalf("Failed to create SLI manager: %v", err)
    }
    sliHelper = sli.NewHelper(manager)
    
    // Start background monitoring
    ctx := context.Background()
    go monitorErrorBudget(ctx)
    go validateSLOCompliance(ctx)
    
    // Start HTTP server
    http.HandleFunc("/devices", trackSLI(handleDevices))
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// Middleware to track SLIs
func trackSLI(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Process request
        next(w, r)
        
        // Record SLI
        sliHelper.TrackRequest(sli.RequestOutcome{
            Success:   w.Header().Get("X-Error") == "",
            Latency:   time.Since(start),
            Operation: r.Method + " " + r.URL.Path,
        })
    }
}

func monitorErrorBudget(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            alerts, _ := sliHelper.ShouldAlert()
            for _, alert := range alerts {
                if alert.Severity == "critical" {
                    pageOnCall(alert)
                } else {
                    alertTeam(alert)
                }
            }
        case <-ctx.Done():
            return
        }
    }
}
```

## Integration with SOD Framework

Combine SLI/SLO tracking with SOD scoring for complete observability:

```go
import (
    "github.com/your-org/shared/sli"
    "github.com/your-org/shared/sod"
)

func handleError(ctx context.Context, err error) {
    // Calculate SOD score
    sodScore, _ := sodHelper.CalculateForError(ctx, getErrorCode(err))
    
    // Track SLI
    sliHelper.TrackRequest(sli.RequestOutcome{
        Success:       false,
        ErrorCode:     getErrorCode(err),
        ErrorSeverity: getSeverityFromSOD(sodScore),
    })
    
    // Check error budget
    budget, _ := sliHelper.CheckBudget()
    
    // Decision logic
    if sod.ShouldPage(sodScore, thresholds) || budget.BudgetPercent < 10 {
        pageOnCall(err, sodScore, budget)
    }
}
```

## License

Internal use only - your-org Platform Team
