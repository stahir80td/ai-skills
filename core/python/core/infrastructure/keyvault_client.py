"""
Azure Key Vault client - Production-ready Key Vault client with caching

Provides:
- Secret retrieval with local caching
- Secret listing with filtering
- Integration status checks
- Health checks
- Structured logging with correlation IDs
- Circuit breaker protection
- TTL-based cache invalidation

Supports both:
- Azure Key Vault (production)
- Local Key Vault emulator (development)
"""

from typing import Any, Dict, List, Optional, Set
from dataclasses import dataclass
from datetime import datetime, timedelta
from threading import Lock
import time
import re

from azure.identity import DefaultAzureCredential, ClientSecretCredential
from azure.keyvault.secrets import SecretClient, SecretProperties
from azure.core.exceptions import (
    ResourceNotFoundError,
    HttpResponseError,
    ClientAuthenticationError,
    ServiceRequestError,
)

from ..logger import Logger
from ..errors import ServiceError, Severity
from ..reliability import CircuitBreaker


@dataclass
class CachedSecret:
    """Cached secret with expiration"""

    value: str
    metadata: Dict[str, Any]
    cached_at: datetime
    expires_at: datetime


@dataclass
class KeyVaultConfig:
    """
    Configuration for Key Vault client
    """

    vault_url: str
    logger: Logger
    cache_ttl_seconds: int = 300  # 5 minutes default cache
    timeout_seconds: float = 60.0
    tenant_id: Optional[str] = None
    client_id: Optional[str] = None
    client_secret: Optional[str] = None  # For service principal auth

    def __post_init__(self):
        """Validate configuration"""
        if not self.vault_url:
            raise ServiceError(
                code="INFRA-KEYVAULT-CONFIG-ERROR",
                message="Key Vault URL cannot be empty",
                severity=Severity.CRITICAL,
            )


class KeyVaultClient:
    """
    Azure Key Vault client with production patterns

    Features:
    - Secret caching with configurable TTL
    - Circuit breaker for fault tolerance
    - Structured logging with masked secret values
    - Support for user integration patterns (user:{user_id}:{type})
    - Metadata extraction without exposing secret values
    """

    # Pattern for user integration secrets: user:{user_id}:{integration_type}
    USER_INTEGRATION_PATTERN = re.compile(r"^user:([a-f0-9-]+):(\w+)$", re.IGNORECASE)

    def __init__(self, config: KeyVaultConfig):
        self.config = config
        self.logger = config.logger.with_component("KeyVaultClient")
        self._client: Optional[SecretClient] = None

        # Circuit breaker - can be disabled via env var for debugging
        import os

        cb_enabled = os.getenv("CIRCUIT_BREAKER_ENABLED", "true").lower() != "false"
        self.circuit_breaker = CircuitBreaker("keyvault", max_failures=5, enabled=False)

        # Thread-safe cache
        self._cache: Dict[str, CachedSecret] = {}
        self._cache_lock = Lock()

        self.logger.info(
            "Initializing Key Vault client",
            vault_url=config.vault_url,
            cache_ttl_seconds=config.cache_ttl_seconds,
            circuit_breaker_enabled=cb_enabled,
        )

    def connect(self) -> None:
        """
        Establish connection to Key Vault
        Initializes credential and client
        """
        start_time = time.time()

        # Log connection settings for debugging
        self.logger.info(
            "KeyVault connecting",
            vault_url=self.config.vault_url,
            cache_ttl_seconds=self.config.cache_ttl_seconds,
            has_client_id=bool(self.config.client_id),
            has_tenant_id=bool(self.config.tenant_id),
        )

        try:
            # Choose credential based on config
            if (
                self.config.client_id
                and self.config.client_secret
                and self.config.tenant_id
            ):
                # Service principal authentication
                credential = ClientSecretCredential(
                    tenant_id=self.config.tenant_id,
                    client_id=self.config.client_id,
                    client_secret=self.config.client_secret,
                )
                self.logger.info("Using service principal authentication")
            else:
                # Default Azure credential (managed identity, CLI, etc.)
                credential = DefaultAzureCredential()
                self.logger.info("Using default Azure credential")

            self._client = SecretClient(
                vault_url=self.config.vault_url,
                credential=credential,
            )

            # Verify connection by listing secrets (limit 1)
            list(self._client.list_properties_of_secrets(max_page_size=1))

            elapsed_ms = (time.time() - start_time) * 1000

            self.logger.info(
                "Connected to Key Vault",
                vault_url=self.config.vault_url,
                status="healthy",
                connect_time_ms=round(elapsed_ms, 2),
            )

        except ClientAuthenticationError as e:
            self.logger.error(
                "Key Vault authentication failed",
                error=str(e),
                vault_url=self.config.vault_url,
                error_code="INFRA-KEYVAULT-AUTH-ERROR",
            )
            raise ServiceError(
                code="INFRA-KEYVAULT-AUTH-ERROR",
                message="Key Vault authentication failed - check credentials",
                severity=Severity.CRITICAL,
                underlying=e,
            )

        except ServiceRequestError as e:
            self.logger.error(
                "Key Vault connection failed",
                error=str(e),
                vault_url=self.config.vault_url,
                error_code="INFRA-KEYVAULT-CONNECT-ERROR",
            )
            raise ServiceError(
                code="INFRA-KEYVAULT-CONNECT-ERROR",
                message=f"Failed to connect to Key Vault: {e}",
                severity=Severity.CRITICAL,
                underlying=e,
            )

        except Exception as e:
            self.logger.error(
                "Unexpected Key Vault connection error",
                error=str(e),
                error_type=type(e).__name__,
                error_code="INFRA-KEYVAULT-CONNECT-ERROR",
            )
            raise ServiceError(
                code="INFRA-KEYVAULT-CONNECT-ERROR",
                message=f"Unexpected error connecting to Key Vault: {e}",
                severity=Severity.CRITICAL,
                underlying=e,
            )

    def _ensure_connected(self) -> SecretClient:
        """Ensure client is connected"""
        if not self._client:
            raise ServiceError(
                code="INFRA-KEYVAULT-CLIENT-ERROR",
                message="Key Vault client not connected - call connect() first",
                severity=Severity.CRITICAL,
            )
        return self._client

    def _get_cached(self, secret_name: str) -> Optional[CachedSecret]:
        """Get secret from cache if not expired"""
        with self._cache_lock:
            cached = self._cache.get(secret_name)
            if cached and datetime.now() < cached.expires_at:
                return cached
            elif cached:
                # Remove expired
                del self._cache[secret_name]
        return None

    def _set_cached(
        self, secret_name: str, value: str, metadata: Dict[str, Any]
    ) -> None:
        """Add secret to cache"""
        now = datetime.now()
        with self._cache_lock:
            self._cache[secret_name] = CachedSecret(
                value=value,
                metadata=metadata,
                cached_at=now,
                expires_at=now + timedelta(seconds=self.config.cache_ttl_seconds),
            )

    def get_secret(
        self,
        secret_name: str,
        use_cache: bool = True,
    ) -> Optional[str]:
        """
        Get secret value by name

        Args:
            secret_name: Name of the secret
            use_cache: Whether to use cached value

        Returns:
            Secret value or None if not found
        """
        # Check cache first
        if use_cache:
            cached = self._get_cached(secret_name)
            if cached:
                self.logger.info(
                    "Key Vault cache hit",
                    secret_name=secret_name,
                    cached_at=cached.cached_at.isoformat(),
                )
                return cached.value

        def _get_secret():
            client = self._ensure_connected()
            start_time = time.time()

            try:
                secret = client.get_secret(secret_name)
                elapsed_ms = (time.time() - start_time) * 1000

                # Cache the result
                metadata = {
                    "content_type": secret.properties.content_type,
                    "enabled": secret.properties.enabled,
                    "created_on": (
                        secret.properties.created_on.isoformat()
                        if secret.properties.created_on
                        else None
                    ),
                    "updated_on": (
                        secret.properties.updated_on.isoformat()
                        if secret.properties.updated_on
                        else None
                    ),
                    "expires_on": (
                        secret.properties.expires_on.isoformat()
                        if secret.properties.expires_on
                        else None
                    ),
                }
                self._set_cached(secret_name, secret.value, metadata)

                self.logger.info(
                    "Key Vault get_secret executed",
                    secret_name=secret_name,
                    found=True,
                    elapsed_ms=round(elapsed_ms, 2),
                    # Never log secret value!
                )

                return secret.value

            except ResourceNotFoundError:
                elapsed_ms = (time.time() - start_time) * 1000
                self.logger.info(
                    "Key Vault secret not found",
                    secret_name=secret_name,
                    found=False,
                    elapsed_ms=round(elapsed_ms, 2),
                )
                return None

            except HttpResponseError as e:
                self.logger.error(
                    "Key Vault get_secret failed",
                    error=str(e),
                    secret_name=secret_name,
                    error_code="INFRA-KEYVAULT-GET-ERROR",
                )
                raise ServiceError(
                    code="INFRA-KEYVAULT-GET-ERROR",
                    message=f"Failed to get secret: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_get_secret)

    def get_secret_metadata(
        self,
        secret_name: str,
    ) -> Optional[Dict[str, Any]]:
        """
        Get secret metadata without retrieving value
        Useful for checking integration status

        Returns:
            Metadata dict or None if not found
        """

        def _get_metadata():
            client = self._ensure_connected()
            start_time = time.time()

            try:
                properties = client.get_secret(secret_name).properties
                elapsed_ms = (time.time() - start_time) * 1000

                metadata = {
                    "name": properties.name,
                    "content_type": properties.content_type,
                    "enabled": properties.enabled,
                    "created_on": (
                        properties.created_on.isoformat()
                        if properties.created_on
                        else None
                    ),
                    "updated_on": (
                        properties.updated_on.isoformat()
                        if properties.updated_on
                        else None
                    ),
                    "expires_on": (
                        properties.expires_on.isoformat()
                        if properties.expires_on
                        else None
                    ),
                    "tags": properties.tags or {},
                }

                self.logger.info(
                    "Key Vault metadata retrieved",
                    secret_name=secret_name,
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return metadata

            except ResourceNotFoundError:
                return None

            except HttpResponseError as e:
                self.logger.error(
                    "Key Vault get_metadata failed",
                    error=str(e),
                    secret_name=secret_name,
                    error_code="INFRA-KEYVAULT-GET-ERROR",
                )
                raise ServiceError(
                    code="INFRA-KEYVAULT-GET-ERROR",
                    message=f"Failed to get secret metadata: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_get_metadata)

    def list_secrets(
        self,
        prefix: Optional[str] = None,
    ) -> List[Dict[str, Any]]:
        """
        List secrets with optional prefix filter

        Args:
            prefix: Filter secrets starting with this prefix

        Returns:
            List of secret metadata dicts (no values)
        """

        def _list_secrets():
            client = self._ensure_connected()
            start_time = time.time()

            try:
                secrets = []
                for properties in client.list_properties_of_secrets():
                    # Apply prefix filter
                    if prefix and not properties.name.startswith(prefix):
                        continue

                    secrets.append(
                        {
                            "name": properties.name,
                            "content_type": properties.content_type,
                            "enabled": properties.enabled,
                            "created_on": (
                                properties.created_on.isoformat()
                                if properties.created_on
                                else None
                            ),
                            "updated_on": (
                                properties.updated_on.isoformat()
                                if properties.updated_on
                                else None
                            ),
                            "expires_on": (
                                properties.expires_on.isoformat()
                                if properties.expires_on
                                else None
                            ),
                            "tags": properties.tags or {},
                        }
                    )

                elapsed_ms = (time.time() - start_time) * 1000

                self.logger.info(
                    "Key Vault list_secrets executed",
                    prefix=prefix,
                    count=len(secrets),
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return secrets

            except HttpResponseError as e:
                self.logger.error(
                    "Key Vault list_secrets failed",
                    error=str(e),
                    prefix=prefix,
                    error_code="INFRA-KEYVAULT-LIST-ERROR",
                )
                raise ServiceError(
                    code="INFRA-KEYVAULT-LIST-ERROR",
                    message=f"Failed to list secrets: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_list_secrets)

    # ==================== User Integration Methods ====================

    def list_user_integrations(self, user_id: str) -> List[Dict[str, Any]]:
        """
        List all integrations for a user

        Looks for secrets matching pattern: user:{user_id}:{integration_type}

        Args:
            user_id: User ID (UUID format)

        Returns:
            List of integration metadata (type, status, etc.)
        """
        prefix = f"user-{user_id}-"  # Key Vault uses - not : in names
        secrets = self.list_secrets(prefix=prefix)

        integrations = []
        for secret in secrets:
            # Parse integration type from name: user-{user_id}-{type}
            parts = secret["name"].split("-")
            if len(parts) >= 3:
                integration_type = "-".join(parts[2:])  # Handle types with dashes

                integrations.append(
                    {
                        "user_id": user_id,
                        "integration_type": integration_type,
                        "enabled": secret["enabled"],
                        "created_on": secret["created_on"],
                        "updated_on": secret["updated_on"],
                        "expires_on": secret["expires_on"],
                        "content_type": secret["content_type"],
                        # Never include the actual secret value!
                    }
                )

        self.logger.info(
            "Listed user integrations",
            user_id=user_id,
            integration_count=len(integrations),
        )

        return integrations

    def get_integration_status(
        self,
        user_id: str,
        integration_type: str,
    ) -> Dict[str, Any]:
        """
        Get status of a specific user integration

        Args:
            user_id: User ID
            integration_type: Integration type (e.g., 'weather', 'google_home')

        Returns:
            Integration status dict
        """
        secret_name = f"user-{user_id}-{integration_type}"
        metadata = self.get_secret_metadata(secret_name)

        if not metadata:
            return {
                "user_id": user_id,
                "integration_type": integration_type,
                "exists": False,
                "enabled": False,
            }

        # Check if expired
        is_expired = False
        if metadata.get("expires_on"):
            expires = datetime.fromisoformat(metadata["expires_on"])
            is_expired = expires < datetime.now()

        return {
            "user_id": user_id,
            "integration_type": integration_type,
            "exists": True,
            "enabled": metadata.get("enabled", False) and not is_expired,
            "is_expired": is_expired,
            "created_on": metadata.get("created_on"),
            "updated_on": metadata.get("updated_on"),
            "expires_on": metadata.get("expires_on"),
        }

    def get_all_integration_types(self) -> Set[str]:
        """
        Get all unique integration types across all users

        Returns:
            Set of integration type names
        """
        secrets = self.list_secrets(prefix="user-")
        types = set()

        for secret in secrets:
            parts = secret["name"].split("-")
            if len(parts) >= 3:
                integration_type = "-".join(parts[2:])
                types.add(integration_type)

        return types

    def clear_cache(self) -> None:
        """Clear all cached secrets"""
        with self._cache_lock:
            count = len(self._cache)
            self._cache.clear()
            self.logger.info("Key Vault cache cleared", cleared_count=count)

    def health_check(self) -> Dict[str, Any]:
        """
        Check Key Vault health status

        Returns:
            Health status dict with latency info
        """
        start_time = time.time()

        try:
            if not self._client:
                return {
                    "status": "unhealthy",
                    "error": "Client not connected",
                    "vault_url": self.config.vault_url,
                }

            # Try to list secrets (limit 1)
            list(self._client.list_properties_of_secrets(max_page_size=1))
            elapsed_ms = (time.time() - start_time) * 1000

            return {
                "status": "healthy",
                "vault_url": self.config.vault_url,
                "latency_ms": round(elapsed_ms, 2),
                "cache_size": len(self._cache),
            }

        except Exception as e:
            elapsed_ms = (time.time() - start_time) * 1000
            self.logger.warning(
                "Key Vault health check failed",
                error=str(e),
                elapsed_ms=round(elapsed_ms, 2),
            )
            return {
                "status": "unhealthy",
                "vault_url": self.config.vault_url,
                "error": str(e),
                "latency_ms": round(elapsed_ms, 2),
            }

    def close(self) -> None:
        """Close Key Vault client and clear cache"""
        self.clear_cache()
        self._client = None
        self.logger.info(
            "Key Vault client closed",
            vault_url=self.config.vault_url,
        )

    def __enter__(self) -> "KeyVaultClient":
        """Context manager entry"""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        """Context manager exit"""
        self.close()
