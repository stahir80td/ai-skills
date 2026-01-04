package reliability

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	Multiplier      float64
	Jitter          bool
	RetryableErrors []error
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

var (
	retryAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "retry_attempts_total",
			Help: "Total retry attempts",
		},
		[]string{"name", "attempt"},
	)

	retrySuccess = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "retry_success_total",
			Help: "Total successful retries",
		},
		[]string{"name", "attempt"},
	)

	retryFailure = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "retry_failure_total",
			Help: "Total failed retries (after all attempts)",
		},
		[]string{"name"},
	)

	retryDelay = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "retry_delay_seconds",
			Help:    "Delay between retry attempts",
			Buckets: []float64{.001, .01, .1, .5, 1, 2, 5, 10, 30},
		},
		[]string{"name"},
	)
)

// Retry executes function with exponential backoff retry logic
func Retry(ctx context.Context, name string, config RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Track attempt
		retryAttempts.WithLabelValues(name, attemptLabel(attempt)).Inc()

		// Execute function
		err := fn()
		if err == nil {
			retrySuccess.WithLabelValues(name, attemptLabel(attempt)).Inc()
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err, config.RetryableErrors) {
			retryFailure.WithLabelValues(name).Inc()
			return err
		}

		// Don't sleep after last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate backoff delay
		delay := calculateBackoff(attempt, config)
		retryDelay.WithLabelValues(name).Observe(delay.Seconds())

		// Wait with context support
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			retryFailure.WithLabelValues(name).Inc()
			return ctx.Err()
		}
	}

	retryFailure.WithLabelValues(name).Inc()
	return lastErr
}

// RetryWithResult executes function that returns a result with retry logic
func RetryWithResult[T any](ctx context.Context, name string, config RetryConfig, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		retryAttempts.WithLabelValues(name, attemptLabel(attempt)).Inc()

		res, err := fn()
		if err == nil {
			retrySuccess.WithLabelValues(name, attemptLabel(attempt)).Inc()
			return res, nil
		}

		lastErr = err

		if !isRetryable(err, config.RetryableErrors) {
			retryFailure.WithLabelValues(name).Inc()
			return result, err
		}

		if attempt == config.MaxAttempts-1 {
			break
		}

		delay := calculateBackoff(attempt, config)
		retryDelay.WithLabelValues(name).Observe(delay.Seconds())

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			retryFailure.WithLabelValues(name).Inc()
			return result, ctx.Err()
		}
	}

	retryFailure.WithLabelValues(name).Inc()
	return result, lastErr
}

func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	// Exponential backoff: initialDelay * (multiplier ^ attempt)
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Add jitter to prevent thundering herd
	if config.Jitter {
		jitter := rand.Float64() * delay * 0.3 // Up to 30% jitter
		delay = delay + jitter
	}

	return time.Duration(delay)
}

func isRetryable(err error, retryableErrors []error) bool {
	if len(retryableErrors) == 0 {
		// If no specific errors defined, retry on any error
		return true
	}

	for _, retryableErr := range retryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}
	return false
}

func attemptLabel(attempt int) string {
	switch attempt {
	case 0:
		return "1"
	case 1:
		return "2"
	case 2:
		return "3"
	default:
		return "4+"
	}
}
