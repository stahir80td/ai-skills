using AiPatterns.Domain.Models;

namespace AiPatterns.Domain.Interfaces;

/// <summary>
/// Telemetry repository interface demonstrating ScyllaDB via Core.Infrastructure
/// </summary>
public interface ITelemetryRepository
{
    Task<IEnumerable<DeviceTelemetry>> GetByDeviceIdAsync(string deviceId, DateTime? startTime = null, DateTime? endTime = null);
    Task<IEnumerable<DeviceTelemetry>> GetByTimeRangeAsync(DateTime startTime, DateTime endTime);
    Task<DeviceTelemetry> InsertAsync(DeviceTelemetry telemetry);
    Task InsertBatchAsync(IEnumerable<DeviceTelemetry> telemetryBatch);
    Task<double?> GetLatestValueAsync(string deviceId, string metric);
    Task<IEnumerable<DeviceTelemetry>> GetAggregatedDataAsync(string deviceId, string metric, TimeSpan window, DateTime startTime, DateTime endTime);
    Task<long> CountRecordsAsync(string deviceId, DateTime? startTime = null, DateTime? endTime = null);
    Task DeleteOldDataAsync(DateTime beforeTime);
}