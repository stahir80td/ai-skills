---
name: ai-scaffold-service-dotnet
description: >
  Creates a complete .NET microservice from scratch following AI patterns.
  Use when asked to create, scaffold, or generate a new .NET service, API, or microservice.
  Generates complete project structure with Domain, Infrastructure, and API layers.
  Includes all Core package integrations, middleware, Helm charts, and Docker setup.
---

# AI .NET Service Scaffolding

## Pre-Scaffolding Questions

Before generating, ask the user:

```
Before I scaffold your .NET service, I need to confirm:

1. **Service Name**: What should we call this service?
   (e.g., "order-service", "inventory-manager")

2. **Project Location**: Where should I create the project?
   (e.g., "C:\dev\order-service" - must be NEW folder)

3. **Data Stores**: Which do you need?
   - [ ] SQL Server (transactional)
   - [ ] MongoDB (documents)
   - [ ] ScyllaDB (time-series)
   - [ ] Redis (caching)
   - [ ] Kafka (events)

4. **GitHub Token**: Required for Core packages.
   Get one at: https://github.com/settings/tokens/new (read:packages scope)
```

---

## Project Structure (SOD Pattern)

```
{service-name}/
├── NuGet.config                          # GitHub Packages source
├── {ServiceName}.sln
├── Dockerfile
├── Taskfile.yml
├── src/
│   ├── {ServiceName}.Api/
│   │   ├── {ServiceName}.Api.csproj
│   │   ├── Program.cs                    # DI registration
│   │   ├── Controllers/
│   │   │   ├── {Entity}Controller.cs
│   │   │   └── SliController.cs          # REQUIRED
│   │   └── Middleware/
│   │       ├── CorrelationIdMiddleware.cs
│   │       ├── ErrorHandlerMiddleware.cs
│   │       └── SliMiddleware.cs
│   ├── {ServiceName}.Domain/
│   │   ├── {ServiceName}.Domain.csproj
│   │   ├── Models/{Entity}.cs
│   │   ├── Services/{Entity}Service.cs
│   │   ├── Interfaces/
│   │   └── Errors/{ServiceName}Errors.cs
│   └── {ServiceName}.Infrastructure/
│       ├── {ServiceName}.Infrastructure.csproj
│       ├── Repositories/{Entity}Repository.cs
│       ├── Cache/{Entity}Cache.cs
│       └── Kafka/{Entity}EventPublisher.cs
├── {service-name}-ui/                    # React frontend (optional)
└── helm/{service-name}/                  # Kubernetes deployment
```

---

## Required Files

### NuGet.config

```xml
<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" />
    <add key="github" value="https://nuget.pkg.github.com/your-github-org/index.json" />
  </packageSources>
  <packageSourceCredentials>
    <github>
      <add key="Username" value="ai-user" />
      <add key="ClearTextPassword" value="%GITHUB_TOKEN%" />
    </github>
  </packageSourceCredentials>
</configuration>
```

### .csproj Template

```xml
<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
  </PropertyGroup>
  
  <ItemGroup>
    <PackageReference Include="Core.Config" Version="1.0.4" />
    <PackageReference Include="Core.Errors" Version="1.0.4" />
    <PackageReference Include="Core.Logger" Version="1.0.4" />
    <PackageReference Include="Core.Metrics" Version="1.0.4" />
    <PackageReference Include="Core.Sli" Version="1.0.4" />
    <PackageReference Include="Core.Infrastructure" Version="1.0.4" />
    <PackageReference Include="Core.Reliability" Version="1.0.4" />
  </ItemGroup>
</Project>
```

### Dockerfile

```dockerfile
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
ARG GITHUB_TOKEN
ENV GITHUB_TOKEN=$GITHUB_TOKEN
WORKDIR /src
COPY NuGet.config .
COPY *.sln .
COPY src/ src/
RUN dotnet restore
RUN dotnet publish -c Release -o /app

FROM mcr.microsoft.com/dotnet/aspnet:8.0
WORKDIR /app
COPY --from=build /app .
EXPOSE 80
ENTRYPOINT ["dotnet", "{ServiceName}.Api.dll"]
```

---

## Program.cs Template

```csharp
using Core.Logger;
using Core.Infrastructure;
using Core.Infrastructure.Kafka;
using Core.Sli;
using {ServiceName}.Api.Middleware;
using {ServiceName}.Domain.Services;
using {ServiceName}.Infrastructure.Repositories;

var builder = WebApplication.CreateBuilder(args);

// ========================================
// Core.Logger
// ========================================
builder.Services.AddSingleton<ServiceLogger>(sp => 
    new ServiceLogger("{service-name}", builder.Configuration));

// ========================================
// Core.Infrastructure Clients
// ========================================
builder.Services.AddSingleton<ISqlServerClient>(sp =>
    new SqlServerClient(new SqlServerConfig 
    { 
        ConnectionString = builder.Configuration.GetConnectionString("SqlServer")! 
    }));

builder.Services.AddSingleton<IRedisClient>(sp =>
    new RedisClient(new RedisConfig 
    { 
        ConnectionString = builder.Configuration.GetConnectionString("Redis")! 
    }));

builder.Services.AddSingleton<IKafkaProducer>(sp =>
    new KafkaProducer(new KafkaConfig 
    { 
        BootstrapServers = builder.Configuration["Kafka:BootstrapServers"]! 
    }));

// ========================================
// Core.Sli
// ========================================
builder.Services.AddSingleton<SliTracker>();

// ========================================
// Domain Services
// ========================================
builder.Services.AddScoped<{Entity}Service>();

// ========================================
// Infrastructure
// ========================================
builder.Services.AddScoped<{Entity}Repository>();

// ========================================
// CORS (for UI)
// ========================================
builder.Services.AddCors(options =>
{
    options.AddPolicy("AllowUI", policy =>
    {
        policy.WithOrigins("http://localhost:3000", "http://localhost:3001")
            .AllowAnyHeader()
            .AllowAnyMethod()
            .AllowCredentials();
    });
});

builder.Services.AddControllers();
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

var app = builder.Build();

// ========================================
// Middleware Pipeline (ORDER MATTERS!)
// ========================================
app.UseCors("AllowUI");
app.UseMiddleware<CorrelationIdMiddleware>();  // FIRST
app.UseMiddleware<ErrorHandlerMiddleware>();   // SECOND
app.UseMiddleware<SliMiddleware>();            // THIRD

app.UseSwagger();
app.UseSwaggerUI();
app.MapControllers();

// Health endpoints
app.MapGet("/health/live", () => Results.Ok(new { status = "healthy" }));
app.MapGet("/health/ready", () => Results.Ok(new { status = "ready" }));

app.Run();
```

---

## Required Middleware

### CorrelationIdMiddleware.cs

```csharp
public class CorrelationIdMiddleware
{
    private readonly RequestDelegate _next;
    
    public CorrelationIdMiddleware(RequestDelegate next) => _next = next;
    
    public async Task InvokeAsync(HttpContext context)
    {
        var correlationId = context.Request.Headers["X-Correlation-ID"].FirstOrDefault()
            ?? Guid.NewGuid().ToString();
        
        context.Items["CorrelationId"] = correlationId;
        context.Response.Headers["X-Correlation-ID"] = correlationId;
        
        await _next(context);
    }
}
```

### ErrorHandlerMiddleware.cs

```csharp
using Core.Errors;

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
        catch (ServiceError error)
        {
            _logger.Error("Service error: {ErrorCode} - {Message}", error.Code, error.Message);
            
            context.Response.StatusCode = (int)error.StatusCode;
            context.Response.ContentType = "application/json";
            
            await context.Response.WriteAsJsonAsync(new
            {
                error = error.Code,
                message = error.Message,
                correlationId = context.Items["CorrelationId"]?.ToString()
            });
        }
    }
}
```

### SliMiddleware.cs

```csharp
using Core.Sli;

public class SliMiddleware
{
    private readonly RequestDelegate _next;
    private readonly SliTracker _sliTracker;
    
    public SliMiddleware(RequestDelegate next, SliTracker sliTracker)
    {
        _next = next;
        _sliTracker = sliTracker;
    }
    
    public async Task InvokeAsync(HttpContext context)
    {
        var stopwatch = Stopwatch.StartNew();
        var endpoint = $"{context.Request.Method} {context.Request.Path}";
        
        try
        {
            await _next(context);
            stopwatch.Stop();
            
            _sliTracker.RecordRequest(endpoint, stopwatch.ElapsedMilliseconds, 
                context.Response.StatusCode < 500);
        }
        catch
        {
            stopwatch.Stop();
            _sliTracker.RecordRequest(endpoint, stopwatch.ElapsedMilliseconds, false);
            throw;
        }
    }
}
```

---

## Required Endpoints

Every service MUST expose:

| Endpoint | Purpose |
|----------|---------|
| `/health/live` | Kubernetes liveness probe |
| `/health/ready` | Kubernetes readiness probe |
| `/metrics` | Prometheus scraping |
| `/api/v1/sli` | SLI metrics dashboard |
| `/swagger` | API documentation |

---

## Taskfile.yml

```yaml
version: '3'

vars:
  SERVICE_NAME: {service-name}

tasks:
  deploy-local:
    desc: Build and deploy to local Kubernetes
    cmds:
      - docker build -t {{.SERVICE_NAME}}:local --build-arg GITHUB_TOKEN=$env:GITHUB_TOKEN .
      - kubectl create namespace sandbox --dry-run=client -o yaml | kubectl apply -f -
      - helm upgrade --install {{.SERVICE_NAME}} ./helm/{{.SERVICE_NAME}} -n sandbox
      - kubectl wait --for=condition=ready pod -l app={{.SERVICE_NAME}} -n sandbox --timeout=180s
      
  run-local:
    desc: Port forward API and UI
    cmds:
      - |
        Start-Job {kubectl port-forward svc/{{.SERVICE_NAME}}-api 8080:80 -n sandbox}
        Start-Job {kubectl port-forward svc/{{.SERVICE_NAME}}-ui 3000:80 -n sandbox}
        Write-Host "API: http://localhost:8080"
        Write-Host "UI: http://localhost:3000"
```

---

## Reference Implementation

Always read these pattern files before generating:
- `patterns/dotnet/AiPatterns/Program.cs`
- `patterns/dotnet/AiPatterns/Domain/Services/PatternsService.cs`
- `patterns/dotnet/AiPatterns/Infrastructure/Messaging/EventPublisher.cs`
- `patterns/dotnet/AiPatterns/Infrastructure/Repositories/OrderRepository.cs`
