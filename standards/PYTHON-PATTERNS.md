# Python Patterns for AI Services

This document defines the standard patterns for Python services including SLI, SOD, SRE, contextual logging, error repository, and Prometheus metrics.

---

## Service Level Indicators (SLI)

Track key performance metrics using prometheus_client:

```python
from prometheus_client import Counter, Histogram, Gauge, Summary
from functools import wraps
import time

# Availability SLI - percentage of successful requests
REQUESTS_TOTAL = Counter(
    'order_service_requests_total',
    'Total requests',
    ['method', 'endpoint', 'status']
)

# Latency SLI - request duration histogram
REQUEST_DURATION = Histogram(
    'order_service_request_duration_seconds',
    'Request duration',
    ['method', 'endpoint'],
    buckets=[.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10]
)

# Throughput SLI - orders processed per second
ORDERS_PROCESSED = Counter(
    'order_service_orders_processed_total',
    'Orders processed',
    ['status']
)


def record_request(method: str, endpoint: str, status_code: int, duration: float):
    """Record a request for SLI metrics."""
    status = 'success' if 200 <= status_code < 400 else 'error'
    REQUESTS_TOTAL.labels(method=method, endpoint=endpoint, status=status).inc()
    REQUEST_DURATION.labels(method=method, endpoint=endpoint).observe(duration)


def record_order_processed(status: str):
    """Record an order processed."""
    ORDERS_PROCESSED.labels(status=status).inc()
```

### SLI Middleware (FastAPI)

```python
from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware
import time


class SLIMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        start_time = time.perf_counter()
        
        response = await call_next(request)
        
        duration = time.perf_counter() - start_time
        
        # Get route path pattern (not the actual path with params)
        route_path = request.scope.get('path', request.url.path)
        for route in request.app.routes:
            if hasattr(route, 'path') and route.matches(request.scope)[0]:
                route_path = route.path
                break
        
        record_request(
            method=request.method,
            endpoint=route_path,
            status_code=response.status_code,
            duration=duration
        )
        
        return response
```

---

## Service Oriented Design (SOD)

Structure services with clear separation of concerns:

```
src/
├── main.py                    # Entry point
├── api/
│   ├── __init__.py
│   ├── routes/               # HTTP endpoints
│   │   ├── __init__.py
│   │   └── orders.py
│   └── middleware/           # Cross-cutting concerns
│       ├── __init__.py
│       └── correlation.py
├── domain/
│   ├── __init__.py
│   ├── models/               # Domain entities
│   │   ├── __init__.py
│   │   └── order.py
│   ├── services/             # Business logic
│   │   ├── __init__.py
│   │   └── order_service.py
│   └── errors/               # Domain errors
│       ├── __init__.py
│       └── order_errors.py
└── infrastructure/
    ├── __init__.py
    ├── repositories/         # Data access
    │   ├── __init__.py
    │   └── order_repository.py
    ├── kafka/               # Event publishing
    │   ├── __init__.py
    │   └── producer.py
    └── redis/               # Caching
        ├── __init__.py
        └── cache.py
```

### Domain Service Pattern

```python
from dataclasses import dataclass
from typing import List
import structlog

from domain.models.order import Order, OrderItem
from domain.errors.order_errors import OrderErrors
from infrastructure.repositories.order_repository import OrderRepository
from infrastructure.kafka.producer import EventPublisher


@dataclass
class CreateOrderRequest:
    customer_id: str
    items: List[OrderItem]


class OrderService:
    def __init__(
        self,
        repository: OrderRepository,
        publisher: EventPublisher,
        logger: structlog.BoundLogger
    ):
        self.repository = repository
        self.publisher = publisher
        self.logger = logger
    
    async def create_order(self, request: CreateOrderRequest) -> Order:
        # Business validation
        if not request.items:
            raise OrderErrors.empty_order()
        
        # Domain logic
        order = Order.create(request.customer_id, request.items)
        
        # Persist
        await self.repository.save(order)
        
        # Publish event
        await self.publisher.publish(
            topic='orders.order.created',
            event={
                'order_id': order.id,
                'customer_id': order.customer_id,
                'total': order.total
            }
        )
        
        self.logger.info(
            "order_created",
            order_id=order.id,
            customer_id=order.customer_id
        )
        
        return order
```

---

## Site Reliability Engineering (SRE)

### Health Checks

```python
from fastapi import APIRouter, Response
from typing import Dict, Any
import asyncio

router = APIRouter()


class HealthChecker:
    def __init__(self, name: str):
        self.name = name
    
    async def check(self) -> bool:
        raise NotImplementedError


class SQLHealthChecker(HealthChecker):
    def __init__(self, db_pool):
        super().__init__("sql")
        self.db_pool = db_pool
    
    async def check(self) -> bool:
        try:
            async with self.db_pool.acquire() as conn:
                await conn.execute("SELECT 1")
            return True
        except Exception:
            return False


class RedisHealthChecker(HealthChecker):
    def __init__(self, redis_client):
        super().__init__("redis")
        self.redis_client = redis_client
    
    async def check(self) -> bool:
        try:
            await self.redis_client.ping()
            return True
        except Exception:
            return False


class HealthService:
    def __init__(self, checkers: list[HealthChecker]):
        self.checkers = checkers
    
    async def check_liveness(self) -> Dict[str, Any]:
        return {"status": "ok"}
    
    async def check_readiness(self) -> Dict[str, Any]:
        results = {}
        all_healthy = True
        
        checks = await asyncio.gather(
            *[checker.check() for checker in self.checkers],
            return_exceptions=True
        )
        
        for checker, result in zip(self.checkers, checks):
            if isinstance(result, Exception):
                results[checker.name] = str(result)
                all_healthy = False
            elif result:
                results[checker.name] = "ok"
            else:
                results[checker.name] = "unhealthy"
                all_healthy = False
        
        return {
            "status": "ok" if all_healthy else "degraded",
            "checks": results
        }


@router.get("/health/live")
async def liveness():
    return {"status": "ok"}


@router.get("/health/ready")
async def readiness(health_service: HealthService):
    result = await health_service.check_readiness()
    status_code = 200 if result["status"] == "ok" else 503
    return Response(
        content=json.dumps(result),
        media_type="application/json",
        status_code=status_code
    )
```

### Circuit Breaker

```python
import asyncio
from datetime import datetime, timedelta
from enum import Enum
from typing import Callable, TypeVar, Any
from functools import wraps

T = TypeVar('T')


class CircuitState(Enum):
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"


class CircuitBreaker:
    def __init__(
        self,
        failure_threshold: int = 5,
        reset_timeout: timedelta = timedelta(seconds=30)
    ):
        self.failure_threshold = failure_threshold
        self.reset_timeout = reset_timeout
        self.failure_count = 0
        self.last_failure: datetime | None = None
        self.state = CircuitState.CLOSED
        self._lock = asyncio.Lock()
    
    async def execute(self, func: Callable[..., T], *args, **kwargs) -> T:
        async with self._lock:
            # Check if we should transition from open to half-open
            if (
                self.state == CircuitState.OPEN
                and self.last_failure
                and datetime.now() - self.last_failure > self.reset_timeout
            ):
                self.state = CircuitState.HALF_OPEN
            
            if self.state == CircuitState.OPEN:
                raise CircuitBreakerOpenError("Circuit breaker is open")
        
        try:
            result = await func(*args, **kwargs)
            
            async with self._lock:
                self.failure_count = 0
                self.state = CircuitState.CLOSED
            
            return result
        
        except Exception as e:
            async with self._lock:
                self.failure_count += 1
                self.last_failure = datetime.now()
                
                if self.failure_count >= self.failure_threshold:
                    self.state = CircuitState.OPEN
            
            raise


class CircuitBreakerOpenError(Exception):
    pass


def circuit_breaker(cb: CircuitBreaker):
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            return await cb.execute(func, *args, **kwargs)
        return wrapper
    return decorator
```

### Retry with Exponential Backoff

```python
import asyncio
from typing import TypeVar, Callable, Type
from functools import wraps
import random

T = TypeVar('T')


async def retry_with_backoff(
    func: Callable[..., T],
    max_retries: int = 3,
    initial_wait: float = 1.0,
    max_wait: float = 30.0,
    exponential_base: float = 2,
    jitter: bool = True,
    retry_exceptions: tuple[Type[Exception], ...] = (Exception,)
) -> T:
    """Retry a function with exponential backoff."""
    last_exception = None
    
    for attempt in range(max_retries + 1):
        try:
            return await func()
        except retry_exceptions as e:
            last_exception = e
            
            if attempt < max_retries:
                wait = min(initial_wait * (exponential_base ** attempt), max_wait)
                if jitter:
                    wait = wait * (0.5 + random.random())
                await asyncio.sleep(wait)
    
    raise last_exception


def retry(
    max_retries: int = 3,
    initial_wait: float = 1.0,
    retry_exceptions: tuple[Type[Exception], ...] = (Exception,)
):
    """Decorator for retry with exponential backoff."""
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            return await retry_with_backoff(
                lambda: func(*args, **kwargs),
                max_retries=max_retries,
                initial_wait=initial_wait,
                retry_exceptions=retry_exceptions
            )
        return wrapper
    return decorator
```

---

## Contextual Logging

Always include context in log entries using structlog:

```python
import structlog
from contextvars import ContextVar
from typing import Any
import uuid

# Context variables for request-scoped data
correlation_id_var: ContextVar[str] = ContextVar('correlation_id', default='')
user_id_var: ContextVar[str] = ContextVar('user_id', default='')


def configure_logging(service_name: str):
    """Configure structlog with JSON output."""
    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.JSONRenderer()
        ],
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(),
        wrapper_class=structlog.make_filtering_bound_logger(20),  # INFO level
        cache_logger_on_first_use=True,
    )
    
    # Bind service name globally
    structlog.contextvars.bind_contextvars(service=service_name)


def get_logger() -> structlog.BoundLogger:
    """Get a logger with current context."""
    return structlog.get_logger()


# Middleware for correlation ID
from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware


class CorrelationIdMiddleware(BaseHTTPMiddleware):
    CORRELATION_ID_HEADER = "X-Correlation-ID"
    
    async def dispatch(self, request: Request, call_next):
        correlation_id = request.headers.get(
            self.CORRELATION_ID_HEADER,
            str(uuid.uuid4())
        )
        
        # Set context var
        correlation_id_var.set(correlation_id)
        structlog.contextvars.bind_contextvars(correlation_id=correlation_id)
        
        response = await call_next(request)
        
        # Add to response headers
        response.headers[self.CORRELATION_ID_HEADER] = correlation_id
        
        # Clear context
        structlog.contextvars.unbind_contextvars('correlation_id')
        
        return response


# Example usage in a service
class OrderService:
    def __init__(self):
        self.logger = get_logger()
    
    async def process_order(self, order_id: str):
        log = self.logger.bind(order_id=order_id)
        
        log.info("processing_order_started")
        
        try:
            # ... processing logic
            log.info("processing_order_completed", duration_ms=elapsed)
        except Exception as e:
            log.error("processing_order_failed", error=str(e))
            raise
```

---

## Error Repository

Centralized error definitions with codes:

```python
from dataclasses import dataclass
from typing import Any, Dict, Optional
from http import HTTPStatus


@dataclass
class ServiceError(Exception):
    code: str
    message: str
    status_code: int = HTTPStatus.INTERNAL_SERVER_ERROR
    details: Optional[Dict[str, Any]] = None
    
    def __str__(self):
        return f"[{self.code}] {self.message}"
    
    def to_dict(self) -> Dict[str, Any]:
        result = {
            "code": self.code,
            "message": self.message,
        }
        if self.details:
            result["details"] = self.details
        return result


class OrderErrors:
    """Error registry for order-related errors."""
    
    @staticmethod
    def not_found(order_id: str) -> ServiceError:
        return ServiceError(
            code="ORD-001",
            message="Order not found",
            status_code=HTTPStatus.NOT_FOUND,
            details={"order_id": order_id}
        )
    
    @staticmethod
    def empty_order() -> ServiceError:
        return ServiceError(
            code="ORD-002",
            message="Empty order not allowed",
            status_code=HTTPStatus.BAD_REQUEST
        )
    
    @staticmethod
    def invalid_status_transition(from_status: str, to_status: str) -> ServiceError:
        return ServiceError(
            code="ORD-003",
            message="Invalid order status transition",
            status_code=HTTPStatus.CONFLICT,
            details={"from": from_status, "to": to_status}
        )
    
    @staticmethod
    def insufficient_inventory(sku: str, requested: int, available: int) -> ServiceError:
        return ServiceError(
            code="ORD-004",
            message="Insufficient inventory",
            status_code=HTTPStatus.CONFLICT,
            details={"sku": sku, "requested": requested, "available": available}
        )
    
    @staticmethod
    def payment_failed(reason: str) -> ServiceError:
        return ServiceError(
            code="ORD-005",
            message="Payment failed",
            status_code=HTTPStatus.PAYMENT_REQUIRED,
            details={"reason": reason}
        )
```

### Error Handler Middleware

```python
from fastapi import Request, FastAPI
from fastapi.responses import JSONResponse
from datetime import datetime
import structlog

logger = structlog.get_logger()


def configure_error_handlers(app: FastAPI):
    @app.exception_handler(ServiceError)
    async def service_error_handler(request: Request, exc: ServiceError):
        logger.warning(
            "service_error",
            code=exc.code,
            message=exc.message,
            details=exc.details,
            path=request.url.path
        )
        
        return JSONResponse(
            status_code=exc.status_code,
            content={
                "error": {
                    **exc.to_dict(),
                    "timestamp": datetime.utcnow().isoformat()
                }
            }
        )
    
    @app.exception_handler(Exception)
    async def generic_error_handler(request: Request, exc: Exception):
        logger.error(
            "unhandled_error",
            error=str(exc),
            error_type=type(exc).__name__,
            path=request.url.path
        )
        
        return JSONResponse(
            status_code=HTTPStatus.INTERNAL_SERVER_ERROR,
            content={
                "error": {
                    "code": "SYS-001",
                    "message": "Internal server error",
                    "timestamp": datetime.utcnow().isoformat()
                }
            }
        )
```

---

## Prometheus Metrics Endpoint

### Setup in main.py

```python
from fastapi import FastAPI
from prometheus_client import make_asgi_app, Counter, Histogram, Gauge, Summary
from starlette.middleware.base import BaseHTTPMiddleware

app = FastAPI(title="Order Service")

# Mount Prometheus metrics endpoint
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)

# Add SLI middleware
app.add_middleware(SLIMiddleware)

# Configure logging
configure_logging("order-service")

# Configure error handlers
configure_error_handlers(app)

# Health endpoints
app.include_router(health_router)

# API routes
app.include_router(order_router, prefix="/api/v1/orders")
```

### Custom Business Metrics

```python
from prometheus_client import Counter, Gauge, Histogram, Summary

# Counter - things that only go up
ORDERS_CREATED = Counter(
    'orders_created_total',
    'Total orders created',
    ['source', 'customer_type']
)

# Gauge - things that go up and down
PENDING_ORDERS = Gauge(
    'orders_pending',
    'Current pending orders count'
)

# Histogram - distributions (latency, sizes)
ORDER_VALUE = Histogram(
    'order_value_dollars',
    'Order value distribution',
    buckets=[10, 25, 50, 100, 250, 500, 1000, 2500, 5000]
)

# Summary - percentiles
PROCESSING_TIME = Summary(
    'order_processing_seconds',
    'Order processing time'
)


class OrderMetrics:
    @staticmethod
    def record_order_created(source: str, customer_type: str, value: float):
        ORDERS_CREATED.labels(source=source, customer_type=customer_type).inc()
        ORDER_VALUE.observe(value)
        PENDING_ORDERS.inc()
    
    @staticmethod
    def record_order_completed(processing_seconds: float):
        PENDING_ORDERS.dec()
        PROCESSING_TIME.observe(processing_seconds)
```

---

## Complete main.py Example

```python
from fastapi import FastAPI
from prometheus_client import make_asgi_app
import structlog
import uvicorn

from api.routes import orders, health
from api.middleware.correlation import CorrelationIdMiddleware
from api.middleware.sli import SLIMiddleware
from domain.errors.base import configure_error_handlers
from infrastructure.logging import configure_logging

# Configure logging first
configure_logging("order-service")
logger = structlog.get_logger()

# Create FastAPI app
app = FastAPI(
    title="Order Service",
    description="Order management service",
    version="1.0.0"
)

# Middleware (order matters - first added = outermost)
app.add_middleware(CorrelationIdMiddleware)
app.add_middleware(SLIMiddleware)

# Error handlers
configure_error_handlers(app)

# Mount Prometheus metrics
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)

# Health endpoints
app.include_router(health.router, tags=["Health"])

# API routes
app.include_router(
    orders.router,
    prefix="/api/v1/orders",
    tags=["Orders"]
)


@app.on_event("startup")
async def startup():
    logger.info("application_starting")
    # Initialize database connections, etc.


@app.on_event("shutdown")
async def shutdown():
    logger.info("application_shutting_down")
    # Cleanup connections


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8000,
        reload=True
    )
```
