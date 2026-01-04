using AiPatterns.Domain.Models;

namespace AiPatterns.Domain.Interfaces;

/// <summary>
/// Repository interface following SOD patterns
/// </summary>
public interface IProductRepository
{
    Task<Product?> GetByIdAsync(Guid id);
    Task<IEnumerable<Product>> GetAllAsync();
    Task<IEnumerable<Product>> GetByCategoryAsync(string category);
    Task<IEnumerable<Product>> GetByStatusAsync(ProductStatus status);
    Task<Product> CreateAsync(Product product);
    Task<Product> UpdateAsync(Product product);
    Task DeleteAsync(Guid id);
    Task<bool> ExistsAsync(Guid id);
    Task<int> GetStockQuantityAsync(Guid id);
    
    // Additional method used by ProductService
    Task<Product> AddAsync(Product product);
}