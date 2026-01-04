package keyvault

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.uber.org/zap"
)

// Client interface for KeyVault operations
type Client interface {
	// GetSecret retrieves a secret by name
	GetSecret(ctx context.Context, name string) (*Secret, error)

	// SetSecret stores or updates a secret
	SetSecret(ctx context.Context, name string, value string, tags map[string]string) error

	// DeleteSecret removes a secret
	DeleteSecret(ctx context.Context, name string) error

	// ListSecrets returns all secret names matching a prefix
	ListSecrets(ctx context.Context, prefix string) ([]string, error)

	// Health checks if KeyVault is accessible
	Health(ctx context.Context) error

	// Close releases any resources
	Close(ctx context.Context) error
}

// client implements the Client interface for Azure KeyVault Emulator
type client struct {
	httpClient *http.Client
	vaultURL   string
	logger     *logger.ContextLogger
	timeout    time.Duration

	// Authentication
	token       string
	tokenExpiry time.Time
	tokenMu     sync.RWMutex
}

// NewClient creates a new KeyVault client
func NewClient(cfg ClientConfig, log *logger.Logger) (Client, error) {
	componentLogger := log.WithComponent("KeyVaultClient")

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		componentLogger.Error("Invalid configuration",
			zap.Error(err),
			zap.String("error_code", ErrCodeConfigInvalid))
		return nil, fmt.Errorf("invalid keyvault config: %w", err)
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	if cfg.TLSConfig.CertPath != "" && cfg.TLSConfig.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSConfig.CertPath, cfg.TLSConfig.KeyPath)
		if err != nil {
			componentLogger.Error("Failed to load TLS certificates",
				zap.Error(err),
				zap.String("error_code", ErrCodeTLSError),
				zap.String("cert_path", cfg.TLSConfig.CertPath))
			return nil, fmt.Errorf("failed to load TLS certificates: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Create HTTP client with custom transport
	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	componentLogger.Info("KeyVault client initialized",
		zap.String("vault_url", cfg.VaultURL),
		zap.Duration("timeout", cfg.Timeout),
		zap.Bool("tls_skip_verify", cfg.InsecureSkipVerify))

	c := &client{
		httpClient: httpClient,
		vaultURL:   strings.TrimSuffix(cfg.VaultURL, "/"),
		logger:     componentLogger,
		timeout:    cfg.Timeout,
	}

	// Fetch initial authentication token
	if err := c.refreshToken(context.Background()); err != nil {
		componentLogger.Error("Failed to fetch authentication token",
			zap.Error(err),
			zap.String("error_code", ErrCodeConnectionFailed),
			zap.String("vault_url", cfg.VaultURL))
		return nil, fmt.Errorf("failed to fetch keyvault token: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := c.Health(ctx); err != nil {
		componentLogger.Error("KeyVault health check failed on initialization",
			zap.Error(err),
			zap.String("error_code", ErrCodeConnectionFailed),
			zap.String("vault_url", cfg.VaultURL))
		return nil, fmt.Errorf("keyvault health check failed: %w", err)
	}

	componentLogger.Info("Successfully connected to KeyVault emulator",
		zap.String("vault_url", cfg.VaultURL),
		zap.String("status", "healthy"))

	return c, nil
}

// refreshToken fetches a new bearer token from the emulator's /token endpoint
func (c *client) refreshToken(ctx context.Context) error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Check if token is still valid (with 1 minute buffer)
	if c.token != "" && time.Now().Add(time.Minute).Before(c.tokenExpiry) {
		return nil
	}

	url := fmt.Sprintf("%s/token", c.vaultURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	tokenBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response: %w", err)
	}

	c.token = strings.TrimSpace(string(tokenBytes))
	// Token is valid for 24 hours according to emulator, but refresh more often
	c.tokenExpiry = time.Now().Add(12 * time.Hour)

	c.logger.Debug("Authentication token refreshed",
		zap.Time("expires_at", c.tokenExpiry))

	return nil
}

// getToken returns a valid bearer token, refreshing if necessary
func (c *client) getToken(ctx context.Context) (string, error) {
	c.tokenMu.RLock()
	if c.token != "" && time.Now().Add(time.Minute).Before(c.tokenExpiry) {
		token := c.token
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()

	// Need to refresh
	if err := c.refreshToken(ctx); err != nil {
		return "", err
	}

	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token, nil
}

// GetSecret retrieves a secret by name from KeyVault
func (c *client) GetSecret(ctx context.Context, name string) (*Secret, error) {
	start := time.Now()

	// Get authentication token
	token, err := c.getToken(ctx)
	if err != nil {
		c.logger.Error("Failed to get authentication token",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretGetFailed))
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Azure KeyVault API: GET {vaultUri}/secrets/{secret-name}?api-version=7.4
	url := fmt.Sprintf("%s/secrets/%s?api-version=7.4", c.vaultURL, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error("Failed to create request",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretGetFailed))
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute request",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretGetFailed),
			zap.Duration("duration", time.Since(start)))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.logger.Debug("Secret not found",
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretNotFound),
			zap.Duration("duration", time.Since(start)))
		return nil, nil // Return nil, nil for not found (not an error)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("KeyVault returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("secret_name", name),
			zap.String("response", string(body)),
			zap.String("error_code", ErrCodeSecretGetFailed))
		return nil, fmt.Errorf("keyvault returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Azure KeyVault response format
	var kvResponse struct {
		Value      string `json:"value"`
		ID         string `json:"id"`
		Attributes struct {
			Enabled bool   `json:"enabled"`
			Created int64  `json:"created"`
			Updated int64  `json:"updated"`
			Exp     *int64 `json:"exp,omitempty"`
			Nbf     *int64 `json:"nbf,omitempty"`
		} `json:"attributes"`
		Tags map[string]string `json:"tags"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&kvResponse); err != nil {
		c.logger.Error("Failed to decode response",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretGetFailed))
		return nil, err
	}

	secret := &Secret{
		Name:    name,
		Value:   kvResponse.Value,
		Enabled: kvResponse.Attributes.Enabled,
		Tags:    kvResponse.Tags,
	}

	// Convert Unix timestamps to time.Time
	if kvResponse.Attributes.Created > 0 {
		t := time.Unix(kvResponse.Attributes.Created, 0)
		secret.CreatedOn = &t
	}
	if kvResponse.Attributes.Updated > 0 {
		t := time.Unix(kvResponse.Attributes.Updated, 0)
		secret.UpdatedOn = &t
	}
	if kvResponse.Attributes.Exp != nil {
		t := time.Unix(*kvResponse.Attributes.Exp, 0)
		secret.ExpiresOn = &t
	}
	if kvResponse.Attributes.Nbf != nil {
		t := time.Unix(*kvResponse.Attributes.Nbf, 0)
		secret.NotBefore = &t
	}

	c.logger.Debug("Secret retrieved successfully",
		zap.String("secret_name", name),
		zap.Bool("enabled", secret.Enabled),
		zap.Duration("duration", time.Since(start)))

	return secret, nil
}

// SetSecret stores or updates a secret in KeyVault
func (c *client) SetSecret(ctx context.Context, name string, value string, tags map[string]string) error {
	start := time.Now()

	// Get authentication token
	token, err := c.getToken(ctx)
	if err != nil {
		c.logger.Error("Failed to get authentication token",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretSetFailed))
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Azure KeyVault API: PUT {vaultUri}/secrets/{secret-name}?api-version=7.4
	url := fmt.Sprintf("%s/secrets/%s?api-version=7.4", c.vaultURL, name)

	payload := map[string]interface{}{
		"value": value,
		"attributes": map[string]interface{}{
			"enabled": true,
		},
	}
	if tags != nil {
		payload["tags"] = tags
	}

	body, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error("Failed to marshal request body",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretSetFailed))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(body)))
	if err != nil {
		c.logger.Error("Failed to create request",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretSetFailed))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute request",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretSetFailed),
			zap.Duration("duration", time.Since(start)))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		c.logger.Error("KeyVault returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("secret_name", name),
			zap.String("response", string(respBody)),
			zap.String("error_code", ErrCodeSecretSetFailed))
		return fmt.Errorf("keyvault returned status %d: %s", resp.StatusCode, string(respBody))
	}

	c.logger.Info("Secret stored successfully",
		zap.String("secret_name", name),
		zap.Int("value_length", len(value)),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// DeleteSecret removes a secret from KeyVault
func (c *client) DeleteSecret(ctx context.Context, name string) error {
	start := time.Now()

	// Get authentication token
	token, err := c.getToken(ctx)
	if err != nil {
		c.logger.Error("Failed to get authentication token",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretDeleteFailed))
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Azure KeyVault API: DELETE {vaultUri}/secrets/{secret-name}?api-version=7.4
	url := fmt.Sprintf("%s/secrets/%s?api-version=7.4", c.vaultURL, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		c.logger.Error("Failed to create request",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretDeleteFailed))
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute request",
			zap.Error(err),
			zap.String("secret_name", name),
			zap.String("error_code", ErrCodeSecretDeleteFailed),
			zap.Duration("duration", time.Since(start)))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		c.logger.Error("KeyVault returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("secret_name", name),
			zap.String("response", string(respBody)),
			zap.String("error_code", ErrCodeSecretDeleteFailed))
		return fmt.Errorf("keyvault returned status %d: %s", resp.StatusCode, string(respBody))
	}

	c.logger.Info("Secret deleted successfully",
		zap.String("secret_name", name),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// ListSecrets returns all secret names matching a prefix
func (c *client) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	start := time.Now()

	// Get authentication token
	token, err := c.getToken(ctx)
	if err != nil {
		c.logger.Error("Failed to get authentication token",
			zap.Error(err),
			zap.String("prefix", prefix),
			zap.String("error_code", ErrCodeSecretListFailed))
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Azure KeyVault API: GET {vaultUri}/secrets?api-version=7.4
	url := fmt.Sprintf("%s/secrets?api-version=7.4", c.vaultURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error("Failed to create request",
			zap.Error(err),
			zap.String("prefix", prefix),
			zap.String("error_code", ErrCodeSecretListFailed))
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute request",
			zap.Error(err),
			zap.String("prefix", prefix),
			zap.String("error_code", ErrCodeSecretListFailed),
			zap.Duration("duration", time.Since(start)))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.logger.Error("KeyVault returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("prefix", prefix),
			zap.String("response", string(respBody)),
			zap.String("error_code", ErrCodeSecretListFailed))
		return nil, fmt.Errorf("keyvault returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse list response
	var listResponse struct {
		Value []struct {
			ID string `json:"id"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		c.logger.Error("Failed to decode response",
			zap.Error(err),
			zap.String("prefix", prefix),
			zap.String("error_code", ErrCodeSecretListFailed))
		return nil, err
	}

	// Extract secret names and filter by prefix
	var names []string
	for _, item := range listResponse.Value {
		// ID format: {vaultUri}/secrets/{name}
		parts := strings.Split(item.ID, "/secrets/")
		if len(parts) == 2 {
			name := parts[1]
			if prefix == "" || strings.HasPrefix(name, prefix) {
				names = append(names, name)
			}
		}
	}

	c.logger.Debug("Secrets listed successfully",
		zap.String("prefix", prefix),
		zap.Int("count", len(names)),
		zap.Duration("duration", time.Since(start)))

	return names, nil
}

// Health checks if KeyVault is accessible
func (c *client) Health(ctx context.Context) error {
	start := time.Now()

	// Get authentication token
	token, err := c.getToken(ctx)
	if err != nil {
		c.logger.Error("Health check failed - authentication error",
			zap.Error(err),
			zap.String("error_code", ErrCodeHealthCheckFailed))
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Try to list secrets as a health check
	url := fmt.Sprintf("%s/secrets?api-version=7.4&maxresults=1", c.vaultURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error("Health check failed - request creation error",
			zap.Error(err),
			zap.String("error_code", ErrCodeHealthCheckFailed))
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Health check failed - connection error",
			zap.Error(err),
			zap.String("error_code", ErrCodeHealthCheckFailed),
			zap.Duration("duration", time.Since(start)))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Health check failed - non-200 status",
			zap.Int("status_code", resp.StatusCode),
			zap.String("error_code", ErrCodeHealthCheckFailed),
			zap.Duration("duration", time.Since(start)))
		return fmt.Errorf("keyvault health check returned status %d", resp.StatusCode)
	}

	c.logger.Debug("Health check passed",
		zap.Duration("duration", time.Since(start)))

	return nil
}

// Close releases any resources
func (c *client) Close(ctx context.Context) error {
	c.httpClient.CloseIdleConnections()
	c.logger.Info("KeyVault client closed")
	return nil
}
