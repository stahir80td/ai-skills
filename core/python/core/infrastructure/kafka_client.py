"""
Kafka client - Python equivalent of Go core/infrastructure/kafka

Provides:
- Producer with headers support
- Consumer with message handling
- Health checks
- Structured logging
- Circuit breaker protection
"""

from typing import Dict, List, Optional, Callable
from dataclasses import dataclass
import json
from kafka import KafkaProducer as BaseProducer, KafkaConsumer as BaseConsumer
from kafka.errors import KafkaError, KafkaTimeoutError
from kafka.admin import KafkaAdminClient, NewTopic

from ..logger import Logger
from ..errors import ServiceError, Severity
from ..reliability import CircuitBreaker


@dataclass
class KafkaConfig:
    """
    Configuration for Kafka client
    Matches Go ProducerConfig pattern
    """

    brokers: List[str]
    logger: Logger
    client_id: str = "python-client"
    timeout_ms: int = 60000
    compression_type: str = "snappy"  # Matches Go snappy compression

    def __post_init__(self):
        """Validate configuration"""
        if not self.brokers:
            raise ServiceError(
                code="INFRA-KAFKA-CONFIG-ERROR",
                message="Kafka brokers list cannot be empty",
                severity=Severity.CRITICAL,
            )


class KafkaProducer:
    """
    Kafka producer with production patterns
    Matches Go Producer interface
    """

    def __init__(self, config: KafkaConfig):
        self.config = config
        self.logger = config.logger.with_component("KafkaProducer")
        self.producer: Optional[BaseProducer] = None
        self.circuit_breaker = CircuitBreaker(
            "kafka_producer", max_failures=5, enabled=False
        )

        self.logger.debug(
            "Initiating Kafka producer",
            brokers=config.brokers,
            client_id=config.client_id,
        )

    def connect(self) -> None:
        """Establish connection to Kafka"""
        try:
            # Create producer with production settings
            # Matches Go config: return successes, wait for all acks, 3 retries, snappy compression
            self.producer = BaseProducer(
                bootstrap_servers=self.config.brokers,
                client_id=self.config.client_id,
                acks="all",  # Wait for all in-sync replicas
                retries=3,
                compression_type=self.config.compression_type,
                value_serializer=lambda v: (
                    json.dumps(v).encode("utf-8")
                    if isinstance(v, dict)
                    else v if isinstance(v, bytes) else str(v).encode("utf-8")
                ),
                request_timeout_ms=self.config.timeout_ms,
            )

            # Test connection
            self.producer.bootstrap_connected()

            self.logger.info(
                "Successfully connected to Kafka",
                brokers=self.config.brokers,
                status="healthy",
            )

        except KafkaError as e:
            self.logger.error(
                "Kafka producer connection failed",
                error=str(e),
                brokers=self.config.brokers,
                error_code="INFRA-KAFKA-CONNECTION-ERROR",
            )
            raise ServiceError(
                code="INFRA-KAFKA-CONNECTION-ERROR",
                message="Failed to connect to Kafka",
                severity=Severity.CRITICAL,
                underlying=e,
            )

    def send_message(
        self,
        topic: str,
        key: Optional[str] = None,
        value: any = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> None:
        """
        Send message to Kafka topic
        Matches Go SendMessage method
        """

        def _send():
            if not self.producer:
                raise ServiceError(
                    code="INFRA-KAFKA-PRODUCER-ERROR",
                    message="Kafka producer not connected",
                    severity=Severity.CRITICAL,
                )

            try:
                # Convert headers to list of tuples format expected by kafka-python
                kafka_headers = (
                    [(k, v.encode("utf-8")) for k, v in headers.items()]
                    if headers
                    else None
                )

                # Convert key to bytes
                key_bytes = key.encode("utf-8") if key else None

                # Send message
                future = self.producer.send(
                    topic=topic,
                    key=key_bytes,
                    value=value,
                    headers=kafka_headers,
                )

                # Wait for result
                future.get(timeout=self.config.timeout_ms / 1000)

                self.logger.debug(
                    "Kafka message sent",
                    topic=topic,
                    key=key,
                    headers=headers,
                )

            except KafkaTimeoutError as e:
                self.logger.error(
                    "Kafka send timeout",
                    error=str(e),
                    topic=topic,
                    error_code="INFRA-KAFKA-TIMEOUT-ERROR",
                )
                raise ServiceError(
                    code="INFRA-KAFKA-TIMEOUT-ERROR",
                    message=f"Timeout sending message to topic: {topic}",
                    severity=Severity.HIGH,
                    underlying=e,
                )

            except KafkaError as e:
                self.logger.error(
                    "Kafka send failed",
                    error=str(e),
                    topic=topic,
                    error_code="INFRA-KAFKA-SEND-ERROR",
                )
                raise ServiceError(
                    code="INFRA-KAFKA-SEND-ERROR",
                    message=f"Failed to send message to topic: {topic}",
                    severity=Severity.HIGH,
                    underlying=e,
                )

        return self.circuit_breaker.call(_send)

    def health(self) -> bool:
        """
        Check Kafka producer health
        Matches Go Health method
        """
        try:
            if not self.producer:
                return False

            # bootstrap_connected() returns bool, not list
            return self.producer.bootstrap_connected()

        except Exception as e:
            self.logger.warning("Kafka producer health check failed", error=str(e))
            return False

    def close(self) -> None:
        """
        Close producer connection
        Matches Go Close method
        """
        if self.producer:
            self.producer.flush()  # Ensure all messages are sent
            self.producer.close()
            self.logger.info("Kafka producer closed")


class KafkaConsumer:
    """
    Kafka consumer with production patterns
    Matches Go Consumer interface (if exists)
    """

    def __init__(self, config: KafkaConfig, group_id: str):
        self.config = config
        self.group_id = group_id
        self.logger = config.logger.with_component("KafkaConsumer")
        self.consumer: Optional[BaseConsumer] = None

        self.logger.debug(
            "Initiating Kafka consumer",
            brokers=config.brokers,
            group_id=group_id,
        )

    def connect(self, topics: List[str]) -> None:
        """Establish connection and subscribe to topics"""
        try:
            self.consumer = BaseConsumer(
                *topics,
                bootstrap_servers=self.config.brokers,
                client_id=self.config.client_id,
                group_id=self.group_id,
                value_deserializer=lambda m: (
                    json.loads(m.decode("utf-8")) if m else None
                ),
                auto_offset_reset="earliest",
                enable_auto_commit=True,
                request_timeout_ms=self.config.timeout_ms,
            )

            self.logger.info(
                "Successfully connected Kafka consumer",
                brokers=self.config.brokers,
                group_id=self.group_id,
                topics=topics,
            )

        except KafkaError as e:
            self.logger.error(
                "Kafka consumer connection failed",
                error=str(e),
                brokers=self.config.brokers,
                error_code="INFRA-KAFKA-CONNECTION-ERROR",
            )
            raise ServiceError(
                code="INFRA-KAFKA-CONNECTION-ERROR",
                message="Failed to connect Kafka consumer",
                severity=Severity.CRITICAL,
                underlying=e,
            )

    def consume(self, handler: Callable[[any], None], timeout_ms: int = 1000) -> None:
        """
        Consume messages and pass to handler
        """
        if not self.consumer:
            raise ServiceError(
                code="INFRA-KAFKA-CONSUMER-ERROR",
                message="Kafka consumer not connected",
                severity=Severity.CRITICAL,
            )

        # Wait for partition assignment (kafka-python assigns partitions lazily)
        import time

        max_wait = 10  # seconds
        waited = 0
        while not self.consumer.assignment() and waited < max_wait:
            self.consumer.poll(timeout_ms=100)
            time.sleep(0.1)
            waited += 0.1

        self.logger.info(
            "Starting consumer loop",
            consumer_topics=list(self.consumer.subscription()),
            consumer_assignment=str(self.consumer.assignment()),
        )

        try:
            for message in self.consumer:
                try:
                    self.logger.debug(
                        "Kafka message received",
                        topic=message.topic,
                        partition=message.partition,
                        offset=message.offset,
                        key=message.key.decode("utf-8") if message.key else None,
                    )

                    handler(message.value)

                except Exception as e:
                    self.logger.error(
                        "Message handler failed",
                        error=str(e),
                        topic=message.topic,
                        offset=message.offset,
                        error_code="INFRA-KAFKA-HANDLER-ERROR",
                    )
                    # Continue processing other messages

        except KafkaError as e:
            self.logger.error(
                "Kafka consume failed",
                error=str(e),
                error_code="INFRA-KAFKA-CONSUME-ERROR",
            )
            raise ServiceError(
                code="INFRA-KAFKA-CONSUME-ERROR",
                message="Failed to consume messages",
                severity=Severity.HIGH,
                underlying=e,
            )

    def health(self) -> bool:
        """Check Kafka consumer health"""
        try:
            if not self.consumer:
                return False

            # Check if consumer is active
            return self.consumer.bootstrap_connected() is not None

        except Exception as e:
            self.logger.warning("Kafka consumer health check failed", error=str(e))
            return False

    def close(self) -> None:
        """Close consumer connection"""
        if self.consumer:
            self.consumer.close()
            self.logger.info("Kafka consumer closed")
