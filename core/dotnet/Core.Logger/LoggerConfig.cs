namespace Core.Logger;

/// <summary>
/// Configuration for the structured logger
/// </summary>
public class LoggerConfig
{
    /// <summary>
    /// Name of the service for log context
    /// </summary>
    public string ServiceName { get; set; } = "unknown-service";

    /// <summary>
    /// Environment (development, staging, production)
    /// </summary>
    public string Environment { get; set; } = "development";

    /// <summary>
    /// Version of the service
    /// </summary>
    public string Version { get; set; } = "1.0.0";

    /// <summary>
    /// Minimum log level (Verbose, Debug, Information, Warning, Error, Fatal)
    /// </summary>
    public string LogLevel { get; set; } = "Information";

    /// <summary>
    /// Include caller information (file:line)
    /// </summary>
    public bool EnableCaller { get; set; } = true;

    /// <summary>
    /// Include stacktrace for errors
    /// </summary>
    public bool EnableStacktrace { get; set; } = true;

    /// <summary>
    /// Output format: "json" or "console"
    /// </summary>
    public string OutputFormat { get; set; } = "json";
}
