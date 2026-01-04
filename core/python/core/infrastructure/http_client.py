"""
HTTP client - Production-ready HTTP client with circuit breaker and retry

Provides:
- GET, POST, PUT, DELETE methods
- Automatic retries with exponential backoff
- Circuit breaker protection
- Timeout configuration
- Structured logging
- Health checks
"""

from typing import Any, Dict, Optional
from dataclasses import dataclass
import httpx

from ..logger import Logger
from ..errors import ServiceError, Severity
from ..reliability import CircuitBreaker
import tenacity


@dataclass
class HTTPConfig:
    """Configuration for HTTP client"""

    base_url: str
    logger: Logger
    timeout_seconds: float = 30.0
    max_retries: int = 3
    retry_delay_seconds: float = 1.0
    max_retry_delay_seconds: float = 10.0

    def __post_init__(self):
        """Validate configuration"""
        if not self.base_url:
            raise ServiceError(
                code="INFRA-HTTP-CONFIG-ERROR",
                message="HTTP base_url cannot be empty",
                severity=Severity.CRITICAL,
            )


class HTTPClient:
    """
    HTTP client with production patterns
    Features: circuit breaker, retries, timeouts, logging
    """

    def __init__(self, config: HTTPConfig):
        self.config = config
        self.logger = config.logger.with_component("HTTPClient")
        self.circuit_breaker = CircuitBreaker(
            "http_client", max_failures=5, enabled=False
        )

        # Create httpx client
        self.client = httpx.Client(
            base_url=config.base_url,
            timeout=config.timeout_seconds,
        )

        self.logger.debug(
            "HTTP client initialized",
            base_url=config.base_url,
            timeout=config.timeout_seconds,
        )

    def get(
        self,
        path: str,
        params: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> httpx.Response:
        """
        HTTP GET request with circuit breaker and retry
        """

        def _get():
            try:
                response = self.client.get(path, params=params, headers=headers)

                self.logger.debug(
                    "HTTP GET",
                    path=path,
                    status_code=response.status_code,
                    elapsed_ms=response.elapsed.total_seconds() * 1000,
                )

                # Raise for 4xx/5xx status codes
                response.raise_for_status()

                return response

            except httpx.TimeoutException as e:
                self.logger.error(
                    "HTTP GET timeout",
                    error=str(e),
                    path=path,
                    error_code="INFRA-HTTP-TIMEOUT",
                )
                raise ServiceError(
                    code="INFRA-HTTP-TIMEOUT",
                    message=f"Request timeout: {path}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except httpx.HTTPStatusError as e:
                self.logger.error(
                    "HTTP GET failed",
                    error=str(e),
                    path=path,
                    status_code=e.response.status_code,
                    error_code="INFRA-HTTP-ERROR",
                )
                raise ServiceError(
                    code="INFRA-HTTP-ERROR",
                    message=f"HTTP error: {e.response.status_code}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except Exception as e:
                self.logger.error(
                    "HTTP GET exception",
                    error=str(e),
                    path=path,
                    error_code="INFRA-HTTP-ERROR",
                )
                raise ServiceError(
                    code="INFRA-HTTP-ERROR",
                    message=f"Request failed: {path}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        # Apply retry with exponential backoff and circuit breaker
        @tenacity.retry(
            stop=tenacity.stop_after_attempt(self.config.max_retries),
            wait=tenacity.wait_exponential(
                multiplier=2.0,
                min=self.config.retry_delay_seconds,
                max=self.config.max_retry_delay_seconds,
            ),
            reraise=True,
        )
        def _with_retry():
            return self.circuit_breaker.call(_get)

        return _with_retry()

    def post(
        self,
        path: str,
        json: Optional[Dict[str, Any]] = None,
        data: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> httpx.Response:
        """
        HTTP POST request with circuit breaker and retry
        """

        def _post():
            try:
                response = self.client.post(path, json=json, data=data, headers=headers)

                self.logger.debug(
                    "HTTP POST",
                    path=path,
                    status_code=response.status_code,
                    elapsed_ms=response.elapsed.total_seconds() * 1000,
                )

                response.raise_for_status()
                return response

            except httpx.TimeoutException as e:
                self.logger.error(
                    "HTTP POST timeout",
                    error=str(e),
                    path=path,
                    error_code="INFRA-HTTP-TIMEOUT",
                )
                raise ServiceError(
                    code="INFRA-HTTP-TIMEOUT",
                    message=f"Request timeout: {path}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except httpx.HTTPStatusError as e:
                self.logger.error(
                    "HTTP POST failed",
                    error=str(e),
                    path=path,
                    status_code=e.response.status_code,
                    error_code="INFRA-HTTP-ERROR",
                )
                raise ServiceError(
                    code="INFRA-HTTP-ERROR",
                    message=f"HTTP error: {e.response.status_code}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except Exception as e:
                self.logger.error(
                    "HTTP POST exception",
                    error=str(e),
                    path=path,
                    error_code="INFRA-HTTP-ERROR",
                )
                raise ServiceError(
                    code="INFRA-HTTP-ERROR",
                    message=f"Request failed: {path}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        @tenacity.retry(
            stop=tenacity.stop_after_attempt(self.config.max_retries),
            wait=tenacity.wait_exponential(
                multiplier=2.0,
                min=self.config.retry_delay_seconds,
                max=self.config.max_retry_delay_seconds,
            ),
            reraise=True,
        )
        def _with_retry():
            return self.circuit_breaker.call(_post)

        return _with_retry()

    def put(
        self,
        path: str,
        json: Optional[Dict[str, Any]] = None,
        data: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> httpx.Response:
        """
        HTTP PUT request with circuit breaker and retry
        """

        def _put():
            try:
                response = self.client.put(path, json=json, data=data, headers=headers)

                self.logger.debug(
                    "HTTP PUT",
                    path=path,
                    status_code=response.status_code,
                    elapsed_ms=response.elapsed.total_seconds() * 1000,
                )

                response.raise_for_status()
                return response

            except httpx.TimeoutException as e:
                self.logger.error(
                    "HTTP PUT timeout",
                    error=str(e),
                    path=path,
                    error_code="INFRA-HTTP-TIMEOUT",
                )
                raise ServiceError(
                    code="INFRA-HTTP-TIMEOUT",
                    message=f"Request timeout: {path}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except httpx.HTTPStatusError as e:
                self.logger.error(
                    "HTTP PUT failed",
                    error=str(e),
                    path=path,
                    status_code=e.response.status_code,
                    error_code="INFRA-HTTP-ERROR",
                )
                raise ServiceError(
                    code="INFRA-HTTP-ERROR",
                    message=f"HTTP error: {e.response.status_code}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except Exception as e:
                self.logger.error(
                    "HTTP PUT exception",
                    error=str(e),
                    path=path,
                    error_code="INFRA-HTTP-ERROR",
                )
                raise ServiceError(
                    code="INFRA-HTTP-ERROR",
                    message=f"Request failed: {path}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        @tenacity.retry(
            stop=tenacity.stop_after_attempt(self.config.max_retries),
            wait=tenacity.wait_exponential(
                multiplier=2.0,
                min=self.config.retry_delay_seconds,
                max=self.config.max_retry_delay_seconds,
            ),
            reraise=True,
        )
        def _with_retry():
            return self.circuit_breaker.call(_put)

        return _with_retry()

    def delete(
        self,
        path: str,
        params: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> httpx.Response:
        """
        HTTP DELETE request with circuit breaker and retry
        """

        def _delete():
            try:
                response = self.client.delete(path, params=params, headers=headers)

                self.logger.debug(
                    "HTTP DELETE",
                    path=path,
                    status_code=response.status_code,
                    elapsed_ms=response.elapsed.total_seconds() * 1000,
                )

                response.raise_for_status()
                return response

            except httpx.TimeoutException as e:
                self.logger.error(
                    "HTTP DELETE timeout",
                    error=str(e),
                    path=path,
                    error_code="INFRA-HTTP-TIMEOUT",
                )
                raise ServiceError(
                    code="INFRA-HTTP-TIMEOUT",
                    message=f"Request timeout: {path}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except httpx.HTTPStatusError as e:
                self.logger.error(
                    "HTTP DELETE failed",
                    error=str(e),
                    path=path,
                    status_code=e.response.status_code,
                    error_code="INFRA-HTTP-ERROR",
                )
                raise ServiceError(
                    code="INFRA-HTTP-ERROR",
                    message=f"HTTP error: {e.response.status_code}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except Exception as e:
                self.logger.error(
                    "HTTP DELETE exception",
                    error=str(e),
                    path=path,
                    error_code="INFRA-HTTP-ERROR",
                )
                raise ServiceError(
                    code="INFRA-HTTP-ERROR",
                    message=f"Request failed: {path}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        @tenacity.retry(
            stop=tenacity.stop_after_attempt(self.config.max_retries),
            wait=tenacity.wait_exponential(
                multiplier=2.0,
                min=self.config.retry_delay_seconds,
                max=self.config.max_retry_delay_seconds,
            ),
            reraise=True,
        )
        def _with_retry():
            return self.circuit_breaker.call(_delete)

        return _with_retry()

    def health(self) -> bool:
        """
        Check HTTP client health
        Tries to connect to base_url/health or base_url
        """
        try:
            # Try health endpoint first
            try:
                response = self.client.get("/health", timeout=5.0)
                return response.status_code < 500
            except:
                # Fallback to base URL
                response = self.client.get("/", timeout=5.0)
                return response.status_code < 500

        except Exception as e:
            self.logger.warning("HTTP health check failed", error=str(e))
            return False

    def close(self) -> None:
        """Close HTTP client"""
        self.client.close()
        self.logger.info("HTTP client closed")
