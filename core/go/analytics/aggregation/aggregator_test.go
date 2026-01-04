package aggregation

import (
	"context"
	"testing"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
)

func TestComputeHourlyStats(t *testing.T) {
	// Setup logger
	log, err := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	agg := NewAggregator(Config{Logger: log})

	// Test event
	now := time.Now()
	event := &Event{
		DeviceID:  "device-1",
		EventType: "temperature",
		Severity:  "high",
		Timestamp: now,
		Metadata: map[string]interface{}{
			"temperature":   25.5,
			"battery_level": 85.0,
			"energy_kwh":    0.5,
		},
	}

	features := agg.ExtractFeatures(event)
	stats := agg.ComputeHourlyStats(context.Background(), event, features)

	// Assertions
	if stats.DeviceID != "device-1" {
		t.Errorf("Expected device_id device-1, got %s", stats.DeviceID)
	}

	if stats.EventCount != 1 {
		t.Errorf("Expected event_count 1, got %d", stats.EventCount)
	}

	if stats.HighCount != 1 {
		t.Errorf("Expected high_count 1, got %d", stats.HighCount)
	}

	if stats.TempSum != 25.5 {
		t.Errorf("Expected temp_sum 25.5, got %f", stats.TempSum)
	}

	if stats.BatterySum != 85.0 {
		t.Errorf("Expected battery_sum 85.0, got %f", stats.BatterySum)
	}

	if stats.EnergySum != 0.5 {
		t.Errorf("Expected energy_sum 0.5, got %f", stats.EnergySum)
	}

	expectedHour := now.Truncate(time.Hour)
	if !stats.HourBucket.Equal(expectedHour) {
		t.Errorf("Expected hour_bucket %v, got %v", expectedHour, stats.HourBucket)
	}
}

func TestMergeHourlyStats(t *testing.T) {
	log, err := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	agg := NewAggregator(Config{Logger: log})

	hourBucket := time.Now().Truncate(time.Hour)

	stats1 := &HourlyStats{
		DeviceID:      "device-1",
		HourBucket:    hourBucket,
		EventCount:    10,
		CriticalCount: 2,
		TempSum:       250.0,
		TempCount:     10,
		TempMin:       20.0,
		TempMax:       30.0,
		BatterySum:    850.0,
		BatteryCount:  10,
		BatteryMin:    80.0,
		BatteryMax:    90.0,
	}

	stats2 := &HourlyStats{
		DeviceID:     "device-1",
		HourBucket:   hourBucket,
		EventCount:   5,
		HighCount:    1,
		TempSum:      125.0,
		TempCount:    5,
		TempMin:      19.0, // Lower than stats1
		TempMax:      31.0, // Higher than stats1
		BatterySum:   425.0,
		BatteryCount: 5,
		BatteryMin:   82.0,
		BatteryMax:   88.0,
	}

	merged := agg.MergeHourlyStats(context.Background(), []*HourlyStats{stats1, stats2})

	// Assertions
	if merged.EventCount != 15 {
		t.Errorf("Expected event_count 15, got %d", merged.EventCount)
	}

	if merged.CriticalCount != 2 {
		t.Errorf("Expected critical_count 2, got %d", merged.CriticalCount)
	}

	if merged.HighCount != 1 {
		t.Errorf("Expected high_count 1, got %d", merged.HighCount)
	}

	if merged.TempSum != 375.0 {
		t.Errorf("Expected temp_sum 375.0, got %f", merged.TempSum)
	}

	if merged.TempMin != 19.0 {
		t.Errorf("Expected temp_min 19.0, got %f", merged.TempMin)
	}

	if merged.TempMax != 31.0 {
		t.Errorf("Expected temp_max 31.0, got %f", merged.TempMax)
	}
}

func TestComputeDailyStats(t *testing.T) {
	log, err := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	agg := NewAggregator(Config{Logger: log})

	now := time.Now()
	dayStart := now.Truncate(24 * time.Hour)

	hourlyStats := []*HourlyStats{
		{
			DeviceID:      "device-1",
			HourBucket:    dayStart,
			EventCount:    100,
			CriticalCount: 5,
			HighCount:     10,
			TempSum:       2500.0,
			TempCount:     100,
			TempMin:       20.0,
			TempMax:       25.0,
			BatterySum:    8500.0,
			BatteryCount:  100,
			BatteryMin:    80.0,
			EnergySum:     10.0,
		},
		{
			DeviceID:      "device-1",
			HourBucket:    dayStart.Add(time.Hour),
			EventCount:    80,
			CriticalCount: 2,
			HighCount:     5,
			TempSum:       2000.0,
			TempCount:     80,
			TempMin:       19.0, // Lower
			TempMax:       26.0, // Higher
			BatterySum:    6800.0,
			BatteryCount:  80,
			BatteryMin:    79.0, // Lower
			EnergySum:     8.0,
		},
	}

	daily := agg.ComputeDailyStats(context.Background(), hourlyStats)

	// Assertions
	if daily.DeviceID != "device-1" {
		t.Errorf("Expected device_id device-1, got %s", daily.DeviceID)
	}

	if daily.EventCount != 180 {
		t.Errorf("Expected event_count 180, got %d", daily.EventCount)
	}

	if daily.AlertCount != 22 { // 5+10+2+5
		t.Errorf("Expected alert_count 22, got %d", daily.AlertCount)
	}

	if daily.MaxHourlyCount != 100 {
		t.Errorf("Expected max_hourly_count 100, got %d", daily.MaxHourlyCount)
	}

	expectedAvg := 180.0 / 2.0
	if daily.HourlyAvgCount != expectedAvg {
		t.Errorf("Expected hourly_avg_count %f, got %f", expectedAvg, daily.HourlyAvgCount)
	}

	if daily.TempMin != 19.0 {
		t.Errorf("Expected temp_min 19.0, got %f", daily.TempMin)
	}

	if daily.TempMax != 26.0 {
		t.Errorf("Expected temp_max 26.0, got %f", daily.TempMax)
	}

	expectedTempAvg := 4500.0 / 180.0
	if daily.TempAvg != expectedTempAvg {
		t.Errorf("Expected temp_avg %f, got %f", expectedTempAvg, daily.TempAvg)
	}

	if daily.EnergyTotal != 18.0 {
		t.Errorf("Expected energy_total 18.0, got %f", daily.EnergyTotal)
	}

	expectedUptime := int64(2 * 3600)
	if daily.UptimeSeconds != expectedUptime {
		t.Errorf("Expected uptime_seconds %d, got %d", expectedUptime, daily.UptimeSeconds)
	}
}

func TestExtractFeatures(t *testing.T) {
	log, err := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	agg := NewAggregator(Config{Logger: log})

	event := &Event{
		DeviceID:  "device-1",
		EventType: "sensor",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"temperature":   25.5,
			"humidity":      65,
			"battery_level": float32(85.5),
			"status":        "active", // Non-numeric, should be skipped
		},
	}

	features := agg.ExtractFeatures(event)

	if len(features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(features))
	}

	if features["temperature"] != 25.5 {
		t.Errorf("Expected temperature 25.5, got %f", features["temperature"])
	}

	if features["humidity"] != 65.0 {
		t.Errorf("Expected humidity 65.0, got %f", features["humidity"])
	}

	if features["battery_level"] != 85.5 {
		t.Errorf("Expected battery_level 85.5, got %f", features["battery_level"])
	}

	if _, exists := features["status"]; exists {
		t.Error("Non-numeric status should not be in features")
	}
}
