using Core.Logger;
using Microsoft.AspNetCore.Mvc;
using AiPatterns.Domain.Interfaces;
using AiPatterns.Domain.Models;

namespace AiPatterns.Api.Controllers;

/// <summary>
/// Product management API demonstrating AI patterns
/// </summary>
[ApiController]
[Route("api/v1/[controller]")]
public class ProductsController : ControllerBase
{
    private readonly IProductService _productService;
    private readonly ServiceLogger _logger;

    public ProductsController(IProductService productService, ServiceLogger logger)
    {
        _productService = productService;
        _logger = logger;
    }

    /// <summary>
    /// Get all products with caching and pagination
    /// </summary>
    [HttpGet]
    public async Task<ActionResult<IEnumerable<ProductDto>>> GetAllAsync(
        [FromQuery] int page = 1,
        [FromQuery] int pageSize = 10)
    {
        var contextLogger = _logger.WithContext(component: "GetAllProducts");

        var products = await _productService.GetAllAsync();
        var productDtos = products.Select(p => ProductDto.FromDomain(p)).ToList();

        contextLogger.Information("All products retrieved successfully");

        return Ok(productDtos);
    }

    /// <summary>
    /// Get product by ID
    /// </summary>
    [HttpGet("{id:guid}")]
    public async Task<ActionResult<ProductDto>> GetByIdAsync(Guid id)
    {
        var contextLogger = _logger.WithContext(component: "GetProductById");
        
        var product = await _productService.GetByIdAsync(id);
        if (product == null)
        {
            return NotFound();
        }
        
        var productDto = ProductDto.FromDomain(product);
        contextLogger.Information("Product retrieved successfully");
        
        return Ok(productDto);
    }

    /// <summary>
    /// Create a new product
    /// </summary>
    [HttpPost]
    public async Task<ActionResult<ProductDto>> CreateAsync([FromBody] CreateProductRequest request)
    {
        var contextLogger = _logger.WithContext(component: "CreateProduct");
        
        var product = await _productService.CreateAsync(request.Name, request.Description, request.Price, request.Category, request.StockQuantity);
        var productDto = ProductDto.FromDomain(product);
        
        contextLogger.Information("Product created successfully");
        
        return CreatedAtAction(nameof(GetByIdAsync), new { id = product.Id }, productDto);
    }

    /// <summary>
    /// Update product status
    /// </summary>
    [HttpPatch("{id:guid}/status")]
    public async Task<ActionResult> UpdateStatusAsync(Guid id, [FromBody] UpdateStatusRequest request)
    {
        var contextLogger = _logger.WithContext(component: "UpdateProductStatus");
        
        await _productService.UpdateStatusAsync(id, request.Status);
        
        contextLogger.Information("Product status updated successfully");
        
        return NoContent();
    }
}

/// <summary>
/// DTO for product representation
/// </summary>
public class ProductDto
{
    public Guid Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public string Description { get; set; } = string.Empty;
    public decimal Price { get; set; }
    public string Status { get; set; } = string.Empty;
    public int StockQuantity { get; set; }
    public string Category { get; set; } = string.Empty;
    public DateTime CreatedAt { get; set; }
    public DateTime UpdatedAt { get; set; }
    
    public static ProductDto FromDomain(Product product)
    {
        return new ProductDto
        {
            Id = product.Id,
            Name = product.Name,
            Description = product.Description,
            Price = product.Price,
            Status = product.Status.ToString(),
            StockQuantity = product.StockQuantity,
            Category = product.Category,
            CreatedAt = product.CreatedAt,
            UpdatedAt = product.UpdatedAt
        };
    }
}

/// <summary>
/// Request DTO for creating products
/// </summary>
public class CreateProductRequest
{
    public string Name { get; set; } = string.Empty;
    public string Description { get; set; } = string.Empty;
    public decimal Price { get; set; }
    public string Category { get; set; } = string.Empty;
    public int StockQuantity { get; set; }
}

/// <summary>
/// Request DTO for updating product status
/// </summary>
public class UpdateStatusRequest
{
    public ProductStatus Status { get; set; }
}