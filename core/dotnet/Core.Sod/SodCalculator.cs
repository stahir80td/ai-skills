using Prometheus;

namespace Core.Sod;

/// <summary>
/// SOD Score result
/// </summary>
public class SodScore
{
    /// <summary>Total SOD score (S × O × D)</summary>
    public int Total { get; set; }

    /// <summary>Severity score (1-10)</summary>
    public int Severity { get; set; }

    /// <summary>Occurrence score (1-10)</summary>
    public int Occurrence { get; set; }

    /// <summary>Detectability score (1-10)</summary>
    public int Detect { get; set; }

    /// <summary>Normalized score (0-1)</summary>
    public double Normalized => Total / 1000.0;

    /// <summary>Base score before adjustments</summary>
    public int BaseScore { get; set; }

    /// <summary>Score after runtime adjustments</summary>
    public int AdjustedScore { get; set; }

    /// <summary>Adjustment factor applied</summary>
    public double AdjustmentFactor { get; set; }

    /// <summary>Reason for severity score</summary>
    public string? SeverityReason { get; set; }

    /// <summary>Reason for occurrence score</summary>
    public string? OccurrenceReason { get; set; }

    /// <summary>Reason for detectability score</summary>
    public string? DetectReason { get; set; }
}

/// <summary>
/// Error context for SOD calculation
/// </summary>
public class ErrorContext
{
    /// <summary>Environment (production, staging, dev)</summary>
    public string Environment { get; set; } = "production";

    /// <summary>Whether it's business hours</summary>
    public bool IsBusinessHours { get; set; }

    /// <summary>Current system load (0-100)</summary>
    public double SystemLoad { get; set; }

    /// <summary>Recent error rate for this error type</summary>
    public double RecentErrorRate { get; set; }

    /// <summary>Number of affected users</summary>
    public int AffectedUsers { get; set; }

    /// <summary>Whether there's data loss potential</summary>
    public bool DataLossPotential { get; set; }
}

/// <summary>
/// Error configuration for SOD calculation
/// </summary>
public class ErrorConfig
{
    /// <summary>Base severity score (1-10)</summary>
    public int BaseSeverity { get; set; } = 5;

    /// <summary>Base occurrence score (1-10)</summary>
    public int BaseOccurrence { get; set; } = 5;

    /// <summary>Base detectability score (1-10)</summary>
    public int BaseDetect { get; set; } = 5;

    /// <summary>Whether monitoring is enabled</summary>
    public bool MonitoringEnabled { get; set; } = true;

    /// <summary>Whether alerting is enabled</summary>
    public bool AlertingEnabled { get; set; } = true;

    /// <summary>Whether auto-detection is enabled</summary>
    public bool AutoDetect { get; set; }
}

/// <summary>
/// Load thresholds for runtime adjustments
/// </summary>
public class LoadThresholds
{
    public double Low { get; set; } = 30;
    public double Medium { get; set; } = 60;
    public double High { get; set; } = 80;
}

/// <summary>
/// SOD Calculator for Symptom-Oriented Diagnosis
/// </summary>
public class SodCalculator
{
    private readonly Dictionary<string, ErrorConfig> _errorConfigs = new();
    private readonly LoadThresholds _loadThresholds = new();

    private static readonly Gauge SodScoreGauge = Prometheus.Metrics.CreateGauge(
        "sod_score",
        "SOD score for error",
        new GaugeConfiguration { LabelNames = new[] { "error_code" } });

    private static readonly Counter ErrorOccurrences = Prometheus.Metrics.CreateCounter(
        "sod_error_occurrences_total",
        "Total error occurrences tracked by SOD",
        new CounterConfiguration { LabelNames = new[] { "error_code", "severity" } });

    /// <summary>
    /// Registers an error configuration
    /// </summary>
    public void RegisterError(string errorCode, ErrorConfig config)
    {
        _errorConfigs[errorCode] = config;
    }

    /// <summary>
    /// Calculates SOD score for an error
    /// </summary>
    public SodScore CalculateScore(string errorCode, ErrorContext context)
    {
        if (!_errorConfigs.TryGetValue(errorCode, out var config))
        {
            config = new ErrorConfig(); // Use defaults
        }

        // Calculate base scores
        var severity = CalculateSeverity(config, context);
        var occurrence = CalculateOccurrence(config, context);
        var detect = CalculateDetect(config);

        // Calculate total
        var baseScore = severity * occurrence * detect;

        // Apply runtime adjustments
        var adjustmentFactor = CalculateAdjustmentFactor(context);
        var adjustedScore = (int)(baseScore * adjustmentFactor);

        // Clamp to 0-1000
        adjustedScore = Math.Clamp(adjustedScore, 0, 1000);

        var score = new SodScore
        {
            Total = adjustedScore,
            Severity = severity,
            Occurrence = occurrence,
            Detect = detect,
            BaseScore = baseScore,
            AdjustedScore = adjustedScore,
            AdjustmentFactor = adjustmentFactor,
            SeverityReason = GetSeverityReason(context),
            OccurrenceReason = GetOccurrenceReason(context),
            DetectReason = GetDetectReason(config)
        };

        // Record metrics
        SodScoreGauge.WithLabels(errorCode).Set(adjustedScore);
        ErrorOccurrences.WithLabels(errorCode, SeverityToString(severity)).Inc();

        return score;
    }

    private int CalculateSeverity(ErrorConfig config, ErrorContext context)
    {
        var severity = config.BaseSeverity;

        // Increase severity for data loss potential
        if (context.DataLossPotential)
        {
            severity = Math.Min(10, severity + 3);
        }

        // Increase severity based on affected users
        if (context.AffectedUsers > 1000)
        {
            severity = Math.Min(10, severity + 2);
        }
        else if (context.AffectedUsers > 100)
        {
            severity = Math.Min(10, severity + 1);
        }

        return Math.Clamp(severity, 1, 10);
    }

    private int CalculateOccurrence(ErrorConfig config, ErrorContext context)
    {
        var occurrence = config.BaseOccurrence;

        // Adjust based on recent error rate
        if (context.RecentErrorRate > 10)
        {
            occurrence = Math.Min(10, occurrence + 3);
        }
        else if (context.RecentErrorRate > 5)
        {
            occurrence = Math.Min(10, occurrence + 2);
        }
        else if (context.RecentErrorRate > 1)
        {
            occurrence = Math.Min(10, occurrence + 1);
        }

        return Math.Clamp(occurrence, 1, 10);
    }

    private int CalculateDetect(ErrorConfig config)
    {
        var detect = config.BaseDetect;

        if (!config.MonitoringEnabled)
        {
            detect = 10; // Hard to detect without monitoring
        }
        else if (!config.AlertingEnabled)
        {
            detect = Math.Max(detect, 7); // Harder without alerts
        }
        else if (config.AutoDetect)
        {
            detect = Math.Min(detect, 3); // Easy to detect with automation
        }

        return Math.Clamp(detect, 1, 10);
    }

    private double CalculateAdjustmentFactor(ErrorContext context)
    {
        var factor = 1.0;

        // Environment multiplier
        factor *= context.Environment.ToLowerInvariant() switch
        {
            "production" => 1.5,
            "staging" => 1.0,
            "dev" or "development" => 0.5,
            _ => 1.0
        };

        // Business hours multiplier
        if (context.IsBusinessHours)
        {
            factor *= 1.3;
        }

        // System load multiplier
        if (context.SystemLoad > _loadThresholds.High)
        {
            factor *= 1.4;
        }
        else if (context.SystemLoad > _loadThresholds.Medium)
        {
            factor *= 1.2;
        }

        return factor;
    }

    private string GetSeverityReason(ErrorContext context)
    {
        if (context.DataLossPotential)
            return "Data loss potential detected";
        if (context.AffectedUsers > 1000)
            return "High user impact (>1000 users)";
        if (context.AffectedUsers > 100)
            return "Moderate user impact (>100 users)";
        return "Base severity";
    }

    private string GetOccurrenceReason(ErrorContext context)
    {
        if (context.RecentErrorRate > 10)
            return "High error rate (>10%)";
        if (context.RecentErrorRate > 5)
            return "Elevated error rate (>5%)";
        if (context.RecentErrorRate > 1)
            return "Moderate error rate (>1%)";
        return "Base occurrence rate";
    }

    private string GetDetectReason(ErrorConfig config)
    {
        if (!config.MonitoringEnabled)
            return "No monitoring configured";
        if (!config.AlertingEnabled)
            return "No alerting configured";
        if (config.AutoDetect)
            return "Auto-detection enabled";
        return "Manual detection required";
    }

    private static string SeverityToString(int severity) => severity switch
    {
        >= 9 => "CRITICAL",
        >= 7 => "HIGH",
        >= 4 => "MEDIUM",
        >= 2 => "LOW",
        _ => "INFO"
    };
}
