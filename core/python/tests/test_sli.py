"""Tests for SLI module"""

import pytest
from datetime import datetime, timezone
from core.sli import SLITracker, RequestOutcome, SLIMetrics, create_nop_tracker


def test_request_outcome_creation():
    """Test creating a request outcome"""
    outcome = RequestOutcome(
        success=True,
        error_code="",
        latency_seconds=0.123,
        operation="query",
    )

    assert outcome.success is True
    assert outcome.latency_seconds == 0.123
    assert outcome.operation == "query"
    assert isinstance(outcome.timestamp, datetime)


def test_request_outcome_failure():
    """Test creating a failed request outcome"""
    outcome = RequestOutcome(
        success=False,
        error_code="DB-001",
        error_severity="HIGH",
        operation="insert",
    )

    assert outcome.success is False
    assert outcome.error_code == "DB-001"
    assert outcome.error_severity == "HIGH"


def test_sli_tracker_creation(test_registry):
    """Test creating SLI tracker"""
    tracker = SLITracker("test-service", registry=test_registry)

    assert tracker.service_name == "test-service"


def test_sli_tracker_record_success(test_registry):
    """Test recording successful request"""
    tracker = SLITracker("test-service", registry=test_registry)

    outcome = RequestOutcome(
        success=True,
        latency_seconds=0.05,
        operation="process",
    )

    # Should not raise exception
    tracker.record_request(outcome)


def test_sli_tracker_record_failure(test_registry):
    """Test recording failed request"""
    tracker = SLITracker("test-service", registry=test_registry)

    outcome = RequestOutcome(
        success=False,
        error_code="ERR-001",
        error_severity="CRITICAL",
        operation="save",
    )

    # Should not raise exception
    tracker.record_request(outcome)


def test_sli_tracker_record_latency(test_registry):
    """Test recording latency"""
    tracker = SLITracker("test-service", registry=test_registry)

    # Should not raise exception
    tracker.record_latency(0.234, operation="compute")


def test_sli_tracker_get_metrics(test_registry):
    """Test getting metrics snapshot"""
    tracker = SLITracker("test-service", registry=test_registry)

    metrics = tracker.get_metrics()

    assert isinstance(metrics, SLIMetrics)
    assert metrics.total_requests >= 0
    assert metrics.availability >= 0


def test_sli_metrics_dataclass():
    """Test SLI metrics dataclass"""
    metrics = SLIMetrics(
        total_requests=100,
        success_requests=95,
        failed_requests=5,
        availability=95.0,
        latency_p95=120.5,
        latency_p99=250.3,
    )

    assert metrics.total_requests == 100
    assert metrics.success_requests == 95
    assert metrics.availability == 95.0
    assert metrics.latency_p95 == 120.5


def test_create_nop_tracker():
    """Test creating no-op tracker for testing"""
    tracker = create_nop_tracker("test-service")

    # Should work without errors
    outcome = RequestOutcome(success=True, latency_seconds=0.1)
    tracker.record_request(outcome)
    tracker.record_latency(0.2)
