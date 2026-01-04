package scylladb

import (
	"context"
	"fmt"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

// SessionConfig for ScyllaDB
type SessionConfig struct {
	Hosts          []string
	Keyspace       string
	Logger         *logger.Logger
	Timeout        time.Duration // Query timeout (MINIMUM 60s)
	ConnectTimeout time.Duration // Connection timeout (MINIMUM 60s)
}

// Session interface for ScyllaDB operations
type Session interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	QueryIter(ctx context.Context, query string, args ...interface{}) Iterator
	Health(ctx context.Context) error
	Close(ctx context.Context) error
}

// Row represents a single row result from a query
type Row interface {
	Scan(dest ...interface{}) error
}

// Iterator represents an iterator for query results
type Iterator interface {
	Scan(dest ...interface{}) bool
	Close() error
}

// session implements the Session interface using gocql
type session struct {
	gocqlSession *gocql.Session
	logger       *logger.Logger
}

// NewSession creates a new ScyllaDB session using gocql
func NewSession(cfg SessionConfig) (Session, error) {
	if len(cfg.Hosts) == 0 {
		err := fmt.Errorf("ScyllaDB hosts list cannot be empty")
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("ScyllaDBSession").Error("Invalid configuration - no hosts",
				zap.Error(err),
				zap.String("error_code", "INFRA-SCYLLADB-CONFIG-ERROR"))
		}
		return nil, err
	}
	if cfg.Keyspace == "" {
		err := fmt.Errorf("ScyllaDB keyspace cannot be empty")
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("ScyllaDBSession").Error("Invalid configuration - no keyspace",
				zap.Error(err),
				zap.String("error_code", "INFRA-SCYLLADB-CONFIG-ERROR"))
		}
		return nil, err
	}

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("ScyllaDBSession").Debug("Initiating ScyllaDB session",
			zap.Strings("hosts", cfg.Hosts),
			zap.String("keyspace", cfg.Keyspace),
			zap.Int("host_count", len(cfg.Hosts)))
	}

	// Create gocql cluster configuration
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = cfg.Timeout               // NO HARDCODING - from config
	cluster.ConnectTimeout = cfg.ConnectTimeout // NO HARDCODING - from config
	cluster.ProtoVersion = 4
	cluster.NumConns = 2

	// Create the session
	gocqlSession, err := cluster.CreateSession()
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("ScyllaDBSession").Error("Failed to create ScyllaDB session",
				zap.Error(err),
				zap.Strings("hosts", cfg.Hosts),
				zap.String("error_code", "INFRA-SCYLLADB-CONNECTION-ERROR"))
		}
		return nil, fmt.Errorf("failed to connect to ScyllaDB: %w", err)
	}

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("ScyllaDBSession").Info("Connected to ScyllaDB",
			zap.Strings("hosts", cfg.Hosts),
			zap.String("keyspace", cfg.Keyspace))
	}

	return &session{
		gocqlSession: gocqlSession,
		logger:       cfg.Logger,
	}, nil
}

// QueryContext executes a query and returns results
func (s *session) QueryContext(ctx context.Context, query string, args ...interface{}) error {
	if s.gocqlSession == nil {
		return fmt.Errorf("session not initialized")
	}

	q := s.gocqlSession.Query(query, args...)
	// Bind context if possible
	q = q.WithContext(ctx)

	return q.Exec()
}

// ExecContext executes a query without returning results
func (s *session) ExecContext(ctx context.Context, query string, args ...interface{}) error {
	if s.gocqlSession == nil {
		return fmt.Errorf("session not initialized")
	}

	q := s.gocqlSession.Query(query, args...)
	// Bind context if possible
	q = q.WithContext(ctx)

	return q.Exec()
}

// QueryRow executes a query expected to return at most one row
func (s *session) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	if s.gocqlSession == nil {
		return &row{err: fmt.Errorf("session not initialized")}
	}

	q := s.gocqlSession.Query(query, args...)
	q = q.WithContext(ctx)

	return &row{query: q}
}

// row implements the Row interface
type row struct {
	query *gocql.Query
	err   error
}

// Scan copies the columns from the row into the values pointed at by dest
func (r *row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}

	if r.query == nil {
		return fmt.Errorf("query not initialized")
	}

	// Execute query and scan first row
	iter := r.query.Iter()
	if !iter.Scan(dest...) {
		if err := iter.Close(); err != nil {
			return err
		}
		return fmt.Errorf("no rows found")
	}

	return iter.Close()
}

// QueryIter executes a query and returns an iterator for multiple rows
func (s *session) QueryIter(ctx context.Context, query string, args ...interface{}) Iterator {
	if s.gocqlSession == nil {
		return &iterator{err: fmt.Errorf("session not initialized")}
	}

	q := s.gocqlSession.Query(query, args...)
	q = q.WithContext(ctx)

	return &iterator{iter: q.Iter()}
}

// iterator implements the Iterator interface
type iterator struct {
	iter *gocql.Iter
	err  error
}

// Scan scans the next row into dest, returns false when no more rows
func (i *iterator) Scan(dest ...interface{}) bool {
	if i.err != nil {
		return false
	}

	if i.iter == nil {
		return false
	}

	return i.iter.Scan(dest...)
}

// Close closes the iterator and returns any errors
func (i *iterator) Close() error {
	if i.err != nil {
		return i.err
	}

	if i.iter == nil {
		return nil
	}

	return i.iter.Close()
}

// Health checks the health of the ScyllaDB session
func (s *session) Health(ctx context.Context) error {
	if s.gocqlSession == nil {
		return fmt.Errorf("session not initialized")
	}

	return s.gocqlSession.Query("SELECT * FROM system.local LIMIT 1").WithContext(ctx).Exec()
}

// Close closes the ScyllaDB session
func (s *session) Close(ctx context.Context) error {
	if s.gocqlSession == nil {
		return nil
	}

	s.gocqlSession.Close()
	if s.logger != nil {
		s.logger.WithComponent("ScyllaDBSession").Info("ScyllaDB session closed")
	}

	return nil
}
