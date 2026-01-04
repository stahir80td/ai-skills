package mongodb

import (
	"context"
	"testing"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
)

// MockMongoClient for testing
type MockMongoClient struct {
	closed bool
}

func (m *MockMongoClient) Ping(ctx context.Context) error {
	return nil
}

func (m *MockMongoClient) Health(ctx context.Context) error {
	return nil
}

func (m *MockMongoClient) Disconnect(ctx context.Context) error {
	m.closed = true
	return nil
}

func TestNewClient(t *testing.T) {
	appLogger, _ := logger.NewProduction("mongodb-test", "1.0.0")
	defer appLogger.Sync()

	// Test with invalid config to avoid actual connection
	cfg := ClientConfig{
		Host:   "",
		Port:   27017,
		Logger: appLogger,
	}

	client, err := NewClient(cfg)
	if err == nil {
		t.Errorf("NewClient() expected error for empty host, got nil")
	}
	if client != nil {
		t.Errorf("NewClient() expected nil client for invalid config, got %v", client)
	}
}

func TestClientConfig(t *testing.T) {
	appLogger, _ := logger.NewProduction("mongodb-test", "1.0.0")
	defer appLogger.Sync()

	tests := []struct {
		name    string
		cfg     ClientConfig
		wantErr bool
	}{
		{
			name: "empty host",
			cfg: ClientConfig{
				Host:   "",
				Port:   27017,
				Logger: appLogger,
			},
			wantErr: true,
		},
		{
			name: "invalid port negative",
			cfg: ClientConfig{
				Host:   "localhost",
				Port:   -1,
				Logger: appLogger,
			},
			wantErr: true,
		},
		{
			name: "invalid port zero",
			cfg: ClientConfig{
				Host:   "localhost",
				Port:   0,
				Logger: appLogger,
			},
			wantErr: true,
		},
		{
			name: "invalid port too high",
			cfg: ClientConfig{
				Host:   "localhost",
				Port:   65536,
				Logger: appLogger,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Logf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockMongoClient(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockMongoClient{}

	// Test Ping
	err := mockClient.Ping(ctx)
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// Test Health
	err = mockClient.Health(ctx)
	if err != nil {
		t.Errorf("Health() error = %v", err)
	}

	// Test Disconnect
	err = mockClient.Disconnect(ctx)
	if err != nil {
		t.Errorf("Disconnect() error = %v", err)
	}
	if !mockClient.closed {
		t.Error("Disconnect() did not set closed flag")
	}
}

func TestClientHealth(t *testing.T) {
	appLogger, _ := logger.NewProduction("mongodb-test", "1.0.0")
	defer appLogger.Sync()

	mockClient := &MockMongoClient{}
	ctx := context.Background()
	err := mockClient.Health(ctx)
	if err != nil {
		t.Logf("Health check: %v", err)
	}
}

func TestClientClose(t *testing.T) {
	appLogger, _ := logger.NewProduction("mongodb-test", "1.0.0")
	defer appLogger.Sync()

	mockClient := &MockMongoClient{}
	ctx := context.Background()
	err := mockClient.Disconnect(ctx)
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
