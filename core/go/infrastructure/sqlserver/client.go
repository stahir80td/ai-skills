package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	_ "github.com/microsoft/go-mssqldb"
	"go.uber.org/zap"
)

// Config for SQL Server
type ClientConfig struct {
	Server      string
	Database    string
	User        string
	Password    string
	Logger      *logger.Logger
	PingTimeout time.Duration // MINIMUM 60s - NO HARDCODING
}

// NewClient creates a new SQL Server client connection
// Returns a configured *sql.DB that can be used directly
func NewClient(cfg ClientConfig) (*sql.DB, error) {
	if cfg.Server == "" {
		err := fmt.Errorf("SQL Server address cannot be empty")
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("SQLServerClient").Error("Invalid configuration - missing server",
				zap.Error(err),
				zap.String("error_code", "INFRA-SQLSERVER-CONFIG-ERROR"))
		}
		return nil, err
	}
	if cfg.Database == "" {
		err := fmt.Errorf("SQL Server database name cannot be empty")
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("SQLServerClient").Error("Invalid configuration - missing database",
				zap.Error(err),
				zap.String("error_code", "INFRA-SQLSERVER-CONFIG-ERROR"))
		}
		return nil, err
	}

	// Build connection string
	// Format: sqlserver://user:password@server?database=database
	connString := fmt.Sprintf("sqlserver://%s:%s@%s?database=%s",
		cfg.User, cfg.Password, cfg.Server, cfg.Database)

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("SQLServerClient").Debug("Initiating SQL Server connection",
			zap.String("server", cfg.Server),
			zap.String("database", cfg.Database),
			zap.String("user", cfg.User))
	}

	db, err := sql.Open("mssql", connString)
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("SQLServerClient").Error("SQL Server driver initialization failed",
				zap.Error(err),
				zap.String("error_code", "INFRA-SQLSERVER-DRIVER-ERROR"),
				zap.String("server", cfg.Server))
		}
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout) // NO HARDCODING - from config
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("SQLServerClient").Error("SQL Server ping failed",
				zap.Error(err),
				zap.String("error_code", "INFRA-SQLSERVER-CONNECT-ERROR"),
				zap.String("server", cfg.Server),
				zap.String("database", cfg.Database))
		}
		return nil, fmt.Errorf("failed to connect to SQL Server: %w", err)
	}

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("SQLServerClient").Info("Successfully connected to SQL Server",
			zap.String("server", cfg.Server),
			zap.String("database", cfg.Database),
			zap.String("status", "healthy"),
			zap.Int("max_open_conns", 25),
			zap.Int("max_idle_conns", 5))
	}

	return db, nil
}
