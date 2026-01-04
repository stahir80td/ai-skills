package sod

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// prometheusMetrics implements MetricsCollector for Prometheus
type prometheusMetrics struct {
	sodScore          *prometheus.GaugeVec
	sodScoreHistogram *prometheus.HistogramVec
	errorCount        *prometheus.CounterVec
	mttd              *prometheus.HistogramVec
	mttr              *prometheus.HistogramVec
	severityScore     *prometheus.GaugeVec
	occurrenceScore   *prometheus.GaugeVec
	detectScore       *prometheus.GaugeVec
}

// NewPrometheusMetrics creates a Prometheus-based metrics collector
func NewPrometheusMetrics(serviceName string) MetricsCollector {
	return &prometheusMetrics{
		sodScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sod_score_total",
				Help: "Current SOD score (Severity x Occurrence x Detectability) for errors",
			},
			[]string{"service", "error_code", "severity_level"},
		),
		sodScoreHistogram: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "sod_score_distribution",
				Help:    "Distribution of SOD scores over time",
				Buckets: []float64{0, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000},
			},
			[]string{"service", "error_code"},
		),
		errorCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sod_error_occurrences_total",
				Help: "Total number of error occurrences tracked by SOD framework",
			},
			[]string{"service", "error_code", "severity"},
		),
		mttd: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "sod_mean_time_to_detect_seconds",
				Help:    "Mean time to detect errors in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
			},
			[]string{"service", "error_code"},
		),
		mttr: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "sod_mean_time_to_resolve_seconds",
				Help:    "Mean time to resolve errors in seconds",
				Buckets: []float64{60, 300, 600, 1800, 3600, 7200, 14400, 28800, 86400},
			},
			[]string{"service", "error_code"},
		),
		severityScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sod_severity_score",
				Help: "Severity component of SOD score (1-10)",
			},
			[]string{"service", "error_code"},
		),
		occurrenceScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sod_occurrence_score",
				Help: "Occurrence component of SOD score (1-10)",
			},
			[]string{"service", "error_code"},
		),
		detectScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sod_detect_score",
				Help: "Detectability component of SOD score (1-10)",
			},
			[]string{"service", "error_code"},
		),
	}
}

// RecordSODScore records an SOD score for an error
func (m *prometheusMetrics) RecordSODScore(errorCode string, score Score) {
	serviceName := "unknown" // Will be injected via constructor in production

	severityLevel := "low"
	if score.Severity >= 9 {
		severityLevel = "critical"
	} else if score.Severity >= 7 {
		severityLevel = "high"
	} else if score.Severity >= 4 {
		severityLevel = "medium"
	}

	m.sodScore.WithLabelValues(serviceName, errorCode, severityLevel).Set(float64(score.Total))
	m.sodScoreHistogram.WithLabelValues(serviceName, errorCode).Observe(float64(score.Total))

	// Record component scores
	m.severityScore.WithLabelValues(serviceName, errorCode).Set(float64(score.Severity))
	m.occurrenceScore.WithLabelValues(serviceName, errorCode).Set(float64(score.Occurrence))
	m.detectScore.WithLabelValues(serviceName, errorCode).Set(float64(score.Detect))
}

// RecordErrorOccurrence tracks error occurrence
func (m *prometheusMetrics) RecordErrorOccurrence(errorCode string, severity string) {
	serviceName := "unknown"
	m.errorCount.WithLabelValues(serviceName, errorCode, severity).Inc()
}

// RecordMTTD tracks Mean Time To Detect
func (m *prometheusMetrics) RecordMTTD(errorCode string, duration time.Duration) {
	serviceName := "unknown"
	m.mttd.WithLabelValues(serviceName, errorCode).Observe(duration.Seconds())
}

// RecordMTTR tracks Mean Time To Resolve
func (m *prometheusMetrics) RecordMTTR(errorCode string, duration time.Duration) {
	serviceName := "unknown"
	m.mttr.WithLabelValues(serviceName, errorCode).Observe(duration.Seconds())
}

// nopMetrics is a no-op implementation for testing or when metrics are disabled
type nopMetrics struct{}

// NewNopMetrics creates a no-op metrics collector
func NewNopMetrics() MetricsCollector {
	return &nopMetrics{}
}

func (m *nopMetrics) RecordSODScore(errorCode string, score Score)            {}
func (m *nopMetrics) RecordErrorOccurrence(errorCode string, severity string) {}
func (m *nopMetrics) RecordMTTD(errorCode string, duration time.Duration)     {}
func (m *nopMetrics) RecordMTTR(errorCode string, duration time.Duration)     {}
