using AiPatterns.Domain.Models;

namespace AiPatterns.Domain.Interfaces;

/// <summary>
/// Cache interface for product data
/// </summary>
public interface IProductCache
{
    // Generic cache methods
    Task<T?> GetAsync<T>(string key) where T : class;
    Task SetAsync<T>(string key, T value, TimeSpan? expiry = null) where T : class;
    Task RemoveAsync(string key);
    Task<bool> ExistsAsync(string key);
    Task ClearPatternAsync(string pattern);
    
    // Product-specific methods used by ProductService
    Task<IEnumerable<Product>?> GetProductListAsync(string key);
    Task SetProductListAsync(string key, IEnumerable<Product> products, TimeSpan? expiry = null);
    Task<Product?> GetProductAsync(Guid id);
    Task SetProductAsync(Guid id, Product product, TimeSpan? expiry = null);
    Task InvalidateProductListsAsync();
}