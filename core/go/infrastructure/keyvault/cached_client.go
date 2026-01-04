package keyvault

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/infrastructure/redis"
	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.uber.org/zap"
)

// CachedClient implements cache-aside pattern for KeyVault operations
type CachedClient interface {
	Client

	// GetUserIntegration retrieves a user's integration secret with caching
	GetUserIntegration(ctx context.Context, userID string, integrationType IntegrationType) (*UserIntegration, error)

	// SetUserIntegration stores a user's integration secret
	SetUserIntegration(ctx context.Context, userID string, integrationType IntegrationType, value string, expiresAt *time.Time) error

	// DeleteUserIntegration removes a user's integration secret
	DeleteUserIntegration(ctx context.Context, userID string, integrationType IntegrationType) error

	// ListUserIntegrations returns all integrations for a user
	ListUserIntegrations(ctx context.Context, userID string) ([]UserIntegration, error)

	// GetCacheStats returns cache performance metrics
	GetCacheStats() *CacheStats

	// InvalidateCache invalidates a specific cache entry
	InvalidateCache(ctx context.Context, key string) error
}

// cachedClient implements CachedClient with Redis cache-aside
type cachedClient struct {
	kvClient    Client
	redisClient redis.Client
	logger      *logger.ContextLogger
	cacheTTL    time.Duration
	cachePrefix string

	// Cache statistics
	cacheHits   int64
	cacheMisses int64
	lastSync    time.Time
}

// NewCachedClient creates a new KeyVault client with Redis caching
func NewCachedClient(cfg CachedClientConfig, log *logger.Logger) (CachedClient, error) {
	componentLogger := log.WithComponent("KeyVaultCachedClient")

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		componentLogger.Error("Invalid configuration",
			zap.Error(err),
			zap.String("error_code", ErrCodeConfigInvalid))
		return nil, fmt.Errorf("invalid cached keyvault config: %w", err)
	}

	// Create base KeyVault client
	kvClient, err := NewClient(cfg.KeyVault, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyvault client: %w", err)
	}

	// Create Redis client
	redisClient, err := redis.NewClient(redis.ClientConfig{
		Host:        cfg.Redis.Host,
		Port:        cfg.Redis.Port,
		Logger:      log,
		PingTimeout: 60 * time.Second, // MINIMUM 60s - NO HARDCODING
	})
	if err != nil {
		kvClient.Close(context.Background())
		componentLogger.Error("Failed to create Redis client for caching",
			zap.Error(err),
			zap.String("error_code", ErrCodeCacheReadFailed))
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	cachePrefix := cfg.CachePrefix
	if cachePrefix == "" {
		cachePrefix = "keyvault:"
	}

	componentLogger.Info("Cached KeyVault client initialized",
		zap.Duration("cache_ttl", cfg.CacheTTL),
		zap.String("cache_prefix", cachePrefix),
		zap.String("redis_host", cfg.Redis.Host),
		zap.Int("redis_port", cfg.Redis.Port))

	return &cachedClient{
		kvClient:    kvClient,
		redisClient: redisClient,
		logger:      componentLogger,
		cacheTTL:    cfg.CacheTTL,
		cachePrefix: cachePrefix,
		lastSync:    time.Now(),
	}, nil
}

// cacheKey generates a cache key with prefix
func (c *cachedClient) cacheKey(name string) string {
	return c.cachePrefix + name
}

// userIntegrationKey generates a secret name for user integrations
func userIntegrationKey(userID string, integrationType IntegrationType) string {
	return fmt.Sprintf("user:%s:%s", userID, integrationType)
}

// GetSecret retrieves a secret with cache-aside pattern
func (c *cachedClient) GetSecret(ctx context.Context, name string) (*Secret, error) {
	start := time.Now()
	cacheKey := c.cacheKey(name)

	// Try cache first
	cached, err := c.redisClient.Get(ctx, cacheKey)
	if err != nil {
		c.logger.Warn("Cache read failed, falling back to KeyVault",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeCacheReadFailed))
	} else if cached != "" {
		// Cache hit
		atomic.AddInt64(&c.cacheHits, 1)

		var secret Secret
		if err := json.Unmarshal([]byte(cached), &secret); err != nil {
			c.logger.Warn("Failed to unmarshal cached secret",
				zap.Error(err),
				zap.String("secret_name", name))
		} else {
			c.logger.Debug("Cache hit",
				zap.String("secret_name", name),
				zap.Duration("duration", time.Since(start)))
			return &secret, nil
		}
	}

	// Cache miss - fetch from KeyVault
	atomic.AddInt64(&c.cacheMisses, 1)

	secret, err := c.kvClient.GetSecret(ctx, name)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, nil // Not found
	}

	// Store in cache
	secretJSON, err := json.Marshal(secret)
	if err != nil {
		c.logger.Warn("Failed to marshal secret for caching",
			zap.Error(err),
			zap.String("secret_name", name))
	} else {
		if err := c.redisClient.Set(ctx, cacheKey, string(secretJSON)); err != nil {
			c.logger.Warn("Failed to cache secret",
				zap.Error(err),
				zap.String("secret_name", name),
				zap.String("error_code", ErrCodeCacheWriteFailed))
		} else {
			// Set TTL
			if err := c.redisClient.Expire(ctx, cacheKey, c.cacheTTL); err != nil {
				c.logger.Warn("Failed to set cache TTL",
					zap.Error(err),
					zap.String("secret_name", name))
			}
		}
	}

	c.logger.Debug("Cache miss - fetched from KeyVault",
		zap.String("secret_name", name),
		zap.Duration("duration", time.Since(start)))

	return secret, nil
}

// SetSecret stores a secret and invalidates cache
func (c *cachedClient) SetSecret(ctx context.Context, name string, value string, tags map[string]string) error {
	// Write to KeyVault first
	if err := c.kvClient.SetSecret(ctx, name, value, tags); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := c.cacheKey(name)
	if err := c.redisClient.Del(ctx, cacheKey); err != nil {
		c.logger.Warn("Failed to invalidate cache after set",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeCacheInvalidate))
	}

	c.lastSync = time.Now()
	return nil
}

// DeleteSecret removes a secret and invalidates cache
func (c *cachedClient) DeleteSecret(ctx context.Context, name string) error {
	// Delete from KeyVault first
	if err := c.kvClient.DeleteSecret(ctx, name); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := c.cacheKey(name)
	if err := c.redisClient.Del(ctx, cacheKey); err != nil {
		c.logger.Warn("Failed to invalidate cache after delete",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeCacheInvalidate))
	}

	return nil
}

// ListSecrets returns all secret names matching a prefix
func (c *cachedClient) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	// List operations bypass cache - go directly to KeyVault
	return c.kvClient.ListSecrets(ctx, prefix)
}

// GetUserIntegration retrieves a user's integration secret with caching
func (c *cachedClient) GetUserIntegration(ctx context.Context, userID string, integrationType IntegrationType) (*UserIntegration, error) {
	secretName := userIntegrationKey(userID, integrationType)

	secret, err := c.GetSecret(ctx, secretName)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		// Integration not configured
		return &UserIntegration{
			Type:   integrationType,
			Status: StatusNotConfigured,
		}, nil
	}

	// Build integration response
	integration := &UserIntegration{
		Type:         integrationType,
		Status:       StatusConnected,
		MaskedKey:    maskSecret(secret.Value),
		ConfiguredAt: secret.CreatedOn,
		ExpiresAt:    secret.ExpiresOn,
		Metadata:     secret.Tags,
	}

	// Check if expired
	if secret.ExpiresOn != nil && secret.ExpiresOn.Before(time.Now()) {
		integration.Status = StatusExpired
	}

	return integration, nil
}

// SetUserIntegration stores a user's integration secret
func (c *cachedClient) SetUserIntegration(ctx context.Context, userID string, integrationType IntegrationType, value string, expiresAt *time.Time) error {
	secretName := userIntegrationKey(userID, integrationType)

	tags := map[string]string{
		"user_id":          userID,
		"integration_type": string(integrationType),
	}

	if err := c.SetSecret(ctx, secretName, value, tags); err != nil {
		return err
	}

	c.logger.Info("User integration configured",
		zap.String("user_id", userID),
		zap.String("integration_type", string(integrationType)),
		zap.Bool("has_expiry", expiresAt != nil))

	return nil
}

// DeleteUserIntegration removes a user's integration secret
func (c *cachedClient) DeleteUserIntegration(ctx context.Context, userID string, integrationType IntegrationType) error {
	secretName := userIntegrationKey(userID, integrationType)

	if err := c.DeleteSecret(ctx, secretName); err != nil {
		return err
	}

	c.logger.Info("User integration removed",
		zap.String("user_id", userID),
		zap.String("integration_type", string(integrationType)))

	return nil
}

// ListUserIntegrations returns all integrations for a user
func (c *cachedClient) ListUserIntegrations(ctx context.Context, userID string) ([]UserIntegration, error) {
	prefix := fmt.Sprintf("user:%s:", userID)

	secretNames, err := c.ListSecrets(ctx, prefix)
	if err != nil {
		return nil, err
	}

	// All possible integration types
	allTypes := []IntegrationType{
		IntegrationWeather,
		IntegrationGoogleHome,
		IntegrationAlexa,
		IntegrationIFTTT,
		IntegrationEnergy,
		IntegrationSMS,
		IntegrationMQTT,
		IntegrationSmartThings,
	}

	// Build a map of configured integrations
	configuredMap := make(map[IntegrationType]bool)
	for _, name := range secretNames {
		// Extract integration type from secret name
		parts := strings.Split(name, ":")
		if len(parts) == 3 {
			intType := IntegrationType(parts[2])
			configuredMap[intType] = true
		}
	}

	// Build full list with status
	var integrations []UserIntegration
	for _, intType := range allTypes {
		if configuredMap[intType] {
			// Get full integration details
			integration, err := c.GetUserIntegration(ctx, userID, intType)
			if err != nil {
				c.logger.Warn("Failed to get integration details",
					zap.Error(err),
					zap.String("user_id", userID),
					zap.String("integration_type", string(intType)))
				continue
			}
			integrations = append(integrations, *integration)
		} else {
			// Not configured
			integrations = append(integrations, UserIntegration{
				Type:   intType,
				Status: StatusNotConfigured,
			})
		}
	}

	return integrations, nil
}

// GetCacheStats returns cache performance metrics
func (c *cachedClient) GetCacheStats() *CacheStats {
	hits := atomic.LoadInt64(&c.cacheHits)
	misses := atomic.LoadInt64(&c.cacheMisses)
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return &CacheStats{
		Hits:     hits,
		Misses:   misses,
		HitRate:  hitRate,
		LastSync: c.lastSync,
	}
}

// InvalidateCache invalidates a specific cache entry
func (c *cachedClient) InvalidateCache(ctx context.Context, key string) error {
	cacheKey := c.cacheKey(key)
	return c.redisClient.Del(ctx, cacheKey)
}

// Health checks both KeyVault and Redis
func (c *cachedClient) Health(ctx context.Context) error {
	// Check KeyVault
	if err := c.kvClient.Health(ctx); err != nil {
		return fmt.Errorf("keyvault unhealthy: %w", err)
	}

	// Check Redis
	if err := c.redisClient.Health(ctx); err != nil {
		return fmt.Errorf("redis cache unhealthy: %w", err)
	}

	return nil
}

// Close releases all resources
func (c *cachedClient) Close(ctx context.Context) error {
	var errs []error

	if err := c.kvClient.Close(ctx); err != nil {
		errs = append(errs, fmt.Errorf("keyvault close error: %w", err))
	}

	if err := c.redisClient.Close(ctx); err != nil {
		errs = append(errs, fmt.Errorf("redis close error: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	c.logger.Info("Cached KeyVault client closed",
		zap.Int64("total_cache_hits", atomic.LoadInt64(&c.cacheHits)),
		zap.Int64("total_cache_misses", atomic.LoadInt64(&c.cacheMisses)))

	return nil
}

// maskSecret returns a masked version of a secret (last 3 characters visible)
func maskSecret(value string) string {
	if len(value) <= 3 {
		return "***"
	}
	return "***" + value[len(value)-3:]
}
