using Prometheus;

namespace Core.Sli;

/// <summary>
/// Request outcome for SLI tracking
/// </summary>
public class RequestOutcome
{
    /// <summary>
    /// Operation name (e.g., "GetDevice", "ProcessMessage")
    /// </summary>
    public string Operation { get; set; } = "default";

    /// <summary>
    /// Whether the request succeeded
    /// </summary>
    public bool Success { get; set; }

    /// <summary>
    /// Request latency
    /// </summary>
    public TimeSpan Latency { get; set; }

    /// <summary>
    /// Error code if failed
    /// </summary>
    public string? ErrorCode { get; set; }

    /// <summary>
    /// Error severity if failed
    /// </summary>
    public string? ErrorSeverity { get; set; }
}

/// <summary>
/// SLI metrics snapshot
/// </summary>
public class SliMetrics
{
    /// <summary>
    /// Window start time
    /// </summary>
    public DateTime WindowStart { get; set; }

    /// <summary>
    /// Window end time
    /// </summary>
    public DateTime WindowEnd { get; set; }

    /// <summary>
    /// Availability percentage (0-100)
    /// </summary>
    public double Availability { get; set; }

    /// <summary>
    /// P95 latency in milliseconds
    /// </summary>
    public double LatencyP95Ms { get; set; }

    /// <summary>
    /// P99 latency in milliseconds
    /// </summary>
    public double LatencyP99Ms { get; set; }

    /// <summary>
    /// Error rate percentage (0-100)
    /// </summary>
    public double ErrorRate { get; set; }

    /// <summary>
    /// Throughput (requests per second)
    /// </summary>
    public double Throughput { get; set; }
}

/// <summary>
/// Interface for SLI tracking
/// </summary>
public interface ISliTracker
{
    /// <summary>
    /// Records a request outcome
    /// </summary>
    void RecordRequest(RequestOutcome outcome);

    /// <summary>
    /// Records request latency
    /// </summary>
    void RecordLatency(TimeSpan duration, string operation);

    /// <summary>
    /// Records throughput events
    /// </summary>
    void RecordThroughput(int count, string operation);
}

/// <summary>
/// Prometheus-based SLI tracker
/// </summary>
public class PrometheusSliTracker : ISliTracker
{
    private readonly string _serviceName;

    private static readonly Counter RequestsTotal = Prometheus.Metrics.CreateCounter(
        "sli_requests_total",
        "Total number of requests for SLI tracking",
        new CounterConfiguration { LabelNames = new[] { "service", "operation" } });

    private static readonly Counter RequestsSuccess = Prometheus.Metrics.CreateCounter(
        "sli_requests_success_total",
        "Total number of successful requests",
        new CounterConfiguration { LabelNames = new[] { "service", "operation" } });

    private static readonly Counter RequestsFailed = Prometheus.Metrics.CreateCounter(
        "sli_requests_failed_total",
        "Total number of failed requests",
        new CounterConfiguration { LabelNames = new[] { "service", "operation", "error_code", "severity" } });

    private static readonly Histogram RequestDuration = Prometheus.Metrics.CreateHistogram(
        "sli_request_duration_seconds",
        "Request duration in seconds for SLI tracking",
        new HistogramConfiguration
        {
            LabelNames = new[] { "service", "operation" },
            Buckets = new[] { 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10 }
        });

    private static readonly Gauge ThroughputRate = Prometheus.Metrics.CreateGauge(
        "sli_throughput_rate",
        "Current throughput rate (requests or messages per second)",
        new GaugeConfiguration { LabelNames = new[] { "service", "operation", "type" } });

    private static readonly Gauge SliAvailability = Prometheus.Metrics.CreateGauge(
        "sli_availability_percent",
        "Current availability SLI in percent",
        new GaugeConfiguration { LabelNames = new[] { "service" } });

    private static readonly Gauge SliLatencyP95 = Prometheus.Metrics.CreateGauge(
        "sli_latency_p95_milliseconds",
        "P95 latency SLI in milliseconds",
        new GaugeConfiguration { LabelNames = new[] { "service", "operation" } });

    private static readonly Gauge SliLatencyP99 = Prometheus.Metrics.CreateGauge(
        "sli_latency_p99_milliseconds",
        "P99 latency SLI in milliseconds",
        new GaugeConfiguration { LabelNames = new[] { "service", "operation" } });

    private static readonly Gauge SliErrorRate = Prometheus.Metrics.CreateGauge(
        "sli_error_rate_percent",
        "Current error rate SLI in percent",
        new GaugeConfiguration { LabelNames = new[] { "service" } });

    /// <summary>
    /// Creates a new Prometheus SLI tracker
    /// </summary>
    public PrometheusSliTracker(string serviceName)
    {
        _serviceName = serviceName;
    }

    /// <inheritdoc/>
    public void RecordRequest(RequestOutcome outcome)
    {
        var operation = string.IsNullOrEmpty(outcome.Operation) ? "default" : outcome.Operation;

        RequestsTotal.WithLabels(_serviceName, operation).Inc();

        if (outcome.Success)
        {
            RequestsSuccess.WithLabels(_serviceName, operation).Inc();
        }
        else
        {
            RequestsFailed.WithLabels(
                _serviceName,
                operation,
                outcome.ErrorCode ?? "unknown",
                outcome.ErrorSeverity ?? "unknown"
            ).Inc();
        }

        if (outcome.Latency > TimeSpan.Zero)
        {
            RequestDuration.WithLabels(_serviceName, operation).Observe(outcome.Latency.TotalSeconds);
        }
    }

    /// <inheritdoc/>
    public void RecordLatency(TimeSpan duration, string operation)
    {
        operation = string.IsNullOrEmpty(operation) ? "default" : operation;
        RequestDuration.WithLabels(_serviceName, operation).Observe(duration.TotalSeconds);
    }

    /// <inheritdoc/>
    public void RecordThroughput(int count, string operation)
    {
        operation = string.IsNullOrEmpty(operation) ? "default" : operation;
        ThroughputRate.WithLabels(_serviceName, operation, "requests").Inc(count);
    }

    /// <summary>
    /// Updates the availability SLI gauge
    /// </summary>
    public void SetAvailability(double percent)
    {
        SliAvailability.WithLabels(_serviceName).Set(percent);
    }

    /// <summary>
    /// Updates the P95 latency SLI gauge
    /// </summary>
    public void SetLatencyP95(string operation, double milliseconds)
    {
        SliLatencyP95.WithLabels(_serviceName, operation).Set(milliseconds);
    }

    /// <summary>
    /// Updates the P99 latency SLI gauge
    /// </summary>
    public void SetLatencyP99(string operation, double milliseconds)
    {
        SliLatencyP99.WithLabels(_serviceName, operation).Set(milliseconds);
    }

    /// <summary>
    /// Updates the error rate SLI gauge
    /// </summary>
    public void SetErrorRate(double percent)
    {
        SliErrorRate.WithLabels(_serviceName).Set(percent);
    }
}

/// <summary>
/// SLI budget tracking for error budgets
/// </summary>
public class SliBudget
{
    /// <summary>
    /// Service name
    /// </summary>
    public string ServiceName { get; }

    /// <summary>
    /// Target availability (e.g., 99.9)
    /// </summary>
    public double TargetAvailability { get; }

    /// <summary>
    /// Budget window (typically 30 days)
    /// </summary>
    public TimeSpan BudgetWindow { get; }

    /// <summary>
    /// Total allowed downtime in the budget window
    /// </summary>
    public TimeSpan AllowedDowntime => BudgetWindow * (1 - TargetAvailability / 100);

    /// <summary>
    /// Creates a new SLI budget
    /// </summary>
    public SliBudget(string serviceName, double targetAvailability, TimeSpan budgetWindow)
    {
        ServiceName = serviceName;
        TargetAvailability = targetAvailability;
        BudgetWindow = budgetWindow;
    }

    /// <summary>
    /// Calculates remaining error budget
    /// </summary>
    public TimeSpan RemainingBudget(TimeSpan downtimeUsed)
    {
        return AllowedDowntime - downtimeUsed;
    }

    /// <summary>
    /// Calculates budget burn rate
    /// </summary>
    public double BurnRate(TimeSpan downtimeUsed, TimeSpan elapsed)
    {
        if (elapsed <= TimeSpan.Zero) return 0;

        var expectedBurn = AllowedDowntime * (elapsed / BudgetWindow);
        if (expectedBurn <= TimeSpan.Zero) return 0;

        return downtimeUsed / expectedBurn;
    }
}
