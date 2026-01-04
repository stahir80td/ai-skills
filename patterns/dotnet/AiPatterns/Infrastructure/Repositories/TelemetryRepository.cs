using Core.Infrastructure;
using Core.Logger;
using AiPatterns.Domain.Interfaces;
using AiPatterns.Domain.Models;

namespace AiPatterns.Infrastructure.Repositories;

/// <summary>
/// Telemetry repository using Core.Infrastructure.ScyllaDB - demonstrates time-series data patterns
/// ScyllaDB is optimal for high-throughput time-series data like IoT telemetry
/// </summary>
public class TelemetryRepository : ITelemetryRepository
{
    private readonly ScyllaDBClient _scyllaClient;
    private readonly ServiceLogger _logger;

    public TelemetryRepository(ScyllaDBClient scyllaClient, ServiceLogger logger)
    {
        _scyllaClient = scyllaClient;
        _logger = logger;
    }

    public async Task<IEnumerable<DeviceTelemetry>> GetByDeviceIdAsync(string deviceId, DateTime? startTime = null, DateTime? endTime = null)
    {
        var contextLogger = _logger.WithContext(component: "TelemetryRepository.GetByDeviceId");
        var start = startTime ?? DateTime.UtcNow.AddDays(-7);
        var end = endTime ?? DateTime.UtcNow;

        contextLogger.Debug("Fetching telemetry for device: {DeviceId}, Range: {Start} to {End}", deviceId, start, end);

        var cql = @"
            SELECT device_id, timestamp, metric, value, unit, quality, tags, correlation_id
            FROM telemetry
            WHERE device_id = ? AND timestamp >= ? AND timestamp <= ?
            ORDER BY timestamp DESC
            LIMIT 1000";

        var results = await _scyllaClient.QueryAsync<DeviceTelemetry>(cql, new { deviceId, start, end });
        var telemetryList = results.ToList();

        contextLogger.Information("Retrieved telemetry records: {Count} for device: {DeviceId}", telemetryList.Count, deviceId);
        return telemetryList;
    }

    public async Task<IEnumerable<DeviceTelemetry>> GetByTimeRangeAsync(DateTime startTime, DateTime endTime)
    {
        var contextLogger = _logger.WithContext(component: "TelemetryRepository.GetByTimeRange");
        contextLogger.Debug("Fetching telemetry by time range: {Start} to {End}", startTime, endTime);

        var cql = @"
            SELECT device_id, timestamp, metric, value, unit, quality, tags, correlation_id
            FROM telemetry
            WHERE timestamp >= ? AND timestamp <= ?
            ALLOW FILTERING
            LIMIT 10000";

        var results = await _scyllaClient.QueryAsync<DeviceTelemetry>(cql, new { startTime, endTime });
        var telemetryList = results.ToList();

        contextLogger.Information("Retrieved telemetry records by time range: {Count}", telemetryList.Count);
        return telemetryList;
    }

    public async Task<DeviceTelemetry> InsertAsync(DeviceTelemetry telemetry)
    {
        var contextLogger = _logger.WithContext(component: "TelemetryRepository.Insert");
        contextLogger.Debug("Inserting telemetry: {DeviceId}, {Timestamp}", telemetry.DeviceId, telemetry.Timestamp);

        var tagsJson = telemetry.Tags != null 
            ? System.Text.Json.JsonSerializer.Serialize(telemetry.Tags) 
            : null;

        var cql = @"
            INSERT INTO telemetry (device_id, timestamp, metric, value, unit, quality, tags, correlation_id)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)";

        var parameters = new
        {
            telemetry.DeviceId,
            telemetry.Timestamp,
            telemetry.Metric,
            telemetry.Value,
            telemetry.Unit,
            telemetry.Quality,
            Tags = tagsJson,
            telemetry.CorrelationId
        };

        await _scyllaClient.ExecuteAsync(cql, parameters);
        contextLogger.Information("Telemetry inserted: {DeviceId}, {Metric}", telemetry.DeviceId, telemetry.Metric);
        return telemetry;
    }

    public async Task InsertBatchAsync(IEnumerable<DeviceTelemetry> telemetryBatch)
    {
        var contextLogger = _logger.WithContext(component: "TelemetryRepository.InsertBatch");
        var batchList = telemetryBatch.ToList();
        contextLogger.Debug("Inserting telemetry batch: {Count} records", batchList.Count);

        foreach (var telemetry in batchList)
        {
            await InsertAsync(telemetry);
        }

        contextLogger.Information("Telemetry batch inserted: {Count} records", batchList.Count);
    }

    public async Task<double?> GetLatestValueAsync(string deviceId, string metric)
    {
        var contextLogger = _logger.WithContext(component: "TelemetryRepository.GetLatestValue");
        contextLogger.Debug("Fetching latest value for device: {DeviceId}, metric: {Metric}", deviceId, metric);

        var cql = @"
            SELECT value FROM telemetry
            WHERE device_id = ? AND metric = ?
            ORDER BY timestamp DESC
            LIMIT 1
            ALLOW FILTERING";

        var results = await _scyllaClient.QueryAsync<DeviceTelemetry>(cql, new { deviceId, metric });
        var latest = results.FirstOrDefault();

        if (latest != null)
        {
            contextLogger.Debug("Found latest value: {Value}", latest.Value);
            return latest.Value;
        }

        return null;
    }

    public async Task<IEnumerable<DeviceTelemetry>> GetAggregatedDataAsync(string deviceId, string metric, TimeSpan window, DateTime startTime, DateTime endTime)
    {
        var contextLogger = _logger.WithContext(component: "TelemetryRepository.GetAggregatedData");
        contextLogger.Debug("Fetching aggregated data for device: {DeviceId}, metric: {Metric}", deviceId, metric);

        // For ScyllaDB, aggregation is typically done at the application level or using materialized views
        // Here we return raw data; actual aggregation would be done by the caller
        var cql = @"
            SELECT device_id, timestamp, metric, value, unit, quality, tags, correlation_id
            FROM telemetry
            WHERE device_id = ? AND metric = ? AND timestamp >= ? AND timestamp <= ?
            ALLOW FILTERING
            LIMIT 10000";

        var results = await _scyllaClient.QueryAsync<DeviceTelemetry>(cql, new { deviceId, metric, startTime, endTime });
        return results;
    }

    public async Task<long> CountRecordsAsync(string deviceId, DateTime? startTime = null, DateTime? endTime = null)
    {
        var contextLogger = _logger.WithContext(component: "TelemetryRepository.CountRecords");
        var start = startTime ?? DateTime.UtcNow.AddDays(-30);
        var end = endTime ?? DateTime.UtcNow;

        contextLogger.Debug("Counting telemetry records for device: {DeviceId}", deviceId);

        // ScyllaDB COUNT is expensive; in production, use counters or pre-computed values
        var data = await GetByDeviceIdAsync(deviceId, start, end);
        var count = data.Count();

        contextLogger.Debug("Counted telemetry records: {Count}", count);
        return count;
    }

    public async Task DeleteOldDataAsync(DateTime beforeTime)
    {
        var contextLogger = _logger.WithContext(component: "TelemetryRepository.DeleteOldData");
        contextLogger.Information("Deleting telemetry data before: {BeforeTime}", beforeTime);

        // Note: ScyllaDB TTL is preferred for automatic data expiration
        var cql = @"DELETE FROM telemetry WHERE timestamp < ? ALLOW FILTERING";
        await _scyllaClient.ExecuteAsync(cql, new { beforeTime });

        contextLogger.Information("Old telemetry data deleted before: {BeforeTime}", beforeTime);
    }
}
