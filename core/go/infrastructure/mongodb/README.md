# MongoDB Infrastructure Package

## Overview

The `mongodb` package provides a standardized interface for MongoDB operations with connection pooling, structured logging, and health checks.

## Configuration

```go
type ClientConfig struct {
    Host   string        // Required: MongoDB host address (e.g., "localhost")
    Port   int           // Required: MongoDB port (1-65535, typically 27017)
    Logger *zap.Logger   // Optional: Structured logger instance
}
```

## Usage

### Creating a Client

```go
logger, _ := zap.NewProduction()
defer logger.Sync()

cfg := mongodb.ClientConfig{
    Host:   "mongodb-cluster-1",
    Port:   27017,
    Logger: logger,
}

client, err := mongodb.NewClient(cfg)
if err != nil {
    log.Fatalf("Failed to create MongoDB client: %v", err)
}
defer client.Close(context.Background())
```

### Operations

```go
// Find documents
err := client.Find(ctx, "db.collection")

// Insert documents  
err := client.Insert(ctx, documentData)

// Health check
err := client.Health(ctx)
```

## Features

- **Connection Pooling**: Efficient connection management
- **Structured Logging**: All operations logged with timing and error context
- **Health Checks**: Replica set and connectivity verification
- **Graceful Shutdown**: Context-aware close operations
- **Configuration Validation**: Host and port validation on init

## Logging

All MongoDB operations produce structured logs with:

```json
{
  "component": "mongodb",
  "duration_ms": 123.456,
  "host": "mongodb-cluster-1",
  "port": 27017,
  "operation": "find"
}
```

## Integration with Core Services

### Health Check Registration

```go
healthChecker.Register("mongodb-client", func(ctx context.Context) error {
    return client.Health(ctx)
})
```

### Reliability Patterns

Future versions will integrate:
- Circuit breaker for connection failures
- Retry with exponential backoff
- Connection pool monitoring
- Bulkhead isolation per collection/database

## Error Codes

- `MONGODB_CONNECTION_FAILED`: Cannot establish connection
- `MONGODB_AUTHENTICATION_FAILED`: Invalid credentials
- `MONGODB_OPERATION_FAILED`: Query or insert operation failed
- `MONGODB_INVALID_CONFIG`: Configuration validation failed

## Best Practices

1. **Reuse Clients**: Create one client per connection string, share across requests
2. **Context Timeouts**: Always use context with timeouts for operations
3. **Error Handling**: Check errors from Find/Insert/Health operations
4. **Graceful Shutdown**: Always call Close() in shutdown handlers

## Testing

```bash
go test ./infrastructure/mongodb -v
```

## Dependencies

- **go.mongodb.org/mongo-driver** v1.16.0 - MongoDB official driver
- **core/logger** - Structured logging
- **core/errors** - Error registry system
