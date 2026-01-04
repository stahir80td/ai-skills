package keyvault

import (
	"github.com/your-github-org/ai-scaffolder/core/go/errors"
)

// Error codes for KeyVault operations
// Following SOD (Severity × Occurrence × Detectability) scoring pattern
const (
	// Configuration errors
	ErrCodeConfigInvalid   = "INFRA-KEYVAULT-001"
	ErrCodeVaultURLMissing = "INFRA-KEYVAULT-002"
	ErrCodeTimeoutInvalid  = "INFRA-KEYVAULT-003"

	// Connection errors
	ErrCodeConnectionFailed  = "INFRA-KEYVAULT-010"
	ErrCodeTLSError          = "INFRA-KEYVAULT-011"
	ErrCodeHealthCheckFailed = "INFRA-KEYVAULT-012"

	// Operation errors
	ErrCodeSecretNotFound     = "INFRA-KEYVAULT-020"
	ErrCodeSecretSetFailed    = "INFRA-KEYVAULT-021"
	ErrCodeSecretGetFailed    = "INFRA-KEYVAULT-022"
	ErrCodeSecretDeleteFailed = "INFRA-KEYVAULT-023"
	ErrCodeSecretListFailed   = "INFRA-KEYVAULT-024"

	// Cache errors
	ErrCodeCacheWriteFailed = "INFRA-KEYVAULT-030"
	ErrCodeCacheReadFailed  = "INFRA-KEYVAULT-031"
	ErrCodeCacheInvalidate  = "INFRA-KEYVAULT-032"

	// Integration errors
	ErrCodeUserIntegrationNotFound = "INFRA-KEYVAULT-040"
	ErrCodeIntegrationExpired      = "INFRA-KEYVAULT-041"
)

// RegisterErrors registers all KeyVault error definitions with the error registry
func RegisterErrors(registry *errors.ErrorRegistry) {
	// Configuration errors
	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeConfigInvalid,
		Severity:    errors.SeverityHigh,
		Description: "KeyVault client configuration is invalid",
		SODScore:    72, // 8 × 3 × 3
		Severity_S:  8,
		Occurrence:  3,
		Detect_D:    3,
		Mitigation:  "Verify all required configuration parameters are set correctly",
		Example:     "Missing VaultURL or invalid timeout value",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeVaultURLMissing,
		Severity:    errors.SeverityHigh,
		Description: "KeyVault URL is not configured",
		SODScore:    48, // 8 × 2 × 3
		Severity_S:  8,
		Occurrence:  2,
		Detect_D:    3,
		Mitigation:  "Set KEYVAULT_URL environment variable or configure in helm values",
		Example:     "VaultURL is empty in configuration",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeTimeoutInvalid,
		Severity:    errors.SeverityMedium,
		Description: "KeyVault timeout configuration is below minimum threshold",
		SODScore:    24, // 4 × 2 × 3
		Severity_S:  4,
		Occurrence:  2,
		Detect_D:    3,
		Mitigation:  "Set timeout to at least 10 seconds",
		Example:     "Timeout set to 5s, minimum is 10s",
	})

	// Connection errors
	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeConnectionFailed,
		Severity:    errors.SeverityCritical,
		Description: "Failed to connect to KeyVault emulator",
		SODScore:    160, // 10 × 4 × 4
		Severity_S:  10,
		Occurrence:  4,
		Detect_D:    4,
		Mitigation:  "Verify KeyVault emulator is running and accessible at configured URL",
		Example:     "Connection refused to https://iot-keyvault:4997",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeTLSError,
		Severity:    errors.SeverityHigh,
		Description: "TLS certificate validation failed for KeyVault connection",
		SODScore:    96, // 8 × 4 × 3
		Severity_S:  8,
		Occurrence:  4,
		Detect_D:    3,
		Mitigation:  "Install KeyVault emulator certificates or enable InsecureSkipVerify for local dev",
		Example:     "x509: certificate signed by unknown authority",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeHealthCheckFailed,
		Severity:    errors.SeverityHigh,
		Description: "KeyVault health check failed",
		SODScore:    80, // 8 × 5 × 2
		Severity_S:  8,
		Occurrence:  5,
		Detect_D:    2,
		Mitigation:  "Check KeyVault emulator logs and network connectivity",
		Example:     "Health endpoint returned non-200 status",
	})

	// Operation errors
	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeSecretNotFound,
		Severity:    errors.SeverityMedium,
		Description: "Requested secret does not exist in KeyVault",
		SODScore:    24, // 4 × 3 × 2
		Severity_S:  4,
		Occurrence:  3,
		Detect_D:    2,
		Mitigation:  "Verify secret name is correct and has been created",
		Example:     "Secret 'user:123:weather-api-key' not found",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeSecretSetFailed,
		Severity:    errors.SeverityHigh,
		Description: "Failed to store secret in KeyVault",
		SODScore:    72, // 8 × 3 × 3
		Severity_S:  8,
		Occurrence:  3,
		Detect_D:    3,
		Mitigation:  "Check KeyVault permissions and emulator status",
		Example:     "Failed to set secret 'user:456:google-home-token'",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeSecretGetFailed,
		Severity:    errors.SeverityHigh,
		Description: "Failed to retrieve secret from KeyVault",
		SODScore:    80, // 8 × 5 × 2
		Severity_S:  8,
		Occurrence:  5,
		Detect_D:    2,
		Mitigation:  "Verify secret exists and KeyVault is accessible",
		Example:     "Network timeout while fetching secret",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeSecretDeleteFailed,
		Severity:    errors.SeverityMedium,
		Description: "Failed to delete secret from KeyVault",
		SODScore:    36, // 6 × 2 × 3
		Severity_S:  6,
		Occurrence:  2,
		Detect_D:    3,
		Mitigation:  "Verify secret exists and has proper permissions for deletion",
		Example:     "Delete operation failed for secret 'user:789:ifttt-webhook'",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeSecretListFailed,
		Severity:    errors.SeverityMedium,
		Description: "Failed to list secrets from KeyVault",
		SODScore:    30, // 5 × 3 × 2
		Severity_S:  5,
		Occurrence:  3,
		Detect_D:    2,
		Mitigation:  "Check KeyVault connectivity and list permissions",
		Example:     "Timeout while listing user integration secrets",
	})

	// Cache errors
	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeCacheWriteFailed,
		Severity:    errors.SeverityLow,
		Description: "Failed to write secret to Redis cache",
		SODScore:    12, // 3 × 2 × 2
		Severity_S:  3,
		Occurrence:  2,
		Detect_D:    2,
		Mitigation:  "Cache miss will occur, secret will be fetched directly from KeyVault",
		Example:     "Redis SET failed for keyvault:user:123:weather-api-key",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeCacheReadFailed,
		Severity:    errors.SeverityLow,
		Description: "Failed to read secret from Redis cache",
		SODScore:    12, // 3 × 2 × 2
		Severity_S:  3,
		Occurrence:  2,
		Detect_D:    2,
		Mitigation:  "Fallback to KeyVault fetch, slightly higher latency",
		Example:     "Redis GET failed for keyvault:user:456:alexa-token",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeCacheInvalidate,
		Severity:    errors.SeverityLow,
		Description: "Failed to invalidate cache entry after secret update",
		SODScore:    18, // 3 × 3 × 2
		Severity_S:  3,
		Occurrence:  3,
		Detect_D:    2,
		Mitigation:  "Stale cache entry will be served until TTL expires",
		Example:     "Redis DEL failed after updating secret",
	})

	// Integration errors
	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeUserIntegrationNotFound,
		Severity:    errors.SeverityMedium,
		Description: "User integration not configured",
		SODScore:    20, // 4 × 2.5 × 2
		Severity_S:  4,
		Occurrence:  3,
		Detect_D:    2,
		Mitigation:  "User needs to configure the integration in their profile settings",
		Example:     "User has not configured Google Home integration",
	})

	registry.Register(&errors.ErrorDefinition{
		Code:        ErrCodeIntegrationExpired,
		Severity:    errors.SeverityMedium,
		Description: "User integration token has expired",
		SODScore:    30, // 5 × 3 × 2
		Severity_S:  5,
		Occurrence:  3,
		Detect_D:    2,
		Mitigation:  "User needs to re-authenticate with the external service",
		Example:     "Google Home OAuth token expired 7 days ago",
	})
}
