"""
Infrastructure clients for external dependencies

Provides consistent, production-ready clients for:
- ScyllaDB/Cassandra
- Redis
- Kafka
- HTTP services
- MongoDB (with vector search)
- SQL Server / Azure SQL
- Azure Key Vault
- Azure OpenAI (LLM)

All clients follow patterns:
- Dependency injection with config
- Structured logging with correlation IDs
- Health checks
- Circuit breakers
- Error handling with ServiceError
"""

from .scylladb import ScyllaDBClient, ScyllaDBConfig
from .redis_client import RedisClient, RedisConfig
from .kafka_client import KafkaProducer, KafkaConsumer, KafkaConfig
from .http_client import HTTPClient, HTTPConfig
from .mongodb_client import MongoDBClient, MongoDBConfig
from .sqlserver_client import SQLServerClient, SQLServerConfig
from .keyvault_client import KeyVaultClient, KeyVaultConfig
from .llm_client import AzureOpenAIClient, LLMConfig, LLMResponse, EmbeddingResponse

__all__ = [
    # ScyllaDB
    "ScyllaDBClient",
    "ScyllaDBConfig",
    # Redis
    "RedisClient",
    "RedisConfig",
    # Kafka
    "KafkaProducer",
    "KafkaConsumer",
    "KafkaConfig",
    # HTTP
    "HTTPClient",
    "HTTPConfig",
    # MongoDB
    "MongoDBClient",
    "MongoDBConfig",
    # SQL Server
    "SQLServerClient",
    "SQLServerConfig",
    # Key Vault
    "KeyVaultClient",
    "KeyVaultConfig",
    # Azure OpenAI
    "AzureOpenAIClient",
    "LLMConfig",
    "LLMResponse",
    "EmbeddingResponse",
]
