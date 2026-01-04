package features

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.uber.org/zap"
)

// DataPoint represents a time-series data point
type DataPoint struct {
	Timestamp time.Time
	Value     float64
}

// Calculator provides feature engineering utilities
type Calculator struct {
	logger *logger.Logger
}

// Config holds calculator configuration
type Config struct {
	Logger *logger.Logger
}

// NewCalculator creates a new feature calculator with dependency injection
func NewCalculator(cfg Config) *Calculator {
	return &Calculator{
		logger: cfg.Logger,
	}
}

// RollingStats computes rolling statistics over a time window
type RollingStats struct {
	Count  int64
	Sum    float64
	Mean   float64
	Min    float64
	Max    float64
	StdDev float64
}

// ComputeRollingAverage calculates rolling average for the given window
func (c *Calculator) ComputeRollingAverage(ctx context.Context, points []DataPoint, windowSize int) []DataPoint {
	if len(points) < windowSize || windowSize <= 0 {
		c.logger.Debug("Insufficient data for rolling average",
			zap.Int("points", len(points)),
			zap.Int("window_size", windowSize),
		)
		return []DataPoint{}
	}

	result := make([]DataPoint, 0, len(points)-windowSize+1)

	for i := windowSize - 1; i < len(points); i++ {
		var sum float64
		for j := i - windowSize + 1; j <= i; j++ {
			sum += points[j].Value
		}
		avg := sum / float64(windowSize)

		result = append(result, DataPoint{
			Timestamp: points[i].Timestamp,
			Value:     avg,
		})
	}

	c.logger.Debug("Computed rolling average",
		zap.Int("input_points", len(points)),
		zap.Int("output_points", len(result)),
		zap.Int("window_size", windowSize),
	)

	return result
}

// ComputeRollingStats calculates comprehensive rolling statistics
func (c *Calculator) ComputeRollingStats(ctx context.Context, points []DataPoint, windowSize int) *RollingStats {
	if len(points) < windowSize || windowSize <= 0 {
		return &RollingStats{}
	}

	// Use last window
	window := points[len(points)-windowSize:]

	stats := &RollingStats{
		Count: int64(windowSize),
		Min:   window[0].Value,
		Max:   window[0].Value,
	}

	var sum, sumSq float64

	for _, point := range window {
		val := point.Value
		sum += val
		sumSq += val * val

		if val < stats.Min {
			stats.Min = val
		}
		if val > stats.Max {
			stats.Max = val
		}
	}

	stats.Sum = sum
	stats.Mean = sum / float64(windowSize)

	// Standard deviation: sqrt(E[X^2] - E[X]^2)
	variance := (sumSq / float64(windowSize)) - (stats.Mean * stats.Mean)
	if variance > 0 {
		stats.StdDev = math.Sqrt(variance)
	}

	c.logger.Debug("Computed rolling stats",
		zap.Int("window_size", windowSize),
		zap.Float64("mean", stats.Mean),
		zap.Float64("stddev", stats.StdDev),
	)

	return stats
}

// ComputePercentile calculates the specified percentile (0-100)
func (c *Calculator) ComputePercentile(ctx context.Context, points []DataPoint, percentile float64) float64 {
	if len(points) == 0 {
		c.logger.Warn("No data points for percentile calculation")
		return 0
	}

	if percentile < 0 || percentile > 100 {
		c.logger.Warn("Invalid percentile value",
			zap.Float64("percentile", percentile),
		)
		return 0
	}

	// Extract values and sort
	values := make([]float64, len(points))
	for i, p := range points {
		values[i] = p.Value
	}
	sort.Float64s(values)

	// Calculate percentile index
	rank := (percentile / 100.0) * float64(len(values)-1)
	lowerIndex := int(math.Floor(rank))
	upperIndex := int(math.Ceil(rank))

	if lowerIndex == upperIndex {
		return values[lowerIndex]
	}

	// Linear interpolation
	fraction := rank - float64(lowerIndex)
	result := values[lowerIndex] + fraction*(values[upperIndex]-values[lowerIndex])

	c.logger.Debug("Computed percentile",
		zap.Float64("percentile", percentile),
		zap.Float64("result", result),
		zap.Int("data_points", len(points)),
	)

	return result
}

// ComputePercentiles calculates P50, P95, P99 in one pass
func (c *Calculator) ComputePercentiles(ctx context.Context, points []DataPoint) map[string]float64 {
	if len(points) == 0 {
		return map[string]float64{
			"p50": 0,
			"p95": 0,
			"p99": 0,
		}
	}

	values := make([]float64, len(points))
	for i, p := range points {
		values[i] = p.Value
	}
	sort.Float64s(values)

	percentiles := map[string]float64{
		"p50": c.calculatePercentileFromSorted(values, 50),
		"p95": c.calculatePercentileFromSorted(values, 95),
		"p99": c.calculatePercentileFromSorted(values, 99),
	}

	c.logger.Debug("Computed percentiles",
		zap.Float64("p50", percentiles["p50"]),
		zap.Float64("p95", percentiles["p95"]),
		zap.Float64("p99", percentiles["p99"]),
	)

	return percentiles
}

func (c *Calculator) calculatePercentileFromSorted(sorted []float64, percentile float64) float64 {
	rank := (percentile / 100.0) * float64(len(sorted)-1)
	lowerIndex := int(math.Floor(rank))
	upperIndex := int(math.Ceil(rank))

	if lowerIndex == upperIndex {
		return sorted[lowerIndex]
	}

	fraction := rank - float64(lowerIndex)
	return sorted[lowerIndex] + fraction*(sorted[upperIndex]-sorted[lowerIndex])
}

// ComputeRateOfChange calculates the rate of change between consecutive points
func (c *Calculator) ComputeRateOfChange(ctx context.Context, points []DataPoint) []DataPoint {
	if len(points) < 2 {
		return []DataPoint{}
	}

	result := make([]DataPoint, 0, len(points)-1)

	for i := 1; i < len(points); i++ {
		timeDiff := points[i].Timestamp.Sub(points[i-1].Timestamp).Seconds()
		if timeDiff == 0 {
			continue
		}

		valueDiff := points[i].Value - points[i-1].Value
		rate := valueDiff / timeDiff

		result = append(result, DataPoint{
			Timestamp: points[i].Timestamp,
			Value:     rate,
		})
	}

	c.logger.Debug("Computed rate of change",
		zap.Int("input_points", len(points)),
		zap.Int("output_points", len(result)),
	)

	return result
}

// ComputeDelta calculates the absolute change between consecutive points
func (c *Calculator) ComputeDelta(ctx context.Context, points []DataPoint) []DataPoint {
	if len(points) < 2 {
		return []DataPoint{}
	}

	result := make([]DataPoint, 0, len(points)-1)

	for i := 1; i < len(points); i++ {
		delta := points[i].Value - points[i-1].Value

		result = append(result, DataPoint{
			Timestamp: points[i].Timestamp,
			Value:     delta,
		})
	}

	c.logger.Debug("Computed delta",
		zap.Int("input_points", len(points)),
		zap.Int("output_points", len(result)),
	)

	return result
}

// TimeBasedFeatures extracts time-based features from a timestamp
type TimeBasedFeatures struct {
	HourOfDay     int
	DayOfWeek     int
	DayOfMonth    int
	Month         int
	IsWeekend     bool
	IsBusinessDay bool
}

// ExtractTimeFeatures extracts time-based features for ML models
func (c *Calculator) ExtractTimeFeatures(ctx context.Context, t time.Time) *TimeBasedFeatures {
	features := &TimeBasedFeatures{
		HourOfDay:  t.Hour(),
		DayOfWeek:  int(t.Weekday()),
		DayOfMonth: t.Day(),
		Month:      int(t.Month()),
	}

	// Weekend detection (Saturday=6, Sunday=0)
	features.IsWeekend = features.DayOfWeek == 0 || features.DayOfWeek == 6

	// Business day (Monday-Friday, excluding weekends)
	features.IsBusinessDay = features.DayOfWeek >= 1 && features.DayOfWeek <= 5

	c.logger.Debug("Extracted time features",
		zap.Int("hour", features.HourOfDay),
		zap.Int("day_of_week", features.DayOfWeek),
		zap.Bool("is_weekend", features.IsWeekend),
	)

	return features
}

// ComputeExponentialMovingAverage calculates EMA with the given alpha (smoothing factor)
func (c *Calculator) ComputeExponentialMovingAverage(ctx context.Context, points []DataPoint, alpha float64) []DataPoint {
	if len(points) == 0 {
		return []DataPoint{}
	}

	if alpha <= 0 || alpha > 1 {
		c.logger.Warn("Invalid alpha value, using default 0.3",
			zap.Float64("alpha", alpha),
		)
		alpha = 0.3
	}

	result := make([]DataPoint, len(points))
	result[0] = points[0]

	for i := 1; i < len(points); i++ {
		ema := alpha*points[i].Value + (1-alpha)*result[i-1].Value
		result[i] = DataPoint{
			Timestamp: points[i].Timestamp,
			Value:     ema,
		}
	}

	c.logger.Debug("Computed exponential moving average",
		zap.Int("points", len(points)),
		zap.Float64("alpha", alpha),
	)

	return result
}

// DetectOutliers identifies outliers using IQR method
func (c *Calculator) DetectOutliers(ctx context.Context, points []DataPoint, threshold float64) []DataPoint {
	if len(points) < 4 {
		return []DataPoint{}
	}

	// Calculate Q1, Q3
	q1 := c.ComputePercentile(ctx, points, 25)
	q3 := c.ComputePercentile(ctx, points, 75)
	iqr := q3 - q1

	lowerBound := q1 - threshold*iqr
	upperBound := q3 + threshold*iqr

	outliers := make([]DataPoint, 0)
	for _, point := range points {
		if point.Value < lowerBound || point.Value > upperBound {
			outliers = append(outliers, point)
		}
	}

	c.logger.Debug("Detected outliers",
		zap.Int("total_points", len(points)),
		zap.Int("outliers", len(outliers)),
		zap.Float64("lower_bound", lowerBound),
		zap.Float64("upper_bound", upperBound),
	)

	return outliers
}
