package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.uber.org/zap"
)

// ConsumerConfig for Kafka consumer
type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
	Logger  *logger.Logger
}

// MessageHandler processes consumed messages
type MessageHandler func(ctx context.Context, msg *sarama.ConsumerMessage) error

// Consumer interface for consuming messages
type Consumer interface {
	Start(ctx context.Context, handler MessageHandler) error
	Close(ctx context.Context) error
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler MessageHandler
	logger  *logger.Logger
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-claim.Messages():
			if msg == nil {
				continue
			}

			if h.logger != nil {
				h.logger.WithComponent("KafkaConsumer").Info("Message received",
					zap.String("topic", msg.Topic),
					zap.Int32("partition", msg.Partition),
					zap.Int64("offset", msg.Offset),
					zap.Int("value_size", len(msg.Value)))
			}

			if err := h.handler(ctx, msg); err != nil {
				if h.logger != nil {
					h.logger.WithComponent("KafkaConsumer").Error("Message processing failed",
						zap.Error(err),
						zap.String("topic", msg.Topic),
						zap.Int32("partition", msg.Partition),
						zap.Int64("offset", msg.Offset))
				}
				// Continue processing even if one message fails
			} else {
				if h.logger != nil {
					h.logger.WithComponent("KafkaConsumer").Info("Message processed successfully",
						zap.String("topic", msg.Topic),
						zap.Int32("partition", msg.Partition),
						zap.Int64("offset", msg.Offset))
				}
			}

			// Mark message as processed
			session.MarkMessage(msg, "")
		}
	}
}

// consumer implements the Consumer interface
type consumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	logger        *logger.Logger
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg ConsumerConfig) (Consumer, error) {
	if len(cfg.Brokers) == 0 {
		err := fmt.Errorf("Kafka brokers list cannot be empty")
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("KafkaConsumer").Error("Invalid configuration - no brokers",
				zap.Error(err),
				zap.String("error_code", "INFRA-KAFKA-CONFIG-ERROR"))
		}
		return nil, err
	}

	if cfg.Topic == "" {
		err := fmt.Errorf("Kafka topic cannot be empty")
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("KafkaConsumer").Error("Invalid configuration - no topic",
				zap.Error(err),
				zap.String("error_code", "INFRA-KAFKA-CONFIG-ERROR"))
		}
		return nil, err
	}

	if cfg.GroupID == "" {
		err := fmt.Errorf("Kafka group ID cannot be empty")
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("KafkaConsumer").Error("Invalid configuration - no group ID",
				zap.Error(err),
				zap.String("error_code", "INFRA-KAFKA-CONFIG-ERROR"))
		}
		return nil, err
	}

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("KafkaConsumer").Info("Initializing Kafka consumer",
			zap.Strings("brokers", cfg.Brokers),
			zap.String("topic", cfg.Topic),
			zap.String("group_id", cfg.GroupID))
	}

	// Configure Sarama
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Version = sarama.V2_8_0_0

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.WithComponent("KafkaConsumer").Error("Failed to create Kafka consumer",
				zap.Error(err),
				zap.Strings("brokers", cfg.Brokers),
				zap.String("error_code", "INFRA-KAFKA-CONNECTION-ERROR"))
		}
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	if cfg.Logger != nil {
		cfg.Logger.WithComponent("KafkaConsumer").Info("Successfully created Kafka consumer",
			zap.Strings("brokers", cfg.Brokers),
			zap.String("topic", cfg.Topic),
			zap.String("group_id", cfg.GroupID),
			zap.String("status", "healthy"))
	}

	return &consumer{
		consumerGroup: consumerGroup,
		topic:         cfg.Topic,
		logger:        cfg.Logger,
	}, nil
}

// Start begins consuming messages
func (c *consumer) Start(ctx context.Context, handler MessageHandler) error {
	if c.consumerGroup == nil {
		return fmt.Errorf("consumer group not initialized")
	}

	// Create handler
	h := &consumerGroupHandler{
		handler: handler,
		logger:  c.logger,
	}

	if c.logger != nil {
		c.logger.WithComponent("KafkaConsumer").Info("Starting consumer", zap.String("topic", c.topic))
	}

	// Consume messages in a loop
	for {
		select {
		case <-ctx.Done():
			if c.logger != nil {
				c.logger.WithComponent("KafkaConsumer").Info("Consumer context cancelled")
			}
			return ctx.Err()
		default:
			if err := c.consumerGroup.Consume(ctx, []string{c.topic}, h); err != nil {
				if c.logger != nil {
					c.logger.WithComponent("KafkaConsumer").Error("Consumer error", zap.Error(err))
				}
				return fmt.Errorf("consumer error: %w", err)
			}

			// Check if context was cancelled
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}
	}
}

// Close closes the Kafka consumer
func (c *consumer) Close(ctx context.Context) error {
	if c.consumerGroup == nil {
		return nil
	}

	if err := c.consumerGroup.Close(); err != nil {
		if c.logger != nil {
			c.logger.WithComponent("KafkaConsumer").Error("Failed to close consumer", zap.Error(err))
		}
		return err
	}

	if c.logger != nil {
		c.logger.WithComponent("KafkaConsumer").Info("Kafka consumer closed")
	}

	return nil
}
