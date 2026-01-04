---
name: ai-logging-patterns
description: >
  Logging patterns using Core.Logger for .NET and core/go/logger for Go.
  Use when adding logging to services, debugging issues, or reviewing log output.
  Ensures structured JSON logging with correlation IDs, proper log levels, and contextual information.
  NEVER use raw Serilog, ILogger<T>, logrus, or zerolog directly.
---

# AI Logging Patterns

## Core Principles

1. **Use Core packages only** - Never raw Serilog/ILogger<T>/logrus/zerolog
2. **Correlation IDs everywhere** - Propagate X-Correlation-ID through all logs
3. **Structured logging** - Use named parameters, not string interpolation
4. **Contextual logging** - Add relevant context (user_id, order_id, etc.)
5. **Appropriate levels** - Debug for development, Info for business events, Warn for anomalies, Error for failures

---

## .NET Logging with Core.Logger

### Basic Setup

```csharp
using Core.Logger;

// DI Registration
builder.Services.AddSingleton<ServiceLogger>(sp => 
    new ServiceLogger("order-service", builder.Configuration));
```

### Service Usage

```csharp
using Core.Logger;

public class OrderService
{
    private readonly ServiceLogger _logger;

    public OrderService(ServiceLogger logger)
    {
        _logger = logger;
    }

    public async Task<Order> CreateOrder(CreateOrderRequest request, string correlationId)
    {
        // Create contextual logger
        var log = _logger.WithContext(
            correlationId: correlationId,
            component: nameof(OrderService));

        log.Information("Creating order for customer {CustomerId}", request.CustomerId);
        
        try
        {
            var order = new Order
            {
                Id = Guid.NewGuid(),
                CustomerId = request.CustomerId,
                Total = request.Total
            };

            await _repository.CreateAsync(order);
            
            log.Information("Order {OrderId} created successfully with total {Total:C}",
                order.Id, order.Total);
            
            return order;
        }
        catch (Exception ex)
        {
            log.Error(ex, "Failed to create order for customer {CustomerId}", 
                request.CustomerId);
            throw;
        }
    }
}
```

### Log Levels

```csharp
// DEBUG - Development/troubleshooting details
_logger.Debug("Validating request payload: {Payload}", JsonSerializer.Serialize(request));

// INFORMATION - Business events, state changes
_logger.Information("Order {OrderId} status changed from {OldStatus} to {NewStatus}",
    orderId, oldStatus, newStatus);

// WARNING - Anomalies, degraded performance, retry attempts
_logger.Warning("Cache miss for order {OrderId}, falling back to database", orderId);
_logger.Warning("External API response time {Duration}ms exceeds threshold", duration);

// ERROR - Failures requiring attention
_logger.Error(ex, "Database query failed for customer {CustomerId}", customerId);

// FATAL - Application cannot continue
_logger.Fatal("Unable to connect to database after {RetryCount} attempts", retryCount);
```

### Contextual Logging Patterns

```csharp
// For entry/exit logging
public async Task<Order> GetOrder(Guid orderId)
{
    _logger.Debug("→ GetOrder started for {OrderId}", orderId);
    
    var order = await _repository.GetByIdAsync(orderId);
    
    _logger.Debug("← GetOrder completed for {OrderId}", orderId);
    return order;
}

// For external calls
public async Task<PaymentResult> ProcessPayment(Payment payment)
{
    _logger.Information("Calling payment gateway for amount {Amount:C}", payment.Amount);
    var stopwatch = Stopwatch.StartNew();
    
    try
    {
        var result = await _paymentGateway.ProcessAsync(payment);
        stopwatch.Stop();
        
        _logger.Information("Payment gateway responded in {Duration}ms with status {Status}",
            stopwatch.ElapsedMilliseconds, result.Status);
        
        return result;
    }
    catch (Exception ex)
    {
        stopwatch.Stop();
        _logger.Error(ex, "Payment gateway failed after {Duration}ms", 
            stopwatch.ElapsedMilliseconds);
        throw;
    }
}
```

---

## Go Logging with core/go/logger

### Basic Setup

```go
import (
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
    "go.uber.org/zap"
)

// Development mode
log, err := logger.NewDevelopment("order-service", "1.0.0")
if err != nil {
    panic(err)
}
defer log.Sync()

// Production mode
log, err := logger.NewProduction("order-service", "1.0.0")
```

### Service Usage

```go
import (
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
    "go.uber.org/zap"
)

type OrderService struct {
    logger *logger.Logger
    repo   *OrderRepository
}

func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    // Get contextual logger with correlation ID
    log := s.logger.WithCorrelation(ctx)
    
    log.Info("Creating order",
        zap.String("customer_id", req.CustomerID))
    
    order := &Order{
        ID:         uuid.New().String(),
        CustomerID: req.CustomerID,
        Total:      req.Total,
    }
    
    if err := s.repo.Create(ctx, order); err != nil {
        log.Error("Failed to create order",
            zap.String("customer_id", req.CustomerID),
            zap.Error(err))
        return nil, err
    }
    
    log.Info("Order created successfully",
        zap.String("order_id", order.ID),
        zap.Float64("total", order.Total))
    
    return order, nil
}
```

### Log Levels

```go
// DEBUG - Development/troubleshooting
log.Debug("Validating request payload",
    zap.Any("payload", request))

// INFO - Business events
log.Info("Order status changed",
    zap.String("order_id", orderID),
    zap.String("old_status", oldStatus),
    zap.String("new_status", newStatus))

// WARN - Anomalies
log.Warn("Cache miss, falling back to database",
    zap.String("order_id", orderID))

// ERROR - Failures
log.Error("Database query failed",
    zap.String("customer_id", customerID),
    zap.Error(err))

// FATAL - Application cannot continue (calls os.Exit)
log.Fatal("Unable to connect to database",
    zap.Int("retry_count", retryCount))
```

### Contextual Logging Patterns

```go
// Entry/exit logging
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*Order, error) {
    log := s.logger.WithCorrelation(ctx)
    
    log.Debug("→ GetOrder started", zap.String("order_id", orderID))
    
    order, err := s.repo.GetByID(ctx, orderID)
    if err != nil {
        log.Error("GetOrder failed", zap.String("order_id", orderID), zap.Error(err))
        return nil, err
    }
    
    log.Debug("← GetOrder completed", zap.String("order_id", orderID))
    return order, nil
}

// External call logging with timing
func (s *OrderService) ProcessPayment(ctx context.Context, payment *Payment) (*PaymentResult, error) {
    log := s.logger.WithCorrelation(ctx)
    
    log.Info("Calling payment gateway",
        zap.Float64("amount", payment.Amount))
    
    start := time.Now()
    result, err := s.paymentGateway.Process(ctx, payment)
    duration := time.Since(start)
    
    if err != nil {
        log.Error("Payment gateway failed",
            zap.Duration("duration", duration),
            zap.Error(err))
        return nil, err
    }
    
    log.Info("Payment gateway responded",
        zap.Duration("duration", duration),
        zap.String("status", result.Status))
    
    return result, nil
}
```

---

## Correlation ID Propagation

### HTTP Request (Middleware)

```csharp
// .NET
public class CorrelationIdMiddleware
{
    public async Task InvokeAsync(HttpContext context)
    {
        var correlationId = context.Request.Headers["X-Correlation-ID"].FirstOrDefault()
            ?? Guid.NewGuid().ToString();
        
        context.Items["CorrelationId"] = correlationId;
        context.Response.Headers["X-Correlation-ID"] = correlationId;
        
        await _next(context);
    }
}
```

```go
// Go
func CorrelationID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        correlationID := r.Header.Get("X-Correlation-ID")
        if correlationID == "" {
            correlationID = uuid.New().String()
        }
        
        ctx := context.WithValue(r.Context(), logger.CorrelationIDKey, correlationID)
        w.Header().Set("X-Correlation-ID", correlationID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Kafka Events

```csharp
// .NET - Include in headers
var message = new KafkaMessage<string>
{
    Key = order.Id.ToString(),
    Value = JsonSerializer.Serialize(eventPayload),
    Headers = new Dictionary<string, string>
    {
        { "correlation_id", correlationId }  // ALWAYS include
    }
};
```

```go
// Go - Include in headers
headers := map[string]string{
    "correlation_id": ctx.Value(logger.CorrelationIDKey).(string),
}
producer.SendMessage(ctx, topic, key, payload, headers)
```

---

## Log Output Format (JSON)

```json
{
  "timestamp": "2024-01-15T10:30:45.123Z",
  "level": "info",
  "service": "order-service",
  "version": "1.0.0",
  "correlation_id": "abc-123-def",
  "component": "OrderService",
  "message": "Order created successfully",
  "order_id": "ord-456",
  "customer_id": "cust-789",
  "total": 99.99
}
```
