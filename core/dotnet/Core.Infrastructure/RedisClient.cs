using Microsoft.Extensions.Diagnostics.HealthChecks;
using StackExchange.Redis;
using System.Text.Json;

namespace Core.Infrastructure;

/// <summary>
/// Redis client configuration
/// </summary>
public class RedisConfig
{
    /// <summary>Redis host</summary>
    public string Host { get; set; } = "localhost";

    /// <summary>Redis port</summary>
    public int Port { get; set; } = 6379;

    /// <summary>Connection timeout (minimum 60s)</summary>
    public TimeSpan ConnectTimeout { get; set; } = TimeSpan.FromSeconds(60);

    /// <summary>Operation timeout</summary>
    public TimeSpan OperationTimeout { get; set; } = TimeSpan.FromSeconds(5);

    /// <summary>Connection string (overrides Host/Port if set)</summary>
    public string? ConnectionString { get; set; }
}

/// <summary>
/// Redis client interface
/// </summary>
public interface IRedisClient : IAsyncDisposable
{
    Task<string?> GetAsync(string key);
    Task<T?> GetAsync<T>(string key) where T : class;
    Task SetAsync(string key, string value, TimeSpan? expiry = null);
    Task SetAsync<T>(string key, T value, TimeSpan? expiry = null) where T : class;
    Task<bool> DeleteAsync(string key);
    Task<bool> ExistsAsync(string key);
    Task<long> IncrementAsync(string key, long value = 1);
    Task<bool> SetAddAsync(string key, string value);
    Task<bool> SetRemoveAsync(string key, string value);
    Task<string[]> SetMembersAsync(string key);
    Task<HealthCheckResult> HealthCheckAsync();
}

/// <summary>
/// Redis client implementation using StackExchange.Redis
/// </summary>
public class RedisClient : IRedisClient
{
    private readonly ConnectionMultiplexer _connection;
    private readonly IDatabase _database;
    private readonly string _host;
    private readonly int _port;

    /// <summary>
    /// Creates a new Redis client
    /// </summary>
    public static async Task<RedisClient> CreateAsync(RedisConfig config)
    {
        if (string.IsNullOrEmpty(config.Host))
            throw new ArgumentException("Redis host cannot be empty");

        if (config.Port <= 0 || config.Port > 65535)
            throw new ArgumentException($"Redis port must be between 1 and 65535, got {config.Port}");

        var connectionString = config.ConnectionString ??
            $"{config.Host}:{config.Port},connectTimeout={(int)config.ConnectTimeout.TotalMilliseconds}";

        var options = ConfigurationOptions.Parse(connectionString);
        options.ConnectTimeout = (int)config.ConnectTimeout.TotalMilliseconds;
        options.SyncTimeout = (int)config.OperationTimeout.TotalMilliseconds;

        var connection = await ConnectionMultiplexer.ConnectAsync(options);
        return new RedisClient(connection, config.Host, config.Port);
    }

    private RedisClient(ConnectionMultiplexer connection, string host, int port)
    {
        _connection = connection;
        _database = connection.GetDatabase();
        _host = host;
        _port = port;
    }

    public async Task<string?> GetAsync(string key)
    {
        var value = await _database.StringGetAsync(key);
        return value.HasValue ? value.ToString() : null;
    }

    public async Task<T?> GetAsync<T>(string key) where T : class
    {
        var value = await GetAsync(key);
        if (string.IsNullOrEmpty(value)) return null;
        return JsonSerializer.Deserialize<T>(value);
    }

    public async Task SetAsync(string key, string value, TimeSpan? expiry = null)
    {
        await _database.StringSetAsync(key, value, expiry);
    }

    public async Task SetAsync<T>(string key, T value, TimeSpan? expiry = null) where T : class
    {
        var json = JsonSerializer.Serialize(value);
        await SetAsync(key, json, expiry);
    }

    public async Task<bool> DeleteAsync(string key)
    {
        return await _database.KeyDeleteAsync(key);
    }

    public async Task<bool> ExistsAsync(string key)
    {
        return await _database.KeyExistsAsync(key);
    }

    public async Task<long> IncrementAsync(string key, long value = 1)
    {
        return await _database.StringIncrementAsync(key, value);
    }

    public async Task<bool> SetAddAsync(string key, string value)
    {
        return await _database.SetAddAsync(key, value);
    }

    public async Task<bool> SetRemoveAsync(string key, string value)
    {
        return await _database.SetRemoveAsync(key, value);
    }

    public async Task<string[]> SetMembersAsync(string key)
    {
        var members = await _database.SetMembersAsync(key);
        return members.Select(m => m.ToString()).ToArray();
    }

    public async Task<HealthCheckResult> HealthCheckAsync()
    {
        try
        {
            var latency = await _database.PingAsync();
            return HealthCheckResult.Healthy($"Redis connected to {_host}:{_port}, latency: {latency.TotalMilliseconds}ms");
        }
        catch (Exception ex)
        {
            return HealthCheckResult.Unhealthy($"Redis connection failed: {ex.Message}", ex);
        }
    }

    public async ValueTask DisposeAsync()
    {
        await _connection.CloseAsync();
        _connection.Dispose();
    }
}
