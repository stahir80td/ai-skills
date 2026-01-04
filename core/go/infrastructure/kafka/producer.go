package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.uber.org/zap"
)

// Config for Kafka producer
type ProducerConfig struct {
	Brokers []string
	Logger  *logger.Logger
}

// Producer interface for sending messages
type Producer interface {
	SendMessage(ctx context.Context, topic, key string, value []byte, headers map[string]string) error
	Close(ctx context.Context) error
	Health(ctx context.Context) error
}

// syncProducer implements the Producer interface
type syncProducer struct {
	producer sarama.SyncProducer
	logger   *logger.Logger
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg ProducerConfig) (Producer, error) {
	if len(cfg.Brokers) == 0 {
		err := fmt.Errorf("Kafka brokers list cannot be empty")
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("KafkaProducer").Error("Invalid configuration - no brokers",
				zap.Error(err),
				zap.String("error_code", "INFRA-KAFKA-CONFIG-ERROR"))
		}
		return nil, err
	}

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("KafkaProducer").Info("Initializing Kafka producer",
			zap.Strings("brokers", cfg.Brokers),
			zap.Int("broker_count", len(cfg.Brokers)))
	}

	// Configure Sarama
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Compression = sarama.CompressionSnappy

	// Create sync producer
	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("KafkaProducer").Error("Failed to create Kafka producer",
				zap.Error(err),
				zap.Strings("brokers", cfg.Brokers),
				zap.String("error_code", "INFRA-KAFKA-CONNECTION-ERROR"))
		}
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("KafkaProducer").Info("Successfully created Kafka producer",
			zap.Strings("brokers", cfg.Brokers),
			zap.String("status", "healthy"))
	}

	return &syncProducer{
		producer: producer,
		logger:   cfg.Logger,
	}, nil
}

// SendMessage sends a message to Kafka
func (p *syncProducer) SendMessage(ctx context.Context, topic, key string, value []byte, headers map[string]string) error {
	// Build message headers
	var recordHeaders []sarama.RecordHeader
	for k, v := range headers {
		recordHeaders = append(recordHeaders, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(key),
		Value:   sarama.ByteEncoder(value),
		Headers: recordHeaders,
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		if p.logger != nil {
			p.logger.WithComponent("KafkaProducer").Error("Failed to send message",
				zap.Error(err),
				zap.String("topic", topic),
				zap.String("key", key))
		}
		return fmt.Errorf("failed to send message: %w", err)
	}

	if p.logger != nil {
		p.logger.WithComponent("KafkaProducer").Debug("Message sent successfully",
			zap.String("topic", topic),
			zap.Int32("partition", partition),
			zap.Int64("offset", offset))
	}

	return nil
}

// Health checks if Kafka is healthy
func (p *syncProducer) Health(ctx context.Context) error {
	// Sarama doesn't provide a direct health check, but we can verify the producer is not nil
	if p.producer == nil {
		return fmt.Errorf("producer not initialized")
	}
	return nil
}

// Close closes the Kafka producer
func (p *syncProducer) Close(ctx context.Context) error {
	if p.producer == nil {
		return nil
	}

	if err := p.producer.Close(); err != nil {
		if p.logger != nil {
			p.logger.WithComponent("KafkaProducer").Error("Failed to close producer", zap.Error(err))
		}
		return err
	}

	if p.logger != nil {
		p.logger.WithComponent("KafkaProducer").Info("Kafka producer closed")
	}

	return nil
}
