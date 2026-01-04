"""Tests for reliability module"""

import pytest
import asyncio
from core.reliability import (
    CircuitBreaker,
    CircuitState,
    RetryPolicy,
    retry_on_exception,
)


def test_circuit_breaker_creation(test_registry):
    """Test creating circuit breaker"""
    cb = CircuitBreaker("test-breaker", max_failures=3, registry=test_registry)

    assert cb.name == "test-breaker"
    assert cb.max_failures == 3
    assert cb.get_state() == CircuitState.CLOSED


def test_circuit_breaker_success(test_registry):
    """Test circuit breaker with successful calls"""
    cb = CircuitBreaker("test-cb", registry=test_registry)

    def successful_operation():
        return "success"

    result = cb.call(successful_operation)
    assert result == "success"
    assert cb.get_state() == CircuitState.CLOSED


def test_circuit_breaker_failure(test_registry):
    """Test circuit breaker with failures"""
    cb = CircuitBreaker("test-cb", max_failures=2, registry=test_registry)

    def failing_operation():
        raise ValueError("Operation failed")

    # First failure
    with pytest.raises(ValueError):
        cb.call(failing_operation)
    assert cb.get_state() == CircuitState.CLOSED

    # Second failure - should open circuit
    with pytest.raises(ValueError):
        cb.call(failing_operation)
    assert cb.get_state() == CircuitState.OPEN


def test_circuit_breaker_open_rejects(test_registry):
    """Test that open circuit rejects requests"""
    cb = CircuitBreaker(
        "test-cb", max_failures=1, timeout_seconds=60, registry=test_registry
    )

    def failing_operation():
        raise ValueError("Fail")

    # Trigger opening
    with pytest.raises(ValueError):
        cb.call(failing_operation)

    assert cb.get_state() == CircuitState.OPEN

    # Next call should be rejected immediately
    with pytest.raises(Exception, match="Circuit breaker.*is OPEN"):
        cb.call(lambda: "should not execute")


@pytest.mark.asyncio
async def test_circuit_breaker_async_success(test_registry):
    """Test async circuit breaker with success"""
    cb = CircuitBreaker("async-cb", registry=test_registry)

    async def async_operation():
        await asyncio.sleep(0.01)
        return "async success"

    result = await cb.call_async(async_operation)
    assert result == "async success"


@pytest.mark.asyncio
async def test_circuit_breaker_async_failure(test_registry):
    """Test async circuit breaker with failure"""
    cb = CircuitBreaker("async-cb", max_failures=2, registry=test_registry)

    async def async_failing():
        await asyncio.sleep(0.01)
        raise ValueError("Async fail")

    # Trigger failures
    with pytest.raises(ValueError):
        await cb.call_async(async_failing)

    with pytest.raises(ValueError):
        await cb.call_async(async_failing)

    assert cb.get_state() == CircuitState.OPEN


def test_retry_policy_exponential():
    """Test exponential backoff retry policy"""
    retry_decorator = RetryPolicy.with_exponential_backoff(
        max_attempts=3,
        initial_delay=0.01,
        max_delay=1.0,
    )

    attempt_count = 0

    @retry_decorator
    def flaky_operation():
        nonlocal attempt_count
        attempt_count += 1
        if attempt_count < 3:
            raise ValueError("Not yet")
        return "success"

    result = flaky_operation()
    assert result == "success"
    assert attempt_count == 3


def test_retry_policy_fixed_delay():
    """Test fixed delay retry policy"""
    retry_decorator = RetryPolicy.with_fixed_delay(
        max_attempts=2,
        delay=0.01,
    )

    attempt_count = 0

    @retry_decorator
    def operation():
        nonlocal attempt_count
        attempt_count += 1
        if attempt_count < 2:
            raise ValueError("Try again")
        return "done"

    result = operation()
    assert result == "done"
    assert attempt_count == 2


def test_retry_on_exception_decorator():
    """Test retry_on_exception decorator"""
    attempt_count = 0

    @retry_on_exception(max_attempts=3, initial_delay=0.01)
    def decorated_operation():
        nonlocal attempt_count
        attempt_count += 1
        if attempt_count < 2:
            raise RuntimeError("Retry me")
        return "completed"

    result = decorated_operation()
    assert result == "completed"
    assert attempt_count == 2


def test_retry_exhausted():
    """Test that retry gives up after max attempts"""

    @retry_on_exception(max_attempts=2, initial_delay=0.01)
    def always_failing():
        raise ValueError("Always fails")

    with pytest.raises(ValueError):
        always_failing()
