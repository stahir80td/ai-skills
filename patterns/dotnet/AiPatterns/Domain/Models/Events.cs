namespace AiPatterns.Domain.Models;

/// <summary>
/// Event models for Kafka messaging
/// </summary>
/// 
public abstract class BaseEvent
{
    public Guid EventId { get; set; } = Guid.NewGuid();
    public DateTime Timestamp { get; set; } = DateTime.UtcNow;
    public string EventType { get; set; } = string.Empty;
    public string Source { get; set; } = string.Empty;
    public string CorrelationId { get; set; } = string.Empty;
    public Dictionary<string, string> Metadata { get; set; } = new();
}

/// <summary>
/// Order-related events
/// </summary>
public class OrderEvent : BaseEvent
{
    public Guid OrderId { get; set; }
    public Guid CustomerId { get; set; }
    public OrderStatus Status { get; set; }
    public decimal TotalAmount { get; set; }
    public int ItemCount { get; set; }
    public string ShippingAddress { get; set; } = string.Empty;

    public static OrderEvent OrderCreated(Order order, string source = "order-service")
    {
        return new OrderEvent
        {
            EventType = "OrderCreated",
            Source = source,
            OrderId = order.Id,
            CustomerId = order.CustomerId,
            Status = order.Status,
            TotalAmount = order.TotalAmount,
            ItemCount = order.Items.Count,
            ShippingAddress = order.ShippingAddress
        };
    }

    public static OrderEvent OrderStatusChanged(Order order, OrderStatus previousStatus, string source = "order-service")
    {
        return new OrderEvent
        {
            EventType = "OrderStatusChanged",
            Source = source,
            OrderId = order.Id,
            CustomerId = order.CustomerId,
            Status = order.Status,
            TotalAmount = order.TotalAmount,
            ItemCount = order.Items.Count,
            Metadata = new Dictionary<string, string>
            {
                ["previousStatus"] = previousStatus.ToString(),
                ["newStatus"] = order.Status.ToString()
            }
        };
    }
}

/// <summary>
/// User-related events
/// </summary>
public class UserEvent : BaseEvent
{
    public Guid UserId { get; set; }
    public string Email { get; set; } = string.Empty;
    public string FirstName { get; set; } = string.Empty;
    public string LastName { get; set; } = string.Empty;

    public static UserEvent UserRegistered(UserProfile profile, string source = "user-service")
    {
        return new UserEvent
        {
            EventType = "UserRegistered",
            Source = source,
            UserId = profile.Id,
            Email = profile.Email,
            FirstName = profile.FirstName,
            LastName = profile.LastName
        };
    }

    public static UserEvent UserProfileUpdated(UserProfile profile, string source = "user-service")
    {
        return new UserEvent
        {
            EventType = "UserProfileUpdated",
            Source = source,
            UserId = profile.Id,
            Email = profile.Email,
            FirstName = profile.FirstName,
            LastName = profile.LastName
        };
    }

    public static UserEvent UserLoggedIn(Guid userId, string email, string source = "auth-service")
    {
        return new UserEvent
        {
            EventType = "UserLoggedIn",
            Source = source,
            UserId = userId,
            Email = email
        };
    }
}

/// <summary>
/// Telemetry-related events
/// </summary>
public class TelemetryEvent : BaseEvent
{
    public string DeviceId { get; set; } = string.Empty;
    public string Metric { get; set; } = string.Empty;
    public double Value { get; set; }
    public string Unit { get; set; } = string.Empty;
    public string Location { get; set; } = string.Empty;
    public string DataType { get; set; } = string.Empty;
    public string Quality { get; set; } = string.Empty;

    public static TelemetryEvent TelemetryReceived(DeviceTelemetry telemetry, string source = "iot-ingestion")
    {
        return new TelemetryEvent
        {
            EventType = "TelemetryReceived",
            Source = source,
            DeviceId = telemetry.DeviceId,
            Metric = telemetry.Metric,
            Value = telemetry.Value,
            Unit = telemetry.Unit,
            Location = telemetry.Location,
            DataType = telemetry.DataType,
            Quality = telemetry.Quality,
            CorrelationId = telemetry.CorrelationId.ToString()
        };
    }

    public static TelemetryEvent AnomalyDetected(DeviceTelemetry telemetry, string anomalyType, double severity, string source = "anomaly-detector")
    {
        return new TelemetryEvent
        {
            EventType = "AnomalyDetected",
            Source = source,
            DeviceId = telemetry.DeviceId,
            Metric = telemetry.Metric,
            Value = telemetry.Value,
            Unit = telemetry.Unit,
            Location = telemetry.Location,
            DataType = telemetry.DataType,
            Quality = telemetry.Quality,
            Metadata = new Dictionary<string, string>
            {
                ["anomalyType"] = anomalyType,
                ["severity"] = severity.ToString(),
                ["threshold"] = "exceeded"
            }
        };
    }
}

/// <summary>
/// System-related events
/// </summary>
public class SystemEvent : BaseEvent
{
    public string Component { get; set; } = string.Empty;
    public string Level { get; set; } = string.Empty; // INFO, WARN, ERROR, CRITICAL
    public string Message { get; set; } = string.Empty;
    public Dictionary<string, object> Context { get; set; } = new();

    public static SystemEvent ServiceStarted(string serviceName, string version, string source)
    {
        return new SystemEvent
        {
            EventType = "ServiceStarted",
            Source = source,
            Component = serviceName,
            Level = "INFO",
            Message = $"{serviceName} service started successfully",
            Context = new Dictionary<string, object>
            {
                ["version"] = version,
                ["environment"] = Environment.GetEnvironmentVariable("ASPNETCORE_ENVIRONMENT") ?? "Unknown"
            }
        };
    }

    public static SystemEvent HealthCheckFailed(string component, string checkName, string error, string source)
    {
        return new SystemEvent
        {
            EventType = "HealthCheckFailed",
            Source = source,
            Component = component,
            Level = "ERROR",
            Message = $"Health check failed: {checkName}",
            Context = new Dictionary<string, object>
            {
                ["checkName"] = checkName,
                ["error"] = error,
                ["severity"] = "HIGH"
            }
        };
    }
}

/// <summary>
/// Cache models for Redis
/// </summary>
public class LeaderboardEntry
{
    public string UserId { get; set; } = string.Empty;
    public string DisplayName { get; set; } = string.Empty;
    public double Score { get; set; }
    public int Rank { get; set; }
    public DateTime UpdatedAt { get; set; }

    public static LeaderboardEntry Create(string userId, string displayName, double score)
    {
        return new LeaderboardEntry
        {
            UserId = userId,
            DisplayName = displayName,
            Score = score,
            UpdatedAt = DateTime.UtcNow
        };
    }
}

public class SessionData
{
    public string SessionId { get; set; } = string.Empty;
    public Guid UserId { get; set; }
    public string UserEmail { get; set; } = string.Empty;
    public Dictionary<string, object> Properties { get; set; } = new();
    public DateTime CreatedAt { get; set; }
    public DateTime ExpiresAt { get; set; }

    public static SessionData Create(string sessionId, Guid userId, string userEmail, TimeSpan expiry)
    {
        return new SessionData
        {
            SessionId = sessionId,
            UserId = userId,
            UserEmail = userEmail,
            CreatedAt = DateTime.UtcNow,
            ExpiresAt = DateTime.UtcNow.Add(expiry)
        };
    }
}