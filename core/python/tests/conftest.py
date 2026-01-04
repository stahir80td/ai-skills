"""Test configuration and fixtures"""

import pytest
from prometheus_client import CollectorRegistry


@pytest.fixture
def test_registry():
    """Create a test-specific Prometheus registry"""
    return CollectorRegistry()


@pytest.fixture
def sample_values():
    """Sample numeric values for testing"""
    return [10.0, 15.0, 20.0, 25.0, 30.0, 35.0, 40.0, 45.0, 50.0]


@pytest.fixture
def sample_timestamps():
    """Sample timestamps for testing"""
    from datetime import datetime, timedelta

    base = datetime.now()
    return [base + timedelta(seconds=i * 60) for i in range(10)]
