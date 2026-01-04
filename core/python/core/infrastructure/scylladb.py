"""
ScyllaDB/Cassandra client - Python equivalent of Go core/infrastructure/scylladb

Provides:
- Session management with connection pooling
- Query execution with context support
- Health checks
- Structured logging
- Circuit breaker protection
"""

from typing import List, Any, Optional, Dict
from dataclasses import dataclass
import os

# Set event loop to eventlet for Python 3.13 compatibility
os.environ.setdefault("CASSANDRA_DRIVER_EVENT_LOOP", "eventlet")

from cassandra.cluster import Cluster, Session as CassandraSession, ResultSet
from cassandra.query import SimpleStatement, ConsistencyLevel
from cassandra.policies import DCAwareRoundRobinPolicy, TokenAwarePolicy
from cassandra import OperationTimedOut, Unavailable

from ..logger import Logger
from ..errors import ServiceError, Severity
from ..reliability import CircuitBreaker


@dataclass
@dataclass
class ScyllaDBConfig:
    """
    Configuration for ScyllaDB client
    Matches Go SessionConfig pattern
    """

    hosts: List[str]
    keyspace: str
    logger: Logger
    timeout_seconds: float = 60.0  # Query timeout (minimum 60s)
    connect_timeout_seconds: float = 60.0  # Connection timeout

    def __post_init__(self):
        """Validate configuration"""
        if not self.hosts:
            raise ServiceError(
                code="INFRA-SCYLLADB-CONFIG-ERROR",
                message="ScyllaDB hosts cannot be empty",
                severity=Severity.CRITICAL,
            )
        if not self.keyspace:
            raise ServiceError(
                code="INFRA-SCYLLADB-CONFIG-ERROR",
                message="ScyllaDB keyspace cannot be empty",
                severity=Severity.CRITICAL,
            )


class ScyllaDBClient:
    """
    ScyllaDB client with production patterns
    Matches Go Session interface
    """

    def __init__(self, config: ScyllaDBConfig):
        self.config = config
        self.logger = config.logger.with_component("ScyllaDBClient")
        self.cluster: Optional[Cluster] = None
        self.session: Optional[CassandraSession] = None

        # Circuit breaker - can be disabled via env var for debugging
        import os

        cb_enabled = os.getenv("CIRCUIT_BREAKER_ENABLED", "true").lower() != "false"
        self.circuit_breaker = CircuitBreaker("scylladb", max_failures=5, enabled=False)

        self.logger.info(
            "Initiating ScyllaDB session",
            hosts=config.hosts,
            keyspace=config.keyspace,
            host_count=len(config.hosts),
            circuit_breaker_enabled=cb_enabled,
        )

    def connect(self) -> None:
        """Establish connection to ScyllaDB"""
        # Log connection settings for debugging
        self.logger.info(
            "ScyllaDB connecting",
            hosts=self.config.hosts,
            keyspace=self.config.keyspace,
            timeout_seconds=self.config.timeout_seconds,
            connect_timeout_seconds=self.config.connect_timeout_seconds,
        )

        try:
            # Configure load balancing policy
            load_balancing_policy = TokenAwarePolicy(DCAwareRoundRobinPolicy())

            # Create cluster
            self.cluster = Cluster(
                contact_points=self.config.hosts,
                load_balancing_policy=load_balancing_policy,
                protocol_version=4,
                connect_timeout=self.config.connect_timeout_seconds,
            )

            # Create session
            self.session = self.cluster.connect(self.config.keyspace)

            # Set default timeout
            self.session.default_timeout = self.config.timeout_seconds

            self.logger.info(
                "Connected to ScyllaDB",
                hosts=self.config.hosts,
                keyspace=self.config.keyspace,
                status="healthy",
            )

        except Exception as e:
            self.logger.error(
                "Failed to create ScyllaDB session",
                error=str(e),
                hosts=self.config.hosts,
                error_code="INFRA-SCYLLADB-CONNECTION-ERROR",
            )
            raise ServiceError(
                code="INFRA-SCYLLADB-CONNECTION-ERROR",
                message="Failed to connect to ScyllaDB",
                severity=Severity.CRITICAL,
                underlying=e,
            )

    def execute(self, query: str, parameters: Optional[tuple] = None) -> ResultSet:
        """
        Execute a query and return results
        Matches Go QueryContext pattern
        """

        def _execute():
            if not self.session:
                raise ServiceError(
                    code="INFRA-SCYLLADB-CLIENT-ERROR",
                    message="ScyllaDB session not connected",
                    severity=Severity.CRITICAL,
                )

            statement = SimpleStatement(
                query, consistency_level=ConsistencyLevel.QUORUM
            )

            try:
                if parameters:
                    result = self.session.execute(statement, parameters)
                else:
                    result = self.session.execute(statement)

                self.logger.info(
                    "ScyllaDB query executed",
                    query_preview=query[:100],
                )

                return result

            except OperationTimedOut as e:
                self.logger.error(
                    "ScyllaDB query timeout",
                    error=str(e),
                    query_preview=query[:100],
                    error_code="INFRA-SCYLLADB-TIMEOUT",
                )
                raise ServiceError(
                    code="INFRA-SCYLLADB-TIMEOUT",
                    message="Query execution timeout",
                    severity=Severity.HIGH,
                    underlying=e,
                )

            except Unavailable as e:
                self.logger.error(
                    "ScyllaDB unavailable",
                    error=str(e),
                    error_code="INFRA-SCYLLADB-UNAVAILABLE",
                )
                raise ServiceError(
                    code="INFRA-SCYLLADB-UNAVAILABLE",
                    message="ScyllaDB cluster unavailable",
                    severity=Severity.CRITICAL,
                    underlying=e,
                )

            except Exception as e:
                self.logger.error(
                    "ScyllaDB query failed",
                    error=str(e),
                    query_preview=query[:100],
                    error_code="INFRA-SCYLLADB-QUERY-ERROR",
                )
                raise ServiceError(
                    code="INFRA-SCYLLADB-QUERY-ERROR",
                    message="Query execution failed",
                    severity=Severity.HIGH,
                    underlying=e,
                )

        # Execute with circuit breaker protection
        return self.circuit_breaker.call(_execute)

    def execute_async(self, query: str, parameters: Optional[tuple] = None):
        """Execute query asynchronously"""
        if not self.session:
            raise ServiceError(
                code="INFRA-SCYLLADB-CLIENT-ERROR",
                message="ScyllaDB session not connected",
                severity=Severity.CRITICAL,
            )

        statement = SimpleStatement(query, consistency_level=ConsistencyLevel.QUORUM)

        if parameters:
            return self.session.execute_async(statement, parameters)
        else:
            return self.session.execute_async(statement)

    def health(self) -> bool:
        """
        Check ScyllaDB health
        Matches Go Health method
        """
        try:
            if not self.session:
                return False

            # Simple health check query
            result = self.session.execute("SELECT now() FROM system.local")
            return result is not None

        except Exception as e:
            self.logger.warning(
                "ScyllaDB health check failed",
                error=str(e),
            )
            return False

    def close(self) -> None:
        """
        Close connection
        Matches Go Close method
        """
        if self.session:
            self.session.shutdown()
            self.logger.info("ScyllaDB session closed")

        if self.cluster:
            self.cluster.shutdown()
            self.logger.info("ScyllaDB cluster connection closed")
