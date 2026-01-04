---
name: ai-error-handling
description: >
  Error handling patterns using Core.Errors for .NET and core/go/errors for Go.
  Use when creating error codes, handling exceptions, or designing error responses.
  Ensures all errors have codes, severity levels, and HTTP status mappings.
  NEVER throw generic Exception or use fmt.Errorf without error codes.
---

# AI Error Handling Patterns

## Core Principles

1. **Every error has a code** - Format: `{SERVICE}-{NUMBER}` (e.g., ORD-001)
2. **Errors have severity** - Low, Medium, High, Critical
3. **Errors map to HTTP status** - 400, 404, 409, 500, etc.
4. **Errors include context** - What entity/ID failed
5. **Never throw generic exceptions** - Always use ServiceError

---

## .NET Error Handling with Core.Errors

### Error Registry Pattern

```csharp
using Core.Errors;
using System.Net;

namespace OrderService.Domain.Errors;

public static class OrderErrors
{
    private static readonly ErrorRegistry _registry = new();

    static OrderErrors()
    {
        // Register all error codes at startup
        _registry.Register("ORD-001", "Order not found", HttpStatusCode.NotFound);
        _registry.Register("ORD-002", "Invalid order status transition", HttpStatusCode.Conflict);
        _registry.Register("ORD-003", "Insufficient inventory", HttpStatusCode.UnprocessableEntity);
        _registry.Register("ORD-004", "Payment declined", HttpStatusCode.PaymentRequired);
        _registry.Register("ORD-005", "Order already cancelled", HttpStatusCode.Conflict);
        _registry.Register("ORD-006", "Customer not found", HttpStatusCode.NotFound);
        _registry.Register("ORD-007", "Invalid order total", HttpStatusCode.BadRequest);
        _registry.Register("ORD-500", "Internal order processing error", HttpStatusCode.InternalServerError);
    }

    // Factory methods with context
    public static ServiceError NotFound(Guid orderId) =>
        _registry.CreateError("ORD-001", new { orderId });

    public static ServiceError InvalidStatusTransition(Guid orderId, string fromStatus, string toStatus) =>
        _registry.CreateError("ORD-002", new { orderId, fromStatus, toStatus });

    public static ServiceError InsufficientInventory(Guid productId, int requested, int available) =>
        _registry.CreateError("ORD-003", new { productId, requested, available });

    public static ServiceError PaymentDeclined(Guid orderId, string reason) =>
        _registry.CreateError("ORD-004", new { orderId, reason });

    public static ServiceError AlreadyCancelled(Guid orderId) =>
        _registry.CreateError("ORD-005", new { orderId });

    public static ServiceError CustomerNotFound(Guid customerId) =>
        _registry.CreateError("ORD-006", new { customerId });

    public static ServiceError InvalidTotal(decimal total) =>
        _registry.CreateError("ORD-007", new { total });

    public static ServiceError InternalError(string operation) =>
        _registry.CreateError("ORD-500", new { operation });
}
```

### Usage in Services

```csharp
using Core.Errors;
using OrderService.Domain.Errors;

public class OrderService
{
    public async Task<Order> GetOrderAsync(Guid orderId)
    {
        var order = await _repository.GetByIdAsync(orderId);
        
        if (order == null)
        {
            throw OrderErrors.NotFound(orderId);  // ✅ Throws ServiceError with code
        }
        
        return order;
    }

    public async Task UpdateStatusAsync(Guid orderId, string newStatus)
    {
        var order = await GetOrderAsync(orderId);
        
        if (!IsValidTransition(order.Status, newStatus))
        {
            throw OrderErrors.InvalidStatusTransition(orderId, order.Status, newStatus);
        }
        
        order.Status = newStatus;
        await _repository.UpdateAsync(order);
    }

    public async Task CancelOrderAsync(Guid orderId)
    {
        var order = await GetOrderAsync(orderId);
        
        if (order.Status == "Cancelled")
        {
            throw OrderErrors.AlreadyCancelled(orderId);
        }
        
        // Process cancellation
    }
}
```

### Error Handling Middleware

```csharp
using Core.Errors;
using Core.Logger;

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
        catch (ServiceError error)
        {
            var correlationId = context.Items["CorrelationId"]?.ToString();
            
            _logger.Warning("Service error: {ErrorCode} - {Message}",
                error.Code, error.Message);

            context.Response.StatusCode = (int)error.StatusCode;
            context.Response.ContentType = "application/json";

            await context.Response.WriteAsJsonAsync(new
            {
                error = error.Code,
                message = error.Message,
                details = error.Context,
                correlationId
            });
        }
        catch (Exception ex)
        {
            var correlationId = context.Items["CorrelationId"]?.ToString();
            
            _logger.Error(ex, "Unhandled exception");

            context.Response.StatusCode = 500;
            context.Response.ContentType = "application/json";

            await context.Response.WriteAsJsonAsync(new
            {
                error = "SYS-500",
                message = "An unexpected error occurred",
                correlationId
            });
        }
    }
}
```

---

## Go Error Handling with core/go/errors

### Error Definitions

```go
package errors

import (
    "github.com/your-github-org/ai-scaffolder/core/go/errors"
)

// Define service-specific errors
var (
    ErrOrderNotFound = errors.New("ORD-001", errors.SeverityMedium, "Order not found")
    ErrInvalidStatus = errors.New("ORD-002", errors.SeverityMedium, "Invalid order status transition")
    ErrInsufficientInventory = errors.New("ORD-003", errors.SeverityMedium, "Insufficient inventory")
    ErrPaymentDeclined = errors.New("ORD-004", errors.SeverityHigh, "Payment declined")
    ErrAlreadyCancelled = errors.New("ORD-005", errors.SeverityLow, "Order already cancelled")
    ErrCustomerNotFound = errors.New("ORD-006", errors.SeverityMedium, "Customer not found")
    ErrInvalidTotal = errors.New("ORD-007", errors.SeverityLow, "Invalid order total")
    ErrInternal = errors.New("ORD-500", errors.SeverityHigh, "Internal order processing error")
)

// Factory functions with context
func OrderNotFound(orderID string) *errors.ServiceError {
    return ErrOrderNotFound.WithContext("order_id", orderID)
}

func InvalidStatusTransition(orderID, fromStatus, toStatus string) *errors.ServiceError {
    return ErrInvalidStatus.
        WithContext("order_id", orderID).
        WithContext("from_status", fromStatus).
        WithContext("to_status", toStatus)
}

func InsufficientInventory(productID string, requested, available int) *errors.ServiceError {
    return ErrInsufficientInventory.
        WithContext("product_id", productID).
        WithContext("requested", requested).
        WithContext("available", available)
}
```

### Usage in Services

```go
package services

import (
    "context"
    
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
    orderErrors "myservice/internal/domain/errors"
)

type OrderService struct {
    repo   *OrderRepository
    logger *logger.Logger
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*Order, error) {
    order, err := s.repo.GetByID(ctx, orderID)
    
    if err != nil {
        return nil, orderErrors.OrderNotFound(orderID)  // ✅ Returns ServiceError
    }
    
    return order, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID, newStatus string) error {
    order, err := s.GetOrder(ctx, orderID)
    if err != nil {
        return err
    }
    
    if !s.isValidTransition(order.Status, newStatus) {
        return orderErrors.InvalidStatusTransition(orderID, order.Status, newStatus)
    }
    
    order.Status = newStatus
    return s.repo.Update(ctx, order)
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID string) error {
    order, err := s.GetOrder(ctx, orderID)
    if err != nil {
        return err
    }
    
    if order.Status == "Cancelled" {
        return orderErrors.ErrAlreadyCancelled.WithContext("order_id", orderID)
    }
    
    // Process cancellation
    return nil
}
```

### Error Handling Middleware

```go
package middleware

import (
    "encoding/json"
    "net/http"
    
    "github.com/your-github-org/ai-scaffolder/core/go/errors"
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
    "go.uber.org/zap"
)

type ErrorResponse struct {
    Error         string                 `json:"error"`
    Message       string                 `json:"message"`
    Details       map[string]interface{} `json:"details,omitempty"`
    CorrelationID string                 `json:"correlationId,omitempty"`
}

func ErrorHandler(log *logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if recovered := recover(); recovered != nil {
                    correlationID, _ := r.Context().Value(logger.CorrelationIDKey).(string)
                    
                    if svcErr, ok := recovered.(*errors.ServiceError); ok {
                        log.Warn("Service error",
                            zap.String("code", svcErr.Code),
                            zap.String("message", svcErr.Message),
                            zap.String("correlation_id", correlationID))
                        
                        w.Header().Set("Content-Type", "application/json")
                        w.WriteHeader(svcErr.HTTPStatus())
                        
                        json.NewEncoder(w).Encode(ErrorResponse{
                            Error:         svcErr.Code,
                            Message:       svcErr.Message,
                            Details:       svcErr.Context,
                            CorrelationID: correlationID,
                        })
                        return
                    }
                    
                    // Unknown error
                    log.Error("Unhandled panic",
                        zap.Any("error", recovered),
                        zap.String("correlation_id", correlationID))
                    
                    w.Header().Set("Content-Type", "application/json")
                    w.WriteHeader(http.StatusInternalServerError)
                    
                    json.NewEncoder(w).Encode(ErrorResponse{
                        Error:         "SYS-500",
                        Message:       "An unexpected error occurred",
                        CorrelationID: correlationID,
                    })
                }
            }()
            
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## Error Code Standards

### Code Format

```
{SERVICE}-{CATEGORY}{NUMBER}

Examples:
ORD-001     Order service, first error
ORD-002     Order service, second error
PAY-001     Payment service, first error
USR-001     User service, first error
SYS-500     System-wide internal error
```

### HTTP Status Mapping

| Error Category | HTTP Status | Example Codes |
|---------------|-------------|---------------|
| Not Found | 404 | ORD-001, USR-001 |
| Validation | 400 | ORD-007, USR-002 |
| Conflict | 409 | ORD-002, ORD-005 |
| Business Rule | 422 | ORD-003, PAY-002 |
| Payment | 402 | PAY-001 |
| Internal | 500 | SYS-500, ORD-500 |

### Severity Levels

| Severity | When to Use |
|----------|-------------|
| Low | User error, validation failure |
| Medium | Business rule violation, not found |
| High | External service failure, payment issues |
| Critical | Data corruption, security issues |

---

## API Error Response Format

```json
{
  "error": "ORD-001",
  "message": "Order not found",
  "details": {
    "order_id": "550e8400-e29b-41d4-a716-446655440000"
  },
  "correlationId": "abc-123-def"
}
```

---

## ❌ FORBIDDEN Patterns

```csharp
// ❌ WRONG - Generic exception
throw new Exception("Order not found");
throw new InvalidOperationException("Bad status");
throw new ArgumentException("Invalid ID");

// ❌ WRONG - String messages without codes
throw new ApplicationException("Something went wrong");
```

```go
// ❌ WRONG - Generic errors
return fmt.Errorf("order not found: %s", orderID)
return errors.New("failed to process")

// ❌ WRONG - Panic without ServiceError
panic("something went wrong")
```

## ✅ CORRECT Patterns

```csharp
// ✅ CORRECT - ServiceError with code
throw OrderErrors.NotFound(orderId);
throw OrderErrors.InvalidStatusTransition(orderId, "Pending", "Shipped");
```

```go
// ✅ CORRECT - ServiceError with code and context
return orderErrors.OrderNotFound(orderID)
return orderErrors.InvalidStatusTransition(orderID, oldStatus, newStatus)
```
