using Core.Config;
using Core.Logger;
using Microsoft.Extensions.Diagnostics.HealthChecks;
using Microsoft.Data.SqlClient;
using Dapper;

namespace Core.Infrastructure.SqlServer;

/// <summary>
/// SQL Server configuration
/// </summary>
public class SqlServerConfig : IValidatable
{
    /// <summary>Connection string</summary>
    public string ConnectionString { get; set; } = "";
    
    /// <summary>Command timeout in seconds (minimum 60s)</summary>
    public int CommandTimeout { get; set; } = 60;
    
    /// <summary>Enable connection pooling</summary>
    public bool EnableConnectionPooling { get; set; } = true;
    
    /// <summary>Maximum pool size</summary>
    public int MaxPoolSize { get; set; } = 25;
    
    /// <summary>Maximum idle connections</summary>
    public int MaxIdleConnections { get; set; } = 5;
    
    /// <summary>Connection lifetime in minutes</summary>
    public int ConnectionLifetimeMinutes { get; set; } = 5;
    
    /// <summary>Health check timeout in seconds</summary>
    public int HealthCheckTimeoutSeconds { get; set; } = 5;
    
    /// <summary>Connection timeout in seconds (minimum 60s)</summary>
    public int ConnectionTimeout { get; set; } = 60;

    public ValidationResult Validate()
    {
        var errors = new List<string>();
        
        if (string.IsNullOrWhiteSpace(ConnectionString))
            errors.Add("ConnectionString is required");
            
        if (CommandTimeout < 60)
            errors.Add("CommandTimeout must be at least 60 seconds");
            
        if (ConnectionTimeout < 60)
            errors.Add("ConnectionTimeout must be at least 60 seconds");
            
        if (MaxPoolSize <= 0)
            errors.Add("MaxPoolSize must be positive");
            
        if (MaxIdleConnections <= 0)
            errors.Add("MaxIdleConnections must be positive");
        
        return errors.Any() ? ValidationResult.Failed(errors.ToArray()) : ValidationResult.Success();
    }
}

/// <summary>
/// Connection diagnostics information
/// </summary>
/// <summary>
/// SQL Server client interface
/// </summary>
public interface ISqlServerClient
{
    /// <summary>Execute a query and return results</summary>
    Task<IEnumerable<T>> QueryAsync<T>(string sql, object? parameters = null, CancellationToken cancellationToken = default);
    
    /// <summary>Execute a command and return affected rows</summary>
    Task<int> ExecuteAsync(string sql, object? parameters = null, CancellationToken cancellationToken = default);
    
    /// <summary>Check connection health with diagnostics</summary>
    Task<Core.Infrastructure.ConnectionDiagnostics> DiagnoseConnectionAsync(CancellationToken cancellationToken = default);
    
    /// <summary>Simple health check</summary>
    Task<bool> HealthAsync(CancellationToken cancellationToken = default);
}

/// <summary>
/// Simple SQL Server client implementation
/// </summary>
public class SqlServerClient : ISqlServerClient
{
    private readonly SqlServerConfig _config;
    private readonly ServiceLogger _logger;
    private readonly string _componentName = "SQLServerClient";

    public SqlServerClient(SqlServerConfig config, ServiceLogger logger)
    {
        _config = config ?? throw new ArgumentNullException(nameof(config));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
        
        var validation = config.Validate();
        if (!validation.IsValid)
        {
            _logger.Error("Invalid SQL Server configuration", new { 
                errors = validation.Errors,
                errorCode = "INFRA-SQLSERVER-CONFIG-ERROR",
                component = _componentName
            });
            throw new InvalidOperationException($"Invalid configuration: {string.Join(", ", validation.Errors)}");
        }
            
        _logger.Information("SqlServerClient initialized", new { 
            component = _componentName,
            maxPoolSize = config.MaxPoolSize,
            maxIdleConnections = config.MaxIdleConnections,
            connectionTimeout = config.ConnectionTimeout,
            status = "healthy"
        });
    }

    public async Task<IEnumerable<T>> QueryAsync<T>(string sql, object? parameters = null, CancellationToken cancellationToken = default)
    {
        _logger.Information("Executing SQL query", new { 
            component = _componentName,
            sql = sql?.Substring(0, Math.Min(100, sql.Length)) + (sql?.Length > 100 ? "..." : ""),
            hasParameters = parameters != null
        });
        
        try
        {
            using var connection = new SqlConnection(_config.ConnectionString);
            await connection.OpenAsync(cancellationToken);
            
            // Use Dapper for proper ORM mapping with parameter binding
            var commandDefinition = new CommandDefinition(
                sql, 
                parameters, 
                commandTimeout: _config.CommandTimeout,
                cancellationToken: cancellationToken
            );
            
            var results = await connection.QueryAsync<T>(commandDefinition);
            var resultList = results.ToList();
            
            _logger.Information("SQL query executed successfully", new { 
                component = _componentName,
                resultCount = resultList.Count
            });
            
            return resultList;
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "SQL query failed", new { 
                component = _componentName,
                sql = sql?.Substring(0, Math.Min(100, sql.Length)),
                errorCode = "INFRA-SQLSERVER-QUERY-ERROR"
            });
            throw;
        }
    }

    public async Task<int> ExecuteAsync(string sql, object? parameters = null, CancellationToken cancellationToken = default)
    {
        _logger.Information("Executing SQL command", new { 
            component = _componentName,
            sql = sql?.Substring(0, Math.Min(100, sql.Length)) + (sql?.Length > 100 ? "..." : ""),
            hasParameters = parameters != null
        });
        
        try
        {
            using var connection = new SqlConnection(_config.ConnectionString);
            await connection.OpenAsync(cancellationToken);
            
            // Use Dapper for proper parameter binding
            var commandDefinition = new CommandDefinition(
                sql, 
                parameters, 
                commandTimeout: _config.CommandTimeout,
                cancellationToken: cancellationToken
            );
            
            var rowsAffected = await connection.ExecuteAsync(commandDefinition);
            
            _logger.Information("SQL command executed successfully", new { 
                component = _componentName,
                rowsAffected
            });
            
            return rowsAffected;
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "SQL command failed", new { 
                component = _componentName,
                sql = sql?.Substring(0, Math.Min(100, sql.Length)),
                errorCode = "INFRA-SQLSERVER-COMMAND-ERROR"
            });
            throw;
        }
    }

    public async Task<Core.Infrastructure.ConnectionDiagnostics> DiagnoseConnectionAsync(CancellationToken cancellationToken = default)
    {
        var stopwatch = System.Diagnostics.Stopwatch.StartNew();
        
        _logger.Information("Starting SQL Server connection diagnostic", new { 
            component = _componentName 
        });
        
        try
        {
            using var connection = new SqlConnection(_config.ConnectionString);
            await connection.OpenAsync(cancellationToken);
            
            var serverVersion = connection.ServerVersion;
            var databaseName = connection.Database;
            
            stopwatch.Stop();
            
            _logger.Information("SQL Server connection diagnostic completed", new { 
                component = _componentName,
                databaseName, 
                serverVersion, 
                connectionTimeMs = stopwatch.ElapsedMilliseconds,
                status = "healthy"
            });
            
            return new Core.Infrastructure.ConnectionDiagnostics
            {
                IsHealthy = true,
                DatabaseName = databaseName,
                ServerVersion = serverVersion,
                ConnectionTime = stopwatch.Elapsed
            };
        }
        catch (Exception ex)
        {
            stopwatch.Stop();
            
            _logger.Warning("SQL Server connection diagnostic failed", ex, new {
                component = _componentName,
                errorCode = "INFRA-SQLSERVER-CONNECT-ERROR",
                connectionTimeMs = stopwatch.ElapsedMilliseconds
            });
            
            return new Core.Infrastructure.ConnectionDiagnostics
            {
                IsHealthy = false,
                ConnectionTime = stopwatch.Elapsed,
                ErrorMessage = ex.Message
            };
        }
    }

    public async Task<bool> HealthAsync(CancellationToken cancellationToken = default)
    {
        var diagnostics = await DiagnoseConnectionAsync(cancellationToken);
        return diagnostics.IsHealthy;
    }
}