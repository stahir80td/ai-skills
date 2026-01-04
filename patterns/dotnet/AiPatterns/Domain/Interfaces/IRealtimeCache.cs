using AiPatterns.Domain.Models;

namespace AiPatterns.Domain.Interfaces;

/// <summary>
/// Real-time cache interface demonstrating Redis via Core.Infrastructure
/// </summary>
public interface IRealtimeCache
{
    Task<T?> GetAsync<T>(string key) where T : class;
    Task SetAsync<T>(string key, T value, TimeSpan? expiry = null) where T : class;
    Task RemoveAsync(string key);
    Task<bool> ExistsAsync(string key);
    Task<long> IncrementCounterAsync(string key, long value = 1);
    Task<bool> SetAddAsync(string setKey, string value);
    Task<bool> SetRemoveAsync(string setKey, string value);
    Task<string[]> SetMembersAsync(string setKey);
    Task ClearPatternAsync(string pattern);
    
    // Real-time specific methods
    Task<LeaderboardEntry[]> GetLeaderboardAsync(string category, int top = 10);
    Task UpdateLeaderboardAsync(string category, string userId, double score);
    Task SetSessionDataAsync(string sessionId, object data, TimeSpan? expiry = null);
    Task<T?> GetSessionDataAsync<T>(string sessionId) where T : class;
    Task InvalidateSessionAsync(string sessionId);
}