using Core.Sod;

namespace Core.Tests;

public class SodTests
{
    [Fact]
    public void SodCalculator_Creates()
    {
        // Act
        var calculator = new SodCalculator();

        // Assert
        Assert.NotNull(calculator);
    }

    [Fact]
    public void SodCalculator_CalculateScore_ReturnsValidScore()
    {
        // Arrange
        var calculator = new SodCalculator();
        calculator.RegisterError("TEST-001", new ErrorConfig
        {
            BaseSeverity = 5,
            BaseOccurrence = 5,
            BaseDetect = 5
        });

        var context = new ErrorContext
        {
            Environment = "production",
            IsBusinessHours = false,
            SystemLoad = 50
        };

        // Act
        var score = calculator.CalculateScore("TEST-001", context);

        // Assert
        Assert.True(score.Total >= 0 && score.Total <= 1000);
        Assert.True(score.Severity >= 1 && score.Severity <= 10);
        Assert.True(score.Occurrence >= 1 && score.Occurrence <= 10);
        Assert.True(score.Detect >= 1 && score.Detect <= 10);
    }

    [Fact]
    public void SodCalculator_ProductionEnvironment_IncreasesScore()
    {
        // Arrange
        var calculator = new SodCalculator();
        calculator.RegisterError("TEST-001", new ErrorConfig
        {
            BaseSeverity = 5,
            BaseOccurrence = 5,
            BaseDetect = 5
        });

        var prodContext = new ErrorContext { Environment = "production" };
        var devContext = new ErrorContext { Environment = "dev" };

        // Act
        var prodScore = calculator.CalculateScore("TEST-001", prodContext);
        var devScore = calculator.CalculateScore("TEST-001", devContext);

        // Assert - production should have higher adjustment factor
        Assert.True(prodScore.AdjustmentFactor > devScore.AdjustmentFactor);
    }

    [Fact]
    public void SodCalculator_BusinessHours_IncreasesScore()
    {
        // Arrange
        var calculator = new SodCalculator();
        calculator.RegisterError("TEST-001", new ErrorConfig());

        var bizHours = new ErrorContext { Environment = "production", IsBusinessHours = true };
        var nonBizHours = new ErrorContext { Environment = "production", IsBusinessHours = false };

        // Act
        var bizScore = calculator.CalculateScore("TEST-001", bizHours);
        var nonBizScore = calculator.CalculateScore("TEST-001", nonBizHours);

        // Assert
        Assert.True(bizScore.Total > nonBizScore.Total);
    }

    [Fact]
    public void SodCalculator_HighLoad_IncreasesScore()
    {
        // Arrange
        var calculator = new SodCalculator();
        calculator.RegisterError("TEST-001", new ErrorConfig());

        var highLoad = new ErrorContext { Environment = "staging", SystemLoad = 90 };
        var lowLoad = new ErrorContext { Environment = "staging", SystemLoad = 20 };

        // Act
        var highScore = calculator.CalculateScore("TEST-001", highLoad);
        var lowScore = calculator.CalculateScore("TEST-001", lowLoad);

        // Assert
        Assert.True(highScore.Total > lowScore.Total);
    }

    [Fact]
    public void SodCalculator_DataLoss_IncreasesSeverity()
    {
        // Arrange
        var calculator = new SodCalculator();
        calculator.RegisterError("TEST-001", new ErrorConfig { BaseSeverity = 5 });

        var withDataLoss = new ErrorContext { Environment = "staging", DataLossPotential = true };
        var noDataLoss = new ErrorContext { Environment = "staging", DataLossPotential = false };

        // Act
        var dataLossScore = calculator.CalculateScore("TEST-001", withDataLoss);
        var noDataLossScore = calculator.CalculateScore("TEST-001", noDataLoss);

        // Assert
        Assert.True(dataLossScore.Severity > noDataLossScore.Severity);
    }

    [Fact]
    public void SodCalculator_HighErrorRate_IncreasesOccurrence()
    {
        // Arrange
        var calculator = new SodCalculator();
        calculator.RegisterError("TEST-001", new ErrorConfig { BaseOccurrence = 3 });

        var highRate = new ErrorContext { Environment = "staging", RecentErrorRate = 15 };
        var lowRate = new ErrorContext { Environment = "staging", RecentErrorRate = 0.5 };

        // Act
        var highRateScore = calculator.CalculateScore("TEST-001", highRate);
        var lowRateScore = calculator.CalculateScore("TEST-001", lowRate);

        // Assert
        Assert.True(highRateScore.Occurrence > lowRateScore.Occurrence);
    }

    [Fact]
    public void SodCalculator_NoMonitoring_MaxDetectability()
    {
        // Arrange
        var calculator = new SodCalculator();
        calculator.RegisterError("TEST-001", new ErrorConfig
        {
            BaseDetect = 3,
            MonitoringEnabled = false
        });

        // Act
        var score = calculator.CalculateScore("TEST-001", new ErrorContext { Environment = "staging" });

        // Assert - without monitoring, detectability should be 10 (hardest to detect)
        Assert.Equal(10, score.Detect);
    }

    [Fact]
    public void SodCalculator_UnregisteredError_UsesDefaults()
    {
        // Arrange
        var calculator = new SodCalculator();
        var context = new ErrorContext { Environment = "staging" };

        // Act
        var score = calculator.CalculateScore("UNKNOWN-ERROR", context);

        // Assert - should not throw and should use default config
        Assert.True(score.Total > 0);
    }

    [Fact]
    public void SodScore_Normalized_IsBetweenZeroAndOne()
    {
        // Arrange
        var score = new SodScore { Total = 500 };

        // Assert
        Assert.Equal(0.5, score.Normalized);
    }

    [Fact]
    public void ErrorContext_DefaultValues()
    {
        // Arrange
        var context = new ErrorContext();

        // Assert
        Assert.Equal("production", context.Environment);
        Assert.False(context.IsBusinessHours);
        Assert.Equal(0, context.SystemLoad);
        Assert.Equal(0, context.RecentErrorRate);
        Assert.Equal(0, context.AffectedUsers);
        Assert.False(context.DataLossPotential);
    }
}
