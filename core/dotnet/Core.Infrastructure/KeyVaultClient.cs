using Core.Config;
using Core.Logger;
using Azure.Security.KeyVault.Secrets;
using Azure.Identity;

namespace Core.Infrastructure.KeyVault;

/// <summary>
/// KeyVault authentication method
/// </summary>
public enum KeyVaultAuthMethod
{
    /// <summary>Use Managed Identity (recommended for production)</summary>
    ManagedIdentity,
    /// <summary>Use Azure CLI credentials (good for development)</summary>
    AzureCli,
    /// <summary>Use Service Principal</summary>
    ServicePrincipal
}

/// <summary>
/// Azure KeyVault configuration
/// </summary>
public class KeyVaultConfig : IValidatable
{
    /// <summary>KeyVault URL</summary>
    public string VaultUrl { get; set; } = "";
    
    /// <summary>Authentication method</summary>
    public KeyVaultAuthMethod AuthMethod { get; set; } = KeyVaultAuthMethod.ManagedIdentity;
    
    /// <summary>Service Principal Client ID (if using ServicePrincipal auth)</summary>
    public string? ClientId { get; set; }
    
    /// <summary>Service Principal Client Secret (if using ServicePrincipal auth)</summary>
    public string? ClientSecret { get; set; }
    
    /// <summary>Service Principal Tenant ID (if using ServicePrincipal auth)</summary>
    public string? TenantId { get; set; }
    
    /// <summary>Cache TTL for secrets</summary>
    public TimeSpan CacheTtl { get; set; } = TimeSpan.FromMinutes(5);
    
    /// <summary>Enable local caching of secrets</summary>
    public bool EnableCaching { get; set; } = true;
    
    /// <summary>Health check timeout in seconds</summary>
    public int HealthCheckTimeoutSeconds { get; set; } = 10;

    public ValidationResult Validate()
    {
        var errors = new List<string>();
        
        if (string.IsNullOrWhiteSpace(VaultUrl))
            errors.Add("VaultUrl is required");
            
        if (AuthMethod == KeyVaultAuthMethod.ServicePrincipal)
        {
            if (string.IsNullOrWhiteSpace(ClientId))
                errors.Add("ClientId is required for ServicePrincipal authentication");
            if (string.IsNullOrWhiteSpace(ClientSecret))
                errors.Add("ClientSecret is required for ServicePrincipal authentication");
            if (string.IsNullOrWhiteSpace(TenantId))
                errors.Add("TenantId is required for ServicePrincipal authentication");
        }
        
        return errors.Any() ? ValidationResult.Failed(errors.ToArray()) : ValidationResult.Success();
    }
}

/// <summary>
/// Secret with metadata
/// </summary>
public class SecretWithMetadata
{
    /// <summary>Secret name</summary>
    public string Name { get; set; } = "";
    
    /// <summary>Secret value</summary>
    public string Value { get; set; } = "";
    
    /// <summary>Secret version</summary>
    public string Version { get; set; } = "";
    
    /// <summary>When the secret was created</summary>
    public DateTime? CreatedOn { get; set; }
    
    /// <summary>When the secret was updated</summary>
    public DateTime? UpdatedOn { get; set; }
}

/// <summary>
/// Cache statistics
/// </summary>
public class CacheStats
{
    /// <summary>Number of cached secrets</summary>
    public int CachedSecrets { get; set; }
    
    /// <summary>Cache hit ratio (0.0 to 1.0)</summary>
    public double HitRatio { get; set; }
}

/// <summary>
/// KeyVault client interface
/// </summary>
public interface IKeyVaultClient
{
    /// <summary>Get a secret value</summary>
    Task<string?> GetSecretAsync(string secretName, CancellationToken cancellationToken = default);
    
    /// <summary>Get a secret with metadata</summary>
    Task<SecretWithMetadata?> GetSecretWithMetadataAsync(string secretName, CancellationToken cancellationToken = default);
    
    /// <summary>Set a secret value</summary>
    Task SetSecretAsync(string secretName, string value, CancellationToken cancellationToken = default);
    
    /// <summary>Check KeyVault health</summary>
    Task<bool> HealthAsync(CancellationToken cancellationToken = default);
    
    /// <summary>Get cache statistics</summary>
    CacheStats GetCacheStats();
    
    /// <summary>Clear the local cache</summary>
    void ClearCache();
}

/// <summary>
/// Simple Azure KeyVault client implementation
/// </summary>
public class KeyVaultClient : IKeyVaultClient
{
    private readonly SecretClient _secretClient;
    private readonly ServiceLogger _logger;
    private readonly KeyVaultConfig _config;
    private readonly string _componentName = "keyvault";
    private readonly Dictionary<string, (SecretWithMetadata Secret, DateTime CachedAt)> _cache = new();
    private readonly object _cacheLock = new object();
    private int _cacheHits = 0;
    private int _cacheRequests = 0;

    public KeyVaultClient(KeyVaultConfig config, ServiceLogger logger)
    {
        _config = config ?? throw new ArgumentNullException(nameof(config));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
        
        var validation = config.Validate();
        if (!validation.IsValid)
            throw new InvalidOperationException($"Invalid configuration: {string.Join(", ", validation.Errors)}");

        var credential = CreateCredential(config);
        _secretClient = new SecretClient(new Uri(config.VaultUrl), credential);
        
        _logger.Information("KeyVaultClient initialized", new { 
            vaultUrl = config.VaultUrl,
            authMethod = config.AuthMethod.ToString(),
            cachingEnabled = config.EnableCaching
        });
    }

    private Azure.Core.TokenCredential CreateCredential(KeyVaultConfig config)
    {
        return config.AuthMethod switch
        {
            KeyVaultAuthMethod.ManagedIdentity => new ManagedIdentityCredential(),
            KeyVaultAuthMethod.AzureCli => new AzureCliCredential(),
            KeyVaultAuthMethod.ServicePrincipal => new ClientSecretCredential(
                config.TenantId!, 
                config.ClientId!, 
                config.ClientSecret!),
            _ => throw new ArgumentException($"Unsupported auth method: {config.AuthMethod}")
        };
    }

    public async Task<string?> GetSecretAsync(string secretName, CancellationToken cancellationToken = default)
    {
        var secretWithMetadata = await GetSecretWithMetadataAsync(secretName, cancellationToken);
        return secretWithMetadata?.Value;
    }

    public async Task<SecretWithMetadata?> GetSecretWithMetadataAsync(string secretName, CancellationToken cancellationToken = default)
    {
        if (string.IsNullOrWhiteSpace(secretName))
            throw new ArgumentException("Secret name cannot be empty", nameof(secretName));

        _logger.Information("Getting KeyVault secret", new { secretName });

        // Check cache first if caching is enabled
        if (_config.EnableCaching)
        {
            lock (_cacheLock)
            {
                _cacheRequests++;
                if (_cache.TryGetValue(secretName, out var cached))
                {
                    if (DateTime.UtcNow - cached.CachedAt < _config.CacheTtl)
                    {
                        _cacheHits++;
                        _logger.Information("KeyVault secret cache hit", new { secretName });
                        return cached.Secret;
                    }
                    else
                    {
                        _cache.Remove(secretName);
                    }
                }
            }
        }

        try
        {
            var response = await _secretClient.GetSecretAsync(secretName, null, cancellationToken);
            var secret = response.Value;
            
            var result = new SecretWithMetadata
            {
                Name = secret.Name,
                Value = secret.Value,
                Version = secret.Properties.Version,
                CreatedOn = secret.Properties.CreatedOn?.DateTime,
                UpdatedOn = secret.Properties.UpdatedOn?.DateTime
            };

            // Cache the result if caching is enabled
            if (_config.EnableCaching)
            {
                lock (_cacheLock)
                {
                    _cache[secretName] = (result, DateTime.UtcNow);
                }
            }

            _logger.Information("KeyVault secret retrieved successfully", new { 
                secretName, 
                version = result.Version 
            });

            return result;
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "Failed to get KeyVault secret", new { secretName });
            throw;
        }
    }

    public async Task SetSecretAsync(string secretName, string value, CancellationToken cancellationToken = default)
    {
        if (string.IsNullOrWhiteSpace(secretName))
            throw new ArgumentException("Secret name cannot be empty", nameof(secretName));
            
        if (string.IsNullOrWhiteSpace(value))
            throw new ArgumentException("Secret value cannot be empty", nameof(value));

        _logger.Information("Setting KeyVault secret", new { secretName });

        try
        {
            await _secretClient.SetSecretAsync(secretName, value, cancellationToken);

            // Invalidate cache entry if caching is enabled
            if (_config.EnableCaching)
            {
                lock (_cacheLock)
                {
                    _cache.Remove(secretName);
                }
            }

            _logger.Information("KeyVault secret set successfully", new { secretName });
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "Failed to set KeyVault secret", new { secretName });
            throw;
        }
    }

    public async Task<bool> HealthAsync(CancellationToken cancellationToken = default)
    {
        try
        {
            _logger.Debug("Performing KeyVault health check", new { 
                component = _componentName 
            });
            
            // Simple health check with timeout - try to list secrets (with minimal permissions required)
            using var cts = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
            cts.CancelAfter(TimeSpan.FromSeconds(_config.HealthCheckTimeoutSeconds));
            
            var properties = _secretClient.GetPropertiesOfSecretsAsync(cts.Token);
            await foreach (var _ in properties)
            {
                // Just check if we can access the vault
                break;
            }
            
            _logger.Debug("KeyVault health check passed", new { 
                component = _componentName
            });
            
            return true;
        }
        catch (Exception ex)
        {
            _logger.Warning("KeyVault health check failed", ex, new {
                component = _componentName,
                errorCode = "INFRA-KEYVAULT-HEALTH-ERROR"
            });
            return false;
        }
    }

    public CacheStats GetCacheStats()
    {
        lock (_cacheLock)
        {
            return new CacheStats
            {
                CachedSecrets = _cache.Count,
                HitRatio = _cacheRequests > 0 ? (double)_cacheHits / _cacheRequests : 0.0
            };
        }
    }

    public void ClearCache()
    {
        lock (_cacheLock)
        {
            var count = _cache.Count;
            _cache.Clear();
            _cacheHits = 0;
            _cacheRequests = 0;
            
            _logger.Information("KeyVault cache cleared", new { clearedSecrets = count });
        }
    }
}