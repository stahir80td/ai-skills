using Core.Infrastructure.SqlServer;
using Core.Logger;
using AiPatterns.Domain.Interfaces;
using AiPatterns.Domain.Models;

namespace AiPatterns.Infrastructure.Repositories;

/// <summary>
/// Order repository using Core.Infrastructure.SqlServer - demonstrates transactional data patterns
/// SQL Server is optimal for ACID-compliant transactional data like orders
/// </summary>
public class OrderRepository : IOrderRepository
{
    private readonly ISqlServerClient _sqlClient;
    private readonly ServiceLogger _logger;

    public OrderRepository(ISqlServerClient sqlClient, ServiceLogger logger)
    {
        _sqlClient = sqlClient;
        _logger = logger;
    }

    public async Task<Order?> GetByIdAsync(Guid id)
    {
        var contextLogger = _logger.WithContext(component: "OrderRepository.GetById");
        contextLogger.Debug("Fetching order: {OrderId}", id);

        var sql = @"
            SELECT Id, CustomerId, OrderDate, Status, TotalAmount, 
                   ShippingAddress, CreatedAt, UpdatedAt
            FROM Orders WHERE Id = @Id";

        var orders = await _sqlClient.QueryAsync<Order>(sql, new { Id = id });
        var order = orders.FirstOrDefault();

        if (order != null)
        {
            // Note: Order.Items has private setter, cannot assign directly from raw SQL query.
            // In production, use EF Core with proper relationship mapping or modify Order model.
            // For now, returning order without items populated from separate query.
            contextLogger.Debug("Found order: {OrderId}", id);
        }

        return order;
    }

    public async Task<IEnumerable<Order>> GetAllAsync()
    {
        var contextLogger = _logger.WithContext(component: "OrderRepository.GetAll");
        contextLogger.Debug("Fetching all orders");

        var sql = @"
            SELECT Id, CustomerId, OrderDate, Status, TotalAmount, 
                   ShippingAddress, CreatedAt, UpdatedAt
            FROM Orders ORDER BY CreatedAt DESC";

        var orders = await _sqlClient.QueryAsync<Order>(sql);
        contextLogger.Information("Retrieved all orders: {Count}", orders.Count());
        return orders;
    }

    public async Task<IEnumerable<Order>> GetByCustomerIdAsync(Guid customerId)
    {
        var contextLogger = _logger.WithContext(component: "OrderRepository.GetByCustomerId");
        contextLogger.Debug("Fetching orders for customer: {CustomerId}", customerId);

        var sql = @"
            SELECT Id, CustomerId, OrderDate, Status, TotalAmount, 
                   ShippingAddress, CreatedAt, UpdatedAt
            FROM Orders WHERE CustomerId = @CustomerId ORDER BY CreatedAt DESC";

        var orders = await _sqlClient.QueryAsync<Order>(sql, new { CustomerId = customerId });
        contextLogger.Information("Retrieved orders for customer: {CustomerId}, Count: {Count}", customerId, orders.Count());
        return orders;
    }

    public async Task<IEnumerable<Order>> GetByStatusAsync(OrderStatus status)
    {
        var contextLogger = _logger.WithContext(component: "OrderRepository.GetByStatus");
        contextLogger.Debug("Fetching orders by status: {Status}", status);

        var sql = @"
            SELECT Id, CustomerId, OrderDate, Status, TotalAmount, 
                   ShippingAddress, CreatedAt, UpdatedAt
            FROM Orders WHERE Status = @Status ORDER BY CreatedAt DESC";

        var orders = await _sqlClient.QueryAsync<Order>(sql, new { Status = status.ToString() });
        contextLogger.Information("Retrieved orders by status: {Status}, Count: {Count}", status, orders.Count());
        return orders;
    }

    public async Task<Order> CreateAsync(Order order)
    {
        var contextLogger = _logger.WithContext(component: "OrderRepository.Create");
        contextLogger.Debug("Creating order: {OrderId}", order.Id);

        var sql = @"
            INSERT INTO Orders (Id, CustomerId, OrderDate, Status, TotalAmount, 
                               ShippingAddress, CreatedAt, UpdatedAt)
            VALUES (@Id, @CustomerId, @OrderDate, @Status, @TotalAmount, 
                    @ShippingAddress, @CreatedAt, @UpdatedAt)";

        await _sqlClient.ExecuteAsync(sql, new
        {
            order.Id,
            order.CustomerId,
            order.OrderDate,
            Status = order.Status.ToString(),
            order.TotalAmount,
            order.ShippingAddress,
            order.CreatedAt,
            order.UpdatedAt
        });

        foreach (var item in order.Items)
        {
            await InsertOrderItemAsync(order.Id, item);
        }

        contextLogger.Information("Order created: {OrderId}, Total: {Total}", order.Id, order.TotalAmount);
        return order;
    }

    public async Task<Order> UpdateAsync(Order order)
    {
        var contextLogger = _logger.WithContext(component: "OrderRepository.Update");
        contextLogger.Debug("Updating order: {OrderId}", order.Id);

        var sql = @"
            UPDATE Orders SET 
                Status = @Status, TotalAmount = @TotalAmount,
                ShippingAddress = @ShippingAddress, UpdatedAt = @UpdatedAt
            WHERE Id = @Id";

        await _sqlClient.ExecuteAsync(sql, new
        {
            order.Id,
            Status = order.Status.ToString(),
            order.TotalAmount,
            order.ShippingAddress,
            order.UpdatedAt
        });

        contextLogger.Information("Order updated: {OrderId}", order.Id);
        return order;
    }

    public async Task DeleteAsync(Guid id)
    {
        var contextLogger = _logger.WithContext(component: "OrderRepository.Delete");
        contextLogger.Debug("Deleting order: {OrderId}", id);

        await _sqlClient.ExecuteAsync("DELETE FROM OrderItems WHERE OrderId = @Id", new { Id = id });
        await _sqlClient.ExecuteAsync("DELETE FROM Orders WHERE Id = @Id", new { Id = id });

        contextLogger.Information("Order deleted: {OrderId}", id);
    }

    public async Task<bool> ExistsAsync(Guid id)
    {
        var order = await GetByIdAsync(id);
        return order != null;
    }

    public async Task<decimal> GetTotalRevenueAsync()
    {
        var contextLogger = _logger.WithContext(component: "OrderRepository.GetTotalRevenue");
        
        var sql = "SELECT COALESCE(SUM(TotalAmount), 0) FROM Orders WHERE Status = 'Delivered'";
        var results = await _sqlClient.QueryAsync<decimal>(sql);
        var revenue = results.FirstOrDefault();

        contextLogger.Information("Total revenue calculated: {Revenue}", revenue);
        return revenue;
    }

    public async Task<int> GetOrderCountByStatusAsync(OrderStatus status)
    {
        var contextLogger = _logger.WithContext(component: "OrderRepository.GetOrderCountByStatus");
        
        var sql = "SELECT COUNT(*) FROM Orders WHERE Status = @Status";
        var results = await _sqlClient.QueryAsync<int>(sql, new { Status = status.ToString() });
        var count = results.FirstOrDefault();

        contextLogger.Debug("Order count by status: {Status} = {Count}", status, count);
        return count;
    }

    private async Task<IEnumerable<OrderItem>> GetOrderItemsAsync(Guid orderId)
    {
        var sql = @"
            SELECT Id, ProductId, ProductName, Quantity, Price
            FROM OrderItems WHERE OrderId = @OrderId";

        return await _sqlClient.QueryAsync<OrderItem>(sql, new { OrderId = orderId });
    }

    private async Task InsertOrderItemAsync(Guid orderId, OrderItem item)
    {
        var sql = @"
            INSERT INTO OrderItems (Id, OrderId, ProductId, ProductName, Quantity, Price)
            VALUES (@Id, @OrderId, @ProductId, @ProductName, @Quantity, @Price)";

        await _sqlClient.ExecuteAsync(sql, new
        {
            Id = item.Id == Guid.Empty ? Guid.NewGuid() : item.Id,
            OrderId = orderId,
            item.ProductId,
            item.ProductName,
            item.Quantity,
            item.Price
        });
    }
}
