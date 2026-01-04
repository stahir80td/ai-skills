using Core.Config;
using Core.Logger;
using MongoDB.Driver;
using MongoDB.Bson;

// Alias for the MongoDB driver client to avoid naming conflict
using MongoDBClient = MongoDB.Driver.MongoClient;

namespace Core.Infrastructure.MongoDB;

/// <summary>
/// MongoDB configuration
/// </summary>
public class MongoConfig : IValidatable
{
    /// <summary>Connection string</summary>
    public string ConnectionString { get; set; } = "mongodb://localhost:27017";
    
    /// <summary>Database name</summary>
    public string DatabaseName { get; set; } = "myapp";
    
    /// <summary>Connection timeout (minimum 60s)</summary>
    public TimeSpan ConnectionTimeout { get; set; } = TimeSpan.FromSeconds(60);
    
    /// <summary>Server selection timeout (minimum 60s)</summary>
    public TimeSpan ServerSelectionTimeout { get; set; } = TimeSpan.FromSeconds(60);
    
    /// <summary>Socket timeout (minimum 60s)</summary>
    public TimeSpan SocketTimeout { get; set; } = TimeSpan.FromSeconds(60);
    
    /// <summary>Maximum connection pool size</summary>
    public int MaxConnectionPoolSize { get; set; } = 100;
    
    /// <summary>Health check timeout in seconds</summary>
    public int HealthCheckTimeoutSeconds { get; set; } = 10;

    public ValidationResult Validate()
    {
        var errors = new List<string>();
        
        if (string.IsNullOrWhiteSpace(ConnectionString))
            errors.Add("ConnectionString is required");
            
        if (string.IsNullOrWhiteSpace(DatabaseName))
            errors.Add("DatabaseName is required");
            
        if (ConnectionTimeout < TimeSpan.FromSeconds(60))
            errors.Add("ConnectionTimeout must be at least 60 seconds");
            
        if (ServerSelectionTimeout < TimeSpan.FromSeconds(60))
            errors.Add("ServerSelectionTimeout must be at least 60 seconds");
            
        if (SocketTimeout < TimeSpan.FromSeconds(60))
            errors.Add("SocketTimeout must be at least 60 seconds");
        
        return errors.Any() ? ValidationResult.Failed(errors.ToArray()) : ValidationResult.Success();
    }
}

/// <summary>
/// MongoDB client interface
/// </summary>
public interface IMongoClient
{
    /// <summary>Find a single document</summary>
    Task<T?> FindOneAsync<T>(string collectionName, FilterDefinition<T> filter, CancellationToken cancellationToken = default);
    
    /// <summary>Insert a document</summary>
    Task InsertOneAsync<T>(string collectionName, T document, CancellationToken cancellationToken = default);
    
    /// <summary>Update a document</summary>
    Task<bool> UpdateOneAsync<T>(string collectionName, FilterDefinition<T> filter, UpdateDefinition<T> update, CancellationToken cancellationToken = default);
    
    /// <summary>Delete a document</summary>
    Task<bool> DeleteOneAsync<T>(string collectionName, FilterDefinition<T> filter, CancellationToken cancellationToken = default);
    
    /// <summary>Check database health</summary>
    Task<bool> HealthAsync(CancellationToken cancellationToken = default);
    
    /// <summary>Get database statistics</summary>
    Task<BsonDocument> GetStatsAsync(CancellationToken cancellationToken = default);
}

/// <summary>
/// Simple MongoDB client implementation
/// </summary>
public class MongoClient : IMongoClient
{
    private readonly MongoDBClient _client;
    private readonly IMongoDatabase _database;
    private readonly ServiceLogger _logger;
    private readonly MongoConfig _config;
    private readonly string _componentName = "mongodb";

    public MongoClient(MongoConfig config, ServiceLogger logger)
    {
        _config = config ?? throw new ArgumentNullException(nameof(config));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
        
        var validation = config.Validate();
        if (!validation.IsValid)
        {
            _logger.Error("Invalid MongoDB configuration", new { 
                errors = validation.Errors,
                errorCode = "INFRA-MONGODB-CONFIG-ERROR",
                component = _componentName
            });
            throw new InvalidOperationException($"Invalid configuration: {string.Join(", ", validation.Errors)}");
        }

        // Create client settings with enhanced timeouts
        var settings = MongoClientSettings.FromConnectionString(_config.ConnectionString);
        settings.ConnectTimeout = config.ConnectionTimeout;
        settings.ServerSelectionTimeout = config.ServerSelectionTimeout;
        settings.SocketTimeout = config.SocketTimeout;
        settings.MaxConnectionPoolSize = config.MaxConnectionPoolSize;
        
        _client = new MongoDBClient(settings);
        _database = _client.GetDatabase(_config.DatabaseName);
        
        _logger.Information("MongoClient initialized", new { 
            component = _componentName,
            databaseName = config.DatabaseName,
            connectionTimeout = config.ConnectionTimeout,
            serverSelectionTimeout = config.ServerSelectionTimeout,
            maxPoolSize = config.MaxConnectionPoolSize,
            status = "healthy"
        });
    }

    public async Task<T?> FindOneAsync<T>(string collectionName, FilterDefinition<T> filter, CancellationToken cancellationToken = default)
    {
        _logger.Information("Finding MongoDB document", new { 
            component = _componentName,
            collectionName,
            documentType = typeof(T).Name
        });
        
        try
        {
            var collection = _database.GetCollection<T>(collectionName);
            var result = await collection.Find(filter).FirstOrDefaultAsync(cancellationToken);
            
            _logger.Information("MongoDB find completed", new { 
                component = _componentName,
                collectionName,
                found = result != null
            });
            
            return result;
        }
        catch (Exception ex)
        {
            _logger.Warning("Failed to find MongoDB document", ex, new { 
                component = _componentName,
                collectionName,
                errorCode = "INFRA-MONGODB-FIND-ERROR"
            });
            throw;
        }
    }

    public async Task InsertOneAsync<T>(string collectionName, T document, CancellationToken cancellationToken = default)
    {
        _logger.Information("Inserting MongoDB document", new { collectionName });
        
        try
        {
            var collection = _database.GetCollection<T>(collectionName);
            await collection.InsertOneAsync(document, null, cancellationToken);
            
            _logger.Information("MongoDB document inserted successfully", new { collectionName });
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "Failed to insert MongoDB document", new { collectionName });
            throw;
        }
    }

    public async Task<bool> UpdateOneAsync<T>(string collectionName, FilterDefinition<T> filter, UpdateDefinition<T> update, CancellationToken cancellationToken = default)
    {
        _logger.Information("Updating MongoDB document", new { collectionName });
        
        try
        {
            var collection = _database.GetCollection<T>(collectionName);
            var result = await collection.UpdateOneAsync(filter, update, null, cancellationToken);
            
            var success = result.ModifiedCount > 0;
            _logger.Information("MongoDB document update completed", new { collectionName, modifiedCount = result.ModifiedCount });
            
            return success;
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "Failed to update MongoDB document", new { collectionName });
            throw;
        }
    }

    public async Task<bool> DeleteOneAsync<T>(string collectionName, FilterDefinition<T> filter, CancellationToken cancellationToken = default)
    {
        _logger.Information("Deleting MongoDB document", new { collectionName });
        
        try
        {
            var collection = _database.GetCollection<T>(collectionName);
            var result = await collection.DeleteOneAsync(filter, null, cancellationToken);
            
            var success = result.DeletedCount > 0;
            _logger.Information("MongoDB document delete completed", new { collectionName, deletedCount = result.DeletedCount });
            
            return success;
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "Failed to delete MongoDB document", new { collectionName });
            throw;
        }
    }

    public async Task<bool> HealthAsync(CancellationToken cancellationToken = default)
    {
        try
        {
            _logger.Debug("Performing MongoDB health check", new { 
                component = _componentName 
            });
            
            // Simple ping operation with timeout
            using var cts = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
            cts.CancelAfter(TimeSpan.FromSeconds(_config.HealthCheckTimeoutSeconds));
            
            await _database.RunCommandAsync<BsonDocument>(new BsonDocument("ping", 1), null, cts.Token);
            
            _logger.Debug("MongoDB health check passed", new { 
                component = _componentName
            });
            
            return true;
        }
        catch (Exception ex)
        {
            _logger.Warning("MongoDB health check failed", ex, new {
                component = _componentName,
                errorCode = "INFRA-MONGODB-HEALTH-ERROR"
            });
            return false;
        }
    }

    public async Task<BsonDocument> GetStatsAsync(CancellationToken cancellationToken = default)
    {
        try
        {
            var stats = await _database.RunCommandAsync<BsonDocument>(new BsonDocument("dbStats", 1), null, cancellationToken);
            return stats;
        }
        catch (Exception ex)
        {
            _logger.Warning("Failed to get MongoDB stats", ex);
            return new BsonDocument();
        }
    }
}