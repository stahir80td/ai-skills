using Core.Infrastructure;
using Core.Logger;
using AiPatterns.Domain.Interfaces;
using AiPatterns.Domain.Models;

namespace AiPatterns.Infrastructure.Cache;

/// <summary>
/// Real-time cache using Core.Infrastructure.Redis - demonstrates caching patterns
/// </summary>
public class RealtimeCache : IRealtimeCache
{
    private readonly IRedisClient _redisClient;
    private readonly ServiceLogger _logger;

    public RealtimeCache(IRedisClient redisClient, ServiceLogger logger)
    {
        _redisClient = redisClient;
        _logger = logger;
    }

    public async Task<T?> GetAsync<T>(string key) where T : class
    {
        var contextLogger = _logger.WithContext(component: "RealtimeCache.Get");

        try
        {
            var result = await _redisClient.GetAsync<T>(key);
            
            if (result != null)
            {
                contextLogger.Information("Cache hit from Redis for key {Key}", key);
            }
            else
            {
                contextLogger.Debug("Cache miss from Redis for key {Key}", key);
            }

            return result;
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error retrieving from Redis cache for key {Key}", key);
            return null;
        }
    }

    public async Task SetAsync<T>(string key, T value, TimeSpan? expiry = null) where T : class
    {
        var contextLogger = _logger.WithContext(component: "RealtimeCache.Set");

        try
        {
            await _redisClient.SetAsync(key, value, expiry);
            contextLogger.Debug("Value cached in Redis for key {Key}", key);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error caching value in Redis for key {Key}", key);
        }
    }

    public async Task RemoveAsync(string key)
    {
        var contextLogger = _logger.WithContext(component: "RealtimeCache.Remove");

        try
        {
            await _redisClient.DeleteAsync(key);
            contextLogger.Debug("Key removed from Redis cache: {Key}", key);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error removing key from Redis cache: {Key}", key);
        }
    }

    public async Task<bool> ExistsAsync(string key)
    {
        try
        {
            return await _redisClient.ExistsAsync(key);
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "Error checking key existence in Redis: {Key}", key);
            return false;
        }
    }

    public async Task<long> IncrementCounterAsync(string key, long value = 1)
    {
        var contextLogger = _logger.WithContext(component: "RealtimeCache.IncrementCounter");

        try
        {
            var result = await _redisClient.IncrementAsync(key, value);
            contextLogger.Debug("Counter incremented in Redis: {Key} = {Value}", key, result);
            return result;
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error incrementing counter in Redis: {Key}", key);
            return 0;
        }
    }

    public async Task<bool> SetAddAsync(string setKey, string value)
    {
        try
        {
            return await _redisClient.SetAddAsync(setKey, value);
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "Error adding to Redis set: {SetKey}", setKey);
            return false;
        }
    }

    public async Task<bool> SetRemoveAsync(string setKey, string value)
    {
        try
        {
            return await _redisClient.SetRemoveAsync(setKey, value);
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "Error removing from Redis set: {SetKey}", setKey);
            return false;
        }
    }

    public async Task<string[]> SetMembersAsync(string setKey)
    {
        try
        {
            return await _redisClient.SetMembersAsync(setKey);
        }
        catch (Exception ex)
        {
            _logger.Error(ex, "Error getting Redis set members: {SetKey}", setKey);
            return Array.Empty<string>();
        }
    }

    public async Task ClearPatternAsync(string pattern)
    {
        var contextLogger = _logger.WithContext(component: "RealtimeCache.ClearPattern");

        try
        {
            contextLogger.Information("Pattern clear requested for Redis cache: {Pattern}", pattern);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error clearing Redis cache pattern: {Pattern}", pattern);
        }
    }

    public async Task<LeaderboardEntry[]> GetLeaderboardAsync(string category, int top = 10)
    {
        var contextLogger = _logger.WithContext(component: "RealtimeCache.GetLeaderboard");

        try
        {
            var leaderboardKey = $"leaderboard:{category}";
            var entries = await GetAsync<LeaderboardEntry[]>(leaderboardKey);
            
            if (entries != null)
            {
                contextLogger.Information("Leaderboard retrieved from Redis cache: {Category}, {Count} entries", category, entries.Length);
                return entries.Take(top).ToArray();
            }

            contextLogger.Debug("Leaderboard not found in cache: {Category}", category);
            return Array.Empty<LeaderboardEntry>();
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error retrieving leaderboard from Redis: {Category}", category);
            return Array.Empty<LeaderboardEntry>();
        }
    }

    public async Task UpdateLeaderboardAsync(string category, string userId, double score)
    {
        var contextLogger = _logger.WithContext(component: "RealtimeCache.UpdateLeaderboard");

        try
        {
            var leaderboardKey = $"leaderboard:{category}";
            var currentEntries = await GetLeaderboardAsync(category);
            var entriesList = currentEntries.ToList();

            entriesList.RemoveAll(e => e.UserId == userId);
            entriesList.Add(LeaderboardEntry.Create(userId, userId, score));

            var updatedEntries = entriesList
                .OrderByDescending(e => e.Score)
                .Take(100)
                .Select((entry, index) => 
                {
                    entry.Rank = index + 1;
                    return entry;
                })
                .ToArray();

            await SetAsync(leaderboardKey, updatedEntries, TimeSpan.FromHours(24));

            contextLogger.Information("Leaderboard updated in Redis cache: {Category}, user {UserId}", category, userId);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error updating leaderboard in Redis: {Category}", category);
        }
    }

    public async Task SetSessionDataAsync(string sessionId, object data, TimeSpan? expiry = null)
    {
        var contextLogger = _logger.WithContext(component: "RealtimeCache.SetSession");

        try
        {
            var sessionKey = $"session:{sessionId}";
            if (data is AiPatterns.Domain.Models.SessionData sessionData)
            {
                await SetAsync(sessionKey, sessionData, expiry ?? TimeSpan.FromMinutes(30));
            }
            contextLogger.Information("Session data cached in Redis: {SessionId}", sessionId);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error caching session data in Redis: {SessionId}", sessionId);
        }
    }

    public async Task<T?> GetSessionDataAsync<T>(string sessionId) where T : class
    {
        var contextLogger = _logger.WithContext(component: "RealtimeCache.GetSession");

        try
        {
            var sessionKey = $"session:{sessionId}";
            var result = await GetAsync<T>(sessionKey);
            
            if (result != null)
            {
                contextLogger.Information("Session data retrieved from Redis cache: {SessionId}", sessionId);
            }
            else
            {
                contextLogger.Debug("Session not found in cache: {SessionId}", sessionId);
            }

            return result;
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error retrieving session data from Redis: {SessionId}", sessionId);
            return null;
        }
    }

    public async Task InvalidateSessionAsync(string sessionId)
    {
        var contextLogger = _logger.WithContext(component: "RealtimeCache.InvalidateSession");

        try
        {
            var sessionKey = $"session:{sessionId}";
            await RemoveAsync(sessionKey);
            
            contextLogger.Information("Session invalidated in Redis cache: {SessionId}", sessionId);
        }
        catch (Exception ex)
        {
            contextLogger.Error(ex, "Error invalidating session in Redis: {SessionId}", sessionId);
        }
    }
}
