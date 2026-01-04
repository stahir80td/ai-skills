using Core.Logger;

namespace Core.Tests;

public class LoggerTests
{
    [Fact]
    public void NewProduction_CreatesLogger_WithProductionSettings()
    {
        // Act
        using var logger = ServiceLogger.NewProduction("test-service", "1.0.0");

        // Assert - no exception thrown
        Assert.NotNull(logger);
    }

    [Fact]
    public void NewDevelopment_CreatesLogger_WithDevSettings()
    {
        // Act
        using var logger = ServiceLogger.NewDevelopment("test-service", "1.0.0");

        // Assert - no exception thrown
        Assert.NotNull(logger);
    }

    [Fact]
    public void WithContext_CreatesContextLogger()
    {
        // Arrange
        using var logger = ServiceLogger.NewDevelopment("test-service", "1.0.0");

        // Act
        var contextLogger = logger.WithContext("test-correlation-id", "TestComponent");

        // Assert
        Assert.NotNull(contextLogger);
    }

    [Fact]
    public void WithComponent_CreatesContextLogger()
    {
        // Arrange
        using var logger = ServiceLogger.NewDevelopment("test-service", "1.0.0");

        // Act
        var contextLogger = logger.WithComponent("TestComponent");

        // Assert
        Assert.NotNull(contextLogger);
    }

    [Fact]
    public void LoggerConfig_DefaultValues_AreSet()
    {
        // Arrange
        var config = new LoggerConfig();

        // Assert
        Assert.Equal("unknown-service", config.ServiceName);
        Assert.Equal("development", config.Environment);
        Assert.Equal("1.0.0", config.Version);
        Assert.Equal("Information", config.LogLevel);
        Assert.True(config.EnableCaller);
        Assert.True(config.EnableStacktrace);
        Assert.Equal("json", config.OutputFormat);
    }

    [Fact]
    public void Logger_Logs_WithoutException()
    {
        // Arrange
        using var logger = ServiceLogger.NewDevelopment("test-service", "1.0.0");

        // Act & Assert - should not throw
        logger.Debug("Debug message");
        logger.Information("Info message");
        logger.Warning("Warning message");
        logger.Error("Error message");
    }

    [Fact]
    public void ContextLogger_Logs_WithoutException()
    {
        // Arrange
        using var logger = ServiceLogger.NewDevelopment("test-service", "1.0.0");
        var contextLogger = logger.WithContext("corr-123", "TestComponent");

        // Act & Assert - should not throw
        contextLogger.Debug("Debug message");
        contextLogger.Information("Info message");
        contextLogger.Warning("Warning message");
        contextLogger.Error("Error message");
    }
}
