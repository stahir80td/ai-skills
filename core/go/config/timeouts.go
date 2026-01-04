package config

import (
	"os"
	"strconv"
	"time"
)

// TimeoutConfig holds all configurable timeout values for a service
type TimeoutConfig struct {
	// HTTP Server Timeouts
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HTTPIdleTimeout  time.Duration

	// HTTP Client Timeouts
	HTTPClientTimeout time.Duration

	// Context Timeouts for Request Handlers
	HTTPHandlerTimeout time.Duration

	// Database Connection Timeouts
	MongoDBPingTimeout            time.Duration
	MongoDBConnectionTimeout      time.Duration
	MongoDBServerSelectionTimeout time.Duration
	MongoDBSocketTimeout          time.Duration

	SQLConnectionTimeout time.Duration
	SQLQueryTimeout      time.Duration
	SQLPingTimeout       time.Duration

	ScyllaDBTimeout        time.Duration
	ScyllaDBConnectTimeout time.Duration

	// Redis Timeouts
	RedisDialTimeout  time.Duration
	RedisReadTimeout  time.Duration
	RedisWriteTimeout time.Duration
	RedisPingTimeout  time.Duration

	// Kafka Timeouts
	KafkaDialTimeout  time.Duration
	KafkaWriteTimeout time.Duration

	// gRPC Timeouts
	GRPCConnectionTimeout time.Duration
	GRPCRequestTimeout    time.Duration

	// WebSocket Timeouts
	WebSocketReadDeadline  time.Duration
	WebSocketWriteDeadline time.Duration
	WebSocketPingInterval  time.Duration

	// Bulkhead/Circuit Breaker Timeouts
	BulkheadTimeout time.Duration
	CircuitTimeout  time.Duration

	// Shutdown Timeouts
	ShutdownTimeout           time.Duration
	HTTPServerShutdownTimeout time.Duration

	// Cache TTLs
	CacheDefaultTTL      time.Duration
	DeviceListCacheTTL   time.Duration
	DeviceStatusCacheTTL time.Duration

	// Other Timeouts
	ValidationTimeout         time.Duration
	PublishTimeout            time.Duration
	SessionTimeout            time.Duration
	SessionCheckInterval      time.Duration
	ReconnectInterval         time.Duration
	HealthCheckTimeout        time.Duration
	HTTPClientFallbackTimeout time.Duration
}

// LoadTimeoutConfig loads timeout configuration from environment variables with sensible defaults
// All timeouts are specified in seconds via environment variables
func LoadTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		// HTTP Server Timeouts - defaults suitable for MongoDB Atlas M0
		HTTPReadTimeout:  parseDurationSeconds("HTTP_READ_TIMEOUT", 60),
		HTTPWriteTimeout: parseDurationSeconds("HTTP_WRITE_TIMEOUT", 60),
		HTTPIdleTimeout:  parseDurationSeconds("HTTP_IDLE_TIMEOUT", 120),

		// HTTP Client Timeouts
		HTTPClientTimeout: parseDurationSeconds("HTTP_CLIENT_TIMEOUT", 60),

		// Context Timeouts for Request Handlers
		HTTPHandlerTimeout: parseDurationSeconds("HTTP_HANDLER_TIMEOUT", 60),

		// MongoDB Timeouts - optimized for Atlas M0 free tier (MINIMUM 60s)
		MongoDBPingTimeout:            parseDurationSeconds("MONGODB_PING_TIMEOUT", 60),
		MongoDBConnectionTimeout:      parseDurationSeconds("MONGODB_CONNECTION_TIMEOUT", 60),
		MongoDBServerSelectionTimeout: parseDurationSeconds("MONGODB_SERVER_SELECTION_TIMEOUT", 60),
		MongoDBSocketTimeout:          parseDurationSeconds("MONGODB_SOCKET_TIMEOUT", 60),

		// SQL Timeouts (MINIMUM 60s)
		SQLConnectionTimeout: parseDurationSeconds("SQL_CONNECTION_TIMEOUT", 60),
		SQLQueryTimeout:      parseDurationSeconds("SQL_QUERY_TIMEOUT", 60),
		SQLPingTimeout:       parseDurationSeconds("SQL_PING_TIMEOUT", 60),

		// ScyllaDB Timeouts (MINIMUM 60s)
		ScyllaDBTimeout:        parseDurationSeconds("SCYLLADB_TIMEOUT", 60),
		ScyllaDBConnectTimeout: parseDurationSeconds("SCYLLADB_CONNECT_TIMEOUT", 60),

		// Redis Timeouts (MINIMUM 60s)
		RedisDialTimeout:  parseDurationSeconds("REDIS_DIAL_TIMEOUT", 60),
		RedisReadTimeout:  parseDurationSeconds("REDIS_READ_TIMEOUT", 60),
		RedisWriteTimeout: parseDurationSeconds("REDIS_WRITE_TIMEOUT", 60),
		RedisPingTimeout:  parseDurationSeconds("REDIS_PING_TIMEOUT", 60),

		// Kafka Timeouts (MINIMUM 60s)
		KafkaDialTimeout:  parseDurationSeconds("KAFKA_DIAL_TIMEOUT", 60),
		KafkaWriteTimeout: parseDurationSeconds("KAFKA_WRITE_TIMEOUT", 60),

		// gRPC Timeouts (MINIMUM 60s)
		GRPCConnectionTimeout: parseDurationSeconds("GRPC_CONNECTION_TIMEOUT", 60),
		GRPCRequestTimeout:    parseDurationSeconds("GRPC_REQUEST_TIMEOUT", 60),

		// WebSocket Timeouts (MINIMUM 60s)
		WebSocketReadDeadline:  parseDurationSeconds("WEBSOCKET_READ_DEADLINE", 60),
		WebSocketWriteDeadline: parseDurationSeconds("WEBSOCKET_WRITE_DEADLINE", 60),
		WebSocketPingInterval:  parseDurationSeconds("WEBSOCKET_PING_INTERVAL", 60),

		// Bulkhead/Circuit Breaker Timeouts (MINIMUM 60s)
		BulkheadTimeout: parseDurationSeconds("BULKHEAD_TIMEOUT", 60),
		CircuitTimeout:  parseDurationSeconds("CIRCUIT_TIMEOUT", 60),

		// Shutdown Timeouts (MINIMUM 60s)
		ShutdownTimeout:           parseDurationSeconds("SHUTDOWN_TIMEOUT", 60),
		HTTPServerShutdownTimeout: parseDurationSeconds("HTTP_SERVER_SHUTDOWN_TIMEOUT", 60),

		// Cache TTLs (MINIMUM 60s)
		CacheDefaultTTL:      parseDurationSeconds("CACHE_DEFAULT_TTL", 300),
		DeviceListCacheTTL:   parseDurationSeconds("DEVICE_LIST_CACHE_TTL", 60),
		DeviceStatusCacheTTL: parseDurationSeconds("DEVICE_STATUS_CACHE_TTL", 60),

		// Other Timeouts (MINIMUM 60s)
		ValidationTimeout:         parseDurationSeconds("VALIDATION_TIMEOUT", 60),
		PublishTimeout:            parseDurationSeconds("PUBLISH_TIMEOUT", 60),
		SessionTimeout:            parseDurationSeconds("SESSION_TIMEOUT", 300),
		SessionCheckInterval:      parseDurationSeconds("SESSION_CHECK_INTERVAL", 60),
		ReconnectInterval:         parseDurationSeconds("RECONNECT_INTERVAL", 60),
		HealthCheckTimeout:        parseDurationSeconds("HEALTH_CHECK_TIMEOUT", 60),
		HTTPClientFallbackTimeout: parseDurationSeconds("HTTP_CLIENT_FALLBACK_TIMEOUT", 60),
	}
}

// parseDurationSeconds parses a duration from environment variable (in seconds) or uses default
func parseDurationSeconds(envVar string, defaultSeconds int) time.Duration {
	if val := os.Getenv(envVar); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return time.Duration(defaultSeconds) * time.Second
}

// GetEnvInt gets an integer from environment variable or returns default
func GetEnvInt(envVar string, defaultValue int) int {
	if val := os.Getenv(envVar); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}
