# AI Scaffolder - Agent Skills

This folder contains **Agent Skills** that teach GitHub Copilot how to generate ai-compliant code for 100+ developers.

## What are Agent Skills?

Agent Skills are folders containing instructions, scripts, and resources that Copilot **automatically loads** when relevant to your prompt. Unlike a single massive instruction file, skills provide:

- **Progressive disclosure** - Only relevant skills are loaded
- **Better compliance** - Focused instructions are followed more reliably
- **Bundled resources** - Each skill can include templates and examples
- **Maintainability** - Easier to update specific patterns

## Skills Structure

```
.github/skills/
├── README.md                           # This file
│
├── ai-development-workflow/           # 🔴 CRITICAL - 13-Phase Development Workflow
│   └── SKILL.md                        # Architecture → Build → Quality → Test → Docker → Seed → Verify
│
├── ai-unit-testing/                   # 🔴 CRITICAL - Unit Tests with 80% Coverage (Phase 3)
│   └── SKILL.md                        # xUnit, testify, pytest, vitest, coverage thresholds
│
├── ai-docker-images/                  # 🔴 CRITICAL - Official Docker Images Only
│   └── SKILL.md                        # NO Bitnami, NO Confluent - mcr.microsoft.com, mongo, apache/kafka
│
├── ai-error-handling/                 # Error codes and handling patterns
│   └── SKILL.md                        # ServiceError patterns, error codes (.NET & Go)
│
├── ai-core-packages-go/               # 🔴 CRITICAL - Go Package Enforcement
│   └── SKILL.md                        # core/go/logger, core/go/errors, core/go/infrastructure
│
├── ai-core-packages-dotnet/           # 🔴 CRITICAL - .NET Package Enforcement
│   └── SKILL.md                        # Core.Logger, Core.Errors, Core.Infrastructure
│
├── ai-infrastructure-clients/         # Data access patterns (.NET & Go)
│   └── SKILL.md                        # Redis, Kafka, MongoDB, SQL Server, ScyllaDB
│
├── ai-helm-charts/                    # Kubernetes Deployment
│   └── SKILL.md                        # Self-contained Helm charts (NO Bitnami)
│
├── ai-react-ui/                       # React Frontend Development
│   └── SKILL.md                        # Dark Tech Theme, Vite + Tailwind + Zustand
│
├── ai-logging-patterns/               # Logging patterns (.NET & Go)
│   └── SKILL.md                        # Structured JSON logging, correlation IDs
│
├── ai-scaffold-service-dotnet/        # Full .NET Service Scaffolding
│   └── SKILL.md                        # Complete microservice generation
│
├── ai-scaffold-service-go/            # Full Go Service Scaffolding
│   └── SKILL.md                        # Complete microservice generation
│
└── ai-sli-middleware/                 # SLI Tracking (.NET & Go)
    └── SKILL.md                        # Availability, latency, throughput metrics
```

## 13-Phase Development Workflow

The `ai-development-workflow` skill enforces a comprehensive 13-phase process:

```
╔═══════════════════════════════════════════════════════════════════════════════════════╗
║                        AI 13-PHASE DEVELOPMENT WORKFLOW                              ║
╠═══════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                       ║
║   PHASE 0:  Architecture Analysis (MANDATORY FIRST!)                                  ║
║             → Define data platforms, Kafka topics, generate ARCHITECTURE.md           ║
║   ──────────────────────────────────────────────────────────────────────────          ║
║   PHASE 1:  Generate Code & Seed Files                                                ║
║   PHASE 2:  Build Locally & Fix Errors                                                ║
║   PHASE 2.5: Code Quality (Format & Lint)         ← dotnet format, golangci-lint      ║
║   PHASE 3:  Run Unit Tests (80% coverage)         ← xUnit, testify, pytest            ║
║   PHASE 4:  Build Docker Image                                                        ║
║   PHASE 5:  Deploy with Docker Compose (Infra + App)                                  ║
║   PHASE 6:  Seed ALL Databases                    ← SQL, MongoDB, ScyllaDB, Kafka     ║
║   ──────────────────────────────────────────────────────────────────────────          ║
║   PHASE 7:  Integration Tests (API)               ← Health, CRUD, error handling      ║
║   PHASE 8:  Integration Tests (UI)                ← Build, load, navigation           ║
║   ──────────────────────────────────────────────────────────────────────────          ║
║   PHASE 9:  Create Helm Chart (only after tests pass)                                 ║
║   PHASE 10: Final Status & Delivery Report                                            ║
║                                                                                       ║
║   ❌ DO NOT SKIP ANY PHASE!                                                           ║
║   ❌ DO NOT CREATE HELM CHART UNTIL PHASES 7-8 PASS!                                  ║
║   ❌ DO NOT DECLARE SUCCESS UNTIL ALL PHASES COMPLETE!                                ║
║                                                                                       ║
╚═══════════════════════════════════════════════════════════════════════════════════════╝
```

## How Skills are Triggered

Copilot reads the `description` field in each skill's YAML frontmatter and decides when to load it:

| When Developer Says... | Skill Loaded |
|------------------------|--------------|
| "Create a service", "implement system", "scaffold" | `ai-development-workflow` (13-phase process) |
| "Create a .NET service" | `ai-scaffold-service-dotnet`, `ai-core-packages-dotnet` |
| "Create a Go service" | `ai-scaffold-service-go`, `ai-core-packages-go` |
| "unit tests", "test coverage", "80% coverage" | `ai-unit-testing` |
| "Add logging to this service" | `ai-logging-patterns` |
| "Create error codes", "handle errors" | `ai-error-handling` |
| "Add Redis", "Add Kafka", "Add MongoDB" | `ai-infrastructure-clients` |
| "Add SLI tracking", "Add metrics" | `ai-sli-middleware` |
| "Create Helm chart", "Kubernetes deployment" | `ai-helm-charts` |
| "Create Dockerfile", "docker-compose" | `ai-docker-images` |
| "Create React UI", "frontend", "dashboard" | `ai-react-ui` |

## Critical Skills (Always Enforced)

The `ai-core-packages-*` skills are the most important - they enforce:

### ❌ NEVER Use These Packages
- `Serilog` / `logrus` / `zerolog` → Use Core.Logger
- `StackExchange.Redis` / `go-redis` → Use Core.Infrastructure
- `Confluent.Kafka` / `sarama` → Use Core.Infrastructure
- `MongoDB.Driver` / `mongo-driver` → Use Core.Infrastructure
- `Polly` (raw) → Use Core.Reliability

### ✅ ALWAYS Use Core Packages

**.NET (from GitHub Packages):**
- `Core.Logger` - Structured JSON logging
- `Core.Errors` - Standardized error handling
- `Core.Infrastructure` - Redis, Kafka, MongoDB, SQL Server, ScyllaDB clients
- `Core.Metrics` - Prometheus metrics
- `Core.Sli` - SLI tracking middleware
- `Core.Config` - Configuration management
- `Core.Reliability` - Circuit breaker, retry, timeout

**Go (from core/go module):**
- `core/go/logger` - Structured JSON logging
- `core/go/errors` - Standardized error handling
- `core/go/infrastructure/*` - Database and messaging clients
- `core/go/metrics` - Prometheus metrics
- `core/go/sli` - SLI tracking middleware

## Docker Images (MANDATORY)

The `ai-docker-images` skill enforces official images only:

| Service | Required Image | ❌ DO NOT USE |
|---------|---------------|---------------|
| SQL Server | `mcr.microsoft.com/mssql/server:2022-latest` | Bitnami |
| MongoDB | `mongo:7` | Bitnami |
| Redis | `redis:7-alpine` | Bitnami |
| Kafka | `apache/kafka:latest` (KRaft mode) | Bitnami, Confluent |
| ScyllaDB | `scylladb/scylla:latest` | Cassandra |

## For Engineers

Just use Copilot normally! When you ask it to generate code, it will automatically:
1. Detect you're in the AI Scaffolder workspace
2. Load the relevant skills based on your request
3. Follow the 13-phase workflow for complete services
4. Generate code that uses Core packages correctly

### Example Prompts

```
"Implement a bookstore management system with books, authors, customers, rentals, and purchases"
→ Triggers: ai-development-workflow, ai-scaffold-service-dotnet, ai-core-packages-dotnet

"Create a Go microservice for order processing"
→ Triggers: ai-development-workflow, ai-scaffold-service-go, ai-core-packages-go

"Add unit tests with 80% coverage"
→ Triggers: ai-unit-testing

"Create a React dashboard for the service"
→ Triggers: ai-react-ui
```

## For Maintainers

To update a skill:
1. Edit the `SKILL.md` file in the skill folder
2. Update the `description` field to control when it's triggered
3. Test by asking Copilot to perform the related task
4. Verify the skill is loaded and instructions are followed

### Skill File Format

```markdown
---
name: ai-skill-name
description: Clear description of when to use this skill. Include keywords that trigger loading.
---

# Skill Title

## Instructions
...
```
