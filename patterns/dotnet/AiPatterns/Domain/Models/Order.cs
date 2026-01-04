namespace AiPatterns.Domain.Models;

/// <summary>
/// Order entity for SQL Server storage - demonstrates transactional data
/// </summary>
public class Order
{
    public Guid Id { get; private set; }
    public Guid CustomerId { get; private set; }
    public DateTime OrderDate { get; private set; }
    public OrderStatus Status { get; private set; }
    public decimal TotalAmount { get; private set; }
    public string ShippingAddress { get; private set; } = string.Empty;
    public List<OrderItem> Items { get; private set; } = new();
    public DateTime CreatedAt { get; private set; }
    public DateTime UpdatedAt { get; private set; }

    // Private constructor for EF Core
    private Order() { }

    public static Order Create(Guid customerId, string shippingAddress, IEnumerable<OrderItem> items)
    {
        var order = new Order
        {
            Id = Guid.NewGuid(),
            CustomerId = customerId,
            OrderDate = DateTime.UtcNow,
            Status = OrderStatus.Pending,
            ShippingAddress = shippingAddress,
            Items = items.ToList(),
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow
        };

        order.TotalAmount = order.Items.Sum(i => i.Quantity * i.Price);
        return order;
    }

    public void UpdateStatus(OrderStatus newStatus)
    {
        if (!CanTransitionTo(newStatus))
            throw new InvalidOperationException($"Cannot transition from {Status} to {newStatus}");

        Status = newStatus;
        UpdatedAt = DateTime.UtcNow;
    }

    private bool CanTransitionTo(OrderStatus newStatus)
    {
        return Status switch
        {
            OrderStatus.Pending => newStatus is OrderStatus.Processing or OrderStatus.Cancelled,
            OrderStatus.Processing => newStatus is OrderStatus.Shipped or OrderStatus.Cancelled,
            OrderStatus.Shipped => newStatus is OrderStatus.Delivered,
            OrderStatus.Delivered => false,
            OrderStatus.Cancelled => false,
            _ => false
        };
    }
}

public class OrderItem
{
    public Guid Id { get; private set; }
    public Guid ProductId { get; private set; }
    public string ProductName { get; private set; } = string.Empty;
    public int Quantity { get; private set; }
    public decimal Price { get; private set; }
    public decimal Total => Quantity * Price;

    // Private constructor for EF Core
    private OrderItem() { }

    public static OrderItem Create(Guid productId, string productName, int quantity, decimal price)
    {
        return new OrderItem
        {
            Id = Guid.NewGuid(),
            ProductId = productId,
            ProductName = productName,
            Quantity = quantity,
            Price = price
        };
    }
}

public enum OrderStatus
{
    Pending,
    Processing,
    Shipped,
    Delivered,
    Cancelled
}