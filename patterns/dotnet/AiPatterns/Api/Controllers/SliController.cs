using Microsoft.AspNetCore.Mvc;
using Core.Sli;
using AiPatterns.Domain.Errors;

namespace AiPatterns.Api.Controllers;

/// <summary>
/// SLI controller demonstrating AI patterns:
/// - Exposes SLI metrics endpoint for monitoring
/// - Returns current service health metrics
/// - Required endpoint for SRE monitoring
/// </summary>
[ApiController]
[Route("api/v1/[controller]")]
[Produces("application/json")]
public class SliController : ControllerBase
{
    private readonly ISliTracker _sli;

    public SliController(ISliTracker sli)
    {
        _sli = sli;
    }

    /// <summary>
    /// Get basic SLI status for this service
    /// </summary>
    /// <returns>Basic service status</returns>
    [HttpGet]
    [ProducesResponseType(typeof(object), 200)]
    public ActionResult GetSli()
    {
        var status = new
        {
            service = "patterns-service",
            version = "1.0.0",
            timestamp = DateTime.UtcNow,
            status = "healthy"
        };
        return Ok(status);
    }

    /// <summary>
    /// Get all registered error codes for documentation
    /// </summary>
    [HttpGet("errors")]
    [ProducesResponseType(typeof(Dictionary<string, string>), 200)]
    public ActionResult<Dictionary<string, string>> GetErrorCodes()
    {
        var errorCodes = ProductErrors.GetAllErrorCodes();
        return Ok(errorCodes);
    }
}