# .NET Patterns for AI Services

This document defines the standard patterns for .NET services including SLI, SOD, SRE, contextual logging, error repository, and Prometheus metrics.

---

## Service Level Indicators (SLI)

Track key performance metrics for your service:

```csharp
using Core.Sli;
using Prometheus;

public class OrderServiceSli
{
    // Availability SLI - percentage of successful requests
    private static readonly Counter RequestsTotal = Metrics
        .CreateCounter("order_service_requests_total", "Total requests",
            new CounterConfiguration { LabelNames = new[] { "method", "endpoint", "status" } });

    // Latency SLI - request duration histogram
    private static readonly Histogram RequestDuration = Metrics
        .CreateHistogram("order_service_request_duration_seconds", "Request duration",
            new HistogramConfiguration
            {
                LabelNames = new[] { "method", "endpoint" },
                Buckets = new[] { .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10 }
            });

    // Throughput SLI - orders processed per second
    private static readonly Counter OrdersProcessed = Metrics
        .CreateCounter("order_service_orders_processed_total", "Orders processed",
            new CounterConfiguration { LabelNames = new[] { "status" } });

    public void RecordRequest(string method, string endpoint, int statusCode, double durationSeconds)
    {
        var status = statusCode >= 200 && statusCode < 400 ? "success" : "error";
        RequestsTotal.WithLabels(method, endpoint, status).Inc();
        RequestDuration.WithLabels(method, endpoint).Observe(durationSeconds);
    }

    public void RecordOrderProcessed(string status)
    {
        OrdersProcessed.WithLabels(status).Inc();
    }
}
```

### SLI Middleware

```csharp
public class SliMiddleware
{
    private readonly RequestDelegate _next;
    private readonly OrderServiceSli _sli;

    public SliMiddleware(RequestDelegate next, OrderServiceSli sli)
    {
        _next = next;
        _sli = sli;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        var stopwatch = Stopwatch.StartNew();
        
        try
        {
            await _next(context);
        }
        finally
        {
            stopwatch.Stop();
            _sli.RecordRequest(
                context.Request.Method,
                context.Request.Path,
                context.Response.StatusCode,
                stopwatch.Elapsed.TotalSeconds);
        }
    }
}
```

---

## Service Oriented Design (SOD)

Structure services with clear separation of concerns:

```
src/
├── {ServiceName}.Api/           # HTTP/gRPC layer
│   ├── Controllers/             # Request handling
│   ├── Middleware/              # Cross-cutting concerns
│   └── Program.cs               # Composition root
├── {ServiceName}.Domain/        # Business logic (no dependencies)
│   ├── Models/                  # Domain entities
│   ├── Services/                # Business rules
│   ├── Errors/                  # Domain-specific errors
│   └── Interfaces/              # Repository contracts
└── {ServiceName}.Infrastructure/ # External integrations
    ├── Repositories/            # Data access
    ├── Kafka/                   # Event publishing
    └── External/                # Third-party APIs
```

### Domain Service Pattern

```csharp
// Domain/Services/OrderService.cs
public class OrderService : IOrderService
{
    private readonly IOrderRepository _repository;
    private readonly IEventPublisher _eventPublisher;
    private readonly ServiceLogger _logger;

    public OrderService(
        IOrderRepository repository,
        IEventPublisher eventPublisher,
        ServiceLogger logger)
    {
        _repository = repository;
        _eventPublisher = eventPublisher;
        _logger = logger;
    }

    public async Task<Order> CreateOrderAsync(CreateOrderRequest request)
    {
        // Business validation
        if (request.Items.Count == 0)
            throw OrderErrors.EmptyOrder();

        // Domain logic
        var order = Order.Create(request.CustomerId, request.Items);
        
        // Persist
        await _repository.SaveAsync(order);
        
        // Publish event
        await _eventPublisher.PublishAsync(new OrderCreatedEvent(order));
        
        _logger.Info("Order created", new { orderId = order.Id, customerId = order.CustomerId });
        
        return order;
    }
}
```

---

## Site Reliability Engineering (SRE)

### Health Checks

```csharp
// Program.cs
builder.Services.AddHealthChecks()
    .AddSqlServer(connectionString, name: "sql", tags: new[] { "db", "ready" })
    .AddRedis(redisConnection, name: "redis", tags: new[] { "cache", "ready" })
    .AddKafka(kafkaConfig, name: "kafka", tags: new[] { "messaging", "ready" });

app.MapHealthChecks("/health/live", new HealthCheckOptions
{
    Predicate = _ => false // Just checks if app responds
});

app.MapHealthChecks("/health/ready", new HealthCheckOptions
{
    Predicate = check => check.Tags.Contains("ready")
});
```

### Circuit Breaker

```csharp
using Polly;
using Polly.CircuitBreaker;

public class ResilientHttpClient
{
    private readonly AsyncCircuitBreakerPolicy _circuitBreaker;
    private readonly HttpClient _httpClient;

    public ResilientHttpClient(HttpClient httpClient)
    {
        _httpClient = httpClient;
        _circuitBreaker = Policy
            .Handle<HttpRequestException>()
            .CircuitBreakerAsync(
                exceptionsAllowedBeforeBreaking: 5,
                durationOfBreak: TimeSpan.FromSeconds(30),
                onBreak: (ex, duration) => 
                    Console.WriteLine($"Circuit opened for {duration.TotalSeconds}s"),
                onReset: () => 
                    Console.WriteLine("Circuit closed"));
    }

    public async Task<T> GetAsync<T>(string url)
    {
        return await _circuitBreaker.ExecuteAsync(async () =>
        {
            var response = await _httpClient.GetAsync(url);
            response.EnsureSuccessStatusCode();
            return await response.Content.ReadFromJsonAsync<T>();
        });
    }
}
```

### Retry Policy

```csharp
var retryPolicy = Policy
    .Handle<SqlException>()
    .Or<TimeoutException>()
    .WaitAndRetryAsync(
        retryCount: 3,
        sleepDurationProvider: attempt => TimeSpan.FromSeconds(Math.Pow(2, attempt)),
        onRetry: (exception, timeSpan, retryCount, context) =>
        {
            _logger.Warning($"Retry {retryCount} after {timeSpan.TotalSeconds}s", 
                new { exception = exception.Message });
        });
```

---

## Contextual Logging

Always include context in log entries:

```csharp
using Core.Logger;

public class ServiceLogger
{
    private readonly ILogger _logger;
    private readonly string _serviceName;

    public ServiceLogger(ILogger<ServiceLogger> logger, IConfiguration config)
    {
        _logger = logger;
        _serviceName = config["ServiceName"] ?? "unknown";
    }

    public void Info(string message, object? context = null)
    {
        using (_logger.BeginScope(new Dictionary<string, object>
        {
            ["service"] = _serviceName,
            ["timestamp"] = DateTime.UtcNow.ToString("O")
        }))
        {
            if (context != null)
                _logger.LogInformation("{Message} {@Context}", message, context);
            else
                _logger.LogInformation("{Message}", message);
        }
    }

    public void Error(string message, Exception ex, object? context = null)
    {
        using (_logger.BeginScope(new Dictionary<string, object>
        {
            ["service"] = _serviceName,
            ["timestamp"] = DateTime.UtcNow.ToString("O"),
            ["exceptionType"] = ex.GetType().Name
        }))
        {
            _logger.LogError(ex, "{Message} {@Context}", message, context);
        }
    }
}
```

### Correlation ID Middleware

```csharp
public class CorrelationIdMiddleware
{
    private readonly RequestDelegate _next;
    private const string CorrelationIdHeader = "X-Correlation-ID";

    public CorrelationIdMiddleware(RequestDelegate next)
    {
        _next = next;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        var correlationId = context.Request.Headers[CorrelationIdHeader].FirstOrDefault()
            ?? Guid.NewGuid().ToString();

        context.Items["CorrelationId"] = correlationId;
        context.Response.Headers[CorrelationIdHeader] = correlationId;

        using (LogContext.PushProperty("CorrelationId", correlationId))
        {
            await _next(context);
        }
    }
}
```

---

## Error Repository

Centralized error definitions with codes:

```csharp
// Domain/Errors/OrderErrors.cs
using Core.Errors;

public static class OrderErrors
{
    private static readonly ErrorRegistry Registry = new();

    static OrderErrors()
    {
        // Register all error codes at startup
        Registry.Register("ORD-001", "Order not found", HttpStatusCode.NotFound);
        Registry.Register("ORD-002", "Empty order not allowed", HttpStatusCode.BadRequest);
        Registry.Register("ORD-003", "Invalid order status transition", HttpStatusCode.Conflict);
        Registry.Register("ORD-004", "Insufficient inventory", HttpStatusCode.Conflict);
        Registry.Register("ORD-005", "Payment failed", HttpStatusCode.PaymentRequired);
    }

    public static ServiceException NotFound(Guid orderId) =>
        Registry.CreateException("ORD-001", new { orderId });

    public static ServiceException EmptyOrder() =>
        Registry.CreateException("ORD-002");

    public static ServiceException InvalidStatusTransition(string from, string to) =>
        Registry.CreateException("ORD-003", new { from, to });

    public static ServiceException InsufficientInventory(string sku, int requested, int available) =>
        Registry.CreateException("ORD-004", new { sku, requested, available });

    public static ServiceException PaymentFailed(string reason) =>
        Registry.CreateException("ORD-005", new { reason });
}
```

### Error Handler Middleware

```csharp
public class ErrorHandlerMiddleware
{
    private readonly RequestDelegate _next;
    private readonly ServiceLogger _logger;

    public ErrorHandlerMiddleware(RequestDelegate next, ServiceLogger logger)
    {
        _next = next;
        _logger = logger;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        try
        {
            await _next(context);
        }
        catch (ServiceException ex)
        {
            _logger.Warning("Service error", new { code = ex.Code, message = ex.Message });
            await WriteErrorResponse(context, ex.StatusCode, ex.Code, ex.Message, ex.Details);
        }
        catch (Exception ex)
        {
            _logger.Error("Unhandled exception", ex);
            await WriteErrorResponse(context, 500, "SYS-001", "Internal server error", null);
        }
    }

    private static async Task WriteErrorResponse(
        HttpContext context, int statusCode, string code, string message, object? details)
    {
        context.Response.StatusCode = statusCode;
        context.Response.ContentType = "application/json";
        
        var error = new
        {
            error = new { code, message, details, timestamp = DateTime.UtcNow }
        };
        
        await context.Response.WriteAsJsonAsync(error);
    }
}
```

---

## Prometheus Metrics Endpoint

### Setup in Program.cs

```csharp
using Prometheus;

var builder = WebApplication.CreateBuilder(args);

// Add metrics
builder.Services.AddSingleton<OrderServiceSli>();

var app = builder.Build();

// Enable metrics endpoint
app.UseHttpMetrics(); // Automatic HTTP metrics
app.MapMetrics("/metrics"); // Expose /metrics endpoint

app.Run();
```

### Custom Business Metrics

```csharp
public class OrderMetrics
{
    // Counter - things that only go up
    private static readonly Counter OrdersCreated = Metrics
        .CreateCounter("orders_created_total", "Total orders created",
            new CounterConfiguration { LabelNames = new[] { "source", "customer_type" } });

    // Gauge - things that go up and down
    private static readonly Gauge PendingOrders = Metrics
        .CreateGauge("orders_pending", "Current pending orders count");

    // Histogram - distributions (latency, sizes)
    private static readonly Histogram OrderValue = Metrics
        .CreateHistogram("order_value_dollars", "Order value distribution",
            new HistogramConfiguration
            {
                Buckets = new[] { 10, 25, 50, 100, 250, 500, 1000, 2500, 5000 }
            });

    // Summary - percentiles
    private static readonly Summary ProcessingTime = Metrics
        .CreateSummary("order_processing_seconds", "Order processing time",
            new SummaryConfiguration
            {
                Objectives = new[]
                {
                    new QuantileEpsilonPair(0.5, 0.05),
                    new QuantileEpsilonPair(0.9, 0.01),
                    new QuantileEpsilonPair(0.99, 0.001)
                }
            });

    public void RecordOrderCreated(string source, string customerType, decimal value)
    {
        OrdersCreated.WithLabels(source, customerType).Inc();
        OrderValue.Observe((double)value);
        PendingOrders.Inc();
    }

    public void RecordOrderCompleted(double processingSeconds)
    {
        PendingOrders.Dec();
        ProcessingTime.Observe(processingSeconds);
    }
}
```

---

## Complete Program.cs Example

```csharp
using Core.Config;
using Core.Logger;
using Core.Errors;
using Core.Metrics;
using Prometheus;
using Serilog;

var builder = WebApplication.CreateBuilder(args);

// Configure Serilog
Log.Logger = new LoggerConfiguration()
    .ReadFrom.Configuration(builder.Configuration)
    .Enrich.FromLogContext()
    .Enrich.WithProperty("Service", "OrderService")
    .WriteTo.Console(new CompactJsonFormatter())
    .CreateLogger();

builder.Host.UseSerilog();

// Add services
builder.Services.AddControllers();
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

// Add health checks
builder.Services.AddHealthChecks()
    .AddSqlServer(builder.Configuration.GetConnectionString("SqlServer")!)
    .AddRedis(builder.Configuration.GetConnectionString("Redis")!);

// Add custom services
builder.Services.AddSingleton<ServiceLogger>();
builder.Services.AddSingleton<OrderServiceSli>();
builder.Services.AddSingleton<OrderMetrics>();
builder.Services.AddScoped<IOrderService, OrderService>();
builder.Services.AddScoped<IOrderRepository, OrderRepository>();

var app = builder.Build();

// Middleware pipeline
app.UseMiddleware<CorrelationIdMiddleware>();
app.UseMiddleware<ErrorHandlerMiddleware>();
app.UseMiddleware<SliMiddleware>();

app.UseSerilogRequestLogging();

if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

// Prometheus metrics
app.UseHttpMetrics();
app.MapMetrics("/metrics");

// Health checks
app.MapHealthChecks("/health/live");
app.MapHealthChecks("/health/ready");

app.MapControllers();

app.Run();
```
