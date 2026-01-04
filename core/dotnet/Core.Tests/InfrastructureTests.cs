using Core.Infrastructure;

namespace Core.Tests;

public class InfrastructureTests
{
    [Fact]
    public void RedisConfig_DefaultValues()
    {
        // Arrange
        var config = new RedisConfig();

        // Assert
        Assert.Equal("localhost", config.Host);
        Assert.Equal(6379, config.Port);
        Assert.Equal(TimeSpan.FromSeconds(60), config.ConnectTimeout);
        Assert.Equal(TimeSpan.FromSeconds(5), config.OperationTimeout);
        Assert.Null(config.ConnectionString);
    }

    [Fact]
    public async Task RedisClient_Create_ThrowsForEmptyHost()
    {
        // Arrange
        var config = new RedisConfig { Host = "" };

        // Act & Assert
        await Assert.ThrowsAsync<ArgumentException>(() => RedisClient.CreateAsync(config));
    }

    [Fact]
    public async Task RedisClient_Create_ThrowsForInvalidPort()
    {
        // Arrange
        var config = new RedisConfig { Host = "localhost", Port = -1 };

        // Act & Assert
        await Assert.ThrowsAsync<ArgumentException>(() => RedisClient.CreateAsync(config));
    }

    [Fact]
    public void HealthChecker_Creates()
    {
        // Act
        var checker = new HealthChecker(TimeSpan.FromSeconds(60));

        // Assert
        Assert.NotNull(checker);
    }

    [Fact]
    public void HealthChecker_RegistersCheck()
    {
        // Arrange
        var checker = new HealthChecker(TimeSpan.FromSeconds(60));

        // Act - should not throw
        checker.Register("test", async _ =>
        {
            await Task.CompletedTask;
            return Microsoft.Extensions.Diagnostics.HealthChecks.HealthCheckResult.Healthy();
        });
    }

    [Fact]
    public async Task HealthChecker_RunsChecks()
    {
        // Arrange
        var checker = new HealthChecker(TimeSpan.FromSeconds(60));
        checker.Register("test1", async _ =>
        {
            await Task.CompletedTask;
            return Microsoft.Extensions.Diagnostics.HealthChecks.HealthCheckResult.Healthy("OK");
        });
        checker.Register("test2", async _ =>
        {
            await Task.CompletedTask;
            return Microsoft.Extensions.Diagnostics.HealthChecks.HealthCheckResult.Healthy("OK");
        });

        // Act
        var response = await checker.CheckAsync();

        // Assert
        Assert.Equal(HealthStatus.Healthy, response.Status);
        Assert.Equal(2, response.Checks.Count);
    }

    [Fact]
    public async Task HealthChecker_ReportsUnhealthy()
    {
        // Arrange
        var checker = new HealthChecker(TimeSpan.FromSeconds(60));
        checker.Register("healthy", async _ =>
        {
            await Task.CompletedTask;
            return Microsoft.Extensions.Diagnostics.HealthChecks.HealthCheckResult.Healthy();
        });
        checker.Register("unhealthy", async _ =>
        {
            await Task.CompletedTask;
            return Microsoft.Extensions.Diagnostics.HealthChecks.HealthCheckResult.Unhealthy("Failed");
        });

        // Act
        var response = await checker.CheckAsync();

        // Assert
        Assert.Equal(HealthStatus.Unhealthy, response.Status);
    }

    [Fact]
    public async Task HealthChecker_HandlesExceptions()
    {
        // Arrange
        var checker = new HealthChecker(TimeSpan.FromSeconds(60));
        checker.Register("failing", _ => throw new Exception("Test exception"));

        // Act
        var response = await checker.CheckAsync();

        // Assert
        Assert.Equal(HealthStatus.Unhealthy, response.Status);
        Assert.NotNull(response.Checks["failing"].Error);
    }

    [Fact]
    public async Task LivenessCheck_ReturnsHealthy()
    {
        // Arrange
        var check = new LivenessCheck();
        var context = new Microsoft.Extensions.Diagnostics.HealthChecks.HealthCheckContext();

        // Act
        var result = await check.CheckHealthAsync(context);

        // Assert
        Assert.Equal(Microsoft.Extensions.Diagnostics.HealthChecks.HealthStatus.Healthy, result.Status);
    }

    [Fact]
    public async Task ReadinessCheck_ReturnsBasedOnDelegate()
    {
        // Arrange
        var isReady = false;
        var check = new ReadinessCheck(() => isReady);
        var context = new Microsoft.Extensions.Diagnostics.HealthChecks.HealthCheckContext();

        // Act - not ready
        var result1 = await check.CheckHealthAsync(context);

        // Change state
        isReady = true;

        // Act - ready
        var result2 = await check.CheckHealthAsync(context);

        // Assert
        Assert.Equal(Microsoft.Extensions.Diagnostics.HealthChecks.HealthStatus.Unhealthy, result1.Status);
        Assert.Equal(Microsoft.Extensions.Diagnostics.HealthChecks.HealthStatus.Healthy, result2.Status);
    }

    [Fact]
    public void HealthCheckResponse_DefaultValues()
    {
        // Arrange
        var response = new HealthCheckResponse();

        // Assert
        Assert.Equal(HealthStatus.Healthy, response.Status); // default enum value
        Assert.NotNull(response.Checks);
        Assert.Empty(response.Checks);
    }

    [Fact]
    public void HealthCheckResultItem_RequiredProperties()
    {
        // Arrange & Act
        var result = new HealthCheckResultItem
        {
            Name = "test-check",
            Status = HealthStatus.Healthy,
            Duration = TimeSpan.FromMilliseconds(50)
        };

        // Assert
        Assert.Equal("test-check", result.Name);
        Assert.Equal(HealthStatus.Healthy, result.Status);
        Assert.Equal(TimeSpan.FromMilliseconds(50), result.Duration);
    }
}
