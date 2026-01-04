package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Global singleton metrics collectors (to avoid duplicate registration)
	serviceRequestsTotalOnce sync.Once
	serviceRequestsTotal     *prometheus.CounterVec

	serviceLatencyOnce sync.Once
	serviceLatency     *prometheus.HistogramVec

	serviceErrorsOnce sync.Once
	serviceErrors     *prometheus.CounterVec
)

// ServiceMeshMetrics tracks inter-service communication for service mesh observability
type ServiceMeshMetrics struct {
	sourceService string

	// Service-to-service request metrics
	serviceRequestsTotal *prometheus.CounterVec
	serviceLatency       *prometheus.HistogramVec
	serviceErrors        *prometheus.CounterVec
}

// NewServiceMeshMetrics creates metrics for tracking service mesh communication
func NewServiceMeshMetrics(sourceService string) *ServiceMeshMetrics {
	// Initialize global metrics once
	serviceRequestsTotalOnce.Do(func() {
		serviceRequestsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "service_requests_total",
				Help: "Total requests between services for service mesh topology",
			},
			[]string{"source_service", "target_service", "method", "status"},
		)
	})

	serviceLatencyOnce.Do(func() {
		serviceLatency = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "service_request_duration_seconds",
				Help:    "Service-to-service request latency in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			},
			[]string{"source_service", "target_service", "method"},
		)
	})

	serviceErrorsOnce.Do(func() {
		serviceErrors = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "service_errors_total",
				Help: "Total errors in service-to-service communication",
			},
			[]string{"source_service", "target_service", "error_type"},
		)
	})

	return &ServiceMeshMetrics{
		sourceService: sourceService,

		serviceRequestsTotal: serviceRequestsTotal,
		serviceLatency:       serviceLatency,
		serviceErrors:        serviceErrors,
	}
}

// TrackRequest records a service-to-service request
func (m *ServiceMeshMetrics) TrackRequest(targetService, method string, statusCode int, duration time.Duration) {
	status := http.StatusText(statusCode)
	if statusCode == 0 {
		status = "unknown"
	}

	m.serviceRequestsTotal.WithLabelValues(
		m.sourceService,
		targetService,
		method,
		status,
	).Inc()

	m.serviceLatency.WithLabelValues(
		m.sourceService,
		targetService,
		method,
	).Observe(duration.Seconds())

	// Track errors separately
	if statusCode >= 400 {
		errorType := "client_error"
		if statusCode >= 500 {
			errorType = "server_error"
		}
		m.serviceErrors.WithLabelValues(
			m.sourceService,
			targetService,
			errorType,
		).Inc()
	}
}

// TrackError records a service-to-service communication error (network, timeout, etc.)
func (m *ServiceMeshMetrics) TrackError(targetService, errorType string) {
	m.serviceErrors.WithLabelValues(
		m.sourceService,
		targetService,
		errorType,
	).Inc()
}

// HTTPClientMiddleware wraps http.Client to automatically track service mesh metrics
type HTTPClientMiddleware struct {
	metrics       *ServiceMeshMetrics
	targetService string
	client        *http.Client
}

// NewHTTPClientMiddleware creates an HTTP client that automatically tracks service mesh metrics
func NewHTTPClientMiddleware(sourceService, targetService string, client *http.Client, fallbackTimeout time.Duration) *HTTPClientMiddleware {
	if client == nil {
		client = &http.Client{Timeout: fallbackTimeout} // NO HARDCODING - from config (MINIMUM 60s)
	}

	return &HTTPClientMiddleware{
		metrics:       NewServiceMeshMetrics(sourceService),
		targetService: targetService,
		client:        client,
	}
}

// Do executes the HTTP request and tracks metrics
func (m *HTTPClientMiddleware) Do(req *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := m.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		m.metrics.TrackError(m.targetService, "network_error")
		return nil, err
	}

	m.metrics.TrackRequest(
		m.targetService,
		req.Method,
		resp.StatusCode,
		duration,
	)

	return resp, nil
}

// Get performs a GET request with metrics tracking
func (m *HTTPClientMiddleware) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return m.Do(req)
}

// Post performs a POST request with metrics tracking
func (m *HTTPClientMiddleware) Post(ctx context.Context, url, contentType string, body interface{}) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return m.Do(req)
}
