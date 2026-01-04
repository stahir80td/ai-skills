package reliability

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"
)

// RateLimiter implements token bucket rate limiting per service/endpoint
type RateLimiter struct {
	name    string
	limiter *rate.Limiter
	burst   int

	// Metrics
	requestsTotal   *prometheus.CounterVec
	rejectedTotal   *prometheus.CounterVec
	allowedTotal    *prometheus.CounterVec
	tokensAvailable prometheus.Gauge
}

var (
	rateLimitRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_requests_total",
			Help: "Total requests to rate limiter",
		},
		[]string{"name"},
	)

	rateLimitRejected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_rejected_total",
			Help: "Total requests rejected by rate limiter",
		},
		[]string{"name"},
	)

	rateLimitAllowed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_allowed_total",
			Help: "Total requests allowed by rate limiter",
		},
		[]string{"name"},
	)

	rateLimitTokensAvailable = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rate_limit_tokens_available",
			Help: "Number of tokens currently available",
		},
		[]string{"name"},
	)
)

// NewRateLimiter creates a new rate limiter with specified requests per second and burst
func NewRateLimiter(name string, requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		name:            name,
		limiter:         rate.NewLimiter(rate.Limit(requestsPerSecond), burst),
		burst:           burst,
		requestsTotal:   rateLimitRequests,
		rejectedTotal:   rateLimitRejected,
		allowedTotal:    rateLimitAllowed,
		tokensAvailable: rateLimitTokensAvailable.WithLabelValues(name),
	}
}

// Allow checks if a request can proceed
func (rl *RateLimiter) Allow() bool {
	rl.requestsTotal.WithLabelValues(rl.name).Inc()
	allowed := rl.limiter.Allow()

	// Update tokens available metric
	tokens := rl.limiter.Tokens()
	rl.tokensAvailable.Set(tokens)

	if allowed {
		rl.allowedTotal.WithLabelValues(rl.name).Inc()
	} else {
		rl.rejectedTotal.WithLabelValues(rl.name).Inc()
	}

	return allowed
}

// Wait blocks until request can proceed or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.requestsTotal.WithLabelValues(rl.name).Inc()

	err := rl.limiter.Wait(ctx)
	if err != nil {
		rl.rejectedTotal.WithLabelValues(rl.name).Inc()
		return err
	}

	rl.allowedTotal.WithLabelValues(rl.name).Inc()
	tokens := rl.limiter.Tokens()
	rl.tokensAvailable.Set(tokens)

	return nil
}

// Bulkhead limits concurrent execution to prevent resource exhaustion
type Bulkhead struct {
	name      string
	semaphore chan struct{}
	timeout   time.Duration

	mu             sync.RWMutex
	activeRequests int
	maxConcurrency int

	// Metrics
	activeGauge   prometheus.Gauge
	rejectedTotal *prometheus.CounterVec
	timeoutTotal  *prometheus.CounterVec
	queuedTotal   *prometheus.CounterVec
	executedTotal *prometheus.CounterVec
}

var (
	bulkheadActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bulkhead_active_requests",
			Help: "Number of active requests in bulkhead",
		},
		[]string{"name"},
	)

	bulkheadRejected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bulkhead_rejected_total",
			Help: "Total requests rejected by bulkhead",
		},
		[]string{"name", "reason"},
	)

	bulkheadTimeout = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bulkhead_timeout_total",
			Help: "Total requests that timed out waiting for bulkhead",
		},
		[]string{"name"},
	)

	bulkheadQueued = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bulkhead_queued_total",
			Help: "Total requests queued in bulkhead",
		},
		[]string{"name"},
	)

	bulkheadExecuted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bulkhead_executed_total",
			Help: "Total requests executed through bulkhead",
		},
		[]string{"name"},
	)
)

// NewBulkhead creates a new bulkhead with max concurrency limit
func NewBulkhead(name string, maxConcurrency int, timeout time.Duration) *Bulkhead {
	return &Bulkhead{
		name:           name,
		semaphore:      make(chan struct{}, maxConcurrency),
		timeout:        timeout,
		maxConcurrency: maxConcurrency,
		activeGauge:    bulkheadActive.WithLabelValues(name),
		rejectedTotal:  bulkheadRejected,
		timeoutTotal:   bulkheadTimeout,
		queuedTotal:    bulkheadQueued,
		executedTotal:  bulkheadExecuted,
	}
}

// Execute runs function with concurrency limiting
func (b *Bulkhead) Execute(fn func() error) error {
	return b.ExecuteWithContext(context.Background(), func(ctx context.Context) error {
		return fn()
	})
}

// ExecuteWithContext runs function with concurrency limiting and context support
func (b *Bulkhead) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	// Try to acquire semaphore with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	b.queuedTotal.WithLabelValues(b.name).Inc()

	select {
	case b.semaphore <- struct{}{}:
		// Acquired semaphore
		b.incrementActive()
		defer func() {
			<-b.semaphore
			b.decrementActive()
		}()

		b.executedTotal.WithLabelValues(b.name).Inc()
		return fn(ctx)

	case <-timeoutCtx.Done():
		// Timeout waiting for semaphore
		b.timeoutTotal.WithLabelValues(b.name).Inc()
		b.rejectedTotal.WithLabelValues(b.name, "timeout").Inc()
		return context.DeadlineExceeded

	case <-ctx.Done():
		// Context cancelled
		b.rejectedTotal.WithLabelValues(b.name, "cancelled").Inc()
		return ctx.Err()
	}
}

func (b *Bulkhead) incrementActive() {
	b.mu.Lock()
	b.activeRequests++
	active := b.activeRequests
	b.mu.Unlock()
	b.activeGauge.Set(float64(active))
}

func (b *Bulkhead) decrementActive() {
	b.mu.Lock()
	b.activeRequests--
	active := b.activeRequests
	b.mu.Unlock()
	b.activeGauge.Set(float64(active))
}

// GetActiveCount returns current number of active requests
func (b *Bulkhead) GetActiveCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.activeRequests
}
