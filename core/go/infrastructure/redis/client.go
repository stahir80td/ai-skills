package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Config for Redis
type ClientConfig struct {
	Host        string
	Port        int
	Logger      *logger.Logger
	PingTimeout time.Duration // MINIMUM 60s - NO HARDCODING
}

// Client interface for Redis operations
type Client interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}) error
	Del(ctx context.Context, keys ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SRem(ctx context.Context, key string, members ...interface{}) error
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	Expire(ctx context.Context, key string, duration time.Duration) error
	Health(ctx context.Context) error
	Close(ctx context.Context) error
}

// redisClient implements the Client interface
type redisClient struct {
	client *redis.Client
	logger *logger.Logger
}

// NewClient creates a new Redis client
func NewClient(cfg ClientConfig) (Client, error) {
	if cfg.Host == "" {
		err := fmt.Errorf("Redis host cannot be empty")
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("RedisClient").Error("Invalid configuration - missing host",
				zap.Error(err),
				zap.String("error_code", "INFRA-REDIS-CONFIG-ERROR"))
		}
		return nil, err
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		err := fmt.Errorf("Redis port must be between 1 and 65535, got %d", cfg.Port)
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("RedisClient").Error("Invalid configuration - bad port",
				zap.Error(err),
				zap.Int("port", cfg.Port),
				zap.String("error_code", "INFRA-REDIS-CONFIG-ERROR"))
		}
		return nil, err
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("RedisClient").Debug("Initiating Redis connection",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port))
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		PoolSize: 10,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout) // NO HARDCODING - from config
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("RedisClient").Error("Redis connection failed",
				zap.Error(err),
				zap.String("error_code", "INFRA-REDIS-CONNECT-ERROR"),
				zap.String("host", cfg.Host),
				zap.Int("port", cfg.Port))
		}
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("RedisClient").Info("Successfully connected to Redis",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("status", "healthy"),
			zap.Int("pool_size", 10))
	}

	return &redisClient{
		client: client,
		logger: cfg.Logger,
	}, nil
}

// Get retrieves a value from Redis
func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key doesn't exist, return empty string
	}
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis_get_failed", zap.String("key", key), zap.Error(err))
		}
		return "", err
	}
	return val, nil
}

// Set stores a value in Redis
func (r *redisClient) Set(ctx context.Context, key string, value interface{}) error {
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	default:
		// Marshal to JSON for complex types
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			if r.logger != nil {
				r.logger.Error("redis_set_marshal_failed", zap.String("key", key), zap.Error(err))
			}
			return err
		}
		strValue = string(jsonBytes)
	}

	if err := r.client.Set(ctx, key, strValue, 0).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("redis_set_failed", zap.String("key", key), zap.Error(err))
		}
		return err
	}
	return nil
}

// Del deletes keys from Redis
func (r *redisClient) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("redis_del_failed", zap.Strings("keys", keys), zap.Error(err))
		}
		return err
	}
	return nil
}

// SMembers returns all members of a set
func (r *redisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	members, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis_smembers_failed", zap.String("key", key), zap.Error(err))
		}
		return nil, err
	}
	return members, nil
}

// SAdd adds members to a set
func (r *redisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if len(members) == 0 {
		return nil
	}
	if err := r.client.SAdd(ctx, key, members...).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("redis_sadd_failed", zap.String("key", key), zap.Error(err))
		}
		return err
	}
	return nil
}

// SRem removes members from a set
func (r *redisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	if len(members) == 0 {
		return nil
	}
	if err := r.client.SRem(ctx, key, members...).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("redis_srem_failed", zap.String("key", key), zap.Error(err))
		}
		return err
	}
	return nil
}

// LRange returns a range of elements from a list
func (r *redisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	vals, err := r.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis_lrange_failed", zap.String("key", key), zap.Int64("start", start), zap.Int64("stop", stop), zap.Error(err))
		}
		return nil, err
	}
	return vals, nil
}

// Expire sets expiration time on a key
func (r *redisClient) Expire(ctx context.Context, key string, duration time.Duration) error {
	if err := r.client.Expire(ctx, key, duration).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("redis_expire_failed", zap.String("key", key), zap.Error(err))
		}
		return err
	}
	return nil
}

// Health checks if Redis is healthy
func (r *redisClient) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *redisClient) Close(ctx context.Context) error {
	return r.client.Close()
}
