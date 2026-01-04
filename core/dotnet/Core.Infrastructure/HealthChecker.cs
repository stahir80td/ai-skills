using Microsoft.Extensions.Diagnostics.HealthChecks;

namespace Core.Infrastructure;

/// <summary>
/// Health check status
/// </summary>
public enum HealthStatus
{
    Healthy,
    Degraded,
    Unhealthy
}

/// <summary>
/// Individual health check result
/// </summary>
public class HealthCheckResultItem
{
    public required string Name { get; init; }
    public HealthStatus Status { get; init; }
    public string? Message { get; init; }
    public TimeSpan Duration { get; init; }
    public string? Error { get; init; }
}

/// <summary>
/// Overall health check response
/// </summary>
public class HealthCheckResponse
{
    public HealthStatus Status { get; set; }
    public DateTime Timestamp { get; set; } = DateTime.UtcNow;
    public Dictionary<string, HealthCheckResultItem> Checks { get; set; } = new();
    public TimeSpan TotalDuration { get; set; }
}

/// <summary>
/// Health check function delegate
/// </summary>
public delegate Task<HealthCheckResult> HealthCheckFunc(CancellationToken cancellationToken);

/// <summary>
/// Health checker for running multiple health checks
/// </summary>
public class HealthChecker
{
    private readonly Dictionary<string, HealthCheckFunc> _checks = new();
    private readonly TimeSpan _timeout;

    /// <summary>
    /// Creates a new health checker
    /// </summary>
    /// <param name="timeout">Timeout for each health check (minimum 60s recommended)</param>
    public HealthChecker(TimeSpan timeout)
    {
        _timeout = timeout;
    }

    /// <summary>
    /// Registers a health check
    /// </summary>
    public void Register(string name, HealthCheckFunc checkFunc)
    {
        _checks[name] = checkFunc;
    }

    /// <summary>
    /// Registers a health check from an IHealthCheck
    /// </summary>
    public void Register(string name, IHealthCheck healthCheck)
    {
        _checks[name] = async ct =>
        {
            var context = new HealthCheckContext { Registration = new HealthCheckRegistration(name, healthCheck, null, null) };
            return await healthCheck.CheckHealthAsync(context, ct);
        };
    }

    /// <summary>
    /// Runs all health checks
    /// </summary>
    public async Task<HealthCheckResponse> CheckAsync(CancellationToken cancellationToken = default)
    {
        var startTime = DateTime.UtcNow;
        var response = new HealthCheckResponse();
        var tasks = new List<Task<(string Name, HealthCheckResultItem Result)>>();

        foreach (var (name, checkFunc) in _checks)
        {
            tasks.Add(RunCheckAsync(name, checkFunc, cancellationToken));
        }

        var results = await Task.WhenAll(tasks);

        foreach (var (name, result) in results)
        {
            response.Checks[name] = result;
        }

        // Determine overall status
        if (response.Checks.Values.Any(c => c.Status == HealthStatus.Unhealthy))
        {
            response.Status = HealthStatus.Unhealthy;
        }
        else if (response.Checks.Values.Any(c => c.Status == HealthStatus.Degraded))
        {
            response.Status = HealthStatus.Degraded;
        }
        else
        {
            response.Status = HealthStatus.Healthy;
        }

        response.TotalDuration = DateTime.UtcNow - startTime;
        return response;
    }

    private async Task<(string Name, HealthCheckResultItem Result)> RunCheckAsync(
        string name,
        HealthCheckFunc checkFunc,
        CancellationToken cancellationToken)
    {
        var startTime = DateTime.UtcNow;

        try
        {
            using var timeoutCts = new CancellationTokenSource(_timeout);
            using var linkedCts = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken, timeoutCts.Token);

            var result = await checkFunc(linkedCts.Token);
            var duration = DateTime.UtcNow - startTime;

            return (name, new HealthCheckResultItem
            {
                Name = name,
                Status = result.Status switch
                {
                    Microsoft.Extensions.Diagnostics.HealthChecks.HealthStatus.Healthy => HealthStatus.Healthy,
                    Microsoft.Extensions.Diagnostics.HealthChecks.HealthStatus.Degraded => HealthStatus.Degraded,
                    _ => HealthStatus.Unhealthy
                },
                Message = result.Description,
                Duration = duration
            });
        }
        catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested)
        {
            var duration = DateTime.UtcNow - startTime;
            return (name, new HealthCheckResultItem
            {
                Name = name,
                Status = HealthStatus.Unhealthy,
                Message = "Health check timed out",
                Duration = duration,
                Error = $"Timeout after {_timeout.TotalSeconds}s"
            });
        }
        catch (Exception ex)
        {
            var duration = DateTime.UtcNow - startTime;
            return (name, new HealthCheckResultItem
            {
                Name = name,
                Status = HealthStatus.Unhealthy,
                Message = "Health check failed",
                Duration = duration,
                Error = ex.Message
            });
        }
    }
}

/// <summary>
/// Liveness probe check (is the service alive?)
/// </summary>
public class LivenessCheck : IHealthCheck
{
    public Task<HealthCheckResult> CheckHealthAsync(HealthCheckContext context, CancellationToken cancellationToken = default)
    {
        return Task.FromResult(HealthCheckResult.Healthy("Service is alive"));
    }
}

/// <summary>
/// Readiness probe check (is the service ready to accept traffic?)
/// </summary>
public class ReadinessCheck : IHealthCheck
{
    private readonly Func<bool> _isReady;

    public ReadinessCheck(Func<bool> isReady)
    {
        _isReady = isReady;
    }

    public Task<HealthCheckResult> CheckHealthAsync(HealthCheckContext context, CancellationToken cancellationToken = default)
    {
        if (_isReady())
        {
            return Task.FromResult(HealthCheckResult.Healthy("Service is ready"));
        }
        return Task.FromResult(HealthCheckResult.Unhealthy("Service is not ready"));
    }
}
