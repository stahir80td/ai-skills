namespace AiPatterns.Domain.Models;

/// <summary>
/// Device telemetry for ScyllaDB storage - demonstrates time-series data
/// </summary>
public class DeviceTelemetry
{
    public string DeviceId { get; set; } = string.Empty;
    public DateTime Timestamp { get; set; }
    public string Metric { get; set; } = string.Empty;
    public double Value { get; set; }
    public string Unit { get; set; } = string.Empty;
    public Dictionary<string, string> Tags { get; set; } = new();
    public string Location { get; set; } = string.Empty;
    public string DataType { get; set; } = string.Empty; // temperature, pressure, vibration, etc.
    public double? MinValue { get; set; }
    public double? MaxValue { get; set; }
    public string Quality { get; set; } = "good"; // good, bad, uncertain
    public Guid CorrelationId { get; set; }

    public static DeviceTelemetry Create(string deviceId, string metric, double value, string unit, string dataType = "sensor")
    {
        return new DeviceTelemetry
        {
            DeviceId = deviceId,
            Timestamp = DateTime.UtcNow,
            Metric = metric,
            Value = value,
            Unit = unit,
            DataType = dataType,
            CorrelationId = Guid.NewGuid()
        };
    }

    public DeviceTelemetry WithLocation(string location)
    {
        Location = location;
        return this;
    }

    public DeviceTelemetry WithTags(Dictionary<string, string> tags)
    {
        Tags = tags;
        return this;
    }

    public DeviceTelemetry WithValidationRange(double minValue, double maxValue)
    {
        MinValue = minValue;
        MaxValue = maxValue;

        // Set quality based on range
        if (Value < minValue || Value > maxValue)
        {
            Quality = "bad";
        }

        return this;
    }
}

/// <summary>
/// Aggregated telemetry data for time-window queries
/// </summary>
public class AggregatedTelemetry
{
    public string DeviceId { get; set; } = string.Empty;
    public string Metric { get; set; } = string.Empty;
    public DateTime WindowStart { get; set; }
    public DateTime WindowEnd { get; set; }
    public TimeSpan WindowSize { get; set; }
    public long Count { get; set; }
    public double Average { get; set; }
    public double Min { get; set; }
    public double Max { get; set; }
    public double Sum { get; set; }
    public double StdDev { get; set; }
    public string Unit { get; set; } = string.Empty;
}

/// <summary>
/// Device metadata
/// </summary>
public class Device
{
    public string DeviceId { get; set; } = string.Empty;
    public string DeviceName { get; set; } = string.Empty;
    public string DeviceType { get; set; } = string.Empty;
    public string Location { get; set; } = string.Empty;
    public string Manufacturer { get; set; } = string.Empty;
    public string Model { get; set; } = string.Empty;
    public string FirmwareVersion { get; set; } = string.Empty;
    public DateTime LastSeen { get; set; }
    public bool IsOnline { get; set; }
    public Dictionary<string, string> Properties { get; set; } = new();
    public DateTime CreatedAt { get; set; }
    public DateTime UpdatedAt { get; set; }

    public static Device Create(string deviceId, string deviceName, string deviceType, string location)
    {
        return new Device
        {
            DeviceId = deviceId,
            DeviceName = deviceName,
            DeviceType = deviceType,
            Location = location,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow,
            LastSeen = DateTime.UtcNow,
            IsOnline = true
        };
    }

    public void UpdateLastSeen()
    {
        LastSeen = DateTime.UtcNow;
        IsOnline = true;
        UpdatedAt = DateTime.UtcNow;
    }

    public void SetOffline()
    {
        IsOnline = false;
        UpdatedAt = DateTime.UtcNow;
    }
}