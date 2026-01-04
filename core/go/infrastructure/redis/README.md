# Redis Infrastructure Package

## Overview

The `redis` package provides a standardized interface for Redis cache operations with built-in logging, health checks, and connection management.

## Configuration

```go
type ClientConfig struct {
    Host   string        // Required: Redis host address (e.g., "localhost")
    Port   int           // Required: Redis port (1-65535, typically 6379)
    Logger *zap.Logger   // Optional: Structured logger instance
}
```

## Usage

### Creating a Client

```go
logger, _ := zap.NewProduction()
defer logger.Sync()

cfg := redis.ClientConfig{
    Host:   "redis-cluster-1",
    Port:   6379,
    Logger: logger,
}

client, err := redis.NewClient(cfg)
if err != nil {
    log.Fatalf("Failed to create Redis client: %v", err)
}
defer client.Close(context.Background())
```

### Cache Operations

```go
// Get value
value, err := client.Get(ctx, "cache-key")

// Set value
err := client.Set(ctx, "cache-key", "cache-value")

// Health check
err := client.Health(ctx)
```

## Features

- **Key-Value Operations**: Get and Set with automatic serialization
- **Structured Logging**: All operations logged with duration and component tags
- **Health Checks**: Redis connectivity and response time verification
- **Graceful Shutdown**: Context-aware close operations
- **Configuration Validation**: Host and port validation on initialization

## Logging

All Redis operations produce structured logs:

```json
{
  "component": "redis",
  "duration_ms": 5.234,
  "host": "redis-cluster-1",
  "port": 6379,
  "operation": "get",
  "key": "cache-key"
}
```

## Integration with Core Services

### Health Check Registration

```go
healthChecker.Register("redis-client", func(ctx context.Context) error {
    return client.Health(ctx)
})
```

### Metrics

Operations tracked via Prometheus metrics:
- `redis_operation_duration_seconds` - Operation timing
- `redis_operation_errors_total` - Error count by operation
- `redis_cache_hit_ratio` - Cache hit/miss ratio

### Reliability Patterns

Future versions will integrate:
- Circuit breaker for unavailable Redis instances
- Automatic reconnection with exponential backoff
- Connection pool management
- Cache-aside pattern support

## Error Codes

- `REDIS_CONNECTION_FAILED`: Cannot connect to Redis
- `REDIS_OPERATION_TIMEOUT`: Operation exceeded timeout
- `REDIS_COMMAND_FAILED`: Redis command execution failed
- `REDIS_INVALID_CONFIG`: Configuration validation failed

## Best Practices

1. **Connection Reuse**: Maintain single client connection per Redis instance
2. **Timeout Contexts**: Always provide context with reasonable timeouts
3. **Error Handling**: Handle both connection and command errors gracefully
4. **Key Naming**: Use consistent key namespace (e.g., "app:feature:key")
5. **TTL Management**: Set appropriate TTLs to prevent unbounded cache growth

## Testing

```bash
go test ./infrastructure/redis -v
```

## Dependencies

- **github.com/redis/go-redis/v9** v9.7.0 - Redis client library
- **core/logger** - Structured logging
- **core/errors** - Error registry system
- **core/metrics** - Prometheus metric collection
