using Core.Logger;
using AiPatterns.Domain.Interfaces;
using AiPatterns.Domain.Models;
using System.Text.Json;

namespace AiPatterns.Infrastructure.Messaging;

/// <summary>
/// Event consumer - Note: Core.Infrastructure does not include a KafkaConsumer.
/// This implementation provides event handler registration and manual message processing.
/// In production, use Confluent.Kafka directly for consumer implementation.
/// </summary>
public class EventConsumer : IEventConsumer
{
    private readonly ServiceLogger _logger;
    private CancellationTokenSource? _cts;
    private bool _isConsuming;

    public event Func<OrderEvent, Task>? OrderEventReceived;
    public event Func<UserEvent, Task>? UserEventReceived;
    public event Func<TelemetryEvent, Task>? TelemetryEventReceived;
    public event Func<SystemEvent, Task>? SystemEventReceived;

    public EventConsumer(ServiceLogger logger)
    {
        _logger = logger;
    }

    public async Task StartConsumingAsync(string[] topics, CancellationToken cancellationToken)
    {
        var contextLogger = _logger.WithContext(component: "EventConsumer.StartConsuming");
        contextLogger.Information("Starting to consume from topics: {Topics}", string.Join(", ", topics));

        _cts = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
        _isConsuming = true;

        // Note: Core.Infrastructure doesn't provide a KafkaConsumer
        // In a real implementation, you would use Confluent.Kafka.Consumer here
        // This is a placeholder that simulates the consumer pattern

        contextLogger.Warning("KafkaConsumer not available in Core.Infrastructure. Use Confluent.Kafka directly for consuming.");
        
        // Keep the task alive until cancellation
        try
        {
            await Task.Delay(Timeout.Infinite, _cts.Token);
        }
        catch (OperationCanceledException)
        {
            contextLogger.Information("Consumer stopping due to cancellation");
        }
    }

    public async Task StopConsumingAsync()
    {
        var contextLogger = _logger.WithContext(component: "EventConsumer.StopConsuming");
        contextLogger.Information("Stopping consumer");

        _isConsuming = false;
        _cts?.Cancel();

        await Task.CompletedTask;
        contextLogger.Information("Consumer stopped");
    }

    /// <summary>
    /// Process a message manually - useful for testing or external integration
    /// </summary>
    public async Task ProcessMessageAsync(string topic, string messageJson)
    {
        var contextLogger = _logger.WithContext(component: "EventConsumer.ProcessMessage");
        contextLogger.Debug("Processing message from topic: {Topic}", topic);

        try
        {
            if (topic.Contains("orders") && OrderEventReceived != null)
            {
                var orderEvent = JsonSerializer.Deserialize<OrderEvent>(messageJson);
                if (orderEvent != null)
                {
                    await OrderEventReceived.Invoke(orderEvent);
                    contextLogger.Information("Order event processed: {EventType}", orderEvent.EventType);
                }
            }
            else if (topic.Contains("users") && UserEventReceived != null)
            {
                var userEvent = JsonSerializer.Deserialize<UserEvent>(messageJson);
                if (userEvent != null)
                {
                    await UserEventReceived.Invoke(userEvent);
                    contextLogger.Information("User event processed: {EventType}", userEvent.EventType);
                }
            }
            else if (topic.Contains("telemetry") && TelemetryEventReceived != null)
            {
                var telemetryEvent = JsonSerializer.Deserialize<TelemetryEvent>(messageJson);
                if (telemetryEvent != null)
                {
                    await TelemetryEventReceived.Invoke(telemetryEvent);
                    contextLogger.Information("Telemetry event processed: {EventType}", telemetryEvent.EventType);
                }
            }
            else if (topic.Contains("system") && SystemEventReceived != null)
            {
                var systemEvent = JsonSerializer.Deserialize<SystemEvent>(messageJson);
                if (systemEvent != null)
                {
                    await SystemEventReceived.Invoke(systemEvent);
                    contextLogger.Information("System event processed: {EventType}", systemEvent.EventType);
                }
            }
            else
            {
                contextLogger.Warning("No handler registered for topic: {Topic}", topic);
            }
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error processing message from topic: {Topic}", topic);
            throw;
        }
    }
}
