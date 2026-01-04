package reliability

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half_open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// CircuitBreaker prevents cascading failures by opening after threshold failures
type CircuitBreaker struct {
	name             string
	maxFailures      uint32
	timeout          time.Duration
	halfOpenRequests uint32

	mu               sync.RWMutex
	state            CircuitState
	failures         uint32
	lastFailTime     time.Time
	halfOpenAttempts uint32

	// Metrics
	stateGauge        prometheus.Gauge
	requestsTotal     *prometheus.CounterVec
	errorsTotal       *prometheus.CounterVec
	stateChangesTotal *prometheus.CounterVec
}

var (
	circuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=half_open, 2=open)",
		},
		[]string{"name"},
	)

	circuitBreakerRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_requests_total",
			Help: "Total requests through circuit breaker",
		},
		[]string{"name", "state", "result"},
	)

	circuitBreakerErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_errors_total",
			Help: "Total errors in circuit breaker",
		},
		[]string{"name"},
	)

	circuitBreakerStateChanges = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_state_changes_total",
			Help: "Total circuit breaker state transitions",
		},
		[]string{"name", "from", "to"},
	)
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures uint32, timeout time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:              name,
		maxFailures:       maxFailures,
		timeout:           timeout,
		halfOpenRequests:  3, // Allow 3 requests in half-open state
		state:             StateClosed,
		stateGauge:        circuitBreakerState.WithLabelValues(name),
		requestsTotal:     circuitBreakerRequests,
		errorsTotal:       circuitBreakerErrors,
		stateChangesTotal: circuitBreakerStateChanges,
	}
	cb.stateGauge.Set(float64(StateClosed))
	return cb
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// Check if circuit should transition from open to half-open
	if cb.state == StateOpen && time.Since(cb.lastFailTime) > cb.timeout {
		cb.setState(StateHalfOpen)
		cb.halfOpenAttempts = 0
	}

	// Reject if circuit is open
	if cb.state == StateOpen {
		cb.mu.Unlock()
		cb.requestsTotal.WithLabelValues(cb.name, "open", "rejected").Inc()
		return errors.New("circuit breaker is open")
	}

	currentState := cb.state
	cb.mu.Unlock()

	// Execute function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onError(currentState)
		cb.requestsTotal.WithLabelValues(cb.name, currentState.String(), "error").Inc()
		return err
	}

	cb.onSuccess(currentState)
	cb.requestsTotal.WithLabelValues(cb.name, currentState.String(), "success").Inc()
	return nil
}

// ExecuteWithContext runs the function with context support
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	return cb.Execute(func() error {
		return fn(ctx)
	})
}

func (cb *CircuitBreaker) onSuccess(state CircuitState) {
	if state == StateHalfOpen {
		cb.halfOpenAttempts++
		if cb.halfOpenAttempts >= cb.halfOpenRequests {
			cb.setState(StateClosed)
			cb.failures = 0
		}
	} else if state == StateClosed {
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) onError(state CircuitState) {
	cb.failures++
	cb.lastFailTime = time.Now()
	cb.errorsTotal.WithLabelValues(cb.name).Inc()

	if state == StateHalfOpen {
		cb.setState(StateOpen)
	} else if cb.failures >= cb.maxFailures {
		cb.setState(StateOpen)
	}
}

func (cb *CircuitBreaker) setState(newState CircuitState) {
	oldState := cb.state
	cb.state = newState
	cb.stateGauge.Set(float64(newState))
	cb.stateChangesTotal.WithLabelValues(cb.name, oldState.String(), newState.String()).Inc()
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.setState(StateClosed)
	cb.failures = 0
	cb.halfOpenAttempts = 0
}
