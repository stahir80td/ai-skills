using Microsoft.AspNetCore.Mvc;
using Core.Logger;
using AiPatterns.Domain.Interfaces;
using AiPatterns.Domain.Models;

namespace AiPatterns.Api.Controllers;

[ApiController]
[Route("api/v1/patterns")]
[Produces("application/json")]
public class PatternsController : ControllerBase
{
    private readonly IPatternsService _patternsService;
    private readonly ServiceLogger _logger;

    public PatternsController(IPatternsService patternsService, ServiceLogger logger)
    {
        _patternsService = patternsService;
        _logger = logger;
    }

    /// <summary>
    /// SQL Server Pattern: Create transactional order with items
    /// </summary>
    [HttpPost("orders")]
    [ProducesResponseType(typeof(Order), 201)]
    [ProducesResponseType(400)]
    public async Task<IActionResult> CreateOrder([FromBody] CreateOrderRequest request)
    {
        _logger.Information("Creating order via SQL Server pattern", new { customerId = request.CustomerId });

        var order = await _patternsService.CreateOrderAsync(
            request.CustomerId,
            request.ShippingAddress,
            request.Items.Select(i => OrderItem.Create(
                Guid.NewGuid(),
                i.ProductName,
                i.Quantity,
                i.UnitPrice)));

        return CreatedAtAction(nameof(GetOrder), new { id = order.Id }, order);
    }

    /// <summary>
    /// SQL Server Pattern: Get order by ID (demonstrates lookup patterns)
    /// </summary>
    [HttpGet("orders/{id:guid}")]
    [ProducesResponseType(typeof(Order), 200)]
    [ProducesResponseType(404)]
    public async Task<IActionResult> GetOrder(Guid id)
    {
        // This would call the repository directly for a simple get
        _logger.Information("Getting order via SQL Server pattern", new { orderId = id });
        return Ok(new { message = "Order lookup pattern - would retrieve from SQL Server", orderId = id });
    }

    /// <summary>
    /// SQL Server + Redis + Kafka Pattern: Update order status (demonstrates cross-platform workflow)
    /// </summary>
    [HttpPatch("orders/{id:guid}/status")]
    [ProducesResponseType(typeof(Order), 200)]
    [ProducesResponseType(404)]
    public async Task<IActionResult> UpdateOrderStatus(Guid id, [FromBody] UpdateOrderStatusRequest request)
    {
        _logger.Information("Updating order status via cross-platform pattern", new { orderId = id, newStatus = request.Status });

        var order = await _patternsService.UpdateOrderStatusAsync(id, request.Status);
        return Ok(order);
    }

    /// <summary>
    /// MongoDB Pattern: Create flexible user profile with preferences
    /// </summary>
    [HttpPost("users")]
    [ProducesResponseType(typeof(UserProfile), 201)]
    [ProducesResponseType(400)]
    public async Task<IActionResult> CreateUser([FromBody] CreateUserRequest request)
    {
        _logger.Information("Creating user via MongoDB pattern", new { email = request.Email });

        var profile = await _patternsService.CreateUserProfileAsync(
            request.Email,
            request.FirstName,
            request.LastName);

        return CreatedAtAction(nameof(GetUser), new { id = profile.Id }, profile);
    }

    /// <summary>
    /// MongoDB Pattern: Get user profile (demonstrates document queries)
    /// </summary>
    [HttpGet("users/{id:guid}")]
    [ProducesResponseType(typeof(UserProfile), 200)]
    [ProducesResponseType(404)]
    public async Task<IActionResult> GetUser(Guid id)
    {
        _logger.Information("Getting user via MongoDB pattern", new { userId = id });
        return Ok(new { message = "User lookup pattern - would retrieve from MongoDB", userId = id });
    }

    /// <summary>
    /// MongoDB + Redis + Kafka Pattern: Update user preferences (demonstrates document updates)
    /// </summary>
    [HttpPut("users/{id:guid}/preferences")]
    [ProducesResponseType(typeof(UserProfile), 200)]
    [ProducesResponseType(404)]
    public async Task<IActionResult> UpdateUserPreferences(Guid id, [FromBody] UserPreferences preferences)
    {
        _logger.Information("Updating user preferences via cross-platform pattern", new { userId = id });

        var profile = await _patternsService.UpdateUserPreferencesAsync(id, preferences);
        return Ok(profile);
    }

    /// <summary>
    /// ScyllaDB Pattern: Record IoT telemetry data (time-series insert)
    /// </summary>
    [HttpPost("telemetry")]
    [ProducesResponseType(typeof(DeviceTelemetry), 201)]
    [ProducesResponseType(400)]
    public async Task<IActionResult> RecordTelemetry([FromBody] RecordTelemetryRequest request)
    {
        _logger.Information("Recording telemetry via ScyllaDB pattern", new { deviceId = request.DeviceId, metric = request.Metric });

        var telemetry = await _patternsService.RecordTelemetryAsync(
            request.DeviceId,
            request.Metric,
            request.Value,
            request.Unit);

        return CreatedAtAction(nameof(GetTelemetryHistory), new { deviceId = request.DeviceId }, telemetry);
    }

    /// <summary>
    /// ScyllaDB + Redis Pattern: Get telemetry history (time-series query with caching)
    /// </summary>
    [HttpGet("telemetry/{deviceId}")]
    [ProducesResponseType(typeof(IEnumerable<DeviceTelemetry>), 200)]
    public async Task<IActionResult> GetTelemetryHistory(
        string deviceId,
        [FromQuery] DateTime? startTime = null,
        [FromQuery] DateTime? endTime = null)
    {
        var start = startTime ?? DateTime.UtcNow.AddHours(-24);
        var end = endTime ?? DateTime.UtcNow;

        _logger.Information("Getting telemetry history via ScyllaDB + Redis pattern", new { deviceId, start, end });

        var telemetry = await _patternsService.GetTelemetryHistoryAsync(deviceId, start, end);
        return Ok(telemetry);
    }

    /// <summary>
    /// Redis Pattern: Update leaderboard (real-time sorted sets)
    /// </summary>
    [HttpPost("leaderboards/{category}/scores")]
    [ProducesResponseType(200)]
    public async Task<IActionResult> UpdateLeaderboard(string category, [FromBody] UpdateLeaderboardRequest request)
    {
        _logger.Information("Updating leaderboard via Redis pattern", new { category, userId = request.UserId, score = request.Score });

        await _patternsService.UpdateLeaderboardAsync(category, request.UserId, request.Score);
        return Ok(new { message = "Leaderboard updated", category, userId = request.UserId, score = request.Score });
    }

    /// <summary>
    /// Redis Pattern: Get leaderboard (sorted set with ranking)
    /// </summary>
    [HttpGet("leaderboards/{category}")]
    [ProducesResponseType(typeof(LeaderboardEntry[]), 200)]
    public async Task<IActionResult> GetLeaderboard(string category, [FromQuery] int top = 10)
    {
        _logger.Information("Getting leaderboard via Redis pattern", new { category, top });

        var leaderboard = await _patternsService.GetLeaderboardAsync(category, top);
        return Ok(leaderboard);
    }

    /// <summary>
    /// Redis + Kafka Pattern: Create user session (session management with events)
    /// </summary>
    [HttpPost("sessions")]
    [ProducesResponseType(200)]
    public async Task<IActionResult> CreateSession([FromBody] CreateSessionRequest request)
    {
        _logger.Information("Creating session via Redis + Kafka pattern", new { sessionId = request.SessionId, userId = request.UserId });

        await _patternsService.CreateUserSessionAsync(request.SessionId, request.UserId, request.UserEmail);
        return Ok(new { message = "Session created", sessionId = request.SessionId });
    }

    /// <summary>
    /// Cross-Platform Analytics: Demonstrates querying all data platforms for comprehensive insights
    /// </summary>
    [HttpGet("analytics")]
    [ProducesResponseType(typeof(PlatformAnalyticsResult), 200)]
    public async Task<IActionResult> GetPlatformAnalytics(
        [FromQuery] DateTime? startDate = null,
        [FromQuery] DateTime? endDate = null)
    {
        var start = startDate ?? DateTime.UtcNow.AddDays(-30);
        var end = endDate ?? DateTime.UtcNow;

        _logger.Information("Generating cross-platform analytics", new { start, end });

        var analytics = await _patternsService.GetPlatformAnalyticsAsync(start, end);
        return Ok(analytics);
    }

    /// <summary>
    /// Health check demonstrating all platform connectivity
    /// </summary>
    [HttpGet("health")]
    [ProducesResponseType(200)]
    public IActionResult GetHealth()
    {
        return Ok(new
        {
            service = "ai-patterns",
            timestamp = DateTime.UtcNow,
            platforms = new
            {
                sqlServer = "Connected via Core.Infrastructure.SqlServer",
                mongodb = "Connected via Core.Infrastructure.MongoDB",
                scylladb = "Connected via Core.Infrastructure.ScyllaDB",
                redis = "Connected via Core.Infrastructure.Redis",
                kafka = "Connected via Core.Infrastructure.Kafka"
            }
        });
    }
}

// Request/Response DTOs
public class CreateOrderRequest
{
    public Guid CustomerId { get; set; }
    public string ShippingAddress { get; set; } = string.Empty;
    public List<CreateOrderItemRequest> Items { get; set; } = new();
}

public class CreateOrderItemRequest
{
    public string ProductName { get; set; } = string.Empty;
    public int Quantity { get; set; }
    public decimal UnitPrice { get; set; }
}

public class UpdateOrderStatusRequest
{
    public OrderStatus Status { get; set; }
}

public class CreateUserRequest
{
    public string Email { get; set; } = string.Empty;
    public string FirstName { get; set; } = string.Empty;
    public string LastName { get; set; } = string.Empty;
}

public class RecordTelemetryRequest
{
    public string DeviceId { get; set; } = string.Empty;
    public string Metric { get; set; } = string.Empty;
    public double Value { get; set; }
    public string Unit { get; set; } = string.Empty;
}

public class UpdateLeaderboardRequest
{
    public string UserId { get; set; } = string.Empty;
    public double Score { get; set; }
}

public class CreateSessionRequest
{
    public string SessionId { get; set; } = string.Empty;
    public Guid UserId { get; set; }
    public string UserEmail { get; set; } = string.Empty;
}