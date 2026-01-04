package sli

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Global metrics - created once and shared across all trackers
	globalRequestsTotal   *prometheus.CounterVec
	globalRequestsSuccess *prometheus.CounterVec
	globalRequestsFailed  *prometheus.CounterVec
	globalRequestDuration *prometheus.HistogramVec
	globalThroughputRate  *prometheus.GaugeVec
	globalSliAvailability *prometheus.GaugeVec
	globalSliLatencyP95   *prometheus.GaugeVec
	globalSliLatencyP99   *prometheus.GaugeVec
	globalSliErrorRate    *prometheus.GaugeVec

	// Ensure metrics are initialized only once
	metricsOnce sync.Once
)

// initGlobalMetrics initializes all global metrics once
func initGlobalMetrics() {
	metricsOnce.Do(func() {
		globalRequestsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sli_requests_total",
				Help: "Total number of requests for SLI tracking",
			},
			[]string{"service", "operation"},
		)

		globalRequestsSuccess = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sli_requests_success_total",
				Help: "Total number of successful requests",
			},
			[]string{"service", "operation"},
		)

		globalRequestsFailed = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sli_requests_failed_total",
				Help: "Total number of failed requests",
			},
			[]string{"service", "operation", "error_code", "severity"},
		)

		globalRequestDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "sli_request_duration_seconds",
				Help:    "Request duration in seconds for SLI tracking",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"service", "operation"},
		)

		globalThroughputRate = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sli_throughput_rate",
				Help: "Current throughput rate (requests or messages per second)",
			},
			[]string{"service", "operation", "type"},
		)

		globalSliAvailability = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sli_availability_percent",
				Help: "Current availability SLI in percent",
			},
			[]string{"service"},
		)

		globalSliLatencyP95 = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sli_latency_p95_milliseconds",
				Help: "P95 latency SLI in milliseconds",
			},
			[]string{"service", "operation"},
		)

		globalSliLatencyP99 = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sli_latency_p99_milliseconds",
				Help: "P99 latency SLI in milliseconds",
			},
			[]string{"service", "operation"},
		)

		globalSliErrorRate = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sli_error_rate_percent",
				Help: "Current error rate SLI in percent",
			},
			[]string{"service"},
		)
	})
}

// prometheusTracker implements Tracker using Prometheus metrics
type prometheusTracker struct {
	serviceName string

	// References to global metrics
	requestsTotal   *prometheus.CounterVec
	requestsSuccess *prometheus.CounterVec
	requestsFailed  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	throughputRate  *prometheus.GaugeVec
	sliAvailability *prometheus.GaugeVec
	sliLatencyP95   *prometheus.GaugeVec
	sliLatencyP99   *prometheus.GaugeVec
	sliErrorRate    *prometheus.GaugeVec
}

// NewPrometheusTracker creates a Prometheus-based SLI tracker
func NewPrometheusTracker(serviceName string) Tracker {
	// Initialize global metrics once
	initGlobalMetrics()

	return &prometheusTracker{
		serviceName:     serviceName,
		requestsTotal:   globalRequestsTotal,
		requestsSuccess: globalRequestsSuccess,
		requestsFailed:  globalRequestsFailed,
		requestDuration: globalRequestDuration,
		throughputRate:  globalThroughputRate,
		sliAvailability: globalSliAvailability,
		sliLatencyP95:   globalSliLatencyP95,
		sliLatencyP99:   globalSliLatencyP99,
		sliErrorRate:    globalSliErrorRate,
	}
}

// RecordRequest tracks a request with its outcome
func (t *prometheusTracker) RecordRequest(ctx context.Context, outcome RequestOutcome) {
	operation := outcome.Operation
	if operation == "" {
		operation = "default"
	}

	// Increment total requests
	t.requestsTotal.WithLabelValues(t.serviceName, operation).Inc()

	// Track success/failure
	if outcome.Success {
		t.requestsSuccess.WithLabelValues(t.serviceName, operation).Inc()
	} else {
		t.requestsFailed.WithLabelValues(
			t.serviceName,
			operation,
			outcome.ErrorCode,
			outcome.ErrorSeverity,
		).Inc()
	}

	// Record latency if available
	if outcome.Latency > 0 {
		t.requestDuration.WithLabelValues(t.serviceName, operation).Observe(outcome.Latency.Seconds())
	}
}

// RecordLatency tracks request latency
func (t *prometheusTracker) RecordLatency(ctx context.Context, duration time.Duration, operation string) {
	if operation == "" {
		operation = "default"
	}

	t.requestDuration.WithLabelValues(t.serviceName, operation).Observe(duration.Seconds())
}

// RecordThroughput tracks throughput events
func (t *prometheusTracker) RecordThroughput(ctx context.Context, count int, operation string) {
	if operation == "" {
		operation = "default"
	}

	t.throughputRate.WithLabelValues(t.serviceName, operation, "requests").Add(float64(count))
}

// GetMetrics returns current SLI metrics snapshot
// Note: This would typically query Prometheus API for real-time data
// For now, returns a placeholder - real implementation would use Prometheus API client
func (t *prometheusTracker) GetMetrics(ctx context.Context) (*Metrics, error) {
	// In production, this would query Prometheus API:
	// - rate(sli_requests_total[5m])
	// - histogram_quantile(0.95, sli_request_duration_seconds_bucket)
	// - rate(sli_requests_failed_total[5m]) / rate(sli_requests_total[5m])

	return &Metrics{
		WindowEnd:   time.Now(),
		WindowStart: time.Now().Add(-5 * time.Minute),
	}, nil
}

// nopTracker is a no-op implementation for testing
type nopTracker struct{}

// NewNopTracker creates a no-op tracker
func NewNopTracker() Tracker {
	return &nopTracker{}
}

func (t *nopTracker) RecordRequest(ctx context.Context, outcome RequestOutcome)                   {}
func (t *nopTracker) RecordLatency(ctx context.Context, duration time.Duration, operation string) {}
func (t *nopTracker) RecordThroughput(ctx context.Context, count int, operation string)           {}
func (t *nopTracker) GetMetrics(ctx context.Context) (*Metrics, error)                            { return &Metrics{}, nil }
