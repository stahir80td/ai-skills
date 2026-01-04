"""
HomeGuard Core Package
Python implementation matching Go core package patterns

Provides:
- Structured logging with correlation IDs
- SLI tracking (availability, latency, error rate)
- Service metrics (Prometheus)
- Circuit breaker and retry logic
- Error registry with SOD scores
- Analytics utilities
"""

__version__ = "1.0.0"

from .logger import Logger, setup_logger
from .metrics import ServiceMetrics
from .sli import SLITracker, RequestOutcome
from .errors import ServiceError, ErrorRegistry, Severity
from .reliability import CircuitBreaker, RetryPolicy

__all__ = [
    "Logger",
    "setup_logger",
    "ServiceMetrics",
    "SLITracker",
    "RequestOutcome",
    "ServiceError",
    "ErrorRegistry",
    "Severity",
    "CircuitBreaker",
    "RetryPolicy",
]
