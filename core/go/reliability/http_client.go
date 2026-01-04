package reliability

import (
	"context"
	"net/http"
	"time"
)

// HTTPClient wraps http.Client with reliability patterns
type HTTPClient struct {
	client         *http.Client
	circuitBreaker *CircuitBreaker
	rateLimiter    *RateLimiter
	bulkhead       *Bulkhead
	retryConfig    RetryConfig
	name           string
}

// HTTPClientConfig configures the reliable HTTP client
type HTTPClientConfig struct {
	Name               string
	BaseTimeout        time.Duration
	CircuitMaxFailures uint32
	CircuitTimeout     time.Duration
	RateLimit          float64 // requests per second
	RateBurst          int
	MaxConcurrency     int
	BulkheadTimeout    time.Duration
	EnableRetry        bool
	RetryConfig        RetryConfig
}

// DefaultHTTPClientConfig returns sensible defaults for HTTP client
func DefaultHTTPClientConfig(name string) HTTPClientConfig {
	return HTTPClientConfig{
		Name:               name,
		BaseTimeout:        30 * time.Second,
		CircuitMaxFailures: 5,
		CircuitTimeout:     60 * time.Second,
		RateLimit:          100, // 100 req/s
		RateBurst:          200, // burst of 200 for traffic spikes
		MaxConcurrency:     50,  // max 50 concurrent requests
		BulkheadTimeout:    5 * time.Second,
		EnableRetry:        true,
		RetryConfig:        DefaultRetryConfig(),
	}
}

// NewHTTPClient creates a new reliable HTTP client
func NewHTTPClient(config HTTPClientConfig) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: config.BaseTimeout,
		},
		circuitBreaker: NewCircuitBreaker(
			config.Name+"-circuit",
			config.CircuitMaxFailures,
			config.CircuitTimeout,
		),
		rateLimiter: NewRateLimiter(
			config.Name+"-rate",
			config.RateLimit,
			config.RateBurst,
		),
		bulkhead: NewBulkhead(
			config.Name+"-bulkhead",
			config.MaxConcurrency,
			config.BulkheadTimeout,
		),
		retryConfig: config.RetryConfig,
		name:        config.Name,
	}
}

// Do executes HTTP request with all reliability patterns
func (hc *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// Rate limiting
	if !hc.rateLimiter.Allow() {
		return nil, &ReliabilityError{
			Type:    "rate_limit",
			Message: "request rate limit exceeded",
		}
	}

	var resp *http.Response
	var err error

	// Bulkhead + Circuit Breaker + Retry
	bulkheadErr := hc.bulkhead.ExecuteWithContext(ctx, func(ctx context.Context) error {
		return hc.circuitBreaker.ExecuteWithContext(ctx, func(ctx context.Context) error {
			// Clone request for retries
			retryReq := req.Clone(ctx)

			resp, err = hc.client.Do(retryReq)
			if err != nil {
				return err
			}

			// Consider 5xx as errors for circuit breaker
			if resp.StatusCode >= 500 {
				return &ReliabilityError{
					Type:       "server_error",
					Message:    "server returned 5xx",
					StatusCode: resp.StatusCode,
				}
			}

			return nil
		})
	})

	if bulkheadErr != nil {
		return nil, bulkheadErr
	}

	return resp, err
}

// DoWithRetry executes HTTP request with retry logic
func (hc *HTTPClient) DoWithRetry(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	return RetryWithResult(ctx, hc.name, hc.retryConfig, func() (*http.Response, error) {
		// Clone request for each retry attempt
		retryReq := req.Clone(ctx)
		return hc.Do(retryReq)
	})
}

// ReliabilityError represents reliability pattern errors
type ReliabilityError struct {
	Type       string
	Message    string
	StatusCode int
}

func (e *ReliabilityError) Error() string {
	if e.StatusCode > 0 {
		return e.Type + ": " + e.Message + " (status: " + string(rune(e.StatusCode)) + ")"
	}
	return e.Type + ": " + e.Message
}

// GetCircuitBreaker returns the circuit breaker for manual control
func (hc *HTTPClient) GetCircuitBreaker() *CircuitBreaker {
	return hc.circuitBreaker
}

// GetBulkhead returns the bulkhead for monitoring
func (hc *HTTPClient) GetBulkhead() *Bulkhead {
	return hc.bulkhead
}
