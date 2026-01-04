using Core.Sli;

namespace Core.Tests;

public class SliTests
{
    [Fact]
    public void PrometheusSliTracker_Creates()
    {
        // Act
        var tracker = new PrometheusSliTracker("test-service");

        // Assert
        Assert.NotNull(tracker);
    }

    [Fact]
    public void PrometheusSliTracker_RecordRequest_Success()
    {
        // Arrange
        var tracker = new PrometheusSliTracker("test-service");
        var outcome = new RequestOutcome
        {
            Operation = "GetUser",
            Success = true,
            Latency = TimeSpan.FromMilliseconds(50)
        };

        // Act & Assert - should not throw
        tracker.RecordRequest(outcome);
    }

    [Fact]
    public void PrometheusSliTracker_RecordRequest_Failure()
    {
        // Arrange
        var tracker = new PrometheusSliTracker("test-service");
        var outcome = new RequestOutcome
        {
            Operation = "GetUser",
            Success = false,
            Latency = TimeSpan.FromMilliseconds(100),
            ErrorCode = "USER-404",
            ErrorSeverity = "MEDIUM"
        };

        // Act & Assert - should not throw
        tracker.RecordRequest(outcome);
    }

    [Fact]
    public void PrometheusSliTracker_RecordLatency()
    {
        // Arrange
        var tracker = new PrometheusSliTracker("test-service");

        // Act & Assert - should not throw
        tracker.RecordLatency(TimeSpan.FromMilliseconds(150), "ProcessOrder");
    }

    [Fact]
    public void PrometheusSliTracker_RecordThroughput()
    {
        // Arrange
        var tracker = new PrometheusSliTracker("test-service");

        // Act & Assert - should not throw
        tracker.RecordThroughput(100, "MessageProcessing");
    }

    [Fact]
    public void RequestOutcome_DefaultOperation_IsDefault()
    {
        // Arrange
        var outcome = new RequestOutcome();

        // Assert
        Assert.Equal("default", outcome.Operation);
    }

    [Fact]
    public void SliBudget_CalculatesAllowedDowntime()
    {
        // Arrange - 99.9% availability over 30 days
        var budget = new SliBudget("test-service", 99.9, TimeSpan.FromDays(30));

        // Assert - 0.1% of 30 days = 43.2 minutes
        var expectedDowntime = TimeSpan.FromDays(30) * 0.001;
        Assert.Equal(expectedDowntime, budget.AllowedDowntime);
    }

    [Fact]
    public void SliBudget_RemainingBudget_CalculatesCorrectly()
    {
        // Arrange
        var budget = new SliBudget("test-service", 99.9, TimeSpan.FromDays(30));
        var downtimeUsed = TimeSpan.FromMinutes(20);

        // Act
        var remaining = budget.RemainingBudget(downtimeUsed);

        // Assert
        Assert.Equal(budget.AllowedDowntime - downtimeUsed, remaining);
    }

    [Fact]
    public void SliBudget_BurnRate_AtExpectedPace()
    {
        // Arrange
        var budget = new SliBudget("test-service", 99.9, TimeSpan.FromDays(30));
        // If we've used proportional downtime to elapsed time, burn rate should be 1.0
        var elapsed = TimeSpan.FromDays(15); // Half the window
        var expectedBurn = budget.AllowedDowntime / 2; // Half the budget

        // Act
        var burnRate = budget.BurnRate(expectedBurn, elapsed);

        // Assert - should be approximately 1.0
        Assert.Equal(1.0, burnRate, precision: 2);
    }

    [Fact]
    public void SliBudget_BurnRate_DoublePace()
    {
        // Arrange
        var budget = new SliBudget("test-service", 99.9, TimeSpan.FromDays(30));
        var elapsed = TimeSpan.FromDays(15); // Half the window
        var usedDouble = budget.AllowedDowntime; // Used entire budget in half time

        // Act
        var burnRate = budget.BurnRate(usedDouble, elapsed);

        // Assert - should be approximately 2.0
        Assert.Equal(2.0, burnRate, precision: 2);
    }

    [Fact]
    public void SliMetrics_DefaultValues()
    {
        // Arrange
        var metrics = new SliMetrics();

        // Assert
        Assert.Equal(0, metrics.Availability);
        Assert.Equal(0, metrics.LatencyP95Ms);
        Assert.Equal(0, metrics.ErrorRate);
        Assert.Equal(0, metrics.Throughput);
    }
}
