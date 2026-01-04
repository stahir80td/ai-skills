namespace AiPatterns.Api.Middleware;

/// <summary>
/// Correlation ID accessor for non-HTTP layers
/// Allows infrastructure and domain layers to access correlation ID
/// </summary>
public interface ICorrelationIdAccessor
{
    string GetCorrelationId();
}

public class CorrelationIdAccessor : ICorrelationIdAccessor
{
    private readonly IHttpContextAccessor _httpContextAccessor;

    public CorrelationIdAccessor(IHttpContextAccessor httpContextAccessor)
    {
        _httpContextAccessor = httpContextAccessor;
    }

    public string GetCorrelationId()
    {
        return _httpContextAccessor.HttpContext?.Items["CorrelationId"]?.ToString() 
            ?? Guid.NewGuid().ToString();
    }
}