# ScyllaDB Infrastructure Package

## Overview

The `scylladb` package provides a standardized interface for ScyllaDB (Cassandra-compatible) time-series database operations with built-in logging, health checks, and high-availability support.

## Configuration

```go
type SessionConfig struct {
    Hosts    []string      // Required: ScyllaDB node addresses (e.g., ["node1:9042", "node2:9042"])
    Keyspace string        // Required: Cassandra keyspace name
    Logger   *zap.Logger   // Optional: Structured logger instance
}
```

## Usage

### Creating a Session

```go
logger, _ := zap.NewProduction()
defer logger.Sync()

cfg := scylladb.SessionConfig{
    Hosts:    []string{"scylladb-1:9042", "scylladb-2:9042", "scylladb-3:9042"},
    Keyspace: "device_events",
    Logger:   logger,
}

session, err := scylladb.NewSession(cfg)
if err != nil {
    log.Fatalf("Failed to create ScyllaDB session: %v", err)
}
defer session.Close(context.Background())
```

### Time-Series Operations

```go
// Query time-series data
err := session.QueryContext(ctx, "SELECT * FROM events WHERE device_id = ? AND time >= ?", deviceID, timestamp)

// Insert metrics
err := session.ExecContext(ctx, "INSERT INTO metrics (id, time, value) VALUES (?, ?, ?)", id, time, value)

// Health check
err := session.Health(ctx)
```

## Features

- **High Availability**: Multi-node cluster support with automatic failover
- **Time-Series Optimized**: Designed for high-volume timestamp data
- **Structured Logging**: All operations logged with duration and consistency level
- **Health Checks**: Cluster connectivity and quorum verification
- **Graceful Shutdown**: Context-aware session closure
- **Configuration Validation**: Required fields validated on initialization

## Logging

All ScyllaDB operations produce structured logs:

```json
{
  "component": "scylladb",
  "duration_ms": 15.789,
  "hosts": ["scylladb-1:9042", "scylladb-2:9042"],
  "keyspace": "device_events",
  "operation": "query",
  "consistency_level": "LOCAL_QUORUM"
}
```

## Integration with Core Services

### Health Check Registration

```go
healthChecker.Register("scylladb-session", func(ctx context.Context) error {
    return session.Health(ctx)
})
```

### Metrics

Operations tracked via Prometheus metrics:
- `scylladb_query_duration_seconds` - Query execution time
- `scylladb_query_errors_total` - Failed query count by type
- `scylladb_cluster_nodes_available` - Number of available nodes
- `scylladb_batch_inserts_total` - Batch insert count

### Reliability Patterns

Future versions will integrate:
- Circuit breaker for cluster unavailability
- Automatic retry with exponential backoff
- Batch insert optimization
- Connection pooling per node
- Bulkhead isolation for concurrent operations

## Consistency Levels

Supported consistency levels for different use cases:

- `ONE` - Fastest, minimal durability (not recommended for production)
- `LOCAL_ONE` - Local datacenter single replica
- `LOCAL_QUORUM` - Local datacenter majority (recommended for high-availability)
- `QUORUM` - Cross-datacenter majority (for critical data)
- `ALL` - All replicas (highest consistency, lowest performance)

## Error Codes

- `SCYLLADB_CLUSTER_UNAVAILABLE`: Cannot connect to any cluster node
- `SCYLLADB_QUERY_FAILED`: Query execution error
- `SCYLLADB_TIMEOUT`: Query exceeded timeout
- `SCYLLADB_INSUFFICIENT_REPLICAS`: Not enough replicas for consistency level
- `SCYLLADB_INVALID_CONFIG`: Configuration validation failed

## Best Practices

1. **Session Reuse**: Maintain single session per keyspace, share across requests
2. **Batch Operations**: Use batches for inserting multiple related events
3. **TTL Management**: Set appropriate TTLs on time-series tables for retention
4. **Consistency Tuning**: Use LOCAL_QUORUM for most IoT use cases
5. **Partition Keys**: Design partition keys for even data distribution
6. **Context Timeouts**: Always provide context with reasonable timeouts
7. **Prepared Statements**: Use parameterized queries to prevent injection

## Time-Series Best Practices

```go
// Bad: String concatenation
query := fmt.Sprintf("INSERT INTO metrics (id, time, value) VALUES ('%s', %d, %f)", id, time, value)

// Good: Parameterized query
query := "INSERT INTO metrics (id, time, value) VALUES (?, ?, ?)"
session.ExecContext(ctx, query, id, time, value)
```

## Testing

```bash
go test ./infrastructure/scylladb -v
```

## Dependencies

- **github.com/gocql/gocql** v1.6.0 - Cassandra driver (compatible with ScyllaDB)
- **core/logger** - Structured logging
- **core/errors** - Error registry system
- **core/metrics** - Prometheus metric collection

## Performance Tuning

- **Batch Size**: Optimize batch insert size (typically 100-1000 rows per batch)
- **Async Writes**: Use async writes for non-critical metrics
- **Compression**: Enable network compression for large queries
- **Connection Pooling**: Configure per-host connection pool sizes
- **Replication Factor**: Set appropriate RF for data durability (RF=3 recommended)
