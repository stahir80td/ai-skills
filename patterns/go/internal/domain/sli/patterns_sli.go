package sli

import (
	"context"
	"sync"
	"time"

	coresli "github.com/your-github-org/ai-scaffolder/core/go/sli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PatternsSli implements SLI tracking for the patterns service
// Demonstrates comprehensive SLI/SLO monitoring patterns
type PatternsSli struct {
	tracker coresli.Tracker

	// Business metrics - products/orders
	productsCreated *prometheus.CounterVec
	activeProducts  prometheus.Gauge
	orderValue      *prometheus.HistogramVec
	ordersByStatus  *prometheus.CounterVec

	// Infrastructure metrics
	databaseOps   *prometheus.CounterVec
	cacheOps      *prometheus.CounterVec
	externalCalls *prometheus.CounterVec

	// Telemetry metrics
	telemetryRecords  *prometheus.CounterVec
	anomaliesDetected *prometheus.CounterVec
}

var (
	// Singleton instance
	instance *PatternsSli
	once     sync.Once
)

// NewPatternsSli creates a new patterns SLI tracker
func NewPatternsSli(serviceName string) *PatternsSli {
	once.Do(func() {
		instance = &PatternsSli{
			tracker: coresli.NewPrometheusTracker(serviceName),

			// Business metrics
			productsCreated: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "patterns",
					Name:      "products_created_total",
					Help:      "Total products created",
				},
				[]string{"category", "status"},
			),

			activeProducts: promauto.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "patterns",
					Name:      "products_active",
					Help:      "Current active products count",
				},
			),

			orderValue: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: "patterns",
					Name:      "order_value_dollars",
					Help:      "Order value distribution",
					Buckets:   []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
				},
				[]string{"category"},
			),

			ordersByStatus: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "patterns",
					Name:      "orders_by_status_total",
					Help:      "Orders by status transitions",
				},
				[]string{"from_status", "to_status"},
			),

			// Infrastructure metrics
			databaseOps: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "patterns",
					Name:      "database_operations_total",
					Help:      "Database operations by type and status",
				},
				[]string{"operation", "table", "status"},
			),

			cacheOps: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "patterns",
					Name:      "cache_operations_total",
					Help:      "Cache operations by type and status",
				},
				[]string{"operation", "status"},
			),

			externalCalls: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "patterns",
					Name:      "external_calls_total",
					Help:      "External service calls by service and status",
				},
				[]string{"service", "operation", "status"},
			),

			// Telemetry metrics
			telemetryRecords: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "patterns",
					Name:      "telemetry_records_total",
					Help:      "Telemetry records by device and metric type",
				},
				[]string{"device_type", "metric"},
			),

			anomaliesDetected: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "patterns",
					Name:      "anomalies_detected_total",
					Help:      "Anomalies detected by type",
				},
				[]string{"anomaly_type", "severity"},
			),
		}
	})
	return instance
}

// RecordRequest delegates to the core SLI tracker
func (p *PatternsSli) RecordRequest(ctx context.Context, outcome coresli.RequestOutcome) {
	p.tracker.RecordRequest(ctx, outcome)
}

// RecordLatency delegates to the core SLI tracker
func (p *PatternsSli) RecordLatency(ctx context.Context, duration time.Duration, operation string) {
	p.tracker.RecordLatency(ctx, duration, operation)
}

// RecordThroughput delegates to the core SLI tracker
func (p *PatternsSli) RecordThroughput(ctx context.Context, count int, operation string) {
	p.tracker.RecordThroughput(ctx, count, operation)
}

// RecordProductCreated records a product creation
func (p *PatternsSli) RecordProductCreated(category, status string, price float64) {
	p.productsCreated.WithLabelValues(category, status).Inc()
	p.orderValue.WithLabelValues(category).Observe(price)

	if status == "active" {
		p.activeProducts.Inc()
	}
}

// RecordProductStatusChanged records a product status change
func (p *PatternsSli) RecordProductStatusChanged(oldStatus, newStatus string) {
	p.ordersByStatus.WithLabelValues(oldStatus, newStatus).Inc()

	if oldStatus != "active" && newStatus == "active" {
		p.activeProducts.Inc()
	} else if oldStatus == "active" && newStatus != "active" {
		p.activeProducts.Dec()
	}
}

// RecordOrderCreated records an order creation with value
func (p *PatternsSli) RecordOrderCreated(category string, totalAmount float64) {
	p.orderValue.WithLabelValues(category).Observe(totalAmount)
}

// RecordDatabaseOperation records a database operation
func (p *PatternsSli) RecordDatabaseOperation(operation, table string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	p.databaseOps.WithLabelValues(operation, table, status).Inc()
}

// RecordCacheOperation records a cache operation
func (p *PatternsSli) RecordCacheOperation(operation string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	p.cacheOps.WithLabelValues(operation, status).Inc()
}

// RecordExternalCall records an external service call
func (p *PatternsSli) RecordExternalCall(service, operation string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	p.externalCalls.WithLabelValues(service, operation, status).Inc()
}

// RecordTelemetry records a telemetry record
func (p *PatternsSli) RecordTelemetry(deviceType, metric string) {
	p.telemetryRecords.WithLabelValues(deviceType, metric).Inc()
}

// RecordAnomaly records an anomaly detection
func (p *PatternsSli) RecordAnomaly(anomalyType, severity string) {
	p.anomaliesDetected.WithLabelValues(anomalyType, severity).Inc()
}

// GetMetrics returns current SLI metrics
func (p *PatternsSli) GetMetrics(ctx context.Context) (*coresli.Metrics, error) {
	return p.tracker.GetMetrics(ctx)
}

// RecordOrderCreationSuccess records a successful order creation with latency
func (p *PatternsSli) RecordOrderCreationSuccess(duration time.Duration) {
	p.tracker.RecordRequest(context.Background(), coresli.RequestOutcome{
		Success:   true,
		Operation: "order_creation",
		Latency:   duration,
		Timestamp: time.Now(),
	})
	p.tracker.RecordLatency(context.Background(), duration, "order_creation")
}

// RecordOrderCreationFailure records a failed order creation
func (p *PatternsSli) RecordOrderCreationFailure() {
	p.tracker.RecordRequest(context.Background(), coresli.RequestOutcome{
		Success:   false,
		Operation: "order_creation",
		Timestamp: time.Now(),
	})
}

// RecordTelemetryIngestionSuccess records a successful telemetry ingestion
func (p *PatternsSli) RecordTelemetryIngestionSuccess(duration time.Duration) {
	p.tracker.RecordRequest(context.Background(), coresli.RequestOutcome{
		Success:   true,
		Operation: "telemetry_ingestion",
		Latency:   duration,
		Timestamp: time.Now(),
	})
	p.tracker.RecordLatency(context.Background(), duration, "telemetry_ingestion")
}

// RecordTelemetryIngestionFailure records a failed telemetry ingestion
func (p *PatternsSli) RecordTelemetryIngestionFailure() {
	p.tracker.RecordRequest(context.Background(), coresli.RequestOutcome{
		Success:   false,
		Operation: "telemetry_ingestion",
		Timestamp: time.Now(),
	})
}
