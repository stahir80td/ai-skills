using AiPatterns.Domain.Models;

namespace AiPatterns.Domain.Interfaces;

/// <summary>
/// Event publisher interface demonstrating Kafka via Core.Infrastructure
/// </summary>
public interface IEventPublisher
{
    Task PublishOrderEventAsync(OrderEvent orderEvent);
    Task PublishUserEventAsync(UserEvent userEvent);
    Task PublishTelemetryEventAsync(TelemetryEvent telemetryEvent);
    Task PublishSystemEventAsync(SystemEvent systemEvent);
    Task PublishBatchEventsAsync<T>(IEnumerable<T> events, string topic) where T : class;
}

/// <summary>
/// Event consumer interface demonstrating Kafka via Core.Infrastructure  
/// </summary>
public interface IEventConsumer
{
    Task StartConsumingAsync(string[] topics, CancellationToken cancellationToken);
    Task StopConsumingAsync();
    event Func<OrderEvent, Task> OrderEventReceived;
    event Func<UserEvent, Task> UserEventReceived;
    event Func<TelemetryEvent, Task> TelemetryEventReceived;
    event Func<SystemEvent, Task> SystemEventReceived;
}