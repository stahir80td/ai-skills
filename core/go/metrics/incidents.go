package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// IncidentMetrics tracks incident lifecycle for MTTR and incident response dashboards
type IncidentMetrics struct {
	serviceName     string
	activeGauge     *prometheus.GaugeVec
	mttrHistogram   *prometheus.HistogramVec
	totalCounter    *prometheus.CounterVec
	activeIncidents sync.Map // incident_id -> *ActiveIncident
	mu              sync.RWMutex
}

// ActiveIncident represents an ongoing incident
type ActiveIncident struct {
	IncidentID   string
	Service      string
	Severity     string
	IncidentType string
	StartTime    time.Time
	Context      context.Context
	Cancel       context.CancelFunc
}

var (
	incidentActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "incident_active",
			Help: "Number of active incidents by service, severity, and type",
		},
		[]string{"service", "severity", "incident_type"},
	)

	incidentMTTR = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "incident_mttr_minutes",
			Help:    "Mean Time To Resolution in minutes by service and severity",
			Buckets: []float64{1, 5, 10, 15, 30, 60, 120, 240, 480, 960}, // 1min to 16 hours
		},
		[]string{"service", "severity"},
	)

	incidentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "incident_total",
			Help: "Total number of incidents by service, severity, and type",
		},
		[]string{"service", "severity", "incident_type"},
	)

	// Additional incident lifecycle metrics
	incidentDetectedTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "incident_detected_timestamp",
			Help: "Unix timestamp when incident was detected",
		},
		[]string{"incident_id", "service"},
	)

	incidentAcknowledgedTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "incident_acknowledged_timestamp",
			Help: "Unix timestamp when incident was acknowledged",
		},
		[]string{"incident_id", "service"},
	)

	incidentInvestigatingTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "incident_investigating_timestamp",
			Help: "Unix timestamp when incident investigation started",
		},
		[]string{"incident_id", "service"},
	)

	incidentResolvedTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "incident_resolved_timestamp",
			Help: "Unix timestamp when incident was resolved",
		},
		[]string{"incident_id", "service"},
	)
)

// NewIncidentMetrics creates a new incident metrics tracker for a service
func NewIncidentMetrics(serviceName string) *IncidentMetrics {
	return &IncidentMetrics{
		serviceName:   serviceName,
		activeGauge:   incidentActive,
		mttrHistogram: incidentMTTR,
		totalCounter:  incidentTotal,
	}
}

// StartIncident registers a new incident and starts tracking it
// Returns incident ID and context for cancellation
func (im *IncidentMetrics) StartIncident(severity, incidentType string) (string, context.Context) {
	incidentID := generateIncidentID(im.serviceName)
	ctx, cancel := context.WithCancel(context.Background())

	incident := &ActiveIncident{
		IncidentID:   incidentID,
		Service:      im.serviceName,
		Severity:     severity,
		IncidentType: incidentType,
		StartTime:    time.Now(),
		Context:      ctx,
		Cancel:       cancel,
	}

	im.activeIncidents.Store(incidentID, incident)

	// Increment active incidents gauge
	im.activeGauge.WithLabelValues(im.serviceName, severity, incidentType).Inc()

	// Increment total incidents counter
	im.totalCounter.WithLabelValues(im.serviceName, severity, incidentType).Inc()

	// Record detection timestamp
	incidentDetectedTimestamp.WithLabelValues(incidentID, im.serviceName).Set(float64(incident.StartTime.Unix()))

	return incidentID, ctx
}

// AcknowledgeIncident marks an incident as acknowledged
func (im *IncidentMetrics) AcknowledgeIncident(incidentID string) {
	if incident, ok := im.activeIncidents.Load(incidentID); ok {
		inc := incident.(*ActiveIncident)
		incidentAcknowledgedTimestamp.WithLabelValues(incidentID, im.serviceName).Set(float64(time.Now().Unix()))
		_ = inc // Prevent unused variable warning
	}
}

// StartInvestigation marks the incident as under investigation
func (im *IncidentMetrics) StartInvestigation(incidentID string) {
	if incident, ok := im.activeIncidents.Load(incidentID); ok {
		inc := incident.(*ActiveIncident)
		incidentInvestigatingTimestamp.WithLabelValues(incidentID, im.serviceName).Set(float64(time.Now().Unix()))
		_ = inc
	}
}

// ResolveIncident marks an incident as resolved and calculates MTTR
func (im *IncidentMetrics) ResolveIncident(incidentID string) {
	value, ok := im.activeIncidents.Load(incidentID)
	if !ok {
		return
	}

	incident := value.(*ActiveIncident)
	resolvedTime := time.Now()
	mttrMinutes := resolvedTime.Sub(incident.StartTime).Minutes()

	// Record MTTR
	im.mttrHistogram.WithLabelValues(im.serviceName, incident.Severity).Observe(mttrMinutes)

	// Record resolution timestamp
	incidentResolvedTimestamp.WithLabelValues(incidentID, im.serviceName).Set(float64(resolvedTime.Unix()))

	// Decrement active incidents gauge
	im.activeGauge.WithLabelValues(im.serviceName, incident.Severity, incident.IncidentType).Dec()

	// Cancel incident context
	incident.Cancel()

	// Remove from active incidents
	im.activeIncidents.Delete(incidentID)

	// Clean up lifecycle metrics
	go func() {
		time.Sleep(5 * time.Minute) // Keep metrics for 5 minutes for dashboards
		incidentDetectedTimestamp.DeleteLabelValues(incidentID, im.serviceName)
		incidentAcknowledgedTimestamp.DeleteLabelValues(incidentID, im.serviceName)
		incidentInvestigatingTimestamp.DeleteLabelValues(incidentID, im.serviceName)
		incidentResolvedTimestamp.DeleteLabelValues(incidentID, im.serviceName)
	}()
}

// GetActiveIncidents returns all active incidents for the service
func (im *IncidentMetrics) GetActiveIncidents() []*ActiveIncident {
	incidents := []*ActiveIncident{}
	im.activeIncidents.Range(func(key, value interface{}) bool {
		incident := value.(*ActiveIncident)
		incidents = append(incidents, incident)
		return true
	})
	return incidents
}

// AutoResolveIncident creates an incident and auto-resolves it after the context is done
// Useful for wrapping error-prone operations
func (im *IncidentMetrics) AutoResolveIncident(ctx context.Context, severity, incidentType string, fn func(incidentID string, incidentCtx context.Context) error) error {
	incidentID, incidentCtx := im.StartIncident(severity, incidentType)
	im.AcknowledgeIncident(incidentID)
	im.StartInvestigation(incidentID)

	err := fn(incidentID, incidentCtx)

	// Auto-resolve when function completes (successful or not)
	im.ResolveIncident(incidentID)

	return err
}

// Helper function to generate unique incident IDs
func generateIncidentID(serviceName string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", serviceName, timestamp)
}
