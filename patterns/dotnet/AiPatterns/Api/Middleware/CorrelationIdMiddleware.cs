using Serilog.Context;

namespace AiPatterns.Api.Middleware;

/// <summary>
/// Correlation ID middleware demonstrating AI patterns:
/// - Captures or generates correlation IDs for distributed tracing
/// - Pushes to Serilog LogContext for automatic inclusion in all logs
/// - Returns correlation ID in response headers
/// </summary>
public class CorrelationIdMiddleware
{
    private readonly RequestDelegate _next;
    private const string CorrelationIdHeader = "X-Correlation-ID";

    public CorrelationIdMiddleware(RequestDelegate next)
    {
        _next = next;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        // Get correlation ID from header or generate new one
        var correlationId = context.Request.Headers[CorrelationIdHeader].FirstOrDefault()
            ?? Guid.NewGuid().ToString();

        // Store in HttpContext for access by other components
        context.Items["CorrelationId"] = correlationId;
        
        // Return correlation ID in response header for client tracing
        context.Response.Headers[CorrelationIdHeader] = correlationId;

        // Push to Serilog LogContext - ALL logs will automatically include this
        using (LogContext.PushProperty("CorrelationId", correlationId))
        {
            await _next(context);
        }
    }
}