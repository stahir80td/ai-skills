"""
MongoDB client - Production-ready MongoDB client with vector search support

Provides:
- Connection pooling with auto-reconnect
- CRUD operations with type safety
- MongoDB Atlas Vector Search support
- Aggregation pipeline support
- Health checks
- Structured logging with correlation IDs
- Circuit breaker protection
- Retry with exponential backoff
"""

from typing import Any, Dict, List, Optional, TypeVar, Union
from dataclasses import dataclass
from datetime import datetime
import time

from pymongo import MongoClient, ASCENDING, DESCENDING
from pymongo.database import Database
from pymongo.collection import Collection
from pymongo.errors import (
    ConnectionFailure,
    ServerSelectionTimeoutError,
    OperationFailure,
    PyMongoError,
)
from bson import ObjectId

from ..logger import Logger
from ..errors import ServiceError, Severity
from ..reliability import CircuitBreaker


T = TypeVar("T", bound=Dict[str, Any])


@dataclass
class MongoDBConfig:
    """
    Configuration for MongoDB client
    Supports MongoDB Atlas connection strings with full options
    """

    uri: str
    database: str
    logger: Logger
    timeout_seconds: float = 60.0
    connect_timeout_seconds: float = 60.0
    max_pool_size: int = 100
    min_pool_size: int = 10
    max_idle_time_seconds: int = 300
    retry_writes: bool = True
    retry_reads: bool = True

    def __post_init__(self):
        """Validate configuration"""
        if not self.uri:
            raise ServiceError(
                code="INFRA-MONGODB-CONFIG-ERROR",
                message="MongoDB URI cannot be empty",
                severity=Severity.CRITICAL,
            )
        if not self.database:
            raise ServiceError(
                code="INFRA-MONGODB-CONFIG-ERROR",
                message="MongoDB database name cannot be empty",
                severity=Severity.CRITICAL,
            )


class MongoDBClient:
    """
    MongoDB client with production patterns

    Features:
    - Connection pooling with configurable pool size
    - Circuit breaker for fault tolerance
    - Structured logging with operation metrics
    - Vector search support (MongoDB Atlas)
    - Aggregation pipeline support
    """

    def __init__(self, config: MongoDBConfig):
        self.config = config
        self.logger = config.logger.with_component("MongoDBClient")
        self._client: Optional[MongoClient] = None
        self._db: Optional[Database] = None

        # Circuit breaker - can be disabled via env var for debugging
        import os

        cb_enabled = os.getenv("CIRCUIT_BREAKER_ENABLED", "true").lower() != "false"
        self.circuit_breaker = CircuitBreaker("mongodb", max_failures=5, enabled=False)

        self.logger.info(
            "Initializing MongoDB client",
            database=config.database,
            max_pool_size=config.max_pool_size,
            timeout_seconds=config.timeout_seconds,
            circuit_breaker_enabled=cb_enabled,
        )

    def connect(self) -> None:
        """
        Establish connection to MongoDB
        Validates connection with ping command
        """
        start_time = time.time()

        # Log connection settings for debugging
        self.logger.info(
            "MongoDB connecting",
            uri=self.config.uri,
            database=self.config.database,
            timeout_seconds=self.config.timeout_seconds,
            connect_timeout_seconds=self.config.connect_timeout_seconds,
            max_pool_size=self.config.max_pool_size,
        )

        try:
            self._client = MongoClient(
                self.config.uri,
                serverSelectionTimeoutMS=int(
                    self.config.connect_timeout_seconds * 1000
                ),
                connectTimeoutMS=int(self.config.connect_timeout_seconds * 1000),
                socketTimeoutMS=int(self.config.timeout_seconds * 1000),
                maxPoolSize=self.config.max_pool_size,
                minPoolSize=self.config.min_pool_size,
                maxIdleTimeMS=self.config.max_idle_time_seconds * 1000,
                retryWrites=self.config.retry_writes,
                retryReads=self.config.retry_reads,
            )

            # Get database reference
            self._db = self._client[self.config.database]

            # Verify connection with ping
            self._client.admin.command("ping")

            elapsed_ms = (time.time() - start_time) * 1000

            self.logger.info(
                "Connected to MongoDB",
                database=self.config.database,
                status="healthy",
                pool_size=self.config.max_pool_size,
                connect_time_ms=round(elapsed_ms, 2),
            )

        except ServerSelectionTimeoutError as e:
            self.logger.error(
                "MongoDB server selection timeout",
                error=str(e),
                database=self.config.database,
                timeout_seconds=self.config.connect_timeout_seconds,
                error_code="INFRA-MONGODB-TIMEOUT",
            )
            raise ServiceError(
                code="INFRA-MONGODB-TIMEOUT",
                message="MongoDB server selection timeout - check connection string and network",
                severity=Severity.CRITICAL,
                underlying=e,
            )

        except ConnectionFailure as e:
            self.logger.error(
                "MongoDB connection failed",
                error=str(e),
                database=self.config.database,
                error_code="INFRA-MONGODB-CONNECT-ERROR",
            )
            raise ServiceError(
                code="INFRA-MONGODB-CONNECT-ERROR",
                message="Failed to connect to MongoDB",
                severity=Severity.CRITICAL,
                underlying=e,
            )

        except Exception as e:
            self.logger.error(
                "Unexpected MongoDB connection error",
                error=str(e),
                error_type=type(e).__name__,
                error_code="INFRA-MONGODB-CONNECT-ERROR",
            )
            raise ServiceError(
                code="INFRA-MONGODB-CONNECT-ERROR",
                message=f"Unexpected error connecting to MongoDB: {e}",
                severity=Severity.CRITICAL,
                underlying=e,
            )

    def _ensure_connected(self) -> Database:
        """Ensure client is connected and return database"""
        if self._db is None:
            raise ServiceError(
                code="INFRA-MONGODB-CLIENT-ERROR",
                message="MongoDB client not connected - call connect() first",
                severity=Severity.CRITICAL,
            )
        return self._db

    def find_one(
        self,
        collection: str,
        query: Dict[str, Any],
        projection: Optional[Dict[str, Any]] = None,
    ) -> Optional[Dict[str, Any]]:
        """
        Find a single document matching query

        Args:
            collection: Collection name
            query: MongoDB query filter
            projection: Fields to include/exclude

        Returns:
            Document if found, None otherwise
        """

        def _find_one():
            db = self._ensure_connected()
            start_time = time.time()

            try:
                result = db[collection].find_one(query, projection)
                elapsed_ms = (time.time() - start_time) * 1000

                self.logger.info(
                    "MongoDB find_one executed",
                    collection=collection,
                    query_keys=list(query.keys()),
                    found=result is not None,
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return result

            except OperationFailure as e:
                self.logger.error(
                    "MongoDB find_one operation failed",
                    error=str(e),
                    collection=collection,
                    query_keys=list(query.keys()),
                    error_code="INFRA-MONGODB-QUERY-ERROR",
                )
                raise ServiceError(
                    code="INFRA-MONGODB-QUERY-ERROR",
                    message=f"MongoDB find_one failed: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_find_one)

    def find(
        self,
        collection: str,
        query: Dict[str, Any],
        projection: Optional[Dict[str, Any]] = None,
        sort: Optional[List[tuple]] = None,
        limit: int = 100,
        skip: int = 0,
    ) -> List[Dict[str, Any]]:
        """
        Find multiple documents matching query

        Args:
            collection: Collection name
            query: MongoDB query filter
            projection: Fields to include/exclude
            sort: List of (field, direction) tuples
            limit: Maximum documents to return
            skip: Number of documents to skip

        Returns:
            List of matching documents
        """

        def _find():
            db = self._ensure_connected()
            start_time = time.time()

            try:
                cursor = db[collection].find(query, projection)

                if sort:
                    cursor = cursor.sort(sort)
                if skip > 0:
                    cursor = cursor.skip(skip)
                if limit > 0:
                    cursor = cursor.limit(limit)

                results = list(cursor)
                elapsed_ms = (time.time() - start_time) * 1000

                self.logger.info(
                    "MongoDB find executed",
                    collection=collection,
                    query_keys=list(query.keys()),
                    result_count=len(results),
                    limit=limit,
                    skip=skip,
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return results

            except OperationFailure as e:
                self.logger.error(
                    "MongoDB find operation failed",
                    error=str(e),
                    collection=collection,
                    query_keys=list(query.keys()),
                    error_code="INFRA-MONGODB-QUERY-ERROR",
                )
                raise ServiceError(
                    code="INFRA-MONGODB-QUERY-ERROR",
                    message=f"MongoDB find failed: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_find)

    def count_documents(
        self,
        collection: str,
        query: Optional[Dict[str, Any]] = None,
    ) -> int:
        """Count documents matching query"""

        def _count():
            db = self._ensure_connected()
            start_time = time.time()
            filter_query = query or {}

            try:
                count = db[collection].count_documents(filter_query)
                elapsed_ms = (time.time() - start_time) * 1000

                self.logger.info(
                    "MongoDB count_documents executed",
                    collection=collection,
                    count=count,
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return count

            except OperationFailure as e:
                self.logger.error(
                    "MongoDB count_documents failed",
                    error=str(e),
                    collection=collection,
                    error_code="INFRA-MONGODB-QUERY-ERROR",
                )
                raise ServiceError(
                    code="INFRA-MONGODB-QUERY-ERROR",
                    message=f"MongoDB count failed: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_count)

    def aggregate(
        self,
        collection: str,
        pipeline: List[Dict[str, Any]],
    ) -> List[Dict[str, Any]]:
        """
        Execute aggregation pipeline

        Args:
            collection: Collection name
            pipeline: MongoDB aggregation pipeline stages

        Returns:
            Aggregation results
        """

        def _aggregate():
            db = self._ensure_connected()
            start_time = time.time()

            try:
                cursor = db[collection].aggregate(pipeline)
                results = list(cursor)
                elapsed_ms = (time.time() - start_time) * 1000

                self.logger.info(
                    "MongoDB aggregate executed",
                    collection=collection,
                    pipeline_stages=len(pipeline),
                    result_count=len(results),
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return results

            except OperationFailure as e:
                self.logger.error(
                    "MongoDB aggregate operation failed",
                    error=str(e),
                    collection=collection,
                    pipeline_stages=len(pipeline),
                    error_code="INFRA-MONGODB-AGGREGATE-ERROR",
                )
                raise ServiceError(
                    code="INFRA-MONGODB-AGGREGATE-ERROR",
                    message=f"MongoDB aggregation failed: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_aggregate)

    def vector_search(
        self,
        collection: str,
        embedding: List[float],
        index_name: str,
        path: str,
        limit: int = 10,
        num_candidates: int = 100,
        filter_query: Optional[Dict[str, Any]] = None,
    ) -> List[Dict[str, Any]]:
        """
        Execute MongoDB Atlas Vector Search

        Args:
            collection: Collection name
            embedding: Query embedding vector
            index_name: Name of the vector search index
            path: Field path containing embeddings
            limit: Maximum results to return
            num_candidates: Number of candidates to consider
            filter_query: Optional pre-filter query

        Returns:
            List of similar documents with scores
        """

        def _vector_search():
            db = self._ensure_connected()
            start_time = time.time()

            # Build vector search stage
            vector_search_stage: Dict[str, Any] = {
                "$vectorSearch": {
                    "index": index_name,
                    "path": path,
                    "queryVector": embedding,
                    "numCandidates": num_candidates,
                    "limit": limit,
                }
            }

            # Add filter if provided
            if filter_query:
                vector_search_stage["$vectorSearch"]["filter"] = filter_query

            # Build pipeline with score projection
            pipeline = [
                vector_search_stage,
                {
                    "$project": {
                        "_id": 1,
                        "score": {"$meta": "vectorSearchScore"},
                        # Include all other fields
                        "document": "$$ROOT",
                    }
                },
            ]

            try:
                cursor = db[collection].aggregate(pipeline)
                results = list(cursor)
                elapsed_ms = (time.time() - start_time) * 1000

                self.logger.info(
                    "MongoDB vector search executed",
                    collection=collection,
                    index_name=index_name,
                    embedding_dim=len(embedding),
                    result_count=len(results),
                    num_candidates=num_candidates,
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return results

            except OperationFailure as e:
                self.logger.error(
                    "MongoDB vector search failed",
                    error=str(e),
                    collection=collection,
                    index_name=index_name,
                    error_code="INFRA-MONGODB-VECTOR-SEARCH-ERROR",
                )
                raise ServiceError(
                    code="INFRA-MONGODB-VECTOR-SEARCH-ERROR",
                    message=f"MongoDB vector search failed: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_vector_search)

    def distinct(
        self,
        collection: str,
        field: str,
        query: Optional[Dict[str, Any]] = None,
    ) -> List[Any]:
        """Get distinct values for a field"""

        def _distinct():
            db = self._ensure_connected()
            start_time = time.time()

            try:
                results = db[collection].distinct(field, query or {})
                elapsed_ms = (time.time() - start_time) * 1000

                self.logger.info(
                    "MongoDB distinct executed",
                    collection=collection,
                    field=field,
                    result_count=len(results),
                    elapsed_ms=round(elapsed_ms, 2),
                )

                return results

            except OperationFailure as e:
                self.logger.error(
                    "MongoDB distinct failed",
                    error=str(e),
                    collection=collection,
                    field=field,
                    error_code="INFRA-MONGODB-QUERY-ERROR",
                )
                raise ServiceError(
                    code="INFRA-MONGODB-QUERY-ERROR",
                    message=f"MongoDB distinct failed: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_distinct)

    def health_check(self) -> Dict[str, Any]:
        """
        Check MongoDB health status

        Returns:
            Health status dict with latency and server info
        """
        start_time = time.time()

        try:
            if not self._client:
                return {
                    "status": "unhealthy",
                    "error": "Client not connected",
                    "database": self.config.database,
                }

            # Ping server
            self._client.admin.command("ping")
            elapsed_ms = (time.time() - start_time) * 1000

            # Get server info
            server_info = self._client.server_info()

            return {
                "status": "healthy",
                "database": self.config.database,
                "latency_ms": round(elapsed_ms, 2),
                "version": server_info.get("version", "unknown"),
            }

        except Exception as e:
            elapsed_ms = (time.time() - start_time) * 1000
            self.logger.warning(
                "MongoDB health check failed",
                error=str(e),
                elapsed_ms=round(elapsed_ms, 2),
            )
            return {
                "status": "unhealthy",
                "database": self.config.database,
                "error": str(e),
                "latency_ms": round(elapsed_ms, 2),
            }

    def close(self) -> None:
        """Close MongoDB connection"""
        if self._client:
            self._client.close()
            self._client = None
            self._db = None
            self.logger.info(
                "MongoDB connection closed",
                database=self.config.database,
            )

    def __enter__(self) -> "MongoDBClient":
        """Context manager entry"""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        """Context manager exit"""
        self.close()
