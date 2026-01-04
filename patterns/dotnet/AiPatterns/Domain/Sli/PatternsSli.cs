using Core.Sli;
using Prometheus;

namespace AiPatterns.Domain.Sli;

/// <summary>
/// SLI tracker demonstrating AI patterns
/// Tracks availability, latency, throughput, and business metrics
/// </summary>
public class PatternsSli : ISliTracker
{
    private readonly PrometheusSliTracker _sliTracker;

    // Business metrics - products
    private static readonly Counter ProductsCreated = Metrics
        .CreateCounter("patterns_products_created_total", "Products created",
            new CounterConfiguration { LabelNames = new[] { "category", "status" } });

    private static readonly Gauge ActiveProducts = Metrics
        .CreateGauge("patterns_products_active", "Current active products count");

    private static readonly Histogram ProductPrice = Metrics
        .CreateHistogram("patterns_product_price_dollars", "Product price distribution",
            new HistogramConfiguration
            {
                LabelNames = new[] { "category" },
                Buckets = new[] { 10.0, 25.0, 50.0, 100.0, 250.0, 500.0, 1000.0, 2500.0, 5000.0 }
            });

    // Infrastructure metrics
    private static readonly Counter DatabaseOperations = Metrics
        .CreateCounter("patterns_database_operations_total", "Database operations",
            new CounterConfiguration { LabelNames = new[] { "operation", "table", "status" } });

    private static readonly Counter CacheOperations = Metrics
        .CreateCounter("patterns_cache_operations_total", "Cache operations",
            new CounterConfiguration { LabelNames = new[] { "operation", "status" } });

    private static readonly Counter ExternalCalls = Metrics
        .CreateCounter("patterns_external_calls_total", "External service calls",
            new CounterConfiguration { LabelNames = new[] { "service", "operation", "status" } });

    public PatternsSli()
    {
        _sliTracker = new PrometheusSliTracker("patterns-service");
    }

    public void RecordRequest(RequestOutcome outcome)
    {
        _sliTracker.RecordRequest(outcome);
    }

    public void RecordLatency(TimeSpan duration, string operation)
    {
        _sliTracker.RecordLatency(duration, operation);
    }

    public void RecordThroughput(int count, string operation)
    {
        _sliTracker.RecordThroughput(count, operation);
    }

    public void RecordProductCreated(string category, string status, decimal price)
    {
        ProductsCreated.WithLabels(category, status).Inc();
        ProductPrice.WithLabels(category).Observe((double)price);
        
        if (status == "active")
        {
            ActiveProducts.Inc();
        }
    }

    public void RecordProductStatusChanged(string oldStatus, string newStatus)
    {
        if (oldStatus != "active" && newStatus == "active")
        {
            ActiveProducts.Inc();
        }
        else if (oldStatus == "active" && newStatus != "active")
        {
            ActiveProducts.Dec();
        }
    }

    public void RecordDatabaseOperation(string operation, string table, bool success)
    {
        var status = success ? "success" : "error";
        DatabaseOperations.WithLabels(operation, table, status).Inc();
    }

    public void RecordCacheOperation(string operation, bool success)
    {
        var status = success ? "success" : "error";
        CacheOperations.WithLabels(operation, status).Inc();
    }

    public void RecordExternalCall(string service, string operation, bool success)
    {
        var status = success ? "success" : "error";
        ExternalCalls.WithLabels(service, operation, status).Inc();
    }
}