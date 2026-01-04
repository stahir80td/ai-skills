---
name: ai-core-packages-dotnet
description: >
  CRITICAL enforcement of AI Core NuGet packages for .NET development.
  Use this skill whenever generating, reviewing, or modifying .NET code in AI projects.
  Ensures developers use Core.Logger, Core.Errors, Core.Infrastructure instead of raw
  packages like Serilog, StackExchange.Redis, Confluent.Kafka, MongoDB.Driver.
  This skill MUST be applied to all .NET code generation, scaffolding, and refactoring.
---

# 🚨 AI Core Package Enforcement (.NET)

## ABSOLUTE RULES - NO EXCEPTIONS

### Required NuGet Packages (EVERY .csproj)

```xml
<ItemGroup>
  <PackageReference Include="Core.Config" Version="1.0.5" />
  <PackageReference Include="Core.Errors" Version="1.0.5" />
  <PackageReference Include="Core.Logger" Version="1.0.5" />
  <PackageReference Include="Core.Metrics" Version="1.0.5" />
  <PackageReference Include="Core.Sli" Version="1.0.5" />
  <PackageReference Include="Core.Infrastructure" Version="1.0.5" />
  <PackageReference Include="Core.Reliability" Version="1.0.5" />
</ItemGroup>
```

### Required NuGet.config (Project Root)

```xml
<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" />
    <add key="github" value="https://nuget.pkg.github.com/your-github-org/index.json" />
  </packageSources>
  <packageSourceCredentials>
    <github>
      <add key="Username" value="ai-user" />
      <add key="ClearTextPassword" value="%GITHUB_TOKEN%" />
    </github>
  </packageSourceCredentials>
</configuration>
```

---

## ❌ FORBIDDEN - NEVER ADD THESE PACKAGES

If you see yourself adding any of these, STOP and use the Core equivalent:

| ❌ FORBIDDEN Package | ✅ Use Instead |
|---------------------|----------------|
| `Serilog` | `Core.Logger` → `ServiceLogger` |
| `Serilog.Sinks.*` | Already in Core.Logger |
| `Microsoft.Extensions.Logging` | `Core.Logger` → `ServiceLogger` |
| `NLog` | `Core.Logger` → `ServiceLogger` |
| `StackExchange.Redis` | `Core.Infrastructure` → `IRedisClient` |
| `Confluent.Kafka` | `Core.Infrastructure` → `IKafkaProducer` |
| `MongoDB.Driver` | `Core.Infrastructure` → `IMongoClient` |
| `Microsoft.Data.SqlClient` | `Core.Infrastructure` → `ISqlServerClient` |
| `Polly` (raw) | `Core.Reliability` → `CircuitBreaker`, `RetryPolicy` |
| `prometheus-net` | `Core.Metrics` → `ServiceMetrics` |

---

## ❌ FORBIDDEN Code Patterns

```csharp
// ❌ WRONG - Direct Serilog
using Serilog;
Log.Information("message");

// ❌ WRONG - ILogger<T>
private readonly ILogger<MyService> _logger;
_logger.LogInformation("message");

// ❌ WRONG - Direct Redis
using StackExchange.Redis;
var redis = ConnectionMultiplexer.Connect("localhost");

// ❌ WRONG - Direct Kafka
using Confluent.Kafka;
var producer = new ProducerBuilder<string, string>(config).Build();

// ❌ WRONG - Direct MongoDB
using MongoDB.Driver;
var client = new MongoClient(connectionString);

// ❌ WRONG - Generic exceptions
throw new Exception("Order not found");
throw new InvalidOperationException("Bad state");
```

---

## ✅ CORRECT Code Patterns

```csharp
// ✅ CORRECT - ServiceLogger from Core.Logger
using Core.Logger;

public class OrderService
{
    private readonly ServiceLogger _logger;
    
    public OrderService(ServiceLogger logger)
    {
        _logger = logger;
    }
    
    public async Task<Order> CreateOrder(CreateOrderRequest request)
    {
        _logger.Information("Creating order for customer {CustomerId}", request.CustomerId);
    }
}
```

```csharp
// ✅ CORRECT - Error codes from Core.Errors
using Core.Errors;

public static class OrderErrors
{
    private static readonly ErrorRegistry _registry = new();
    
    static OrderErrors()
    {
        _registry.Register("ORD-001", "Order not found", HttpStatusCode.NotFound);
        _registry.Register("ORD-002", "Invalid status", HttpStatusCode.Conflict);
    }
    
    public static ServiceError NotFound(Guid id) => 
        _registry.CreateError("ORD-001", new { orderId = id });
}

// Usage: throw OrderErrors.NotFound(orderId);
```

```csharp
// ✅ CORRECT - Infrastructure from Core.Infrastructure
using Core.Infrastructure;
using Core.Infrastructure.Kafka;
using Core.Infrastructure.MongoDB;

public class OrderRepository
{
    private readonly ISqlServerClient _sql;
    private readonly IRedisClient _redis;
    private readonly IKafkaProducer _kafka;
    private readonly Core.Infrastructure.MongoDB.IMongoClient _mongo;
    
    public OrderRepository(
        ISqlServerClient sql, 
        IRedisClient redis, 
        IKafkaProducer kafka,
        Core.Infrastructure.MongoDB.IMongoClient mongo)
    {
        _sql = sql;
        _redis = redis;
        _kafka = kafka;
        _mongo = mongo;
    }
}
```

---

## Program.cs DI Registration Pattern

```csharp
using Core.Logger;
using Core.Infrastructure;
using Core.Infrastructure.Kafka;
using Core.Infrastructure.MongoDB;
using Core.Sli;

var builder = WebApplication.CreateBuilder(args);

// Core.Logger
builder.Services.AddSingleton<ServiceLogger>(sp => 
    new ServiceLogger("order-service", builder.Configuration));

// Core.Infrastructure - SQL Server
builder.Services.AddSingleton<ISqlServerClient>(sp =>
    new SqlServerClient(new SqlServerConfig 
    { 
        ConnectionString = builder.Configuration.GetConnectionString("SqlServer")! 
    }));

// Core.Infrastructure - Redis
builder.Services.AddSingleton<IRedisClient>(sp =>
    new RedisClient(new RedisConfig 
    { 
        ConnectionString = builder.Configuration.GetConnectionString("Redis")! 
    }));

// Core.Infrastructure - Kafka
builder.Services.AddSingleton<IKafkaProducer>(sp =>
    new KafkaProducer(new KafkaConfig 
    { 
        BootstrapServers = builder.Configuration["Kafka:BootstrapServers"]! 
    }));

// Core.Infrastructure - MongoDB
builder.Services.AddSingleton<Core.Infrastructure.MongoDB.IMongoClient>(sp =>
    new Core.Infrastructure.MongoDB.MongoClient(new MongoConfig
    {
        ConnectionString = builder.Configuration.GetConnectionString("MongoDB")!,
        DatabaseName = builder.Configuration["MongoDB:Database"]!
    }));

// Core.Sli
builder.Services.AddSingleton<SliTracker>();
```

---

## Pre-Generation Checklist

Before outputting ANY .NET code, verify:

- [ ] NuGet.config exists with GitHub Packages source
- [ ] .csproj has ALL Core.* package references
- [ ] NO forbidden packages in .csproj
- [ ] Using `ServiceLogger` not `ILogger<T>` or `Log.*`
- [ ] Using `ServiceError` with codes, not generic exceptions
- [ ] Using `IKafkaProducer` not `Confluent.Kafka`
- [ ] Using `IRedisClient` not `StackExchange.Redis`
- [ ] Using `IMongoClient` not `MongoDB.Driver`
- [ ] Using `ISqlServerClient` not raw `SqlConnection`

---

## Reference Files (Read Before Generating)

When generating .NET services, read these files first:
- `patterns/dotnet/AiPatterns/Program.cs` - Complete DI setup
- `patterns/dotnet/AiPatterns/Infrastructure/Messaging/EventPublisher.cs` - Kafka usage
- `patterns/dotnet/AiPatterns/Infrastructure/Repositories/OrderRepository.cs` - SQL usage
- `core/dotnet/Core.Infrastructure/KafkaClient.cs` - Interface definitions
