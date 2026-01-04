using Core.Config;
using Core.Logger;
using Confluent.Kafka;

namespace Core.Infrastructure.Kafka;

/// <summary>
/// Kafka configuration
/// </summary>
public class KafkaConfig : IValidatable
{
    /// <summary>Bootstrap servers</summary>
    public string BootstrapServers { get; set; } = "localhost:9092";
    
    /// <summary>Security protocol</summary>
    public SecurityProtocol SecurityProtocol { get; set; } = SecurityProtocol.Plaintext;
    
    /// <summary>Enable idempotent producer</summary>
    public bool EnableIdempotence { get; set; } = true;
    
    /// <summary>Message timeout in milliseconds</summary>
    public int MessageTimeoutMs { get; set; } = 30000;
    
    /// <summary>Request timeout in milliseconds (minimum 60s)</summary>
    public int RequestTimeoutMs { get; set; } = 60000;
    
    /// <summary>Session timeout in milliseconds (minimum 60s)</summary>
    public int SessionTimeoutMs { get; set; } = 60000;
    
    /// <summary>Heartbeat interval in milliseconds</summary>
    public int HeartbeatIntervalMs { get; set; } = 20000;
    
    /// <summary>Health check timeout in seconds</summary>
    public int HealthCheckTimeoutSeconds { get; set; } = 10;

    public ValidationResult Validate()
    {
        var errors = new List<string>();
        
        if (string.IsNullOrWhiteSpace(BootstrapServers))
            errors.Add("BootstrapServers is required");
            
        if (MessageTimeoutMs <= 0)
            errors.Add("MessageTimeoutMs must be positive");
            
        if (RequestTimeoutMs < 60000)
            errors.Add("RequestTimeoutMs must be at least 60 seconds");
            
        if (SessionTimeoutMs < 60000)
            errors.Add("SessionTimeoutMs must be at least 60 seconds");
        
        return errors.Any() ? ValidationResult.Failed(errors.ToArray()) : ValidationResult.Success();
    }
}

/// <summary>
/// Simple Kafka message wrapper
/// </summary>
public class KafkaMessage<T>
{
    /// <summary>Message key</summary>
    public string? Key { get; set; }
    
    /// <summary>Message value</summary>
    public T? Value { get; set; }
    
    /// <summary>Message headers</summary>
    public Dictionary<string, string> Headers { get; set; } = new();
}

/// <summary>
/// Kafka producer interface
/// </summary>
public interface IKafkaProducer
{
    /// <summary>Produce a message to a topic</summary>
    Task<DeliveryResult<string, string>> ProduceAsync(string topic, KafkaMessage<string> message, CancellationToken cancellationToken = default);
    
    /// <summary>Check producer health</summary>
    Task<bool> HealthAsync(CancellationToken cancellationToken = default);
    
    /// <summary>Dispose of the producer</summary>
    void Dispose();
}

/// <summary>
/// Simple Kafka producer implementation
/// </summary>
public class KafkaProducer : IKafkaProducer, IDisposable
{
    private readonly IProducer<string, string> _producer;
    private readonly ServiceLogger _logger;
    private readonly KafkaConfig _config;
    private readonly string _componentName = "kafka";
    private volatile bool _disposed;

    public KafkaProducer(KafkaConfig config, ServiceLogger logger)
    {
        _config = config ?? throw new ArgumentNullException(nameof(config));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
        
        var validation = config.Validate();
        if (!validation.IsValid)
        {
            _logger.Error("Invalid Kafka configuration", new { 
                errors = validation.Errors,
                errorCode = "INFRA-KAFKA-CONFIG-ERROR",
                component = _componentName
            });
            throw new InvalidOperationException($"Invalid configuration: {string.Join(", ", validation.Errors)}");
        }

        var producerConfig = new ProducerConfig
        {
            BootstrapServers = config.BootstrapServers,
            SecurityProtocol = config.SecurityProtocol,
            EnableIdempotence = config.EnableIdempotence,
            MessageTimeoutMs = config.MessageTimeoutMs,
            RequestTimeoutMs = config.RequestTimeoutMs
            // Note: SessionTimeoutMs and HeartbeatIntervalMs are consumer-specific
        };

        _producer = new ProducerBuilder<string, string>(producerConfig)
            .SetErrorHandler((_, error) => _logger.Warning("Kafka producer error", new { 
                error = error.Reason,
                component = _componentName,
                errorCode = "INFRA-KAFKA-PRODUCER-ERROR"
            }))
            .Build();

        _logger.Information("KafkaProducer initialized", new { 
            component = _componentName,
            bootstrapServers = config.BootstrapServers,
            securityProtocol = config.SecurityProtocol.ToString(),
            requestTimeout = config.RequestTimeoutMs,
            sessionTimeout = config.SessionTimeoutMs,
            status = "healthy"
        });
    }

    public async Task<DeliveryResult<string, string>> ProduceAsync(string topic, KafkaMessage<string> message, CancellationToken cancellationToken = default)
    {
        if (_disposed)
            throw new ObjectDisposedException(nameof(KafkaProducer));

        _logger.Information("Producing Kafka message", new { 
            component = _componentName,
            topic, 
            key = message.Key,
            hasHeaders = message.Headers.Any()
        });

        try
        {
            var kafkaMessage = new Message<string, string>
            {
                Key = message.Key,
                Value = message.Value
            };

            // Add headers
            if (message.Headers.Any())
            {
                kafkaMessage.Headers = new Headers();
                foreach (var header in message.Headers)
                {
                    kafkaMessage.Headers.Add(header.Key, System.Text.Encoding.UTF8.GetBytes(header.Value));
                }
            }

            var result = await _producer.ProduceAsync(topic, kafkaMessage, cancellationToken);
            
            _logger.Information("Kafka message produced successfully", new { 
                component = _componentName,
                topic, 
                key = message.Key,
                partition = result.Partition.Value,
                offset = result.Offset.Value,
                status = "success"
            });

            return result;
        }
        catch (Exception ex)
        {
            _logger.Warning("Failed to produce Kafka message", ex, new { 
                component = _componentName,
                topic, 
                key = message.Key,
                errorCode = "INFRA-KAFKA-PUBLISH-ERROR"
            });
            throw;
        }
    }

    public async Task<bool> HealthAsync(CancellationToken cancellationToken = default)
    {
        if (_disposed)
            return false;

        try
        {
            _logger.Debug("Performing Kafka health check", new { 
                component = _componentName 
            });
            
            // Use metadata query for health check with timeout
            using var cts = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
            cts.CancelAfter(TimeSpan.FromSeconds(_config.HealthCheckTimeoutSeconds));
            
            // Create temporary admin client for health check
            using var adminClient = new AdminClientBuilder(new AdminClientConfig
            {
                BootstrapServers = _producer.Name // Get bootstrap servers from producer
            }).Build();
            
            var metadata = adminClient.GetMetadata(TimeSpan.FromSeconds(_config.HealthCheckTimeoutSeconds));
            
            _logger.Debug("Kafka health check passed", new { 
                component = _componentName,
                brokerCount = metadata.Brokers.Count,
                topicCount = metadata.Topics.Count
            });
            
            return true;
        }
        catch (Exception ex)
        {
            _logger.Warning("Kafka health check failed", ex, new {
                component = _componentName,
                errorCode = "INFRA-KAFKA-HEALTH-ERROR"
            });
            return false;
        }
    }

    public void Dispose()
    {
        if (!_disposed)
        {
            _logger.Information("Disposing KafkaProducer", new { component = _componentName });
            _producer?.Dispose();
            _disposed = true;
        }
    }
}