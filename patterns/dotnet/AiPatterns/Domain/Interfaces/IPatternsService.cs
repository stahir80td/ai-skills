using AiPatterns.Domain.Models;

namespace AiPatterns.Domain.Interfaces;

public interface IPatternsService
{
    // SQL Server (Transactional) - Order Management
    Task<Order> CreateOrderAsync(Guid customerId, string shippingAddress, IEnumerable<OrderItem> items);
    Task<Order> UpdateOrderStatusAsync(Guid orderId, OrderStatus newStatus);

    // MongoDB (Document) - User Profile Management
    Task<UserProfile> CreateUserProfileAsync(string email, string firstName, string lastName);
    Task<UserProfile> UpdateUserPreferencesAsync(Guid userId, UserPreferences preferences);

    // ScyllaDB (Time-Series) - IoT Telemetry
    Task<DeviceTelemetry> RecordTelemetryAsync(string deviceId, string metric, double value, string unit);
    Task<IEnumerable<DeviceTelemetry>> GetTelemetryHistoryAsync(string deviceId, DateTime startTime, DateTime endTime);

    // Redis (Real-time Cache) - Leaderboards and Sessions
    Task UpdateLeaderboardAsync(string category, string userId, double score);
    Task<LeaderboardEntry[]> GetLeaderboardAsync(string category, int top = 10);
    Task CreateUserSessionAsync(string sessionId, Guid userId, string userEmail);

    // Cross-Platform Analytics
    Task<PlatformAnalyticsResult> GetPlatformAnalyticsAsync(DateTime startDate, DateTime endDate);
}

// Analytics models for comprehensive cross-platform analysis
public class PlatformAnalyticsResult
{
    public OrderAnalytics OrderAnalytics { get; set; } = new();
    public UserAnalytics UserAnalytics { get; set; } = new();
    public TelemetryAnalytics TelemetryAnalytics { get; set; } = new();
    public CacheAnalytics CacheAnalytics { get; set; } = new();
    public DateTime GeneratedAt { get; set; }
}

public class OrderAnalytics
{
    public int TotalOrders { get; set; }
    public decimal TotalRevenue { get; set; }
    public decimal AverageOrderValue { get; set; }
    public int PendingOrders { get; set; }
    public int CompletedOrders { get; set; }
}

public class UserAnalytics
{
    public int TotalUsers { get; set; }
    public int ActiveUsers { get; set; }
    public int NewUsers { get; set; }
    public string[] TopCategories { get; set; } = Array.Empty<string>();
}

public class TelemetryAnalytics
{
    public int TotalDataPoints { get; set; }
    public int UniqueDevices { get; set; }
    public double AverageValue { get; set; }
    public string[] TopMetrics { get; set; } = Array.Empty<string>();
}

public class CacheAnalytics
{
    public int TotalKeys { get; set; }
    public double HitRate { get; set; }
    public string MemoryUsed { get; set; } = string.Empty;
    public int ActiveSessions { get; set; }
}

// Session management for Redis patterns
public class SessionData
{
    public string SessionId { get; set; } = string.Empty;
    public Guid UserId { get; set; }
    public string UserEmail { get; set; } = string.Empty;
    public DateTime CreatedAt { get; set; }
    public DateTime ExpiresAt { get; set; }
    public Dictionary<string, object> Data { get; set; } = new();

    public static SessionData Create(string sessionId, Guid userId, string userEmail, TimeSpan expiration)
    {
        var now = DateTime.UtcNow;
        return new SessionData
        {
            SessionId = sessionId,
            UserId = userId,
            UserEmail = userEmail,
            CreatedAt = now,
            ExpiresAt = now.Add(expiration)
        };
    }
}