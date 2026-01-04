package mongodb

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// RetryConfig defines retry behavior for MongoDB operations
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Logger       *logger.Logger
}

// DefaultRetryConfig for MongoDB Atlas M0 free tier
func DefaultRetryConfig(log *logger.Logger) RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Logger:       log,
	}
}

// IsRetryableError checks if a MongoDB error should be retried
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Retryable errors for Atlas M0
	retryableErrors := []string{
		"server selection error",
		"ReplicaSetNoPrimary",
		"no reachable servers",
		"connection refused",
		"i/o timeout",
		"connection reset by peer",
		"topology is closed",
		"PoolClosedError",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errMsg, retryable) {
			return true
		}
	}

	// Check for specific MongoDB driver errors
	if errors.Is(err, mongo.ErrClientDisconnected) {
		return true
	}

	return false
}

// WithRetry wraps a MongoDB operation with exponential backoff retry logic
// Optimized for MongoDB Atlas M0 free tier transient failures
func WithRetry(ctx context.Context, cfg RetryConfig, operation func() error) error {
	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		err := operation()

		if err == nil {
			// Success
			if attempt > 1 && cfg.Logger != nil {
				cfg.Logger.WithComponent("MongoDBRetry").Info("Operation succeeded after retry",
					zap.Int("attempt", attempt),
					zap.Int("total_attempts", cfg.MaxAttempts))
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryableError(err) {
			if cfg.Logger != nil {
				cfg.Logger.WithComponent("MongoDBRetry").Debug("Error is not retryable",
					zap.Error(err),
					zap.Int("attempt", attempt))
			}
			return err
		}

		// Last attempt - don't sleep
		if attempt == cfg.MaxAttempts {
			if cfg.Logger != nil {
				cfg.Logger.WithComponent("MongoDBRetry").Error("All retry attempts exhausted",
					zap.Error(err),
					zap.Int("attempts", cfg.MaxAttempts),
					zap.String("error_code", "INFRA-MONGODB-RETRY-EXHAUSTED"))
			}
			break
		}

		// Log retry attempt
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("MongoDBRetry").Warn("Retryable error detected, waiting before retry",
				zap.Error(err),
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", cfg.MaxAttempts),
				zap.Duration("delay", delay),
				zap.String("error_type", "retryable"))
		}

		// Wait with exponential backoff
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Double the delay for next attempt (exponential backoff)
			delay *= 2
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
		}
	}

	return lastErr
}

// WithRetryFunc wraps a MongoDB operation that returns a value with retry logic
func WithRetryFunc[T any](ctx context.Context, cfg RetryConfig, operation func() (T, error)) (T, error) {
	var result T
	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		res, err := operation()

		if err == nil {
			if attempt > 1 && cfg.Logger != nil {
				cfg.Logger.WithComponent("MongoDBRetry").Info("Operation succeeded after retry",
					zap.Int("attempt", attempt))
			}
			return res, nil
		}

		lastErr = err

		if !IsRetryableError(err) {
			return result, err
		}

		if attempt == cfg.MaxAttempts {
			if cfg.Logger != nil {
				cfg.Logger.WithComponent("MongoDBRetry").Error("All retry attempts exhausted",
					zap.Error(err),
					zap.Int("attempts", cfg.MaxAttempts))
			}
			break
		}

		if cfg.Logger != nil {
			cfg.Logger.WithComponent("MongoDBRetry").Warn("Retrying operation",
				zap.Error(err),
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay))
		}

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(delay):
			delay *= 2
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
		}
	}

	return result, lastErr
}
