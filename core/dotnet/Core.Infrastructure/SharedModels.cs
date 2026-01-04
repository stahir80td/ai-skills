namespace Core.Infrastructure
{
    /// <summary>
    /// Diagnostic information about a database connection
    /// </summary>
    public class ConnectionDiagnostics
    {
        public bool IsHealthy { get; set; }
        public string DatabaseName { get; set; } = "";
        public string ServerVersion { get; set; } = "";
        public TimeSpan ConnectionTime { get; set; }
        public string? ErrorMessage { get; set; }
    }
}