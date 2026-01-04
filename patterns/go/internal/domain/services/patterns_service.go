package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/infrastructure/kafka"
	"github.com/your-github-org/ai-scaffolder/core/go/infrastructure/redis"
	"github.com/your-github-org/ai-scaffolder/core/go/infrastructure/scylladb"
	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"github.com/your-github-org/ai-scaffolder/core/go/reliability"
	"github.com/your-github-org/ai-scaffolder/patterns/go/internal/domain/errors"
	"github.com/your-github-org/ai-scaffolder/patterns/go/internal/domain/models"
	"github.com/your-github-org/ai-scaffolder/patterns/go/internal/domain/sli"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// PatternsService implements business logic using Core infrastructure packages
// This demonstrates how to properly use all Core packages for a production service
type PatternsService struct {
	// Core Infrastructure Clients
	sqlDB         *sql.DB          // Core.Infrastructure.SqlServer
	mongoClient   *mongo.Client    // Core.Infrastructure.MongoDB
	mongoDatabase string           // MongoDB database name
	scyllaSession scylladb.Session // Core.Infrastructure.ScyllaDB
	redisClient   redis.Client     // Core.Infrastructure.Redis
	kafkaProducer kafka.Producer   // Core.Infrastructure.Kafka

	// Core packages
	logger *logger.Logger   // Core.Logger
	sli    *sli.PatternsSli // Core.Sli

	// Circuit breakers (Core.Reliability)
	mongoCircuitBreaker  *reliability.CircuitBreaker
	scyllaCircuitBreaker *reliability.CircuitBreaker
	kafkaCircuitBreaker  *reliability.CircuitBreaker
}

// NewPatternsService creates a new patterns service with Core infrastructure clients
func NewPatternsService(
	sqlDB *sql.DB,
	mongoClient *mongo.Client,
	mongoDatabase string,
	scyllaSession scylladb.Session,
	redisClient redis.Client,
	kafkaProducer kafka.Producer,
	log *logger.Logger,
	sliTracker *sli.PatternsSli,
) *PatternsService {
	return &PatternsService{
		sqlDB:                sqlDB,
		mongoClient:          mongoClient,
		mongoDatabase:        mongoDatabase,
		scyllaSession:        scyllaSession,
		redisClient:          redisClient,
		kafkaProducer:        kafkaProducer,
		logger:               log,
		sli:                  sliTracker,
		mongoCircuitBreaker:  reliability.NewCircuitBreaker("mongodb", 5, 30*time.Second),
		scyllaCircuitBreaker: reliability.NewCircuitBreaker("scylladb", 5, 30*time.Second),
		kafkaCircuitBreaker:  reliability.NewCircuitBreaker("kafka", 5, 30*time.Second),
	}
}

// =============================================================================
// SQL Server Operations - Orders (Transactional Data)
// Demonstrates: Core.Infrastructure.SqlServer usage
// =============================================================================

// CreateOrder creates a new order in SQL Server
func (s *PatternsService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	log := s.logger.WithContext(ctx)
	start := time.Now()

	log.Info("Creating order",
		zap.String("customer_id", req.CustomerID.String()),
		zap.Int("item_count", len(req.Items)))

	// Validate request
	if req.CustomerID == uuid.Nil {
		s.sli.RecordOrderCreationFailure()
		return nil, errors.ErrInvalidCustomerID
	}
	if len(req.Items) == 0 {
		s.sli.RecordOrderCreationFailure()
		return nil, errors.ErrEmptyOrderItems
	}

	// Convert request items to model items
	orderItems := make([]models.OrderItem, len(req.Items))
	for i, item := range req.Items {
		orderItems[i] = models.NewOrderItem(
			uuid.New(), // generate product ID
			item.ProductName,
			item.Quantity,
			item.UnitPrice,
		)
	}

	// Create order model
	order := models.NewOrder(req.CustomerID, req.ShippingAddress, orderItems)

	// Insert into SQL Server using Core.Infrastructure.SqlServer
	query := `
		INSERT INTO Orders (Id, CustomerID, TotalAmount, Currency, Status, ShippingAddress, CreatedAt, UpdatedAt)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8)`

	_, err := s.sqlDB.ExecContext(ctx, query,
		sql.Named("p1", order.ID),
		sql.Named("p2", order.CustomerID),
		sql.Named("p3", order.TotalAmount),
		sql.Named("p4", order.Currency),
		sql.Named("p5", string(order.Status)),
		sql.Named("p6", order.ShippingAddress),
		sql.Named("p7", order.CreatedAt),
		sql.Named("p8", order.UpdatedAt),
	)
	if err != nil {
		log.Error("Failed to create order in SQL Server", zap.Error(err))
		s.sli.RecordOrderCreationFailure()
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Publish event via Kafka using Core.Infrastructure.Kafka
	if s.kafkaProducer != nil {
		event := models.NewOrderCreatedEvent(order, "ai-patterns")
		if err := s.publishOrderEvent(ctx, event); err != nil {
			log.Warn("Failed to publish order created event", zap.Error(err))
		}
	}

	s.sli.RecordOrderCreationSuccess(time.Since(start))
	log.Info("Order created successfully",
		zap.String("order_id", order.ID.String()),
		zap.Duration("duration", time.Since(start)))

	return order, nil
}

// GetOrder retrieves an order by ID from SQL Server
func (s *PatternsService) GetOrder(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	log := s.logger.WithContext(ctx)

	log.Debug("Getting order", zap.String("order_id", id.String()))

	query := `
		SELECT Id, CustomerID, TotalAmount, Currency, Status, ShippingAddress, CreatedAt, UpdatedAt
		FROM Orders
		WHERE Id = @p1`

	row := s.sqlDB.QueryRowContext(ctx, query, sql.Named("p1", id))

	var order models.Order
	var status string
	err := row.Scan(
		&order.ID, &order.CustomerID, &order.TotalAmount, &order.Currency,
		&status, &order.ShippingAddress, &order.CreatedAt, &order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrOrderNotFound
	}
	if err != nil {
		log.Error("Failed to get order from SQL Server", zap.Error(err))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	order.Status = models.OrderStatus(status)
	return &order, nil
}

// UpdateOrderStatus updates an order's status in SQL Server
func (s *PatternsService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, newStatus models.OrderStatus) (*models.Order, error) {
	log := s.logger.WithContext(ctx)

	log.Info("Updating order status",
		zap.String("order_id", id.String()),
		zap.String("new_status", string(newStatus)))

	// Get current order
	order, err := s.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	previousStatus := order.Status

	// Update in SQL Server
	query := `UPDATE Orders SET Status = @p1, UpdatedAt = @p2 WHERE Id = @p3`
	_, err = s.sqlDB.ExecContext(ctx, query,
		sql.Named("p1", string(newStatus)),
		sql.Named("p2", time.Now()),
		sql.Named("p3", id),
	)
	if err != nil {
		log.Error("Failed to update order status", zap.Error(err))
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	order.Status = newStatus
	order.UpdatedAt = time.Now()

	// Publish status change event via Kafka
	if s.kafkaProducer != nil {
		event := models.NewOrderStatusChangedEvent(order, previousStatus, "ai-patterns")
		if err := s.publishOrderEvent(ctx, event); err != nil {
			log.Warn("Failed to publish order status changed event", zap.Error(err))
		}
	}

	return order, nil
}

// =============================================================================
// MongoDB Operations - User Profiles (Document Data)
// Demonstrates: Core.Infrastructure.MongoDB usage
// =============================================================================

// CreateUser creates a new user profile in MongoDB
func (s *PatternsService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserProfile, error) {
	log := s.logger.WithContext(ctx)
	start := time.Now()

	log.Info("Creating user profile",
		zap.String("email", req.Email))

	// Validate request
	if req.Email == "" {
		return nil, errors.ErrInvalidEmail
	}

	// Create user profile
	profile := models.NewUserProfile(req.Email, req.FirstName, req.LastName)

	// Execute with circuit breaker (Core.Reliability)
	err := s.mongoCircuitBreaker.Execute(func() error {
		collection := s.mongoClient.Database(s.mongoDatabase).Collection("user_profiles")
		_, err := collection.InsertOne(ctx, profile)
		return err
	})

	if err != nil {
		log.Error("Failed to create user in MongoDB", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Publish event via Kafka
	if s.kafkaProducer != nil {
		event := models.NewUserRegisteredEvent(profile, "ai-patterns")
		if err := s.publishUserEvent(ctx, event); err != nil {
			log.Warn("Failed to publish user registered event", zap.Error(err))
		}
	}

	log.Info("User profile created successfully",
		zap.String("user_id", profile.ID.String()),
		zap.Duration("duration", time.Since(start)))

	return profile, nil
}

// GetUser retrieves a user profile by ID from MongoDB
func (s *PatternsService) GetUser(ctx context.Context, id uuid.UUID) (*models.UserProfile, error) {
	log := s.logger.WithContext(ctx)

	log.Debug("Getting user profile", zap.String("user_id", id.String()))

	var profile models.UserProfile

	err := s.mongoCircuitBreaker.Execute(func() error {
		collection := s.mongoClient.Database(s.mongoDatabase).Collection("user_profiles")
		filter := bson.M{"_id": id.String()}
		return collection.FindOne(ctx, filter).Decode(&profile)
	})

	if err == mongo.ErrNoDocuments {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		log.Error("Failed to get user from MongoDB", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &profile, nil
}

// UpdateUserPreferences updates user preferences in MongoDB
func (s *PatternsService) UpdateUserPreferences(ctx context.Context, id uuid.UUID, prefs models.UserPreferences) error {
	log := s.logger.WithContext(ctx)

	log.Info("Updating user preferences", zap.String("user_id", id.String()))

	err := s.mongoCircuitBreaker.Execute(func() error {
		collection := s.mongoClient.Database(s.mongoDatabase).Collection("user_profiles")
		filter := bson.M{"_id": id.String()}
		update := bson.M{
			"$set": bson.M{
				"preferences": prefs,
				"updated_at":  time.Now(),
			},
		}
		_, err := collection.UpdateOne(ctx, filter, update)
		return err
	})

	if err != nil {
		log.Error("Failed to update user preferences", zap.Error(err))
		return fmt.Errorf("failed to update preferences: %w", err)
	}

	return nil
}

// =============================================================================
// ScyllaDB Operations - Telemetry (Time-Series Data)
// Demonstrates: Core.Infrastructure.ScyllaDB usage
// =============================================================================

// RecordTelemetry records device telemetry in ScyllaDB
func (s *PatternsService) RecordTelemetry(ctx context.Context, req *models.RecordTelemetryRequest) (*models.DeviceTelemetry, error) {
	log := s.logger.WithContext(ctx)
	start := time.Now()

	log.Info("Recording telemetry",
		zap.String("device_id", req.DeviceID),
		zap.String("metric", req.Metric))

	// Create telemetry record
	telemetry := &models.DeviceTelemetry{
		CorrelationID: uuid.New(),
		DeviceID:      req.DeviceID,
		Metric:        req.Metric,
		Value:         req.Value,
		Unit:          req.Unit,
		Timestamp:     time.Now(),
	}

	// Insert into ScyllaDB using Core.Infrastructure.ScyllaDB
	err := s.scyllaCircuitBreaker.Execute(func() error {
		query := `
			INSERT INTO device_telemetry (correlation_id, device_id, metric, value, unit, timestamp)
			VALUES (?, ?, ?, ?, ?, ?)`
		return s.scyllaSession.ExecContext(ctx, query,
			telemetry.CorrelationID,
			telemetry.DeviceID,
			telemetry.Metric,
			telemetry.Value,
			telemetry.Unit,
			telemetry.Timestamp,
		)
	})

	if err != nil {
		log.Error("Failed to record telemetry in ScyllaDB", zap.Error(err))
		s.sli.RecordTelemetryIngestionFailure()
		return nil, fmt.Errorf("failed to record telemetry: %w", err)
	}

	// Publish telemetry event via Kafka
	if s.kafkaProducer != nil {
		event := models.NewTelemetryReceivedEvent(telemetry, "ai-patterns")
		if err := s.publishTelemetryEvent(ctx, event); err != nil {
			log.Warn("Failed to publish telemetry event", zap.Error(err))
		}
	}

	s.sli.RecordTelemetryIngestionSuccess(time.Since(start))
	log.Info("Telemetry recorded successfully",
		zap.String("device_id", telemetry.DeviceID),
		zap.Duration("duration", time.Since(start)))

	return telemetry, nil
}

// GetTelemetryHistory retrieves telemetry history from ScyllaDB
func (s *PatternsService) GetTelemetryHistory(ctx context.Context, deviceID string, startTime, endTime time.Time) ([]*models.DeviceTelemetry, error) {
	log := s.logger.WithContext(ctx)

	log.Debug("Getting telemetry history",
		zap.String("device_id", deviceID),
		zap.Time("start", startTime),
		zap.Time("end", endTime))

	var results []*models.DeviceTelemetry

	err := s.scyllaCircuitBreaker.Execute(func() error {
		query := `
			SELECT correlation_id, device_id, metric, value, unit, timestamp
			FROM device_telemetry
			WHERE device_id = ? AND timestamp >= ? AND timestamp <= ?
			ORDER BY timestamp DESC
			LIMIT 1000`

		iter := s.scyllaSession.QueryIter(ctx, query, deviceID, startTime, endTime)
		defer iter.Close()

		var t models.DeviceTelemetry
		for iter.Scan(&t.CorrelationID, &t.DeviceID, &t.Metric, &t.Value, &t.Unit, &t.Timestamp) {
			record := t // copy
			results = append(results, &record)
		}
		return iter.Close()
	})

	if err != nil {
		log.Error("Failed to get telemetry from ScyllaDB", zap.Error(err))
		return nil, fmt.Errorf("failed to get telemetry: %w", err)
	}

	return results, nil
}

// =============================================================================
// Redis Operations - Cache & Real-Time Data
// Demonstrates: Core.Infrastructure.Redis usage
// =============================================================================

// UpdateLeaderboard updates a leaderboard entry in Redis
func (s *PatternsService) UpdateLeaderboard(ctx context.Context, category, userID string, score float64) error {
	log := s.logger.WithContext(ctx)

	log.Info("Updating leaderboard",
		zap.String("category", category),
		zap.String("user_id", userID),
		zap.Float64("score", score))

	// Store in Redis using Core.Infrastructure.Redis
	key := fmt.Sprintf("leaderboard:%s", category)
	entry := map[string]interface{}{
		"user_id": userID,
		"score":   score,
	}

	if err := s.redisClient.Set(ctx, fmt.Sprintf("%s:%s", key, userID), entry); err != nil {
		log.Error("Failed to update leaderboard in Redis", zap.Error(err))
		return fmt.Errorf("failed to update leaderboard: %w", err)
	}

	// Add to set for tracking
	if err := s.redisClient.SAdd(ctx, key, userID); err != nil {
		log.Warn("Failed to add to leaderboard set", zap.Error(err))
	}

	return nil
}

// GetLeaderboard retrieves leaderboard entries from Redis
func (s *PatternsService) GetLeaderboard(ctx context.Context, category string, top int) ([]models.LeaderboardEntry, error) {
	log := s.logger.WithContext(ctx)

	log.Debug("Getting leaderboard",
		zap.String("category", category),
		zap.Int("top", top))

	key := fmt.Sprintf("leaderboard:%s", category)

	// Get members from set using Core.Infrastructure.Redis
	members, err := s.redisClient.SMembers(ctx, key)
	if err != nil {
		log.Error("Failed to get leaderboard members from Redis", zap.Error(err))
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	var entries []models.LeaderboardEntry
	for i, userID := range members {
		if i >= top {
			break
		}

		// Get entry data
		data, err := s.redisClient.Get(ctx, fmt.Sprintf("%s:%s", key, userID))
		if err != nil || data == "" {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			continue
		}

		entries = append(entries, models.LeaderboardEntry{
			UserID: userID,
			Score:  entry["score"].(float64),
			Rank:   i + 1,
		})
	}

	return entries, nil
}

// CreateSession creates a user session in Redis
func (s *PatternsService) CreateSession(ctx context.Context, req *models.CreateSessionRequest) (*models.Session, error) {
	log := s.logger.WithContext(ctx)

	log.Info("Creating session",
		zap.String("user_id", req.UserID.String()))

	session := &models.Session{
		SessionID: uuid.New().String(),
		UserID:    req.UserID,
		UserEmail: req.UserEmail,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Store in Redis using Core.Infrastructure.Redis
	key := fmt.Sprintf("session:%s", session.SessionID)
	if err := s.redisClient.Set(ctx, key, session); err != nil {
		log.Error("Failed to create session in Redis", zap.Error(err))
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Set expiration
	if err := s.redisClient.Expire(ctx, key, 24*time.Hour); err != nil {
		log.Warn("Failed to set session expiration", zap.Error(err))
	}

	// Publish session created event via Kafka
	if s.kafkaProducer != nil {
		event := models.NewUserLoggedInEvent(session.UserID, session.UserEmail, "ai-patterns")
		if err := s.publishUserEvent(ctx, event); err != nil {
			log.Warn("Failed to publish session created event", zap.Error(err))
		}
	}

	return session, nil
}

// GetSession retrieves a session from Redis
func (s *PatternsService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	log := s.logger.WithContext(ctx)

	log.Debug("Getting session", zap.String("session_id", sessionID))

	key := fmt.Sprintf("session:%s", sessionID)
	data, err := s.redisClient.Get(ctx, key)
	if err != nil {
		log.Error("Failed to get session from Redis", zap.Error(err))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if data == "" {
		return nil, errors.ErrSessionNotFound
	}

	var session models.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		log.Error("Failed to unmarshal session", zap.Error(err))
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	return &session, nil
}

// =============================================================================
// Kafka Event Publishing
// Demonstrates: Core.Infrastructure.Kafka usage
// =============================================================================

func (s *PatternsService) publishOrderEvent(ctx context.Context, event *models.OrderEvent) error {
	return s.kafkaCircuitBreaker.Execute(func() error {
		payload, _ := json.Marshal(event)
		headers := map[string]string{
			"event_type":     event.EventType,
			"correlation_id": event.CorrelationID,
		}
		return s.kafkaProducer.SendMessage(ctx, "orders.events", event.OrderID.String(), payload, headers)
	})
}

func (s *PatternsService) publishUserEvent(ctx context.Context, event *models.UserEvent) error {
	return s.kafkaCircuitBreaker.Execute(func() error {
		payload, _ := json.Marshal(event)
		headers := map[string]string{
			"event_type":     event.EventType,
			"correlation_id": event.CorrelationID,
		}
		return s.kafkaProducer.SendMessage(ctx, "users.events", event.UserID.String(), payload, headers)
	})
}

func (s *PatternsService) publishTelemetryEvent(ctx context.Context, event *models.TelemetryEvent) error {
	return s.kafkaCircuitBreaker.Execute(func() error {
		payload, _ := json.Marshal(event)
		headers := map[string]string{
			"event_type":     event.EventType,
			"correlation_id": event.CorrelationID,
		}
		return s.kafkaProducer.SendMessage(ctx, "telemetry.events", event.DeviceID, payload, headers)
	})
}

// =============================================================================
// Cross-Platform Analytics
// Demonstrates: Using multiple Core infrastructure packages together
// =============================================================================

// GetAnalytics retrieves analytics across all platforms
func (s *PatternsService) GetAnalytics(ctx context.Context, startDate, endDate time.Time) (*models.PlatformAnalyticsResult, error) {
	log := s.logger.WithContext(ctx)

	log.Info("Generating cross-platform analytics",
		zap.Time("start", startDate),
		zap.Time("end", endDate))

	result := &models.PlatformAnalyticsResult{
		StartDate:   startDate,
		EndDate:     endDate,
		GeneratedAt: time.Now(),
	}

	// SQL Server analytics
	if s.sqlDB != nil {
		sqlAnalytics, err := s.getSQLServerAnalytics(ctx, startDate, endDate)
		if err != nil {
			log.Warn("Failed to get SQL Server analytics", zap.Error(err))
		} else {
			result.SQLServer = sqlAnalytics
		}
	}

	// MongoDB analytics
	if s.mongoClient != nil {
		mongoAnalytics, err := s.getMongoDBAnalytics(ctx, startDate, endDate)
		if err != nil {
			log.Warn("Failed to get MongoDB analytics", zap.Error(err))
		} else {
			result.MongoDB = mongoAnalytics
		}
	}

	// ScyllaDB analytics
	if s.scyllaSession != nil {
		scyllaAnalytics, err := s.getScyllaDBAnalytics(ctx, startDate, endDate)
		if err != nil {
			log.Warn("Failed to get ScyllaDB analytics", zap.Error(err))
		} else {
			result.ScyllaDB = scyllaAnalytics
		}
	}

	// Redis analytics
	if s.redisClient != nil {
		redisAnalytics, err := s.getRedisAnalytics(ctx)
		if err != nil {
			log.Warn("Failed to get Redis analytics", zap.Error(err))
		} else {
			result.Redis = redisAnalytics
		}
	}

	return result, nil
}

func (s *PatternsService) getSQLServerAnalytics(ctx context.Context, start, end time.Time) (*models.SQLServerAnalytics, error) {
	query := `
		SELECT 
			COUNT(*) as total_orders,
			COALESCE(SUM(TotalAmount), 0) as total_revenue,
			COALESCE(AVG(TotalAmount), 0) as avg_order_value
		FROM Orders
		WHERE CreatedAt BETWEEN @p1 AND @p2`

	row := s.sqlDB.QueryRowContext(ctx, query, sql.Named("p1", start), sql.Named("p2", end))

	var analytics models.SQLServerAnalytics
	if err := row.Scan(&analytics.TotalOrders, &analytics.TotalRevenue, &analytics.AverageOrderValue); err != nil {
		return nil, err
	}

	return &analytics, nil
}

func (s *PatternsService) getMongoDBAnalytics(ctx context.Context, start, end time.Time) (*models.MongoDBAnalytics, error) {
	collection := s.mongoClient.Database(s.mongoDatabase).Collection("user_profiles")

	totalUsers, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	newRegistrations, err := collection.CountDocuments(ctx, bson.M{
		"created_at": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
	if err != nil {
		return nil, err
	}

	return &models.MongoDBAnalytics{
		TotalUsers:       totalUsers,
		NewRegistrations: newRegistrations,
	}, nil
}

func (s *PatternsService) getScyllaDBAnalytics(ctx context.Context, start, end time.Time) (*models.ScyllaDBAnalytics, error) {
	var analytics models.ScyllaDBAnalytics

	query := `SELECT COUNT(*) FROM device_telemetry WHERE timestamp >= ? AND timestamp <= ? ALLOW FILTERING`
	row := s.scyllaSession.QueryRow(ctx, query, start, end)
	if err := row.Scan(&analytics.TotalRecords); err != nil {
		return nil, err
	}

	return &analytics, nil
}

func (s *PatternsService) getRedisAnalytics(ctx context.Context) (*models.RedisAnalytics, error) {
	// Count active sessions by checking keys
	sessions, err := s.redisClient.SMembers(ctx, "active_sessions")
	if err != nil {
		return &models.RedisAnalytics{}, nil
	}

	return &models.RedisAnalytics{
		ActiveSessions: int64(len(sessions)),
	}, nil
}

// =============================================================================
// Health Check
// Demonstrates: Health checking with Core infrastructure packages
// =============================================================================

// HealthCheck checks the health of all infrastructure components
func (s *PatternsService) HealthCheck(ctx context.Context) map[string]string {
	health := make(map[string]string)

	// Check SQL Server using Core.Infrastructure.SqlServer
	if s.sqlDB != nil {
		if err := s.sqlDB.PingContext(ctx); err != nil {
			health["sqlserver"] = "unhealthy: " + err.Error()
		} else {
			health["sqlserver"] = "healthy"
		}
	}

	// Check MongoDB using Core.Infrastructure.MongoDB
	if s.mongoClient != nil {
		if err := s.mongoClient.Ping(ctx, nil); err != nil {
			health["mongodb"] = "unhealthy: " + err.Error()
		} else {
			health["mongodb"] = "healthy"
		}
	}

	// Check ScyllaDB using Core.Infrastructure.ScyllaDB
	if s.scyllaSession != nil {
		if err := s.scyllaSession.Health(ctx); err != nil {
			health["scylladb"] = "unhealthy: " + err.Error()
		} else {
			health["scylladb"] = "healthy"
		}
	}

	// Check Redis using Core.Infrastructure.Redis
	if s.redisClient != nil {
		if err := s.redisClient.Health(ctx); err != nil {
			health["redis"] = "unhealthy: " + err.Error()
		} else {
			health["redis"] = "healthy"
		}
	}

	// Check Kafka using Core.Infrastructure.Kafka
	if s.kafkaProducer != nil {
		if err := s.kafkaProducer.Health(ctx); err != nil {
			health["kafka"] = "unhealthy: " + err.Error()
		} else {
			health["kafka"] = "healthy"
		}
	}

	return health
}
