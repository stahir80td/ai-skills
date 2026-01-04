"""
Tests for Kafka client
"""

import pytest
from unittest.mock import Mock, patch, MagicMock
from kafka.errors import KafkaError, KafkaTimeoutError

from core.logger import Logger
from core.errors import ServiceError

# Handle cassandra driver import issues that affect core.infrastructure (Python 3.13 compatibility)
try:
    from core.infrastructure import KafkaProducer, KafkaConsumer, KafkaConfig

    INFRASTRUCTURE_AVAILABLE = True
except ImportError as e:
    INFRASTRUCTURE_AVAILABLE = False
    pytestmark = pytest.mark.skip(
        reason=f"Infrastructure import failed (cassandra driver issue): {e}"
    )


@pytest.fixture
def logger():
    """Create a test logger"""
    return Logger("test-kafka", "INFO")


@pytest.fixture
def config(logger):
    """Create a valid KafkaConfig"""
    return KafkaConfig(
        brokers=["localhost:9092"],
        logger=logger,
        client_id="test-client",
        timeout_ms=60000,
    )


def test_config_validation_empty_brokers(logger):
    """Test that empty brokers raises ServiceError"""
    with pytest.raises(ServiceError) as exc_info:
        KafkaConfig(
            brokers=[],
            logger=logger,
        )


def test_producer_initialization(config):
    """Test producer initializes with valid config"""
    producer = KafkaProducer(config)
    assert producer.config == config
    assert producer.producer is None


@patch("core.infrastructure.kafka_client.BaseProducer")
def test_producer_connect_success(mock_producer_class, config):
    """Test successful producer connection"""
    mock_producer_instance = Mock()
    mock_producer_instance.bootstrap_connected.return_value = True
    mock_producer_class.return_value = mock_producer_instance

    producer = KafkaProducer(config)
    producer.connect()

    assert producer.producer is not None
    mock_producer_class.assert_called_once()


@patch("core.infrastructure.kafka_client.BaseProducer")
def test_producer_connect_failure(mock_producer_class, config):
    """Test producer connection failure"""
    mock_producer_class.side_effect = KafkaError("Connection failed")

    producer = KafkaProducer(config)

    with pytest.raises(ServiceError) as exc_info:
        producer.connect()

    assert exc_info.value.code == "INFRA-KAFKA-CONNECTION-ERROR"


def test_send_message_not_connected(config):
    """Test send_message without connection raises error"""
    producer = KafkaProducer(config)

    with pytest.raises(ServiceError) as exc_info:
        producer.send_message("test_topic", value="test")

    assert exc_info.value.code == "INFRA-KAFKA-PRODUCER-ERROR"


@patch("core.infrastructure.kafka_client.BaseProducer")
def test_send_message_success(mock_producer_class, config):
    """Test successful message send"""
    mock_future = Mock()
    mock_future.get.return_value = Mock()

    mock_producer_instance = Mock()
    mock_producer_instance.bootstrap_connected.return_value = True
    mock_producer_instance.send.return_value = mock_future
    mock_producer_class.return_value = mock_producer_instance

    producer = KafkaProducer(config)
    producer.connect()

    producer.send_message("test_topic", key="key1", value="value1")

    mock_producer_instance.send.assert_called_once()
    mock_future.get.assert_called_once()


@patch("core.infrastructure.kafka_client.BaseProducer")
def test_send_message_with_headers(mock_producer_class, config):
    """Test message send with headers"""
    mock_future = Mock()
    mock_future.get.return_value = Mock()

    mock_producer_instance = Mock()
    mock_producer_instance.bootstrap_connected.return_value = True
    mock_producer_instance.send.return_value = mock_future
    mock_producer_class.return_value = mock_producer_instance

    producer = KafkaProducer(config)
    producer.connect()

    headers = {"correlation_id": "12345", "service": "test"}
    producer.send_message("test_topic", value="value1", headers=headers)

    # Verify headers were passed
    call_args = mock_producer_instance.send.call_args
    assert call_args[1]["headers"] is not None


@patch("core.infrastructure.kafka_client.BaseProducer")
def test_send_message_timeout(mock_producer_class, config):
    """Test message send timeout"""
    mock_future = Mock()
    mock_future.get.side_effect = KafkaTimeoutError("Timeout")

    mock_producer_instance = Mock()
    mock_producer_instance.bootstrap_connected.return_value = True
    mock_producer_instance.send.return_value = mock_future
    mock_producer_class.return_value = mock_producer_instance

    producer = KafkaProducer(config)
    producer.connect()

    with pytest.raises(ServiceError) as exc_info:
        producer.send_message("test_topic", value="value1")

    assert exc_info.value.code == "INFRA-KAFKA-TIMEOUT-ERROR"


@patch("core.infrastructure.kafka_client.BaseProducer")
def test_send_message_error(mock_producer_class, config):
    """Test message send error"""
    mock_future = Mock()
    mock_future.get.side_effect = KafkaError("Send failed")

    mock_producer_instance = Mock()
    mock_producer_instance.bootstrap_connected.return_value = True
    mock_producer_instance.send.return_value = mock_future
    mock_producer_class.return_value = mock_producer_instance

    producer = KafkaProducer(config)
    producer.connect()

    with pytest.raises(ServiceError) as exc_info:
        producer.send_message("test_topic", value="value1")

    assert exc_info.value.code == "INFRA-KAFKA-SEND-ERROR"


@patch("core.infrastructure.kafka_client.BaseProducer")
def test_producer_health_check_success(mock_producer_class, config):
    """Test successful producer health check"""
    mock_producer_instance = Mock()
    mock_producer_instance.bootstrap_connected.return_value = True
    mock_producer_class.return_value = mock_producer_instance

    producer = KafkaProducer(config)
    producer.connect()

    assert producer.health() is True


@patch("core.infrastructure.kafka_client.BaseProducer")
def test_producer_health_check_failure(mock_producer_class, config):
    """Test failed producer health check"""
    mock_producer_instance = Mock()
    mock_producer_instance.bootstrap_connected.side_effect = [
        True,
        Exception("Health check failed"),
    ]
    mock_producer_class.return_value = mock_producer_instance

    producer = KafkaProducer(config)
    producer.connect()

    assert producer.health() is False


def test_producer_health_not_connected(config):
    """Test producer health without connection"""
    producer = KafkaProducer(config)
    assert producer.health() is False


@patch("core.infrastructure.kafka_client.BaseProducer")
def test_producer_close(mock_producer_class, config):
    """Test closing producer"""
    mock_producer_instance = Mock()
    mock_producer_instance.bootstrap_connected.return_value = True
    mock_producer_class.return_value = mock_producer_instance

    producer = KafkaProducer(config)
    producer.connect()
    producer.close()

    mock_producer_instance.flush.assert_called_once()
    mock_producer_instance.close.assert_called_once()


def test_consumer_initialization(config):
    """Test consumer initializes with valid config"""
    consumer = KafkaConsumer(config, group_id="test-group")
    assert consumer.config == config
    assert consumer.group_id == "test-group"
    assert consumer.consumer is None


@patch("core.infrastructure.kafka_client.BaseConsumer")
def test_consumer_connect_success(mock_consumer_class, config):
    """Test successful consumer connection"""
    mock_consumer_instance = Mock()
    mock_consumer_class.return_value = mock_consumer_instance

    consumer = KafkaConsumer(config, group_id="test-group")
    consumer.connect(["test_topic"])

    assert consumer.consumer is not None
    mock_consumer_class.assert_called_once()


@patch("core.infrastructure.kafka_client.BaseConsumer")
def test_consumer_connect_failure(mock_consumer_class, config):
    """Test consumer connection failure"""
    mock_consumer_class.side_effect = KafkaError("Connection failed")

    consumer = KafkaConsumer(config, group_id="test-group")

    with pytest.raises(ServiceError) as exc_info:
        consumer.connect(["test_topic"])

    assert exc_info.value.code == "INFRA-KAFKA-CONNECTION-ERROR"


def test_consume_not_connected(config):
    """Test consume without connection raises error"""
    consumer = KafkaConsumer(config, group_id="test-group")

    with pytest.raises(ServiceError) as exc_info:
        consumer.consume(lambda msg: None)

    assert exc_info.value.code == "INFRA-KAFKA-CONSUMER-ERROR"


@patch("core.infrastructure.kafka_client.BaseConsumer")
def test_consume_success(mock_consumer_class, config):
    """Test successful message consumption"""
    # Create mock messages
    mock_message1 = Mock()
    mock_message1.topic = "test_topic"
    mock_message1.partition = 0
    mock_message1.offset = 100
    mock_message1.key = b"key1"
    mock_message1.value = {"data": "value1"}

    mock_consumer_instance = Mock()
    mock_consumer_instance.__iter__ = Mock(return_value=iter([mock_message1]))
    mock_consumer_instance.subscription.return_value = ["test_topic"]
    mock_consumer_instance.assignment.return_value = []
    mock_consumer_class.return_value = mock_consumer_instance

    consumer = KafkaConsumer(config, group_id="test-group")
    consumer.connect(["test_topic"])

    # Handler to collect messages
    messages = []

    def handler(msg):
        messages.append(msg)
        # Stop after one message for testing
        raise StopIteration()

    try:
        consumer.consume(handler)
    except StopIteration:
        pass

    assert len(messages) == 1
    assert messages[0] == {"data": "value1"}


@patch("core.infrastructure.kafka_client.BaseConsumer")
def test_consumer_health_check_success(mock_consumer_class, config):
    """Test successful consumer health check"""
    mock_consumer_instance = Mock()
    mock_consumer_instance.bootstrap_connected.return_value = True
    mock_consumer_class.return_value = mock_consumer_instance

    consumer = KafkaConsumer(config, group_id="test-group")
    consumer.connect(["test_topic"])

    assert consumer.health() is True


def test_consumer_health_not_connected(config):
    """Test consumer health without connection"""
    consumer = KafkaConsumer(config, group_id="test-group")
    assert consumer.health() is False


@patch("core.infrastructure.kafka_client.BaseConsumer")
def test_consumer_close(mock_consumer_class, config):
    """Test closing consumer"""
    mock_consumer_instance = Mock()
    mock_consumer_class.return_value = mock_consumer_instance

    consumer = KafkaConsumer(config, group_id="test-group")
    consumer.connect(["test_topic"])
    consumer.close()

    mock_consumer_instance.close.assert_called_once()


@patch("core.infrastructure.kafka_client.BaseProducer")
def test_producer_circuit_breaker_integration(mock_producer_class, config):
    """Test circuit breaker triggers after failures"""
    mock_future = Mock()
    mock_future.get.side_effect = KafkaError("Simulated error")

    mock_producer_instance = Mock()
    mock_producer_instance.bootstrap_connected.return_value = True
    mock_producer_instance.send.return_value = mock_future
    mock_producer_class.return_value = mock_producer_instance

    producer = KafkaProducer(config)
    producer.connect()

    # Trigger multiple failures
    for _ in range(6):
        try:
            producer.send_message("test_topic", value="test")
        except:
            pass

    # Circuit should be open
    assert producer.circuit_breaker.state == "open"
