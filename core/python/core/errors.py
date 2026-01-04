"""
Error Registry - Python equivalent of Go core/errors

Provides:
- Structured error handling
- Error codes with severity
- SOD scoring (Severity × Occurrence × Detectability)
- Context attachment
"""

from enum import Enum
from typing import Any, Dict, Optional


class Severity(str, Enum):
    """Error severity levels - matches Go constants"""

    CRITICAL = "CRITICAL"  # System unavailable, data loss, security breach
    HIGH = "HIGH"  # Major functionality broken, significant impact
    MEDIUM = "MEDIUM"  # Moderate impact, workaround available
    LOW = "LOW"  # Minor issue, minimal impact
    INFO = "INFO"  # Informational, not an error


class ServiceError(Exception):
    """
    Structured error with code, severity, and context
    Matches Go ServiceError struct
    """

    def __init__(
        self,
        code: str,
        message: str,
        severity: Severity = Severity.MEDIUM,
        underlying: Optional[Exception] = None,
        context: Optional[Dict[str, Any]] = None,
    ):
        self.code = code
        self.message = message
        self.severity = severity
        self.underlying = underlying
        self.context = context or {}
        super().__init__(self._format_message())

    def _format_message(self) -> str:
        """Format error message consistently"""
        if self.underlying:
            return f"[{self.code}] {self.severity.value}: {self.message} (caused by: {self.underlying})"
        return f"[{self.code}] {self.severity.value}: {self.message}"

    def with_context(self, key: str, value: Any) -> "ServiceError":
        """Add context to error - matches Go WithContext"""
        self.context[key] = value
        return self

    def get_context(self, key: str) -> Optional[Any]:
        """Get context value - matches Go GetContext"""
        return self.context.get(key)

    def __str__(self) -> str:
        return self._format_message()

    def __repr__(self) -> str:
        return f"ServiceError(code='{self.code}', severity='{self.severity.value}', message='{self.message}')"


class ErrorDefinition:
    """
    Error definition with SOD scores
    Matches Go ErrorDefinition struct
    """

    def __init__(
        self,
        code: str,
        severity: Severity,
        description: str,
        severity_s: int,  # 1-10
        occurrence: int,  # 1-10
        detect_d: int,  # 1-10
        mitigation: str,
        example: str = "",
    ):
        self.code = code
        self.severity = severity
        self.description = description
        self.severity_s = severity_s
        self.occurrence = occurrence
        self.detect_d = detect_d
        self.sod_score = severity_s * occurrence * detect_d  # 1-1000
        self.mitigation = mitigation
        self.example = example

    def __repr__(self) -> str:
        return f"ErrorDefinition(code='{self.code}', sod_score={self.sod_score})"


class ErrorRegistry:
    """
    Registry for error definitions with SOD scoring
    Matches Go ErrorRegistry pattern
    """

    def __init__(self):
        self.definitions: Dict[str, ErrorDefinition] = {}

    def register(self, definition: ErrorDefinition) -> None:
        """Register an error definition"""
        self.definitions[definition.code] = definition

    def get(self, code: str) -> Optional[ErrorDefinition]:
        """Get error definition by code"""
        return self.definitions.get(code)

    def new_error(self, code: str, message: str) -> ServiceError:
        """
        Create a new ServiceError from registered definition
        Matches Go NewError pattern
        """
        definition = self.definitions.get(code)
        if definition:
            return ServiceError(
                code=code,
                message=message,
                severity=definition.severity,
            )
        # Fallback if code not registered
        return ServiceError(code=code, message=message, severity=Severity.MEDIUM)

    def wrap_error(self, err: Exception, code: str, message: str) -> ServiceError:
        """
        Wrap an existing error with ServiceError
        Matches Go WrapError pattern
        """
        definition = self.definitions.get(code)
        severity = definition.severity if definition else Severity.MEDIUM
        return ServiceError(
            code=code,
            message=message,
            severity=severity,
            underlying=err,
        )

    def list_codes(self) -> list[str]:
        """List all registered error codes"""
        return list(self.definitions.keys())

    def get_sod_score(self, code: str) -> Optional[int]:
        """Get SOD score for an error code"""
        definition = self.definitions.get(code)
        return definition.sod_score if definition else None
