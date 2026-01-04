---
name: ai-development-workflow
description: CRITICAL 13-phase development workflow for .NET and Go services. Starts with MANDATORY architecture analysis (Phase 0), then enforces build verification, code quality (format/lint), Docker Compose deployment, integration testing (API + UI), and final delivery. ALL tests MUST pass before Helm chart. DO NOT skip phases. DO NOT declare success until ALL phases pass. Use when scaffolding complete services.
---

# AI 13-Phase Development Workflow

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  MANDATORY DEVELOPMENT WORKFLOW  ⚠️                             ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   EVERY SERVICE GENERATION MUST FOLLOW ALL 13 PHASES!                                    ║
║                                                                                          ║
║   PHASE 0:  Architecture Analysis (MANDATORY FIRST!)  ← ai-architect skill              ║
║   ─────────────────────────────────────────────────────────────────────────              ║
║   PHASE 1:  Generate Code & Seed Files                                                   ║
║   PHASE 2:  Build Locally & Fix Errors                                                   ║
║   PHASE 2.5: Code Quality (Format & Lint)         ← ai-code-quality skill               ║
║   PHASE 3:  Run Unit Tests (80% coverage)         ← ai-unit-testing skill               ║
║   PHASE 4:  Build Docker Image                                                           ║
║   PHASE 5:  Deploy with Docker Compose (Infra + App)                                     ║
║   PHASE 6:  Seed ALL Databases                                                           ║
║   PHASE 7:  Run Integration Tests (API)           ← ai-integration-testing skill        ║
║   PHASE 8:  Run Integration Tests (UI)            ← ALL tests must pass!                 ║
║   PHASE 9:  Create Helm Chart (only after tests pass)                                    ║
║   PHASE 10: Final Status & Delivery Report                                               ║
║                                                                                          ║
║   ❌ DO NOT SKIP ANY PHASE!                                                              ║
║   ❌ DO NOT GENERATE CODE UNTIL PHASE 0 IS COMPLETE!                                     ║
║   ❌ DO NOT CREATE HELM CHART UNTIL PHASES 7-8 PASS!                                     ║
║   ❌ DO NOT DECLARE SUCCESS UNTIL ALL PHASES COMPLETE!                                   ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## Phase Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              AI DEVELOPMENT WORKFLOW                                   │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│   PHASE 0: ARCHITECTURE                                                                 │
│   ├── Ask clarifying questions                                                          │
│   ├── Decide data platforms (SQL/MongoDB/ScyllaDB/Redis/Kafka)                          │
│   ├── Define Kafka topics                                                               │
│   ├── Generate ARCHITECTURE.md                                                          │
│   └── Get user confirmation ─────────────────────────────────────────────────┐          │
│                                                                               │          │
│   PHASE 1-2: BUILD LOCALLY                                                    │          │
│   ├── Phase 1: Generate Code + Seed Scripts                                   │          │
│   └── Phase 2: Build & Fix Errors ───────────────────────────────────────┐    │          │
│                                                                           │    │          │
│   PHASE 2.5: CODE QUALITY (Format & Lint)  ← ai-code-quality skill       │    │          │
│   ├── Run formatters (dotnet format, gofmt, prettier)                     │    │          │
│   ├── Run linters (analyzers, golangci-lint, eslint)                      │    │          │
│   └── Fix any linting errors before proceeding ──────────────────────┐    │    │          │
│                                                                       │    │    │          │
│   PHASE 3-4: TEST & DOCKER                                            │    │    │          │
│   ├── Phase 3: Run Unit Tests                                         │    │    │          │
│   └── Phase 4: Build Docker Image ───────────────────────────────┐    │    │    │          │
│                                                                   │    │    │    │          │
│   PHASE 5-6: DEPLOY INFRASTRUCTURE + APP                          │    │    │    │          │
│   ├── Phase 5: docker-compose up (ALL services)                   │    │    │    │          │
│   └── Phase 6: Seed ALL databases ───────────────────────────┐    │    │    │    │          │
│                                                               │    │    │    │    │          │
│   PHASE 7-8: INTEGRATION TESTING GATE                         │    │    │    │    │          │
│   ┌───────────────────────────────────────────────────────┐   │    │    │    │    │          │
│   │  ⚠️  ALL TESTS MUST PASS BEFORE HELM CHART!           │   │    │    │    │    │          │
│   │  ├── Phase 7: API Tests (health, CRUD, error handling)│   │    │    │    │    │          │
│   │  └── Phase 8: UI Tests (build, load, navigation, E2E) │   │    │    │    │    │          │
│   │                                                        │   │    │    │    │    │          │
│   │  If ANY test fails → Fix → Rebuild → Redeploy → Retest│   │    │    │    │    │          │
│   └───────────────────────────────────────────────────────┘   │    │    │    │    │          │
│                           │                                    │    │    │    │    │          │
│                           ▼                                    │    │    │    │    │          │
│   PHASE 9-10: HELM & DELIVERY (only after tests pass)          │    │    │    │    │          │
│   ├── Phase 9: Generate Helm Chart                             │    │    │    │    │          │
│   └── Phase 10: Final Status Report + Launch Instructions      │    │    │    │    │          │
│                                                                │    │    │    │    │          │
└────────────────────────────────────────────────────────────────┴────┴────┴────┴────┴──────────┘
```

---

## PHASE 0: Architecture Analysis (MANDATORY FIRST!)

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  ARCHITECTURE BEFORE CODE  ⚠️                                   ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   📖 READ AND FOLLOW: ai-architect skill                                                ║
║                                                                                          ║
║   BEFORE generating ANY code, you MUST:                                                  ║
║                                                                                          ║
║   1. UNDERSTAND the use-case (ask clarifying questions if needed)                        ║
║   2. DECIDE which data goes in which platform:                                           ║
║      • SQL Server  → Transactional data, relationships, ACID                             ║
║      • MongoDB     → Documents, hierarchical data, flexible schema                       ║
║      • ScyllaDB    → Time-series, events, high-throughput writes                         ║
║      • Redis       → Cache, sessions, rate limiting, real-time                           ║
║      • Kafka       → Event streaming, async communication, audit                         ║
║   3. DEFINE all Kafka topics needed                                                      ║
║   4. GENERATE ARCHITECTURE.md document                                                   ║
║   5. GET USER CONFIRMATION before proceeding                                             ║
║                                                                                          ║
║   ❌ DO NOT proceed to Phase 1 until architecture is approved!                           ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### Phase 0 Outputs:
- `ARCHITECTURE.md` - Complete architecture document with:
  - Data platform allocation per entity
  - Kafka topic definitions
  - Data flow diagram
  - API endpoint summary
  - Caching strategy
  - Event contracts

### Phase 0 Checklist:
- [ ] Asked clarifying questions about data characteristics
- [ ] Identified which entities need ACID transactions (→ SQL Server)
- [ ] Identified hierarchical/flexible data (→ MongoDB)
- [ ] Identified time-series/event data (→ ScyllaDB)
- [ ] Identified hot/cached data (→ Redis)
- [ ] Defined all Kafka topics for events
- [ ] Generated ARCHITECTURE.md
- [ ] User confirmed architecture

---

## PHASE 1: Generate Code & Seed Files

### Required Files to Generate

After architecture is approved, generate the main service code and supporting files:

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        📖 SEED SCRIPTS (MANDATORY)                                        ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   Read the ai-data-seeding skill for COMPLETE templates!                                ║
║                                                                                          ║
║   Generate ALL scripts that apply to your service's data platforms:                      ║
║                                                                                          ║
║   scripts/                                                                               ║
║   ├── seed-all.ps1          # Master script (ALWAYS generate)                            ║
║   ├── seed-sqlserver.sql    # If using SQL Server                                        ║
║   ├── seed-mongodb.js       # If using MongoDB                                           ║
║   ├── seed-scylladb.cql     # If using ScyllaDB                                          ║
║   ├── seed-kafka.ps1        # If using Kafka                                             ║
║   └── seed-redis.ps1        # If using Redis                                             ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

#### 1.1 Database Seed Scripts

Generate seed scripts for EVERY data platform used by the service.
See **ai-data-seeding** skill for complete templates including:
- SQL Server: Table creation, indexes, foreign keys, sample data
- MongoDB: Collections, schema validation, indexes, sample documents
- ScyllaDB: Keyspace, tables with proper partition keys, TTL, sample data
- Kafka: Topic creation with partitions and replication
- Redis: Configuration, feature flags, cache warming

#### 1.2 Kafka Topics Script (scripts/create-kafka-topics.ps1)

```powershell
$topics = @(
    "{service-name}.{entity}.created",
    "{service-name}.{entity}.updated",
    "{service-name}.{entity}.deleted"
)

foreach ($topic in $topics) {
    docker exec -it {project}-kafka-1 /opt/kafka/bin/kafka-topics.sh `
        --bootstrap-server localhost:9092 `
        --create --topic $topic `
        --partitions 3 --replication-factor 1 --if-not-exists
}
```

#### 1.3 Docker Compose (docker-compose.yml)

Must include ALL required infrastructure with official images only:
- API service (built from Dockerfile)
- SQL Server: `mcr.microsoft.com/mssql/server:2022-latest`
- MongoDB: `mongo:7`
- ScyllaDB: `scylladb/scylla:latest`
- Redis: `redis:7-alpine`
- Kafka: `apache/kafka:latest` (KRaft mode - NO Zookeeper!)

---

## PHASE 2: Build Locally & Fix Errors

### .NET Build

```powershell
cd {project-path}
dotnet restore
dotnet build --no-restore
```

### Go Build

```bash
cd services/go/{service-name}
go mod download
go mod tidy
go build -o bin/server ./cmd/server
```

### Error Handling Loop

```
╔════════════════════════════════════════════════════════════════╗
║                     BUILD ERROR LOOP                            ║
╠════════════════════════════════════════════════════════════════╣
║                                                                 ║
║   while (build_fails) {                                         ║
║       1. Read error messages carefully                          ║
║       2. Fix EACH error in source files                         ║
║       3. Rebuild                                                ║
║   }                                                             ║
║                                                                 ║
║   ⚠️  DO NOT PROCEED TO PHASE 3 UNTIL BUILD SUCCEEDS!           ║
║                                                                 ║
╚════════════════════════════════════════════════════════════════╝
```

### Common Build Errors

| Language | Error | Fix |
|----------|-------|-----|
| .NET | `CS0246: Type not found` | Add missing `using` statement |
| .NET | `CS0234: Namespace not found` | Add package to .csproj |
| Go | `undefined: X` | Add missing import |
| Go | `cannot find module` | Run `go mod tidy` |

---

## PHASE 2.5: Code Quality (Format & Lint)

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  CODE QUALITY IS MANDATORY  ⚠️                                  ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   📖 READ AND FOLLOW: ai-code-quality skill                                             ║
║                                                                                          ║
║   AFTER successful build, BEFORE running tests, you MUST:                                ║
║                                                                                          ║
║   1. RUN formatters to auto-fix style issues                                             ║
║   2. RUN linters to catch potential bugs                                                 ║
║   3. FIX any linting errors                                                              ║
║                                                                                          ║
║   ❌ DO NOT proceed to Phase 3 until all formatting/linting passes!                      ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### .NET Code Quality

```powershell
# Format code
dotnet format --verbosity normal

# Build with analyzers (treat warnings as errors)
dotnet build /p:EnforceCodeStyleInBuild=true
```

### Go Code Quality

```bash
# Format code
gofmt -w -s .
goimports -w .

# Run linter
golangci-lint run
```

### TypeScript/React Code Quality

```bash
cd {service-name}-ui

# Format with Prettier
npx prettier --write "src/**/*.{ts,tsx,css,json}"

# Lint with ESLint
npx eslint src/ --fix

# Type check
npx tsc --noEmit
```

### Python Code Quality

```bash
# Format code
black .
isort .

# Lint
ruff check .
mypy src/
```

### Code Quality Checklist

```
PHASE 2.5 CHECKLIST:
□ Formatters have been run (auto-fix applied)
□ All linters pass with no errors
□ Warnings are either fixed or documented
□ Type checking passes (TypeScript/Python)
```

**If linting fails:**
1. Read error messages carefully
2. Fix style/bug issues in source files
3. Re-run linters until all pass
4. Then proceed to Phase 3

---

## PHASE 3: Run Unit Tests (80% Coverage Required)

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  UNIT TESTING IS MANDATORY  ⚠️                                  ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   📖 READ AND FOLLOW: ai-unit-testing skill                                             ║
║                                                                                          ║
║   ALL code MUST have unit tests with ≥80% coverage!                                      ║
║                                                                                          ║
║   ❌ DO NOT proceed to Phase 4 until all tests pass!                                     ║
║   ❌ DO NOT proceed with <80% coverage without explicit approval!                        ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### .NET Unit Tests

```powershell
# Run tests with coverage
dotnet test --collect:"XPlat Code Coverage" --results-directory ./coverage

# Verify 80% threshold
$coverage = [xml](Get-Content ./coverage/**/coverage.cobertura.xml)
$lineRate = [double]$coverage.coverage.'line-rate' * 100
if ($lineRate -lt 80) { Write-Host "❌ Coverage $lineRate% below 80%"; exit 1 }
Write-Host "✅ Coverage: $lineRate%"
```

### Go Unit Tests

```bash
# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# Verify 80% threshold
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then echo "❌ Coverage below 80%"; exit 1; fi
echo "✅ Coverage: ${COVERAGE}%"
```

### TypeScript/React Unit Tests

```bash
cd {service-name}-ui
npm run test:ci  # Runs with --coverage and 80% threshold
```

### Python Unit Tests

```bash
pytest --cov=src --cov-report=term-missing --cov-fail-under=80
```

**If tests fail:**
1. Read test output to identify failures
2. Fix the code or test
3. Add more tests if coverage < 80%
4. Repeat until ALL tests pass AND coverage ≥ 80%

---

## PHASE 4: Build Docker Image

### .NET

```powershell
$env:GITHUB_TOKEN = "your-token"
docker build -t {service-name}:latest --build-arg GITHUB_TOKEN=$env:GITHUB_TOKEN .
```

### Go

```bash
docker build -t {service-name}:latest .
# Or with docker-compose
docker-compose build {service-name}
```

**If Docker build fails:**
1. Check Dockerfile syntax
2. Verify all files are present
3. Check build errors in container logs
4. Fix and repeat PHASE 4

---

## PHASE 5: Deploy with Docker Compose (Infrastructure + Application)

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  DEPLOY BOTH INFRA AND APP  ⚠️                                  ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   Docker Compose MUST start ALL of these:                                                ║
║                                                                                          ║
║   INFRASTRUCTURE:                                                                        ║
║   ├── SQL Server (mcr.microsoft.com/mssql/server:2022-latest)                            ║
║   ├── MongoDB (mongo:7)                                                                  ║
║   ├── ScyllaDB (scylladb/scylla:latest) - if used                                        ║
║   ├── Redis (redis:7-alpine)                                                             ║
║   └── Kafka (apache/kafka:latest) - KRaft mode, NO Zookeeper                             ║
║                                                                                          ║
║   APPLICATION:                                                                           ║
║   ├── Backend API Service (from Dockerfile)                                              ║
║   └── Frontend (optional - can run via npm run dev)                                      ║
║                                                                                          ║
║   ALL containers MUST be healthy before proceeding!                                      ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### 5.1 Start All Services

```powershell
# Start all infrastructure AND application
docker-compose up -d

# Verify all containers are running
docker-compose ps

# Watch logs for startup errors
docker-compose logs -f {service-name}
```

### 5.2 Verify Container Health

```powershell
# Check container status
$containers = docker-compose ps --format json | ConvertFrom-Json
foreach ($container in $containers) {
    Write-Host "$($container.Name): $($container.State)"
}

# Wait for services to be healthy (with timeout)
$timeout = 120
$elapsed = 0
while ($elapsed -lt $timeout) {
    $unhealthy = docker-compose ps | Select-String "unhealthy|starting"
    if (-not $unhealthy) {
        Write-Host "✅ All containers healthy" -ForegroundColor Green
        break
    }
    Start-Sleep -Seconds 5
    $elapsed += 5
    Write-Host "⏳ Waiting for containers... ($elapsed/$timeout seconds)"
}
```

### 5.3 Infrastructure Connectivity Check

```powershell
# SQL Server
docker exec {project}-sqlserver-1 /opt/mssql-tools18/bin/sqlcmd `
    -S localhost -U sa -P "YourStrong!Password" -C -Q "SELECT 1"

# MongoDB
docker exec {project}-mongodb-1 mongosh --quiet --eval "db.runCommand({ping:1})"

# Redis
docker exec {project}-redis-1 redis-cli ping

# Kafka
docker exec {project}-kafka-1 /opt/kafka/bin/kafka-topics.sh `
    --bootstrap-server localhost:9092 --list
```

**If containers fail to start:**
1. Check logs: `docker-compose logs {service-name}`
2. Verify environment variables in docker-compose.yml
3. Check health of dependencies
4. Fix configuration and repeat PHASE 5

---

## PHASE 6: Seed ALL Databases (MANDATORY!)

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  DATABASE SEEDING IS REQUIRED  ⚠️                               ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   ALL databases MUST be initialized with schema and seed data before testing!            ║
║   The API WILL FAIL if databases/keyspaces/topics don't exist!                           ║
║                                                                                          ║
║   📖 REFER TO: ai-data-seeding skill for comprehensive templates!                       ║
║                                                                                          ║
║   Required seed scripts (generate ALL that apply to your service):                       ║
║     ✅ scripts/seed-sqlserver.sql  - SQL Server tables and data                          ║
║     ✅ scripts/seed-mongodb.js     - MongoDB collections and documents                   ║
║     ✅ scripts/seed-scylladb.cql   - ScyllaDB keyspace, tables, and data                 ║
║     ✅ scripts/seed-kafka.ps1      - Kafka topic creation                                ║
║     ✅ scripts/seed-redis.ps1      - Redis initialization and cache priming             ║
║     ✅ scripts/seed-all.ps1        - MASTER script that runs ALL above                   ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### 6.0 Run Master Seed Script (RECOMMENDED)

```powershell
# Single command to seed ALL platforms
./scripts/seed-all.ps1 -ComposeProject "{service-name}"
```

### 6.1 SQL Server

```powershell
docker exec -it {project}-sqlserver-1 /opt/mssql-tools18/bin/sqlcmd `
    -S localhost -U sa -P "YourStrong!Password" -C `
    -Q "CREATE DATABASE {DatabaseName}"

# Run seed script
docker exec -it {project}-sqlserver-1 /opt/mssql-tools18/bin/sqlcmd `
    -S localhost -U sa -P "YourStrong!Password" -C `
    -d {DatabaseName} -i /scripts/seed-sqlserver.sql
```

### 6.2 MongoDB

```powershell
# Run MongoDB seed script
docker exec -it {project}-mongodb-1 mongosh --quiet < scripts/seed-mongodb.js
```

### 6.3 ScyllaDB

```powershell
# Wait for ScyllaDB to be ready (takes longer than other DBs)
Start-Sleep -Seconds 30

# Run ScyllaDB seed script
docker exec -it {project}-scylladb-1 cqlsh -f /scripts/seed-scylladb.cql
```

### 6.4 Kafka Topics

```powershell
# Run Kafka topic creation script
./scripts/seed-kafka.ps1 -ComposeProject "{project}"

# Verify topics
docker exec -it {project}-kafka-1 /opt/kafka/bin/kafka-topics.sh `
    --bootstrap-server localhost:9092 --list
```

### 6.5 Redis

```powershell
# Run Redis initialization script
./scripts/seed-redis.ps1 -ComposeProject "{project}"

# Verify connection
docker exec -it {project}-redis-1 redis-cli ping
# Expected: PONG
```

---

## PHASE 7: Run Integration Tests - API

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  INTEGRATION TESTING GATE  ⚠️                                   ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   📖 READ AND FOLLOW: ai-integration-testing skill                                      ║
║                                                                                          ║
║   ALL API tests MUST pass before proceeding to UI tests!                                 ║
║                                                                                          ║
║   ❌ DO NOT proceed to Phase 8 until ALL API tests pass!                                 ║
║   ❌ DO NOT create Helm chart with failing tests!                                        ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### 7.1 Health Endpoint Tests

```powershell
# Test all health endpoints
$baseUrl = "http://localhost:8080"

# Health check
$health = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
Write-Host "✅ /health: $($health.status)" -ForegroundColor Green

# Liveness probe
$live = Invoke-RestMethod -Uri "$baseUrl/health/live" -Method GET
Write-Host "✅ /health/live: OK" -ForegroundColor Green

# Readiness probe
$ready = Invoke-RestMethod -Uri "$baseUrl/health/ready" -Method GET
Write-Host "✅ /health/ready: OK" -ForegroundColor Green

# Metrics (Prometheus)
$metrics = Invoke-WebRequest -Uri "$baseUrl/metrics" -Method GET
Write-Host "✅ /metrics: $($metrics.Content.Length) bytes" -ForegroundColor Green

# SLI endpoint
$sli = Invoke-RestMethod -Uri "$baseUrl/api/v1/sli" -Method GET
Write-Host "✅ /api/v1/sli: OK" -ForegroundColor Green
```

### 7.2 CRUD Endpoint Tests

```powershell
$baseUrl = "http://localhost:8080/api/v1"

# CREATE Test
$createPayload = @{ name = "Test Entity"; /* other fields */ } | ConvertTo-Json
$created = Invoke-RestMethod -Uri "$baseUrl/{entities}" `
    -Method POST -ContentType "application/json" -Body $createPayload
$entityId = $created.id
Write-Host "✅ CREATE: $entityId" -ForegroundColor Green

# READ ALL Test
$list = Invoke-RestMethod -Uri "$baseUrl/{entities}" -Method GET
Write-Host "✅ READ ALL: $($list.Count) entities" -ForegroundColor Green

# READ ONE Test
$entity = Invoke-RestMethod -Uri "$baseUrl/{entities}/$entityId" -Method GET
Write-Host "✅ READ ONE: $($entity.id)" -ForegroundColor Green

# UPDATE Test
$updatePayload = @{ name = "Updated Entity" } | ConvertTo-Json
$updated = Invoke-RestMethod -Uri "$baseUrl/{entities}/$entityId" `
    -Method PUT -ContentType "application/json" -Body $updatePayload
Write-Host "✅ UPDATE: $($updated.name)" -ForegroundColor Green

# DELETE Test
Invoke-RestMethod -Uri "$baseUrl/{entities}/$entityId" -Method DELETE
Write-Host "✅ DELETE: $entityId removed" -ForegroundColor Green

# NOT FOUND Test (should return 404)
try {
    Invoke-RestMethod -Uri "$baseUrl/{entities}/00000000-0000-0000-0000-000000000000" -Method GET
    Write-Host "❌ NOT FOUND: Should have returned 404" -ForegroundColor Red
    exit 1
} catch {
    if ($_.Exception.Response.StatusCode -eq 404) {
        Write-Host "✅ NOT FOUND: Correctly returned 404" -ForegroundColor Green
    }
}
```

### 7.3 API Test Summary

```
API TEST CHECKLIST:
□ GET /health returns 200
□ GET /health/live returns 200
□ GET /health/ready returns 200
□ GET /metrics returns Prometheus data
□ GET /api/v1/sli returns SLI metrics
□ POST /api/v1/{entities} creates entity (201)
□ GET /api/v1/{entities} lists entities (200)
□ GET /api/v1/{entities}/{id} returns entity (200)
□ PUT /api/v1/{entities}/{id} updates entity (200)
□ DELETE /api/v1/{entities}/{id} deletes entity (204)
□ GET /api/v1/{entities}/invalid returns 404
```

**If API tests fail:**
1. Check API response error messages
2. Review handler/controller code
3. Check service layer and infrastructure connections
4. Verify databases are seeded (PHASE 6!)
5. Fix code → Rebuild Docker → Restart → Re-seed → Retest

---

## PHASE 8: Run Integration Tests - UI

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  UI TESTS MUST PASS  ⚠️                                         ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   📖 READ AND FOLLOW: ai-integration-testing skill                                      ║
║                                                                                          ║
║   ALL UI tests MUST pass before creating Helm chart!                                     ║
║                                                                                          ║
║   ❌ DO NOT proceed to Phase 9 until ALL UI tests pass!                                  ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### 8.1 UI Build Test

```powershell
cd frontend

# Install dependencies
Write-Host "📦 Installing dependencies..." -ForegroundColor Cyan
npm install
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ npm install FAILED" -ForegroundColor Red
    exit 1
}
Write-Host "✅ npm install succeeded" -ForegroundColor Green

# Build for production
Write-Host "🔨 Building UI..." -ForegroundColor Cyan
npm run build 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ npm run build FAILED" -ForegroundColor Red
    exit 1
}
Write-Host "✅ npm run build succeeded" -ForegroundColor Green

# Verify build output
if (-not (Test-Path "./dist/index.html")) {
    Write-Host "❌ Build output not found" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Build output verified" -ForegroundColor Green
```

### 8.2 TypeScript & Lint Check

```powershell
# TypeScript check
Write-Host "📝 Running TypeScript check..." -ForegroundColor Cyan
npx tsc --noEmit
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ TypeScript errors found" -ForegroundColor Red
    exit 1
}
Write-Host "✅ TypeScript check passed" -ForegroundColor Green

# ESLint check
Write-Host "🔍 Running ESLint..." -ForegroundColor Cyan
npm run lint
if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️ ESLint warnings (review but non-blocking)" -ForegroundColor Yellow
}
```

### 8.3 UI Smoke Test Checklist

After starting the dev server (`npm run dev`), verify:

```
UI TEST CHECKLIST:
□ npm install succeeds
□ npm run build succeeds (no TypeScript errors)
□ dist/index.html exists
□ npm run dev starts without errors
□ Dashboard page loads at http://localhost:3000
□ Navigation works to all pages
□ No JavaScript errors in browser console
□ API calls succeed (check Network tab)
□ CRUD operations work through UI
□ Data displays correctly in tables/lists
□ Forms submit successfully
□ Error states display properly
```

### 8.4 End-to-End Flow Test

```powershell
# Run E2E test against running services
# This verifies the full stack works together

# 1. Create via API
$customer = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/customers" `
    -Method POST -ContentType "application/json" `
    -Body '{"name":"E2E Test","email":"e2e@test.com"}'

# 2. Verify appears in UI (manual check or automated with Playwright/Cypress)
Write-Host "✅ Created customer: $($customer.id)" -ForegroundColor Green

# 3. Update via API
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/customers/$($customer.id)" `
    -Method PUT -ContentType "application/json" `
    -Body '{"name":"E2E Updated"}'
Write-Host "✅ Updated customer" -ForegroundColor Green

# 4. Verify update reflected in UI (manual check)

# 5. Create order for customer
$order = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/orders" `
    -Method POST -ContentType "application/json" `
    -Body "{`"customerId`":`"$($customer.id)`",`"items`":[{`"productId`":`"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa`",`"quantity`":1,`"unitPrice`":99.99}]}"
Write-Host "✅ Created order: $($order.id)" -ForegroundColor Green

Write-Host "`n✅ END-TO-END FLOW PASSED" -ForegroundColor Green
```

---

## PHASE 9: Create Helm Chart (Only After Tests Pass!)

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  TESTS MUST PASS FIRST  ⚠️                                      ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   DO NOT create Helm chart unless:                                                       ║
║   ✅ Phase 7 (API Tests) PASSED                                                          ║
║   ✅ Phase 8 (UI Tests) PASSED                                                           ║
║                                                                                          ║
║   If tests failed, go back and fix before proceeding!                                    ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

Generate a self-contained Helm chart (NO Bitnami dependencies):

```
helm/{service-name}/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── _helpers.tpl
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── hpa.yaml
│   └── pdb.yaml
└── charts/           # Self-contained sub-charts
    ├── mongodb/
    ├── redis/
    ├── kafka/
    └── sqlserver/
```

### Deploy to Kubernetes

```powershell
cd helm/{service-name}
helm dependency update
helm upgrade --install {service-name} . -n sandbox --create-namespace
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name={service-name} -n sandbox --timeout=300s
```

---

## PHASE 10: Final Status & Delivery Report

**Only after ALL phases pass (0-10), provide this final status:**

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                         ✅ FINAL STATUS REPORT                                           ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   Phase 0:   Architecture Analysis ............... ✅ PASSED                             ║
║   ─────────────────────────────────────────────────────────────                          ║
║   Phase 1:   Generate Code & Seed Files .......... ✅ PASSED                             ║
║   Phase 2:   Build Locally ....................... ✅ PASSED                             ║
║   Phase 2.5: Code Quality (Format & Lint) ........ ✅ PASSED                             ║
║   Phase 3:   Unit Tests (≥80% coverage) .......... ✅ PASSED                             ║
║   Phase 4:   Docker Build ........................ ✅ PASSED                             ║
║   Phase 5:   Deploy Docker Compose ............... ✅ PASSED                             ║
║   Phase 6:   Seed Databases ...................... ✅ PASSED                             ║
║   ─────────────────────────────────────────────────────────────                          ║
║   Phase 7:   Integration Tests (API) ............. ✅ PASSED                             ║
║   Phase 8:   Integration Tests (UI) .............. ✅ PASSED                             ║
║   ─────────────────────────────────────────────────────────────                          ║
║   Phase 9:   Helm Chart .......................... ✅ PASSED                             ║
║   Phase 10:  Final Delivery ...................... ✅ PASSED                             ║
║                                                                                          ║
║   🎉 ALL PHASES COMPLETE - SERVICE IS READY!                                             ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### Launch Instructions Template

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                         🚀 LAUNCH INSTRUCTIONS                                           ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   To start the application:                                                              ║
║                                                                                          ║
║   1. Start all services:                                                                 ║
║      docker-compose up -d                                                                ║
║                                                                                          ║
║   2. Wait for services to be healthy (30-60 seconds)                                     ║
║                                                                                          ║
║   3. Seed databases (first time only):                                                   ║
║      ./scripts/seed-all.ps1                                                              ║
║                                                                                          ║
║   4. Access Points:                                                                      ║
║      • UI:      http://localhost:3000                                                    ║
║      • API:     http://localhost:8080/api/v1                                             ║
║      • Health:  http://localhost:8080/health                                             ║
║      • Metrics: http://localhost:8080/metrics                                            ║
║      • SLI:     http://localhost:8080/api/v1/sli                                         ║
║                                                                                          ║
║   5. To stop:                                                                            ║
║      docker-compose down                                                                 ║
║                                                                                          ║
║   6. For Kubernetes:                                                                     ║
║      helm upgrade --install {service-name} ./helm/{service-name} -n sandbox              ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## 🔁 ITERATION LOOP

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                          ITERATE UNTIL SUCCESS                                           ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   while (any_phase_fails) {                                                              ║
║       1. Identify the error from logs/output                                             ║
║       2. Fix the code                                                                    ║
║       3. Rebuild: docker-compose build                                                   ║
║       4. Restart: docker-compose up -d                                                   ║
║       5. Re-seed if needed: ./scripts/seed-all.ps1                                       ║
║       6. Retest: repeat failed phase                                                     ║
║   }                                                                                      ║
║                                                                                          ║
║   ❌ DO NOT DECLARE SUCCESS UNTIL ALL PHASES PASS!                                       ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## ✅ FINAL VERIFICATION CHECKLIST

Before telling the user "your service is ready", verify ALL of these:

```
PHASE 1: SEED FILES
□ scripts/seed-database.sql (or appropriate seed script)
□ scripts/create-kafka-topics.ps1
□ docker-compose.yml with ALL infrastructure

PHASE 2: BUILD
□ Build succeeded with 0 errors
□ All compile errors fixed

PHASE 2.5: CODE QUALITY
□ Formatters run (dotnet format, gofmt, prettier)
□ Linters pass (analyzers, golangci-lint, eslint)
□ Type checking passes (TypeScript/Python)
□ No unresolved warnings

PHASE 3: UNIT TESTS
□ All unit tests pass
□ Coverage ≥ 80%

PHASE 4: DOCKER BUILD
□ Docker image built successfully

PHASE 5: CONTAINERS
□ All containers start and stay healthy

PHASE 6: SEEDING
□ SQL Server database created and seeded
□ MongoDB collections created and seeded
□ ScyllaDB keyspace and tables created
□ Kafka topics created
□ Redis accessible

PHASE 7: API
□ /health returns 200
□ /metrics returns Prometheus data
□ /api/v1/sli returns SLI metrics
□ All CRUD endpoints respond correctly

PHASE 8: FRONTEND
□ UI loads without errors
□ All pages accessible
□ CRUD operations work

PHASE 9: HELM
□ Chart structure complete
□ Deploys to Kubernetes successfully

PHASE 10: DOCUMENTATION
□ Launch instructions provided
□ All access URLs listed
□ Troubleshooting guide included
```
