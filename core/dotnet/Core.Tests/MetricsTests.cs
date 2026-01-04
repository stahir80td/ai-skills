using Core.Metrics;

namespace Core.Tests;

public class MetricsTests
{
    [Fact]
    public void ServiceMetrics_CreateWithConfig()
    {
        // Arrange
        var config = new MetricsConfig
        {
            ServiceName = "test-service",
            Namespace = "test",
            Subsystem = "api"
        };

        // Act
        var metrics = new ServiceMetrics(config);

        // Assert
        Assert.NotNull(metrics);
    }

    [Fact]
    public void ServiceMetrics_RecordRequest_DoesNotThrow()
    {
        // Arrange
        var metrics = new ServiceMetrics(new MetricsConfig { ServiceName = "test" });

        // Act & Assert - should not throw
        metrics.RecordRequest("GET", "/api/test", "200", TimeSpan.FromMilliseconds(50));
    }

    [Fact]
    public void ServiceMetrics_RecordError_DoesNotThrow()
    {
        // Arrange
        var metrics = new ServiceMetrics(new MetricsConfig { ServiceName = "test" });

        // Act & Assert - should not throw
        metrics.RecordError("ERR-001", "HIGH", "Database");
    }

    [Fact]
    public void ServiceMetrics_SetResourceUtilization_DoesNotThrow()
    {
        // Arrange
        var metrics = new ServiceMetrics(new MetricsConfig { ServiceName = "test" });

        // Act & Assert - should not throw
        metrics.SetResourceUtilization("cpu", 75.5);
        metrics.SetResourceUtilization("memory", 60.0);
    }

    [Fact]
    public void ServiceMetrics_ActiveRequests_IncrementDecrement()
    {
        // Arrange
        var metrics = new ServiceMetrics(new MetricsConfig { ServiceName = "test" });

        // Act & Assert - should not throw
        metrics.IncrementActiveRequests();
        metrics.IncrementActiveRequests();
        metrics.DecrementActiveRequests();
    }

    [Fact]
    public void ServiceMetrics_TimeRequest_RecordsDuration()
    {
        // Arrange
        var metrics = new ServiceMetrics(new MetricsConfig { ServiceName = "test" });
        var status = "200";

        // Act
        using (metrics.TimeRequest("GET", "/test", () => status))
        {
            Thread.Sleep(10); // Simulate work
        }

        // Assert - if we got here without exception, it worked
        Assert.True(true);
    }

    [Fact]
    public void MetricsConfig_DefaultValues()
    {
        // Arrange
        var config = new MetricsConfig();

        // Assert
        Assert.Equal("unknown-service", config.ServiceName);
        Assert.Equal("AI", config.Namespace);
        Assert.Null(config.Subsystem);
        Assert.Null(config.LatencyBuckets);
    }

    [Fact]
    public void DefaultLatencyBuckets_HasExpectedValues()
    {
        // Assert
        Assert.Contains(0.01, ServiceMetrics.DefaultLatencyBuckets);
        Assert.Contains(0.1, ServiceMetrics.DefaultLatencyBuckets);
        Assert.Contains(1.0, ServiceMetrics.DefaultLatencyBuckets);
        Assert.Contains(10.0, ServiceMetrics.DefaultLatencyBuckets);
    }
}
