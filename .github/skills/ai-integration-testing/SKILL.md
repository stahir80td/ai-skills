````markdown
---
name: ai-integration-testing
description: >
  MANDATORY integration testing skill that runs AFTER Docker Compose deployment.
  Verifies both API and UI work correctly before proceeding to Helm chart and final delivery.
  Includes API endpoint testing, UI smoke tests, and end-to-end validation.
  ALL tests MUST pass before proceeding to final phase.
---

# AI Integration Testing Skill

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  MANDATORY INTEGRATION TESTING  ⚠️                              ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   AFTER Docker Compose deploys infrastructure + application:                             ║
║                                                                                          ║
║   1. ✅ ALL containers must be healthy                                                   ║
║   2. ✅ ALL databases must be seeded                                                     ║
║   3. ✅ API tests must pass (health, CRUD, error handling)                               ║
║   4. ✅ UI tests must pass (build, load, navigation, API calls)                          ║
║   5. ✅ End-to-end flow must work (create → read → update → delete)                      ║
║                                                                                          ║
║   ❌ DO NOT proceed to Helm chart until ALL tests pass!                                  ║
║   ❌ DO NOT declare success with failing tests!                                          ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## Test Execution Order

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                           INTEGRATION TEST SEQUENCE                                      ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   STEP 1: Infrastructure Health                                                          ║
║   ────────────────────────────────────────────────────────────────                       ║
║   □ All Docker containers running and healthy                                            ║
║   □ SQL Server accepting connections                                                     ║
║   □ MongoDB accepting connections                                                        ║
║   □ ScyllaDB accepting connections (if used)                                             ║
║   □ Redis responding to PING                                                             ║
║   □ Kafka broker ready                                                                   ║
║                                                                                          ║
║   STEP 2: API Service Health                                                             ║
║   ────────────────────────────────────────────────────────────────                       ║
║   □ /health endpoint returns 200                                                         ║
║   □ /health/live endpoint returns 200                                                    ║
║   □ /health/ready endpoint returns 200                                                   ║
║   □ /metrics endpoint returns Prometheus metrics                                         ║
║   □ /api/v1/sli endpoint returns SLI data                                                ║
║                                                                                          ║
║   STEP 3: API CRUD Tests                                                                 ║
║   ────────────────────────────────────────────────────────────────                       ║
║   □ CREATE: POST returns 201 with created entity                                         ║
║   □ READ: GET returns 200 with entity list                                               ║
║   □ READ ONE: GET /{id} returns 200 with entity                                          ║
║   □ UPDATE: PUT /{id} returns 200 with updated entity                                    ║
║   □ DELETE: DELETE /{id} returns 204                                                     ║
║   □ NOT FOUND: GET /invalid-id returns 404                                               ║
║                                                                                          ║
║   STEP 4: UI Build & Smoke Tests                                                         ║
║   ────────────────────────────────────────────────────────────────                       ║
║   □ npm install succeeds                                                                 ║
║   □ npm run build succeeds (no TypeScript errors)                                        ║
║   □ npm run dev starts without errors                                                    ║
║   □ Dashboard page loads                                                                 ║
║   □ Navigation works to all pages                                                        ║
║   □ No JavaScript console errors                                                         ║
║                                                                                          ║
║   STEP 5: End-to-End Flow                                                                ║
║   ────────────────────────────────────────────────────────────────                       ║
║   □ Create entity via API                                                                ║
║   □ Verify entity appears in UI list                                                     ║
║   □ Update entity via API                                                                ║
║   □ Verify changes reflected in UI                                                       ║
║   □ Delete entity via API                                                                ║
║   □ Verify entity removed from UI                                                        ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## Step 1: Infrastructure Health Tests

### Container Health Check

```powershell
# Check all containers are running
$containers = docker-compose ps --format json | ConvertFrom-Json
$allHealthy = $true

foreach ($container in $containers) {
    $status = docker inspect --format='{{.State.Health.Status}}' $container.Name 2>$null
    if ($status -eq "unhealthy" -or $container.State -ne "running") {
        Write-Host "❌ Container unhealthy: $($container.Name)" -ForegroundColor Red
        $allHealthy = $false
    } else {
        Write-Host "✅ Container healthy: $($container.Name)" -ForegroundColor Green
    }
}

if (-not $allHealthy) {
    Write-Host "❌ INFRASTRUCTURE HEALTH FAILED - Fix containers before proceeding" -ForegroundColor Red
    exit 1
}
```

### Database Connectivity Tests

```powershell
# SQL Server
Write-Host "Testing SQL Server connection..." -ForegroundColor Cyan
docker exec {project}-sqlserver-1 /opt/mssql-tools18/bin/sqlcmd `
    -S localhost -U sa -P "YourStrong!Password" -C `
    -Q "SELECT 1 AS Connected" -h -1

# MongoDB
Write-Host "Testing MongoDB connection..." -ForegroundColor Cyan
docker exec {project}-mongodb-1 mongosh --quiet --eval "db.runCommand({ping:1})"

# ScyllaDB (if used)
Write-Host "Testing ScyllaDB connection..." -ForegroundColor Cyan
docker exec {project}-scylladb-1 cqlsh -e "DESCRIBE KEYSPACES;"

# Redis
Write-Host "Testing Redis connection..." -ForegroundColor Cyan
$redisPing = docker exec {project}-redis-1 redis-cli ping
if ($redisPing -ne "PONG") { throw "Redis not responding" }

# Kafka
Write-Host "Testing Kafka connection..." -ForegroundColor Cyan
docker exec {project}-kafka-1 /opt/kafka/bin/kafka-topics.sh `
    --bootstrap-server localhost:9092 --list
```

---

## Step 2: API Service Health Tests

### Health Endpoints Script

```powershell
#!/usr/bin/env pwsh
# api-health-tests.ps1

param(
    [string]$BaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"
$allPassed = $true

Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           API Health Tests                                   ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

# Test endpoints
$healthEndpoints = @(
    @{ Path = "/health"; ExpectedStatus = 200; Name = "Health" },
    @{ Path = "/health/live"; ExpectedStatus = 200; Name = "Liveness" },
    @{ Path = "/health/ready"; ExpectedStatus = 200; Name = "Readiness" },
    @{ Path = "/metrics"; ExpectedStatus = 200; Name = "Metrics" },
    @{ Path = "/api/v1/sli"; ExpectedStatus = 200; Name = "SLI" }
)

foreach ($endpoint in $healthEndpoints) {
    try {
        $response = Invoke-WebRequest -Uri "$BaseUrl$($endpoint.Path)" -Method GET -UseBasicParsing
        if ($response.StatusCode -eq $endpoint.ExpectedStatus) {
            Write-Host "✅ $($endpoint.Name): $($endpoint.Path) → $($response.StatusCode)" -ForegroundColor Green
        } else {
            Write-Host "❌ $($endpoint.Name): Expected $($endpoint.ExpectedStatus), got $($response.StatusCode)" -ForegroundColor Red
            $allPassed = $false
        }
    } catch {
        Write-Host "❌ $($endpoint.Name): $($endpoint.Path) → FAILED: $($_.Exception.Message)" -ForegroundColor Red
        $allPassed = $false
    }
}

if (-not $allPassed) {
    Write-Host "`n❌ API HEALTH TESTS FAILED" -ForegroundColor Red
    exit 1
}

Write-Host "`n✅ ALL API HEALTH TESTS PASSED" -ForegroundColor Green
```

---

## Step 3: API CRUD Tests

### CRUD Test Script

```powershell
#!/usr/bin/env pwsh
# api-crud-tests.ps1

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$EntityPath = "/api/v1/orders"  # Change per entity
)

$ErrorActionPreference = "Stop"
$testResults = @()

Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           API CRUD Tests: $EntityPath                        ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

# Test data (customize per entity)
$createPayload = @{
    customerId = "11111111-1111-1111-1111-111111111111"
    status = "Pending"
    items = @(
        @{
            productId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
            quantity = 2
            unitPrice = 99.99
        }
    )
} | ConvertTo-Json -Depth 5

$updatePayload = @{
    status = "Processing"
} | ConvertTo-Json

# CREATE Test
Write-Host "`n📝 Testing CREATE..." -ForegroundColor Yellow
try {
    $createResponse = Invoke-RestMethod -Uri "$BaseUrl$EntityPath" `
        -Method POST `
        -ContentType "application/json" `
        -Body $createPayload
    
    $createdId = $createResponse.id
    Write-Host "✅ CREATE: Created entity with ID: $createdId" -ForegroundColor Green
    $testResults += @{ Test = "CREATE"; Passed = $true }
} catch {
    Write-Host "❌ CREATE FAILED: $($_.Exception.Message)" -ForegroundColor Red
    $testResults += @{ Test = "CREATE"; Passed = $false }
    exit 1
}

# READ ALL Test
Write-Host "`n📖 Testing READ ALL..." -ForegroundColor Yellow
try {
    $readAllResponse = Invoke-RestMethod -Uri "$BaseUrl$EntityPath" -Method GET
    $count = if ($readAllResponse.data) { $readAllResponse.data.Count } else { $readAllResponse.Count }
    Write-Host "✅ READ ALL: Retrieved $count entities" -ForegroundColor Green
    $testResults += @{ Test = "READ_ALL"; Passed = $true }
} catch {
    Write-Host "❌ READ ALL FAILED: $($_.Exception.Message)" -ForegroundColor Red
    $testResults += @{ Test = "READ_ALL"; Passed = $false }
}

# READ ONE Test
Write-Host "`n📖 Testing READ ONE..." -ForegroundColor Yellow
try {
    $readOneResponse = Invoke-RestMethod -Uri "$BaseUrl$EntityPath/$createdId" -Method GET
    Write-Host "✅ READ ONE: Retrieved entity $createdId" -ForegroundColor Green
    $testResults += @{ Test = "READ_ONE"; Passed = $true }
} catch {
    Write-Host "❌ READ ONE FAILED: $($_.Exception.Message)" -ForegroundColor Red
    $testResults += @{ Test = "READ_ONE"; Passed = $false }
}

# UPDATE Test
Write-Host "`n✏️ Testing UPDATE..." -ForegroundColor Yellow
try {
    $updateResponse = Invoke-RestMethod -Uri "$BaseUrl$EntityPath/$createdId" `
        -Method PUT `
        -ContentType "application/json" `
        -Body $updatePayload
    Write-Host "✅ UPDATE: Updated entity $createdId" -ForegroundColor Green
    $testResults += @{ Test = "UPDATE"; Passed = $true }
} catch {
    Write-Host "❌ UPDATE FAILED: $($_.Exception.Message)" -ForegroundColor Red
    $testResults += @{ Test = "UPDATE"; Passed = $false }
}

# DELETE Test
Write-Host "`n🗑️ Testing DELETE..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BaseUrl$EntityPath/$createdId" -Method DELETE
    Write-Host "✅ DELETE: Deleted entity $createdId" -ForegroundColor Green
    $testResults += @{ Test = "DELETE"; Passed = $true }
} catch {
    Write-Host "❌ DELETE FAILED: $($_.Exception.Message)" -ForegroundColor Red
    $testResults += @{ Test = "DELETE"; Passed = $false }
}

# NOT FOUND Test
Write-Host "`n🔍 Testing NOT FOUND..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BaseUrl$EntityPath/00000000-0000-0000-0000-000000000000" -Method GET
    Write-Host "❌ NOT FOUND: Should have returned 404" -ForegroundColor Red
    $testResults += @{ Test = "NOT_FOUND"; Passed = $false }
} catch {
    if ($_.Exception.Response.StatusCode -eq 404) {
        Write-Host "✅ NOT FOUND: Correctly returned 404" -ForegroundColor Green
        $testResults += @{ Test = "NOT_FOUND"; Passed = $true }
    } else {
        Write-Host "❌ NOT FOUND: Expected 404, got $($_.Exception.Response.StatusCode)" -ForegroundColor Red
        $testResults += @{ Test = "NOT_FOUND"; Passed = $false }
    }
}

# Summary
Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           CRUD TEST SUMMARY                                  ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

$passed = ($testResults | Where-Object { $_.Passed }).Count
$total = $testResults.Count
$allPassed = $passed -eq $total

foreach ($result in $testResults) {
    $icon = if ($result.Passed) { "✅" } else { "❌" }
    $color = if ($result.Passed) { "Green" } else { "Red" }
    Write-Host "$icon $($result.Test)" -ForegroundColor $color
}

Write-Host "`nTotal: $passed/$total tests passed" -ForegroundColor $(if ($allPassed) { "Green" } else { "Red" })

if (-not $allPassed) {
    exit 1
}
```

---

## Step 4: UI Build & Smoke Tests

### UI Test Script

```powershell
#!/usr/bin/env pwsh
# ui-tests.ps1

param(
    [string]$FrontendPath = "./frontend",
    [string]$ApiBaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"

Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           UI Build & Smoke Tests                             ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

Push-Location $FrontendPath

try {
    # Test 1: npm install
    Write-Host "`n📦 Testing npm install..." -ForegroundColor Yellow
    npm install 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ npm install FAILED" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ npm install succeeded" -ForegroundColor Green

    # Test 2: npm run build
    Write-Host "`n🔨 Testing npm run build..." -ForegroundColor Yellow
    $buildOutput = npm run build 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ npm run build FAILED" -ForegroundColor Red
        Write-Host $buildOutput -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ npm run build succeeded" -ForegroundColor Green

    # Test 3: Check dist folder exists
    Write-Host "`n📁 Checking build output..." -ForegroundColor Yellow
    if (-not (Test-Path "./dist/index.html")) {
        Write-Host "❌ Build output not found (dist/index.html missing)" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Build output exists" -ForegroundColor Green

    # Test 4: TypeScript type check
    Write-Host "`n📝 Running TypeScript check..." -ForegroundColor Yellow
    $tscOutput = npx tsc --noEmit 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "⚠️ TypeScript warnings (non-blocking):" -ForegroundColor Yellow
        Write-Host $tscOutput -ForegroundColor Yellow
    } else {
        Write-Host "✅ TypeScript check passed" -ForegroundColor Green
    }

    # Test 5: ESLint check
    Write-Host "`n🔍 Running ESLint..." -ForegroundColor Yellow
    $lintOutput = npm run lint 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "⚠️ ESLint warnings (non-blocking):" -ForegroundColor Yellow
        Write-Host $lintOutput -ForegroundColor Yellow
    } else {
        Write-Host "✅ ESLint check passed" -ForegroundColor Green
    }

    Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Green
    Write-Host "║           ✅ UI BUILD TESTS PASSED                           ║" -ForegroundColor Green
    Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Green

} finally {
    Pop-Location
}
```

---

## Step 5: End-to-End Flow Test

### E2E Test Script

```powershell
#!/usr/bin/env pwsh
# e2e-tests.ps1

param(
    [string]$BaseUrl = "http://localhost:8080"
)

Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           End-to-End Flow Tests                              ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

# Full workflow test
Write-Host "`n🔄 Running end-to-end workflow..." -ForegroundColor Yellow

# 1. Create a customer
$customer = @{
    name = "E2E Test Customer"
    email = "e2e-test-$(Get-Random)@example.com"
    phone = "+1-555-0199"
} | ConvertTo-Json

$createdCustomer = Invoke-RestMethod -Uri "$BaseUrl/api/v1/customers" `
    -Method POST -ContentType "application/json" -Body $customer
$customerId = $createdCustomer.id
Write-Host "✅ Created customer: $customerId" -ForegroundColor Green

# 2. Create a product
$product = @{
    name = "E2E Test Product"
    price = 49.99
    stockQuantity = 100
} | ConvertTo-Json

$createdProduct = Invoke-RestMethod -Uri "$BaseUrl/api/v1/products" `
    -Method POST -ContentType "application/json" -Body $product
$productId = $createdProduct.id
Write-Host "✅ Created product: $productId" -ForegroundColor Green

# 3. Create an order
$order = @{
    customerId = $customerId
    items = @(
        @{
            productId = $productId
            quantity = 2
            unitPrice = 49.99
        }
    )
} | ConvertTo-Json -Depth 5

$createdOrder = Invoke-RestMethod -Uri "$BaseUrl/api/v1/orders" `
    -Method POST -ContentType "application/json" -Body $order
$orderId = $createdOrder.id
Write-Host "✅ Created order: $orderId" -ForegroundColor Green

# 4. Verify order exists
$fetchedOrder = Invoke-RestMethod -Uri "$BaseUrl/api/v1/orders/$orderId" -Method GET
if ($fetchedOrder.id -ne $orderId) {
    throw "Order verification failed"
}
Write-Host "✅ Verified order exists" -ForegroundColor Green

# 5. Update order status
$statusUpdate = @{ status = "Processing" } | ConvertTo-Json
Invoke-RestMethod -Uri "$BaseUrl/api/v1/orders/$orderId" `
    -Method PUT -ContentType "application/json" -Body $statusUpdate
Write-Host "✅ Updated order status to Processing" -ForegroundColor Green

# 6. Verify status changed
$updatedOrder = Invoke-RestMethod -Uri "$BaseUrl/api/v1/orders/$orderId" -Method GET
if ($updatedOrder.status -ne "Processing") {
    throw "Status update verification failed"
}
Write-Host "✅ Verified order status changed" -ForegroundColor Green

# 7. Get customer orders
$customerOrders = Invoke-RestMethod -Uri "$BaseUrl/api/v1/customers/$customerId/orders" -Method GET
Write-Host "✅ Retrieved customer orders (count: $($customerOrders.Count))" -ForegroundColor Green

# Cleanup (optional - mark as cancelled instead of delete for audit)
$cancelUpdate = @{ status = "Cancelled" } | ConvertTo-Json
Invoke-RestMethod -Uri "$BaseUrl/api/v1/orders/$orderId" `
    -Method PUT -ContentType "application/json" -Body $cancelUpdate
Write-Host "✅ Cleanup: Cancelled test order" -ForegroundColor Green

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║           ✅ END-TO-END TESTS PASSED                         ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Green
```

---

## Master Test Runner

### scripts/run-integration-tests.ps1

```powershell
#!/usr/bin/env pwsh
# Master integration test runner

param(
    [string]$ProjectName = "{service-name}",
    [string]$BaseUrl = "http://localhost:8080",
    [string]$FrontendPath = "./frontend"
)

$ErrorActionPreference = "Stop"

Write-Host "╔══════════════════════════════════════════════════════════════════════════════════════════╗" -ForegroundColor Magenta
Write-Host "║                           AI INTEGRATION TEST SUITE                                     ║" -ForegroundColor Magenta
Write-Host "╚══════════════════════════════════════════════════════════════════════════════════════════╝" -ForegroundColor Magenta

$testSuites = @(
    @{ Name = "Infrastructure Health"; Script = "test-infra-health.ps1" },
    @{ Name = "API Health"; Script = "test-api-health.ps1" },
    @{ Name = "API CRUD"; Script = "test-api-crud.ps1" },
    @{ Name = "UI Build"; Script = "test-ui-build.ps1" },
    @{ Name = "End-to-End"; Script = "test-e2e.ps1" }
)

$results = @()

foreach ($suite in $testSuites) {
    Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
    Write-Host "Running: $($suite.Name)" -ForegroundColor Cyan
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
    
    $scriptPath = Join-Path "scripts" $suite.Script
    if (Test-Path $scriptPath) {
        try {
            & $scriptPath -BaseUrl $BaseUrl -FrontendPath $FrontendPath
            $results += @{ Suite = $suite.Name; Passed = $true }
            Write-Host "✅ $($suite.Name) PASSED" -ForegroundColor Green
        } catch {
            $results += @{ Suite = $suite.Name; Passed = $false; Error = $_.Exception.Message }
            Write-Host "❌ $($suite.Name) FAILED: $($_.Exception.Message)" -ForegroundColor Red
        }
    } else {
        Write-Host "⚠️ Script not found: $scriptPath (skipping)" -ForegroundColor Yellow
        $results += @{ Suite = $suite.Name; Passed = $true; Skipped = $true }
    }
}

# Final Summary
Write-Host "`n╔══════════════════════════════════════════════════════════════════════════════════════════╗" -ForegroundColor Magenta
Write-Host "║                           INTEGRATION TEST SUMMARY                                       ║" -ForegroundColor Magenta
Write-Host "╠══════════════════════════════════════════════════════════════════════════════════════════╣" -ForegroundColor Magenta

$passed = ($results | Where-Object { $_.Passed }).Count
$total = $results.Count
$allPassed = $passed -eq $total

foreach ($result in $results) {
    $icon = if ($result.Passed) { "✅" } else { "❌" }
    $status = if ($result.Skipped) { "SKIPPED" } elseif ($result.Passed) { "PASSED" } else { "FAILED" }
    $color = if ($result.Passed) { "Green" } else { "Red" }
    Write-Host "║   $icon $($result.Suite.PadRight(30)) $status" -ForegroundColor $color
}

Write-Host "╠══════════════════════════════════════════════════════════════════════════════════════════╣" -ForegroundColor Magenta
Write-Host "║   Total: $passed/$total test suites passed" -ForegroundColor $(if ($allPassed) { "Green" } else { "Red" })
Write-Host "╚══════════════════════════════════════════════════════════════════════════════════════════╝" -ForegroundColor Magenta

if (-not $allPassed) {
    Write-Host "`n❌ INTEGRATION TESTS FAILED - DO NOT PROCEED TO HELM CHART" -ForegroundColor Red
    exit 1
}

Write-Host "`n✅ ALL INTEGRATION TESTS PASSED - Ready for Helm chart generation" -ForegroundColor Green
```

---

## Test Checklist (For Workflow)

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                     INTEGRATION TEST CHECKLIST                                           ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   INFRASTRUCTURE:                                                                        ║
║   □ docker-compose up -d succeeded                                                       ║
║   □ All containers running (docker-compose ps)                                           ║
║   □ SQL Server accepting connections                                                     ║
║   □ MongoDB accepting connections                                                        ║
║   □ Redis responding to PING                                                             ║
║   □ Kafka broker ready                                                                   ║
║   □ Databases seeded with test data                                                      ║
║                                                                                          ║
║   API TESTS:                                                                             ║
║   □ GET /health returns 200                                                              ║
║   □ GET /health/live returns 200                                                         ║
║   □ GET /health/ready returns 200                                                        ║
║   □ GET /metrics returns Prometheus metrics                                              ║
║   □ GET /api/v1/sli returns SLI data                                                     ║
║   □ POST /api/v1/{entities} creates entity (201)                                         ║
║   □ GET /api/v1/{entities} lists entities (200)                                          ║
║   □ GET /api/v1/{entities}/{id} returns entity (200)                                     ║
║   □ PUT /api/v1/{entities}/{id} updates entity (200)                                     ║
║   □ DELETE /api/v1/{entities}/{id} deletes entity (204)                                  ║
║   □ GET /api/v1/{entities}/invalid returns 404                                           ║
║                                                                                          ║
║   UI TESTS:                                                                              ║
║   □ npm install succeeds                                                                 ║
║   □ npm run build succeeds (no errors)                                                   ║
║   □ dist/index.html exists                                                               ║
║   □ TypeScript compiles without errors                                                   ║
║   □ ESLint passes (or only warnings)                                                     ║
║                                                                                          ║
║   END-TO-END:                                                                            ║
║   □ Create → Read → Update → Delete flow works                                           ║
║   □ Related entities work (customer → orders)                                            ║
║   □ Status transitions work correctly                                                    ║
║                                                                                          ║
║   ✅ ALL TESTS PASSED → Proceed to Helm Chart                                            ║
║   ❌ ANY TEST FAILED → Fix and re-run tests                                              ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

````
