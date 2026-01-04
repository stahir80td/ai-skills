using System.Net;
using Core.Errors;
using AiPatterns.Domain.Models;

namespace AiPatterns.Domain.Errors;

/// <summary>
/// Centralized error definitions for Products domain with SOD scoring and operational guidance
/// </summary>
public static class ProductErrors
{
    private static readonly ErrorRegistry Registry = new();

    static ProductErrors()
    {
        // Product entity errors (PRD = Product)
        Registry.Register("PAT-PRD-001", Severity.Medium, "Product not found", HttpStatusCode.NotFound);
        Registry.Register("PAT-PRD-002", Severity.Low, "Product name is required", HttpStatusCode.BadRequest);
        Registry.Register("PAT-PRD-003", Severity.Low, "Product price must be positive", HttpStatusCode.BadRequest);
        Registry.Register("PAT-PRD-004", Severity.Low, "Product category is required", HttpStatusCode.BadRequest);
        Registry.Register("PAT-PRD-005", Severity.Low, "Stock quantity cannot be negative", HttpStatusCode.BadRequest);
        Registry.Register("PAT-PRD-006", Severity.Medium, "Invalid status transition", HttpStatusCode.Conflict);
        Registry.Register("PAT-PRD-007", Severity.Medium, "Cannot delete active product", HttpStatusCode.Conflict);
        Registry.Register("PAT-PRD-008", Severity.Medium, "Insufficient stock", HttpStatusCode.Conflict);
        
        // Infrastructure errors  
        Registry.Register("PAT-INFRA-001", Severity.High, "Database connection failed", HttpStatusCode.ServiceUnavailable);
        Registry.Register("PAT-INFRA-002", Severity.Medium, "Cache operation failed", HttpStatusCode.ServiceUnavailable);
        Registry.Register("PAT-INFRA-003", Severity.High, "External service unavailable", HttpStatusCode.ServiceUnavailable);

        // Validation errors
        Registry.Register("PAT-VAL-001", Severity.Low, "Invalid request format", HttpStatusCode.BadRequest);
        Registry.Register("PAT-VAL-002", Severity.Low, "Missing required parameter", HttpStatusCode.BadRequest);
    }

    // Product-specific factory methods
    public static ServiceError NotFound(Guid productId) =>
        Registry.CreateError("PAT-PRD-001", productId);

    public static ServiceError NameRequired() =>
        Registry.CreateError("PAT-PRD-002");

    public static ServiceError PriceMustBePositive(decimal price) =>
        Registry.CreateError("PAT-PRD-003", price);

    public static ServiceError CategoryRequired() =>
        Registry.CreateError("PAT-PRD-004");

    public static ServiceError StockCannotBeNegative(int quantity) =>
        Registry.CreateError("PAT-PRD-005", quantity);

    public static ServiceError InvalidStatusTransition(ProductStatus from, ProductStatus to) =>
        Registry.CreateError("PAT-PRD-006", from, to);

    public static ServiceError CannotDeleteActiveProduct(Guid productId) =>
        Registry.CreateError("PAT-PRD-007", productId);

    public static ServiceError InsufficientStock(int available, int requested) =>
        Registry.CreateError("PAT-PRD-008", available, requested);

    // Infrastructure factory methods
    public static ServiceError DatabaseConnectionFailed(Exception ex) =>
        Registry.WrapError(ex, "PAT-INFRA-001");

    public static ServiceError CacheOperationFailed(string operation, Exception ex) =>
        Registry.WrapError(ex, "PAT-INFRA-002", operation);

    public static ServiceError ExternalServiceUnavailable(string serviceName, Exception ex) =>
        Registry.WrapError(ex, "PAT-INFRA-003", serviceName);

    // Validation factory methods
    public static ServiceError InvalidRequestFormat(string details) =>
        Registry.CreateError("PAT-VAL-001", details);

    public static ServiceError MissingRequiredParameter(string parameterName) =>
        Registry.CreateError("PAT-VAL-002", parameterName);

    // Method for error discovery/documentation
    public static IEnumerable<string> GetAllErrorCodes() =>
        Registry.GetAll().Keys;
}