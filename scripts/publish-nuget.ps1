# =============================================================================
# Publish .NET Core NuGet Packages to Artifactory
# =============================================================================
# Usage:
#   .\publish-nuget.ps1                    # Pack and publish release
#   .\publish-nuget.ps1 -PreRelease        # Pack and publish pre-release (PR builds)
#   .\publish-nuget.ps1 -PackOnly          # Pack only (no publish)
#
# Environment Variables:
#   NUGET_API_KEY - API key for Artifactory NuGet feed
# =============================================================================

param(
    [switch]$PreRelease,
    [switch]$PackOnly,
    [string]$Version = "1.0.3"
)

$ErrorActionPreference = "Stop"

# Configuration
$NuGetSource = "https://nuget.pkg.github.com/your-github-org/index.json"
$CorePath = "$PSScriptRoot\..\core\dotnet"
$OutputPath = "$CorePath\nupkg"

# Packages to publish (in dependency order)
$Packages = @(
    "Core.Config",
    "Core.Errors", 
    "Core.Logger",
    "Core.Metrics",
    "Core.Infrastructure"
)

# Add pre-release suffix if specified
$VersionSuffix = ""
if ($PreRelease) {
    $BuildNumber = Get-Date -Format "yyyyMMddHHmm"
    $VersionSuffix = "-preview.$BuildNumber"
    Write-Host "Building pre-release packages with suffix: $VersionSuffix" -ForegroundColor Yellow
}

$FullVersion = "$Version$VersionSuffix"
Write-Host "==============================================================================" -ForegroundColor Cyan
Write-Host " Publishing .NET Core Packages v$FullVersion" -ForegroundColor Cyan
Write-Host "==============================================================================" -ForegroundColor Cyan

# Clean output directory
if (Test-Path $OutputPath) {
    Remove-Item "$OutputPath\*.nupkg" -Force
}
else {
    New-Item -ItemType Directory -Path $OutputPath | Out-Null
}

# Build and pack each package
foreach ($Package in $Packages) {
    $ProjectPath = "$CorePath\$Package\$Package.csproj"
    
    if (-not (Test-Path $ProjectPath)) {
        Write-Host "  [SKIP] $Package - project file not found" -ForegroundColor Yellow
        continue
    }
    
    Write-Host ""
    Write-Host "Building $Package..." -ForegroundColor Green
    
    # Update version in csproj
    $CsprojContent = Get-Content $ProjectPath -Raw
    $CsprojContent = $CsprojContent -replace '<Version>[^<]+</Version>', "<Version>$FullVersion</Version>"
    Set-Content -Path $ProjectPath -Value $CsprojContent
    
    # Build and pack
    dotnet pack $ProjectPath `
        --configuration Release `
        --output $OutputPath `
        --no-restore `
        /p:Version=$FullVersion
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  [FAILED] $Package pack failed" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "  [OK] $Package packed successfully" -ForegroundColor Green
}

# Restore NuGet packages first
Write-Host ""
Write-Host "Restoring packages..." -ForegroundColor Cyan
dotnet restore "$CorePath\Core.sln"

# List generated packages
Write-Host ""
Write-Host "Generated packages:" -ForegroundColor Cyan
Get-ChildItem "$OutputPath\*.nupkg" | ForEach-Object {
    Write-Host "  - $($_.Name)" -ForegroundColor White
}

# Publish to Artifactory (unless PackOnly)
if (-not $PackOnly) {
    Write-Host ""
    Write-Host "Publishing to Artifactory..." -ForegroundColor Cyan
    
    # Check for API key
    if (-not $env:NUGET_API_KEY) {
        Write-Host "[WARNING] NUGET_API_KEY not set. Set it to publish packages:" -ForegroundColor Yellow
        Write-Host "  `$env:NUGET_API_KEY = 'your-api-key'" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "To publish manually, run:" -ForegroundColor Yellow
        Get-ChildItem "$OutputPath\*.nupkg" | ForEach-Object {
            Write-Host "  dotnet nuget push `"$($_.FullName)`" --source $NuGetSource --api-key YOUR_API_KEY" -ForegroundColor White
        }
        exit 0
    }
    
    # Push each package
    Get-ChildItem "$OutputPath\*.nupkg" | ForEach-Object {
        Write-Host "  Pushing $($_.Name)..." -ForegroundColor White
        
        dotnet nuget push $_.FullName `
            --source $NuGetSource `
            --api-key $env:NUGET_API_KEY `
            --skip-duplicate
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  [FAILED] Push failed for $($_.Name)" -ForegroundColor Red
            # Continue with other packages
        }
        else {
            Write-Host "  [OK] $($_.Name) published" -ForegroundColor Green
        }
    }
}

Write-Host ""
Write-Host "==============================================================================" -ForegroundColor Cyan
Write-Host " Done! Packages available at:" -ForegroundColor Cyan
Write-Host " $NuGetSource" -ForegroundColor White
Write-Host "==============================================================================" -ForegroundColor Cyan

# Tag git for release (if not pre-release)
if (-not $PreRelease -and -not $PackOnly) {
    Write-Host ""
    Write-Host "To tag this release in git:" -ForegroundColor Yellow
    Write-Host "  git tag -a core/dotnet/v$Version -m `"Core .NET packages v$Version`"" -ForegroundColor White
    Write-Host "  git push origin core/dotnet/v$Version" -ForegroundColor White
}
