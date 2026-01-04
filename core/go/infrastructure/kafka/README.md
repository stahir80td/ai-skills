# Kafka Infrastructure Package

## Overview

The `kafka` package provides a standardized interface for Kafka producer operations with built-in logging, health checks, and reliability patterns.

## Configuration

```go
type ProducerConfig struct {
    Brokers []string      // Required: Kafka broker addresses (e.g., ["localhost:9092"])
    Topic   string        // Optional: Default topic for messages
    Logger  *zap.Logger   // Optional: Structured logger instance
}
```

## Usage

### Creating a Producer

```go
logger, _ := zap.NewProduction()
defer logger.Sync()

cfg := kafka.ProducerConfig{
    Brokers: []string{"kafka-broker-1:9092", "kafka-broker-2:9092"},
    Topic:   "events",
    Logger:  logger,
}

producer, err := kafka.NewProducer(cfg)
if err != nil {
    log.Fatalf("Failed to create Kafka producer: %v", err)
}
defer producer.Close(context.Background())
```

### Sending Messages

```go
// Asynchronous send
err := producer.SendMessage(ctx, []byte("key"), []byte("message data"))

// Or handle with error checking
if err != nil {
    logger.Error("failed_to_send_message",
        zap.Error(err),
        zap.String("component", "kafka"),
    )
}
```

### Health Checks

```go
err := producer.Health(ctx)
if err != nil {
    logger.Warn("kafka_health_check_failed", zap.Error(err))
}
```

## Features

- **Structured Logging**: All operations logged with component tag and duration tracking
- **Error Handling**: Errors mapped through core/errors registry
- **Health Checks**: Kafka broker connectivity verification
- **Graceful Shutdown**: Context-aware close operations
- **Configuration Validation**: Required fields validated on initialization

## Logging

All Kafka operations produce structured logs with the following pattern:

```json
{
  "component": "kafka",
  "duration_ms": 45.123,
  "brokers": ["kafka-1:9092", "kafka-2:9092"],
  "error": null
}
```

## Integration with Core Services

### Health Check Registration

```go
healthChecker.Register("kafka-producer", func(ctx context.Context) error {
    return producer.Health(ctx)
})
```

### Reliability Patterns

Future versions will integrate:
- Circuit breaker for broker failures
- Retry logic with exponential backoff
- Rate limiting for high-volume producers
- Bulkhead isolation for resource management

## Error Codes

- `KAFKA_BROKERS_UNAVAILABLE`: Cannot connect to any broker
- `KAFKA_SEND_FAILED`: Message send operation failed
- `KAFKA_INVALID_CONFIG`: Configuration validation failed

## Testing

Run tests with:

```bash
go test ./infrastructure/kafka -v
```

## Dependencies

- **IBM/sarama** v1.46.3 - Kafka client library
- **core/logger** - Structured logging
- **core/errors** - Error registry system
