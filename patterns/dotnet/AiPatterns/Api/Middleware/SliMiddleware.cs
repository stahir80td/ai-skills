using System.Diagnostics;
using Core.Sli;
using Core.Errors;

namespace AiPatterns.Api.Middleware;

/// <summary>
/// SLI middleware demonstrating AI patterns:
/// - Records availability, latency, and throughput metrics for every request
/// - Integrates with Prometheus for monitoring
/// - Tracks error rates and response times
/// </summary>
public class SliMiddleware
{
    private readonly RequestDelegate _next;
    private readonly ISliTracker _sli;

    public SliMiddleware(RequestDelegate next, ISliTracker sli)
    {
        _next = next;
        _sli = sli;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        var stopwatch = Stopwatch.StartNew();
        var success = true;
        string? errorCode = null;

        try
        {
            await _next(context);
            success = context.Response.StatusCode < 400;
        }
        catch (ServiceError ex)
        {
            success = false;
            errorCode = ex.Code;
            throw; // Re-throw to let error handler middleware handle it
        }
        catch (Exception)
        {
            success = false;
            errorCode = "PAT-SYS-001";
            throw; // Re-throw to let error handler middleware handle it
        }
        finally
        {
            // ALWAYS record metrics - even on exceptions
            stopwatch.Stop();
            
            var operation = $"{context.Request.Method} {context.Request.Path}";
            
            _sli.RecordRequest(new RequestOutcome
            {
                Operation = operation,
                Success = success,
                Latency = stopwatch.Elapsed,
                ErrorCode = errorCode
            });
        }
    }
}