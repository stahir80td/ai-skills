# Core Infrastructure Package

## Overview

The `infrastructure` package provides production-grade, standardized interfaces for all database and message broker integrations used across the AI Scaffolder platform. It eliminates code duplication, enforces consistent logging/error handling, and provides reliable patterns for managing external service dependencies.

## Architecture

```
core/infrastructure/
├── health/          # Centralized health check framework
├── kafka/           # Apache Kafka producer/consumer interfaces
├── keyvault/        # Azure KeyVault Emulator with Redis cache-aside
├── mongodb/         # MongoDB client with connection pooling
├── redis/           # Redis cache client
├── sqlserver/       # SQL Server database client
└── scylladb/        # ScyllaDB time-series database client
```

## Component Overview

| Package | Purpose | Use Cases |
|---------|---------|-----------|
| **health** | Centralized health checks | Liveness/readiness probes, startup validation, monitoring |
| **kafka** | Event streaming | Device events, system events, async processing |
| **keyvault** | Secret management | User integration keys, API tokens, OAuth credentials |
| **mongodb** | Document storage | User profiles, device metadata, configurations |
| **redis** | In-memory cache | Session caching, device state, real-time metrics |
| **sqlserver** | Relational data | Structured records, transactions, complex queries |
| **scylladb** | Time-series data | Device metrics, historical events, analytics |

## Key Features

### 1. Structured Logging

Every operation produces detailed structured logs with:
- Component identification
- Operation duration
- Error context
- Resource identifiers
- Correlation IDs (via core/logger integration)

```go
logger.Info("operation_completed",
    zap.String("component", "kafka"),
    zap.Duration("duration", elapsed),
    zap.String("topic", topic),
    zap.Int64("offset", offset),
)
```

### 2. Unified Health Checks

All packages implement consistent Health() method:

```go
err := producer.Health(context.Background())
err := mongoClient.Health(context.Background())
err := redisClient.Health(context.Background())
```

### 3. Error Handling

Errors integrated with core/errors registry for:
- Consistent error codes
- Error classification
- SLI/SLO tracking
- Incident management

### 4. Configuration Validation

Every client validates configuration on initialization:

```go
// Invalid - caught immediately
cfg := kafka.ProducerConfig{Brokers: []string{}}
producer, err := kafka.NewProducer(cfg) // Error: brokers list empty
```

### 5. Context Support

All operations accept context.Context for:
- Timeout enforcement
- Cancellation propagation
- Correlation ID tracking

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := producer.SendMessage(ctx, key, value)
```

## Integration Guide

### Step 1: Initialize Health Checker

```go
import (
    "github.com/your-org/core/infrastructure/health"
    "github.com/your-org/core/logger"
)

// Create logger
log := logger.NewProduction()
defer log.Sync()

// Create health checker
healthChecker := health.NewChecker(log)
healthChecker.SetTimeout(5 * time.Second)
```

### Step 2: Initialize Infrastructure Clients

```go
import (
    "github.com/your-org/core/infrastructure/kafka"
    "github.com/your-org/core/infrastructure/mongodb"
    "github.com/your-org/core/infrastructure/sqlserver"
    "github.com/your-org/core/infrastructure/scylladb"
)

// Initialize Kafka producer
kafkaProducer, err := kafka.NewProducer(kafka.ProducerConfig{
    Brokers: []string{"kafka-1:9092", "kafka-2:9092"},
    Topic:   "events",
    Logger:  log,
})
if err != nil {
    log.Fatal("kafka_init_failed", zap.Error(err))
}
defer kafkaProducer.Close(context.Background())

// Initialize MongoDB client
mongoClient, err := mongodb.NewClient(mongodb.ClientConfig{
    Host:   "mongodb-cluster",
    Port:   27017,
    Logger: log,
})
if err != nil {
    log.Fatal("mongodb_init_failed", zap.Error(err))
}
defer mongoClient.Close(context.Background())

// Initialize Redis client
redisClient, err := redis.NewClient(redis.ClientConfig{
    Host:   "redis-1",
    Port:   6379,
    Logger: log,
})
if err != nil {
    log.Fatal("redis_init_failed", zap.Error(err))
}
defer redisClient.Close(context.Background())

// Initialize SQL Server client
sqlClient, err := sqlserver.NewClient(sqlserver.ClientConfig{
    Server:   "sqlserver.example.com",
    Database: "your-org",
    Logger:   log,
})
if err != nil {
    log.Fatal("sqlserver_init_failed", zap.Error(err))
}
defer sqlClient.Close(context.Background())

// Initialize ScyllaDB session
scyllaSession, err := scylladb.NewSession(scylladb.SessionConfig{
    Hosts:    []string{"scylladb-1:9042", "scylladb-2:9042", "scylladb-3:9042"},
    Keyspace: "device_events",
    Logger:   log,
})
if err != nil {
    log.Fatal("scylladb_init_failed", zap.Error(err))
}
defer scyllaSession.Close(context.Background())
```

### Step 3: Register Health Checks

```go
healthChecker.Register("kafka-producer", func(ctx context.Context) error {
    return kafkaProducer.Health(ctx)
})

healthChecker.Register("mongodb-client", func(ctx context.Context) error {
    return mongoClient.Health(ctx)
})

healthChecker.Register("redis-cache", func(ctx context.Context) error {
    return redisClient.Health(ctx)
})

healthChecker.Register("sqlserver-db", func(ctx context.Context) error {
    return sqlClient.Health(ctx)
})

healthChecker.Register("scylladb-timeseries", func(ctx context.Context) error {
    return scyllaSession.Health(ctx)
})
```

### Step 4: Verify Startup

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if !healthChecker.IsHealthy(ctx) {
    results := healthChecker.Check(ctx)
    for name, result := range results {
        if result.Status != health.StatusHealthy {
            log.Fatal("startup_failed",
                zap.String("component", name),
                zap.String("error", result.Error),
            )
        }
    }
}
log.Info("all_infrastructure_services_healthy")
```

### Step 5: Expose Health Endpoint

```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    results := healthChecker.Check(r.Context())
    
    if !healthChecker.IsHealthy(r.Context()) {
        w.WriteHeader(http.StatusServiceUnavailable)
    } else {
        w.WriteHeader(http.StatusOK)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
})

http.ListenAndServe(":8080", nil)
```

## Best Practices

### 1. Lifecycle Management

```go
// ✓ Good: Single shared instance
producer, _ := kafka.NewProducer(cfg)
defer producer.Close(ctx)
// Use producer throughout application lifetime

// ✗ Bad: Creating new instance per message
for _, msg := range messages {
    p, _ := kafka.NewProducer(cfg)
    p.SendMessage(ctx, key, value)
    p.Close(ctx)
}
```

### 2. Error Handling

```go
// ✓ Good: Specific error handling
err := producer.SendMessage(ctx, key, value)
if err != nil {
    logger.Error("failed_to_send_event",
        zap.Error(err),
        zap.String("component", "kafka"),
        zap.Bytes("key", key),
    )
    // Handle error appropriately
}

// ✗ Bad: Ignoring errors
_ = producer.SendMessage(ctx, key, value)
```

### 3. Context Timeouts

```go
// ✓ Good: Provide appropriate timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := mongoClient.Find(ctx, query)

// ✗ Bad: No timeout
err := mongoClient.Find(context.Background(), query)
```

### 4. Health Monitoring

```go
// ✓ Good: Regular health checks
ticker := time.NewTicker(10 * time.Second)
defer ticker.Stop()

for range ticker.C {
    results := healthChecker.Check(context.Background())
    for name, result := range results {
        if result.Status != health.StatusHealthy {
            // Trigger alerting, restart, etc.
        }
    }
}

// ✗ Bad: Only checking on startup
healthChecker.Check(ctx) // Once at startup, never again
```

### 5. Resource Cleanup

```go
// ✓ Good: Ensure cleanup on shutdown
func shutdown(ctx context.Context) error {
    shutdownTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    var errs []error
    
    if err := kafkaProducer.Close(shutdownTimeout); err != nil {
        errs = append(errs, err)
    }
    if err := mongoClient.Close(shutdownTimeout); err != nil {
        errs = append(errs, err)
    }
    if err := redisClient.Close(shutdownTimeout); err != nil {
        errs = append(errs, err)
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("shutdown_errors: %v", errs)
    }
    return nil
}
```

## Configuration Management

Recommended configuration approach using environment variables:

```go
func loadConfig() (*Config, error) {
    return &Config{
        Kafka: kafka.ProducerConfig{
            Brokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
            Topic:   os.Getenv("KAFKA_TOPIC"),
        },
        MongoDB: mongodb.ClientConfig{
            Host: os.Getenv("MONGODB_HOST"),
            Port: mustParseInt(os.Getenv("MONGODB_PORT")),
        },
        Redis: redis.ClientConfig{
            Host: os.Getenv("REDIS_HOST"),
            Port: mustParseInt(os.Getenv("REDIS_PORT")),
        },
        // ... other services
    }
}
```

## Monitoring and Observability

### Prometheus Metrics

Each infrastructure package exports metrics via core/metrics:

```
kafka_producer_messages_total{topic="events"}
mongodb_query_duration_seconds{operation="find"}
redis_operation_duration_seconds{command="get"}
sqlserver_query_duration_seconds{query_type="SELECT"}
scylladb_query_duration_seconds{consistency_level="LOCAL_QUORUM"}
health_check_duration_seconds{check_name="kafka-producer"}
```

### Structured Logging

All operations produce structured logs for:
- ELK stack ingestion
- CloudWatch/DataDog integration
- Alerting and anomaly detection

### SLI/SLO Integration

Errors automatically tracked via core/sli for:
- Error rate SLIs
- Latency SLIs
- Availability tracking

## Testing

Run all infrastructure tests:

```bash
go test ./infrastructure/... -v
```

Run specific package tests:

```bash
go test ./infrastructure/health -v
go test ./infrastructure/kafka -v
go test ./infrastructure/mongodb -v
go test ./infrastructure/redis -v
go test ./infrastructure/sqlserver -v
go test ./infrastructure/scylladb -v
```

## Future Enhancements

- Connection pool monitoring and auto-tuning
- Automatic circuit breaker integration
- Batch operation optimizations
- Multi-region failover support
- Performance profiling and optimization
- Advanced caching strategies

## Dependencies

- **IBM/sarama** v1.46.3 - Kafka client
- **go.mongodb.org/mongo-driver** v1.16.0 - MongoDB driver
- **github.com/redis/go-redis/v9** v9.7.0 - Redis driver
- **github.com/microsoft/go-mssqldb** v1.7.2 - SQL Server driver
- **github.com/gocql/gocql** v1.6.0 - ScyllaDB/Cassandra driver
- **core/logger** - Structured logging
- **core/errors** - Error registry
- **core/metrics** - Prometheus metrics
- **core/reliability** - Circuit breaker, retry patterns (future)

## Support

For issues or questions:
1. Check individual package README files
2. Review integration examples in this file
3. Check core/errors for error codes
4. Consult core/logger for logging patterns
