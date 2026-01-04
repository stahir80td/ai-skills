package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ServiceMetrics provides comprehensive metrics for a service following the Four Golden Signals:
// 1. Latency - How long requests take
// 2. Traffic - How many requests the service handles
// 3. Errors - Rate of failed requests
// 4. Saturation - Resource utilization
type ServiceMetrics struct {
	serviceName string

	// Latency (Golden Signal #1)
	requestDuration *prometheus.HistogramVec

	// Traffic (Golden Signal #2)
	requestTotal *prometheus.CounterVec

	// Errors (Golden Signal #3)
	errorTotal *prometheus.CounterVec

	// Saturation (Golden Signal #4)
	resourceUtilization *prometheus.GaugeVec
	activeRequests      prometheus.Gauge
	queueDepth          prometheus.Gauge
}

// Config holds configuration for metrics
type Config struct {
	ServiceName       string
	Namespace         string // Prometheus namespace (default: "iot_homeguard")
	Subsystem         string // Prometheus subsystem (optional)
	LatencyBuckets    []float64
	EnableGoProfiling bool // Enable Go runtime metrics
}

// NewServiceMetrics creates a new metrics instance for a service
func NewServiceMetrics(config Config) *ServiceMetrics {
	if config.Namespace == "" {
		config.Namespace = "iot_homeguard"
	}

	if config.LatencyBuckets == nil {
		// Default buckets: 10ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
		config.LatencyBuckets = []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}
	}

	m := &ServiceMetrics{
		serviceName: config.ServiceName,
	}

	// Latency: Request duration histogram
	m.requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "request_duration_seconds",
			Help:      "Request latency in seconds (Golden Signal: Latency)",
			Buckets:   config.LatencyBuckets,
		},
		[]string{"service", "method", "endpoint", "status"},
	)

	// Traffic: Total requests counter
	m.requestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "requests_total",
			Help:      "Total number of requests (Golden Signal: Traffic)",
		},
		[]string{"service", "method", "endpoint", "status"},
	)

	// Errors: Error counter
	m.errorTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "errors_total",
			Help:      "Total number of errors (Golden Signal: Errors)",
		},
		[]string{"service", "error_code", "severity", "component"},
	)

	// Saturation: Resource utilization gauge
	m.resourceUtilization = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "resource_utilization",
			Help:      "Resource utilization percentage 0-100 (Golden Signal: Saturation)",
		},
		[]string{"service", "resource_type"},
	)

	// Saturation: Active requests gauge
	m.activeRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "active_requests",
			Help:      "Number of requests currently being processed (Golden Signal: Saturation)",
			ConstLabels: prometheus.Labels{
				"service": config.ServiceName,
			},
		},
	)

	// Saturation: Queue depth gauge
	m.queueDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "queue_depth",
			Help:      "Number of items in processing queue (Golden Signal: Saturation)",
			ConstLabels: prometheus.Labels{
				"service": config.ServiceName,
			},
		},
	)

	return m
}

// RecordRequest records a completed request with latency and status
// Golden Signals: Latency + Traffic
func (m *ServiceMetrics) RecordRequest(method, endpoint, status string, duration time.Duration) {
	labels := prometheus.Labels{
		"service":  m.serviceName,
		"method":   method,
		"endpoint": endpoint,
		"status":   status,
	}

	m.requestDuration.With(labels).Observe(duration.Seconds())
	m.requestTotal.With(labels).Inc()
}

// RecordError records an error occurrence
// Golden Signal: Errors
func (m *ServiceMetrics) RecordError(errorCode, severity, component string) {
	labels := prometheus.Labels{
		"service":    m.serviceName,
		"error_code": errorCode,
		"severity":   severity,
		"component":  component,
	}

	m.errorTotal.With(labels).Inc()
}

// UpdateResourceUtilization updates resource utilization percentage (0-100)
// Golden Signal: Saturation
func (m *ServiceMetrics) UpdateResourceUtilization(resourceType string, percentage float64) {
	labels := prometheus.Labels{
		"service":       m.serviceName,
		"resource_type": resourceType,
	}

	m.resourceUtilization.With(labels).Set(percentage)
}

// IncActiveRequests increments the active requests counter
// Golden Signal: Saturation
func (m *ServiceMetrics) IncActiveRequests() {
	m.activeRequests.Inc()
}

// DecActiveRequests decrements the active requests counter
// Golden Signal: Saturation
func (m *ServiceMetrics) DecActiveRequests() {
	m.activeRequests.Dec()
}

// SetQueueDepth sets the current queue depth
// Golden Signal: Saturation
func (m *ServiceMetrics) SetQueueDepth(depth float64) {
	m.queueDepth.Set(depth)
}

// RequestTimer is a helper for timing requests
type RequestTimer struct {
	metrics  *ServiceMetrics
	method   string
	endpoint string
	start    time.Time
}

// NewRequestTimer creates a timer for tracking request duration
func (m *ServiceMetrics) NewRequestTimer(method, endpoint string) *RequestTimer {
	m.IncActiveRequests()
	return &RequestTimer{
		metrics:  m,
		method:   method,
		endpoint: endpoint,
		start:    time.Now(),
	}
}

// Done completes the request timer and records metrics
func (t *RequestTimer) Done(status string) {
	duration := time.Since(t.start)
	t.metrics.RecordRequest(t.method, t.endpoint, status, duration)
	t.metrics.DecActiveRequests()
}

// DoneWithError completes the timer and records both request and error metrics
func (t *RequestTimer) DoneWithError(status, errorCode, severity, component string) {
	duration := time.Since(t.start)
	t.metrics.RecordRequest(t.method, t.endpoint, status, duration)
	t.metrics.RecordError(errorCode, severity, component)
	t.metrics.DecActiveRequests()
}
