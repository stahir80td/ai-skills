"""
Redis client - Python equivalent of Go core/infrastructure/redis

Provides:
- Get/Set operations with type safety
- Set operations (SAdd, SRem, SMembers)
- Expiration handling
- Health checks
- Structured logging
- Circuit breaker protection
"""

from typing import Any, List, Optional
from dataclasses import dataclass
import json
import redis
from redis.exceptions import RedisError, ConnectionError, TimeoutError

from ..logger import Logger
from ..errors import ServiceError, Severity
from ..reliability import CircuitBreaker


@dataclass
class RedisConfig:
    """
    Configuration for Redis client
    Matches Go ClientConfig pattern
    """

    host: str
    port: int
    logger: Logger
    ping_timeout_seconds: float = 60.0  # Connection test timeout
    db: int = 0  # Redis database number
    password: Optional[str] = None

    def __post_init__(self):
        """Validate configuration"""
        if not self.host:
            raise ServiceError(
                code="INFRA-REDIS-CONFIG-ERROR",
                message="Redis host cannot be empty",
                severity=Severity.CRITICAL,
            )
        if self.port <= 0 or self.port > 65535:
            raise ServiceError(
                code="INFRA-REDIS-CONFIG-ERROR",
                message=f"Redis port must be between 1 and 65535, got {self.port}",
                severity=Severity.CRITICAL,
            )


class RedisClient:
    """
    Redis client with production patterns
    Matches Go Client interface
    """

    def __init__(self, config: RedisConfig):
        self.config = config
        self.logger = config.logger.with_component("RedisClient")
        self.client: Optional[redis.Redis] = None

        # Circuit breaker - can be disabled via env var for debugging
        import os

        cb_enabled = os.getenv("CIRCUIT_BREAKER_ENABLED", "true").lower() != "false"
        self.circuit_breaker = CircuitBreaker("redis", max_failures=5, enabled=False)

        self.logger.info(
            "Initiating Redis connection",
            host=config.host,
            port=config.port,
            db=config.db,
            circuit_breaker_enabled=cb_enabled,
        )

    def connect(self) -> None:
        """Establish connection to Redis"""
        # Log connection settings for debugging
        self.logger.info(
            "Redis connecting",
            host=self.config.host,
            port=self.config.port,
            db=self.config.db,
            ping_timeout_seconds=self.config.ping_timeout_seconds,
        )

        try:
            # Create Redis client with connection pool
            self.client = redis.Redis(
                host=self.config.host,
                port=self.config.port,
                db=self.config.db,
                password=self.config.password,
                decode_responses=True,  # Automatically decode bytes to strings
                socket_timeout=self.config.ping_timeout_seconds,
                socket_connect_timeout=self.config.ping_timeout_seconds,
                max_connections=10,
            )

            # Test connection
            self.client.ping()

            self.logger.info(
                "Successfully connected to Redis",
                host=self.config.host,
                port=self.config.port,
                db=self.config.db,
                status="healthy",
                pool_size=10,
            )

        except (ConnectionError, TimeoutError) as e:
            self.logger.error(
                "Redis connection failed",
                error=str(e),
                host=self.config.host,
                port=self.config.port,
                error_code="INFRA-REDIS-CONNECT-ERROR",
            )
            raise ServiceError(
                code="INFRA-REDIS-CONNECT-ERROR",
                message="Failed to connect to Redis",
                severity=Severity.CRITICAL,
                underlying=e,
            )

    def get(self, key: str) -> Optional[str]:
        """
        Get value by key
        Matches Go Get method
        """

        def _get():
            if not self.client:
                raise ServiceError(
                    code="INFRA-REDIS-CLIENT-ERROR",
                    message="Redis client not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                value = self.client.get(key)
                self.logger.info("Redis GET", key=key, found=value is not None)
                return value

            except RedisError as e:
                self.logger.error(
                    "Redis GET failed",
                    error=str(e),
                    key=key,
                    error_code="INFRA-REDIS-GET-ERROR",
                )
                raise ServiceError(
                    code="INFRA-REDIS-GET-ERROR",
                    message=f"Failed to get key: {key}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_get)

    def set(
        self, key: str, value: Any, expiration_seconds: Optional[int] = None
    ) -> bool:
        """
        Set key-value pair with optional expiration
        Matches Go Set method
        """

        def _set():
            if not self.client:
                raise ServiceError(
                    code="INFRA-REDIS-CLIENT-ERROR",
                    message="Redis client not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                # Convert value to string if needed
                if isinstance(value, (dict, list)):
                    value_str = json.dumps(value)
                elif isinstance(value, (int, float, bool)):
                    value_str = str(value)
                else:
                    value_str = value

                if expiration_seconds:
                    result = self.client.setex(key, expiration_seconds, value_str)
                else:
                    result = self.client.set(key, value_str)

                self.logger.info(
                    "Redis SET",
                    key=key,
                    expiration=expiration_seconds,
                    success=result,
                )
                return result

            except RedisError as e:
                self.logger.error(
                    "Redis SET failed",
                    error=str(e),
                    key=key,
                    error_code="INFRA-REDIS-SET-ERROR",
                )
                raise ServiceError(
                    code="INFRA-REDIS-SET-ERROR",
                    message=f"Failed to set key: {key}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_set)

    def delete(self, *keys: str) -> int:
        """
        Delete one or more keys
        Matches Go Del method
        """

        def _delete():
            if not self.client:
                raise ServiceError(
                    code="INFRA-REDIS-CLIENT-ERROR",
                    message="Redis client not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                count = self.client.delete(*keys)
                self.logger.info("Redis DELETE", keys=keys, deleted_count=count)
                return count

            except RedisError as e:
                self.logger.error(
                    "Redis DELETE failed",
                    error=str(e),
                    keys=keys,
                    error_code="INFRA-REDIS-DEL-ERROR",
                )
                raise ServiceError(
                    code="INFRA-REDIS-DEL-ERROR",
                    message=f"Failed to delete keys: {keys}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_delete)

    def smembers(self, key: str) -> List[str]:
        """
        Get all members of a set
        Matches Go SMembers method
        """

        def _smembers():
            if not self.client:
                raise ServiceError(
                    code="INFRA-REDIS-CLIENT-ERROR",
                    message="Redis client not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                members = self.client.smembers(key)
                result = list(members) if members else []
                self.logger.info("Redis SMEMBERS", key=key, count=len(result))
                return result

            except RedisError as e:
                self.logger.error(
                    "Redis SMEMBERS failed",
                    error=str(e),
                    key=key,
                    error_code="INFRA-REDIS-SMEMBERS-ERROR",
                )
                raise ServiceError(
                    code="INFRA-REDIS-SMEMBERS-ERROR",
                    message=f"Failed to get set members: {key}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_smembers)

    def sadd(self, key: str, *members: Any) -> int:
        """
        Add members to a set
        Matches Go SAdd method
        """

        def _sadd():
            if not self.client:
                raise ServiceError(
                    code="INFRA-REDIS-CLIENT-ERROR",
                    message="Redis client not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                count = self.client.sadd(key, *members)
                self.logger.info("Redis SADD", key=key, added_count=count)
                return count

            except RedisError as e:
                self.logger.error(
                    "Redis SADD failed",
                    error=str(e),
                    key=key,
                    error_code="INFRA-REDIS-SADD-ERROR",
                )
                raise ServiceError(
                    code="INFRA-REDIS-SADD-ERROR",
                    message=f"Failed to add to set: {key}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_sadd)

    def srem(self, key: str, *members: Any) -> int:
        """
        Remove members from a set
        Matches Go SRem method
        """

        def _srem():
            if not self.client:
                raise ServiceError(
                    code="INFRA-REDIS-CLIENT-ERROR",
                    message="Redis client not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                count = self.client.srem(key, *members)
                self.logger.info("Redis SREM", key=key, removed_count=count)
                return count

            except RedisError as e:
                self.logger.error(
                    "Redis SREM failed",
                    error=str(e),
                    key=key,
                    error_code="INFRA-REDIS-SREM-ERROR",
                )
                raise ServiceError(
                    code="INFRA-REDIS-SREM-ERROR",
                    message=f"Failed to remove from set: {key}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_srem)

    def expire(self, key: str, seconds: int) -> bool:
        """
        Set expiration on a key
        Matches Go Expire method
        """

        def _expire():
            if not self.client:
                raise ServiceError(
                    code="INFRA-REDIS-CLIENT-ERROR",
                    message="Redis client not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                result = self.client.expire(key, seconds)
                self.logger.info("Redis EXPIRE", key=key, seconds=seconds)
                return result

            except RedisError as e:
                self.logger.error(
                    "Redis EXPIRE failed",
                    error=str(e),
                    key=key,
                    error_code="INFRA-REDIS-EXPIRE-ERROR",
                )
                raise ServiceError(
                    code="INFRA-REDIS-EXPIRE-ERROR",
                    message=f"Failed to set expiration: {key}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_expire)

    def lpush(self, key: str, *values: Any) -> int:
        """
        Push values to the head of a list
        Returns the length of the list after push
        """

        def _lpush():
            if not self.client:
                raise ServiceError(
                    code="INFRA-REDIS-CLIENT-ERROR",
                    message="Redis client not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                count = self.client.lpush(key, *values)
                self.logger.info(
                    "Redis LPUSH", key=key, values_count=len(values), list_length=count
                )
                return count

            except RedisError as e:
                self.logger.error(
                    "Redis LPUSH failed",
                    error=str(e),
                    key=key,
                    error_code="INFRA-REDIS-LPUSH-ERROR",
                )
                raise ServiceError(
                    code="INFRA-REDIS-LPUSH-ERROR",
                    message=f"Failed to push to list: {key}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_lpush)

    def ltrim(self, key: str, start: int, stop: int) -> bool:
        """
        Trim a list to the specified range
        """

        def _ltrim():
            if not self.client:
                raise ServiceError(
                    code="INFRA-REDIS-CLIENT-ERROR",
                    message="Redis client not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                result = self.client.ltrim(key, start, stop)
                self.logger.info("Redis LTRIM", key=key, start=start, stop=stop)
                return result

            except RedisError as e:
                self.logger.error(
                    "Redis LTRIM failed",
                    error=str(e),
                    key=key,
                    error_code="INFRA-REDIS-LTRIM-ERROR",
                )
                raise ServiceError(
                    code="INFRA-REDIS-LTRIM-ERROR",
                    message=f"Failed to trim list: {key}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_ltrim)

    def lrange(self, key: str, start: int, stop: int) -> List[str]:
        """
        Get a range of elements from a list

        Args:
            key: List key
            start: Start index (0-based)
            stop: Stop index (inclusive, -1 for all)

        Returns:
            List of values
        """

        def _lrange():
            if not self.client:
                raise ServiceError(
                    code="INFRA-REDIS-CLIENT-ERROR",
                    message="Redis client not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                result = self.client.lrange(key, start, stop)
                # Decode bytes to strings
                decoded = []
                for item in result:
                    if isinstance(item, bytes):
                        decoded.append(item.decode("utf-8"))
                    else:
                        decoded.append(str(item))
                self.logger.debug(
                    "Redis LRANGE", key=key, start=start, stop=stop, count=len(decoded)
                )
                return decoded

            except RedisError as e:
                self.logger.error(
                    "Redis LRANGE failed",
                    error=str(e),
                    key=key,
                    error_code="INFRA-REDIS-LRANGE-ERROR",
                )
                raise ServiceError(
                    code="INFRA-REDIS-LRANGE-ERROR",
                    message=f"Failed to get list range: {key}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_lrange)

    def health(self) -> bool:
        """
        Check Redis health
        Matches Go Health method
        """
        try:
            if not self.client:
                return False

            response = self.client.ping()
            return response is True

        except Exception as e:
            self.logger.warning("Redis health check failed", error=str(e))
            return False

    def close(self) -> None:
        """
        Close connection
        Matches Go Close method
        """
        if self.client:
            self.client.close()
            self.logger.info("Redis connection closed")
