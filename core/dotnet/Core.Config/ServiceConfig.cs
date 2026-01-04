namespace Core.Config;

/// <summary>
/// Validation result for configuration validation
/// </summary>
public class ValidationResult
{
    /// <summary>Whether validation passed</summary>
    public bool IsValid { get; set; }
    
    /// <summary>List of validation errors</summary>
    public List<string> Errors { get; set; } = new();
    
    /// <summary>Create a successful validation result</summary>
    public static ValidationResult Success() => new() { IsValid = true };
    
    /// <summary>Create a failed validation result with errors</summary>
    public static ValidationResult Failed(params string[] errors) => new() 
    { 
        IsValid = false, 
        Errors = errors.ToList() 
    };
}

/// <summary>
/// Interface for configuration classes that can be validated
/// </summary>
public interface IValidatable
{
    /// <summary>Validate the configuration</summary>
    ValidationResult Validate();
}

/// <summary>
/// Standard timeout configurations for infrastructure components
/// All timeouts are configurable - NO HARDCODING
/// </summary>
public class TimeoutConfig : IValidatable
{
    /// <summary>
    /// Redis connection/operation timeout (minimum: 60s)
    /// </summary>
    public TimeSpan RedisTimeout { get; set; } = TimeSpan.FromSeconds(60);

    /// <summary>
    /// SQL Server connection timeout (minimum: 60s)
    /// </summary>
    public TimeSpan SqlTimeout { get; set; } = TimeSpan.FromSeconds(60);

    /// <summary>
    /// MongoDB connection timeout (minimum: 60s)
    /// </summary>
    public TimeSpan MongoDbTimeout { get; set; } = TimeSpan.FromSeconds(60);

    /// <summary>
    /// Kafka producer/consumer timeout (minimum: 60s)
    /// </summary>
    public TimeSpan KafkaTimeout { get; set; } = TimeSpan.FromSeconds(60);

    /// <summary>
    /// HTTP client timeout for external calls (minimum: 30s)
    /// </summary>
    public TimeSpan HttpTimeout { get; set; } = TimeSpan.FromSeconds(30);

    /// <summary>
    /// Health check timeout (minimum: 60s)
    /// </summary>
    public TimeSpan HealthCheckTimeout { get; set; } = TimeSpan.FromSeconds(60);

    /// <summary>
    /// Graceful shutdown timeout
    /// </summary>
    public TimeSpan ShutdownTimeout { get; set; } = TimeSpan.FromSeconds(30);

    /// <summary>
    /// Validates timeout configurations meet minimum requirements
    /// </summary>
    public ValidationResult Validate()
    {
        var errors = new List<string>();
        
        if (RedisTimeout < TimeSpan.FromSeconds(60))
            errors.Add("RedisTimeout must be at least 60 seconds");
        if (SqlTimeout < TimeSpan.FromSeconds(60))
            errors.Add("SqlTimeout must be at least 60 seconds");
        if (MongoDbTimeout < TimeSpan.FromSeconds(60))
            errors.Add("MongoDbTimeout must be at least 60 seconds");
        if (KafkaTimeout < TimeSpan.FromSeconds(60))
            errors.Add("KafkaTimeout must be at least 60 seconds");
        if (HealthCheckTimeout < TimeSpan.FromSeconds(60))
            errors.Add("HealthCheckTimeout must be at least 60 seconds");
        
        return errors.Any() ? ValidationResult.Failed(errors.ToArray()) : ValidationResult.Success();
    }
}

/// <summary>
/// Service configuration base class
/// </summary>
public class ServiceConfig : IValidatable
{
    /// <summary>
    /// Service name for logging and metrics
    /// </summary>
    public string ServiceName { get; set; } = "unknown-service";

    /// <summary>
    /// Service version
    /// </summary>
    public string Version { get; set; } = "1.0.0";

    /// <summary>
    /// Environment (development, staging, production)
    /// </summary>
    public string Environment { get; set; } = "development";

    /// <summary>
    /// Timeout configurations
    /// </summary>
    public TimeoutConfig Timeouts { get; set; } = new();

    /// <summary>
    /// Whether to enable detailed debug logging
    /// </summary>
    public bool EnableDebugLogging { get; set; }

    /// <summary>
    /// Whether this is a production environment
    /// </summary>
    public bool IsProduction => Environment.Equals("production", StringComparison.OrdinalIgnoreCase);

    /// <summary>
    /// Validate the service configuration
    /// </summary>
    public virtual ValidationResult Validate()
    {
        var errors = new List<string>();
        
        if (string.IsNullOrWhiteSpace(ServiceName))
            errors.Add("ServiceName is required");
            
        if (string.IsNullOrWhiteSpace(Version))
            errors.Add("Version is required");
            
        // Validate nested timeouts configuration
        var timeoutResult = Timeouts.Validate();
        if (!timeoutResult.IsValid)
            errors.AddRange(timeoutResult.Errors);
        
        return errors.Any() ? ValidationResult.Failed(errors.ToArray()) : ValidationResult.Success();
    }
}
