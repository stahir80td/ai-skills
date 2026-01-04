# SQL Server Infrastructure Package

## Overview

The `sqlserver` package provides a standardized interface for SQL Server database operations with connection management, structured logging, and health checks.

## Configuration

```go
type ClientConfig struct {
    Server   string        // Required: SQL Server address (e.g., "localhost")
    Database string        // Required: Database name
    Logger   *zap.Logger   // Optional: Structured logger instance
}
```

## Usage

### Creating a Client

```go
logger, _ := zap.NewProduction()
defer logger.Sync()

cfg := sqlserver.ClientConfig{
    Server:   "sqlserver-1.example.com",
    Database: "your-org",
    Logger:   logger,
}

client, err := sqlserver.NewClient(cfg)
if err != nil {
    log.Fatalf("Failed to create SQL Server client: %v", err)
}
defer client.Close(context.Background())
```

### Database Operations

```go
// Execute query
err := client.QueryContext(ctx, "SELECT * FROM Users WHERE ID = @id", userID)

// Execute command
err := client.ExecContext(ctx, "INSERT INTO Events VALUES (@data)", eventData)

// Health check
err := client.Health(ctx)
```

## Features

- **Connection Pooling**: Efficient connection management with configurable limits
- **Structured Logging**: All operations logged with duration, query, and error context
- **Health Checks**: Server connectivity and query response time verification
- **Graceful Shutdown**: Context-aware close operations
- **Configuration Validation**: Required fields validated on initialization
- **Parameterized Queries**: Built-in support for query parameters

## Logging

All SQL Server operations produce structured logs:

```json
{
  "component": "sqlserver",
  "duration_ms": 234.567,
  "server": "sqlserver-1.example.com",
  "database": "your-org",
  "operation": "query",
  "query_type": "SELECT"
}
```

## Integration with Core Services

### Health Check Registration

```go
healthChecker.Register("sqlserver-client", func(ctx context.Context) error {
    return client.Health(ctx)
})
```

### Metrics

Operations tracked via Prometheus metrics:
- `sqlserver_query_duration_seconds` - Query execution time
- `sqlserver_query_errors_total` - Failed query count
- `sqlserver_connection_pool_size` - Active connection count

### Reliability Patterns

Future versions will integrate:
- Circuit breaker for connection failures
- Automatic retry with exponential backoff
- Connection pool monitoring and optimization
- Query timeout enforcement

## Error Codes

- `SQLSERVER_CONNECTION_FAILED`: Cannot establish connection
- `SQLSERVER_AUTHENTICATION_FAILED`: Invalid credentials
- `SQLSERVER_QUERY_FAILED`: Query execution error
- `SQLSERVER_TIMEOUT`: Query exceeded timeout
- `SQLSERVER_INVALID_CONFIG`: Configuration validation failed

## Best Practices

1. **Connection Pooling**: Reuse single client instance across application
2. **Context Timeouts**: Always provide context with appropriate timeouts
3. **Parameterized Queries**: Use @parameter placeholders to prevent SQL injection
4. **Error Handling**: Log and handle both connection and query errors
5. **Transaction Management**: Use connection.BeginTx() for multi-step operations
6. **Query Optimization**: Monitor slow queries via structured logs

## Testing

```bash
go test ./infrastructure/sqlserver -v
```

## Dependencies

- **github.com/microsoft/go-mssqldb** v1.7.2 - SQL Server driver
- **core/logger** - Structured logging
- **core/errors** - Error registry system
- **core/metrics** - Prometheus metric collection

## Connection String Format

```
Server=<server>;Database=<db>;User Id=<user>;Password=<password>
```

Example:
```
Server=sqlserver-1.example.com;Database=your-org;User Id=sa;Password=YourPassword
```
