package keyvault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
)

// =============================================================================
// Mock KeyVault Server
// =============================================================================

type mockKeyVaultServer struct {
	secrets map[string]*Secret
	token   string
}

func newMockKeyVaultServer() *mockKeyVaultServer {
	return &mockKeyVaultServer{
		secrets: make(map[string]*Secret),
		token:   "mock-jwt-token-for-testing",
	}
}

func (m *mockKeyVaultServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Token endpoint - no auth required
	if r.URL.Path == "/token" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.token))
		return
	}

	// All other endpoints require auth
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "Unauthorized",
				"message": "Authentication required",
			},
		})
		return
	}

	// Verify token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token != m.token {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "Unauthorized",
				"message": "Invalid token",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Route handling
	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/secrets/"):
		m.handleGetSecret(w, r)
	case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/secrets/"):
		m.handleSetSecret(w, r)
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/secrets/"):
		m.handleDeleteSecret(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/secrets":
		m.handleListSecrets(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "NotFound",
				"message": "Resource not found",
			},
		})
	}
}

func (m *mockKeyVaultServer) handleGetSecret(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/secrets/")
	name = strings.Split(name, "?")[0] // Remove query params

	secret, exists := m.secrets[name]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "SecretNotFound",
				"message": "Secret not found: " + name,
			},
		})
		return
	}

	response := map[string]interface{}{
		"value": secret.Value,
		"id":    "https://localhost:4997/secrets/" + name,
		"attributes": map[string]interface{}{
			"enabled": secret.Enabled,
			"created": time.Now().Unix(),
			"updated": time.Now().Unix(),
		},
		"tags": secret.Tags,
	}

	json.NewEncoder(w).Encode(response)
}

func (m *mockKeyVaultServer) handleSetSecret(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/secrets/")
	name = strings.Split(name, "?")[0]

	var payload struct {
		Value      string            `json:"value"`
		Attributes map[string]bool   `json:"attributes"`
		Tags       map[string]string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "BadRequest",
				"message": err.Error(),
			},
		})
		return
	}

	now := time.Now()
	m.secrets[name] = &Secret{
		Name:      name,
		Value:     payload.Value,
		Enabled:   payload.Attributes["enabled"],
		Tags:      payload.Tags,
		CreatedOn: &now,
		UpdatedOn: &now,
	}

	response := map[string]interface{}{
		"value": payload.Value,
		"id":    "https://localhost:4997/secrets/" + name,
		"attributes": map[string]interface{}{
			"enabled": payload.Attributes["enabled"],
			"created": now.Unix(),
			"updated": now.Unix(),
		},
		"tags": payload.Tags,
	}

	json.NewEncoder(w).Encode(response)
}

func (m *mockKeyVaultServer) handleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/secrets/")
	name = strings.Split(name, "?")[0]

	if _, exists := m.secrets[name]; !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "SecretNotFound",
				"message": "Secret not found: " + name,
			},
		})
		return
	}

	delete(m.secrets, name)
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockKeyVaultServer) handleListSecrets(w http.ResponseWriter, r *http.Request) {
	var value []map[string]interface{}
	for name := range m.secrets {
		value = append(value, map[string]interface{}{
			"id": "https://localhost:4997/secrets/" + name,
		})
	}

	if value == nil {
		value = []map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"value": value,
	})
}

// =============================================================================
// Config Tests
// =============================================================================

func TestClientConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ClientConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: ClientConfig{
				VaultURL:           "https://localhost:4997",
				Timeout:            30 * time.Second,
				InsecureSkipVerify: true,
			},
			wantErr: false,
		},
		{
			name: "empty vault URL",
			config: ClientConfig{
				VaultURL: "",
				Timeout:  30 * time.Second,
			},
			wantErr: true,
			errMsg:  "VaultURL cannot be empty",
		},
		{
			name: "timeout too short",
			config: ClientConfig{
				VaultURL: "https://localhost:4997",
				Timeout:  5 * time.Second,
			},
			wantErr: true,
			errMsg:  "Timeout must be at least 10 seconds",
		},
		{
			name: "minimum timeout",
			config: ClientConfig{
				VaultURL: "https://localhost:4997",
				Timeout:  10 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, expected to contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestCachedClientConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  CachedClientConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: CachedClientConfig{
				KeyVault: ClientConfig{
					VaultURL: "https://localhost:4997",
					Timeout:  30 * time.Second,
				},
				Redis: RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
				CacheTTL:    5 * time.Minute,
				CachePrefix: "keyvault:",
			},
			wantErr: false,
		},
		{
			name: "invalid keyvault config",
			config: CachedClientConfig{
				KeyVault: ClientConfig{
					VaultURL: "",
					Timeout:  30 * time.Second,
				},
				Redis: RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
				CacheTTL: 5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "KeyVault config invalid",
		},
		{
			name: "empty redis host",
			config: CachedClientConfig{
				KeyVault: ClientConfig{
					VaultURL: "https://localhost:4997",
					Timeout:  30 * time.Second,
				},
				Redis: RedisConfig{
					Host: "",
					Port: 6379,
				},
				CacheTTL: 5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "Redis host cannot be empty",
		},
		{
			name: "invalid redis port",
			config: CachedClientConfig{
				KeyVault: ClientConfig{
					VaultURL: "https://localhost:4997",
					Timeout:  30 * time.Second,
				},
				Redis: RedisConfig{
					Host: "localhost",
					Port: 0,
				},
				CacheTTL: 5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "Redis port must be between 1 and 65535",
		},
		{
			name: "cache TTL too short",
			config: CachedClientConfig{
				KeyVault: ClientConfig{
					VaultURL: "https://localhost:4997",
					Timeout:  30 * time.Second,
				},
				Redis: RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
				CacheTTL: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "CacheTTL must be at least 60 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, expected to contain %v", err, tt.errMsg)
			}
		})
	}
}

// =============================================================================
// Client Tests with Mock Server
// =============================================================================

func setupTestClient(t *testing.T) (Client, *httptest.Server, func()) {
	mockServer := newMockKeyVaultServer()
	server := httptest.NewTLSServer(mockServer)

	appLogger, err := logger.NewProduction("keyvault-test", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	cfg := ClientConfig{
		VaultURL:           server.URL,
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	}

	client, err := NewClient(cfg, appLogger)
	if err != nil {
		server.Close()
		t.Fatalf("Failed to create client: %v", err)
	}

	cleanup := func() {
		client.Close(context.Background())
		server.Close()
		appLogger.Sync()
	}

	return client, server, cleanup
}

func TestClient_SetAndGetSecret(t *testing.T) {
	client, _, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Test SetSecret
	tags := map[string]string{
		"user_id": "user-123",
		"type":    "weather",
	}
	err := client.SetSecret(ctx, "test-secret", "secret-value-123", tags)
	if err != nil {
		t.Fatalf("SetSecret() error = %v", err)
	}

	// Test GetSecret
	secret, err := client.GetSecret(ctx, "test-secret")
	if err != nil {
		t.Fatalf("GetSecret() error = %v", err)
	}
	if secret == nil {
		t.Fatal("GetSecret() returned nil secret")
	}
	if secret.Value != "secret-value-123" {
		t.Errorf("GetSecret() value = %v, want %v", secret.Value, "secret-value-123")
	}
	if secret.Name != "test-secret" {
		t.Errorf("GetSecret() name = %v, want %v", secret.Name, "test-secret")
	}
}

func TestClient_GetSecret_NotFound(t *testing.T) {
	client, _, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Get non-existent secret
	secret, err := client.GetSecret(ctx, "non-existent")
	if err != nil {
		t.Fatalf("GetSecret() error = %v (expected nil for not found)", err)
	}
	if secret != nil {
		t.Errorf("GetSecret() expected nil for non-existent secret, got %v", secret)
	}
}

func TestClient_DeleteSecret(t *testing.T) {
	client, _, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Create a secret
	err := client.SetSecret(ctx, "to-delete", "value", nil)
	if err != nil {
		t.Fatalf("SetSecret() error = %v", err)
	}

	// Verify it exists
	secret, err := client.GetSecret(ctx, "to-delete")
	if err != nil {
		t.Fatalf("GetSecret() error = %v", err)
	}
	if secret == nil {
		t.Fatal("Secret should exist before delete")
	}

	// Delete it
	err = client.DeleteSecret(ctx, "to-delete")
	if err != nil {
		t.Fatalf("DeleteSecret() error = %v", err)
	}

	// Verify it's gone
	secret, err = client.GetSecret(ctx, "to-delete")
	if err != nil {
		t.Fatalf("GetSecret() after delete error = %v", err)
	}
	if secret != nil {
		t.Error("Secret should not exist after delete")
	}
}

func TestClient_ListSecrets(t *testing.T) {
	client, _, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Create some secrets
	secrets := []string{"user:123:weather", "user:123:alexa", "user:456:weather"}
	for _, name := range secrets {
		err := client.SetSecret(ctx, name, "value", nil)
		if err != nil {
			t.Fatalf("SetSecret(%s) error = %v", name, err)
		}
	}

	// List all secrets
	names, err := client.ListSecrets(ctx, "")
	if err != nil {
		t.Fatalf("ListSecrets() error = %v", err)
	}
	if len(names) != 3 {
		t.Errorf("ListSecrets() returned %d secrets, want 3", len(names))
	}

	// List with prefix
	names, err = client.ListSecrets(ctx, "user:123")
	if err != nil {
		t.Fatalf("ListSecrets(prefix) error = %v", err)
	}
	if len(names) != 2 {
		t.Errorf("ListSecrets(prefix) returned %d secrets, want 2", len(names))
	}
}

func TestClient_Health(t *testing.T) {
	client, _, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	err := client.Health(ctx)
	if err != nil {
		t.Errorf("Health() error = %v", err)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	client, _, cleanup := setupTestClient(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetSecret(ctx, "test")
	if err == nil {
		t.Error("GetSecret() with cancelled context should return error")
	}
}

// =============================================================================
// Secret Type Tests
// =============================================================================

func TestSecret_Fields(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name    string
		secret  Secret
		expired bool
	}{
		{
			name:    "no expiry set",
			secret:  Secret{Name: "test", Enabled: true},
			expired: false,
		},
		{
			name:    "expired",
			secret:  Secret{Name: "test", ExpiresOn: &past, Enabled: true},
			expired: true,
		},
		{
			name:    "not expired",
			secret:  Secret{Name: "test", ExpiresOn: &future, Enabled: true},
			expired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if expired by comparing with now
			isExpired := tt.secret.ExpiresOn != nil && tt.secret.ExpiresOn.Before(time.Now())
			if isExpired != tt.expired {
				t.Errorf("Expired check = %v, want %v", isExpired, tt.expired)
			}
		})
	}
}

// =============================================================================
// Integration Type Tests
// =============================================================================

func TestIntegrationType_Values(t *testing.T) {
	tests := []struct {
		intType IntegrationType
		want    string
	}{
		{IntegrationWeather, "weather"},
		{IntegrationGoogleHome, "google_home"},
		{IntegrationAlexa, "alexa"},
		{IntegrationIFTTT, "ifttt"},
		{IntegrationEnergy, "energy_provider"},
		{IntegrationSMS, "sms_notification"},
		{IntegrationMQTT, "mqtt_broker"},
		{IntegrationSmartThings, "smartthings"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := string(tt.intType); got != tt.want {
				t.Errorf("IntegrationType = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// UserIntegration Tests
// =============================================================================

func TestUserIntegration_Status(t *testing.T) {
	tests := []struct {
		name   string
		ui     UserIntegration
		active bool
	}{
		{
			name:   "connected",
			ui:     UserIntegration{Status: StatusConnected},
			active: true,
		},
		{
			name:   "not configured",
			ui:     UserIntegration{Status: StatusNotConfigured},
			active: false,
		},
		{
			name:   "error",
			ui:     UserIntegration{Status: StatusError},
			active: false,
		},
		{
			name:   "expired",
			ui:     UserIntegration{Status: StatusExpired},
			active: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isActive := tt.ui.Status == StatusConnected
			if isActive != tt.active {
				t.Errorf("IsActive = %v, want %v", isActive, tt.active)
			}
		})
	}
}

// =============================================================================
// Error Code Tests
// =============================================================================

func TestErrorCodes(t *testing.T) {
	// Verify error codes are defined
	codes := []string{
		ErrCodeConfigInvalid,
		ErrCodeVaultURLMissing,
		ErrCodeTimeoutInvalid,
		ErrCodeConnectionFailed,
		ErrCodeTLSError,
		ErrCodeHealthCheckFailed,
		ErrCodeSecretNotFound,
		ErrCodeSecretSetFailed,
		ErrCodeSecretGetFailed,
		ErrCodeSecretDeleteFailed,
		ErrCodeSecretListFailed,
		ErrCodeCacheWriteFailed,
		ErrCodeCacheReadFailed,
		ErrCodeCacheInvalidate,
		ErrCodeUserIntegrationNotFound,
		ErrCodeIntegrationExpired,
	}

	for _, code := range codes {
		if code == "" {
			t.Errorf("Error code should not be empty")
		}
		if !strings.HasPrefix(code, "INFRA-KEYVAULT-") {
			t.Errorf("Error code %s should have prefix INFRA-KEYVAULT-", code)
		}
	}
}
