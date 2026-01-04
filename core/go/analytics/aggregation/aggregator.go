package aggregation

import (
	"context"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.uber.org/zap"
)

// Event represents a raw device event for aggregation
type Event struct {
	DeviceID  string
	EventType string
	Severity  string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// HourlyStats represents hourly aggregated statistics
type HourlyStats struct {
	DeviceID      string
	HourBucket    time.Time
	EventCount    int64
	CriticalCount int64
	HighCount     int64
	MediumCount   int64
	LowCount      int64
	TempSum       float64
	TempCount     int64
	TempMin       float64
	TempMax       float64
	BatterySum    float64
	BatteryCount  int64
	BatteryMin    float64
	BatteryMax    float64
	EnergySum     float64
	EnergyCount   int64
}

// DailyStats represents daily aggregated statistics
type DailyStats struct {
	DeviceID       string
	DayBucket      time.Time
	EventCount     int64
	HourlyAvgCount float64
	MaxHourlyCount int64
	AlertCount     int64
	UptimeSeconds  int64
	TempAvg        float64
	TempMin        float64
	TempMax        float64
	BatteryAvg     float64
	BatteryMin     float64
	EnergyTotal    float64
}

// Aggregator provides time-series aggregation capabilities
type Aggregator struct {
	logger *logger.Logger
}

// Config holds aggregator configuration
type Config struct {
	Logger *logger.Logger
}

// NewAggregator creates a new aggregator with dependency injection
func NewAggregator(cfg Config) *Aggregator {
	return &Aggregator{
		logger: cfg.Logger,
	}
}

// ComputeHourlyStats aggregates event data into hourly buckets
func (a *Aggregator) ComputeHourlyStats(ctx context.Context, event *Event, features map[string]float64) *HourlyStats {
	hourBucket := event.Timestamp.Truncate(time.Hour)

	stats := &HourlyStats{
		DeviceID:   event.DeviceID,
		HourBucket: hourBucket,
		EventCount: 1,
	}

	// Count by severity
	switch event.Severity {
	case "critical":
		stats.CriticalCount = 1
	case "high":
		stats.HighCount = 1
	case "medium":
		stats.MediumCount = 1
	case "low":
		stats.LowCount = 1
	}

	// Aggregate temperature
	if temp, ok := features["temperature"]; ok {
		stats.TempSum = temp
		stats.TempCount = 1
		stats.TempMin = temp
		stats.TempMax = temp
	}

	// Aggregate battery
	if battery, ok := features["battery_level"]; ok {
		stats.BatterySum = battery
		stats.BatteryCount = 1
		stats.BatteryMin = battery
		stats.BatteryMax = battery
	}

	// Aggregate energy
	if energy, ok := features["energy_kwh"]; ok {
		stats.EnergySum = energy
		stats.EnergyCount = 1
	}

	a.logger.Debug("Computed hourly stats",
		zap.String("device_id", event.DeviceID),
		zap.Time("hour_bucket", hourBucket),
		zap.Int64("event_count", stats.EventCount),
	)

	return stats
}

// MergeHourlyStats merges multiple HourlyStats for the same device/hour
func (a *Aggregator) MergeHourlyStats(ctx context.Context, stats []*HourlyStats) *HourlyStats {
	if len(stats) == 0 {
		return nil
	}

	merged := &HourlyStats{
		DeviceID:   stats[0].DeviceID,
		HourBucket: stats[0].HourBucket,
		TempMin:    stats[0].TempMin,
		TempMax:    stats[0].TempMax,
		BatteryMin: stats[0].BatteryMin,
		BatteryMax: stats[0].BatteryMax,
	}

	for _, s := range stats {
		merged.EventCount += s.EventCount
		merged.CriticalCount += s.CriticalCount
		merged.HighCount += s.HighCount
		merged.MediumCount += s.MediumCount
		merged.LowCount += s.LowCount

		// Temperature aggregation
		merged.TempSum += s.TempSum
		merged.TempCount += s.TempCount
		if s.TempMin < merged.TempMin || merged.TempMin == 0 {
			merged.TempMin = s.TempMin
		}
		if s.TempMax > merged.TempMax {
			merged.TempMax = s.TempMax
		}

		// Battery aggregation
		merged.BatterySum += s.BatterySum
		merged.BatteryCount += s.BatteryCount
		if s.BatteryMin < merged.BatteryMin || merged.BatteryMin == 0 {
			merged.BatteryMin = s.BatteryMin
		}
		if s.BatteryMax > merged.BatteryMax {
			merged.BatteryMax = s.BatteryMax
		}

		// Energy aggregation
		merged.EnergySum += s.EnergySum
		merged.EnergyCount += s.EnergyCount
	}

	a.logger.Debug("Merged hourly stats",
		zap.String("device_id", merged.DeviceID),
		zap.Time("hour_bucket", merged.HourBucket),
		zap.Int("stats_count", len(stats)),
		zap.Int64("total_events", merged.EventCount),
	)

	return merged
}

// ComputeDailyStats aggregates hourly stats into daily buckets
func (a *Aggregator) ComputeDailyStats(ctx context.Context, hourlyStats []*HourlyStats) *DailyStats {
	if len(hourlyStats) == 0 {
		return nil
	}

	dayBucket := hourlyStats[0].HourBucket.Truncate(24 * time.Hour)

	stats := &DailyStats{
		DeviceID:   hourlyStats[0].DeviceID,
		DayBucket:  dayBucket,
		TempMin:    hourlyStats[0].TempMin,
		TempMax:    hourlyStats[0].TempMax,
		BatteryMin: hourlyStats[0].BatteryMin,
	}

	var totalEvents int64
	var maxHourly int64
	var tempSum float64
	var tempCount int64
	var batterySum float64
	var batteryCount int64

	for _, hourly := range hourlyStats {
		totalEvents += hourly.EventCount
		stats.AlertCount += hourly.CriticalCount + hourly.HighCount

		if hourly.EventCount > maxHourly {
			maxHourly = hourly.EventCount
		}

		// Temperature
		tempSum += hourly.TempSum
		tempCount += hourly.TempCount
		if hourly.TempMin < stats.TempMin || stats.TempMin == 0 {
			stats.TempMin = hourly.TempMin
		}
		if hourly.TempMax > stats.TempMax {
			stats.TempMax = hourly.TempMax
		}

		// Battery
		batterySum += hourly.BatterySum
		batteryCount += hourly.BatteryCount
		if hourly.BatteryMin < stats.BatteryMin || stats.BatteryMin == 0 {
			stats.BatteryMin = hourly.BatteryMin
		}

		// Energy
		stats.EnergyTotal += hourly.EnergySum
	}

	stats.EventCount = totalEvents
	stats.MaxHourlyCount = maxHourly

	if len(hourlyStats) > 0 {
		stats.HourlyAvgCount = float64(totalEvents) / float64(len(hourlyStats))
	}

	if tempCount > 0 {
		stats.TempAvg = tempSum / float64(tempCount)
	}

	if batteryCount > 0 {
		stats.BatteryAvg = batterySum / float64(batteryCount)
	}

	// Estimate uptime (hours with events * 3600 seconds)
	stats.UptimeSeconds = int64(len(hourlyStats)) * 3600

	a.logger.Debug("Computed daily stats",
		zap.String("device_id", stats.DeviceID),
		zap.Time("day_bucket", dayBucket),
		zap.Int64("event_count", stats.EventCount),
		zap.Float64("temp_avg", stats.TempAvg),
	)

	return stats
}

// ExtractFeatures extracts numeric features from event metadata
func (a *Aggregator) ExtractFeatures(event *Event) map[string]float64 {
	features := make(map[string]float64)

	if event.Metadata == nil {
		return features
	}

	// Extract numeric values from metadata
	for key, value := range event.Metadata {
		switch v := value.(type) {
		case float64:
			features[key] = v
		case float32:
			features[key] = float64(v)
		case int:
			features[key] = float64(v)
		case int64:
			features[key] = float64(v)
		case int32:
			features[key] = float64(v)
		}
	}

	return features
}
