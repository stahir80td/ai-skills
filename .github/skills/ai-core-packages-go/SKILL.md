---
name: ai-core-packages-go
description: >
  CRITICAL enforcement of AI Core Go packages for Go development.
  Use this skill whenever generating, reviewing, or modifying Go code in AI projects.
  Ensures developers use core/go/logger, core/go/errors, core/go/infrastructure instead
  of raw packages like logrus, zerolog, go-redis, sarama, mongo-driver.
  This skill MUST be applied to all Go code generation, scaffolding, and refactoring.
---

# 🚨 AI Core Package Enforcement (Go)

## ABSOLUTE RULES - NO EXCEPTIONS

### Required go.mod Dependency

```go
require (
    github.com/your-github-org/ai-scaffolder/core/go v1.0.3
)
```

### Required Imports

```go
import (
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
    "github.com/your-github-org/ai-scaffolder/core/go/errors"
    "github.com/your-github-org/ai-scaffolder/core/go/metrics"
    "github.com/your-github-org/ai-scaffolder/core/go/sli"
    "github.com/your-github-org/ai-scaffolder/core/go/reliability"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/kafka"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/redis"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/mongodb"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/sqlserver"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/scylladb"
)
```

---

## ❌ FORBIDDEN - NEVER IMPORT THESE PACKAGES

If you see yourself importing any of these, STOP and use the Core equivalent:

| ❌ FORBIDDEN Package | ✅ Use Instead |
|---------------------|----------------|
| `log` (stdlib) | `core/go/logger` → `*logger.Logger` |
| `github.com/sirupsen/logrus` | `core/go/logger` → `*logger.Logger` |
| `github.com/rs/zerolog` | `core/go/logger` → `*logger.Logger` |
| `go.uber.org/zap` (direct) | `core/go/logger` → `*logger.Logger` |
| `github.com/go-redis/redis` | `core/go/infrastructure/redis` |
| `github.com/redis/go-redis/v9` | `core/go/infrastructure/redis` |
| `github.com/IBM/sarama` | `core/go/infrastructure/kafka` |
| `github.com/Shopify/sarama` | `core/go/infrastructure/kafka` |
| `go.mongodb.org/mongo-driver/mongo` | `core/go/infrastructure/mongodb` |
| `github.com/gocql/gocql` | `core/go/infrastructure/scylladb` |
| `database/sql` + mssql driver | `core/go/infrastructure/sqlserver` |

---

## ❌ FORBIDDEN Code Patterns

```go
// ❌ WRONG - Standard library log
import "log"
log.Printf("Processing order %s", orderID)

// ❌ WRONG - Direct logrus
import "github.com/sirupsen/logrus"
logrus.Info("message")

// ❌ WRONG - Direct zerolog
import "github.com/rs/zerolog"
log.Info().Msg("message")

// ❌ WRONG - Direct sarama Kafka
import "github.com/IBM/sarama"
producer, _ := sarama.NewSyncProducer(brokers, config)

// ❌ WRONG - Direct go-redis
import "github.com/redis/go-redis/v9"
client := redis.NewClient(&redis.Options{})

// ❌ WRONG - Direct MongoDB driver
import "go.mongodb.org/mongo-driver/mongo"
client, _ := mongo.Connect(ctx, options.Client().ApplyURI(uri))

// ❌ WRONG - Using panic for errors
if err != nil {
    panic(err)
}

// ❌ WRONG - Generic errors
return fmt.Errorf("order not found: %s", orderID)
```

---

## ✅ CORRECT Code Patterns

### Logger Setup

```go
// ✅ CORRECT - Use core/go/logger
import (
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
    "go.uber.org/zap"
)

func main() {
    // Development mode
    log, err := logger.NewDevelopment("order-service", "1.0.0")
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    // Production mode
    log, err = logger.NewProduction("order-service", "1.0.0")
    
    // Contextual logging with correlation ID
    ctx := context.WithValue(ctx, logger.CorrelationIDKey, "req-123")
    ctxLog := log.WithCorrelation(ctx)
    ctxLog.Info("Processing order", zap.String("order_id", orderID))
}
```

### Error Handling

```go
// ✅ CORRECT - Use core/go/errors
import (
    "github.com/your-github-org/ai-scaffolder/core/go/errors"
)

// Define service-specific errors
var (
    ErrOrderNotFound = errors.New("ORD-001", errors.SeverityMedium, "Order not found")
    ErrInvalidStatus = errors.New("ORD-002", errors.SeverityMedium, "Invalid order status")
    ErrPaymentFailed = errors.New("ORD-003", errors.SeverityHigh, "Payment processing failed")
)

// Usage with context
func GetOrder(ctx context.Context, orderID string) (*Order, error) {
    order, err := repo.FindByID(ctx, orderID)
    if err != nil {
        return nil, errors.Wrap(err, "ORD-001", errors.SeverityMedium, "Order not found").
            WithContext("order_id", orderID).
            WithContext("correlation_id", ctx.Value(logger.CorrelationIDKey))
    }
    return order, nil
}
```

### Infrastructure Clients

```go
// ✅ CORRECT - Infrastructure client initialization
import (
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/kafka"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/mongodb"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/redis"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/scylladb"
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/sqlserver"
)

func main() {
    // SQL Server
    sqlDB, err := sqlserver.NewClient(sqlserver.ClientConfig{
        Server:   cfg.SQLServer.Server,
        Database: cfg.SQLServer.Database,
        User:     cfg.SQLServer.User,
        Password: cfg.SQLServer.Password,
        Logger:   log,
    })
    
    // MongoDB
    mongoClient, err := mongodb.NewClient(mongodb.ClientConfig{
        ConnectionURI: cfg.MongoDB.ConnectionURI,
        Database:      cfg.MongoDB.Database,
        Logger:        log,
    })
    
    // Redis
    redisClient, err := redis.NewClient(redis.ClientConfig{
        Host:   cfg.Redis.Host,
        Port:   cfg.Redis.Port,
        Logger: log,
    })
    
    // ScyllaDB
    scyllaSession, err := scylladb.NewSession(scylladb.SessionConfig{
        Hosts:    cfg.ScyllaDB.Hosts,
        Keyspace: cfg.ScyllaDB.Keyspace,
        Logger:   log,
    })
    
    // Kafka Producer
    kafkaProducer, err := kafka.NewProducer(kafka.ProducerConfig{
        Brokers: cfg.Kafka.Brokers,
        Logger:  log,
    })
}
```

### Kafka Event Publishing

```go
// ✅ CORRECT - Use kafka.Producer from core package
import (
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/kafka"
)

type EventPublisher struct {
    producer kafka.Producer
    logger   *logger.Logger
}

func (p *EventPublisher) PublishOrderCreated(ctx context.Context, order *Order) error {
    payload, _ := json.Marshal(OrderCreatedEvent{
        OrderID:    order.ID,
        CustomerID: order.CustomerID,
        Total:      order.Total,
    })
    
    headers := map[string]string{
        "correlation_id": ctx.Value(logger.CorrelationIDKey).(string),
        "event_type":     "order.created",
    }
    
    return p.producer.SendMessage(ctx, "orders.events", order.ID, payload, headers)
}
```

---

## Pre-Generation Checklist

Before outputting ANY Go code, verify:

- [ ] go.mod has `github.com/your-github-org/ai-scaffolder/core/go` dependency
- [ ] NO forbidden imports (log, logrus, zerolog, sarama, go-redis, mongo-driver)
- [ ] Using `*logger.Logger` from core/go/logger
- [ ] Using `errors.New()` with codes, not `fmt.Errorf()`
- [ ] Using `kafka.Producer` not raw sarama
- [ ] Using `redis.Client` not go-redis
- [ ] Using `mongodb.NewClient` not mongo-driver
- [ ] Using `sqlserver.NewClient` not database/sql directly
- [ ] NO `panic()` for error handling

---

## Reference Files (Read Before Generating)

When generating Go services, read these files first:
- `patterns/go/cmd/patterns/main.go` - Infrastructure setup
- `patterns/go/internal/domain/services/patterns_service.go` - Business logic
- `core/go/logger/logger.go` - Logger interface
- `core/go/errors/errors.go` - Error patterns
- `core/go/infrastructure/kafka/producer.go` - Kafka interface
