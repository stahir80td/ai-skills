using Polly;
using Polly.CircuitBreaker;
using Polly.Retry;
using Polly.Timeout;
using Prometheus;

namespace Core.Reliability;

/// <summary>
/// Circuit breaker states
/// </summary>
public enum CircuitState
{
    Closed,
    Open,
    HalfOpen
}

/// <summary>
/// Circuit breaker configuration
/// </summary>
public class CircuitBreakerConfig
{
    /// <summary>Name for metrics/logging</summary>
    public string Name { get; set; } = "default";

    /// <summary>Number of failures before opening</summary>
    public int FailureThreshold { get; set; } = 5;

    /// <summary>Time to wait before trying again</summary>
    public TimeSpan BreakDuration { get; set; } = TimeSpan.FromSeconds(30);

    /// <summary>Sampling duration for failure counting</summary>
    public TimeSpan SamplingDuration { get; set; } = TimeSpan.FromSeconds(60);

    /// <summary>Minimum throughput before circuit can open</summary>
    public int MinimumThroughput { get; set; } = 10;
}

/// <summary>
/// Circuit breaker with Prometheus metrics
/// </summary>
public class CircuitBreaker
{
    private readonly AsyncCircuitBreakerPolicy _policy;
    private readonly string _name;

    private static readonly Gauge CircuitBreakerState = Prometheus.Metrics.CreateGauge(
        "circuit_breaker_state",
        "Circuit breaker state (0=closed, 1=half_open, 2=open)",
        new GaugeConfiguration { LabelNames = new[] { "name" } });

    private static readonly Counter CircuitBreakerRequests = Prometheus.Metrics.CreateCounter(
        "circuit_breaker_requests_total",
        "Total requests through circuit breaker",
        new CounterConfiguration { LabelNames = new[] { "name", "state", "result" } });

    private static readonly Counter CircuitBreakerStateChanges = Prometheus.Metrics.CreateCounter(
        "circuit_breaker_state_changes_total",
        "Total circuit breaker state transitions",
        new CounterConfiguration { LabelNames = new[] { "name", "from", "to" } });

    /// <summary>
    /// Creates a new circuit breaker
    /// </summary>
    public CircuitBreaker(CircuitBreakerConfig config)
    {
        _name = config.Name;

        _policy = Policy
            .Handle<Exception>()
            .AdvancedCircuitBreakerAsync(
                failureThreshold: 0.5, // 50% failure rate
                samplingDuration: config.SamplingDuration,
                minimumThroughput: config.MinimumThroughput,
                durationOfBreak: config.BreakDuration,
                onBreak: (exception, duration) =>
                {
                    CircuitBreakerState.WithLabels(_name).Set((int)CircuitState.Open);
                    CircuitBreakerStateChanges.WithLabels(_name, "closed", "open").Inc();
                },
                onReset: () =>
                {
                    CircuitBreakerState.WithLabels(_name).Set((int)CircuitState.Closed);
                    CircuitBreakerStateChanges.WithLabels(_name, "half_open", "closed").Inc();
                },
                onHalfOpen: () =>
                {
                    CircuitBreakerState.WithLabels(_name).Set((int)CircuitState.HalfOpen);
                    CircuitBreakerStateChanges.WithLabels(_name, "open", "half_open").Inc();
                });

        CircuitBreakerState.WithLabels(_name).Set((int)CircuitState.Closed);
    }

    /// <summary>
    /// Executes an action with circuit breaker protection
    /// </summary>
    public async Task ExecuteAsync(Func<Task> action)
    {
        var stateLabel = GetStateLabel();
        try
        {
            await _policy.ExecuteAsync(action);
            CircuitBreakerRequests.WithLabels(_name, stateLabel, "success").Inc();
        }
        catch (BrokenCircuitException)
        {
            CircuitBreakerRequests.WithLabels(_name, "open", "rejected").Inc();
            throw;
        }
        catch
        {
            CircuitBreakerRequests.WithLabels(_name, stateLabel, "error").Inc();
            throw;
        }
    }

    /// <summary>
    /// Executes a function with circuit breaker protection
    /// </summary>
    public async Task<T> ExecuteAsync<T>(Func<Task<T>> action)
    {
        var stateLabel = GetStateLabel();
        try
        {
            var result = await _policy.ExecuteAsync(action);
            CircuitBreakerRequests.WithLabels(_name, stateLabel, "success").Inc();
            return result;
        }
        catch (BrokenCircuitException)
        {
            CircuitBreakerRequests.WithLabels(_name, "open", "rejected").Inc();
            throw;
        }
        catch
        {
            CircuitBreakerRequests.WithLabels(_name, stateLabel, "error").Inc();
            throw;
        }
    }

    /// <summary>
    /// Gets current circuit state
    /// </summary>
    public CircuitState State => _policy.CircuitState switch
    {
        Polly.CircuitBreaker.CircuitState.Closed => CircuitState.Closed,
        Polly.CircuitBreaker.CircuitState.Open => CircuitState.Open,
        Polly.CircuitBreaker.CircuitState.HalfOpen => CircuitState.HalfOpen,
        Polly.CircuitBreaker.CircuitState.Isolated => CircuitState.Open,
        _ => CircuitState.Closed
    };

    /// <summary>
    /// Manually resets the circuit breaker
    /// </summary>
    public void Reset()
    {
        _policy.Reset();
        CircuitBreakerState.WithLabels(_name).Set((int)CircuitState.Closed);
    }

    private string GetStateLabel() => State switch
    {
        CircuitState.Closed => "closed",
        CircuitState.Open => "open",
        CircuitState.HalfOpen => "half_open",
        _ => "unknown"
    };
}

/// <summary>
/// Retry configuration
/// </summary>
public class RetryConfig
{
    /// <summary>Name for metrics/logging</summary>
    public string Name { get; set; } = "default";

    /// <summary>Maximum retry attempts</summary>
    public int MaxAttempts { get; set; } = 3;

    /// <summary>Initial delay between retries</summary>
    public TimeSpan InitialDelay { get; set; } = TimeSpan.FromMilliseconds(100);

    /// <summary>Maximum delay between retries</summary>
    public TimeSpan MaxDelay { get; set; } = TimeSpan.FromSeconds(30);

    /// <summary>Backoff multiplier</summary>
    public double Multiplier { get; set; } = 2.0;

    /// <summary>Add random jitter to delays</summary>
    public bool Jitter { get; set; } = true;
}

/// <summary>
/// Retry policy with exponential backoff and metrics
/// </summary>
public class RetryPolicy
{
    private readonly AsyncRetryPolicy _policy;
    private readonly string _name;

    private static readonly Counter RetryAttempts = Prometheus.Metrics.CreateCounter(
        "retry_attempts_total",
        "Total retry attempts",
        new CounterConfiguration { LabelNames = new[] { "name", "attempt" } });

    private static readonly Counter RetrySuccess = Prometheus.Metrics.CreateCounter(
        "retry_success_total",
        "Total successful retries",
        new CounterConfiguration { LabelNames = new[] { "name", "attempt" } });

    private static readonly Counter RetryFailure = Prometheus.Metrics.CreateCounter(
        "retry_failure_total",
        "Total failed retries (after all attempts)",
        new CounterConfiguration { LabelNames = new[] { "name" } });

    private static readonly Histogram RetryDelays = Prometheus.Metrics.CreateHistogram(
        "retry_delay_seconds",
        "Delay between retry attempts",
        new HistogramConfiguration
        {
            LabelNames = new[] { "name" },
            Buckets = new[] { .001, .01, .1, .5, 1, 2, 5, 10, 30 }
        });

    /// <summary>
    /// Creates a new retry policy
    /// </summary>
    public RetryPolicy(RetryConfig config)
    {
        _name = config.Name;

        _policy = Policy
            .Handle<Exception>()
            .WaitAndRetryAsync(
                config.MaxAttempts - 1,
                retryAttempt =>
                {
                    var delay = TimeSpan.FromMilliseconds(
                        config.InitialDelay.TotalMilliseconds * Math.Pow(config.Multiplier, retryAttempt - 1));

                    if (delay > config.MaxDelay)
                        delay = config.MaxDelay;

                    if (config.Jitter)
                    {
                        var jitter = Random.Shared.NextDouble() * 0.3; // Up to 30% jitter
                        delay = TimeSpan.FromMilliseconds(delay.TotalMilliseconds * (1 + jitter));
                    }

                    RetryAttempts.WithLabels(_name, retryAttempt.ToString()).Inc();
                    RetryDelays.WithLabels(_name).Observe(delay.TotalSeconds);

                    return delay;
                },
                (exception, timeSpan, retryCount, context) =>
                {
                    // Logging hook
                });
    }

    /// <summary>
    /// Executes an action with retry logic
    /// </summary>
    public async Task ExecuteAsync(Func<Task> action)
    {
        try
        {
            await _policy.ExecuteAsync(action);
            RetrySuccess.WithLabels(_name, "final").Inc();
        }
        catch
        {
            RetryFailure.WithLabels(_name).Inc();
            throw;
        }
    }

    /// <summary>
    /// Executes a function with retry logic
    /// </summary>
    public async Task<T> ExecuteAsync<T>(Func<Task<T>> action)
    {
        try
        {
            var result = await _policy.ExecuteAsync(action);
            RetrySuccess.WithLabels(_name, "final").Inc();
            return result;
        }
        catch
        {
            RetryFailure.WithLabels(_name).Inc();
            throw;
        }
    }
}

/// <summary>
/// Rate limiter configuration
/// </summary>
public class RateLimiterConfig
{
    /// <summary>Name for metrics/logging</summary>
    public string Name { get; set; } = "default";

    /// <summary>Requests allowed per second</summary>
    public double RequestsPerSecond { get; set; } = 100;

    /// <summary>Burst capacity</summary>
    public int Burst { get; set; } = 10;
}

/// <summary>
/// Token bucket rate limiter with metrics
/// </summary>
public class RateLimiter
{
    private readonly SemaphoreSlim _semaphore;
    private readonly string _name;
    private readonly double _requestsPerSecond;
    private readonly Timer _refillTimer;
    private int _currentTokens;
    private readonly int _maxTokens;
    private readonly object _lock = new();

    private static readonly Counter RateLimitRequests = Prometheus.Metrics.CreateCounter(
        "rate_limit_requests_total",
        "Total requests to rate limiter",
        new CounterConfiguration { LabelNames = new[] { "name" } });

    private static readonly Counter RateLimitRejected = Prometheus.Metrics.CreateCounter(
        "rate_limit_rejected_total",
        "Total requests rejected by rate limiter",
        new CounterConfiguration { LabelNames = new[] { "name" } });

    private static readonly Counter RateLimitAllowed = Prometheus.Metrics.CreateCounter(
        "rate_limit_allowed_total",
        "Total requests allowed by rate limiter",
        new CounterConfiguration { LabelNames = new[] { "name" } });

    private static readonly Gauge RateLimitTokensAvailable = Prometheus.Metrics.CreateGauge(
        "rate_limit_tokens_available",
        "Number of tokens currently available",
        new GaugeConfiguration { LabelNames = new[] { "name" } });

    /// <summary>
    /// Creates a new rate limiter
    /// </summary>
    public RateLimiter(RateLimiterConfig config)
    {
        _name = config.Name;
        _requestsPerSecond = config.RequestsPerSecond;
        _maxTokens = config.Burst;
        _currentTokens = config.Burst;
        _semaphore = new SemaphoreSlim(1, 1);

        // Refill tokens periodically
        var refillInterval = TimeSpan.FromSeconds(1.0 / _requestsPerSecond);
        _refillTimer = new Timer(_ => RefillToken(), null, refillInterval, refillInterval);

        RateLimitTokensAvailable.WithLabels(_name).Set(_currentTokens);
    }

    private void RefillToken()
    {
        lock (_lock)
        {
            if (_currentTokens < _maxTokens)
            {
                _currentTokens++;
                RateLimitTokensAvailable.WithLabels(_name).Set(_currentTokens);
            }
        }
    }

    /// <summary>
    /// Attempts to acquire a token (non-blocking)
    /// </summary>
    public bool TryAcquire()
    {
        RateLimitRequests.WithLabels(_name).Inc();

        lock (_lock)
        {
            if (_currentTokens > 0)
            {
                _currentTokens--;
                RateLimitTokensAvailable.WithLabels(_name).Set(_currentTokens);
                RateLimitAllowed.WithLabels(_name).Inc();
                return true;
            }
        }

        RateLimitRejected.WithLabels(_name).Inc();
        return false;
    }

    /// <summary>
    /// Waits for a token to become available
    /// </summary>
    public async Task<bool> WaitAsync(CancellationToken cancellationToken = default)
    {
        RateLimitRequests.WithLabels(_name).Inc();

        while (!cancellationToken.IsCancellationRequested)
        {
            lock (_lock)
            {
                if (_currentTokens > 0)
                {
                    _currentTokens--;
                    RateLimitTokensAvailable.WithLabels(_name).Set(_currentTokens);
                    RateLimitAllowed.WithLabels(_name).Inc();
                    return true;
                }
            }

            await Task.Delay(TimeSpan.FromMilliseconds(10), cancellationToken);
        }

        RateLimitRejected.WithLabels(_name).Inc();
        return false;
    }
}

/// <summary>
/// Bulkhead for limiting concurrent operations
/// </summary>
public class Bulkhead
{
    private readonly SemaphoreSlim _semaphore;
    private readonly string _name;
    private readonly TimeSpan _timeout;

    private static readonly Gauge BulkheadActive = Prometheus.Metrics.CreateGauge(
        "bulkhead_active_requests",
        "Number of active requests in bulkhead",
        new GaugeConfiguration { LabelNames = new[] { "name" } });

    private static readonly Counter BulkheadRejected = Prometheus.Metrics.CreateCounter(
        "bulkhead_rejected_total",
        "Total requests rejected by bulkhead",
        new CounterConfiguration { LabelNames = new[] { "name", "reason" } });

    /// <summary>
    /// Creates a new bulkhead
    /// </summary>
    public Bulkhead(string name, int maxConcurrency, TimeSpan timeout)
    {
        _name = name;
        _timeout = timeout;
        _semaphore = new SemaphoreSlim(maxConcurrency, maxConcurrency);
    }

    /// <summary>
    /// Executes an action within the bulkhead
    /// </summary>
    public async Task ExecuteAsync(Func<Task> action, CancellationToken cancellationToken = default)
    {
        if (!await _semaphore.WaitAsync(_timeout, cancellationToken))
        {
            BulkheadRejected.WithLabels(_name, "timeout").Inc();
            throw new TimeoutException($"Bulkhead '{_name}' timed out waiting for slot");
        }

        BulkheadActive.WithLabels(_name).Inc();
        try
        {
            await action();
        }
        finally
        {
            BulkheadActive.WithLabels(_name).Dec();
            _semaphore.Release();
        }
    }

    /// <summary>
    /// Executes a function within the bulkhead
    /// </summary>
    public async Task<T> ExecuteAsync<T>(Func<Task<T>> action, CancellationToken cancellationToken = default)
    {
        if (!await _semaphore.WaitAsync(_timeout, cancellationToken))
        {
            BulkheadRejected.WithLabels(_name, "timeout").Inc();
            throw new TimeoutException($"Bulkhead '{_name}' timed out waiting for slot");
        }

        BulkheadActive.WithLabels(_name).Inc();
        try
        {
            return await action();
        }
        finally
        {
            BulkheadActive.WithLabels(_name).Dec();
            _semaphore.Release();
        }
    }
}
