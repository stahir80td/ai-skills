using AiPatterns.Domain.Models;

namespace AiPatterns.Domain.Interfaces;

/// <summary>
/// Order repository interface demonstrating SQL Server via Core.Infrastructure
/// </summary>
public interface IOrderRepository
{
    Task<Order?> GetByIdAsync(Guid id);
    Task<IEnumerable<Order>> GetAllAsync();
    Task<IEnumerable<Order>> GetByCustomerIdAsync(Guid customerId);
    Task<IEnumerable<Order>> GetByStatusAsync(OrderStatus status);
    Task<Order> CreateAsync(Order order);
    Task<Order> UpdateAsync(Order order);
    Task DeleteAsync(Guid id);
    Task<bool> ExistsAsync(Guid id);
    Task<decimal> GetTotalRevenueAsync();
    Task<int> GetOrderCountByStatusAsync(OrderStatus status);
}