using Serilog;
using Serilog.Context;
using Serilog.Events;
using Serilog.Formatting.Compact;

namespace Core.Logger;

/// <summary>
/// Structured logger with SRE capabilities including correlation ID tracking and component logging
/// </summary>
public class ServiceLogger : IDisposable
{
    private readonly ILogger _logger;
    private readonly string _serviceName;
    private bool _disposed;

    /// <summary>
    /// Context key for correlation IDs
    /// </summary>
    public const string CorrelationIdKey = "CorrelationId";

    /// <summary>
    /// Context key for component names
    /// </summary>
    public const string ComponentKey = "Component";

    /// <summary>
    /// Creates a new structured logger with consistent configuration
    /// </summary>
    public ServiceLogger(LoggerConfig config)
    {
        _serviceName = config.ServiceName;

        var loggerConfig = new LoggerConfiguration()
            .Enrich.WithProperty("Service", config.ServiceName)
            .Enrich.WithProperty("Environment", config.Environment)
            .Enrich.WithProperty("Version", config.Version)
            .Enrich.FromLogContext();

        // Set minimum level
        var level = config.LogLevel.ToLowerInvariant() switch
        {
            "verbose" => LogEventLevel.Verbose,
            "debug" => LogEventLevel.Debug,
            "information" or "info" => LogEventLevel.Information,
            "warning" or "warn" => LogEventLevel.Warning,
            "error" => LogEventLevel.Error,
            "fatal" => LogEventLevel.Fatal,
            _ => LogEventLevel.Information
        };
        loggerConfig.MinimumLevel.Is(level);

        // Configure output format
        if (config.OutputFormat.Equals("json", StringComparison.OrdinalIgnoreCase) ||
            config.Environment.Equals("production", StringComparison.OrdinalIgnoreCase))
        {
            loggerConfig.WriteTo.Console(new CompactJsonFormatter());
        }
        else
        {
            loggerConfig.WriteTo.Console(
                outputTemplate: "[{Timestamp:HH:mm:ss} {Level:u3}] {Message:lj} {Properties:j}{NewLine}{Exception}");
        }

        _logger = loggerConfig.CreateLogger();
    }

    /// <summary>
    /// Creates a production logger with standard settings
    /// </summary>
    public static ServiceLogger NewProduction(string serviceName, string version)
    {
        return new ServiceLogger(new LoggerConfig
        {
            ServiceName = serviceName,
            Environment = "production",
            Version = version,
            LogLevel = "Information",
            EnableCaller = true,
            EnableStacktrace = true,
            OutputFormat = "json"
        });
    }

    /// <summary>
    /// Creates a development logger with colorized console output
    /// </summary>
    public static ServiceLogger NewDevelopment(string serviceName, string version)
    {
        return new ServiceLogger(new LoggerConfig
        {
            ServiceName = serviceName,
            Environment = "development",
            Version = version,
            LogLevel = "Debug",
            EnableCaller = true,
            EnableStacktrace = true,
            OutputFormat = "console"
        });
    }

    /// <summary>
    /// Creates a context logger with correlation ID and component
    /// </summary>
    public ContextLogger WithContext(string? correlationId = null, string? component = null)
    {
        return new ContextLogger(_logger, correlationId, component);
    }

    /// <summary>
    /// Creates a logger scoped to a specific component
    /// </summary>
    public ContextLogger WithComponent(string component)
    {
        return new ContextLogger(_logger, null, component);
    }

    /// <summary>
    /// Logs a debug message
    /// </summary>
    public void Debug(string message, params object[] propertyValues)
    {
        _logger.Debug(message, propertyValues);
    }

    /// <summary>
    /// Logs an information message
    /// </summary>
    public void Information(string message, params object[] propertyValues)
    {
        _logger.Information(message, propertyValues);
    }

    /// <summary>
    /// Logs a warning message
    /// </summary>
    public void Warning(string message, params object[] propertyValues)
    {
        _logger.Warning(message, propertyValues);
    }

    /// <summary>
    /// Logs an error message
    /// </summary>
    public void Error(string message, params object[] propertyValues)
    {
        _logger.Error(message, propertyValues);
    }

    /// <summary>
    /// Logs an error with exception
    /// </summary>
    public void Error(Exception ex, string message, params object[] propertyValues)
    {
        _logger.Error(ex, message, propertyValues);
    }

    /// <summary>
    /// Logs a fatal error
    /// </summary>
    public void Fatal(string message, params object[] propertyValues)
    {
        _logger.Fatal(message, propertyValues);
    }

    /// <summary>
    /// Logs a fatal error with exception
    /// </summary>
    public void Fatal(Exception ex, string message, params object[] propertyValues)
    {
        _logger.Fatal(ex, message, propertyValues);
    }

    public void Dispose()
    {
        if (!_disposed)
        {
            (_logger as IDisposable)?.Dispose();
            _disposed = true;
        }
    }
}

/// <summary>
/// Logger with correlation context for request tracking
/// </summary>
public class ContextLogger
{
    private readonly ILogger _logger;
    private readonly string? _correlationId;
    private readonly string? _component;

    internal ContextLogger(ILogger logger, string? correlationId, string? component)
    {
        _logger = logger;
        _correlationId = correlationId;
        _component = component;
    }

    /// <summary>
    /// Creates a new context logger with updated component
    /// </summary>
    public ContextLogger WithComponent(string component)
    {
        return new ContextLogger(_logger, _correlationId, component);
    }

    /// <summary>
    /// Creates a new context logger with updated correlation ID
    /// </summary>
    public ContextLogger WithCorrelationId(string correlationId)
    {
        return new ContextLogger(_logger, correlationId, _component);
    }

    private IDisposable PushContext()
    {
        var disposables = new List<IDisposable>();

        if (!string.IsNullOrEmpty(_correlationId))
        {
            disposables.Add(LogContext.PushProperty(ServiceLogger.CorrelationIdKey, _correlationId));
        }

        if (!string.IsNullOrEmpty(_component))
        {
            disposables.Add(LogContext.PushProperty(ServiceLogger.ComponentKey, _component));
        }

        return new CompositeDisposable(disposables);
    }

    public void Debug(string message, params object[] propertyValues)
    {
        using (PushContext())
        {
            _logger.Debug(message, propertyValues);
        }
    }

    public void Information(string message, params object[] propertyValues)
    {
        using (PushContext())
        {
            _logger.Information(message, propertyValues);
        }
    }

    public void Warning(string message, params object[] propertyValues)
    {
        using (PushContext())
        {
            _logger.Warning(message, propertyValues);
        }
    }

    public void Error(string message, params object[] propertyValues)
    {
        using (PushContext())
        {
            _logger.Error(message, propertyValues);
        }
    }

    public void Error(Exception ex, string message, params object[] propertyValues)
    {
        using (PushContext())
        {
            _logger.Error(ex, message, propertyValues);
        }
    }

    public void Fatal(string message, params object[] propertyValues)
    {
        using (PushContext())
        {
            _logger.Fatal(message, propertyValues);
        }
    }

    public void Fatal(Exception ex, string message, params object[] propertyValues)
    {
        using (PushContext())
        {
            _logger.Fatal(ex, message, propertyValues);
        }
    }

    private class CompositeDisposable : IDisposable
    {
        private readonly List<IDisposable> _disposables;

        public CompositeDisposable(List<IDisposable> disposables)
        {
            _disposables = disposables;
        }

        public void Dispose()
        {
            foreach (var d in _disposables)
            {
                d.Dispose();
            }
        }
    }
}
