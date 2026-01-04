using Core.Errors;
using Core.Logger;

namespace AiPatterns.Api.Middleware;

/// <summary>
/// Error handler middleware demonstrating AI patterns:
/// - Catches all exceptions and converts to standardized error responses
/// - Logs all errors with full context
/// - Returns proper HTTP status codes and error formats
/// </summary>
public class ErrorHandlerMiddleware
{
    private readonly RequestDelegate _next;
    private readonly ServiceLogger _logger;

    public ErrorHandlerMiddleware(RequestDelegate next, ServiceLogger logger)
    {
        _next = next;
        _logger = logger;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        try
        {
            await _next(context);
        }
        catch (ServiceError ex)
        {
            _logger.Warning("Service error occurred", ex, new { 
                code = ex.Code,
                path = context.Request.Path,
                method = context.Request.Method
            });
            
            await WriteErrorResponse(context, (int)ex.GetHttpStatusCode(), ex.Code, ex.Message, ex.Context);
        }
        catch (Exception ex)
        {
            _logger.Error("Unhandled exception occurred", ex, new { 
                path = context.Request.Path,
                method = context.Request.Method
            });
            
            await WriteErrorResponse(context, 500, "PAT-SYS-001", "Internal server error", null);
        }
    }

    private static async Task WriteErrorResponse(
        HttpContext context, 
        int statusCode, 
        string code, 
        string message, 
        object? details)
    {
        // Don't overwrite response if already started
        if (context.Response.HasStarted)
            return;

        context.Response.StatusCode = statusCode;
        context.Response.ContentType = "application/json";

        var error = new
        {
            error = new
            {
                code,
                message,
                details,
                timestamp = DateTime.UtcNow,
                traceId = context.TraceIdentifier,
                correlationId = context.Items["CorrelationId"]?.ToString()
            }
        };

        await context.Response.WriteAsJsonAsync(error);
    }
}