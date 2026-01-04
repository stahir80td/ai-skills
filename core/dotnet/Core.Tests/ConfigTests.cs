using Core.Config;

namespace Core.Tests;

public class ConfigTests
{
    [Fact]
    public void TimeoutConfig_DefaultValues_MeetMinimums()
    {
        // Arrange
        var config = new TimeoutConfig();

        // Assert
        Assert.Equal(TimeSpan.FromSeconds(60), config.RedisTimeout);
        Assert.Equal(TimeSpan.FromSeconds(60), config.SqlTimeout);
        Assert.Equal(TimeSpan.FromSeconds(60), config.MongoDbTimeout);
        Assert.Equal(TimeSpan.FromSeconds(60), config.KafkaTimeout);
        Assert.Equal(TimeSpan.FromSeconds(60), config.HealthCheckTimeout);
        Assert.Equal(TimeSpan.FromSeconds(30), config.HttpTimeout);
        Assert.Equal(TimeSpan.FromSeconds(30), config.ShutdownTimeout);
    }

    [Fact]
    public void TimeoutConfig_Validate_PassesWithDefaults()
    {
        // Arrange
        var config = new TimeoutConfig();

        // Act & Assert - should not throw
        config.Validate();
    }

    [Fact]
    public void TimeoutConfig_Validate_FailsForLowRedisTimeout()
    {
        // Arrange
        var config = new TimeoutConfig { RedisTimeout = TimeSpan.FromSeconds(30) };

        // Act
        var result = config.Validate();

        // Assert
        Assert.False(result.IsValid);
        Assert.Contains(result.Errors, error => error.Contains("RedisTimeout must be at least 60 seconds"));
    }

    [Fact]
    public void TimeoutConfig_Validate_FailsForLowSqlTimeout()
    {
        // Arrange
        var config = new TimeoutConfig { SqlTimeout = TimeSpan.FromSeconds(30) };

        // Act
        var result = config.Validate();

        // Assert
        Assert.False(result.IsValid);
        Assert.Contains(result.Errors, error => error.Contains("SqlTimeout must be at least 60 seconds"));
    }

    [Fact]
    public void ServiceConfig_DefaultValues()
    {
        // Arrange
        var config = new ServiceConfig();

        // Assert
        Assert.Equal("unknown-service", config.ServiceName);
        Assert.Equal("1.0.0", config.Version);
        Assert.Equal("development", config.Environment);
        Assert.False(config.EnableDebugLogging);
        Assert.NotNull(config.Timeouts);
    }

    [Fact]
    public void ServiceConfig_IsProduction_ReturnsTrueForProduction()
    {
        // Arrange
        var config = new ServiceConfig { Environment = "production" };

        // Assert
        Assert.True(config.IsProduction);
    }

    [Fact]
    public void ServiceConfig_IsProduction_ReturnsFalseForDevelopment()
    {
        // Arrange
        var config = new ServiceConfig { Environment = "development" };

        // Assert
        Assert.False(config.IsProduction);
    }

    [Theory]
    [InlineData("production", true)]
    [InlineData("Production", true)]
    [InlineData("PRODUCTION", true)]
    [InlineData("staging", false)]
    [InlineData("development", false)]
    [InlineData("dev", false)]
    public void ServiceConfig_IsProduction_CaseInsensitive(string env, bool expected)
    {
        var config = new ServiceConfig { Environment = env };
        Assert.Equal(expected, config.IsProduction);
    }
}
