using Core.Config;
using Core.Logger;
using Core.Errors;
using Core.Sli;
using Core.Infrastructure;
using Core.Infrastructure.SqlServer;
using Core.Infrastructure.MongoDB;
using Core.Infrastructure.Kafka;
using Prometheus;
using Serilog;
using Serilog.Events;
using Serilog.Formatting.Compact;
using AiPatterns.Domain.Interfaces;
using AiPatterns.Infrastructure.Repositories;
using AiPatterns.Infrastructure.Cache;
using AiPatterns.Infrastructure.Messaging;
using AiPatterns.Domain.Services;
using AiPatterns.Domain.Sli;

var builder = WebApplication.CreateBuilder(args);

// ========================================
// 1. LOGGING SETUP (Core.Logger)
// ========================================
var loggerConfig = new LoggerConfig
{
    ServiceName = builder.Configuration["ServiceName"] ?? "ai-patterns",
    Environment = builder.Environment.EnvironmentName,
    Version = "1.0.0",
    LogLevel = builder.Configuration["Logging:LogLevel:Default"] ?? "Information",
    OutputFormat = "json"
};

var serviceLogger = new ServiceLogger(loggerConfig);
builder.Services.AddSingleton(serviceLogger);

// Also configure Serilog for ASP.NET Core integration
Log.Logger = new LoggerConfiguration()
    .MinimumLevel.Override("Microsoft", LogEventLevel.Warning)
    .MinimumLevel.Override("Microsoft.EntityFrameworkCore", LogEventLevel.Warning)
    .Enrich.FromLogContext()
    .Enrich.WithProperty("Service", loggerConfig.ServiceName)
    .WriteTo.Console(new CompactJsonFormatter())
    .CreateLogger();
builder.Host.UseSerilog();

// ========================================
// 2. SLI TRACKER (Core.Sli)
// ========================================
builder.Services.AddSingleton<PatternsSli>();
builder.Services.AddSingleton<ISliTracker>(sp => sp.GetRequiredService<PatternsSli>());

// ========================================
// 3. HTTP CONTEXT & CORRELATION ID (REQUIRED)
// ========================================
builder.Services.AddHttpContextAccessor();

// ========================================
// 4. CORE.INFRASTRUCTURE CLIENTS (CRITICAL!)
// ========================================

// SQL Server via Core.Infrastructure.SqlServer
var sqlConnectionString = builder.Configuration.GetConnectionString("SqlServer") ?? 
    "Server=localhost,1433;Database=AiPatterns;User Id=sa;Password=AiPatterns2024!;TrustServerCertificate=True;";
var sqlConfig = new SqlServerConfig { ConnectionString = sqlConnectionString };
builder.Services.AddSingleton<ISqlServerClient>(sp => 
    new SqlServerClient(sqlConfig, serviceLogger));

// MongoDB via Core.Infrastructure.MongoDB  
var mongoConnectionString = builder.Configuration.GetConnectionString("MongoDB") ?? 
    "mongodb://localhost:27017";
var mongoDatabase = builder.Configuration["MongoDB:Database"] ?? "AiPatternsDB";
var mongoConfig = new MongoConfig { ConnectionString = mongoConnectionString, DatabaseName = mongoDatabase };
builder.Services.AddSingleton<Core.Infrastructure.MongoDB.IMongoClient>(sp => 
    new Core.Infrastructure.MongoDB.MongoClient(mongoConfig, serviceLogger));

// Redis via Core.Infrastructure (async factory pattern)
var redisHost = builder.Configuration["Redis:Host"] ?? "localhost";
var redisPort = int.TryParse(builder.Configuration["Redis:Port"], out var p) ? p : 6379;
var redisConnectionString = $"{redisHost}:{redisPort}";
var redisConfig = new RedisConfig { Host = redisHost, Port = redisPort };
builder.Services.AddSingleton<IRedisClient>(sp =>
{
    try
    {
        var client = RedisClient.CreateAsync(redisConfig).GetAwaiter().GetResult();
        return client;
    }
    catch (Exception ex)
    {
        serviceLogger.Warning("Redis connection failed, using null client", ex, new { host = redisHost, port = redisPort });
        // Return a null-safe wrapper or rethrow based on requirements
        throw new InvalidOperationException($"Failed to connect to Redis at {redisHost}:{redisPort}. Ensure Redis is running.", ex);
    }
});

// Kafka via Core.Infrastructure.Kafka
var kafkaBootstrapServers = builder.Configuration["Kafka:BootstrapServers"] ?? 
    "localhost:9092";
var kafkaConfig = new KafkaConfig { BootstrapServers = kafkaBootstrapServers };
builder.Services.AddSingleton<IKafkaProducer>(sp => 
    new KafkaProducer(kafkaConfig, serviceLogger));
// Note: IKafkaConsumer not yet implemented in Core.Infrastructure
// Consumer would use Confluent.Kafka directly in production

// ScyllaDB via Core.Infrastructure
var scyllaHosts = builder.Configuration["ScyllaDB:ContactPoints"]?.Split(',') ?? new[] { "localhost" };
var scyllaKeyspace = builder.Configuration["ScyllaDB:Keyspace"] ?? "ai_patterns";
var scyllaConfig = new ScyllaConfig { Hosts = scyllaHosts, Keyspace = scyllaKeyspace };
builder.Services.AddSingleton<ScyllaDBClient>(sp => 
    new ScyllaDBClient(scyllaConfig, serviceLogger));

// ========================================
// 5. REPOSITORY LAYER (All using Core.Infrastructure)
// ========================================
builder.Services.AddScoped<IOrderRepository, OrderRepository>();
builder.Services.AddScoped<IUserProfileRepository, UserProfileRepository>();
builder.Services.AddScoped<ITelemetryRepository, TelemetryRepository>();
builder.Services.AddScoped<IRealtimeCache, RealtimeCache>();
builder.Services.AddScoped<IEventPublisher, EventPublisher>();
builder.Services.AddScoped<IEventConsumer, EventConsumer>();

// ========================================
// 6. DOMAIN SERVICES (Business Logic)
// ========================================
builder.Services.AddScoped<IPatternsService, PatternsService>();

// ========================================
// 7. CONTROLLERS & SWAGGER
// ========================================
builder.Services.AddControllers()
    .AddJsonOptions(options =>
    {
        options.JsonSerializerOptions.PropertyNamingPolicy = 
            System.Text.Json.JsonNamingPolicy.CamelCase;
        options.JsonSerializerOptions.Converters.Add(
            new System.Text.Json.Serialization.JsonStringEnumConverter());
    });
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen(c =>
{
    c.SwaggerDoc("v1", new Microsoft.OpenApi.Models.OpenApiInfo 
    { 
        Title = "AI Patterns API - All Core.Infrastructure Clients", 
        Version = "v1",
        Description = "Comprehensive patterns demonstrating ALL Core.Infrastructure data platforms:\n\n" +
                     "• **SQL Server** (Core.Infrastructure.SqlServer) - Transactional data, ACID compliance\n" +
                     "• **MongoDB** (Core.Infrastructure.MongoDB) - Document storage, flexible schemas\n" +
                     "• **ScyllaDB** (Core.Infrastructure.ScyllaDB) - Time-series data, high throughput\n" +
                     "• **Redis** (Core.Infrastructure.Redis) - Caching, real-time data, leaderboards\n" +
                     "• **Kafka** (Core.Infrastructure.Kafka) - Event streaming, pub/sub messaging\n\n" +
                     "Each endpoint demonstrates optimal patterns for its respective data platform using the AI Core.Infrastructure package."
    });
});

// ========================================
// 8. HEALTH CHECKS (All Data Platforms)
// ========================================
// Note: Health checks are configured but won't block startup if services aren't available
builder.Services.AddHealthChecks()
    .AddCheck("self", () => Microsoft.Extensions.Diagnostics.HealthChecks.HealthCheckResult.Healthy("AI Patterns service is running"));
// Database health checks commented out for demo - uncomment when infrastructure is available:
// .AddSqlServer(sqlConnectionString, name: "sqlserver", tags: new[] { "db", "ready" })
// .AddRedis(redisConnectionString, name: "redis", tags: new[] { "cache", "ready" })
// .AddMongoDb(mongoConnectionString, name: "mongodb", tags: new[] { "db", "ready" })
// .AddKafka(options => { options.BootstrapServers = kafkaBootstrapServers; }, name: "kafka", tags: new[] { "messaging", "ready" });
// Note: ScyllaDB health check would need custom implementation

var app = builder.Build();

// ========================================
// 9. MIDDLEWARE PIPELINE (ORDER MATTERS!)
// ========================================

// Development tools (always available for patterns demo)
app.UseSwagger();
app.UseSwaggerUI(c =>
{
    c.SwaggerEndpoint("/swagger/v1/swagger.json", "AI Patterns API v1");
    c.RoutePrefix = string.Empty; // Serve UI at root
    c.DocumentTitle = "AI Patterns - All Core.Infrastructure Clients";
    c.DefaultModelsExpandDepth(-1); // Hide models section by default
    c.DisplayRequestDuration();
});

// Serilog request logging
app.UseSerilogRequestLogging(options =>
{
    options.MessageTemplate = "HTTP {RequestMethod} {RequestPath} responded {StatusCode} in {Elapsed:0.0000} ms";
    options.GetLevel = (httpContext, elapsed, ex) => ex != null 
        ? LogEventLevel.Error 
        : httpContext.Response.StatusCode > 499 
            ? LogEventLevel.Error 
            : LogEventLevel.Information;
});

// ========================================
// 10. PROMETHEUS METRICS
// ========================================
app.UseHttpMetrics();
app.MapMetrics("/metrics");

// ========================================
// 11. HEALTH ENDPOINTS (SRE)
// ========================================
app.MapHealthChecks("/health/live", new Microsoft.AspNetCore.Diagnostics.HealthChecks.HealthCheckOptions
{
    Predicate = _ => false  // Just checks if app responds
});

app.MapHealthChecks("/health/ready", new Microsoft.AspNetCore.Diagnostics.HealthChecks.HealthCheckOptions
{
    Predicate = check => check.Tags.Contains("ready")
});

// ========================================
// 12. CONTROLLERS
// ========================================
app.MapControllers();

// ========================================
// 13. STARTUP LOGGING
// ========================================
serviceLogger.Information("🚀 AI Patterns service starting with ALL Core.Infrastructure clients!", new { 
    Environment = app.Environment.EnvironmentName,
    Version = "1.0.0",
    Port = builder.Configuration["ASPNETCORE_URLS"] ?? "http://localhost:5000",
    DataPlatforms = new {
        SqlServer = "✅ Core.Infrastructure.SqlServer - Transactional data",
        MongoDB = "✅ Core.Infrastructure.MongoDB - Document storage", 
        ScyllaDB = "✅ Core.Infrastructure.ScyllaDB - Time-series data",
        Redis = "✅ Core.Infrastructure.Redis - Caching & real-time",
        Kafka = "✅ Core.Infrastructure.Kafka - Event streaming"
    },
    SwaggerUI = "http://localhost:5000 (patterns documentation & testing)",
    HealthChecks = new {
        Live = "/health/live",
        Ready = "/health/ready"
    },
    Metrics = "/metrics (Prometheus)"
});

serviceLogger.Information("📋 Available API endpoints:", new {
    Endpoints = new {
        Orders = "POST/PATCH /api/v1/patterns/orders - SQL Server transactional patterns",
        Users = "POST/PUT /api/v1/patterns/users - MongoDB document patterns", 
        Telemetry = "POST/GET /api/v1/patterns/telemetry - ScyllaDB time-series patterns",
        Leaderboards = "POST/GET /api/v1/patterns/leaderboards - Redis real-time patterns",
        Sessions = "POST /api/v1/patterns/sessions - Redis + Kafka session patterns",
        Analytics = "GET /api/v1/patterns/analytics - Cross-platform analytics",
        Health = "GET /api/v1/patterns/health - All platform connectivity"
    }
});

app.Run();