"""Tests for metrics module"""

import pytest
from core.metrics import ServiceMetrics, IncidentMetrics, create_nop_metrics


def test_service_metrics_creation(test_registry):
    """Test creating service metrics"""
    metrics = ServiceMetrics("iot_homeguard", "test_service", registry=test_registry)

    assert metrics.namespace == "iot_homeguard"
    assert metrics.subsystem == "test_service"


def test_service_metrics_record_request(test_registry):
    """Test recording a request"""
    metrics = ServiceMetrics("test", "service", registry=test_registry)

    # Should not raise exception
    metrics.record_request("query", "test-service", "GET", 0.123)


def test_service_metrics_record_error(test_registry):
    """Test recording an error"""
    metrics = ServiceMetrics("test", "service", registry=test_registry)

    # Should not raise exception
    metrics.record_error("database_error", "test-service", "database")


def test_service_metrics_cache_operations(test_registry):
    """Test cache hit/miss recording"""
    metrics = ServiceMetrics("test", "service", registry=test_registry)

    # Should not raise exceptions
    metrics.record_cache_hit("redis", "test-service")
    metrics.record_cache_miss("redis", "test-service")


def test_service_metrics_active_requests(test_registry):
    """Test active request tracking"""
    metrics = ServiceMetrics("test", "service", registry=test_registry)

    # Should not raise exceptions
    metrics.inc_active_requests("process", "test-service")
    metrics.dec_active_requests("process", "test-service")


def test_incident_metrics_creation(test_registry):
    """Test creating incident metrics"""
    metrics = IncidentMetrics("test-service", registry=test_registry)

    assert metrics.service_name == "test-service"


def test_incident_metrics_lifecycle(test_registry):
    """Test incident lifecycle tracking"""
    metrics = IncidentMetrics("test-service", registry=test_registry)

    # Start incident
    metrics.start_incident("high", "database_failure")

    # Resolve incident
    metrics.resolve_incident("high", "database_failure", 15.5)


def test_create_nop_metrics():
    """Test creating no-op metrics for testing"""
    metrics = create_nop_metrics()

    # Should work without errors
    metrics.record_request("test", "service", "GET", 0.1)
    metrics.record_error("test", "service", "component")
