"""
Tests for ScyllaDB client
"""

import pytest
from unittest.mock import Mock, patch, MagicMock

# Skip entire module if cassandra driver has import issues (Python 3.13 compatibility)
try:
    from cassandra.cluster import Session as CassandraSession, ResultSet
    from cassandra import OperationTimedOut, Unavailable
    from core.infrastructure import ScyllaDBClient, ScyllaDBConfig

    CASSANDRA_AVAILABLE = True
except ImportError as e:
    CASSANDRA_AVAILABLE = False
    pytestmark = pytest.mark.skip(reason=f"Cassandra driver not available: {e}")

from core.logger import Logger
from core.errors import ServiceError


@pytest.fixture
def logger():
    """Create a test logger"""
    return Logger("test-scylladb", "INFO")


@pytest.fixture
def config(logger):
    """Create a valid ScyllaDBConfig"""
    return ScyllaDBConfig(
        hosts=["localhost"],
        keyspace="test_keyspace",
        logger=logger,
        timeout_seconds=60.0,
        connect_timeout_seconds=60.0,
    )


def test_config_validation_empty_hosts(logger):
    """Test that empty hosts raises ServiceError"""
    with pytest.raises(ServiceError) as exc_info:
        ScyllaDBConfig(
            hosts=[],
            keyspace="test",
            logger=logger,
        )
    assert exc_info.value.code == "INFRA-SCYLLADB-CONFIG-ERROR"
    assert "hosts cannot be empty" in exc_info.value.message.lower()


def test_config_validation_empty_keyspace(logger):
    """Test that empty keyspace raises ServiceError"""
    with pytest.raises(ServiceError) as exc_info:
        ScyllaDBConfig(
            hosts=["localhost"],
            keyspace="",
            logger=logger,
        )
    assert exc_info.value.code == "INFRA-SCYLLADB-CONFIG-ERROR"
    assert "keyspace cannot be empty" in exc_info.value.message.lower()


def test_client_initialization(config):
    """Test client initializes with valid config"""
    client = ScyllaDBClient(config)
    assert client.config == config
    assert client.session is None
    assert client.cluster is None


@patch("core.infrastructure.scylladb.Cluster")
def test_connect_success(mock_cluster_class, config):
    """Test successful connection to ScyllaDB"""
    # Mock cluster and session
    mock_session = Mock(spec=CassandraSession)
    mock_session.execute.return_value = Mock(spec=ResultSet)
    mock_cluster = Mock()
    mock_cluster.connect.return_value = mock_session
    mock_cluster_class.return_value = mock_cluster

    client = ScyllaDBClient(config)
    client.connect()

    assert client.session is not None
    assert client.cluster is not None
    mock_cluster_class.assert_called_once()
    mock_cluster.connect.assert_called_once_with(config.keyspace)


@patch("core.infrastructure.scylladb.Cluster")
def test_connect_failure(mock_cluster_class, config):
    """Test connection failure handling"""
    mock_cluster = Mock()
    mock_cluster.connect.side_effect = Exception("Connection failed")
    mock_cluster_class.return_value = mock_cluster

    client = ScyllaDBClient(config)

    with pytest.raises(ServiceError) as exc_info:
        client.connect()

    assert exc_info.value.code == "INFRA-SCYLLADB-CONNECTION-ERROR"
    assert "Failed to connect" in exc_info.value.message


def test_execute_not_connected(config):
    """Test execute without connection raises error"""
    client = ScyllaDBClient(config)

    with pytest.raises(ServiceError) as exc_info:
        client.execute("SELECT * FROM test")

    assert exc_info.value.code == "INFRA-SCYLLADB-CLIENT-ERROR"


@patch("core.infrastructure.scylladb.Cluster")
def test_execute_success(mock_cluster_class, config):
    """Test successful query execution"""
    # Mock session
    mock_result = Mock(spec=ResultSet)
    mock_session = Mock(spec=CassandraSession)
    mock_session.execute.return_value = mock_result

    mock_cluster = Mock()
    mock_cluster.connect.return_value = mock_session
    mock_cluster_class.return_value = mock_cluster

    client = ScyllaDBClient(config)
    client.connect()

    # Reset mock to ignore health check
    mock_session.execute.reset_mock()

    result = client.execute("SELECT * FROM test", ("param1",))

    assert result == mock_result
    assert mock_session.execute.call_count == 1


@patch("core.infrastructure.scylladb.Cluster")
def test_execute_timeout(mock_cluster_class, config):
    """Test query timeout handling"""
    mock_session = Mock(spec=CassandraSession)
    mock_session.execute.side_effect = OperationTimedOut("Query timed out")

    mock_cluster = Mock()
    mock_cluster.connect.return_value = mock_session
    mock_cluster_class.return_value = mock_cluster

    client = ScyllaDBClient(config)
    client.connect()

    with pytest.raises(ServiceError) as exc_info:
        client.execute("SELECT * FROM test")

    assert exc_info.value.code == "INFRA-SCYLLADB-TIMEOUT"


@patch("core.infrastructure.scylladb.Cluster")
def test_execute_unavailable(mock_cluster_class, config):
    """Test unavailable error handling"""
    mock_session = Mock(spec=CassandraSession)
    mock_session.execute.side_effect = Unavailable("Not enough replicas")

    mock_cluster = Mock()
    mock_cluster.connect.return_value = mock_session
    mock_cluster_class.return_value = mock_cluster

    client = ScyllaDBClient(config)
    client.connect()

    with pytest.raises(ServiceError) as exc_info:
        client.execute("SELECT * FROM test")

    assert exc_info.value.code == "INFRA-SCYLLADB-UNAVAILABLE"


@patch("core.infrastructure.scylladb.Cluster")
def test_execute_async_success(mock_cluster_class, config):
    """Test async query execution"""
    mock_future = Mock()
    mock_session = Mock(spec=CassandraSession)
    mock_session.execute_async.return_value = mock_future
    mock_session.execute.return_value = Mock()  # Health check

    mock_cluster = Mock()
    mock_cluster.connect.return_value = mock_session
    mock_cluster_class.return_value = mock_cluster

    client = ScyllaDBClient(config)
    client.connect()

    future = client.execute_async("SELECT * FROM test")

    assert future == mock_future
    mock_session.execute_async.assert_called_once()


@patch("core.infrastructure.scylladb.Cluster")
def test_health_check_success(mock_cluster_class, config):
    """Test successful health check"""
    mock_session = Mock(spec=CassandraSession)
    mock_session.execute.return_value = Mock()

    mock_cluster = Mock()
    mock_cluster.connect.return_value = mock_session
    mock_cluster_class.return_value = mock_cluster

    client = ScyllaDBClient(config)
    client.connect()

    assert client.health() is True


@patch("core.infrastructure.scylladb.Cluster")
def test_health_check_failure(mock_cluster_class, config):
    """Test failed health check"""
    mock_session = Mock(spec=CassandraSession)
    mock_session.execute.side_effect = Exception("Health check failed")

    mock_cluster = Mock()
    mock_cluster.connect.return_value = mock_session
    mock_cluster_class.return_value = mock_cluster

    client = ScyllaDBClient(config)
    client.connect()

    assert client.health() is False


def test_health_check_not_connected(config):
    """Test health check without connection"""
    client = ScyllaDBClient(config)
    assert client.health() is False


@patch("core.infrastructure.scylladb.Cluster")
def test_close(mock_cluster_class, config):
    """Test closing connection"""
    mock_session = Mock(spec=CassandraSession)
    mock_session.execute.return_value = Mock()
    mock_cluster = Mock()
    mock_cluster.connect.return_value = mock_session
    mock_cluster_class.return_value = mock_cluster

    client = ScyllaDBClient(config)
    client.connect()
    client.close()

    mock_session.shutdown.assert_called_once()
    mock_cluster.shutdown.assert_called_once()


def test_circuit_breaker_integration(config):
    """Test that circuit breaker is triggered after failures"""
    client = ScyllaDBClient(config)

    # Simulate multiple failures to open circuit
    for _ in range(6):  # More than max_failures
        try:
            client.execute("SELECT * FROM test")
        except:
            pass

    # Circuit should be open now
    assert client.circuit_breaker.state == "open"
