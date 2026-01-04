using Core.Logger;
using AiPatterns.Domain.Interfaces;
using AiPatterns.Domain.Models;
using AiPatterns.Domain.Errors;
using AiPatterns.Domain.Sli;

namespace AiPatterns.Domain.Services;

/// <summary>
/// Product business logic with comprehensive AI patterns
/// </summary>
public class ProductService : IProductService
{
    private readonly IProductRepository _repository;
    private readonly IProductCache _cache;
    private readonly ServiceLogger _logger;
    private readonly PatternsSli _sli;

    public ProductService(
        IProductRepository repository,
        IProductCache cache,
        ServiceLogger logger,
        PatternsSli sli)
    {
        _repository = repository;
        _cache = cache;
        _logger = logger;
        _sli = sli;
    }

    public async Task<IEnumerable<Product>> GetAllAsync()
    {
        var contextLogger = _logger.WithContext(component: "ProductService.GetAll");

        // Try cache first
        var cachedProducts = await _cache.GetProductListAsync("products:all");
        if (cachedProducts != null)
        {
            contextLogger.Information("Products retrieved from cache");
            return cachedProducts;
        }

        // Fallback to database
        var products = await _repository.GetAllAsync();
        var productList = products.ToList();

        // Update cache
        await _cache.SetProductListAsync("products:all", productList);

        contextLogger.Information("Products retrieved from database and cached");
        return productList;
    }

    public async Task<Product> GetByIdAsync(Guid id)
    {
        var contextLogger = _logger.WithContext(component: "ProductService.GetById");

        // Try cache first
        var cachedProduct = await _cache.GetProductAsync(id);
        if (cachedProduct != null)
        {
            contextLogger.Information("Product retrieved from cache");
            return cachedProduct;
        }

        // Fallback to database
        var product = await _repository.GetByIdAsync(id);
        if (product != null)
        {
            await _cache.SetProductAsync(product.Id, product);
            contextLogger.Information("Product retrieved from database and cached");
            return product;
        }
        else
        {
            contextLogger.Warning("Product not found");
            throw ProductErrors.NotFound(id);
        }
    }

    public async Task<Product> CreateAsync(string name, string description, decimal price, string category, int stockQuantity)
    {
        var contextLogger = _logger.WithContext(component: "ProductService.Create");

        // Validation
        if (string.IsNullOrEmpty(name))
            throw ProductErrors.NameRequired();

        if (price <= 0)
            throw ProductErrors.PriceMustBePositive(price);

        if (string.IsNullOrEmpty(category))
            throw ProductErrors.CategoryRequired();

        if (stockQuantity < 0)
            throw ProductErrors.StockCannotBeNegative(stockQuantity);

        var product = Product.Create(name, description, price, category, stockQuantity);
        
        await _repository.AddAsync(product);
        await _cache.SetProductAsync(product.Id, product);
        await _cache.InvalidateProductListsAsync(); // Invalidate list caches

        // Track business metric
        _sli.RecordProductCreated(category, product.Status.ToString().ToLower(), product.Price);

        contextLogger.Information("Product created successfully");

        return product;
    }

    public async Task<Product> UpdateAsync(Guid id, string name, string description, decimal price, string category, int stockQuantity)
    {
        var contextLogger = _logger.WithContext(component: "ProductService.Update");

        var product = await _repository.GetByIdAsync(id);
        if (product == null)
            throw ProductErrors.NotFound(id);

        // Validation
        if (string.IsNullOrEmpty(name))
            throw ProductErrors.NameRequired();

        if (price <= 0)
            throw ProductErrors.PriceMustBePositive(price);

        if (string.IsNullOrEmpty(category))
            throw ProductErrors.CategoryRequired();

        if (stockQuantity < 0)
            throw ProductErrors.StockCannotBeNegative(stockQuantity);

        product.Update(name, description, category);
        product.UpdatePrice(price);
        product.UpdateStock(stockQuantity);
        
        await _repository.UpdateAsync(product);
        await _cache.SetProductAsync(product.Id, product);
        await _cache.InvalidateProductListsAsync();

        contextLogger.Information("Product updated successfully");

        return product;
    }

    public async Task<IEnumerable<Product>> GetByCategoryAsync(string category)
    {
        var contextLogger = _logger.WithContext(component: "ProductService.GetByCategory");
        var products = await _repository.GetByCategoryAsync(category);
        contextLogger.Information("Products retrieved by category");
        return products;
    }

    public async Task<IEnumerable<Product>> GetActiveProductsAsync()
    {
        var contextLogger = _logger.WithContext(component: "ProductService.GetActive");
        var products = await _repository.GetByStatusAsync(ProductStatus.Active);
        contextLogger.Information("Active products retrieved");
        return products;
    }

    public async Task<Product> CreateProductAsync(string name, string description, decimal price, string category, int initialStock = 0)
    {
        return await CreateAsync(name, description, price, category, initialStock);
    }

    public async Task<Product> UpdatePriceAsync(Guid id, decimal newPrice)
    {
        var contextLogger = _logger.WithContext(component: "ProductService.UpdatePrice");
        var product = await _repository.GetByIdAsync(id);
        if (product == null) throw ProductErrors.NotFound(id);
        
        product.UpdatePrice(newPrice);
        var updatedProduct = await _repository.UpdateAsync(product);
        await _cache.SetProductAsync(updatedProduct.Id, updatedProduct);
        
        contextLogger.Information("Product price updated successfully");
        return updatedProduct;
    }

    public async Task<Product> UpdateStockAsync(Guid id, int newQuantity)
    {
        var contextLogger = _logger.WithContext(component: "ProductService.UpdateStock");
        var product = await _repository.GetByIdAsync(id);
        if (product == null) throw ProductErrors.NotFound(id);
        
        product.UpdateStock(newQuantity);
        var updatedProduct = await _repository.UpdateAsync(product);
        await _cache.SetProductAsync(updatedProduct.Id, updatedProduct);
        
        contextLogger.Information("Product stock updated successfully");
        return updatedProduct;
    }

    public async Task<Product> UpdateStatusAsync(Guid id, ProductStatus newStatus)
    {
        var contextLogger = _logger.WithContext(component: "ProductService.UpdateStatus");

        var product = await _repository.GetByIdAsync(id);
        if (product == null) throw ProductErrors.NotFound(id);

        var oldStatus = product.Status;
        if (!product.CanTransitionTo(newStatus))
            throw ProductErrors.InvalidStatusTransition(oldStatus, newStatus);

        product.UpdateStatus(newStatus);
        
        var updatedProduct = await _repository.UpdateAsync(product);
        await _cache.SetProductAsync(updatedProduct.Id, updatedProduct);

        contextLogger.Information("Product status updated successfully");
        return updatedProduct;
    }

    public async Task DeleteProductAsync(Guid id)
    {
        await _repository.DeleteAsync(id);
    }

    public async Task<bool> ReserveStockAsync(Guid id, int quantity)
    {
        var contextLogger = _logger.WithContext(component: "ProductService.ReserveStock");
        var product = await _repository.GetByIdAsync(id);
        if (product == null) return false;
        
        if (product.StockQuantity >= quantity)
        {
            product.UpdateStock(product.StockQuantity - quantity);
            await _repository.UpdateAsync(product);
            await _cache.SetProductAsync(product.Id, product);
            contextLogger.Information("Stock reserved successfully");
            return true;
        }
        
        contextLogger.Warning("Insufficient stock for reservation");
        return false;
    }

    public async Task<bool> ReleaseStockAsync(Guid id, int quantity)
    {
        var contextLogger = _logger.WithContext(component: "ProductService.ReleaseStock");
        var product = await _repository.GetByIdAsync(id);
        if (product == null) return false;
        
        product.UpdateStock(product.StockQuantity + quantity);
        await _repository.UpdateAsync(product);
        await _cache.SetProductAsync(product.Id, product);
        
        contextLogger.Information("Stock released successfully");
        return true;
    }
}