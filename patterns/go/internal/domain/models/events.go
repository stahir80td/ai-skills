package models

import (
	"time"

	"github.com/google/uuid"
)

// BaseEvent contains common event fields
type BaseEvent struct {
	EventID       uuid.UUID         `json:"eventId"`
	Timestamp     time.Time         `json:"timestamp"`
	EventType     string            `json:"eventType"`
	Source        string            `json:"source"`
	CorrelationID string            `json:"correlationId,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// NewBaseEvent creates a new base event with defaults
func NewBaseEvent(eventType, source string) BaseEvent {
	return BaseEvent{
		EventID:   uuid.New(),
		Timestamp: time.Now().UTC(),
		EventType: eventType,
		Source:    source,
		Metadata:  make(map[string]string),
	}
}

// OrderEvent represents order-related events for Kafka
type OrderEvent struct {
	BaseEvent
	OrderID         uuid.UUID   `json:"orderId"`
	CustomerID      uuid.UUID   `json:"customerId"`
	Status          OrderStatus `json:"status"`
	TotalAmount     float64     `json:"totalAmount"`
	ItemCount       int         `json:"itemCount"`
	ShippingAddress string      `json:"shippingAddress,omitempty"`
	PreviousStatus  OrderStatus `json:"previousStatus,omitempty"`
}

// NewOrderCreatedEvent creates an order created event
func NewOrderCreatedEvent(order *Order, source string) *OrderEvent {
	return &OrderEvent{
		BaseEvent:       NewBaseEvent("OrderCreated", source),
		OrderID:         order.ID,
		CustomerID:      order.CustomerID,
		Status:          order.Status,
		TotalAmount:     order.TotalAmount,
		ItemCount:       len(order.Items),
		ShippingAddress: order.ShippingAddress,
	}
}

// NewOrderStatusChangedEvent creates an order status changed event
func NewOrderStatusChangedEvent(order *Order, previousStatus OrderStatus, source string) *OrderEvent {
	event := &OrderEvent{
		BaseEvent:      NewBaseEvent("OrderStatusChanged", source),
		OrderID:        order.ID,
		CustomerID:     order.CustomerID,
		Status:         order.Status,
		TotalAmount:    order.TotalAmount,
		ItemCount:      len(order.Items),
		PreviousStatus: previousStatus,
	}
	event.Metadata["previousStatus"] = string(previousStatus)
	event.Metadata["newStatus"] = string(order.Status)
	return event
}

// UserEvent represents user-related events for Kafka
type UserEvent struct {
	BaseEvent
	UserID    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
}

// NewUserRegisteredEvent creates a user registered event
func NewUserRegisteredEvent(profile *UserProfile, source string) *UserEvent {
	return &UserEvent{
		BaseEvent: NewBaseEvent("UserRegistered", source),
		UserID:    profile.ID,
		Email:     profile.Email,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
	}
}

// NewUserProfileUpdatedEvent creates a user profile updated event
func NewUserProfileUpdatedEvent(profile *UserProfile, source string) *UserEvent {
	return &UserEvent{
		BaseEvent: NewBaseEvent("UserProfileUpdated", source),
		UserID:    profile.ID,
		Email:     profile.Email,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
	}
}

// NewUserLoggedInEvent creates a user logged in event
func NewUserLoggedInEvent(userID uuid.UUID, email, source string) *UserEvent {
	return &UserEvent{
		BaseEvent: NewBaseEvent("UserLoggedIn", source),
		UserID:    userID,
		Email:     email,
	}
}

// TelemetryEvent represents telemetry-related events for Kafka
type TelemetryEvent struct {
	BaseEvent
	DeviceID     string  `json:"deviceId"`
	Metric       string  `json:"metric"`
	Value        float64 `json:"value"`
	Unit         string  `json:"unit"`
	AnomalyType  string  `json:"anomalyType,omitempty"`
	AnomalyScore float64 `json:"anomalyScore,omitempty"`
}

// NewTelemetryReceivedEvent creates a telemetry received event
func NewTelemetryReceivedEvent(telemetry *DeviceTelemetry, source string) *TelemetryEvent {
	event := &TelemetryEvent{
		BaseEvent: NewBaseEvent("TelemetryReceived", source),
		DeviceID:  telemetry.DeviceID,
		Metric:    telemetry.Metric,
		Value:     telemetry.Value,
		Unit:      telemetry.Unit,
	}
	event.CorrelationID = telemetry.CorrelationID.String()
	return event
}

// NewAnomalyDetectedEvent creates an anomaly detected event
func NewAnomalyDetectedEvent(telemetry *DeviceTelemetry, anomalyType string, anomalyScore float64, source string) *TelemetryEvent {
	event := &TelemetryEvent{
		BaseEvent:    NewBaseEvent("AnomalyDetected", source),
		DeviceID:     telemetry.DeviceID,
		Metric:       telemetry.Metric,
		Value:        telemetry.Value,
		Unit:         telemetry.Unit,
		AnomalyType:  anomalyType,
		AnomalyScore: anomalyScore,
	}
	event.CorrelationID = telemetry.CorrelationID.String()
	return event
}

// SystemEvent represents system-level events for Kafka
type SystemEvent struct {
	BaseEvent
	Component string `json:"component"`
	Message   string `json:"message"`
	Level     string `json:"level"` // info, warn, error, critical
	Details   string `json:"details,omitempty"`
}

// NewSystemEvent creates a system event
func NewSystemEvent(component, message, level, source string) *SystemEvent {
	return &SystemEvent{
		BaseEvent: NewBaseEvent("SystemEvent", source),
		Component: component,
		Message:   message,
		Level:     level,
	}
}

// NewServiceStartedEvent creates a service started event
func NewServiceStartedEvent(serviceName, version, source string) *SystemEvent {
	event := NewSystemEvent(serviceName, "Service started", "info", source)
	event.EventType = "ServiceStarted"
	event.Metadata["version"] = version
	return event
}

// NewHealthCheckEvent creates a health check event
func NewHealthCheckEvent(component string, healthy bool, source string) *SystemEvent {
	level := "info"
	message := "Health check passed"
	if !healthy {
		level = "error"
		message = "Health check failed"
	}
	event := NewSystemEvent(component, message, level, source)
	event.EventType = "HealthCheck"
	return event
}

// LeaderboardEntry represents an entry in a leaderboard
type LeaderboardEntry struct {
	UserID string  `json:"userId"`
	Score  float64 `json:"score"`
	Rank   int     `json:"rank"`
}

// UpdateLeaderboardRequest represents the request to update a leaderboard
type UpdateLeaderboardRequest struct {
	UserID string  `json:"userId"`
	Score  float64 `json:"score"`
}

// CreateSessionRequest represents the request to create a session
type CreateSessionRequest struct {
	SessionID string    `json:"sessionId"`
	UserID    uuid.UUID `json:"userId"`
	UserEmail string    `json:"userEmail"`
}

// Session represents a user session stored in Redis
type Session struct {
	SessionID string    `json:"sessionId"`
	UserID    uuid.UUID `json:"userId"`
	UserEmail string    `json:"userEmail"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// PlatformAnalyticsResult represents cross-platform analytics
type PlatformAnalyticsResult struct {
	StartDate   time.Time           `json:"startDate"`
	EndDate     time.Time           `json:"endDate"`
	SQLServer   *SQLServerAnalytics `json:"sqlServer,omitempty"`
	MongoDB     *MongoDBAnalytics   `json:"mongodb,omitempty"`
	ScyllaDB    *ScyllaDBAnalytics  `json:"scylladb,omitempty"`
	Redis       *RedisAnalytics     `json:"redis,omitempty"`
	GeneratedAt time.Time           `json:"generatedAt"`
}

// SQLServerAnalytics represents SQL Server specific analytics
type SQLServerAnalytics struct {
	TotalOrders       int64            `json:"totalOrders"`
	TotalRevenue      float64          `json:"totalRevenue"`
	AverageOrderValue float64          `json:"averageOrderValue"`
	OrdersByStatus    map[string]int64 `json:"ordersByStatus"`
}

// MongoDBAnalytics represents MongoDB specific analytics
type MongoDBAnalytics struct {
	TotalUsers       int64 `json:"totalUsers"`
	ActiveUsers      int64 `json:"activeUsers"`
	NewRegistrations int64 `json:"newRegistrations"`
}

// ScyllaDBAnalytics represents ScyllaDB specific analytics
type ScyllaDBAnalytics struct {
	TotalRecords     int64   `json:"totalRecords"`
	UniqueDevices    int64   `json:"uniqueDevices"`
	RecordsPerSecond float64 `json:"recordsPerSecond"`
}

// RedisAnalytics represents Redis specific analytics
type RedisAnalytics struct {
	CacheHitRate     float64 `json:"cacheHitRate"`
	ActiveSessions   int64   `json:"activeSessions"`
	LeaderboardCount int64   `json:"leaderboardCount"`
}
