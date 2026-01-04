package metrics

import (
	"context"
	"time"
)

// Common incident types
const (
	IncidentTypeServiceDown      = "service_down"
	IncidentTypeDatabaseDown     = "database_down"
	IncidentTypeHighErrorRate    = "high_error_rate"
	IncidentTypeHighLatency      = "high_latency"
	IncidentTypeCircuitOpen      = "circuit_breaker_open"
	IncidentTypeKafkaLag         = "kafka_consumer_lag"
	IncidentTypeMemoryExhaustion = "memory_exhaustion"
	IncidentTypeCPUExhaustion    = "cpu_exhaustion"
	IncidentTypeDiskFull         = "disk_full"
	IncidentTypeNetworkPartition = "network_partition"
)

// Common severity levels
const (
	SeverityCritical = "critical"
	SeverityWarning  = "warning"
	SeverityInfo     = "info"
)

// MonitorErrorRate automatically creates incidents when error rate exceeds threshold
func (im *IncidentMetrics) MonitorErrorRate(ctx context.Context, errorRateThreshold float64, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	var activeIncidentID string

	for {
		select {
		case <-ctx.Done():
			// Resolve incident if still active when monitoring stops
			if activeIncidentID != "" {
				im.ResolveIncident(activeIncidentID)
			}
			return
		case <-ticker.C:
			// This is a placeholder - actual implementation would query Prometheus
			// or track error rate internally
			// For now, services will call TrackError() manually
		}
	}
}

// TrackCriticalError creates an incident for critical errors
func (im *IncidentMetrics) TrackCriticalError(errorType string) (string, context.Context) {
	return im.StartIncident(SeverityCritical, errorType)
}

// TrackDatabaseDowntime creates an incident when database is unavailable
func (im *IncidentMetrics) TrackDatabaseDowntime() (string, context.Context) {
	return im.StartIncident(SeverityCritical, IncidentTypeDatabaseDown)
}

// TrackHighLatency creates an incident when latency exceeds threshold
func (im *IncidentMetrics) TrackHighLatency() (string, context.Context) {
	return im.StartIncident(SeverityWarning, IncidentTypeHighLatency)
}

// TrackKafkaLag creates an incident when Kafka consumer lag is high
func (im *IncidentMetrics) TrackKafkaLag() (string, context.Context) {
	return im.StartIncident(SeverityWarning, IncidentTypeKafkaLag)
}
