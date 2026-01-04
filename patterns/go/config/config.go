package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the patterns service
type Config struct {
	Service   ServiceConfig   `yaml:"service"`
	Logging   LoggingConfig   `yaml:"logging"`
	SQLServer SQLServerConfig `yaml:"sqlserver"`
	MongoDB   MongoDBConfig   `yaml:"mongodb"`
	ScyllaDB  ScyllaDBConfig  `yaml:"scylladb"`
	Redis     RedisConfig     `yaml:"redis"`
	Kafka     KafkaConfig     `yaml:"kafka"`
	SLI       SLIConfig       `yaml:"sli"`
}

// ServiceConfig holds service-level configuration
type ServiceConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Port        int    `yaml:"port"`
	Environment string `yaml:"environment"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level            string `yaml:"level"`
	EnableCaller     bool   `yaml:"enable_caller"`
	EnableStacktrace bool   `yaml:"enable_stacktrace"`
}

// SQLServerConfig holds SQL Server connection configuration
type SQLServerConfig struct {
	Server      string        `yaml:"server"`
	Database    string        `yaml:"database"`
	User        string        `yaml:"user"`
	Password    string        `yaml:"password"`
	PingTimeout time.Duration `yaml:"ping_timeout"`
}

// MongoDBConfig holds MongoDB connection configuration
type MongoDBConfig struct {
	ConnectionURI string        `yaml:"connection_uri"`
	Database      string        `yaml:"database"`
	PingTimeout   time.Duration `yaml:"ping_timeout"`
}

// ScyllaDBConfig holds ScyllaDB connection configuration
type ScyllaDBConfig struct {
	Hosts          []string      `yaml:"hosts"`
	Keyspace       string        `yaml:"keyspace"`
	Timeout        time.Duration `yaml:"timeout"`
	ConnectTimeout time.Duration `yaml:"connect_timeout"`
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	PingTimeout time.Duration `yaml:"ping_timeout"`
}

// KafkaConfig holds Kafka connection configuration
type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
}

// SLIConfig holds SLI/error budget configuration
type SLIConfig struct {
	AvailabilityTarget     float64 `yaml:"availability_target"`
	LatencyP95TargetMs     int     `yaml:"latency_p95_target_ms"`
	LatencyP99TargetMs     int     `yaml:"latency_p99_target_ms"`
	ErrorRateTargetPercent float64 `yaml:"error_rate_target_percent"`
}

// Load reads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	applyDefaults(&cfg)

	return &cfg, nil
}

// LoadFromEnv creates configuration from environment variables
func LoadFromEnv() *Config {
	cfg := &Config{
		Service: ServiceConfig{
			Name:        getEnv("SERVICE_NAME", "ai-patterns"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Port:        getEnvInt("SERVICE_PORT", 8080),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Logging: LoggingConfig{
			Level:            getEnv("LOG_LEVEL", "info"),
			EnableCaller:     getEnvBool("LOG_ENABLE_CALLER", true),
			EnableStacktrace: getEnvBool("LOG_ENABLE_STACKTRACE", true),
		},
		SQLServer: SQLServerConfig{
			Server:      getEnv("SQLSERVER_SERVER", "localhost:1433"),
			Database:    getEnv("SQLSERVER_DATABASE", "AiPatterns"),
			User:        getEnv("SQLSERVER_USER", "sa"),
			Password:    getEnv("SQLSERVER_PASSWORD", "AiPatterns2024!"),
			PingTimeout: getEnvDuration("SQLSERVER_PING_TIMEOUT", 60*time.Second),
		},
		MongoDB: MongoDBConfig{
			ConnectionURI: getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database:      getEnv("MONGODB_DATABASE", "AiPatternsDB"),
			PingTimeout:   getEnvDuration("MONGODB_PING_TIMEOUT", 60*time.Second),
		},
		ScyllaDB: ScyllaDBConfig{
			Hosts:          getEnvSlice("SCYLLADB_HOSTS", []string{"localhost"}),
			Keyspace:       getEnv("SCYLLADB_KEYSPACE", "ai_patterns"),
			Timeout:        getEnvDuration("SCYLLADB_TIMEOUT", 60*time.Second),
			ConnectTimeout: getEnvDuration("SCYLLADB_CONNECT_TIMEOUT", 60*time.Second),
		},
		Redis: RedisConfig{
			Host:        getEnv("REDIS_HOST", "localhost"),
			Port:        getEnvInt("REDIS_PORT", 6379),
			PingTimeout: getEnvDuration("REDIS_PING_TIMEOUT", 60*time.Second),
		},
		Kafka: KafkaConfig{
			Brokers: getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		},
		SLI: SLIConfig{
			AvailabilityTarget:     getEnvFloat("SLI_AVAILABILITY_TARGET", 99.9),
			LatencyP95TargetMs:     getEnvInt("SLI_LATENCY_P95_TARGET_MS", 200),
			LatencyP99TargetMs:     getEnvInt("SLI_LATENCY_P99_TARGET_MS", 500),
			ErrorRateTargetPercent: getEnvFloat("SLI_ERROR_RATE_TARGET", 0.1),
		},
	}

	return cfg
}

// applyDefaults sets default values for missing configuration
func applyDefaults(cfg *Config) {
	if cfg.Service.Name == "" {
		cfg.Service.Name = "ai-patterns"
	}
	if cfg.Service.Version == "" {
		cfg.Service.Version = "1.0.0"
	}
	if cfg.Service.Port == 0 {
		cfg.Service.Port = 8080
	}
	if cfg.Service.Environment == "" {
		cfg.Service.Environment = "development"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.SQLServer.PingTimeout == 0 {
		cfg.SQLServer.PingTimeout = 60 * time.Second
	}
	if cfg.MongoDB.PingTimeout == 0 {
		cfg.MongoDB.PingTimeout = 60 * time.Second
	}
	if cfg.ScyllaDB.Timeout == 0 {
		cfg.ScyllaDB.Timeout = 60 * time.Second
	}
	if cfg.ScyllaDB.ConnectTimeout == 0 {
		cfg.ScyllaDB.ConnectTimeout = 60 * time.Second
	}
	if cfg.Redis.PingTimeout == 0 {
		cfg.Redis.PingTimeout = 60 * time.Second
	}
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		var result float64
		fmt.Sscanf(value, "%f", &result)
		return result
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		return []string{value} // Simplified - in production, properly parse comma-separated values
	}
	return defaultValue
}
