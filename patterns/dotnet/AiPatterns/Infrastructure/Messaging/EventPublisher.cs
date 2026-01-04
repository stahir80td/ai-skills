using Core.Infrastructure.Kafka;
using Core.Logger;
using AiPatterns.Domain.Interfaces;
using AiPatterns.Domain.Models;
using System.Text.Json;

namespace AiPatterns.Infrastructure.Messaging;

/// <summary>
/// Event publisher using Core.Infrastructure.Kafka - demonstrates event-driven patterns
/// </summary>
public class EventPublisher : IEventPublisher
{
    private readonly IKafkaProducer _kafkaProducer;
    private readonly ServiceLogger _logger;

    public EventPublisher(IKafkaProducer kafkaProducer, ServiceLogger logger)
    {
        _kafkaProducer = kafkaProducer;
        _logger = logger;
    }

    public async Task PublishOrderEventAsync(OrderEvent orderEvent)
    {
        var contextLogger = _logger.WithContext(component: "EventPublisher.PublishOrderEvent");

        try
        {
            var topic = $"patterns.orders.{orderEvent.EventType.ToLowerInvariant()}";
            var message = new KafkaMessage<string>
            {
                Key = orderEvent.OrderId.ToString(),
                Value = JsonSerializer.Serialize(orderEvent),
                Headers = new Dictionary<string, string>
                {
                    { "event_type", orderEvent.EventType },
                    { "timestamp", orderEvent.Timestamp.ToString("O") },
                    { "correlation_id", orderEvent.CorrelationId ?? Guid.NewGuid().ToString() }
                }
            };

            await _kafkaProducer.ProduceAsync(topic, message);
            contextLogger.Information("Order event published to Kafka: {EventType}, OrderId: {OrderId}", orderEvent.EventType, orderEvent.OrderId);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Failed to publish order event to Kafka: {EventType}", orderEvent.EventType);
            throw;
        }
    }

    public async Task PublishUserEventAsync(UserEvent userEvent)
    {
        var contextLogger = _logger.WithContext(component: "EventPublisher.PublishUserEvent");

        try
        {
            var topic = $"patterns.users.{userEvent.EventType.ToLowerInvariant()}";
            var message = new KafkaMessage<string>
            {
                Key = userEvent.UserId.ToString(),
                Value = JsonSerializer.Serialize(userEvent),
                Headers = new Dictionary<string, string>
                {
                    { "event_type", userEvent.EventType },
                    { "timestamp", userEvent.Timestamp.ToString("O") },
                    { "correlation_id", userEvent.CorrelationId ?? Guid.NewGuid().ToString() }
                }
            };

            await _kafkaProducer.ProduceAsync(topic, message);
            contextLogger.Information("User event published to Kafka: {EventType}, UserId: {UserId}", userEvent.EventType, userEvent.UserId);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Failed to publish user event to Kafka: {EventType}", userEvent.EventType);
            throw;
        }
    }

    public async Task PublishTelemetryEventAsync(TelemetryEvent telemetryEvent)
    {
        var contextLogger = _logger.WithContext(component: "EventPublisher.PublishTelemetryEvent");

        try
        {
            var topic = $"patterns.telemetry.{telemetryEvent.EventType.ToLowerInvariant()}";
            var message = new KafkaMessage<string>
            {
                Key = telemetryEvent.DeviceId,
                Value = JsonSerializer.Serialize(telemetryEvent),
                Headers = new Dictionary<string, string>
                {
                    { "event_type", telemetryEvent.EventType },
                    { "timestamp", telemetryEvent.Timestamp.ToString("O") },
                    { "device_id", telemetryEvent.DeviceId },
                    { "correlation_id", telemetryEvent.CorrelationId ?? Guid.NewGuid().ToString() }
                }
            };

            await _kafkaProducer.ProduceAsync(topic, message);
            contextLogger.Information("Telemetry event published to Kafka: {EventType}, DeviceId: {DeviceId}", telemetryEvent.EventType, telemetryEvent.DeviceId);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Failed to publish telemetry event to Kafka: {EventType}", telemetryEvent.EventType);
            throw;
        }
    }

    public async Task PublishSystemEventAsync(SystemEvent systemEvent)
    {
        var contextLogger = _logger.WithContext(component: "EventPublisher.PublishSystemEvent");

        try
        {
            var topic = $"patterns.system.{systemEvent.EventType.ToLowerInvariant()}";
            var message = new KafkaMessage<string>
            {
                Key = systemEvent.Source,
                Value = JsonSerializer.Serialize(systemEvent),
                Headers = new Dictionary<string, string>
                {
                    { "event_type", systemEvent.EventType },
                    { "timestamp", systemEvent.Timestamp.ToString("O") },
                    { "source", systemEvent.Source },
                    { "correlation_id", systemEvent.CorrelationId ?? Guid.NewGuid().ToString() }
                }
            };

            await _kafkaProducer.ProduceAsync(topic, message);
            contextLogger.Information("System event published to Kafka: {EventType}, Source: {Source}", systemEvent.EventType, systemEvent.Source);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Failed to publish system event to Kafka: {EventType}", systemEvent.EventType);
            throw;
        }
    }

    public async Task PublishBatchEventsAsync<T>(IEnumerable<T> events, string topic) where T : class
    {
        var contextLogger = _logger.WithContext(component: "EventPublisher.PublishBatchEvents");

        try
        {
            var eventList = events.ToList();
            contextLogger.Information("Publishing batch of {Count} events to topic: {Topic}", eventList.Count, topic);

            foreach (var evt in eventList)
            {
                var message = new KafkaMessage<string>
                {
                    Key = Guid.NewGuid().ToString(),
                    Value = JsonSerializer.Serialize(evt),
                    Headers = new Dictionary<string, string>
                    {
                        { "event_type", typeof(T).Name },
                        { "timestamp", DateTime.UtcNow.ToString("O") }
                    }
                };

                await _kafkaProducer.ProduceAsync(topic, message);
            }

            contextLogger.Information("Batch of {Count} events published to topic: {Topic}", eventList.Count, topic);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Failed to publish batch events to topic: {Topic}", topic);
            throw;
        }
    }
}
