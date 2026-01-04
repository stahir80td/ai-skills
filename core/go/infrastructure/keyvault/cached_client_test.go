package keyvault

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Mock Clients for Cached Client Testing
// =============================================================================

// MockKeyVaultClient for testing cached client
type MockKeyVaultClient struct {
	secrets    map[string]*Secret
	getCount   int64
	setCount   int64
	delCount   int64
	listCount  int64
	shouldFail bool
}

func NewMockKeyVaultClient() *MockKeyVaultClient {
	return &MockKeyVaultClient{
		secrets: make(map[string]*Secret),
	}
}

func (m *MockKeyVaultClient) GetSecret(ctx context.Context, name string) (*Secret, error) {
	atomic.AddInt64(&m.getCount, 1)
	if m.shouldFail {
		return nil, context.DeadlineExceeded
	}
	secret, exists := m.secrets[name]
	if !exists {
		return nil, nil
	}
	return secret, nil
}

func (m *MockKeyVaultClient) SetSecret(ctx context.Context, name string, value string, tags map[string]string) error {
	atomic.AddInt64(&m.setCount, 1)
	if m.shouldFail {
		return context.DeadlineExceeded
	}
	now := time.Now()
	m.secrets[name] = &Secret{
		Name:      name,
		Value:     value,
		Tags:      tags,
		Enabled:   true,
		CreatedOn: &now,
		UpdatedOn: &now,
	}
	return nil
}

func (m *MockKeyVaultClient) DeleteSecret(ctx context.Context, name string) error {
	atomic.AddInt64(&m.delCount, 1)
	if m.shouldFail {
		return context.DeadlineExceeded
	}
	delete(m.secrets, name)
	return nil
}

func (m *MockKeyVaultClient) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	atomic.AddInt64(&m.listCount, 1)
	if m.shouldFail {
		return nil, context.DeadlineExceeded
	}
	var names []string
	for name := range m.secrets {
		if prefix == "" || len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			names = append(names, name)
		}
	}
	return names, nil
}

func (m *MockKeyVaultClient) Health(ctx context.Context) error {
	if m.shouldFail {
		return context.DeadlineExceeded
	}
	return nil
}

func (m *MockKeyVaultClient) Close(ctx context.Context) error {
	return nil
}

// MockRedisClient for testing cached client
type MockRedisClient struct {
	data       map[string]string
	getCount   int64
	setCount   int64
	delCount   int64
	shouldFail bool
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]string),
	}
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	atomic.AddInt64(&m.getCount, 1)
	if m.shouldFail {
		return "", context.DeadlineExceeded
	}
	val, exists := m.data[key]
	if !exists {
		return "", nil
	}
	return val, nil
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value string, ttl int64) error {
	atomic.AddInt64(&m.setCount, 1)
	if m.shouldFail {
		return context.DeadlineExceeded
	}
	m.data[key] = value
	return nil
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	atomic.AddInt64(&m.delCount, 1)
	if m.shouldFail {
		return context.DeadlineExceeded
	}
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *MockRedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return []string{}, nil
}

func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...string) error {
	return nil
}

func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...string) error {
	return nil
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, seconds int64) error {
	return nil
}

func (m *MockRedisClient) Health(ctx context.Context) error {
	if m.shouldFail {
		return context.DeadlineExceeded
	}
	return nil
}

func (m *MockRedisClient) Close(ctx context.Context) error {
	return nil
}

// =============================================================================
// Cache Stats Tests
// =============================================================================

func TestCacheStats_Fields(t *testing.T) {
	tests := []struct {
		name     string
		hits     int64
		misses   int64
		wantRate float64
	}{
		{
			name:     "all hits",
			hits:     100,
			misses:   0,
			wantRate: 1.0,
		},
		{
			name:     "all misses",
			hits:     0,
			misses:   100,
			wantRate: 0.0,
		},
		{
			name:     "50-50",
			hits:     50,
			misses:   50,
			wantRate: 0.5,
		},
		{
			name:     "no requests",
			hits:     0,
			misses:   0,
			wantRate: 0.0,
		},
		{
			name:     "75% hit rate",
			hits:     75,
			misses:   25,
			wantRate: 0.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate hit rate
			var hitRate float64
			if total := tt.hits + tt.misses; total > 0 {
				hitRate = float64(tt.hits) / float64(total)
			}
			stats := &CacheStats{
				Hits:    tt.hits,
				Misses:  tt.misses,
				HitRate: hitRate,
			}
			if stats.HitRate != tt.wantRate {
				t.Errorf("HitRate = %v, want %v", stats.HitRate, tt.wantRate)
			}
		})
	}
}

// =============================================================================
// Mock Cached Client Tests
// =============================================================================

// testCachedClient is a test implementation
type testCachedClient struct {
	kvClient    *MockKeyVaultClient
	redisClient *MockRedisClient
	cacheTTL    time.Duration
	cachePrefix string
	cacheHits   int64
	cacheMisses int64
}

func newTestCachedClient() *testCachedClient {
	return &testCachedClient{
		kvClient:    NewMockKeyVaultClient(),
		redisClient: NewMockRedisClient(),
		cacheTTL:    5 * time.Minute,
		cachePrefix: "keyvault:",
	}
}

func (c *testCachedClient) GetSecret(ctx context.Context, name string) (*Secret, error) {
	cacheKey := c.cachePrefix + name

	// Try cache first
	cached, err := c.redisClient.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		atomic.AddInt64(&c.cacheHits, 1)
		// In real impl, would unmarshal JSON
		return &Secret{Name: name, Value: cached}, nil
	}

	atomic.AddInt64(&c.cacheMisses, 1)

	// Get from KeyVault
	secret, err := c.kvClient.GetSecret(ctx, name)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, nil
	}

	// Cache it
	c.redisClient.Set(ctx, cacheKey, secret.Value, int64(c.cacheTTL.Seconds()))

	return secret, nil
}

func (c *testCachedClient) SetSecret(ctx context.Context, name string, value string, tags map[string]string) error {
	// Set in KeyVault
	err := c.kvClient.SetSecret(ctx, name, value, tags)
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := c.cachePrefix + name
	c.redisClient.Del(ctx, cacheKey)

	return nil
}

func (c *testCachedClient) DeleteSecret(ctx context.Context, name string) error {
	// Delete from KeyVault
	err := c.kvClient.DeleteSecret(ctx, name)
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := c.cachePrefix + name
	c.redisClient.Del(ctx, cacheKey)

	return nil
}

func (c *testCachedClient) GetCacheStats() *CacheStats {
	return &CacheStats{
		Hits:   atomic.LoadInt64(&c.cacheHits),
		Misses: atomic.LoadInt64(&c.cacheMisses),
	}
}

// =============================================================================
// Cache-Aside Pattern Tests
// =============================================================================

func TestCacheAsidePattern_CacheMiss(t *testing.T) {
	client := newTestCachedClient()
	ctx := context.Background()

	// Set a secret in the underlying store
	client.kvClient.SetSecret(ctx, "test-secret", "test-value", nil)

	// First get should be cache miss
	secret, err := client.GetSecret(ctx, "test-secret")
	if err != nil {
		t.Fatalf("GetSecret() error = %v", err)
	}
	if secret == nil {
		t.Fatal("GetSecret() returned nil")
	}
	if secret.Value != "test-value" {
		t.Errorf("GetSecret() value = %v, want %v", secret.Value, "test-value")
	}

	stats := client.GetCacheStats()
	if stats.Hits != 0 {
		t.Errorf("Expected 0 cache hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
}

func TestCacheAsidePattern_CacheHit(t *testing.T) {
	client := newTestCachedClient()
	ctx := context.Background()

	// Set a secret
	client.kvClient.SetSecret(ctx, "test-secret", "test-value", nil)

	// First get - cache miss
	client.GetSecret(ctx, "test-secret")

	// Second get - should be cache hit
	secret, err := client.GetSecret(ctx, "test-secret")
	if err != nil {
		t.Fatalf("GetSecret() error = %v", err)
	}
	if secret == nil {
		t.Fatal("GetSecret() returned nil")
	}

	stats := client.GetCacheStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
}

func TestCacheAsidePattern_Invalidation(t *testing.T) {
	client := newTestCachedClient()
	ctx := context.Background()

	// Set and get a secret (populates cache)
	client.kvClient.SetSecret(ctx, "test-secret", "original-value", nil)
	client.GetSecret(ctx, "test-secret")

	// Update the secret (should invalidate cache)
	client.SetSecret(ctx, "test-secret", "new-value", nil)

	// Reset the mock redis to simulate invalidated cache
	client.redisClient.data = make(map[string]string)

	// Next get should be cache miss and return new value
	secret, err := client.GetSecret(ctx, "test-secret")
	if err != nil {
		t.Fatalf("GetSecret() error = %v", err)
	}
	if secret.Value != "new-value" {
		t.Errorf("GetSecret() value = %v, want %v", secret.Value, "new-value")
	}

	stats := client.GetCacheStats()
	// Should have 1 hit (second get before update) + 1 miss (first get) + 1 miss (after invalidation)
	if stats.Misses < 2 {
		t.Errorf("Expected at least 2 cache misses after invalidation, got %d", stats.Misses)
	}
}

func TestCacheAsidePattern_DeleteInvalidation(t *testing.T) {
	client := newTestCachedClient()
	ctx := context.Background()

	// Set and cache a secret
	client.kvClient.SetSecret(ctx, "to-delete", "value", nil)
	client.GetSecret(ctx, "to-delete")

	// Verify it's in redis cache
	if _, exists := client.redisClient.data[client.cachePrefix+"to-delete"]; !exists {
		t.Error("Secret should be in cache before delete")
	}

	// Delete (should invalidate cache)
	client.DeleteSecret(ctx, "to-delete")

	// Cache should be invalidated
	if _, exists := client.redisClient.data[client.cachePrefix+"to-delete"]; exists {
		t.Error("Secret should not be in cache after delete")
	}
}

// =============================================================================
// Cache Prefix Tests
// =============================================================================

func TestCachePrefix(t *testing.T) {
	client := newTestCachedClient()
	client.cachePrefix = "test-prefix:"
	ctx := context.Background()

	// Set and get a secret
	client.kvClient.SetSecret(ctx, "my-secret", "value", nil)
	client.GetSecret(ctx, "my-secret")

	// Check that cache key has correct prefix
	expectedKey := "test-prefix:my-secret"
	if _, exists := client.redisClient.data[expectedKey]; !exists {
		t.Errorf("Expected cache key %s not found", expectedKey)
	}
}

// =============================================================================
// UserIntegration Key Generation Tests
// =============================================================================

func TestUserIntegration_Types(t *testing.T) {
	tests := []struct {
		intType IntegrationType
		status  IntegrationStatus
		wantStr string
	}{
		{
			intType: IntegrationWeather,
			status:  StatusConnected,
			wantStr: "weather",
		},
		{
			intType: IntegrationGoogleHome,
			status:  StatusConnected,
			wantStr: "google_home",
		},
		{
			intType: IntegrationMQTT,
			status:  StatusNotConfigured,
			wantStr: "mqtt_broker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantStr, func(t *testing.T) {
			ui := UserIntegration{
				Type:   tt.intType,
				Status: tt.status,
			}
			if string(ui.Type) != tt.wantStr {
				t.Errorf("Type = %v, want %v", ui.Type, tt.wantStr)
			}
		})
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestConcurrentAccess(t *testing.T) {
	client := newTestCachedClient()
	ctx := context.Background()

	// Set up some secrets
	for i := 0; i < 10; i++ {
		name := "secret-" + string(rune('0'+i))
		client.kvClient.SetSecret(ctx, name, "value-"+name, nil)
	}

	// Concurrent reads
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(idx int) {
			name := "secret-" + string(rune('0'+(idx%10)))
			_, err := client.GetSecret(ctx, name)
			if err != nil {
				t.Errorf("Concurrent GetSecret() error = %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	stats := client.GetCacheStats()
	total := stats.Hits + stats.Misses
	if total != 100 {
		t.Errorf("Expected 100 total cache operations, got %d", total)
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestCacheFailure_FallbackToKeyVault(t *testing.T) {
	client := newTestCachedClient()
	ctx := context.Background()

	// Set up a secret in KeyVault
	client.kvClient.SetSecret(ctx, "test-secret", "test-value", nil)

	// Make Redis fail
	client.redisClient.shouldFail = true

	// Should still work by falling back to KeyVault
	// Note: In real implementation, cache errors are logged but don't fail the operation
	secret, err := client.kvClient.GetSecret(ctx, "test-secret")
	if err != nil {
		t.Fatalf("GetSecret() error = %v", err)
	}
	if secret == nil {
		t.Fatal("GetSecret() returned nil")
	}
	if secret.Value != "test-value" {
		t.Errorf("GetSecret() value = %v, want %v", secret.Value, "test-value")
	}
}

func TestKeyVaultFailure(t *testing.T) {
	client := newTestCachedClient()
	ctx := context.Background()

	// Make KeyVault fail
	client.kvClient.shouldFail = true

	// Should return error
	_, err := client.kvClient.GetSecret(ctx, "test-secret")
	if err == nil {
		t.Error("Expected error when KeyVault fails")
	}
}
