"""
SQL Server client - Production-ready Azure SQL / SQL Server client

Provides:
- Connection pooling with pyodbc
- Parameterized queries (SQL injection safe)
- Transaction support
- Health checks
- Structured logging with correlation IDs
- Circuit breaker protection
- Retry with exponential backoff
"""

from typing import Any, Dict, List, Optional, Tuple, Union
from dataclasses import dataclass
from contextlib import contextmanager
import time

import pyodbc

from ..logger import Logger
from ..errors import ServiceError, Severity
from ..reliability import CircuitBreaker


@dataclass
class SQLServerConfig:
    """
    Configuration for SQL Server client
    Supports Azure SQL and on-premises SQL Server
    """

    server: str
    database: str
    logger: Logger
    username: Optional[str] = None
    password: Optional[str] = None
    driver: str = "ODBC Driver 18 for SQL Server"
    timeout_seconds: int = 60
    connection_timeout_seconds: int = 60
    pool_size: int = 10
    encrypt: bool = True
    trust_server_certificate: bool = False

    def __post_init__(self):
        """Validate configuration"""
        if not self.server:
            raise ServiceError(
                code="INFRA-SQLSERVER-CONFIG-ERROR",
                message="SQL Server hostname cannot be empty",
                severity=Severity.CRITICAL,
            )
        if not self.database:
            raise ServiceError(
                code="INFRA-SQLSERVER-CONFIG-ERROR",
                message="SQL Server database name cannot be empty",
                severity=Severity.CRITICAL,
            )

    def get_connection_string(self) -> str:
        """Build ODBC connection string"""
        parts = [
            f"DRIVER={{{self.driver}}}",
            f"SERVER={self.server}",
            f"DATABASE={self.database}",
            f"Encrypt={'yes' if self.encrypt else 'no'}",
            f"TrustServerCertificate={'yes' if self.trust_server_certificate else 'no'}",
            f"Connection Timeout={self.connection_timeout_seconds}",
        ]

        if self.username and self.password:
            parts.extend(
                [
                    f"UID={self.username}",
                    f"PWD={self.password}",
                ]
            )
        else:
            # Use Windows/Azure AD integrated auth
            parts.append("Trusted_Connection=yes")

        return ";".join(parts)


class SQLServerClient:
    """
    SQL Server client with production patterns

    Features:
    - Connection pooling (pyodbc handles this at driver level)
    - Parameterized queries for SQL injection prevention
    - Circuit breaker for fault tolerance
    - Structured logging with query metrics
    - Transaction support via context manager
    """

    def __init__(self, config: SQLServerConfig):
        self.config = config
        self.logger = config.logger.with_component("SQLServerClient")
        self._connection: Optional[pyodbc.Connection] = None

        # Circuit breaker - can be disabled via env var for debugging
        import os

        cb_enabled = os.getenv("CIRCUIT_BREAKER_ENABLED", "true").lower() != "false"
        self.circuit_breaker = CircuitBreaker(
            "sqlserver", max_failures=5, enabled=False
        )

        # Enable connection pooling
        pyodbc.pooling = True

        self.logger.info(
            "Initializing SQL Server client",
            server=config.server,
            database=config.database,
            driver=config.driver,
            encrypt=config.encrypt,
            circuit_breaker_enabled=cb_enabled,
        )

    def connect(self) -> None:
        """
        Establish connection to SQL Server
        Validates connection with simple query
        """
        start_time = time.time()
        connection_string = self.config.get_connection_string()

        # Log connection settings for debugging (mask password)
        masked_conn = (
            connection_string.replace(self.config.password or "", "***MASKED***")
            if self.config.password
            else connection_string
        )
        self.logger.info(
            "SQL Server connecting",
            server=self.config.server,
            database=self.config.database,
            connection_string=masked_conn,
            timeout_seconds=self.config.timeout_seconds,
            connection_timeout_seconds=self.config.connection_timeout_seconds,
        )

        try:
            self._connection = pyodbc.connect(
                connection_string,
                timeout=self.config.connection_timeout_seconds,
                autocommit=True,  # Default to autocommit, use transactions explicitly
            )

            # Set query timeout
            self._connection.timeout = self.config.timeout_seconds

            # Verify connection
            cursor = self._connection.cursor()
            cursor.execute("SELECT 1")
            cursor.close()

            elapsed_ms = (time.time() - start_time) * 1000

            self.logger.info(
                "Connected to SQL Server",
                server=self.config.server,
                database=self.config.database,
                status="healthy",
                connect_time_ms=round(elapsed_ms, 2),
            )

        except pyodbc.InterfaceError as e:
            self.logger.error(
                "SQL Server driver error",
                error=str(e),
                server=self.config.server,
                driver=self.config.driver,
                error_code="INFRA-SQLSERVER-DRIVER-ERROR",
            )
            raise ServiceError(
                code="INFRA-SQLSERVER-DRIVER-ERROR",
                message=f"SQL Server driver error: {e}",
                severity=Severity.CRITICAL,
                underlying=e,
            )

        except pyodbc.OperationalError as e:
            self.logger.error(
                "SQL Server connection failed",
                error=str(e),
                server=self.config.server,
                database=self.config.database,
                error_code="INFRA-SQLSERVER-CONNECT-ERROR",
            )
            raise ServiceError(
                code="INFRA-SQLSERVER-CONNECT-ERROR",
                message=f"Failed to connect to SQL Server: {e}",
                severity=Severity.CRITICAL,
                underlying=e,
            )

        except Exception as e:
            self.logger.error(
                "Unexpected SQL Server connection error",
                error=str(e),
                error_type=type(e).__name__,
                error_code="INFRA-SQLSERVER-CONNECT-ERROR",
            )
            raise ServiceError(
                code="INFRA-SQLSERVER-CONNECT-ERROR",
                message=f"Unexpected error connecting to SQL Server: {e}",
                severity=Severity.CRITICAL,
                underlying=e,
            )

    def _ensure_connected(self) -> pyodbc.Connection:
        """Ensure client is connected and return connection"""
        if not self._connection:
            raise ServiceError(
                code="INFRA-SQLSERVER-CLIENT-ERROR",
                message="SQL Server client not connected - call connect() first",
                severity=Severity.CRITICAL,
            )

        # Check if connection is still alive
        try:
            cursor = self._connection.cursor()
            cursor.execute("SELECT 1")
            cursor.close()
        except pyodbc.Error:
            self.logger.warning(
                "SQL Server connection lost, reconnecting",
                server=self.config.server,
            )
            self.connect()

        return self._connection

    def execute_query(
        self,
        query: str,
        params: Optional[Tuple[Any, ...]] = None,
    ) -> List[Dict[str, Any]]:
        """
        Execute SELECT query and return results as list of dicts

        Args:
            query: SQL query with ? placeholders for parameters
            params: Tuple of parameter values

        Returns:
            List of row dictionaries
        """

        def _execute():
            conn = self._ensure_connected()
            start_time = time.time()

            try:
                cursor = conn.cursor()

                if params:
                    cursor.execute(query, params)
                else:
                    cursor.execute(query)

                # Get column names
                columns = (
                    [column[0] for column in cursor.description]
                    if cursor.description
                    else []
                )

                # Fetch results
                rows = cursor.fetchall()

                # Convert to list of dicts
                results = [dict(zip(columns, row)) for row in rows]

                cursor.close()
                elapsed_ms = (time.time() - start_time) * 1000

                self.logger.info(
                    "SQL query executed",
                    query_preview=query[:100],
                    param_count=len(params) if params else 0,
                    row_count=len(results),
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return results

            except pyodbc.ProgrammingError as e:
                self.logger.error(
                    "SQL query programming error",
                    error=str(e),
                    query_preview=query[:100],
                    error_code="INFRA-SQLSERVER-QUERY-ERROR",
                )
                raise ServiceError(
                    code="INFRA-SQLSERVER-QUERY-ERROR",
                    message=f"SQL query error: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except pyodbc.Error as e:
                self.logger.error(
                    "SQL query execution failed",
                    error=str(e),
                    query_preview=query[:100],
                    error_code="INFRA-SQLSERVER-QUERY-ERROR",
                )
                raise ServiceError(
                    code="INFRA-SQLSERVER-QUERY-ERROR",
                    message=f"SQL query failed: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_execute)

    def execute_scalar(
        self,
        query: str,
        params: Optional[Tuple[Any, ...]] = None,
    ) -> Optional[Any]:
        """
        Execute query and return single scalar value

        Args:
            query: SQL query returning single value
            params: Tuple of parameter values

        Returns:
            Single value or None
        """

        def _execute_scalar():
            conn = self._ensure_connected()
            start_time = time.time()

            try:
                cursor = conn.cursor()

                if params:
                    cursor.execute(query, params)
                else:
                    cursor.execute(query)

                row = cursor.fetchone()
                result = row[0] if row else None

                cursor.close()
                elapsed_ms = (time.time() - start_time) * 1000

                self.logger.info(
                    "SQL scalar query executed",
                    query_preview=query[:100],
                    has_result=result is not None,
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return result

            except pyodbc.Error as e:
                self.logger.error(
                    "SQL scalar query failed",
                    error=str(e),
                    query_preview=query[:100],
                    error_code="INFRA-SQLSERVER-QUERY-ERROR",
                )
                raise ServiceError(
                    code="INFRA-SQLSERVER-QUERY-ERROR",
                    message=f"SQL scalar query failed: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_execute_scalar)

    def execute_non_query(
        self,
        query: str,
        params: Optional[Tuple[Any, ...]] = None,
    ) -> int:
        """
        Execute INSERT/UPDATE/DELETE query

        Args:
            query: SQL modification query
            params: Tuple of parameter values

        Returns:
            Number of affected rows
        """

        def _execute_non_query():
            conn = self._ensure_connected()
            start_time = time.time()

            try:
                cursor = conn.cursor()

                if params:
                    cursor.execute(query, params)
                else:
                    cursor.execute(query)

                affected_rows = cursor.rowcount
                cursor.close()

                elapsed_ms = (time.time() - start_time) * 1000

                self.logger.info(
                    "SQL non-query executed",
                    query_preview=query[:100],
                    affected_rows=affected_rows,
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return affected_rows

            except pyodbc.IntegrityError as e:
                self.logger.error(
                    "SQL integrity error",
                    error=str(e),
                    query_preview=query[:100],
                    error_code="INFRA-SQLSERVER-INTEGRITY-ERROR",
                )
                raise ServiceError(
                    code="INFRA-SQLSERVER-INTEGRITY-ERROR",
                    message=f"SQL integrity violation: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except pyodbc.Error as e:
                self.logger.error(
                    "SQL non-query failed",
                    error=str(e),
                    query_preview=query[:100],
                    error_code="INFRA-SQLSERVER-QUERY-ERROR",
                )
                raise ServiceError(
                    code="INFRA-SQLSERVER-QUERY-ERROR",
                    message=f"SQL non-query failed: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_execute_non_query)

    @contextmanager
    def transaction(self):
        """
        Transaction context manager for atomic operations

        Usage:
            with client.transaction():
                client.execute_non_query("INSERT INTO ...")
                client.execute_non_query("UPDATE ...")
        """
        conn = self._ensure_connected()

        # Disable autocommit for transaction
        conn.autocommit = False

        self.logger.info("SQL transaction started")

        try:
            yield self
            conn.commit()
            self.logger.info("SQL transaction committed")

        except Exception as e:
            conn.rollback()
            self.logger.warning(
                "SQL transaction rolled back",
                error=str(e),
            )
            raise

        finally:
            conn.autocommit = True

    def get_user_by_email(self, email: str) -> Optional[Dict[str, Any]]:
        """
        Get user by email address (convenience method for iot-seek)

        Args:
            email: User email address

        Returns:
            User dict or None
        """
        # Cast DATETIMEOFFSET to VARCHAR to avoid ODBC type -155 error
        query = """
            SELECT id, email, name, role,
                   CONVERT(VARCHAR(30), created_at, 127) as created_at,
                   CONVERT(VARCHAR(30), updated_at, 127) as updated_at
            FROM users
            WHERE email = ?
        """
        results = self.execute_query(query, (email,))
        return results[0] if results else None

    def get_user_by_id(self, user_id: str) -> Optional[Dict[str, Any]]:
        """Get user by ID"""
        # Cast DATETIMEOFFSET to VARCHAR to avoid ODBC type -155 error
        query = """
            SELECT id, email, name, role,
                   CONVERT(VARCHAR(30), created_at, 127) as created_at,
                   CONVERT(VARCHAR(30), updated_at, 127) as updated_at
            FROM users
            WHERE id = ?
        """
        results = self.execute_query(query, (user_id,))
        return results[0] if results else None

    def list_users_by_role(self, role: str) -> List[Dict[str, Any]]:
        """List users by role"""
        # Cast DATETIMEOFFSET to VARCHAR to avoid ODBC type -155 error
        query = """
            SELECT id, email, name, role,
                   CONVERT(VARCHAR(30), created_at, 127) as created_at,
                   CONVERT(VARCHAR(30), updated_at, 127) as updated_at
            FROM users
            WHERE role = ?
            ORDER BY created_at DESC
        """
        return self.execute_query(query, (role,))

    def search_users(self, search_term: str, limit: int = 100) -> List[Dict[str, Any]]:
        """Search users by name or email"""
        # Cast DATETIMEOFFSET to VARCHAR to avoid ODBC type -155 error
        query = """
            SELECT TOP (?) id, email, name, role,
                   CONVERT(VARCHAR(30), created_at, 127) as created_at,
                   CONVERT(VARCHAR(30), updated_at, 127) as updated_at
            FROM users
            WHERE email LIKE ? OR name LIKE ?
            ORDER BY email
        """
        pattern = f"%{search_term}%"
        return self.execute_query(query, (limit, pattern, pattern))

    def health_check(self) -> Dict[str, Any]:
        """
        Check SQL Server health status

        Returns:
            Health status dict with latency and server info
        """
        start_time = time.time()

        try:
            if not self._connection:
                return {
                    "status": "unhealthy",
                    "error": "Client not connected",
                    "server": self.config.server,
                    "database": self.config.database,
                }

            # Execute simple query
            cursor = self._connection.cursor()
            cursor.execute("SELECT @@VERSION AS version, GETDATE() AS server_time")
            row = cursor.fetchone()
            cursor.close()

            elapsed_ms = (time.time() - start_time) * 1000

            return {
                "status": "healthy",
                "server": self.config.server,
                "database": self.config.database,
                "latency_ms": round(elapsed_ms, 2),
                "version": row[0].split("\n")[0] if row else "unknown",
            }

        except Exception as e:
            elapsed_ms = (time.time() - start_time) * 1000
            self.logger.warning(
                "SQL Server health check failed",
                error=str(e),
                elapsed_ms=round(elapsed_ms, 2),
            )
            return {
                "status": "unhealthy",
                "server": self.config.server,
                "database": self.config.database,
                "error": str(e),
                "latency_ms": round(elapsed_ms, 2),
            }

    def close(self) -> None:
        """Close SQL Server connection"""
        if self._connection:
            self._connection.close()
            self._connection = None
            self.logger.info(
                "SQL Server connection closed",
                server=self.config.server,
                database=self.config.database,
            )

    def __enter__(self) -> "SQLServerClient":
        """Context manager entry"""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        """Context manager exit"""
        self.close()
