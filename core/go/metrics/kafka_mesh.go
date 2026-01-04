package metrics

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// KafkaMeshMetrics tracks Kafka producer metrics for service mesh observability
// This captures async message flow between services via Kafka topics
type KafkaMeshMetrics struct {
	sourceService     string
	messagesPublished *prometheus.CounterVec
	publishDuration   *prometheus.HistogramVec
	publishErrors     *prometheus.CounterVec
}

var (
	kafkaMessagesPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_published_total",
			Help: "Total messages published to Kafka topics by source service",
		},
		[]string{"source_service", "topic", "status"},
	)

	kafkaPublishDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_publish_duration_seconds",
			Help:    "Duration of Kafka publish operations",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"source_service", "topic"},
	)

	kafkaPublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_publish_errors_total",
			Help: "Total Kafka publish errors by source service",
		},
		[]string{"source_service", "topic", "error_type"},
	)
)

// NewKafkaMeshMetrics creates a new Kafka mesh metrics tracker
func NewKafkaMeshMetrics(sourceService string) *KafkaMeshMetrics {
	return &KafkaMeshMetrics{
		sourceService:     sourceService,
		messagesPublished: kafkaMessagesPublished,
		publishDuration:   kafkaPublishDuration,
		publishErrors:     kafkaPublishErrors,
	}
}

// TrackPublish records metrics for a Kafka publish operation
func (k *KafkaMeshMetrics) TrackPublish(topic string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
		errorType := "unknown"

		// Classify Kafka errors
		if err == sarama.ErrOutOfBrokers {
			errorType = "out_of_brokers"
		} else if err == sarama.ErrNotConnected {
			errorType = "not_connected"
		} else if err == sarama.ErrInsufficientData {
			errorType = "insufficient_data"
		} else if err == sarama.ErrShuttingDown {
			errorType = "shutting_down"
		} else if err == sarama.ErrMessageSizeTooLarge {
			errorType = "message_too_large"
		} else if err == sarama.ErrNotLeaderForPartition {
			errorType = "not_leader"
		} else if err == sarama.ErrRequestTimedOut {
			errorType = "timeout"
		}

		k.publishErrors.WithLabelValues(k.sourceService, topic, errorType).Inc()
	}

	k.messagesPublished.WithLabelValues(k.sourceService, topic, status).Inc()
	k.publishDuration.WithLabelValues(k.sourceService, topic).Observe(duration.Seconds())
}

// WrapProducer wraps a Sarama SyncProducer with automatic metrics tracking
type TrackedSyncProducer struct {
	producer sarama.SyncProducer
	metrics  *KafkaMeshMetrics
}

// NewTrackedSyncProducer wraps a Kafka producer with metrics tracking
func NewTrackedSyncProducer(sourceService string, producer sarama.SyncProducer) *TrackedSyncProducer {
	return &TrackedSyncProducer{
		producer: producer,
		metrics:  NewKafkaMeshMetrics(sourceService),
	}
}

// SendMessage sends a message with automatic metrics tracking
func (t *TrackedSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	start := time.Now()
	partition, offset, err = t.producer.SendMessage(msg)
	duration := time.Since(start)

	t.metrics.TrackPublish(msg.Topic, duration, err)

	return partition, offset, err
}

// SendMessages sends multiple messages with automatic metrics tracking
func (t *TrackedSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	start := time.Now()
	err := t.producer.SendMessages(msgs)
	duration := time.Since(start)

	// Track each message
	for _, msg := range msgs {
		t.metrics.TrackPublish(msg.Topic, duration, err)
	}

	return err
}

// Close closes the underlying producer
func (t *TrackedSyncProducer) Close() error {
	return t.producer.Close()
}

// IsTransactional returns whether the producer is transactional
func (t *TrackedSyncProducer) IsTransactional() bool {
	return t.producer.IsTransactional()
}

// TxnStatus returns the current transaction status
func (t *TrackedSyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	return t.producer.TxnStatus()
}

// BeginTxn begins a transaction
func (t *TrackedSyncProducer) BeginTxn() error {
	return t.producer.BeginTxn()
}

// CommitTxn commits a transaction
func (t *TrackedSyncProducer) CommitTxn() error {
	return t.producer.CommitTxn()
}

// AbortTxn aborts a transaction
func (t *TrackedSyncProducer) AbortTxn() error {
	return t.producer.AbortTxn()
}

// AddOffsetsToTxn adds offsets to the current transaction
func (t *TrackedSyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupID string) error {
	return t.producer.AddOffsetsToTxn(offsets, groupID)
}

// AddMessageToTxn adds a message to the current transaction (this method doesn't exist in sarama.SyncProducer, keeping for potential future use)
func (t *TrackedSyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupID string, metadata *string) error {
	return t.producer.AddMessageToTxn(msg, groupID, metadata)
}
