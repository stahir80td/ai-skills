# AI Patterns - Complete Go Core Package Reference Implementation

> **Comprehensive patterns demonstrating ALL Go Core package capabilities**

This reference implementation showcases optimal usage patterns for **all Go Core package components** within a single HTTP service, following AI architecture standards.

## 🎯 What This Demonstrates

### Core Package Components Used

| Package | Import Path | Use Cases | Patterns Shown |
|---------|-------------|-----------|----------------|
| **Logger** | `core/go/logger` | Structured JSON logging, correlation | Contextual logging, error correlation |
| **Errors** | `core/go/errors` | Error codes, ServiceError patterns | Error registry, SOD scoring |
| **SLI** | `core/go/sli` | SLI tracking, availability, latency | Prometheus metrics, request tracking |
| **SOD** | `core/go/sod` | Severity/Occurrence/Detectability | Risk calculation, error prioritization |
| **Metrics** | `core/go/metrics` | Four Golden Signals | Latency, traffic, errors, saturation |
| **Reliability** | `core/go/reliability` | Circuit breakers, retry, rate limiting | Resilience patterns |
| **Infrastructure** | `core/go/infrastructure/*` | Data platform clients | All 5 data platforms |

### Infrastructure Clients Integration

| Platform | Package | Use Cases | Patterns Shown |
|----------|---------|-----------|----------------|
| **SQL Server** | `infrastructure/sqlserver` | Transactional data, ACID | Orders, transactions |
| **MongoDB** | `infrastructure/mongodb` | Document storage | User profiles, flexible schemas |
| **ScyllaDB** | `infrastructure/scylladb` | Time-series, high throughput | IoT telemetry, aggregations |
| **Redis** | `infrastructure/redis` | Caching, real-time data | Leaderboards, sessions |
| **Kafka** | `infrastructure/kafka` | Event streaming | Domain events, pub/sub |

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    AI PATTERNS SERVICE                     │
│              (All Go Core Package Components)               │
└─────────────────────────────────────────────────────────────┘
                                │
                    ┌───────────┴───────────┐
                    │     HTTP Handlers     │
                    │   (PatternsHandler)   │
                    └───────────┬───────────┘
                                │
                    ┌───────────┴───────────┐
                    │    PatternsService    │
                    │   (Business Logic)    │
                    └───────────┬───────────┘
                                │
        ┌───────┬───────┬───────┼───────┬───────┬───────┐
        │       │       │       │       │       │       │
        ▼       ▼       ▼       ▼       ▼       ▼       ▼
   ┌─────────┐ │ ┌─────────┐ │ ┌─────────┐ │ ┌─────────┐
   │   SQL   │ │ │ MongoDB │ │ │ScyllaDB │ │ │  Redis  │
   │ Server  │ │ │         │ │ │         │ │ │         │
   └─────────┘ │ └─────────┘ │ └─────────┘ │ └─────────┘
   Orders,     │ Users,      │ Telemetry,  │ Cache,
   Transactions│ Profiles    │ Time-series │ Sessions
               │             │             │
               ▼             ▼             ▼
           ┌─────────────────────────────────────┐
           │              Kafka                  │
           │         (Event Streaming)           │
           └─────────────────────────────────────┘
           Domain Events, Notifications, Pub/Sub
```

## 🚀 Quick Start

### 1. Start Dependencies

```bash
# Start all data platforms (Docker Compose recommended)
docker run -d --name sqlserver -p 1433:1433 -e ACCEPT_EULA=Y -e SA_PASSWORD=AiPatterns2024! mcr.microsoft.com/mssql/server:2022-latest
docker run -d --name mongodb -p 27017:27017 mongo:7
docker run -d --name redis -p 6379:6379 redis:7-alpine
docker run -d --name scylla -p 9042:9042 scylladb/scylla:5.4.3 --smp 1
docker run -d --name kafka -p 9092:9092 apache/kafka:3.7.0
```

### 2. Build and Run

```bash
# Navigate to patterns directory
cd patterns/go

# Download dependencies
go mod download

# Run the service
go run ./cmd/patterns

# Service available at http://localhost:8080
```

### 3. Test All Patterns

```bash
# Health check
curl http://localhost:8080/health

# Create an order (SQL Server pattern)
curl -X POST http://localhost:8080/api/v1/patterns/orders \
  -H "Content-Type: application/json" \
  -d '{"customerId":"123e4567-e89b-12d3-a456-426614174000","shippingAddress":"123 Main St","items":[{"productName":"Widget","quantity":2,"unitPrice":29.99}]}'

# Create user profile (MongoDB pattern)
curl -X POST http://localhost:8080/api/v1/patterns/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","firstName":"John","lastName":"Doe"}'

# Record telemetry (ScyllaDB pattern)
curl -X POST http://localhost:8080/api/v1/patterns/telemetry \
  -H "Content-Type: application/json" \
  -d '{"deviceId":"device-001","metric":"temperature","value":72.5,"unit":"fahrenheit"}'

# Update leaderboard (Redis pattern)
curl -X POST http://localhost:8080/api/v1/patterns/leaderboards/gaming/scores \
  -H "Content-Type: application/json" \
  -d '{"userId":"player-001","score":1500}'

# Prometheus metrics
curl http://localhost:8080/metrics
```

## 📊 API Endpoints

### SQL Server Patterns (Transactional)

```http
POST   /api/v1/patterns/orders              # Create order with items
PATCH  /api/v1/patterns/orders/{id}/status  # Update order status
GET    /api/v1/patterns/orders/{id}         # Get order details
```

### MongoDB Patterns (Document)

```http
POST   /api/v1/patterns/users                    # Create user profile
PUT    /api/v1/patterns/users/{id}/preferences   # Update preferences
GET    /api/v1/patterns/users/{id}               # Get user profile
```

### ScyllaDB Patterns (Time-Series)

```http
POST   /api/v1/patterns/telemetry            # Record device telemetry
GET    /api/v1/patterns/telemetry/{deviceId} # Get telemetry history
```

### Redis Patterns (Real-time)

```http
POST   /api/v1/patterns/leaderboards/{category}/scores # Update leaderboard
GET    /api/v1/patterns/leaderboards/{category}        # Get leaderboard
POST   /api/v1/patterns/sessions                       # Create session
```

### Cross-Platform Analytics

```http
GET    /api/v1/patterns/analytics            # Query all platforms
```

### Health & Monitoring

```http
GET    /health           # Connectivity check for all platforms
GET    /health/live      # Liveness probe
GET    /health/ready     # Readiness probe
GET    /metrics          # Prometheus metrics
```

## 📁 Project Structure (Service Oriented Design)

```
patterns/go/
├── cmd/
│   └── patterns/
│       └── main.go                 # Entry point, dependency injection
│
├── internal/
│   ├── api/
│   │   ├── handlers.go             # HTTP handlers
│   │   ├── middleware.go           # Middleware (correlation, metrics)
│   │   └── routes.go               # Route definitions
│   │
│   ├── domain/
│   │   ├── models/                 # Domain entities
│   │   │   ├── order.go            # SQL Server entity
│   │   │   ├── user_profile.go     # MongoDB document
│   │   │   ├── telemetry.go        # ScyllaDB time-series
│   │   │   └── events.go           # Kafka events
│   │   │
│   │   ├── services/
│   │   │   └── patterns_service.go # Cross-platform business logic
│   │   │
│   │   ├── errors/
│   │   │   └── product_errors.go   # Error repository with SOD
│   │   │
│   │   └── sli/
│   │       └── patterns_sli.go     # SLI tracking
│   │
│   └── infrastructure/
│       ├── repositories/
│       │   ├── order_repository.go      # SQL Server
│       │   ├── user_profile_repository.go # MongoDB
│       │   └── telemetry_repository.go  # ScyllaDB
│       │
│       ├── cache/
│       │   └── realtime_cache.go        # Redis
│       │
│       └── messaging/
│           └── event_publisher.go       # Kafka
│
├── config/
│   ├── config.go                   # Configuration loading
│   └── config.yaml                 # Environment configuration
│
├── go.mod
├── go.sum
└── README.md
```

## 🔧 Configuration

Configuration via `config/config.yaml`:

```yaml
service:
  name: ai-patterns
  version: 1.0.0
  port: 8080

logging:
  level: info
  environment: development

sqlserver:
  server: localhost:1433
  database: AiPatterns
  user: sa
  password: AiPatterns2024!
  ping_timeout: 60s

mongodb:
  connection_uri: mongodb://localhost:27017
  database: AiPatternsDB
  ping_timeout: 60s

scylladb:
  hosts:
    - localhost
  keyspace: ai_patterns
  timeout: 60s
  connect_timeout: 60s

redis:
  host: localhost
  port: 6379
  ping_timeout: 60s

kafka:
  brokers:
    - localhost:9092
```

## 🎭 Pattern Examples

### Cross-Platform Workflow Example

When creating an order, the system demonstrates all platforms working together:

```go
func (s *PatternsService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    // Create context logger with correlation
    log := s.logger.WithContext(ctx)
    log.Info("Creating order", zap.String("customer_id", req.CustomerID.String()))

    // 1. SQL Server - Transactional storage
    order := NewOrder(req.CustomerID, req.ShippingAddress, req.Items)
    if err := s.orderRepo.Create(ctx, order); err != nil {
        s.sli.RecordRequest(ctx, sli.RequestOutcome{Operation: "create_order", Success: false, ErrorCode: "ORD-001"})
        return nil, productErrors.DatabaseError(err)
    }

    // 2. Redis - Cache for performance
    if err := s.cache.Set(ctx, fmt.Sprintf("order:%s", order.ID), order, 24*time.Hour); err != nil {
        log.Warn("Failed to cache order", zap.Error(err))
    }

    // 3. Kafka - Event publishing
    event := NewOrderCreatedEvent(order)
    if err := s.eventPublisher.PublishOrderEvent(ctx, event); err != nil {
        log.Warn("Failed to publish order event", zap.Error(err))
    }

    // 4. Track SLI
    s.sli.RecordRequest(ctx, sli.RequestOutcome{Operation: "create_order", Success: true})

    return order, nil
}
```

### Contextual Logging Pattern

```go
// Create logger with context (correlation ID, component)
ctx = context.WithValue(ctx, logger.CorrelationIDKey, uuid.New().String())
ctx = context.WithValue(ctx, logger.ComponentKey, "OrderService")

log := logger.WithContext(ctx)
log.Info("Processing order",
    zap.String("order_id", orderID),
    zap.String("status", "processing"))
```

### Error Repository Pattern

```go
// Register errors with SOD scoring
var ProductErrors = errors.NewErrorRegistry()

func init() {
    ProductErrors.Register(&errors.ErrorDefinition{
        Code:        "PAT-PRD-001",
        Severity:    errors.SeverityMedium,
        Description: "Product not found: %s",
        SODScore:    120, // 4 × 5 × 6
        Severity_S:  4,
        Occurrence:  5,
        Detect_D:    6,
        Mitigation:  "Verify product ID exists in database",
    })
}

// Create error from registry
err := ProductErrors.CreateError("PAT-PRD-001", productID)
```

### SLI Tracking Pattern

```go
// Track request outcomes
sliTracker.RecordRequest(ctx, sli.RequestOutcome{
    Operation:     "get_order",
    Success:       true,
    Latency:       time.Since(start),
    ErrorCode:     "",
    ErrorSeverity: "",
})

// Track latency
sliTracker.RecordLatency(ctx, duration, "database_query")

// Track throughput
sliTracker.RecordThroughput(ctx, messageCount, "kafka_publish")
```

### Circuit Breaker Pattern

```go
// Create circuit breaker
cb := reliability.NewCircuitBreaker("mongodb", 5, 30*time.Second)

// Execute with circuit breaker protection
err := cb.ExecuteWithContext(ctx, func(ctx context.Context) error {
    return mongoClient.InsertOne(ctx, collection, document)
})

if err != nil {
    log.WithError("INFRA-CB-001", errors.SeverityHigh).
        Error("Circuit breaker triggered", zap.Error(err))
}
```

### Prometheus Metrics Pattern

```go
// Create service metrics (Four Golden Signals)
metrics := metrics.NewServiceMetrics(metrics.Config{
    ServiceName: "ai-patterns",
    Namespace:   "iot_your-org",
})

// Record request (Latency + Traffic)
metrics.RecordRequest("POST", "/api/v1/orders", "200", duration)

// Record error (Errors)
metrics.RecordError("PAT-PRD-001", "MEDIUM", "order-service")

// Update saturation
metrics.UpdateResourceUtilization("database_connections", 75.0)
metrics.IncActiveRequests()
defer metrics.DecActiveRequests()
```

## 📈 Prometheus Metrics Exported

```
# Four Golden Signals
iot_your-org_request_duration_seconds{service,method,endpoint,status}
iot_your-org_requests_total{service,method,endpoint,status}
iot_your-org_errors_total{service,error_code,severity,component}
iot_your-org_resource_utilization{service,resource_type}
iot_your-org_active_requests{service}

# SLI Metrics
sli_requests_total{service,operation}
sli_requests_success_total{service,operation}
sli_requests_failed_total{service,operation,error_code,severity}
sli_request_duration_seconds{service,operation}
sli_availability_percent{service}
sli_latency_p95_milliseconds{service,operation}
sli_error_rate_percent{service}

# Circuit Breaker Metrics
circuit_breaker_state{name}
circuit_breaker_requests_total{name,state,result}
circuit_breaker_errors_total{name}
circuit_breaker_state_changes_total{name,from,to}
```

## 🔍 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/domain/services/...

# Run integration tests (requires infrastructure)
go test -tags=integration ./...
```
