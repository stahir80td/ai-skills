"""Tests for errors module"""

import pytest
from core.errors import (
    ServiceError,
    ErrorDefinition,
    ErrorRegistry,
    Severity,
)


def test_service_error_creation():
    """Test creating a service error"""
    error = ServiceError(
        code="TEST-001",
        message="Test error message",
        severity=Severity.HIGH,
    )

    assert error.code == "TEST-001"
    assert error.message == "Test error message"
    assert error.severity == Severity.HIGH
    assert error.underlying is None


def test_service_error_with_underlying():
    """Test service error wrapping another exception"""
    original = ValueError("Original error")
    error = ServiceError(
        code="TEST-002",
        message="Wrapped error",
        severity=Severity.MEDIUM,
        underlying=original,
    )

    assert error.underlying == original
    assert "caused by" in str(error)


def test_service_error_with_context():
    """Test adding context to error"""
    error = ServiceError("TEST-003", "Error with context", Severity.LOW)
    error.with_context("user_id", "user-123")
    error.with_context("device_id", "device-456")

    assert error.get_context("user_id") == "user-123"
    assert error.get_context("device_id") == "device-456"
    assert error.get_context("nonexistent") is None


def test_error_definition_creation():
    """Test creating error definition"""
    definition = ErrorDefinition(
        code="TEST-001",
        severity=Severity.CRITICAL,
        description="Critical test error",
        severity_s=9,
        occurrence=5,
        detect_d=8,
        mitigation="Restart the service",
        example="When database connection fails",
    )

    assert definition.code == "TEST-001"
    assert definition.sod_score == 9 * 5 * 8  # 360
    assert definition.mitigation == "Restart the service"


def test_error_registry_register():
    """Test registering error definitions"""
    registry = ErrorRegistry()

    definition = ErrorDefinition(
        code="REG-001",
        severity=Severity.HIGH,
        description="Registry test error",
        severity_s=7,
        occurrence=3,
        detect_d=5,
        mitigation="Check logs",
    )

    registry.register(definition)

    retrieved = registry.get("REG-001")
    assert retrieved is not None
    assert retrieved.code == "REG-001"
    assert retrieved.sod_score == 7 * 3 * 5  # 105


def test_error_registry_new_error():
    """Test creating error from registry"""
    registry = ErrorRegistry()

    registry.register(
        ErrorDefinition(
            code="NEW-001",
            severity=Severity.CRITICAL,
            description="Test",
            severity_s=10,
            occurrence=10,
            detect_d=10,
            mitigation="Fix it",
        )
    )

    error = registry.new_error("NEW-001", "Something went wrong")

    assert error.code == "NEW-001"
    assert error.severity == Severity.CRITICAL
    assert error.message == "Something went wrong"


def test_error_registry_wrap_error():
    """Test wrapping error from registry"""
    registry = ErrorRegistry()

    registry.register(
        ErrorDefinition(
            code="WRAP-001",
            severity=Severity.MEDIUM,
            description="Wrapping test",
            severity_s=5,
            occurrence=5,
            detect_d=5,
            mitigation="Retry",
        )
    )

    original = RuntimeError("Original error")
    wrapped = registry.wrap_error(original, "WRAP-001", "Wrapped message")

    assert wrapped.code == "WRAP-001"
    assert wrapped.severity == Severity.MEDIUM
    assert wrapped.underlying == original


def test_error_registry_unregistered_code():
    """Test creating error with unregistered code"""
    registry = ErrorRegistry()

    # Should still work with default severity
    error = registry.new_error("UNKNOWN-001", "Unknown error")
    assert error.code == "UNKNOWN-001"
    assert error.severity == Severity.MEDIUM


def test_error_registry_list_codes():
    """Test listing all registered error codes"""
    registry = ErrorRegistry()

    registry.register(
        ErrorDefinition("CODE-001", Severity.LOW, "Test 1", 1, 1, 1, "Fix 1")
    )
    registry.register(
        ErrorDefinition("CODE-002", Severity.MEDIUM, "Test 2", 2, 2, 2, "Fix 2")
    )
    registry.register(
        ErrorDefinition("CODE-003", Severity.HIGH, "Test 3", 3, 3, 3, "Fix 3")
    )

    codes = registry.list_codes()
    assert len(codes) == 3
    assert "CODE-001" in codes
    assert "CODE-002" in codes
    assert "CODE-003" in codes


def test_error_registry_get_sod_score():
    """Test getting SOD score from registry"""
    registry = ErrorRegistry()

    registry.register(
        ErrorDefinition(
            code="SOD-001",
            severity=Severity.CRITICAL,
            description="High SOD score",
            severity_s=9,
            occurrence=8,
            detect_d=7,
            mitigation="Immediate action",
        )
    )

    score = registry.get_sod_score("SOD-001")
    assert score == 9 * 8 * 7  # 504

    # Nonexistent code
    assert registry.get_sod_score("NONEXISTENT") is None
