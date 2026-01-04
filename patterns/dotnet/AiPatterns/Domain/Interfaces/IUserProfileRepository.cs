using AiPatterns.Domain.Models;

namespace AiPatterns.Domain.Interfaces;

/// <summary>
/// User profile repository interface demonstrating MongoDB via Core.Infrastructure
/// </summary>
public interface IUserProfileRepository
{
    Task<UserProfile?> GetByIdAsync(Guid userId);
    Task<UserProfile?> GetByEmailAsync(string email);
    Task<IEnumerable<UserProfile>> GetByPreferencesAsync(string category);
    Task<UserProfile> CreateAsync(UserProfile profile);
    Task<UserProfile> UpdateAsync(UserProfile profile);
    Task DeleteAsync(Guid userId);
    Task<bool> ExistsAsync(Guid userId);
    Task<long> CountProfilesAsync();
    Task<IEnumerable<UserProfile>> SearchProfilesAsync(string query);
}