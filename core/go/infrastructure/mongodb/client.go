package mongodb

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// ClientConfig for MongoDB
type ClientConfig struct {
	Host          string // MongoDB host (for mongodb:// connections)
	Port          int    // MongoDB port (for mongodb:// connections)
	ConnectionURI string // Full connection URI (mongodb:// or mongodb+srv://) - takes precedence over Host/Port
	Database      string
	User          string
	Password      string
	Logger        *logger.Logger

	// Timeout configurations (optional - uses defaults if not set)
	PingTimeout time.Duration // Timeout for initial ping verification (default 15s for Atlas)
}

// NewClient creates a new MongoDB client
func NewClient(cfg ClientConfig) (*mongo.Client, error) {
	var uri string

	// If ConnectionURI is provided, use it directly (supports mongodb+srv://)
	if cfg.ConnectionURI != "" {
		uri = cfg.ConnectionURI
		if cfg.Logger != nil {
			// Mask credentials but show connection type and parameters
			connType := "mongodb://"
			isAtlas := false
			if strings.HasPrefix(cfg.ConnectionURI, "mongodb+srv://") {
				connType = "mongodb+srv:// (Atlas)"
				isAtlas = true
			}

			// Extract host info (without credentials)
			hostInfo := "masked"
			if atIdx := strings.Index(cfg.ConnectionURI, "@"); atIdx > 0 {
				afterAt := cfg.ConnectionURI[atIdx+1:]
				if qIdx := strings.Index(afterAt, "?"); qIdx > 0 {
					hostInfo = afterAt[:qIdx]
				} else {
					hostInfo = afterAt
				}
			}

			cfg.Logger.WithComponent("MongoDBClient").Info("Using connection URI",
				zap.String("connection_type", connType),
				zap.Bool("is_atlas", isAtlas),
				zap.String("host_info", hostInfo),
				zap.String("database", cfg.Database),
				zap.Bool("has_retry_writes", strings.Contains(cfg.ConnectionURI, "retryWrites=true")),
				zap.Bool("has_tls", strings.Contains(cfg.ConnectionURI, "tls=true") || strings.Contains(cfg.ConnectionURI, "ssl=true")))
		}
	} else {
		// Build connection URI from Host/Port (legacy mode)
		if cfg.Host == "" {
			err := fmt.Errorf("MongoDB host cannot be empty")
			if cfg.Logger != nil {
				cfg.Logger.WithComponent("MongoDBClient").Error("Invalid configuration - missing host",
					zap.Error(err),
					zap.String("error_code", "INFRA-MONGODB-CONFIG-ERROR"))
			}
			return nil, err
		}
		if cfg.Port <= 0 || cfg.Port > 65535 {
			err := fmt.Errorf("MongoDB port must be between 1 and 65535, got %d", cfg.Port)
			if cfg.Logger != nil {
				cfg.Logger.WithComponent("MongoDBClient").Error("Invalid configuration - bad port",
					zap.Error(err),
					zap.Int("port", cfg.Port),
					zap.String("error_code", "INFRA-MONGODB-CONFIG-ERROR"))
			}
			return nil, err
		}

		addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))

		if cfg.User != "" && cfg.Password != "" {
			uri = fmt.Sprintf("mongodb://%s:%s@%s", cfg.User, cfg.Password, addr)
		} else {
			uri = fmt.Sprintf("mongodb://%s", addr)
		}

		if cfg.Logger != nil {
			cfg.Logger.WithComponent("MongoDBClient").Debug("Initiating MongoDB connection",
				zap.String("host", cfg.Host),
				zap.Int("port", cfg.Port),
				zap.String("database", cfg.Database))
		}
	}

	// MongoDB Atlas best practice: Use Stable API and simple configuration
	clientOpts := options.Client().ApplyURI(uri)

	if strings.HasPrefix(uri, "mongodb+srv://") {
		// Atlas-specific: Use Stable API version 1 (MongoDB best practice)
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		clientOpts.SetServerAPIOptions(serverAPI)
		clientOpts.SetMaxPoolSize(100)
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("MongoDBClient").Info("Connecting to MongoDB Atlas with Stable API",
				zap.String("api_version", "1"),
				zap.Uint64("max_pool_size", 100))
		}
	} else {
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("MongoDBClient").Info("Connecting to MongoDB (local)")
		}
	}

	// Use context.Background() for production (not context.TODO())
	// MongoDB Go Driver will handle topology discovery in background
	connectStart := time.Now()
	client, err := mongo.Connect(context.Background(), clientOpts)
	connectDuration := time.Since(connectStart)

	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("MongoDBClient").Error("MongoDB connection failed",
				zap.Error(err),
				zap.Duration("duration", connectDuration),
				zap.String("error_code", "INFRA-MONGODB-CONNECT-ERROR"),
				zap.String("host", cfg.Host),
				zap.Int("port", cfg.Port))
		}
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// For Atlas M0: Verify connection with adequate timeout for primary election
	// M0 clusters can take 10-15 seconds during primary elections
	if strings.HasPrefix(uri, "mongodb+srv://") {
		pingTimeout := cfg.PingTimeout
		if pingTimeout == 0 {
			pingTimeout = 60 * time.Second // MINIMUM 60s - NO HARDCODING below this
		}
		pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
		defer cancel()

		if err := client.Ping(pingCtx, nil); err != nil {
			client.Disconnect(context.Background()) // Clean up
			if cfg.Logger != nil {
				cfg.Logger.WithComponent("MongoDBClient").Warn("MongoDB Atlas ping failed - cluster may be paused or no primary available",
					zap.Error(err),
					zap.String("error_code", "INFRA-MONGODB-PING-FAILED"),
					zap.String("hint", "Check Atlas dashboard: https://cloud.mongodb.com/"))
			}
			return nil, fmt.Errorf("MongoDB Atlas cluster unreachable (no primary or paused): %w", err)
		}
	}

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("MongoDBClient").Info("Successfully connected to MongoDB",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("database", cfg.Database),
			zap.Duration("connect_duration", connectDuration),
			zap.String("status", "ready"))
	}

	return client, nil
}
