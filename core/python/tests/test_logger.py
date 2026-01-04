"""Tests for logger module"""

import pytest
from core.logger import Logger, ContextLogger, setup_logger


def test_logger_creation():
    """Test logger can be created"""
    logger = setup_logger("test-service", version="1.0.0", environment="test")
    assert logger is not None
    assert logger.service_name == "test-service"
    assert logger.version == "1.0.0"
    assert logger.environment == "test"


def test_logger_with_correlation():
    """Test creating context logger with correlation ID"""
    logger = setup_logger("test-service")
    ctx_logger = logger.with_correlation("test-corr-123", "TestComponent")

    assert isinstance(ctx_logger, ContextLogger)
    assert ctx_logger.correlation_id == "test-corr-123"
    assert ctx_logger.component == "TestComponent"


def test_logger_with_component():
    """Test creating context logger with component only"""
    logger = setup_logger("test-service")
    ctx_logger = logger.with_component("DataProcessor")

    assert isinstance(ctx_logger, ContextLogger)
    assert ctx_logger.component == "DataProcessor"


def test_logger_methods():
    """Test all logger methods execute without error"""
    logger = setup_logger("test-service", log_level="DEBUG")

    # Should not raise exceptions
    logger.debug("Debug message", test_key="test_value")
    logger.info("Info message", count=42)
    logger.warning("Warning message", status="degraded")
    logger.error("Error message", error_code="TEST-001")


def test_context_logger_methods():
    """Test context logger methods"""
    logger = setup_logger("test-service")
    ctx_logger = logger.with_correlation("corr-456", "TestComp")

    # Should not raise exceptions
    ctx_logger.debug("Debug with context")
    ctx_logger.info("Info with context")
    ctx_logger.warning("Warning with context")
    ctx_logger.error("Error with context")


def test_context_logger_chaining():
    """Test creating new context logger from existing one"""
    logger = setup_logger("test-service")
    ctx_logger1 = logger.with_correlation("corr-789", "Component1")
    ctx_logger2 = ctx_logger1.with_component("Component2")

    assert ctx_logger2.correlation_id == "corr-789"
    assert ctx_logger2.component == "Component2"
