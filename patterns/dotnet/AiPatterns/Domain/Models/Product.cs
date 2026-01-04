namespace AiPatterns.Domain.Models;

/// <summary>
/// Sample domain entity demonstrating AI patterns
/// </summary>
public class Product
{
    public Guid Id { get; private set; }
    public string Name { get; private set; } = string.Empty;
    public string Description { get; private set; } = string.Empty;
    public decimal Price { get; private set; }
    public ProductStatus Status { get; private set; }
    public int StockQuantity { get; private set; }
    public string Category { get; private set; } = string.Empty;
    public DateTime CreatedAt { get; private set; }
    public DateTime UpdatedAt { get; private set; }

    // Private constructor for EF Core
    private Product() { }

    // Factory method for creating new products
    public static Product Create(string name, string description, decimal price, string category, int initialStock = 0)
    {
        var now = DateTime.UtcNow;
        return new Product
        {
            Id = Guid.NewGuid(),
            Name = name,
            Description = description,
            Price = price,
            Category = category,
            StockQuantity = initialStock,
            Status = ProductStatus.Draft,
            CreatedAt = now,
            UpdatedAt = now
        };
    }

    // Business logic methods
    public void UpdatePrice(decimal newPrice)
    {
        if (newPrice <= 0)
            throw new ArgumentException("Price must be greater than zero");

        Price = newPrice;
        UpdatedAt = DateTime.UtcNow;
    }

    public void UpdateStock(int newQuantity)
    {
        if (newQuantity < 0)
            throw new ArgumentException("Stock quantity cannot be negative");

        StockQuantity = newQuantity;
        UpdatedAt = DateTime.UtcNow;
    }

    public bool CanTransitionTo(ProductStatus newStatus)
    {
        return Status switch
        {
            ProductStatus.Draft => newStatus == ProductStatus.Active,
            ProductStatus.Active => newStatus == ProductStatus.Inactive || newStatus == ProductStatus.Discontinued,
            ProductStatus.Inactive => newStatus == ProductStatus.Active || newStatus == ProductStatus.Discontinued,
            ProductStatus.Discontinued => false,
            _ => false
        };
    }

    public void UpdateStatus(ProductStatus newStatus)
    {
        if (!CanTransitionTo(newStatus))
            throw new InvalidOperationException($"Cannot transition from {Status} to {newStatus}");

        Status = newStatus;
        UpdatedAt = DateTime.UtcNow;
    }

    public void Update(string name, string description, string category)
    {
        if (string.IsNullOrWhiteSpace(name))
            throw new ArgumentException("Name cannot be empty");
        if (string.IsNullOrWhiteSpace(description))
            throw new ArgumentException("Description cannot be empty");
        if (string.IsNullOrWhiteSpace(category))
            throw new ArgumentException("Category cannot be empty");

        Name = name;
        Description = description;
        Category = category;
        UpdatedAt = DateTime.UtcNow;
    }

    public bool IsInStock => StockQuantity > 0;
    public bool IsAvailable => Status == ProductStatus.Active && IsInStock;
}

public enum ProductStatus
{
    Draft,
    Active,
    Inactive,
    Discontinued
}