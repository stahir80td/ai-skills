package features

import (
	"context"
	"testing"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
)

func TestComputeRollingAverage(t *testing.T) {
	log, _ := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})

	calc := NewCalculator(Config{Logger: log})

	now := time.Now()
	points := []DataPoint{
		{Timestamp: now, Value: 10},
		{Timestamp: now.Add(time.Minute), Value: 20},
		{Timestamp: now.Add(2 * time.Minute), Value: 30},
		{Timestamp: now.Add(3 * time.Minute), Value: 40},
		{Timestamp: now.Add(4 * time.Minute), Value: 50},
	}

	result := calc.ComputeRollingAverage(context.Background(), points, 3)

	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}

	// First window: (10+20+30)/3 = 20
	if result[0].Value != 20.0 {
		t.Errorf("Expected first average 20.0, got %f", result[0].Value)
	}

	// Second window: (20+30+40)/3 = 30
	if result[1].Value != 30.0 {
		t.Errorf("Expected second average 30.0, got %f", result[1].Value)
	}

	// Third window: (30+40+50)/3 = 40
	if result[2].Value != 40.0 {
		t.Errorf("Expected third average 40.0, got %f", result[2].Value)
	}
}

func TestComputeRollingStats(t *testing.T) {
	log, _ := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})

	calc := NewCalculator(Config{Logger: log})

	now := time.Now()
	points := []DataPoint{
		{Timestamp: now, Value: 10},
		{Timestamp: now.Add(time.Minute), Value: 20},
		{Timestamp: now.Add(2 * time.Minute), Value: 30},
		{Timestamp: now.Add(3 * time.Minute), Value: 40},
		{Timestamp: now.Add(4 * time.Minute), Value: 50},
	}

	stats := calc.ComputeRollingStats(context.Background(), points, 5)

	if stats.Count != 5 {
		t.Errorf("Expected count 5, got %d", stats.Count)
	}

	if stats.Sum != 150.0 {
		t.Errorf("Expected sum 150.0, got %f", stats.Sum)
	}

	if stats.Mean != 30.0 {
		t.Errorf("Expected mean 30.0, got %f", stats.Mean)
	}

	if stats.Min != 10.0 {
		t.Errorf("Expected min 10.0, got %f", stats.Min)
	}

	if stats.Max != 50.0 {
		t.Errorf("Expected max 50.0, got %f", stats.Max)
	}
}

func TestComputePercentile(t *testing.T) {
	log, _ := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})

	calc := NewCalculator(Config{Logger: log})

	now := time.Now()
	points := []DataPoint{
		{Timestamp: now, Value: 10},
		{Timestamp: now.Add(time.Minute), Value: 20},
		{Timestamp: now.Add(2 * time.Minute), Value: 30},
		{Timestamp: now.Add(3 * time.Minute), Value: 40},
		{Timestamp: now.Add(4 * time.Minute), Value: 50},
		{Timestamp: now.Add(5 * time.Minute), Value: 60},
		{Timestamp: now.Add(6 * time.Minute), Value: 70},
		{Timestamp: now.Add(7 * time.Minute), Value: 80},
		{Timestamp: now.Add(8 * time.Minute), Value: 90},
		{Timestamp: now.Add(9 * time.Minute), Value: 100},
	}

	p50 := calc.ComputePercentile(context.Background(), points, 50)
	if p50 != 55.0 {
		t.Errorf("Expected P50 55.0, got %f", p50)
	}

	p95 := calc.ComputePercentile(context.Background(), points, 95)
	if p95 < 95.49 || p95 > 95.51 { // Allow for floating point precision
		t.Errorf("Expected P95 ~95.5, got %f", p95)
	}
}

func TestComputePercentiles(t *testing.T) {
	log, _ := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})

	calc := NewCalculator(Config{Logger: log})

	now := time.Now()
	points := make([]DataPoint, 100)
	for i := 0; i < 100; i++ {
		points[i] = DataPoint{
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			Value:     float64(i + 1),
		}
	}

	percentiles := calc.ComputePercentiles(context.Background(), points)

	if percentiles["p50"] != 50.5 {
		t.Errorf("Expected P50 50.5, got %f", percentiles["p50"])
	}

	if percentiles["p95"] != 95.05 {
		t.Errorf("Expected P95 95.05, got %f", percentiles["p95"])
	}

	if percentiles["p99"] != 99.01 {
		t.Errorf("Expected P99 99.01, got %f", percentiles["p99"])
	}
}

func TestComputeRateOfChange(t *testing.T) {
	log, _ := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})

	calc := NewCalculator(Config{Logger: log})

	now := time.Now()
	points := []DataPoint{
		{Timestamp: now, Value: 0},
		{Timestamp: now.Add(10 * time.Second), Value: 100}, // +100 in 10s = 10/s
		{Timestamp: now.Add(20 * time.Second), Value: 150}, // +50 in 10s = 5/s
		{Timestamp: now.Add(30 * time.Second), Value: 250}, // +100 in 10s = 10/s
	}

	result := calc.ComputeRateOfChange(context.Background(), points)

	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}

	if result[0].Value != 10.0 {
		t.Errorf("Expected rate 10.0/s, got %f", result[0].Value)
	}

	if result[1].Value != 5.0 {
		t.Errorf("Expected rate 5.0/s, got %f", result[1].Value)
	}

	if result[2].Value != 10.0 {
		t.Errorf("Expected rate 10.0/s, got %f", result[2].Value)
	}
}

func TestComputeDelta(t *testing.T) {
	log, _ := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})

	calc := NewCalculator(Config{Logger: log})

	now := time.Now()
	points := []DataPoint{
		{Timestamp: now, Value: 10},
		{Timestamp: now.Add(time.Minute), Value: 15},
		{Timestamp: now.Add(2 * time.Minute), Value: 12},
		{Timestamp: now.Add(3 * time.Minute), Value: 20},
	}

	result := calc.ComputeDelta(context.Background(), points)

	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}

	if result[0].Value != 5.0 {
		t.Errorf("Expected delta 5.0, got %f", result[0].Value)
	}

	if result[1].Value != -3.0 {
		t.Errorf("Expected delta -3.0, got %f", result[1].Value)
	}

	if result[2].Value != 8.0 {
		t.Errorf("Expected delta 8.0, got %f", result[2].Value)
	}
}

func TestExtractTimeFeatures(t *testing.T) {
	log, _ := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})

	calc := NewCalculator(Config{Logger: log})

	// Wednesday, 2025-12-17 14:30:00
	testTime := time.Date(2025, 12, 17, 14, 30, 0, 0, time.UTC)

	features := calc.ExtractTimeFeatures(context.Background(), testTime)

	if features.HourOfDay != 14 {
		t.Errorf("Expected hour 14, got %d", features.HourOfDay)
	}

	if features.DayOfWeek != 3 { // Wednesday
		t.Errorf("Expected day_of_week 3, got %d", features.DayOfWeek)
	}

	if features.DayOfMonth != 17 {
		t.Errorf("Expected day_of_month 17, got %d", features.DayOfMonth)
	}

	if features.Month != 12 {
		t.Errorf("Expected month 12, got %d", features.Month)
	}

	if features.IsWeekend {
		t.Error("Expected is_weekend to be false for Wednesday")
	}

	if !features.IsBusinessDay {
		t.Error("Expected is_business_day to be true for Wednesday")
	}

	// Test weekend
	saturday := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	weekendFeatures := calc.ExtractTimeFeatures(context.Background(), saturday)

	if !weekendFeatures.IsWeekend {
		t.Error("Expected is_weekend to be true for Saturday")
	}

	if weekendFeatures.IsBusinessDay {
		t.Error("Expected is_business_day to be false for Saturday")
	}
}

func TestComputeExponentialMovingAverage(t *testing.T) {
	log, _ := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})

	calc := NewCalculator(Config{Logger: log})

	now := time.Now()
	points := []DataPoint{
		{Timestamp: now, Value: 10},
		{Timestamp: now.Add(time.Minute), Value: 20},
		{Timestamp: now.Add(2 * time.Minute), Value: 30},
	}

	alpha := 0.5
	result := calc.ComputeExponentialMovingAverage(context.Background(), points, alpha)

	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}

	// First value stays the same
	if result[0].Value != 10.0 {
		t.Errorf("Expected first EMA 10.0, got %f", result[0].Value)
	}

	// Second: 0.5*20 + 0.5*10 = 15
	if result[1].Value != 15.0 {
		t.Errorf("Expected second EMA 15.0, got %f", result[1].Value)
	}

	// Third: 0.5*30 + 0.5*15 = 22.5
	if result[2].Value != 22.5 {
		t.Errorf("Expected third EMA 22.5, got %f", result[2].Value)
	}
}

func TestDetectOutliers(t *testing.T) {
	log, _ := logger.New(logger.Config{
		ServiceName: "test",
		Environment: "test",
		LogLevel:    "debug",
	})

	calc := NewCalculator(Config{Logger: log})

	now := time.Now()
	points := []DataPoint{
		{Timestamp: now, Value: 10},
		{Timestamp: now.Add(time.Minute), Value: 12},
		{Timestamp: now.Add(2 * time.Minute), Value: 11},
		{Timestamp: now.Add(3 * time.Minute), Value: 13},
		{Timestamp: now.Add(4 * time.Minute), Value: 100}, // Outlier
		{Timestamp: now.Add(5 * time.Minute), Value: 12},
		{Timestamp: now.Add(6 * time.Minute), Value: 11},
		{Timestamp: now.Add(7 * time.Minute), Value: -50}, // Outlier
	}

	outliers := calc.DetectOutliers(context.Background(), points, 1.5)

	if len(outliers) < 2 {
		t.Errorf("Expected at least 2 outliers, got %d", len(outliers))
	}

	// Check if 100 and -50 are detected as outliers
	hasHighOutlier := false
	hasLowOutlier := false

	for _, o := range outliers {
		if o.Value == 100 {
			hasHighOutlier = true
		}
		if o.Value == -50 {
			hasLowOutlier = true
		}
	}

	if !hasHighOutlier {
		t.Error("Expected to detect 100 as outlier")
	}

	if !hasLowOutlier {
		t.Error("Expected to detect -50 as outlier")
	}
}
