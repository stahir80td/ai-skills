using Core.Logger;
using MongoDB.Driver;
using AiPatterns.Domain.Interfaces;
using AiPatterns.Domain.Models;
using IMongoClientCore = Core.Infrastructure.MongoDB.IMongoClient;

namespace AiPatterns.Infrastructure.Repositories;

/// <summary>
/// User profile repository using Core.Infrastructure.MongoDB - demonstrates document store patterns
/// MongoDB is optimal for flexible schemas and document-oriented data like user profiles
/// </summary>
public class UserProfileRepository : IUserProfileRepository
{
    private readonly IMongoClientCore _mongoClient;
    private readonly ServiceLogger _logger;
    private const string CollectionName = "user_profiles";

    public UserProfileRepository(IMongoClientCore mongoClient, ServiceLogger logger)
    {
        _mongoClient = mongoClient;
        _logger = logger;
    }

    public async Task<UserProfile?> GetByIdAsync(Guid userId)
    {
        var contextLogger = _logger.WithContext(component: "UserProfileRepository.GetById");
        contextLogger.Debug("Fetching user profile: {UserId}", userId);

        var filter = Builders<UserProfile>.Filter.Eq(u => u.Id, userId);
        var profile = await _mongoClient.FindOneAsync(CollectionName, filter);

        if (profile != null)
        {
            contextLogger.Debug("Found user profile: {UserId}", userId);
        }

        return profile;
    }

    public async Task<UserProfile?> GetByEmailAsync(string email)
    {
        var contextLogger = _logger.WithContext(component: "UserProfileRepository.GetByEmail");
        contextLogger.Debug("Fetching user profile by email: {Email}", email);

        var filter = Builders<UserProfile>.Filter.Eq(u => u.Email, email);
        return await _mongoClient.FindOneAsync(CollectionName, filter);
    }

    public async Task<IEnumerable<UserProfile>> GetByPreferencesAsync(string category)
    {
        var contextLogger = _logger.WithContext(component: "UserProfileRepository.GetByPreferences");
        contextLogger.Debug("Fetching user profiles by preference category: {Category}", category);

        // For MongoDB, we'd typically use a filter on preferences
        // Since Core.Infrastructure doesn't have FindManyAsync, we return empty for now
        // In production, you'd extend IMongoClient with FindManyAsync
        return new List<UserProfile>();
    }

    public async Task<UserProfile> CreateAsync(UserProfile profile)
    {
        var contextLogger = _logger.WithContext(component: "UserProfileRepository.Create");
        contextLogger.Debug("Creating user profile: {UserId}", profile.Id);

        profile.CreatedAt = DateTime.UtcNow;
        profile.UpdatedAt = DateTime.UtcNow;

        await _mongoClient.InsertOneAsync(CollectionName, profile);
        contextLogger.Information("User profile created: {UserId}, {Email}", profile.Id, profile.Email);
        return profile;
    }

    public async Task<UserProfile> UpdateAsync(UserProfile profile)
    {
        var contextLogger = _logger.WithContext(component: "UserProfileRepository.Update");
        contextLogger.Debug("Updating user profile: {UserId}", profile.Id);

        profile.UpdatedAt = DateTime.UtcNow;

        var filter = Builders<UserProfile>.Filter.Eq(u => u.Id, profile.Id);
        var update = Builders<UserProfile>.Update
            .Set(u => u.Email, profile.Email)
            .Set(u => u.FirstName, profile.FirstName)
            .Set(u => u.LastName, profile.LastName)
            .Set(u => u.Preferences, profile.Preferences)
            .Set(u => u.Addresses, profile.Addresses)
            .Set(u => u.PaymentMethods, profile.PaymentMethods)
            .Set(u => u.Tags, profile.Tags)
            .Set(u => u.Metadata, profile.Metadata)
            .Set(u => u.LastLoginAt, profile.LastLoginAt)
            .Set(u => u.UpdatedAt, profile.UpdatedAt);

        await _mongoClient.UpdateOneAsync(CollectionName, filter, update);
        contextLogger.Information("User profile updated: {UserId}", profile.Id);
        return profile;
    }

    public async Task DeleteAsync(Guid userId)
    {
        var contextLogger = _logger.WithContext(component: "UserProfileRepository.Delete");
        contextLogger.Debug("Deleting user profile: {UserId}", userId);

        var filter = Builders<UserProfile>.Filter.Eq(u => u.Id, userId);
        await _mongoClient.DeleteOneAsync(CollectionName, filter);

        contextLogger.Information("User profile deleted: {UserId}", userId);
    }

    public async Task<bool> ExistsAsync(Guid userId)
    {
        var profile = await GetByIdAsync(userId);
        return profile != null;
    }

    public async Task<long> CountProfilesAsync()
    {
        var contextLogger = _logger.WithContext(component: "UserProfileRepository.CountProfiles");
        
        // Core.Infrastructure MongoDB doesn't have a Count method
        // In production, you'd extend the interface
        contextLogger.Debug("Counting user profiles (returning simulated value)");
        return 100; // Simulated value
    }

    public async Task<IEnumerable<UserProfile>> SearchProfilesAsync(string query)
    {
        var contextLogger = _logger.WithContext(component: "UserProfileRepository.SearchProfiles");
        contextLogger.Debug("Searching user profiles: {Query}", query);

        // Core.Infrastructure MongoDB doesn't have FindManyAsync
        // In production, you'd extend the interface with text search
        return new List<UserProfile>();
    }
}
