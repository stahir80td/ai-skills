# Core.Infrastructure Package - Comprehensive Infrastructure Clients

## Overview

The .NET Core.Infrastructure package provides production-grade infrastructure clients with standardized patterns, comprehensive error handling, and extensive health checking capabilities. This implementation matches the sophistication of the Go infrastructure package and integrates seamlessly with other Core packages.

## Infrastructure Components

### 1. SQL Server Client (`SqlServerClient.cs`)

**Features:**
- Connection pooling with automatic retry
- Query and command execution with parameter mapping
- Transaction support with automatic rollback
- Connection diagnostics and health checking
- Comprehensive error handling with Core.Errors integration

**Usage:**
```csharp
// Configuration
var config = new SqlServerConfig
{
    ConnectionString = "Server=localhost;Database=myapp;...",
    CommandTimeout = 30,
    EnableConnectionPooling = true,
    MaxPoolSize = 100
};

// Registration
services.AddSingleton<ISqlServerClient>(sp => 
    new SqlServerClient(config, sp.GetService<ServiceLogger>()));

// Usage
var result = await sqlClient.QueryAsync<Customer>(
    "SELECT * FROM customers WHERE city = @city", 
    new { city = "Seattle" });

// Health checking
var diagnostics = await sqlClient.DiagnoseConnectionAsync();
```

### 2. MongoDB Client (`MongoClient.cs`)

**Features:**
- Atlas and standalone connection support
- Document operations (CRUD)
- Aggregation pipeline support
- Health monitoring and statistics
- Comprehensive error handling

**Usage:**
```csharp
// Configuration
var config = new MongoConfig
{
    ConnectionString = "mongodb://localhost:27017",
    DatabaseName = "myapp",
    ConnectionTimeout = TimeSpan.FromSeconds(10)
};

// Registration
services.AddSingleton<IMongoClient>(sp => 
    new MongoClient(config, sp.GetService<ServiceLogger>()));

// Usage
var customer = await mongoClient.FindOneAsync<Customer>(
    "customers", 
    Builders<Customer>.Filter.Eq(c => c.Email, "user@example.com"));

// Aggregation
var pipeline = new BsonDocument[]
{
    new("$match", new BsonDocument("status", "active")),
    new("$group", new BsonDocument("_id", "$category"))
};
var results = await mongoClient.AggregateAsync<CustomerGroup>("customers", pipeline);
```

### 3. ScyllaDB Client (`ScyllaClient.cs`)

**Features:**
- Optimized for time-series data
- Prepared statement caching
- Batch operations support
- CQL execution with comprehensive error handling
- Metadata access and health checking

**Usage:**
```csharp
// Configuration
var config = new ScyllaConfig
{
    ContactPoints = new[] { "127.0.0.1" },
    Port = 9042,
    Keyspace = "timeseries",
    LocalDatacenter = "datacenter1"
};

// Registration
services.AddSingleton<IScyllaClient>(sp => 
    new ScyllaClient(config, sp.GetService<ServiceLogger>()));

// Usage
await scyllaClient.ExecuteAsync(
    "INSERT INTO sensor_data (device_id, timestamp, value) VALUES (?, ?, ?)",
    deviceId, DateTime.UtcNow, temperature);

// Prepared statements
var prepared = await scyllaClient.PrepareAsync(
    "SELECT * FROM sensor_data WHERE device_id = ? AND timestamp >= ?");
var rows = await scyllaClient.ExecutePreparedAsync(prepared, deviceId, startTime);
```

### 4. Kafka Producer (`KafkaClient.cs`)

**Features:**
- Enhanced message wrapper with metadata
- Correlation ID support for distributed tracing
- Comprehensive error handling and retry logic
- Health monitoring
- Production-ready configuration

**Usage:**
```csharp
// Configuration
var config = new KafkaConfig
{
    BootstrapServers = "localhost:9092",
    SecurityProtocol = SecurityProtocol.Plaintext,
    EnableIdempotence = true,
    MessageTimeoutMs = 30000
};

// Registration
services.AddSingleton<IKafkaProducer>(sp => 
    new KafkaProducer(config, sp.GetService<ServiceLogger>()));

// Usage
var message = new KafkaMessage<string>
{
    Value = JsonSerializer.Serialize(new OrderCreated { OrderId = orderId }),
    Key = orderId.ToString(),
    Headers = new Dictionary<string, string>
    {
        { "event-type", "OrderCreated" },
        { "correlation-id", correlationId }
    }
};

await kafkaProducer.ProduceAsync("orders.events", message);
```

### 5. Azure KeyVault Client (`KeyVaultClient.cs`)

**Features:**
- Multiple authentication methods (Managed Identity, Service Principal, Azure CLI)
- Local caching with TTL support
- Secret versioning and metadata access
- Batch operations for efficiency
- Comprehensive cache statistics

**Usage:**
```csharp
// Configuration
var config = new KeyVaultConfig
{
    VaultUrl = "https://myvault.vault.azure.net/",
    AuthMethod = KeyVaultAuthMethod.ManagedIdentity,
    CacheTtl = TimeSpan.FromMinutes(5),
    EnableCaching = true
};

// Registration
services.AddSingleton<IKeyVaultClient>(sp => 
    new KeyVaultClient(config, sp.GetService<ServiceLogger>()));

// Usage
var secret = await keyVaultClient.GetSecretAsync("database-connection-string");
var secrets = await keyVaultClient.GetSecretsAsync(new[] { "api-key", "jwt-secret" });

// Cache management
var stats = keyVaultClient.GetCacheStats();
keyVaultClient.ClearCache();
```

## Health Checking System

### Enhanced Health Checker (`EnhancedHealthChecker.cs`)

**Features:**
- Component-based health checking with detailed diagnostics
- Tag-based health check filtering
- System-wide health reporting with statistics
- Performance metrics (response times, success rates)
- Extensible component registration

**Usage:**
```csharp
// Registration
services.AddSingleton<ISystemHealthChecker, SystemHealthChecker>();

// Register components
healthChecker.RegisterComponent(new SqlServerHealthCheck(sqlClient));
healthChecker.RegisterComponent(new MongoHealthCheck(mongoClient));
healthChecker.RegisterComponent(new ScyllaHealthCheck(scyllaClient));
healthChecker.RegisterComponent(new RedisHealthCheck(redisClient));
healthChecker.RegisterComponent(new KafkaHealthCheck(kafkaProducer));
healthChecker.RegisterComponent(new KeyVaultHealthCheck(keyVaultClient));

// Check all components
var report = await healthChecker.CheckAllAsync();

// Check by category
var dbReport = await healthChecker.CheckByTagAsync("database");

// Check specific component
var sqlResult = await healthChecker.CheckComponentAsync("sqlserver");
```

### Integration with Existing HealthChecker

The existing `HealthChecker.cs` has been enhanced with extension methods for easy infrastructure component registration:

```csharp
// Automatic infrastructure registration
healthChecker.RegisterInfrastructure(serviceProvider);

// This automatically registers health checks for all available infrastructure clients
```

## Core Package Integration

All infrastructure clients integrate with Core packages for:

- **Core.Logger**: Structured logging with correlation IDs
- **Core.Errors**: Standardized error handling with error codes
- **Core.Config**: Configuration validation with `IValidatable`

**Error Handling Example:**
```csharp
// All clients throw standardized ServiceError exceptions
try
{
    await sqlClient.QueryAsync<Customer>("SELECT * FROM customers");
}
catch (ServiceError ex)
{
    // ex.Code: "SQL-001", "SQL-002", etc.
    // ex.Message: Human-readable description
    // ex.Details: Additional context (query, parameters, etc.)
    _logger.Warning("Database query failed", ex, new { 
        code = ex.Code,
        query = "SELECT * FROM customers"
    });
}
```

## Configuration Patterns

All infrastructure clients follow consistent configuration patterns:

```csharp
public class SqlServerConfig : IValidatable
{
    public string ConnectionString { get; set; } = "";
    public int CommandTimeout { get; set; } = 30;
    public bool EnableConnectionPooling { get; set; } = true;
    public int MaxPoolSize { get; set; } = 100;
    
    public void Validate()
    {
        if (string.IsNullOrWhiteSpace(ConnectionString))
            throw new InvalidOperationException("ConnectionString is required");
        
        if (CommandTimeout <= 0)
            throw new InvalidOperationException("CommandTimeout must be positive");
    }
}
```

## Production Readiness Features

### 1. Comprehensive Logging
- Entry/exit logging for all operations
- Performance metrics (execution times)
- Error context with full details
- Correlation ID propagation

### 2. Error Handling
- Standardized error codes (SQL-001, MONGO-002, etc.)
- Detailed error context
- Automatic retry for transient failures
- Circuit breaker patterns (where applicable)

### 3. Health Monitoring
- Component-specific health checks
- Performance metrics collection
- System-wide health reporting
- Tag-based filtering for different probe types

### 4. Performance Optimization
- Connection pooling
- Prepared statement caching
- Local caching (KeyVault)
- Batch operations support

## Example Service Registration

```csharp
// Program.cs - Complete infrastructure setup
public static void Main(string[] args)
{
    var builder = WebApplication.CreateBuilder(args);
    
    // Core packages
    builder.Services.AddSingleton<ServiceLogger>();
    
    // SQL Server
    builder.Services.AddSingleton<ISqlServerClient>(sp =>
        new SqlServerClient(
            builder.Configuration.GetSection("SqlServer").Get<SqlServerConfig>(),
            sp.GetRequiredService<ServiceLogger>()));
    
    // MongoDB
    builder.Services.AddSingleton<IMongoClient>(sp =>
        new MongoClient(
            builder.Configuration.GetSection("MongoDB").Get<MongoConfig>(),
            sp.GetRequiredService<ServiceLogger>()));
    
    // ScyllaDB
    builder.Services.AddSingleton<IScyllaClient>(sp =>
        new ScyllaClient(
            builder.Configuration.GetSection("ScyllaDB").Get<ScyllaConfig>(),
            sp.GetRequiredService<ServiceLogger>()));
    
    // Kafka
    builder.Services.AddSingleton<IKafkaProducer>(sp =>
        new KafkaProducer(
            builder.Configuration.GetSection("Kafka").Get<KafkaConfig>(),
            sp.GetRequiredService<ServiceLogger>()));
    
    // KeyVault
    builder.Services.AddSingleton<IKeyVaultClient>(sp =>
        new KeyVaultClient(
            builder.Configuration.GetSection("KeyVault").Get<KeyVaultConfig>(),
            sp.GetRequiredService<ServiceLogger>()));
    
    // Enhanced health checking
    builder.Services.AddSingleton<ISystemHealthChecker, SystemHealthChecker>();
    
    // Traditional health checking with infrastructure
    builder.Services.AddSingleton(sp =>
    {
        var healthChecker = new HealthChecker(TimeSpan.FromMinutes(1));
        healthChecker.RegisterInfrastructure(sp);
        return healthChecker;
    });
    
    var app = builder.Build();
    
    // Health endpoints
    app.MapHealthChecks("/health/live");
    app.MapHealthChecks("/health/ready");
    
    // Enhanced health endpoint
    app.MapGet("/health/detailed", async (ISystemHealthChecker healthChecker) =>
        await healthChecker.CheckAllAsync());
    
    app.Run();
}
```

## Conclusion

The .NET Core.Infrastructure package now provides comprehensive infrastructure capabilities matching the Go package sophistication. All clients integrate seamlessly with Core packages for logging, error handling, and configuration validation. The enhanced health checking system provides detailed insights into system health with component-level diagnostics and system-wide reporting.

This implementation follows production-ready patterns with comprehensive error handling, performance optimization, and extensive observability features.