package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Core packages
	"github.com/your-github-org/ai-scaffolder/core/go/infrastructure/kafka"
	"github.com/your-github-org/ai-scaffolder/core/go/infrastructure/mongodb"
	"github.com/your-github-org/ai-scaffolder/core/go/infrastructure/redis"
	"github.com/your-github-org/ai-scaffolder/core/go/infrastructure/scylladb"
	"github.com/your-github-org/ai-scaffolder/core/go/infrastructure/sqlserver"
	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"github.com/your-github-org/ai-scaffolder/core/go/metrics"

	// Patterns packages
	"github.com/your-github-org/ai-scaffolder/patterns/go/config"
	"github.com/your-github-org/ai-scaffolder/patterns/go/internal/api"
	"github.com/your-github-org/ai-scaffolder/patterns/go/internal/domain/services"
	"github.com/your-github-org/ai-scaffolder/patterns/go/internal/domain/sli"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func main() {
	// ========================================
	// 1. LOAD CONFIGURATION
	// ========================================
	var cfg *config.Config
	var err error

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err = config.Load(configPath)
	if err != nil {
		cfg = config.LoadFromEnv()
	}

	// ========================================
	// 2. CORE.LOGGER SETUP
	// ========================================
	var log *logger.Logger
	if cfg.Service.Environment == "development" {
		log, err = logger.NewDevelopment(cfg.Service.Name, cfg.Service.Version)
	} else {
		log, err = logger.NewProduction(cfg.Service.Name, cfg.Service.Version)
	}
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting AI Patterns service - demonstrating Core package usage",
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Service.Environment),
		zap.Int("port", cfg.Service.Port))

	// ========================================
	// 3. CORE.METRICS SETUP
	// ========================================
	serviceMetrics := metrics.NewServiceMetrics(metrics.Config{
		ServiceName: cfg.Service.Name,
		Namespace:   "iot_homeguard",
		Subsystem:   "patterns",
	})

	log.Info("Core.Metrics initialized",
		zap.String("namespace", "iot_homeguard"),
		zap.String("subsystem", "patterns"))

	// ========================================
	// 4. CORE.SLI SETUP
	// ========================================
	sliTracker := sli.NewPatternsSli(cfg.Service.Name)
	log.Info("Core.Sli tracker initialized")

	// ========================================
	// 5. CORE.INFRASTRUCTURE CLIENTS SETUP
	// ========================================
	log.Info("Initializing Core.Infrastructure clients...")

	// --- Core.Infrastructure.SqlServer ---
	var sqlDB *sql.DB
	if cfg.SQLServer.Server != "" {
		sqlDB, err = sqlserver.NewClient(sqlserver.ClientConfig{
			Server:      cfg.SQLServer.Server,
			Database:    cfg.SQLServer.Database,
			User:        cfg.SQLServer.User,
			Password:    cfg.SQLServer.Password,
			Logger:      log,
			PingTimeout: 60 * time.Second,
		})
		if err != nil {
			log.Warn("Failed to connect to SQL Server - continuing without it",
				zap.Error(err),
				zap.String("server", cfg.SQLServer.Server))
		} else {
			log.Info("Core.Infrastructure.SqlServer connected",
				zap.String("server", cfg.SQLServer.Server),
				zap.String("database", cfg.SQLServer.Database))
		}
	}

	// --- Core.Infrastructure.MongoDB ---
	var mongoClient *mongo.Client
	if cfg.MongoDB.ConnectionURI != "" {
		mongoClient, err = mongodb.NewClient(mongodb.ClientConfig{
			ConnectionURI: cfg.MongoDB.ConnectionURI,
			Database:      cfg.MongoDB.Database,
			Logger:        log,
			PingTimeout:   60 * time.Second,
		})
		if err != nil {
			log.Warn("Failed to connect to MongoDB - continuing without it",
				zap.Error(err))
		} else {
			log.Info("Core.Infrastructure.MongoDB connected",
				zap.String("database", cfg.MongoDB.Database))
		}
	}

	// --- Core.Infrastructure.ScyllaDB ---
	var scyllaSession scylladb.Session
	if len(cfg.ScyllaDB.Hosts) > 0 {
		scyllaSession, err = scylladb.NewSession(scylladb.SessionConfig{
			Hosts:          cfg.ScyllaDB.Hosts,
			Keyspace:       cfg.ScyllaDB.Keyspace,
			Logger:         log,
			Timeout:        60 * time.Second,
			ConnectTimeout: 60 * time.Second,
		})
		if err != nil {
			log.Warn("Failed to connect to ScyllaDB - continuing without it",
				zap.Error(err),
				zap.Strings("hosts", cfg.ScyllaDB.Hosts))
		} else {
			log.Info("Core.Infrastructure.ScyllaDB connected",
				zap.Strings("hosts", cfg.ScyllaDB.Hosts),
				zap.String("keyspace", cfg.ScyllaDB.Keyspace))
		}
	}

	// --- Core.Infrastructure.Redis ---
	var redisClient redis.Client
	if cfg.Redis.Host != "" {
		redisClient, err = redis.NewClient(redis.ClientConfig{
			Host:        cfg.Redis.Host,
			Port:        cfg.Redis.Port,
			Logger:      log,
			PingTimeout: 60 * time.Second,
		})
		if err != nil {
			log.Warn("Failed to connect to Redis - continuing without it",
				zap.Error(err),
				zap.String("host", cfg.Redis.Host))
		} else {
			log.Info("Core.Infrastructure.Redis connected",
				zap.String("host", cfg.Redis.Host),
				zap.Int("port", cfg.Redis.Port))
		}
	}

	// --- Core.Infrastructure.Kafka ---
	var kafkaProducer kafka.Producer
	if len(cfg.Kafka.Brokers) > 0 {
		kafkaProducer, err = kafka.NewProducer(kafka.ProducerConfig{
			Brokers: cfg.Kafka.Brokers,
			Logger:  log,
		})
		if err != nil {
			log.Warn("Failed to connect to Kafka - continuing without it",
				zap.Error(err),
				zap.Strings("brokers", cfg.Kafka.Brokers))
		} else {
			log.Info("Core.Infrastructure.Kafka producer created",
				zap.Strings("brokers", cfg.Kafka.Brokers))
		}
	}

	log.Info("Core.Infrastructure initialization complete",
		zap.Bool("sqlserver", sqlDB != nil),
		zap.Bool("mongodb", mongoClient != nil),
		zap.Bool("scylladb", scyllaSession != nil),
		zap.Bool("redis", redisClient != nil),
		zap.Bool("kafka", kafkaProducer != nil))

	// ========================================
	// 6. SERVICE LAYER SETUP
	// ========================================
	patternsService := services.NewPatternsService(
		sqlDB,
		mongoClient,
		cfg.MongoDB.Database,
		scyllaSession,
		redisClient,
		kafkaProducer,
		log,
		sliTracker,
	)

	log.Info("PatternsService created with Core infrastructure clients")

	// ========================================
	// 7. HTTP HANDLER & SERVER SETUP
	// ========================================
	handler := api.NewPatternsHandler(patternsService, log, serviceMetrics)
	server := api.NewServer(fmt.Sprintf("%d", cfg.Service.Port), handler, log, serviceMetrics)

	// ========================================
	// 8. GRACEFUL SHUTDOWN SETUP
	// ========================================
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("HTTP server starting",
			zap.Int("port", cfg.Service.Port),
			zap.String("address", server.Addr))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed to start", zap.Error(err))
			os.Exit(1)
		}
	}()

	log.Info("AI Patterns service started successfully",
		zap.String("url", fmt.Sprintf("http://localhost:%d", cfg.Service.Port)),
		zap.String("health", fmt.Sprintf("http://localhost:%d/health", cfg.Service.Port)),
		zap.String("metrics", fmt.Sprintf("http://localhost:%d/metrics", cfg.Service.Port)))

	// Wait for shutdown signal
	<-done
	log.Info("Received shutdown signal")

	// ========================================
	// 9. GRACEFUL SHUTDOWN WITH CORE CLIENTS
	// ========================================
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Info("Shutting down HTTP server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown failed", zap.Error(err))
	}

	// Close Core.Infrastructure clients
	log.Info("Closing Core.Infrastructure clients...")

	if kafkaProducer != nil {
		if err := kafkaProducer.Close(ctx); err != nil {
			log.Error("Failed to close Kafka producer", zap.Error(err))
		}
	}

	if redisClient != nil {
		if err := redisClient.Close(ctx); err != nil {
			log.Error("Failed to close Redis client", zap.Error(err))
		}
	}

	if scyllaSession != nil {
		if err := scyllaSession.Close(ctx); err != nil {
			log.Error("Failed to close ScyllaDB session", zap.Error(err))
		}
	}

	if mongoClient != nil {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Error("Failed to close MongoDB client", zap.Error(err))
		}
	}

	if sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			log.Error("Failed to close SQL Server connection", zap.Error(err))
		}
	}

	log.Info("AI Patterns service stopped - all Core.Infrastructure clients closed")
}
