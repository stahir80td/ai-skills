"""
Tests for Redis client
"""

import pytest
from unittest.mock import Mock, patch
import redis
from redis.exceptions import RedisError, ConnectionError, TimeoutError

from core.logger import Logger
from core.errors import ServiceError

# Handle cassandra driver import issues that affect core.infrastructure (Python 3.13 compatibility)
try:
    from core.infrastructure import RedisClient, RedisConfig

    INFRASTRUCTURE_AVAILABLE = True
except ImportError as e:
    INFRASTRUCTURE_AVAILABLE = False
    pytestmark = pytest.mark.skip(
        reason=f"Infrastructure import failed (cassandra driver issue): {e}"
    )


@pytest.fixture
def logger():
    """Create a test logger"""
    return Logger("test-redis", "INFO")


@pytest.fixture
def config(logger):
    """Create a valid RedisConfig"""
    return RedisConfig(
        host="localhost",
        port=6379,
        logger=logger,
        ping_timeout_seconds=60.0,
        db=0,
    )


def test_config_validation_empty_host(logger):
    """Test that empty host raises ServiceError"""
    with pytest.raises(ServiceError) as exc_info:
        RedisConfig(
            host="",
            port=6379,
            logger=logger,
        )


def test_config_validation_invalid_port(logger):
    """Test that invalid port raises ServiceError"""
    with pytest.raises(ServiceError) as exc_info:
        RedisConfig(
            host="localhost",
            port=70000,
            logger=logger,
        )
    assert exc_info.value.code == "INFRA-REDIS-CONFIG-ERROR"


@patch("core.infrastructure.redis_client.redis.Redis")
def test_connect_success(mock_redis_class, config):
    """Test successful connection to Redis"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    assert client.client is not None
    mock_client.ping.assert_called_once()


@patch("core.infrastructure.redis_client.redis.Redis")
def test_connect_failure(mock_redis_class, config):
    """Test connection failure handling"""
    mock_client = Mock()
    mock_client.ping.side_effect = ConnectionError("Connection failed")
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)

    with pytest.raises(ServiceError) as exc_info:
        client.connect()

    assert exc_info.value.code == "INFRA-REDIS-CONNECT-ERROR"


@patch("core.infrastructure.redis_client.redis.Redis")
def test_get_success(mock_redis_class, config):
    """Test successful GET operation"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.get.return_value = "test_value"
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    result = client.get("test_key")

    assert result == "test_value"
    mock_client.get.assert_called_with("test_key")


@patch("core.infrastructure.redis_client.redis.Redis")
def test_get_not_found(mock_redis_class, config):
    """Test GET with non-existent key"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.get.return_value = None
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    result = client.get("nonexistent_key")

    assert result is None


def test_get_not_connected(config):
    """Test GET without connection raises error"""
    client = RedisClient(config)

    with pytest.raises(ServiceError) as exc_info:
        client.get("test_key")

    assert exc_info.value.code == "INFRA-REDIS-CLIENT-ERROR"


@patch("core.infrastructure.redis_client.redis.Redis")
def test_set_success(mock_redis_class, config):
    """Test successful SET operation"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.set.return_value = True
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    result = client.set("test_key", "test_value")

    assert result is True
    mock_client.set.assert_called_with("test_key", "test_value")


@patch("core.infrastructure.redis_client.redis.Redis")
def test_set_with_expiration(mock_redis_class, config):
    """Test SET with expiration"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.setex.return_value = True
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    result = client.set("test_key", "test_value", expiration_seconds=300)

    assert result is True
    mock_client.setex.assert_called_with("test_key", 300, "test_value")


@patch("core.infrastructure.redis_client.redis.Redis")
def test_set_dict_value(mock_redis_class, config):
    """Test SET with dict value (JSON serialization)"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.set.return_value = True
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    test_dict = {"key": "value", "number": 42}
    result = client.set("test_key", test_dict)

    assert result is True
    # Should serialize dict to JSON
    assert mock_client.set.called


@patch("core.infrastructure.redis_client.redis.Redis")
def test_delete_success(mock_redis_class, config):
    """Test successful DELETE operation"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.delete.return_value = 1
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    result = client.delete("test_key")

    assert result == 1
    mock_client.delete.assert_called_with("test_key")


@patch("core.infrastructure.redis_client.redis.Redis")
def test_delete_multiple_keys(mock_redis_class, config):
    """Test DELETE with multiple keys"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.delete.return_value = 3
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    result = client.delete("key1", "key2", "key3")

    assert result == 3


@patch("core.infrastructure.redis_client.redis.Redis")
def test_smembers_success(mock_redis_class, config):
    """Test successful SMEMBERS operation"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.smembers.return_value = {"member1", "member2", "member3"}
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    result = client.smembers("test_set")

    assert len(result) == 3
    assert "member1" in result


@patch("core.infrastructure.redis_client.redis.Redis")
def test_sadd_success(mock_redis_class, config):
    """Test successful SADD operation"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.sadd.return_value = 2
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    result = client.sadd("test_set", "member1", "member2")

    assert result == 2
    mock_client.sadd.assert_called_with("test_set", "member1", "member2")


@patch("core.infrastructure.redis_client.redis.Redis")
def test_srem_success(mock_redis_class, config):
    """Test successful SREM operation"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.srem.return_value = 1
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    result = client.srem("test_set", "member1")

    assert result == 1


@patch("core.infrastructure.redis_client.redis.Redis")
def test_expire_success(mock_redis_class, config):
    """Test successful EXPIRE operation"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.expire.return_value = True
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    result = client.expire("test_key", 300)

    assert result is True
    mock_client.expire.assert_called_with("test_key", 300)


@patch("core.infrastructure.redis_client.redis.Redis")
def test_health_check_success(mock_redis_class, config):
    """Test successful health check"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    assert client.health() is True


@patch("core.infrastructure.redis_client.redis.Redis")
def test_health_check_failure(mock_redis_class, config):
    """Test failed health check"""
    mock_client = Mock()
    mock_client.ping.side_effect = [True, RedisError("Health check failed")]
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    assert client.health() is False


def test_health_check_not_connected(config):
    """Test health check without connection"""
    client = RedisClient(config)
    assert client.health() is False


@patch("core.infrastructure.redis_client.redis.Redis")
def test_close(mock_redis_class, config):
    """Test closing connection"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()
    client.close()

    mock_client.close.assert_called_once()


@patch("core.infrastructure.redis_client.redis.Redis")
def test_circuit_breaker_integration(mock_redis_class, config):
    """Test circuit breaker triggers after failures"""
    mock_client = Mock()
    mock_client.ping.return_value = True
    mock_client.get.side_effect = RedisError("Simulated error")
    mock_redis_class.return_value = mock_client

    client = RedisClient(config)
    client.connect()

    # Trigger multiple failures
    for _ in range(6):
        try:
            client.get("test_key")
        except:
            pass

    # Circuit should be open
    assert client.circuit_breaker.state == "open"
