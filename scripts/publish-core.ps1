param(
    [switch]$SkipTests = $false
)

$ErrorActionPreference = "Stop"
$scriptDir = $PSScriptRoot
$rootDir = Split-Path -Parent $scriptDir

# =============================================================================
# VERSION MANAGEMENT
# =============================================================================
function Get-NextVersion {
    $versionFile = Join-Path $rootDir "VERSION"
    
    if (-not (Test-Path $versionFile)) {
        Write-Host "Creating VERSION file with initial version 1.0.0" -ForegroundColor Yellow
        "1.0.0" | Out-File -FilePath $versionFile -NoNewline -Encoding UTF8
        return "1.0.0"
    }
    
    $currentVersion = (Get-Content $versionFile -Raw).Trim()
    
    if ($currentVersion -match '^(\d+)\.(\d+)\.(\d+)$') {
        $major = [int]$matches[1]
        $minor = [int]$matches[2]
        $patch = [int]$matches[3]
        $newPatch = $patch + 1
        $newVersion = "$major.$minor.$newPatch"
        
        Write-Host "Version: $currentVersion -> $newVersion" -ForegroundColor Cyan
        $newVersion | Out-File -FilePath $versionFile -NoNewline -Encoding UTF8
        
        return $newVersion
    } else {
        Write-Host "WARNING: Invalid version format, resetting to 1.0.0" -ForegroundColor Yellow
        "1.0.0" | Out-File -FilePath $versionFile -NoNewline -Encoding UTF8
        return "1.0.0"
    }
}

# =============================================================================
# GO PACKAGE
# =============================================================================
function Publish-GoPackage {
    param([string]$Version)
    
    Write-Host ""
    Write-Host "========================================"  -ForegroundColor Cyan
    Write-Host "Publishing Go Core Package v$Version" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    
    $goDir = Join-Path $rootDir "core\go"
    Push-Location $goDir
    
    try {
        # Run tests
        if (-not $SkipTests) {
            Write-Host "Running Go tests..." -ForegroundColor Yellow
            go test ./... -v
            if ($LASTEXITCODE -ne 0) { throw "Go tests failed" }
            Write-Host "Go tests passed" -ForegroundColor Green
        }
        
        # Tag for Go modules
        $tag = "core/go/v$Version"
        Write-Host "Creating git tag: $tag" -ForegroundColor Yellow
        
        git add .
        git commit -m "chore(core/go): bump version to $Version" --allow-empty
        git tag -a $tag -m "Go Core Package v$Version"
        git push origin main --tags
        
        Write-Host "Go package published: $tag" -ForegroundColor Green
    }
    finally {
        Pop-Location
    }
}

# =============================================================================
# PYTHON PACKAGE
# =============================================================================
function Publish-PythonPackage {
    param([string]$Version)
    
    Write-Host ""
    Write-Host "========================================"  -ForegroundColor Cyan
    Write-Host "Publishing Python Core Package v$Version" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    
    $pyDir = Join-Path $rootDir "core\python"
    Push-Location $pyDir
    
    try {
        # Update version in setup.py
        $setupPy = Get-Content "setup.py" -Raw
        $setupPy = $setupPy -replace 'version="[\d\.]+"', "version=`"$Version`""
        $setupPy | Out-File -FilePath "setup.py" -NoNewline -Encoding UTF8
        
        # Run tests
        if (-not $SkipTests) {
            Write-Host "Running Python tests..." -ForegroundColor Yellow
            python -m pytest tests/ -v
            if ($LASTEXITCODE -ne 0) { throw "Python tests failed" }
            Write-Host "Python tests passed" -ForegroundColor Green
        }
        
        # Tag for Python (like Go modules, install from git)
        $tag = "core/python/v$Version"
        Write-Host "Creating git tag: $tag" -ForegroundColor Yellow
        
        git add .
        git commit -m "chore(core/python): bump version to $Version" --allow-empty
        git tag -a $tag -m "Python Core Package v$Version"
        git push origin main --tags
        
        Write-Host "Python package published: $tag" -ForegroundColor Green
        Write-Host "Install with: pip install git+https://github.com/your-github-org/ai-scaffolder.git@$tag#subdirectory=core/python" -ForegroundColor Gray
    }
    finally {
        Pop-Location
    }
}

# =============================================================================
# .NET PACKAGE
# =============================================================================
function Publish-DotNetPackage {
    param([string]$Version)
    
    Write-Host ""
    Write-Host "========================================"  -ForegroundColor Cyan
    Write-Host "Publishing .NET Core Package v$Version" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    
    $dotnetDir = Join-Path $rootDir "core\dotnet"
    Push-Location $dotnetDir
    
    try {
        # Run tests
        if (-not $SkipTests) {
            Write-Host "Running .NET tests..." -ForegroundColor Yellow
            dotnet test Core.Tests/Core.Tests.csproj -v normal
            if ($LASTEXITCODE -ne 0) { throw ".NET tests failed" }
            Write-Host ".NET tests passed" -ForegroundColor Green
        }
        
        # Tag for .NET (reference via git submodule or PackageReference with git source)
        $tag = "core/dotnet/v$Version"
        Write-Host "Creating git tag: $tag" -ForegroundColor Yellow
        
        git add .
        git commit -m "chore(core/dotnet): bump version to $Version" --allow-empty
        git tag -a $tag -m ".NET Core Package v$Version"
        git push origin main --tags
        
        Write-Host ".NET package published: $tag" -ForegroundColor Green
        Write-Host "Reference via git tag or add as submodule" -ForegroundColor Gray
    }
    finally {
        Pop-Location
    }
}

# =============================================================================
# MAIN
# =============================================================================
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "     AI Core Package Publisher        " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$version = Get-NextVersion

Write-Host ""
Write-Host "Publishing version: $version" -ForegroundColor Green
Write-Host ""
Write-Host "Packages:"
Write-Host "  - Go:     github.com/your-github-org/ai-scaffolder/core/go v$version"
Write-Host "  - Python: ai-core==$version"
Write-Host "  - .NET:   AI.Core.* v$version"
Write-Host ""

$confirm = Read-Host "Continue? (y/n)"
if ($confirm -ne "y") {
    Write-Host "Aborted." -ForegroundColor Yellow
    exit 0
}

try {
    Publish-GoPackage -Version $version
    Publish-PythonPackage -Version $version
    Publish-DotNetPackage -Version $version
    
    # Commit version file
    git add VERSION
    git commit -m "chore: bump core packages to v$version" --allow-empty
    git push origin main
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  All packages published: v$version    " -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
}
catch {
    Write-Host ""
    Write-Host "ERROR: $_" -ForegroundColor Red
    exit 1
}
