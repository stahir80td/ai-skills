using Core.Errors;

namespace Core.Tests;

public class ErrorsTests
{
    [Fact]
    public void ServiceError_CreatesWithCodeAndSeverity()
    {
        // Act
        var error = new ServiceError("TEST-001", Severity.High, "Test error message");

        // Assert
        Assert.Equal("TEST-001", error.Code);
        Assert.Equal(Severity.High, error.ErrorSeverity);
        Assert.Equal("Test error message", error.Message);
    }

    [Fact]
    public void ServiceError_WithContext_AddsContext()
    {
        // Arrange
        var error = new ServiceError("TEST-001", Severity.Medium, "Test error");

        // Act
        error.WithContext("user_id", "123")
             .WithContext("request_id", "req-456");

        // Assert
        Assert.Equal("123", error.GetContext<string>("user_id"));
        Assert.Equal("req-456", error.GetContext<string>("request_id"));
    }

    [Fact]
    public void ServiceError_ToString_IncludesCodeAndSeverity()
    {
        // Arrange
        var error = new ServiceError("TEST-001", Severity.Critical, "Critical failure");

        // Act
        var result = error.ToString();

        // Assert
        Assert.Contains("TEST-001", result);
        Assert.Contains("CRITICAL", result);
        Assert.Contains("Critical failure", result);
    }

    [Fact]
    public void ServiceError_WithInnerException_IncludesCausedBy()
    {
        // Arrange
        var inner = new Exception("Inner exception");
        var error = new ServiceError("TEST-002", Severity.High, "Outer error", inner);

        // Act
        var result = error.ToString();

        // Assert
        Assert.Contains("caused by", result);
        Assert.Contains("Inner exception", result);
    }

    [Fact]
    public void ErrorRegistry_RegistersAndRetrievesDefinitions()
    {
        // Arrange
        var registry = new ErrorRegistry();
        var definition = new ErrorDefinition
        {
            Code = "DB-001",
            Severity = Severity.High,
            Description = "Database connection failed",
            SeverityScore = 8,
            OccurrenceScore = 3,
            DetectabilityScore = 2
        };

        // Act
        registry.Register(definition);
        var retrieved = registry.Get("DB-001");

        // Assert
        Assert.NotNull(retrieved);
        Assert.Equal("DB-001", retrieved.Code);
        Assert.Equal(48, retrieved.SODScore); // 8 × 3 × 2 = 48
    }

    [Fact]
    public void ErrorRegistry_CreateError_UsesDefinition()
    {
        // Arrange
        var registry = new ErrorRegistry();
        registry.Register(new ErrorDefinition
        {
            Code = "API-001",
            Severity = Severity.Medium,
            Description = "API request to {0} failed"
        });

        // Act
        var error = registry.CreateError("API-001", "external-service");

        // Assert
        Assert.Equal("API-001", error.Code);
        Assert.Equal(Severity.Medium, error.ErrorSeverity);
        Assert.Contains("external-service", error.Message);
    }

    [Fact]
    public void ErrorRegistry_WrapError_PreservesInnerException()
    {
        // Arrange
        var registry = new ErrorRegistry();
        registry.Register(new ErrorDefinition
        {
            Code = "NET-001",
            Severity = Severity.High,
            Description = "Network failure"
        });
        var innerEx = new Exception("Connection refused");

        // Act
        var error = registry.WrapError(innerEx, "NET-001");

        // Assert
        Assert.Equal("NET-001", error.Code);
        Assert.NotNull(error.InnerException);
        Assert.Equal("Connection refused", error.InnerException.Message);
    }

    [Fact]
    public void SodCalculator_Calculate_ReturnsCorrectScore()
    {
        // Act
        var score = SodCalculator.Calculate(5, 3, 2);

        // Assert
        Assert.Equal(30, score);
    }

    [Theory]
    [InlineData(1, 1, 1, 1)]
    [InlineData(10, 10, 10, 1000)]
    [InlineData(5, 5, 5, 125)]
    public void SodCalculator_Calculate_VariousInputs(int s, int o, int d, int expected)
    {
        Assert.Equal(expected, SodCalculator.Calculate(s, o, d));
    }
}
