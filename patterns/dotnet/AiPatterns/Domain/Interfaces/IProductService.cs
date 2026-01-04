using AiPatterns.Domain.Models;

namespace AiPatterns.Domain.Interfaces;

/// <summary>
/// Service interface following SOD patterns
/// </summary>
public interface IProductService
{
    Task<Product> GetByIdAsync(Guid id);
    Task<IEnumerable<Product>> GetAllAsync();
    Task<IEnumerable<Product>> GetByCategoryAsync(string category);
    Task<IEnumerable<Product>> GetActiveProductsAsync();
    Task<Product> CreateAsync(string name, string description, decimal price, string category, int initialStock = 0);
    Task<Product> CreateProductAsync(string name, string description, decimal price, string category, int initialStock = 0);
    Task<Product> UpdatePriceAsync(Guid id, decimal newPrice);
    Task<Product> UpdateStockAsync(Guid id, int newQuantity);
    Task<Product> UpdateStatusAsync(Guid id, ProductStatus newStatus);
    Task DeleteProductAsync(Guid id);
    Task<bool> ReserveStockAsync(Guid id, int quantity);
    Task<bool> ReleaseStockAsync(Guid id, int quantity);
}