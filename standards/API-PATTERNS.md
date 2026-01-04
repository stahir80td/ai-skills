# API Standards - REST Conventions

> **Consistent REST API patterns for Go, Python, and .NET backends**

This document defines the API standards for generated services. All backends follow these conventions for consistent client integration.

---

## URL Structure

### Base Pattern

```
/api/v1/{resource}
/api/v1/{resource}/{id}
/api/v1/{resource}/{id}/{sub-resource}
```

### Examples

```
GET    /api/v1/orders              # List all orders
POST   /api/v1/orders              # Create an order
GET    /api/v1/orders/{id}         # Get single order
PUT    /api/v1/orders/{id}         # Update order
DELETE /api/v1/orders/{id}         # Delete order
GET    /api/v1/orders/{id}/items   # Get order items
POST   /api/v1/orders/{id}/items   # Add item to order
```

### Naming Rules

| Rule | Good | Bad |
|------|------|-----|
| Plural nouns | `/orders` | `/order`, `/getOrders` |
| Lowercase | `/orders` | `/Orders` |
| Hyphens for multi-word | `/order-items` | `/orderItems`, `/order_items` |
| No verbs in URL | `POST /orders` | `POST /createOrder` |
| Version prefix | `/api/v1/` | `/v1/`, `/api/` |

---

## HTTP Methods

| Method | Purpose | Request Body | Response Body | Idempotent |
|--------|---------|--------------|---------------|------------|
| `GET` | Read | None | Resource(s) | Yes |
| `POST` | Create | New resource | Created resource | No |
| `PUT` | Replace | Full resource | Updated resource | Yes |
| `PATCH` | Partial update | Partial resource | Updated resource | No |
| `DELETE` | Remove | None | None or confirmation | Yes |

---

## Status Codes

### Success Codes

| Code | When to Use | Response Body |
|------|-------------|---------------|
| `200 OK` | Successful GET, PUT, PATCH | Resource data |
| `201 Created` | Successful POST | Created resource + Location header |
| `204 No Content` | Successful DELETE | None |

### Client Error Codes

| Code | When to Use | Example |
|------|-------------|---------|
| `400 Bad Request` | Invalid request body/params | Validation failed |
| `401 Unauthorized` | Missing/invalid auth | No token |
| `403 Forbidden` | Valid auth but no permission | Wrong role |
| `404 Not Found` | Resource doesn't exist | Order not found |
| `409 Conflict` | State conflict | Duplicate email |
| `422 Unprocessable Entity` | Semantic validation error | Invalid status transition |

### Server Error Codes

| Code | When to Use |
|------|-------------|
| `500 Internal Server Error` | Unexpected error |
| `502 Bad Gateway` | Upstream service error |
| `503 Service Unavailable` | Temporary outage |
| `504 Gateway Timeout` | Upstream timeout |

---

## Request/Response Format

### Standard Response Envelope

```json
// Success response
{
  "data": { ... },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "requestId": "req-abc-123"
  }
}

// List response with pagination
{
  "data": [ ... ],
  "meta": {
    "total": 150,
    "page": 1,
    "pageSize": 20,
    "totalPages": 8,
    "timestamp": "2024-01-15T10:30:00Z",
    "requestId": "req-abc-123"
  }
}

// Error response
{
  "error": {
    "code": "ORDER-001",
    "message": "Order not found",
    "details": "No order exists with ID: ord-123",
    "severity": "warning",
    "timestamp": "2024-01-15T10:30:00Z",
    "requestId": "req-abc-123",
    "correlationId": "corr-xyz-789"
  }
}
```

### Go Implementation

```go
// Response types
type Response[T any] struct {
    Data T        `json:"data"`
    Meta MetaInfo `json:"meta"`
}

type ListResponse[T any] struct {
    Data []T          `json:"data"`
    Meta PaginatedMeta `json:"meta"`
}

type ErrorResponse struct {
    Error ErrorInfo `json:"error"`
}

type MetaInfo struct {
    Timestamp   time.Time `json:"timestamp"`
    RequestID   string    `json:"requestId"`
}

type PaginatedMeta struct {
    MetaInfo
    Total      int `json:"total"`
    Page       int `json:"page"`
    PageSize   int `json:"pageSize"`
    TotalPages int `json:"totalPages"`
}

type ErrorInfo struct {
    Code          string    `json:"code"`
    Message       string    `json:"message"`
    Details       string    `json:"details,omitempty"`
    Severity      string    `json:"severity"`
    Timestamp     time.Time `json:"timestamp"`
    RequestID     string    `json:"requestId"`
    CorrelationID string    `json:"correlationId,omitempty"`
}

// Handler helpers
func RespondJSON[T any](w http.ResponseWriter, status int, data T) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(Response[T]{
        Data: data,
        Meta: MetaInfo{
            Timestamp: time.Now().UTC(),
            RequestID: GetRequestID(w),
        },
    })
}

func RespondError(w http.ResponseWriter, status int, err *errors.ServiceError) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{
        Error: ErrorInfo{
            Code:      err.Code,
            Message:   err.Message,
            Severity:  err.Severity,
            Timestamp: time.Now().UTC(),
            RequestID: GetRequestID(w),
        },
    })
}
```

### Python Implementation

```python
from dataclasses import dataclass
from datetime import datetime
from typing import TypeVar, Generic, List, Optional
from fastapi import Response
from pydantic import BaseModel

T = TypeVar('T')

class MetaInfo(BaseModel):
    timestamp: datetime
    request_id: str

class PaginatedMeta(MetaInfo):
    total: int
    page: int
    page_size: int
    total_pages: int

class ApiResponse(BaseModel, Generic[T]):
    data: T
    meta: MetaInfo

class ListApiResponse(BaseModel, Generic[T]):
    data: List[T]
    meta: PaginatedMeta

class ErrorInfo(BaseModel):
    code: str
    message: str
    details: Optional[str] = None
    severity: str
    timestamp: datetime
    request_id: str
    correlation_id: Optional[str] = None

class ErrorResponse(BaseModel):
    error: ErrorInfo
```

---

## Query Parameters

### Pagination

```
GET /api/v1/orders?page=1&pageSize=20
```

| Param | Default | Max | Description |
|-------|---------|-----|-------------|
| `page` | 1 | - | Page number (1-indexed) |
| `pageSize` | 20 | 100 | Items per page |

### Filtering

```
GET /api/v1/orders?status=pending&customerId=cust-123
GET /api/v1/orders?createdAfter=2024-01-01&createdBefore=2024-01-31
```

| Pattern | Example | Description |
|---------|---------|-------------|
| Exact match | `?status=pending` | Field equals value |
| Date range | `?createdAfter=...&createdBefore=...` | Between dates |
| Contains | `?name=*john*` | Contains substring |
| Multiple values | `?status=pending,processing` | IN clause |

### Sorting

```
GET /api/v1/orders?sort=createdAt:desc
GET /api/v1/orders?sort=status:asc,createdAt:desc
```

Format: `field:direction` (comma-separated for multiple)

### Search

```
GET /api/v1/orders?q=smartphone
```

Full-text search across searchable fields.

---

## Headers

### Required Request Headers

```
Content-Type: application/json
Authorization: Bearer <token>
X-Correlation-ID: <uuid>          # Optional, generated if missing
```

### Response Headers

```
Content-Type: application/json
X-Request-ID: <uuid>              # Unique per request
X-Correlation-ID: <uuid>          # Same as request or generated
X-Response-Time: 45ms             # Processing time
```

### Go Middleware

```go
func CorrelationIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        correlationID := r.Header.Get("X-Correlation-ID")
        if correlationID == "" {
            correlationID = uuid.New().String()
        }
        
        requestID := uuid.New().String()
        
        ctx := context.WithValue(r.Context(), "correlationID", correlationID)
        ctx = context.WithValue(ctx, "requestID", requestID)
        
        w.Header().Set("X-Correlation-ID", correlationID)
        w.Header().Set("X-Request-ID", requestID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## Authentication

### Bearer Token

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### Token Claims (JWT)

```json
{
  "sub": "user-123",
  "email": "user@example.com",
  "roles": ["admin", "user"],
  "exp": 1705312200,
  "iat": 1705308600
}
```

### Auth Middleware (Go)

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            RespondError(w, 401, errors.NewServiceError("AUTH-001", "Missing authorization header"))
            return
        }
        
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            RespondError(w, 401, errors.NewServiceError("AUTH-002", "Invalid authorization format"))
            return
        }
        
        claims, err := validateToken(parts[1])
        if err != nil {
            RespondError(w, 401, errors.NewServiceError("AUTH-003", "Invalid token"))
            return
        }
        
        ctx := context.WithValue(r.Context(), "user", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## Validation

### Request Validation

```go
type CreateOrderRequest struct {
    CustomerID string       `json:"customerId" validate:"required,uuid"`
    Items      []OrderItem  `json:"items" validate:"required,min=1,dive"`
    Notes      string       `json:"notes" validate:"max=500"`
}

type OrderItem struct {
    ProductID string  `json:"productId" validate:"required,uuid"`
    Quantity  int     `json:"quantity" validate:"required,min=1,max=100"`
    Price     float64 `json:"price" validate:"required,min=0"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondError(w, 400, errors.NewServiceError("VAL-001", "Invalid JSON body"))
        return
    }
    
    if err := h.validator.Struct(req); err != nil {
        RespondError(w, 400, errors.NewServiceError("VAL-002", formatValidationErrors(err)))
        return
    }
    
    // Process valid request...
}
```

### Validation Error Response

```json
{
  "error": {
    "code": "VAL-002",
    "message": "Validation failed",
    "details": "customerId: required, items[0].quantity: must be at least 1",
    "severity": "warning",
    "timestamp": "2024-01-15T10:30:00Z",
    "requestId": "req-abc-123"
  }
}
```

---

## Error Codes

### Format

```
{DOMAIN}-{NUMBER}
```

### Standard Domains

| Domain | Description | Example |
|--------|-------------|---------|
| `AUTH` | Authentication/Authorization | `AUTH-001` |
| `VAL` | Validation | `VAL-001` |
| `ORDER` | Order domain | `ORDER-001` |
| `USER` | User domain | `USER-001` |
| `INFRA` | Infrastructure | `INFRA-001` |

### Error Registry

```go
var errorRegistry = errors.NewErrorRegistry()

func init() {
    // Auth errors
    errorRegistry.Register(&errors.ErrorDefinition{
        Code:        "AUTH-001",
        Message:     "Missing authorization header",
        Severity:    errors.SeverityWarning,
        HttpStatus:  401,
        Retryable:   false,
    })
    
    // Validation errors
    errorRegistry.Register(&errors.ErrorDefinition{
        Code:        "VAL-001",
        Message:     "Invalid request body",
        Severity:    errors.SeverityWarning,
        HttpStatus:  400,
        Retryable:   false,
    })
    
    // Domain errors
    errorRegistry.Register(&errors.ErrorDefinition{
        Code:        "ORDER-001",
        Message:     "Order not found",
        Severity:    errors.SeverityWarning,
        HttpStatus:  404,
        Retryable:   false,
    })
}
```

---

## Rate Limiting

### Response Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705312200
```

### 429 Too Many Requests

```json
{
  "error": {
    "code": "RATE-001",
    "message": "Rate limit exceeded",
    "details": "Try again in 60 seconds",
    "severity": "warning",
    "timestamp": "2024-01-15T10:30:00Z",
    "requestId": "req-abc-123"
  }
}
```

---

## Health Endpoints

### Liveness (K8s)

```
GET /health/live
```

```json
{
  "status": "ok"
}
```

### Readiness (K8s)

```
GET /health/ready
```

```json
{
  "status": "ok",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "kafka": "ok"
  }
}
```

### Detailed Health

```
GET /health
```

```json
{
  "status": "ok",
  "version": "1.2.3",
  "uptime": "24h30m15s",
  "checks": {
    "database": {
      "status": "ok",
      "latency": "2ms"
    },
    "redis": {
      "status": "ok",
      "latency": "1ms"
    },
    "kafka": {
      "status": "ok",
      "connected": true
    }
  }
}
```

---

## API Documentation

### OpenAPI Spec Location

```
GET /api/v1/docs           # Swagger UI
GET /api/v1/openapi.json   # OpenAPI spec
```

### Generate from Code (Go)

```go
// Use swaggo/swag annotations
// @Summary Create an order
// @Description Create a new order for a customer
// @Tags orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "Order data"
// @Success 201 {object} Response[Order]
// @Failure 400 {object} ErrorResponse
// @Router /orders [post]
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
    // ...
}
```
