package keyvault

import (
	"fmt"
	"time"
)

// ClientConfig for Azure KeyVault Emulator connection
type ClientConfig struct {
	// VaultURL is the URL to the Azure KeyVault Emulator (e.g., "https://localhost:4997")
	VaultURL string

	// Timeout for operations (MINIMUM 10s - NO HARDCODING below this)
	Timeout time.Duration

	// TLS configuration
	TLSConfig TLSConfig

	// Optional: Skip TLS verification (ONLY for local development)
	InsecureSkipVerify bool
}

// TLSConfig for KeyVault connection
type TLSConfig struct {
	// CertPath is the path to the TLS certificate file
	CertPath string

	// KeyPath is the path to the TLS key file
	KeyPath string

	// CAPath is the path to the CA certificate file
	CAPath string
}

// CachedClientConfig for KeyVault with Redis cache-aside pattern
type CachedClientConfig struct {
	// KeyVault client configuration
	KeyVault ClientConfig

	// Redis configuration for caching
	Redis RedisConfig

	// CacheTTL is the default TTL for cached secrets (MINIMUM 60s)
	CacheTTL time.Duration

	// CachePrefix is the prefix for all cache keys (default: "keyvault:")
	CachePrefix string
}

// RedisConfig for cache-aside pattern
type RedisConfig struct {
	// Host is the Redis host
	Host string

	// Port is the Redis port
	Port int
}

// Validate validates the ClientConfig
func (c *ClientConfig) Validate() error {
	if c.VaultURL == "" {
		return fmt.Errorf("VaultURL cannot be empty")
	}

	if c.Timeout < 10*time.Second {
		return fmt.Errorf("Timeout must be at least 10 seconds, got %v", c.Timeout)
	}

	return nil
}

// Validate validates the CachedClientConfig
func (c *CachedClientConfig) Validate() error {
	if err := c.KeyVault.Validate(); err != nil {
		return fmt.Errorf("KeyVault config invalid: %w", err)
	}

	if c.Redis.Host == "" {
		return fmt.Errorf("Redis host cannot be empty")
	}

	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		return fmt.Errorf("Redis port must be between 1 and 65535, got %d", c.Redis.Port)
	}

	if c.CacheTTL < 60*time.Second {
		return fmt.Errorf("CacheTTL must be at least 60 seconds, got %v", c.CacheTTL)
	}

	return nil
}

// DefaultConfig returns a default configuration for local development
func DefaultConfig() CachedClientConfig {
	return CachedClientConfig{
		KeyVault: ClientConfig{
			VaultURL:           "https://localhost:4997",
			Timeout:            30 * time.Second,
			InsecureSkipVerify: true, // Local dev only
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		CacheTTL:    5 * time.Minute,
		CachePrefix: "keyvault:",
	}
}

// IntegrationType represents the type of user integration
type IntegrationType string

const (
	IntegrationWeather     IntegrationType = "weather"
	IntegrationGoogleHome  IntegrationType = "google_home"
	IntegrationAlexa       IntegrationType = "alexa"
	IntegrationIFTTT       IntegrationType = "ifttt"
	IntegrationEnergy      IntegrationType = "energy_provider"
	IntegrationSMS         IntegrationType = "sms_notification"
	IntegrationMQTT        IntegrationType = "mqtt_broker"
	IntegrationSmartThings IntegrationType = "smartthings"
)

// IntegrationStatus represents the status of an integration
type IntegrationStatus string

const (
	StatusConnected     IntegrationStatus = "connected"
	StatusNotConfigured IntegrationStatus = "not_configured"
	StatusExpired       IntegrationStatus = "expired"
	StatusError         IntegrationStatus = "error"
)

// UserIntegration represents a user's external service integration
type UserIntegration struct {
	Type         IntegrationType   `json:"type"`
	Status       IntegrationStatus `json:"status"`
	MaskedKey    string            `json:"masked_key,omitempty"` // Last 4 characters only
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
	ConfiguredAt *time.Time        `json:"configured_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Secret represents a KeyVault secret
type Secret struct {
	Name      string            `json:"name"`
	Value     string            `json:"value"`
	Version   string            `json:"version,omitempty"`
	Enabled   bool              `json:"enabled"`
	ExpiresOn *time.Time        `json:"expires_on,omitempty"`
	NotBefore *time.Time        `json:"not_before,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	CreatedOn *time.Time        `json:"created_on,omitempty"`
	UpdatedOn *time.Time        `json:"updated_on,omitempty"`
}

// CacheStats provides cache performance metrics
type CacheStats struct {
	Hits       int64         `json:"hits"`
	Misses     int64         `json:"misses"`
	HitRate    float64       `json:"hit_rate"`
	LastSync   time.Time     `json:"last_sync"`
	AvgLatency time.Duration `json:"avg_latency"`
}
