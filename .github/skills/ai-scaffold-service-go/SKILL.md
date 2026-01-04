---
name: ai-scaffold-service-go
description: >
  Creates a complete Go microservice from scratch following AI patterns.
  Use when asked to create, scaffold, or generate a new Go service, API, or microservice.
  Generates complete project structure with cmd, internal/domain, and internal/infrastructure layers.
  Includes all Core package integrations, middleware, Helm charts, and Docker setup.
---

# AI Go Service Scaffolding

## Pre-Scaffolding Questions

Before generating, ask the user:

```
Before I scaffold your Go service, I need to confirm:

1. **Service Name**: What should we call this service?
   (e.g., "order-service", "inventory-manager")

2. **Project Location**: Where should I create the project?
   (e.g., "C:\dev\order-service" - must be NEW folder)

3. **Data Stores**: Which do you need?
   - [ ] SQL Server (transactional)
   - [ ] MongoDB (documents)
   - [ ] ScyllaDB (time-series)
   - [ ] Redis (caching)
   - [ ] Kafka (events)

4. **Special Requirements**?
   (e.g., "real-time updates", "10K req/sec", "ML integration")
```

---

## Project Structure

```
{service-name}/
├── go.mod
├── go.sum
├── Dockerfile
├── Taskfile.yml
├── docker-compose.yml
├── cmd/
│   └── {service}/
│       └── main.go                       # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go
│   │   ├── routes.go
│   │   ├── server.go
│   │   └── middleware/
│   │       ├── correlation.go
│   │       ├── error_handler.go
│   │       ├── logging.go
│   │       └── sli.go
│   ├── domain/
│   │   ├── models/
│   │   ├── services/
│   │   ├── errors/
│   │   └── sli/
│   └── infrastructure/
│       ├── repositories/
│       ├── publishers/
│       └── cache/
├── config/
│   ├── config.go
│   └── config.yaml
├── scripts/
│   ├── seed-database.sql
│   └── create-kafka-topics.sh
└── helm/{service-name}/
```

---

## Required Files

### go.mod

```go
module github.com/your-org/{service-name}

go 1.22

require (
    github.com/your-github-org/ai-scaffolder/core/go v1.0.3
    github.com/gorilla/mux v1.8.1
    go.uber.org/zap v1.27.0
)
```

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./cmd/{service}

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /service .
COPY config/config.yaml ./config/

EXPOSE 8080
CMD ["./service"]
```

---

## main.go Template

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/kafka"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/mongodb"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/redis"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/sqlserver"
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
    "github.com/your-github-org/ai-scaffolder/core/go/metrics"
    "github.com/your-github-org/ai-scaffolder/core/go/sli"
    "go.uber.org/zap"
    
    "{module}/config"
    "{module}/internal/api"
    "{module}/internal/domain/services"
    "{module}/internal/infrastructure/repositories"
)

func main() {
    // ========================================
    // Logger Setup
    // ========================================
    log, err := logger.NewProduction("{service-name}", "1.0.0")
    if err != nil {
        panic(err)
    }
    defer log.Sync()
    
    log.Info("Starting {service-name}")

    // ========================================
    // Configuration
    // ========================================
    cfg, err := config.Load("config/config.yaml")
    if err != nil {
        log.Fatal("Failed to load config", zap.Error(err))
    }

    // ========================================
    // Infrastructure Clients
    // ========================================
    
    // SQL Server
    sqlClient, err := sqlserver.NewClient(sqlserver.ClientConfig{
        Server:   cfg.SQLServer.Server,
        Database: cfg.SQLServer.Database,
        User:     cfg.SQLServer.User,
        Password: cfg.SQLServer.Password,
        Logger:   log,
    })
    if err != nil {
        log.Fatal("Failed to connect to SQL Server", zap.Error(err))
    }
    
    // Redis
    redisClient, err := redis.NewClient(redis.ClientConfig{
        Host:   cfg.Redis.Host,
        Port:   cfg.Redis.Port,
        Logger: log,
    })
    if err != nil {
        log.Fatal("Failed to connect to Redis", zap.Error(err))
    }
    
    // Kafka
    kafkaProducer, err := kafka.NewProducer(kafka.ProducerConfig{
        Brokers: cfg.Kafka.Brokers,
        Logger:  log,
    })
    if err != nil {
        log.Fatal("Failed to create Kafka producer", zap.Error(err))
    }
    defer kafkaProducer.Close()

    // ========================================
    // Metrics & SLI
    // ========================================
    serviceMetrics := metrics.NewServiceMetrics(metrics.Config{
        ServiceName: "{service-name}",
        Namespace:   "AI",
    })
    
    sliTracker := sli.NewPrometheusTracker("{service-name}")

    // ========================================
    // Domain Services
    // ========================================
    repo := repositories.New{Entity}Repository(sqlClient, redisClient, log)
    svc := services.New{Entity}Service(repo, kafkaProducer, log)

    // ========================================
    // HTTP Server
    // ========================================
    server := api.NewServer(api.ServerConfig{
        Port:        cfg.Server.Port,
        Logger:      log,
        Metrics:     serviceMetrics,
        SliTracker:  sliTracker,
        Service:     svc,
    })

    // ========================================
    // Graceful Shutdown
    // ========================================
    go func() {
        if err := server.Start(); err != nil {
            log.Fatal("Server failed", zap.Error(err))
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Info("Shutting down...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    server.Shutdown(ctx)
    log.Info("Server stopped")
}
```

---

## Required Middleware

### correlation.go

```go
package middleware

import (
    "context"
    "net/http"

    "github.com/google/uuid"
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
)

func CorrelationID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        correlationID := r.Header.Get("X-Correlation-ID")
        if correlationID == "" {
            correlationID = uuid.New().String()
        }
        
        ctx := context.WithValue(r.Context(), logger.CorrelationIDKey, correlationID)
        w.Header().Set("X-Correlation-ID", correlationID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### error_handler.go

```go
package middleware

import (
    "encoding/json"
    "net/http"

    "github.com/your-github-org/ai-scaffolder/core/go/errors"
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
)

type ErrorResponse struct {
    Error         string `json:"error"`
    Message       string `json:"message"`
    CorrelationID string `json:"correlationId,omitempty"`
}

func ErrorHandler(log *logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    if svcErr, ok := err.(*errors.ServiceError); ok {
                        correlationID, _ := r.Context().Value(logger.CorrelationIDKey).(string)
                        
                        log.Error("Service error",
                            zap.String("code", svcErr.Code),
                            zap.String("message", svcErr.Message),
                            zap.String("correlation_id", correlationID))
                        
                        w.Header().Set("Content-Type", "application/json")
                        w.WriteHeader(svcErr.HTTPStatus())
                        json.NewEncoder(w).Encode(ErrorResponse{
                            Error:         svcErr.Code,
                            Message:       svcErr.Message,
                            CorrelationID: correlationID,
                        })
                        return
                    }
                    panic(err)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

### sli.go

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

---

## Required Endpoints

Every service MUST expose:

| Endpoint | Purpose |
|----------|---------|
| `/health/live` | Kubernetes liveness probe |
| `/health/ready` | Kubernetes readiness probe |
| `/metrics` | Prometheus scraping |
| `/api/v1/sli` | SLI metrics dashboard |

---

## Docker Compose (Local Development)

```yaml
services:
  {service-name}:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CONFIG_PATH=/app/config/config.yaml
    depends_on:
      - redis
      - kafka

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  kafka:
    image: apache/kafka:3.7.0
    ports:
      - "9092:9092"
    environment:
      - KAFKA_NODE_ID=1
      - KAFKA_PROCESS_ROLES=broker,controller
      - KAFKA_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093
      - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka:9092
      - KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER
      - KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      - KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093
      - CLUSTER_ID=MkU3OEVBNTcwNTJENDM2Qk
```

---

## Reference Implementation

Always read these pattern files before generating:
- `patterns/go/cmd/patterns/main.go`
- `patterns/go/internal/domain/services/patterns_service.go`
- `patterns/go/internal/api/handler.go`
- `core/go/logger/logger.go`
- `core/go/infrastructure/kafka/producer.go`
