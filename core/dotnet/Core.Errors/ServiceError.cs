using System.Reflection;
using System.Net;

namespace Core.Errors;

/// <summary>
/// Severity levels for errors
/// </summary>
public static class Severity
{
    /// <summary>System unavailable, data loss, security breach</summary>
    public const string Critical = "CRITICAL";

    /// <summary>Major functionality broken, significant impact</summary>
    public const string High = "HIGH";

    /// <summary>Moderate impact, workaround available</summary>
    public const string Medium = "MEDIUM";

    /// <summary>Minor issue, minimal impact</summary>
    public const string Low = "LOW";

    /// <summary>Informational, not an error</summary>
    public const string Info = "INFO";
}

/// <summary>
/// Structured error with code, severity, and context for SRE observability
/// </summary>
public class ServiceError : Exception
{
    /// <summary>
    /// Error code (e.g., "INGEST-001")
    /// </summary>
    public string Code { get; }

    /// <summary>
    /// Severity level (CRITICAL, HIGH, MEDIUM, LOW, INFO)
    /// </summary>
    public string ErrorSeverity { get; }

    /// <summary>
    /// Additional context (user_id, request_id, etc.)
    /// </summary>
    public Dictionary<string, object> Context { get; }

    /// <summary>
    /// Creates a new ServiceError
    /// </summary>
    public ServiceError(string code, string severity, string message)
        : base(message)
    {
        Code = code;
        ErrorSeverity = severity;
        Context = new Dictionary<string, object>();
    }

    /// <summary>
    /// Creates a new ServiceError wrapping an inner exception
    /// </summary>
    public ServiceError(string code, string severity, string message, Exception innerException)
        : base(message, innerException)
    {
        Code = code;
        ErrorSeverity = severity;
        Context = new Dictionary<string, object>();
    }

    /// <summary>
    /// Adds context to the error (fluent API)
    /// </summary>
    public ServiceError WithContext(string key, object value)
    {
        Context[key] = value;
        return this;
    }

    /// <summary>
    /// Gets a context value by key
    /// </summary>
    public T? GetContext<T>(string key)
    {
        if (Context.TryGetValue(key, out var value) && value is T typedValue)
        {
            return typedValue;
        }
        return default;
    }

    public override string ToString()
    {
        if (InnerException != null)
        {
            return $"[{Code}] {ErrorSeverity}: {Message} (caused by: {InnerException.Message})";
        }
        return $"[{Code}] {ErrorSeverity}: {Message}";
    }
}

/// <summary>
/// Error definition with SOD scores, HTTP mapping, and operational guidance
/// </summary>
public class ErrorDefinition
{
    /// <summary>Error code (e.g., "INGEST-001")</summary>
    public required string Code { get; init; }

    /// <summary>Severity level</summary>
    public required string Severity { get; init; }

    /// <summary>Detailed description with format placeholders</summary>
    public required string Description { get; init; }

    /// <summary>HTTP status code for this error</summary>
    public System.Net.HttpStatusCode HttpStatusCode { get; init; } = System.Net.HttpStatusCode.InternalServerError;

    /// <summary>Severity × Occurrence × Detectability (1-1000)</summary>
    public int SODScore => SeverityScore * OccurrenceScore * DetectabilityScore;

    /// <summary>Severity score (1-10)</summary>
    public int SeverityScore { get; init; } = 5;

    /// <summary>Occurrence score (1-10)</summary>
    public int OccurrenceScore { get; init; } = 5;

    /// <summary>Detectability score (1-10)</summary>
    public int DetectabilityScore { get; init; } = 5;

    /// <summary>How to resolve this error - operational guidance</summary>
    public string Mitigation { get; init; } = string.Empty;

    /// <summary>Example scenario when this error occurs</summary>
    public string Example { get; init; } = string.Empty;

    /// <summary>Related error codes that might occur together</summary>
    public string[] RelatedErrors { get; init; } = Array.Empty<string>();

    /// <summary>Tags for categorization (database, network, validation, etc.)</summary>
    public string[] Tags { get; init; } = Array.Empty<string>();
}

/// <summary>
/// Registry for error definitions with SOD scoring, HTTP mapping, and operational guidance
/// </summary>
public class ErrorRegistry
{
    private readonly Dictionary<string, ErrorDefinition> _definitions = new();

    /// <summary>
    /// Registers an error definition
    /// </summary>
    public void Register(ErrorDefinition definition)
    {
        _definitions[definition.Code] = definition;
    }

    /// <summary>
    /// Convenient registration method with all parameters
    /// </summary>
    public void Register(string code, string severity, string description, 
        System.Net.HttpStatusCode httpStatusCode = System.Net.HttpStatusCode.InternalServerError,
        int severityScore = 5, int occurrenceScore = 5, int detectabilityScore = 5,
        string mitigation = "", string example = "", string[]? tags = null, string[]? relatedErrors = null)
    {
        Register(new ErrorDefinition
        {
            Code = code,
            Severity = severity,
            Description = description,
            HttpStatusCode = httpStatusCode,
            SeverityScore = severityScore,
            OccurrenceScore = occurrenceScore,
            DetectabilityScore = detectabilityScore,
            Mitigation = mitigation,
            Example = example,
            Tags = tags ?? Array.Empty<string>(),
            RelatedErrors = relatedErrors ?? Array.Empty<string>()
        });
    }

    /// <summary>
    /// Registers errors from assembly using attributes (for automatic discovery)
    /// </summary>
    public void RegisterFromAssembly(Assembly assembly)
    {
        var errorTypes = assembly.GetTypes()
            .Where(t => t.GetCustomAttribute<ErrorDefinitionAttribute>() != null);

        foreach (var type in errorTypes)
        {
            var methods = type.GetMethods(BindingFlags.Static | BindingFlags.Public)
                .Where(m => m.GetCustomAttribute<ErrorDefinitionAttribute>() != null);

            foreach (var method in methods)
            {
                var attr = method.GetCustomAttribute<ErrorDefinitionAttribute>();
                if (attr != null)
                {
                    Register(attr.Code, attr.Severity, attr.Description, 
                        attr.HttpStatusCode, attr.SeverityScore, attr.OccurrenceScore, 
                        attr.DetectabilityScore, attr.Mitigation, attr.Example, 
                        attr.Tags, attr.RelatedErrors);
                }
            }
        }
    }

    /// <summary>
    /// Gets an error definition by code
    /// </summary>
    public ErrorDefinition? Get(string code)
    {
        return _definitions.TryGetValue(code, out var def) ? def : null;
    }

    /// <summary>
    /// Gets all registered error definitions
    /// </summary>
    public IReadOnlyDictionary<string, ErrorDefinition> GetAll()
    {
        return _definitions;
    }

    /// <summary>
    /// Gets errors by tag (e.g., "database", "network")
    /// </summary>
    public IEnumerable<ErrorDefinition> GetByTag(string tag)
    {
        return _definitions.Values.Where(d => d.Tags.Contains(tag));
    }

    /// <summary>
    /// Gets errors by severity level
    /// </summary>
    public IEnumerable<ErrorDefinition> GetBySeverity(string severity)
    {
        return _definitions.Values.Where(d => d.Severity == severity);
    }

    /// <summary>
    /// Creates a ServiceError from a registered error definition with context
    /// </summary>
    public ServiceError CreateError(string code, params object[] messageArgs)
    {
        var def = Get(code);
        if (def == null)
        {
            return new ServiceError(code, Severity.Medium, $"Unknown error: {code}");
        }

        var message = messageArgs.Length > 0
            ? string.Format(def.Description, messageArgs)
            : def.Description;

        return new ServiceError(code, def.Severity, message)
            .WithContext("http_status_code", (int)def.HttpStatusCode)
            .WithContext("sod_score", def.SODScore)
            .WithContext("mitigation", def.Mitigation)
            .WithContext("example", def.Example)
            .WithContext("tags", def.Tags);
    }

    /// <summary>
    /// Creates a ServiceError with additional context
    /// </summary>
    public ServiceError CreateError(string code, Dictionary<string, object> context, params object[] messageArgs)
    {
        var serviceError = CreateError(code, messageArgs);
        
        foreach (var kvp in context)
        {
            serviceError.WithContext(kvp.Key, kvp.Value);
        }
        
        return serviceError;
    }

    /// <summary>
    /// Wraps an existing exception using a registered error definition
    /// </summary>
    public ServiceError WrapError(Exception ex, string code, params object[] messageArgs)
    {
        var def = Get(code);
        if (def == null)
        {
            return new ServiceError(code, Severity.Medium, $"Unknown error: {code}", ex);
        }

        var message = messageArgs.Length > 0
            ? string.Format(def.Description, messageArgs)
            : def.Description;

        return new ServiceError(code, def.Severity, message, ex)
            .WithContext("http_status_code", (int)def.HttpStatusCode)
            .WithContext("sod_score", def.SODScore)
            .WithContext("mitigation", def.Mitigation)
            .WithContext("original_exception", ex.GetType().Name);
    }

    /// <summary>
    /// Gets error documentation for OpenAPI/Swagger generation
    /// </summary>
    public Dictionary<string, object> GetErrorDocumentation()
    {
        return _definitions.Values.ToDictionary(
            def => def.Code,
            def => (object)new
            {
                code = def.Code,
                severity = def.Severity,
                description = def.Description,
                httpStatusCode = (int)def.HttpStatusCode,
                sodScore = def.SODScore,
                mitigation = def.Mitigation,
                example = def.Example,
                tags = def.Tags
            }
        );
    }
}

/// <summary>
/// Utility methods for SOD calculation
/// </summary>
public static class SodCalculator
{
    /// <summary>
    /// Calculates the SOD score (Severity × Occurrence × Detectability)
    /// </summary>
    public static int Calculate(int severity, int occurrence, int detectability)
    {
        return severity * occurrence * detectability;
    }
}

/// <summary>
/// Attribute for marking error definition methods for automatic registration
/// </summary>
[AttributeUsage(AttributeTargets.Method, AllowMultiple = false)]
public class ErrorDefinitionAttribute : Attribute
{
    public string Code { get; set; }
    public string Severity { get; set; }
    public string Description { get; set; }
    public HttpStatusCode HttpStatusCode { get; set; } = HttpStatusCode.InternalServerError;
    public int SeverityScore { get; set; } = 5;
    public int OccurrenceScore { get; set; } = 5;
    public int DetectabilityScore { get; set; } = 5;
    public string Mitigation { get; set; } = "";
    public string Example { get; set; } = "";
    public string[] Tags { get; set; } = Array.Empty<string>();
    public string[] RelatedErrors { get; set; } = Array.Empty<string>();

    public ErrorDefinitionAttribute(string code, string severity, string description)
    {
        Code = code;
        Severity = severity;
        Description = description;
    }
}

/// <summary>
/// Extension methods for ServiceError
/// </summary>
public static class ServiceErrorExtensions
{
    /// <summary>
    /// Gets the HTTP status code from error context, fallback to 500
    /// </summary>
    public static HttpStatusCode GetHttpStatusCode(this ServiceError error)
    {
        if (error.GetContext<int>("http_status_code") is int statusCode)
        {
            return (HttpStatusCode)statusCode;
        }
        return HttpStatusCode.InternalServerError;
    }

    /// <summary>
    /// Gets the SOD score from error context
    /// </summary>
    public static int GetSODScore(this ServiceError error)
    {
        return error.GetContext<int>("sod_score");
    }

    /// <summary>
    /// Gets mitigation guidance from error context
    /// </summary>
    public static string GetMitigation(this ServiceError error)
    {
        return error.GetContext<string>("mitigation") ?? "";
    }

    /// <summary>
    /// Checks if error has a specific tag
    /// </summary>
    public static bool HasTag(this ServiceError error, string tag)
    {
        var tags = error.GetContext<string[]>("tags");
        return tags?.Contains(tag) == true;
    }
}
