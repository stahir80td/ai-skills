"""
Reliability Patterns - Python equivalent of Go core/reliability

Provides:
- Circuit breaker pattern
- Retry with exponential backoff
- Timeout handling
"""

from enum import Enum
from typing import Callable, Any, Optional, TypeVar
from datetime import datetime, timedelta
import time
import asyncio
from functools import wraps
from prometheus_client import Counter, Gauge, CollectorRegistry
import tenacity


T = TypeVar("T")


class CircuitState(str, Enum):
    """Circuit breaker states - matches Go CircuitState"""

    CLOSED = "closed"  # Normal operation
    HALF_OPEN = "half_open"  # Testing if service recovered
    OPEN = "open"  # Rejecting requests


class CircuitBreaker:
    """
    Circuit breaker to prevent cascading failures
    Matches Go reliability.CircuitBreaker
    """

    def __init__(
        self,
        name: str,
        max_failures: int = 5,
        timeout_seconds: float = 60.0,
        half_open_requests: int = 3,
        registry: Optional[CollectorRegistry] = None,
        enabled: bool = True,
    ):
        self.name = name
        self.max_failures = max_failures
        self.timeout = timedelta(seconds=timeout_seconds)
        self.half_open_max_requests = half_open_requests
        self.enabled = enabled  # If False, circuit breaker is bypassed

        self.state = CircuitState.CLOSED
        self.failures = 0
        self.last_fail_time: Optional[datetime] = None
        self.half_open_attempts = 0

        # Metrics
        self.state_gauge = Gauge(
            "circuit_breaker_state",
            "Circuit breaker state (0=closed, 1=half_open, 2=open)",
            ["name"],
            registry=registry,
        )
        self.state_gauge.labels(name=name).set(0)

        self.requests_total = Counter(
            "circuit_breaker_requests_total",
            "Total requests through circuit breaker",
            ["name", "state", "result"],
            registry=registry,
        )

        self.errors_total = Counter(
            "circuit_breaker_errors_total",
            "Total errors in circuit breaker",
            ["name"],
            registry=registry,
        )

    def call(self, func: Callable[[], T]) -> T:
        """
        Execute function with circuit breaker protection
        Matches Go CircuitBreaker.Call
        """
        # Bypass if disabled
        if not self.enabled:
            return func()

        # Check if circuit should transition from OPEN to HALF_OPEN
        if self.state == CircuitState.OPEN:
            if (
                self.last_fail_time
                and datetime.now() - self.last_fail_time > self.timeout
            ):
                self._transition_to_half_open()
            else:
                self.requests_total.labels(
                    name=self.name, state="open", result="rejected"
                ).inc()
                raise Exception(f"Circuit breaker '{self.name}' is OPEN")

        # Execute the function
        try:
            result = func()
            self._on_success()
            self.requests_total.labels(
                name=self.name, state=self.state.value, result="success"
            ).inc()
            return result
        except Exception as e:
            self._on_failure()
            self.requests_total.labels(
                name=self.name, state=self.state.value, result="failure"
            ).inc()
            self.errors_total.labels(name=self.name).inc()
            raise e

    async def call_async(self, func: Callable[[], Any]) -> Any:
        """Execute async function with circuit breaker protection"""
        # Bypass if disabled
        if not self.enabled:
            return await func()

        # Check circuit state
        if self.state == CircuitState.OPEN:
            if (
                self.last_fail_time
                and datetime.now() - self.last_fail_time > self.timeout
            ):
                self._transition_to_half_open()
            else:
                self.requests_total.labels(
                    name=self.name, state="open", result="rejected"
                ).inc()
                raise Exception(f"Circuit breaker '{self.name}' is OPEN")

        try:
            result = await func()
            self._on_success()
            self.requests_total.labels(
                name=self.name, state=self.state.value, result="success"
            ).inc()
            return result
        except Exception as e:
            self._on_failure()
            self.requests_total.labels(
                name=self.name, state=self.state.value, result="failure"
            ).inc()
            self.errors_total.labels(name=self.name).inc()
            raise e

    def _on_success(self):
        """Handle successful request"""
        if self.state == CircuitState.HALF_OPEN:
            self.half_open_attempts += 1
            if self.half_open_attempts >= self.half_open_max_requests:
                self._transition_to_closed()
        elif self.state == CircuitState.CLOSED:
            self.failures = 0

    def _on_failure(self):
        """Handle failed request"""
        self.failures += 1
        self.last_fail_time = datetime.now()

        if self.state == CircuitState.HALF_OPEN:
            self._transition_to_open()
        elif self.state == CircuitState.CLOSED and self.failures >= self.max_failures:
            self._transition_to_open()

    def _transition_to_open(self):
        """Transition to OPEN state"""
        self.state = CircuitState.OPEN
        self.state_gauge.labels(name=self.name).set(2)

    def _transition_to_half_open(self):
        """Transition to HALF_OPEN state"""
        self.state = CircuitState.HALF_OPEN
        self.half_open_attempts = 0
        self.state_gauge.labels(name=self.name).set(1)

    def _transition_to_closed(self):
        """Transition to CLOSED state"""
        self.state = CircuitState.CLOSED
        self.failures = 0
        self.half_open_attempts = 0
        self.state_gauge.labels(name=self.name).set(0)

    def get_state(self) -> CircuitState:
        """Get current circuit state"""
        return self.state


class RetryPolicy:
    """
    Retry policy with exponential backoff
    Uses tenacity library (Python equivalent of Go retry patterns)
    """

    @staticmethod
    def with_exponential_backoff(
        max_attempts: int = 3,
        initial_delay: float = 1.0,
        max_delay: float = 60.0,
        multiplier: float = 2.0,
    ):
        """
        Create retry decorator with exponential backoff
        Matches Go reliability retry patterns
        """
        return tenacity.retry(
            stop=tenacity.stop_after_attempt(max_attempts),
            wait=tenacity.wait_exponential(
                multiplier=multiplier, min=initial_delay, max=max_delay
            ),
            reraise=True,
        )

    @staticmethod
    def with_fixed_delay(max_attempts: int = 3, delay: float = 1.0):
        """Create retry decorator with fixed delay"""
        return tenacity.retry(
            stop=tenacity.stop_after_attempt(max_attempts),
            wait=tenacity.wait_fixed(delay),
            reraise=True,
        )


def retry_on_exception(
    max_attempts: int = 3,
    initial_delay: float = 1.0,
    max_delay: float = 60.0,
):
    """
    Decorator for retry with exponential backoff

    Usage:
        @retry_on_exception(max_attempts=5)
        def my_function():
            # code that might fail
    """
    return RetryPolicy.with_exponential_backoff(
        max_attempts=max_attempts,
        initial_delay=initial_delay,
        max_delay=max_delay,
    )
