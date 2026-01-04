using Prometheus;

namespace Core.Metrics;

/// <summary>
/// Configuration for service metrics
/// </summary>
public class MetricsConfig
{
    /// <summary>
    /// Service name for metric labels
    /// </summary>
    public string ServiceName { get; set; } = "unknown-service";

    /// <summary>
    /// Prometheus namespace (default: "AI")
    /// </summary>
    public string Namespace { get; set; } = "AI";

    /// <summary>
    /// Prometheus subsystem (optional)
    /// </summary>
    public string? Subsystem { get; set; }

    /// <summary>
    /// Custom latency buckets (in seconds)
    /// </summary>
    public double[]? LatencyBuckets { get; set; }
}

/// <summary>
/// Service metrics following the Four Golden Signals:
/// 1. Latency - How long requests take
/// 2. Traffic - How many requests the service handles
/// 3. Errors - Rate of failed requests
/// 4. Saturation - Resource utilization
/// </summary>
public class ServiceMetrics
{
    private readonly string _serviceName;

    // Latency (Golden Signal #1)
    private readonly Histogram _requestDuration;

    // Traffic (Golden Signal #2)
    private readonly Counter _requestTotal;

    // Errors (Golden Signal #3)
    private readonly Counter _errorTotal;

    // Saturation (Golden Signal #4)
    private readonly Gauge _resourceUtilization;
    private readonly Gauge _activeRequests;
    private readonly Gauge _queueDepth;

    /// <summary>
    /// Default latency buckets: 10ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
    /// </summary>
    public static readonly double[] DefaultLatencyBuckets = { 0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0 };

    /// <summary>
    /// Creates a new service metrics instance
    /// </summary>
    public ServiceMetrics(MetricsConfig config)
    {
        _serviceName = config.ServiceName;
        var buckets = config.LatencyBuckets ?? DefaultLatencyBuckets;

        // Build metric names with namespace/subsystem
        var prefix = config.Namespace;
        if (!string.IsNullOrEmpty(config.Subsystem))
        {
            prefix = $"{prefix}_{config.Subsystem}";
        }

        // Latency: Request duration histogram
        _requestDuration = Prometheus.Metrics.CreateHistogram(
            $"{prefix}_request_duration_seconds",
            "Request latency in seconds (Golden Signal: Latency)",
            new HistogramConfiguration
            {
                LabelNames = new[] { "service", "method", "endpoint", "status" },
                Buckets = buckets
            });

        // Traffic: Total requests counter
        _requestTotal = Prometheus.Metrics.CreateCounter(
            $"{prefix}_requests_total",
            "Total number of requests (Golden Signal: Traffic)",
            new CounterConfiguration
            {
                LabelNames = new[] { "service", "method", "endpoint", "status" }
            });

        // Errors: Error counter
        _errorTotal = Prometheus.Metrics.CreateCounter(
            $"{prefix}_errors_total",
            "Total number of errors (Golden Signal: Errors)",
            new CounterConfiguration
            {
                LabelNames = new[] { "service", "error_code", "severity", "component" }
            });

        // Saturation: Resource utilization gauge
        _resourceUtilization = Prometheus.Metrics.CreateGauge(
            $"{prefix}_resource_utilization",
            "Resource utilization percentage 0-100 (Golden Signal: Saturation)",
            new GaugeConfiguration
            {
                LabelNames = new[] { "service", "resource_type" }
            });

        // Saturation: Active requests gauge
        _activeRequests = Prometheus.Metrics.CreateGauge(
            $"{prefix}_active_requests",
            "Number of requests currently being processed (Golden Signal: Saturation)",
            new GaugeConfiguration
            {
                LabelNames = new[] { "service" }
            });

        // Saturation: Queue depth gauge
        _queueDepth = Prometheus.Metrics.CreateGauge(
            $"{prefix}_queue_depth",
            "Number of items in processing queue (Golden Signal: Saturation)",
            new GaugeConfiguration
            {
                LabelNames = new[] { "service" }
            });
    }

    /// <summary>
    /// Records a completed request with latency and status
    /// Golden Signals: Latency + Traffic
    /// </summary>
    public void RecordRequest(string method, string endpoint, string status, TimeSpan duration)
    {
        _requestDuration.WithLabels(_serviceName, method, endpoint, status).Observe(duration.TotalSeconds);
        _requestTotal.WithLabels(_serviceName, method, endpoint, status).Inc();
    }

    /// <summary>
    /// Records an error occurrence
    /// Golden Signal: Errors
    /// </summary>
    public void RecordError(string errorCode, string severity, string component)
    {
        _errorTotal.WithLabels(_serviceName, errorCode, severity, component).Inc();
    }

    /// <summary>
    /// Updates resource utilization metric
    /// Golden Signal: Saturation
    /// </summary>
    public void SetResourceUtilization(string resourceType, double percent)
    {
        _resourceUtilization.WithLabels(_serviceName, resourceType).Set(percent);
    }

    /// <summary>
    /// Increments active request count (call before processing)
    /// </summary>
    public void IncrementActiveRequests()
    {
        _activeRequests.WithLabels(_serviceName).Inc();
    }

    /// <summary>
    /// Decrements active request count (call after processing)
    /// </summary>
    public void DecrementActiveRequests()
    {
        _activeRequests.WithLabels(_serviceName).Dec();
    }

    /// <summary>
    /// Updates queue depth metric
    /// </summary>
    public void SetQueueDepth(double depth)
    {
        _queueDepth.WithLabels(_serviceName).Set(depth);
    }

    /// <summary>
    /// Creates a timer that records request duration when disposed
    /// </summary>
    public IDisposable TimeRequest(string method, string endpoint, Func<string> getStatus)
    {
        return new RequestTimer(this, method, endpoint, getStatus);
    }

    private class RequestTimer : IDisposable
    {
        private readonly ServiceMetrics _metrics;
        private readonly string _method;
        private readonly string _endpoint;
        private readonly Func<string> _getStatus;
        private readonly DateTime _start;

        public RequestTimer(ServiceMetrics metrics, string method, string endpoint, Func<string> getStatus)
        {
            _metrics = metrics;
            _method = method;
            _endpoint = endpoint;
            _getStatus = getStatus;
            _start = DateTime.UtcNow;
            _metrics.IncrementActiveRequests();
        }

        public void Dispose()
        {
            var duration = DateTime.UtcNow - _start;
            _metrics.RecordRequest(_method, _endpoint, _getStatus(), duration);
            _metrics.DecrementActiveRequests();
        }
    }
}
