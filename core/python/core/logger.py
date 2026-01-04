"""
Structured Logger - Python equivalent of Go core/logger

Provides:
- Structured JSON logging
- Correlation ID tracking
- Component-based logging
- Consistent with Go logger patterns
"""

import logging
import sys
from datetime import datetime, timezone
from typing import Any, Dict, Optional
import structlog
from pythonjsonlogger import jsonlogger


class ContextFilter(logging.Filter):
    """Filter to add context fields to log records"""

    def __init__(self, service_name: str, version: str, environment: str):
        super().__init__()
        self.service_name = service_name
        self.version = version
        self.environment = environment

    def filter(self, record: logging.LogRecord) -> bool:
        record.service = self.service_name
        record.version = self.version
        record.environment = self.environment
        return True


class Logger:
    """
    Structured logger matching Go core/logger patterns

    Features:
    - JSON structured logging
    - Correlation ID support
    - Component-based logging
    - ISO8601 timestamps
    """

    def __init__(
        self,
        service_name: str,
        version: str = "1.0.0",
        environment: str = "production",
        log_level: str = "INFO",
    ):
        self.service_name = service_name
        self.version = version
        self.environment = environment

        # Configure structured logging
        structlog.configure(
            processors=[
                structlog.contextvars.merge_contextvars,
                structlog.processors.add_log_level,
                structlog.processors.TimeStamper(fmt="iso", utc=True, key="timestamp"),
                structlog.processors.dict_tracebacks,
                structlog.processors.JSONRenderer(),
            ],
            wrapper_class=structlog.make_filtering_bound_logger(
                getattr(logging, log_level.upper())
            ),
            context_class=dict,
            logger_factory=structlog.PrintLoggerFactory(),
            cache_logger_on_first_use=True,
        )

        self.logger = structlog.get_logger()
        # Bind service metadata
        self.logger = self.logger.bind(
            service=service_name,
            version=version,
            environment=environment,
        )

    def with_correlation(
        self, correlation_id: str, component: str = ""
    ) -> "ContextLogger":
        """Create a context logger with correlation ID"""
        return ContextLogger(self.logger, correlation_id, component)

    def with_component(self, component: str) -> "ContextLogger":
        """Create a context logger for a specific component"""
        return ContextLogger(self.logger, "", component)

    def debug(self, message: str, **kwargs):
        """Log debug message"""
        self.logger.debug(message, **kwargs)

    def info(self, message: str, **kwargs):
        """Log info message"""
        self.logger.info(message, **kwargs)

    def warning(self, message: str, **kwargs):
        """Log warning message"""
        self.logger.warning(message, **kwargs)

    def error(self, message: str, **kwargs):
        """Log error message"""
        self.logger.error(message, **kwargs)

    def critical(self, message: str, **kwargs):
        """Log critical message"""
        self.logger.critical(message, **kwargs)


class ContextLogger:
    """
    Context-aware logger with correlation ID and component
    Matches Go ContextLogger pattern
    """

    def __init__(
        self, base_logger: structlog.BoundLogger, correlation_id: str, component: str
    ):
        self.correlation_id = correlation_id
        self.component = component

        # Bind context to logger
        bind_data = {}
        if correlation_id:
            bind_data["correlation_id"] = correlation_id
        if component:
            bind_data["component"] = component

        self.logger = base_logger.bind(**bind_data) if bind_data else base_logger

    def debug(self, message: str, **kwargs):
        """Log debug message with context"""
        self.logger.debug(message, **kwargs)

    def info(self, message: str, **kwargs):
        """Log info message with context"""
        self.logger.info(message, **kwargs)

    def warning(self, message: str, **kwargs):
        """Log warning message with context"""
        self.logger.warning(message, **kwargs)

    def error(self, message: str, **kwargs):
        """Log error message with context"""
        self.logger.error(message, **kwargs)

    def critical(self, message: str, **kwargs):
        """Log critical message with context"""
        self.logger.critical(message, **kwargs)

    def with_component(self, component: str) -> "ContextLogger":
        """Create a new logger with different component"""
        return ContextLogger(self.logger, self.correlation_id, component)


def setup_logger(
    service_name: str,
    version: str = "1.0.0",
    environment: str = "production",
    log_level: str = "INFO",
) -> Logger:
    """
    Factory function to create a configured logger
    Matches Go logger.New() pattern
    """
    return Logger(service_name, version, environment, log_level)


# Module-level logger cache for get_logger
_loggers: Dict[str, Logger] = {}
_default_service_name: str = "iot-service"
_default_environment: str = "development"
_default_log_level: str = "INFO"


def configure_logging(
    service_name: str,
    environment: str = "development",
    log_level: str = "INFO",
) -> None:
    """
    Configure default logging settings for get_logger.
    Call this once at application startup.
    """
    global _default_service_name, _default_environment, _default_log_level
    _default_service_name = service_name
    _default_environment = environment
    _default_log_level = log_level


def get_logger(name: str = __name__) -> Logger:
    """
    Get or create a logger instance by name.

    This provides a simple way to get loggers without passing
    configuration everywhere. Call configure_logging() at startup
    to set defaults.

    Args:
        name: Logger name (typically __name__ of the calling module)

    Returns:
        Configured Logger instance
    """
    if name not in _loggers:
        _loggers[name] = Logger(
            service_name=_default_service_name,
            version="1.0.0",
            environment=_default_environment,
            log_level=_default_log_level,
        )
    return _loggers[name]
