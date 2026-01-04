package models

import (
	"time"

	"github.com/google/uuid"
)

// DeviceTelemetry for ScyllaDB storage - demonstrates time-series data
type DeviceTelemetry struct {
	DeviceID      string            `json:"deviceId"`
	Timestamp     time.Time         `json:"timestamp"`
	Metric        string            `json:"metric"`
	Value         float64           `json:"value"`
	Unit          string            `json:"unit"`
	Tags          map[string]string `json:"tags,omitempty"`
	Location      string            `json:"location,omitempty"`
	DataType      string            `json:"dataType"` // temperature, pressure, vibration, etc.
	MinValue      *float64          `json:"minValue,omitempty"`
	MaxValue      *float64          `json:"maxValue,omitempty"`
	Quality       string            `json:"quality"` // good, bad, uncertain
	CorrelationID uuid.UUID         `json:"correlationId"`
}

// NewDeviceTelemetry creates a new telemetry record
func NewDeviceTelemetry(deviceID, metric string, value float64, unit string) *DeviceTelemetry {
	return &DeviceTelemetry{
		DeviceID:      deviceID,
		Timestamp:     time.Now().UTC(),
		Metric:        metric,
		Value:         value,
		Unit:          unit,
		DataType:      "sensor",
		Quality:       "good",
		Tags:          make(map[string]string),
		CorrelationID: uuid.New(),
	}
}

// WithLocation sets the location for the telemetry
func (t *DeviceTelemetry) WithLocation(location string) *DeviceTelemetry {
	t.Location = location
	return t
}

// WithTags sets tags for the telemetry
func (t *DeviceTelemetry) WithTags(tags map[string]string) *DeviceTelemetry {
	t.Tags = tags
	return t
}

// WithValidationRange sets the validation range and updates quality
func (t *DeviceTelemetry) WithValidationRange(minValue, maxValue float64) *DeviceTelemetry {
	t.MinValue = &minValue
	t.MaxValue = &maxValue

	// Set quality based on range
	if t.Value < minValue || t.Value > maxValue {
		t.Quality = "bad"
	}

	return t
}

// WithDataType sets the data type
func (t *DeviceTelemetry) WithDataType(dataType string) *DeviceTelemetry {
	t.DataType = dataType
	return t
}

// AggregatedTelemetry represents aggregated telemetry data for time-window queries
type AggregatedTelemetry struct {
	DeviceID    string        `json:"deviceId"`
	Metric      string        `json:"metric"`
	WindowStart time.Time     `json:"windowStart"`
	WindowEnd   time.Time     `json:"windowEnd"`
	WindowSize  time.Duration `json:"windowSize"`
	Count       int64         `json:"count"`
	Average     float64       `json:"average"`
	Min         float64       `json:"min"`
	Max         float64       `json:"max"`
	Sum         float64       `json:"sum"`
	StdDev      float64       `json:"stdDev"`
	Unit        string        `json:"unit"`
}

// Device represents device metadata
type Device struct {
	DeviceID        string            `json:"deviceId"`
	DeviceName      string            `json:"deviceName"`
	DeviceType      string            `json:"deviceType"`
	Location        string            `json:"location"`
	Manufacturer    string            `json:"manufacturer"`
	Model           string            `json:"model"`
	FirmwareVersion string            `json:"firmwareVersion"`
	LastSeen        time.Time         `json:"lastSeen"`
	IsOnline        bool              `json:"isOnline"`
	Properties      map[string]string `json:"properties"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

// NewDevice creates a new device
func NewDevice(deviceID, deviceName, deviceType, location string) *Device {
	now := time.Now().UTC()
	return &Device{
		DeviceID:   deviceID,
		DeviceName: deviceName,
		DeviceType: deviceType,
		Location:   location,
		LastSeen:   now,
		IsOnline:   true,
		Properties: make(map[string]string),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// RecordTelemetryRequest represents the request to record telemetry
type RecordTelemetryRequest struct {
	DeviceID string  `json:"deviceId"`
	Metric   string  `json:"metric"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
}

// TelemetryQueryParams represents parameters for querying telemetry
type TelemetryQueryParams struct {
	DeviceID  string    `json:"deviceId"`
	Metric    string    `json:"metric,omitempty"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Limit     int       `json:"limit,omitempty"`
}
