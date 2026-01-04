# KeyVault Infrastructure Package

## Overview

The `keyvault` package provides a production-grade wrapper around the Azure KeyVault Emulator (`jamesgoulddev/azure-keyvault-emulator`) with Redis cache-aside pattern for high-performance secret retrieval in IoT applications.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     IoT Application Layer                        │
│         (API Gateway, Device Service, User Service)              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    CachedClient (keyvault)                       │
│   ┌───────────────────────────────────────────────────────────┐ │
│   │            Cache-Aside Pattern                            │ │
│   │  1. Check Redis → 2. Miss? Fetch KeyVault → 3. Cache it   │ │
│   └───────────────────────────────────────────────────────────┘ │
│                         │                 │                      │
│                         ▼                 ▼                      │
│  ┌───────────────────────────┐  ┌─────────────────────────────┐ │
│  │      Redis (Existing)     │  │  Azure KeyVault Emulator    │ │
│  │   TTL-based caching       │  │  jamesgoulddev/azure-...    │ │
│  │   Prefix: keyvault:*      │  │  Port: 4997                 │ │
│  └───────────────────────────┘  └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Features

### 1. Cache-Aside Pattern

Automatic caching with Redis reduces latency and KeyVault load:

```go
// Read Flow:
// 1. Check Redis cache (keyvault:user:123:weather-api-key)
// 2. Cache HIT? → Return immediately
// 3. Cache MISS? → Fetch from KeyVault → Store in Redis with TTL → Return

// Write Flow:
// 1. Write to KeyVault
// 2. Invalidate Redis cache key (next read will repopulate)
```

### 2. User Integration Secrets (IoT Use Case)

Store and manage user's external service integrations:

| Integration Type | Secret Key Format | Example |
|-----------------|-------------------|---------|
| Weather Service | `user:{id}:weather` | OpenWeatherMap API key |
| Google Home | `user:{id}:google_home` | OAuth token |
| Alexa | `user:{id}:alexa` | Skill linking token |
| IFTTT | `user:{id}:ifttt` | Webhook URL |
| Energy Provider | `user:{id}:energy_provider` | Smart meter token |
| SMS Notifications | `user:{id}:sms_notification` | Phone number + provider key |
| MQTT Broker | `user:{id}:mqtt_broker` | Custom broker credentials |
| SmartThings | `user:{id}:smartthings` | Personal Access Token |

### 3. SLI/SRE Integration

- **Cache Hit Rate**: Track cache performance (`GetCacheStats()`)
- **Latency Metrics**: All operations are timed and logged
- **Health Checks**: Combined KeyVault + Redis health endpoint
- **Error Codes**: Structured error codes for incident management

### 4. SOD Error Registry

All errors follow the SOD (Severity × Occurrence × Detectability) scoring pattern:

| Error Code | Description | SOD Score |
|------------|-------------|-----------|
| INFRA-KEYVAULT-010 | Connection failed | 160 (Critical) |
| INFRA-KEYVAULT-020 | Secret not found | 24 (Medium) |
| INFRA-KEYVAULT-030 | Cache write failed | 12 (Low) |

## Usage

### Basic Client (No Caching)

```go
import (
    "github.com/your-org/core/infrastructure/keyvault"
    "github.com/your-org/core/logger"
)

// Create logger
log := logger.NewProduction()
defer log.Sync()

// Initialize KeyVault client
kvClient, err := keyvault.NewClient(keyvault.ClientConfig{
    VaultURL:           os.Getenv("KEYVAULT_URL"),      // From ConfigMap
    Timeout:            30 * time.Second,
    InsecureSkipVerify: os.Getenv("ENV") == "local",    // Local dev only
}, log)
if err != nil {
    log.Fatal("keyvault_init_failed", zap.Error(err))
}
defer kvClient.Close(context.Background())

// Get a secret
secret, err := kvClient.GetSecret(ctx, "api-key-openai")
if err != nil {
    log.Error("secret_fetch_failed", zap.Error(err))
}
```

### Cached Client (Recommended)

```go
import (
    "github.com/your-org/core/infrastructure/keyvault"
    "github.com/your-org/core/logger"
)

log := logger.NewProduction()
defer log.Sync()

// Initialize cached client (uses existing Redis)
client, err := keyvault.NewCachedClient(keyvault.CachedClientConfig{
    KeyVault: keyvault.ClientConfig{
        VaultURL:           os.Getenv("KEYVAULT_URL"),
        Timeout:            30 * time.Second,
        InsecureSkipVerify: os.Getenv("ENV") == "local",
    },
    Redis: keyvault.RedisConfig{
        Host: os.Getenv("REDIS_HOST"),
        Port: 6379,
    },
    CacheTTL:    5 * time.Minute,
    CachePrefix: "keyvault:",
}, log)
if err != nil {
    log.Fatal("cached_keyvault_init_failed", zap.Error(err))
}
defer client.Close(context.Background())

// Get user integration (with caching)
integration, err := client.GetUserIntegration(ctx, "user-123", keyvault.IntegrationWeather)
if err != nil {
    log.Error("integration_fetch_failed", zap.Error(err))
}

// integration.Status = "connected" | "not_configured" | "expired"
// integration.MaskedKey = "***4f2"
```

### List All User Integrations

```go
// Get all integrations for a user (for UI display)
integrations, err := client.ListUserIntegrations(ctx, "user-123")
if err != nil {
    log.Error("list_integrations_failed", zap.Error(err))
}

for _, i := range integrations {
    fmt.Printf("%s: %s (key: %s)\n", i.Type, i.Status, i.MaskedKey)
}
// Output:
// weather: connected (key: ***4f2)
// google_home: not_configured (key: )
// alexa: connected (key: ***8k1)
// ...
```

### Set User Integration

```go
// Configure a new integration
expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
err := client.SetUserIntegration(ctx, "user-123", keyvault.IntegrationWeather, 
    "owm-abc123xyz789", &expiresAt)
if err != nil {
    log.Error("set_integration_failed", zap.Error(err))
}
```

### Cache Statistics

```go
// Get cache performance metrics
stats := client.GetCacheStats()
fmt.Printf("Cache Hit Rate: %.2f%% (%d hits, %d misses)\n", 
    stats.HitRate, stats.Hits, stats.Misses)
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KEYVAULT_URL` | Azure KeyVault Emulator URL | `https://iot-keyvault:4997` |
| `KEYVAULT_TIMEOUT` | Operation timeout | `30s` |
| `KEYVAULT_CACHE_TTL` | Cache TTL duration | `5m` |
| `KEYVAULT_TLS_SKIP_VERIFY` | Skip TLS verification (dev only) | `false` |

### Helm Values

```yaml
keyvault:
  enabled: true
  image: jamesgoulddev/azure-keyvault-emulator:latest
  persist: true  # Keep secrets across restarts
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  storage:
    size: 1Gi
```

## Health Checks

Register with the infrastructure health checker:

```go
import "github.com/your-org/core/infrastructure/health"

healthChecker := health.NewChecker(log)

healthChecker.Register("keyvault-cached", func(ctx context.Context) error {
    return cachedClient.Health(ctx)  // Checks both KeyVault AND Redis
})
```

## Error Handling

```go
import "github.com/your-org/core/errors"

// Register KeyVault errors with global registry
registry := errors.NewErrorRegistry()
keyvault.RegisterErrors(registry)

// Use in service
secret, err := client.GetSecret(ctx, "my-secret")
if err != nil {
    // Error includes code, severity, and context for SRE
    // [INFRA-KEYVAULT-022] HIGH: Failed to retrieve secret from KeyVault
    log.Error("secret_retrieval_failed",
        zap.Error(err),
        zap.String("secret_name", "my-secret"))
}
```

## Testing

```go
// Use mock client for unit tests
type MockKeyVaultClient struct {
    secrets map[string]*Secret
}

func (m *MockKeyVaultClient) GetSecret(ctx context.Context, name string) (*Secret, error) {
    if s, ok := m.secrets[name]; ok {
        return s, nil
    }
    return nil, nil
}
// ... implement other interface methods
```

## Security Considerations

1. **Never log secret values** - Only log secret names and masked versions
2. **TLS Required** - Always use HTTPS in production
3. **InsecureSkipVerify** - Only for local development, never in production
4. **Cache TTL** - Keep short (5min default) to limit exposure window
5. **Secret Rotation** - Invalidate cache immediately on secret update

## Related Packages

- `core/infrastructure/redis` - Redis client used for caching
- `core/infrastructure/health` - Health check framework
- `core/errors` - Error registry with SOD scoring
- `core/logger` - Structured logging
