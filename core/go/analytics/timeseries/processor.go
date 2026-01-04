package timeseries

import (
	"context"
	"math"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.uber.org/zap"
)

// DataPoint represents a time-series data point
type DataPoint struct {
	Timestamp time.Time
	Value     float64
}

// Window represents a time window
type Window struct {
	Start  time.Time
	End    time.Time
	Points []DataPoint
}

// WindowType defines the type of windowing strategy
type WindowType string

const (
	WindowTypeTumbling WindowType = "tumbling"
	WindowTypeSliding  WindowType = "sliding"
	WindowTypeSession  WindowType = "session"
)

// Processor provides time-series processing capabilities
type Processor struct {
	logger *logger.Logger
}

// Config holds processor configuration
type Config struct {
	Logger *logger.Logger
}

// NewProcessor creates a new time-series processor with dependency injection
func NewProcessor(cfg Config) *Processor {
	return &Processor{
		logger: cfg.Logger,
	}
}

// CreateTumblingWindows creates non-overlapping time windows
func (p *Processor) CreateTumblingWindows(ctx context.Context, points []DataPoint, windowSize time.Duration) []Window {
	if len(points) == 0 {
		return []Window{}
	}

	windows := []Window{}

	// Find first window start (aligned to window size)
	firstTimestamp := points[0].Timestamp
	windowStart := firstTimestamp.Truncate(windowSize)

	currentWindow := Window{
		Start:  windowStart,
		End:    windowStart.Add(windowSize),
		Points: []DataPoint{},
	}

	for _, point := range points {
		// Check if point belongs to current window
		if point.Timestamp.Before(currentWindow.End) {
			currentWindow.Points = append(currentWindow.Points, point)
		} else {
			// Save current window and start new one
			if len(currentWindow.Points) > 0 {
				windows = append(windows, currentWindow)
			}

			// Calculate new window boundaries
			windowStart = point.Timestamp.Truncate(windowSize)
			currentWindow = Window{
				Start:  windowStart,
				End:    windowStart.Add(windowSize),
				Points: []DataPoint{point},
			}
		}
	}

	// Add last window
	if len(currentWindow.Points) > 0 {
		windows = append(windows, currentWindow)
	}

	p.logger.Debug("Created tumbling windows",
		zap.Int("input_points", len(points)),
		zap.Int("windows", len(windows)),
		zap.Duration("window_size", windowSize),
	)

	return windows
}

// CreateSlidingWindows creates overlapping time windows
func (p *Processor) CreateSlidingWindows(ctx context.Context, points []DataPoint, windowSize, slideInterval time.Duration) []Window {
	if len(points) == 0 {
		return []Window{}
	}

	windows := []Window{}

	firstTimestamp := points[0].Timestamp
	lastTimestamp := points[len(points)-1].Timestamp

	// Create windows at slide intervals
	for windowStart := firstTimestamp.Truncate(slideInterval); windowStart.Before(lastTimestamp); windowStart = windowStart.Add(slideInterval) {
		windowEnd := windowStart.Add(windowSize)

		window := Window{
			Start:  windowStart,
			End:    windowEnd,
			Points: []DataPoint{},
		}

		// Add points that fall within this window
		for _, point := range points {
			if !point.Timestamp.Before(windowStart) && point.Timestamp.Before(windowEnd) {
				window.Points = append(window.Points, point)
			}
		}

		if len(window.Points) > 0 {
			windows = append(windows, window)
		}
	}

	p.logger.Debug("Created sliding windows",
		zap.Int("input_points", len(points)),
		zap.Int("windows", len(windows)),
		zap.Duration("window_size", windowSize),
		zap.Duration("slide_interval", slideInterval),
	)

	return windows
}

// CreateSessionWindows creates windows based on gaps in activity
func (p *Processor) CreateSessionWindows(ctx context.Context, points []DataPoint, gapTimeout time.Duration) []Window {
	if len(points) == 0 {
		return []Window{}
	}

	windows := []Window{}

	currentWindow := Window{
		Start:  points[0].Timestamp,
		Points: []DataPoint{points[0]},
	}

	for i := 1; i < len(points); i++ {
		gap := points[i].Timestamp.Sub(points[i-1].Timestamp)

		if gap <= gapTimeout {
			// Continue current session
			currentWindow.Points = append(currentWindow.Points, points[i])
		} else {
			// End current session and start new one
			currentWindow.End = points[i-1].Timestamp
			windows = append(windows, currentWindow)

			currentWindow = Window{
				Start:  points[i].Timestamp,
				Points: []DataPoint{points[i]},
			}
		}
	}

	// Add last window
	currentWindow.End = points[len(points)-1].Timestamp
	windows = append(windows, currentWindow)

	p.logger.Debug("Created session windows",
		zap.Int("input_points", len(points)),
		zap.Int("windows", len(windows)),
		zap.Duration("gap_timeout", gapTimeout),
	)

	return windows
}

// BucketByTime groups data points into time buckets
func (p *Processor) BucketByTime(ctx context.Context, points []DataPoint, bucketSize time.Duration) map[time.Time][]DataPoint {
	buckets := make(map[time.Time][]DataPoint)

	for _, point := range points {
		bucketTime := point.Timestamp.Truncate(bucketSize)
		buckets[bucketTime] = append(buckets[bucketTime], point)
	}

	p.logger.Debug("Bucketed data points by time",
		zap.Int("input_points", len(points)),
		zap.Int("buckets", len(buckets)),
		zap.Duration("bucket_size", bucketSize),
	)

	return buckets
}

// InterpolationType defines the interpolation method
type InterpolationType string

const (
	InterpolationLinear   InterpolationType = "linear"
	InterpolationNearest  InterpolationType = "nearest"
	InterpolationForward  InterpolationType = "forward"
	InterpolationBackward InterpolationType = "backward"
)

// InterpolateMissing fills missing values in time-series data
func (p *Processor) InterpolateMissing(ctx context.Context, points []DataPoint, expectedInterval time.Duration, method InterpolationType) []DataPoint {
	if len(points) < 2 {
		return points
	}

	result := []DataPoint{points[0]}

	for i := 1; i < len(points); i++ {
		prev := points[i-1]
		curr := points[i]

		gap := curr.Timestamp.Sub(prev.Timestamp)

		// Check if there's a gap that needs filling
		if gap > expectedInterval {
			missingCount := int(gap / expectedInterval)

			for j := 1; j < missingCount; j++ {
				missingTime := prev.Timestamp.Add(expectedInterval * time.Duration(j))

				var interpolatedValue float64
				switch method {
				case InterpolationLinear:
					// Linear interpolation
					fraction := float64(j) / float64(missingCount)
					interpolatedValue = prev.Value + fraction*(curr.Value-prev.Value)

				case InterpolationNearest:
					// Use nearest neighbor
					if j < missingCount/2 {
						interpolatedValue = prev.Value
					} else {
						interpolatedValue = curr.Value
					}

				case InterpolationForward:
					// Forward fill
					interpolatedValue = prev.Value

				case InterpolationBackward:
					// Backward fill
					interpolatedValue = curr.Value
				}

				result = append(result, DataPoint{
					Timestamp: missingTime,
					Value:     interpolatedValue,
				})
			}
		}

		result = append(result, curr)
	}

	p.logger.Debug("Interpolated missing values",
		zap.Int("input_points", len(points)),
		zap.Int("output_points", len(result)),
		zap.String("method", string(method)),
	)

	return result
}

// Resample resamples time-series data to a different frequency
func (p *Processor) Resample(ctx context.Context, points []DataPoint, newInterval time.Duration, aggregation string) []DataPoint {
	if len(points) == 0 {
		return []DataPoint{}
	}

	// First, bucket by new interval
	buckets := p.BucketByTime(ctx, points, newInterval)

	// Then aggregate each bucket
	result := make([]DataPoint, 0, len(buckets))

	for bucketTime, bucketPoints := range buckets {
		var aggregatedValue float64

		switch aggregation {
		case "mean", "avg":
			sum := 0.0
			for _, pt := range bucketPoints {
				sum += pt.Value
			}
			aggregatedValue = sum / float64(len(bucketPoints))

		case "sum":
			for _, pt := range bucketPoints {
				aggregatedValue += pt.Value
			}

		case "min":
			aggregatedValue = bucketPoints[0].Value
			for _, pt := range bucketPoints {
				if pt.Value < aggregatedValue {
					aggregatedValue = pt.Value
				}
			}

		case "max":
			aggregatedValue = bucketPoints[0].Value
			for _, pt := range bucketPoints {
				if pt.Value > aggregatedValue {
					aggregatedValue = pt.Value
				}
			}

		case "first":
			aggregatedValue = bucketPoints[0].Value

		case "last":
			aggregatedValue = bucketPoints[len(bucketPoints)-1].Value

		default:
			// Default to mean
			sum := 0.0
			for _, pt := range bucketPoints {
				sum += pt.Value
			}
			aggregatedValue = sum / float64(len(bucketPoints))
		}

		result = append(result, DataPoint{
			Timestamp: bucketTime,
			Value:     aggregatedValue,
		})
	}

	p.logger.Debug("Resampled time-series data",
		zap.Int("input_points", len(points)),
		zap.Int("output_points", len(result)),
		zap.Duration("new_interval", newInterval),
		zap.String("aggregation", aggregation),
	)

	return result
}

// Downsample reduces the number of points while preserving shape
func (p *Processor) Downsample(ctx context.Context, points []DataPoint, targetCount int) []DataPoint {
	if len(points) <= targetCount {
		return points
	}

	// Calculate step size
	step := float64(len(points)) / float64(targetCount)

	result := make([]DataPoint, targetCount)

	for i := 0; i < targetCount; i++ {
		index := int(math.Round(float64(i) * step))
		if index >= len(points) {
			index = len(points) - 1
		}
		result[i] = points[index]
	}

	p.logger.Debug("Downsampled time-series data",
		zap.Int("input_points", len(points)),
		zap.Int("output_points", len(result)),
		zap.Int("target_count", targetCount),
	)

	return result
}

// AlignTimestamps aligns timestamps to a specific interval
func (p *Processor) AlignTimestamps(ctx context.Context, points []DataPoint, interval time.Duration) []DataPoint {
	result := make([]DataPoint, len(points))

	for i, point := range points {
		alignedTime := point.Timestamp.Truncate(interval)
		result[i] = DataPoint{
			Timestamp: alignedTime,
			Value:     point.Value,
		}
	}

	p.logger.Debug("Aligned timestamps",
		zap.Int("points", len(points)),
		zap.Duration("interval", interval),
	)

	return result
}

// CalculateWindowStatistics computes statistics for each window
func (p *Processor) CalculateWindowStatistics(ctx context.Context, windows []Window) []map[string]float64 {
	stats := make([]map[string]float64, len(windows))

	for i, window := range windows {
		if len(window.Points) == 0 {
			stats[i] = map[string]float64{}
			continue
		}

		sum := 0.0
		min := window.Points[0].Value
		max := window.Points[0].Value

		for _, point := range window.Points {
			sum += point.Value
			if point.Value < min {
				min = point.Value
			}
			if point.Value > max {
				max = point.Value
			}
		}

		mean := sum / float64(len(window.Points))

		stats[i] = map[string]float64{
			"count": float64(len(window.Points)),
			"sum":   sum,
			"mean":  mean,
			"min":   min,
			"max":   max,
		}
	}

	p.logger.Debug("Calculated window statistics",
		zap.Int("windows", len(windows)),
	)

	return stats
}
