namespace AiPatterns.Domain.Interfaces;

/// <summary>
/// External pricing service interface
/// </summary>
public interface IPricingService
{
    Task<decimal> GetDiscountedPriceAsync(Guid productId, decimal originalPrice);
    Task<bool> ValidatePriceAsync(Guid productId, decimal price);
}