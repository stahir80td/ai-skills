# Order Management System - Full Stack Generation Prompt

> **Purpose:** This prompt exercises ALL AI Agent Skills to generate a complete Order Management System using every pattern, database, and standard defined in the skills.

---

## 📚 Required Skills Reference

This prompt uses the AI Agent Skills. Load and follow each skill's patterns:

| Skill | Purpose | When Used |
|-------|---------|----------|
| `ai-development-workflow` | 🔴 CRITICAL 13-phase workflow | Overall process governance |
| `ai-core-packages-dotnet` | 🔴 .NET Core package enforcement | All .NET code generation |
| `ai-core-packages-go` | 🔴 Go Core package enforcement | All Go code generation |
| `ai-unit-testing` | 🔴 Unit tests with 80% coverage | Phase 3 testing |
| `ai-scaffold-service-dotnet` | Full .NET microservice scaffolding | Backend service structure |
| `ai-scaffold-service-go` | Full Go microservice scaffolding | Alternative Go backend |
| `ai-logging-patterns` | Structured JSON logging | All logging implementation |
| `ai-error-handling` | Error codes and ServiceError | All error handling |
| `ai-infrastructure-clients` | Redis, Kafka, MongoDB, SQL, ScyllaDB | All data access |
| `ai-sli-middleware` | SLI tracking middleware | Observability endpoints |
| `ai-helm-charts` | Kubernetes Helm patterns | Production deployment |
| `ai-docker-images` | Official Docker images only | Dockerfile & docker-compose |
| `ai-react-ui` | Dark Tech Theme React patterns | Frontend UI generation |

---

## ⚠️ MANDATORY: Follow the 13-Phase Development Workflow

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        USE THE ai-development-workflow SKILL                            ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   You MUST follow ALL 13 phases in order:                                                ║
║                                                                                          ║
║   PHASE 0:  Architecture Analysis (MANDATORY FIRST STEP!)                                ║
║   ─────────────────────────────────────────────────────────────                          ║
║   PHASE 1:  Generate Code & Seed Files                                                   ║
║   PHASE 2:  Build Locally & Fix Errors                                                   ║
║   PHASE 2.5: Code Quality (Format & Lint) ← dotnet format, golangci-lint                 ║
║   PHASE 3:  Run Unit Tests (80% coverage) ← ai-unit-testing skill                       ║
║   PHASE 4:  Build Docker Image                                                           ║
║   PHASE 5:  Deploy Docker Compose (Infrastructure + Application)                         ║
║   PHASE 6:  Seed ALL Databases (SQL, MongoDB, ScyllaDB, Redis, Kafka)                    ║
║   ─────────────────────────────────────────────────────────────                          ║
║   PHASE 7:  Integration Tests (API) ← MUST PASS!                                         ║
║   PHASE 8:  Integration Tests (UI)  ← MUST PASS!                                         ║
║   ─────────────────────────────────────────────────────────────                          ║
║   PHASE 9:  Create Helm Chart (ONLY after tests pass!)                                   ║
║   PHASE 10: Final Status & Delivery                                                      ║
║                                                                                          ║
║   ❌ DO NOT SKIP ANY PHASE!                                                              ║
║   ❌ DO NOT CREATE HELM CHART UNTIL API + UI TESTS PASS!                                 ║
║   ❌ DO NOT DECLARE SUCCESS UNTIL ALL PHASES PASS!                                       ║
║   ❌ FIX ALL ERRORS BEFORE PROCEEDING TO NEXT PHASE!                                     ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## System Overview

Generate a complete **Order Management System** with the following requirements:

### Service Name
- **Service:** `order-management-service`
- **Namespace:** `AI.OrderManagement`
- **Port:** `8080`

---

## Technology Stack

### Backend (.NET 8)
Use the AI Core packages for ALL cross-cutting concerns:
- `Core.Config` - Configuration management
- `Core.Logger` - Structured logging with correlation IDs
- `Core.Errors` - Error codes and ServiceError
- `Core.Metrics` - Prometheus metrics
- `Core.Sli` - SLI tracking middleware
- `Core.Infrastructure` - All database clients
- `Core.Reliability` - Circuit breakers and retry policies

### Frontend (React 18)
Use the AI React patterns:
- React 18 + TypeScript 5 + Vite 5
- Tailwind CSS with Dark Tech Theme
- Zustand 5 for state management
- Recharts for analytics
- Lucide React for icons
- Axios for API calls

---

## Data Architecture

### SQL Server (Relational Data)
Primary transactional store for orders and customers.

```
Tables:
- Orders (OrderId, CustomerId, Status, TotalAmount, CreatedAt, UpdatedAt)
- OrderItems (OrderItemId, OrderId, ProductId, Quantity, UnitPrice)
- Customers (CustomerId, Name, Email, Phone, Address)
- Products (ProductId, Name, Description, Price, StockQuantity)
```

### MongoDB (Document Store)
Store order history snapshots and audit logs.

```
Collections:
- order_snapshots (full order state at each status change)
- audit_logs (who changed what, when)
- customer_preferences (flexible schema for preferences)
```

### ScyllaDB (Time-Series Analytics)
Store order metrics and analytics data.

```
Tables:
- order_metrics_by_day (partition by date, cluster by hour)
- product_sales_by_region (partition by region, cluster by product)
- customer_activity (partition by customer_id, cluster by timestamp)
```

### Redis (Caching & Real-time)
- Cache frequently accessed orders and products
- Store shopping cart sessions
- Real-time inventory counts
- Rate limiting counters

### Apache Kafka (Event Streaming)
Topics for event-driven architecture:

```
Topics:
- order.created (new orders)
- order.status.changed (status updates)
- inventory.updated (stock changes)
- payment.processed (payment events)
- notification.send (email/SMS triggers)
```

---

## API Endpoints

### Orders API
```
POST   /api/v1/orders              - Create new order
GET    /api/v1/orders/{id}         - Get order by ID
GET    /api/v1/orders              - List orders (paginated)
PUT    /api/v1/orders/{id}/status  - Update order status
DELETE /api/v1/orders/{id}         - Cancel order
```

### Products API
```
GET    /api/v1/products            - List products
GET    /api/v1/products/{id}       - Get product details
POST   /api/v1/products            - Create product (admin)
PUT    /api/v1/products/{id}       - Update product (admin)
```

### Customers API
```
GET    /api/v1/customers/{id}      - Get customer
POST   /api/v1/customers           - Create customer
PUT    /api/v1/customers/{id}      - Update customer
GET    /api/v1/customers/{id}/orders - Get customer orders
```

### Analytics API
```
GET    /api/v1/analytics/sales     - Sales dashboard data
GET    /api/v1/analytics/products  - Product performance
GET    /api/v1/analytics/customers - Customer insights
```

### SRE Endpoints
```
GET    /health                     - Health check
GET    /ready                      - Readiness probe
GET    /api/v1/sli                 - SLI metrics
GET    /metrics                    - Prometheus metrics
```

---

## Error Codes

Define these error codes in ErrorRegistry:

| Code | Name | Message |
|------|------|---------|
| ORD001 | OrderNotFound | Order with ID {0} was not found |
| ORD002 | InvalidOrderStatus | Cannot transition from {0} to {1} |
| ORD003 | InsufficientInventory | Product {0} has insufficient stock |
| ORD004 | OrderAlreadyCancelled | Order {0} is already cancelled |
| CUS001 | CustomerNotFound | Customer with ID {0} was not found |
| CUS002 | DuplicateEmail | Customer with email {0} already exists |
| PRD001 | ProductNotFound | Product with ID {0} was not found |
| PRD002 | InvalidPrice | Price must be greater than zero |
| PAY001 | PaymentFailed | Payment processing failed: {0} |
| INV001 | InventoryLocked | Inventory is locked for product {0} |

---

## SLI Requirements

Track these SLIs:

### Availability
- Target: 99.9%
- Measure: Successful requests / Total requests

### Latency
- P50: < 100ms
- P95: < 500ms
- P99: < 1000ms

### Throughput
- Track requests per second by endpoint
- Track orders created per minute

### Error Budget
- Monthly error budget: 0.1%
- Alert at 50% consumption

---

## Logging Requirements

All logs must include:
- `correlation_id` - Trace requests across services
- `service_name` - "order-management-service"
- `environment` - dev/staging/prod
- `timestamp` - ISO 8601 format

Log levels:
- **DEBUG**: Detailed flow, SQL queries (dev only)
- **INFO**: Business events (order created, status changed)
- **WARN**: Recoverable issues (cache miss, retry)
- **ERROR**: Failures requiring attention
- **FATAL**: Service cannot continue

---

## UI Requirements

### Theme
Use Dark Tech Theme:
- Background: `#0a0a0f` (primary), `#12121a` (secondary)
- Accent: Cyan (`#00d4ff`)
- Text: White primary, gray-400 secondary
- Borders: `rgba(0, 212, 255, 0.3)`

### Pages

#### Dashboard (`/`)
- Order statistics cards (total, pending, completed, revenue)
- Real-time order chart (Recharts)
- Recent orders table
- Inventory alerts

#### Orders (`/orders`)
- Filterable order list with DataTable
- Status badges with colors
- Quick actions (view, cancel)
- Order detail modal

#### Products (`/products`)
- Product grid with cards
- Stock level indicators
- Quick edit inline
- Add product modal

#### Customers (`/customers`)
- Customer list with search
- Customer detail with order history
- Activity timeline

#### Analytics (`/analytics`)
- Sales over time chart
- Top products chart
- Customer acquisition funnel
- Regional heat map

### Components Required
- `Button` - Primary/secondary/danger variants
- `Card` - With glow effect on hover
- `DataTable` - Sortable, filterable, paginated
- `Modal` - Slide-in with backdrop blur
- `StatusBadge` - Color-coded status pills
- `Chart` - Line, bar, pie using Recharts
- `Sidebar` - Collapsible navigation
- `Header` - With user menu and notifications

### State Management (Zustand)
```typescript
// Required stores:
- useOrderStore (orders, selectedOrder, filters)
- useProductStore (products, categories)
- useCustomerStore (customers, selectedCustomer)
- useUIStore (sidebar, modals, theme)
- useAuthStore (user, token, permissions)
```

---

## Infrastructure

### Docker Compose (Development)
Include all services:
- SQL Server 2022 (mcr.microsoft.com/mssql/server:2022-latest)
- MongoDB 7 (mongo:7)
- ScyllaDB (scylladb/scylla:latest)
- Redis 7 (redis:7-alpine)
- Kafka KRaft (apache/kafka:latest) - NO ZOOKEEPER
- Order Management Service

### Helm Chart (Production)
Self-contained chart with:
- Deployment with health probes
- Service (ClusterIP)
- ConfigMap for configuration
- Secret for credentials
- HorizontalPodAutoscaler
- PodDisruptionBudget
- ServiceMonitor for Prometheus

---

## Project Structure

### Backend (.NET)
```
src/
├── AI.OrderManagement.Api/
│   ├── Program.cs
│   ├── appsettings.json
│   ├── Controllers/
│   │   ├── OrdersController.cs
│   │   ├── ProductsController.cs
│   │   ├── CustomersController.cs
│   │   └── AnalyticsController.cs
│   ├── Middleware/
│   │   └── (use Core.Sli middleware)
│   └── Models/
│       ├── Requests/
│       └── Responses/
├── AI.OrderManagement.Domain/
│   ├── Entities/
│   ├── Enums/
│   └── Events/
├── AI.OrderManagement.Infrastructure/
│   ├── Repositories/
│   ├── Kafka/
│   └── Cache/
└── AI.OrderManagement.Tests/
```

### Frontend (React)
```
src/
├── components/
│   ├── ui/
│   │   ├── Button.tsx
│   │   ├── Card.tsx
│   │   ├── DataTable.tsx
│   │   └── Modal.tsx
│   ├── layout/
│   │   ├── Sidebar.tsx
│   │   └── Header.tsx
│   └── features/
│       ├── orders/
│       ├── products/
│       └── customers/
├── pages/
│   ├── Dashboard.tsx
│   ├── Orders.tsx
│   ├── Products.tsx
│   ├── Customers.tsx
│   └── Analytics.tsx
├── stores/
│   ├── orderStore.ts
│   ├── productStore.ts
│   └── uiStore.ts
├── services/
│   └── api.ts
├── hooks/
│   ├── useOrders.ts
│   └── useProducts.ts
└── types/
    └── index.ts
```

---

## Package Authentication

### GitHub Personal Access Token
Use the following PAT to authenticate with GitHub Packages for Core libraries:

```
your-github-token
```

### .NET NuGet Restore
Add a `nuget.config` in the solution root:

```xml
<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" />
    <add key="github" value="https://nuget.pkg.github.com/your-github-org/index.json" />
  </packageSources>
  <packageSourceCredentials>
    <github>
      <add key="Username" value="your-github-org" />
      <add key="ClearTextPassword" value="your-github-token" />
    </github>
  </packageSourceCredentials>
</configuration>
```

### Go Package Download
Configure Git to use the PAT for private Go modules:

```bash
# Set GOPRIVATE to skip proxy for org packages
export GOPRIVATE=github.com/your-github-org/*

# Configure git to use PAT for authentication
git config --global url."https://your-github-token@github.com/".insteadOf "https://github.com/"
```

Or create/update `~/.netrc` (Linux/Mac) or `%USERPROFILE%\_netrc` (Windows):

```
machine github.com
login your-github-org
password your-github-token
```

---

## Generation Instructions (Follow 13-Phase Workflow!)

### Skills to Load Before Starting

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  LOAD THESE SKILLS FROM .github/skills/ BEFORE GENERATING ANY CODE:         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1.  ai-development-workflow   → Governs the entire 13-phase process       │
│  2.  ai-core-packages-dotnet   → Enforces Core.* package usage             │
│  3.  ai-scaffold-service-dotnet→ Project structure and templates           │
│  4.  ai-logging-patterns       → ServiceLogger patterns                    │
│  5.  ai-error-handling         → ServiceError and error codes              │
│  6.  ai-infrastructure-clients → SQL, MongoDB, ScyllaDB, Redis, Kafka      │
│  7.  ai-sli-middleware         → SLI tracking and /api/v1/sli endpoint     │
│  8.  ai-unit-testing           → Unit tests with 80% coverage (Phase 3)    │
│  9.  ai-docker-images          → Official images, NO Bitnami               │
│  10. ai-helm-charts            → Self-contained Kubernetes charts          │
│  11. ai-react-ui               → Dark Tech Theme, Vite + Tailwind          │
│                                                                             │
│  (ai-core-packages-go and ai-scaffold-service-go for Go alternative)      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### PHASE 0: Architecture Analysis (MANDATORY FIRST!)
```
Follow the ai-development-workflow skill to:
1. Analyze use case requirements
2. Identify data entities and relationships
3. Determine data platform allocation (SQL, MongoDB, ScyllaDB, Redis, Kafka)
4. Create ARCHITECTURE.md documenting decisions
5. Get user approval before proceeding
```

### PHASE 1: Generate Code & Seed Files

**Use these skills for code generation:**

| Step | Skill | Action |
|------|-------|--------|
| 1 | `ai-scaffold-service-dotnet` | Generate .NET backend project structure |
| 2 | `ai-core-packages-dotnet` | Enforce Core.* packages - NEVER raw packages |
| 3 | `ai-infrastructure-clients` | Configure SQL Server, MongoDB, ScyllaDB, Redis, Kafka |
| 4 | `ai-error-handling` | Implement error codes (ORD001, CUS001, etc.) |
| 5 | `ai-logging-patterns` | Add structured logging with correlation IDs |
| 6 | `ai-sli-middleware` | Add SLI middleware to all endpoints |
| 7 | `ai-docker-images` | Create Dockerfile and docker-compose.yml |
| 8 | `ai-react-ui` | Build React frontend with Dark Tech Theme |

### PHASE 2: Build Locally
```powershell
dotnet restore
dotnet build
# Fix ALL errors before proceeding!
```

### PHASE 2.5: Code Quality (Format & Lint)
```powershell
# .NET: Format and lint
dotnet format --verbosity normal
dotnet build /p:EnforceCodeStyleInBuild=true

# React: Format and lint
cd ui
npx prettier --write "src/**/*.{ts,tsx,css,json}"
npx eslint src/ --fix
npx tsc --noEmit
# Fix ALL lint errors before proceeding!
```

### PHASE 3: Run Unit Tests (80% Coverage Required)
```powershell
# Use ai-unit-testing skill
dotnet test --collect:"XPlat Code Coverage" --results-directory ./coverage
# All tests must pass with ≥80% coverage!
```

### PHASE 4: Build Docker Image
```powershell
docker build -t order-management-service:latest --build-arg GITHUB_TOKEN=$env:GITHUB_TOKEN .
```

### PHASE 5: Deploy Docker Compose (Infrastructure + Application)
```powershell
docker-compose up -d
docker-compose ps           # Verify all containers are running
docker-compose logs -f order-management-service
# Verify all containers are healthy!
```

### PHASE 6: Seed ALL Databases
```powershell
# Run seed scripts for all platforms:
# SQL Server: docker exec -it {container} /opt/mssql-tools18/bin/sqlcmd ...
# MongoDB: docker exec -it {container} mongosh < seed-mongodb.js
# ScyllaDB: docker exec -it {container} cqlsh -f seed-scylladb.cql
# Kafka: Create topics with kafka-topics.sh
# Redis: Verify connection with redis-cli ping
```

### PHASE 7: Integration Tests (API) - MUST PASS!
```powershell
# Test all endpoints:
# Test: /health, /health/live, /health/ready, /metrics, /api/v1/sli
# Test: CRUD operations for all entities
# ALL API tests must pass before proceeding!

$baseUrl = "http://localhost:8080"
Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
Invoke-RestMethod -Uri "$baseUrl/api/v1/orders" -Method GET
# ... test all CRUD operations
```

### PHASE 8: Integration Tests (UI) - MUST PASS!
```powershell
cd ui
npm install
npm run build    # Must succeed
npx tsc --noEmit # Must pass
npm run dev      # Manual verification
# ALL UI tests must pass before proceeding!
```

### PHASE 9: Create Helm Chart (Only After Tests Pass!)
**Skill:** `ai-helm-charts`

```
╔════════════════════════════════════════════════════════════════╗
║  DO NOT CREATE HELM CHART UNLESS:                              ║
║  ✅ Phase 7 (API Tests) PASSED                                 ║
║  ✅ Phase 8 (UI Tests) PASSED                                  ║
╚════════════════════════════════════════════════════════════════╝
```

```powershell
# Generate Helm chart using ai-helm-charts patterns:
# - Self-contained (NO Bitnami subcharts)
# - Official Docker images from ai-docker-images
# - ConfigMaps, Secrets, HPA, PDB included
# - Health probes configured
```

### PHASE 10: Final Status
Only after ALL phases pass, provide the final status report and launch instructions.

---

## Critical Reminders

### Skill-Enforced Rules

| Skill | Forbidden ❌ | Required ✅ |
|-------|-------------|-------------|
| `ai-core-packages-dotnet` | Serilog, Polly, StackExchange.Redis, Confluent.Kafka, MongoDB.Driver | Core.Logger, Core.Reliability, Core.Infrastructure |
| `ai-core-packages-go` | logrus, zerolog, go-redis, sarama, mongo-driver | core/go/logger, core/go/infrastructure/* |
| `ai-docker-images` | Bitnami images, Confluent images, Zookeeper | Official images: mongo:7, redis:7-alpine, apache/kafka |
| `ai-helm-charts` | Bitnami subcharts, Helm dependencies | Self-contained charts with official images |
| `ai-logging-patterns` | Console.WriteLine, fmt.Println, raw ILogger | ServiceLogger with correlation IDs |
| `ai-error-handling` | Generic Exception, fmt.Errorf | ServiceError with error codes |
| `ai-react-ui` | Redux, styled-components, Material UI | Zustand, Tailwind CSS, Dark Tech Theme |
| `ai-unit-testing` | Skip unit tests, <80% coverage | All unit tests pass with ≥80% coverage |

### DO NOT USE:
- ❌ `Serilog` - Use `Core.Logger` (see `ai-logging-patterns`)
- ❌ `Polly` directly - Use `Core.Reliability` (see `ai-core-packages-dotnet`)
- ❌ `StackExchange.Redis` directly - Use `Core.Infrastructure` (see `ai-infrastructure-clients`)
- ❌ `Confluent.Kafka` directly - Use `Core.Infrastructure` (see `ai-infrastructure-clients`)
- ❌ Bitnami Docker images - Use official images (see `ai-docker-images`)
- ❌ Zookeeper - Use Kafka KRaft mode (see `ai-docker-images`)
- ❌ Helm dependencies - Self-contained charts only (see `ai-helm-charts`)
- ❌ Redux/Material UI - Use Zustand/Tailwind (see `ai-react-ui`)
- ❌ Skip unit tests - Must have 80% coverage (see `ai-unit-testing`)
- ❌ Skip integration tests - Must pass before Helm

### ALWAYS USE:
- ✅ `ai-development-workflow` skill for 13-phase process governance
- ✅ `Core.*` packages for all cross-cutting concerns (see `ai-core-packages-dotnet`)
- ✅ `ServiceLogger` for all logging (see `ai-logging-patterns`)
- ✅ `ServiceError` with error codes (see `ai-error-handling`)
- ✅ SLI middleware on all endpoints (see `ai-sli-middleware`)
- ✅ Correlation IDs in all requests (see `ai-logging-patterns`)
- ✅ `ai-unit-testing` skill for Phase 3 with 80% coverage
- ✅ Official Docker images only (see `ai-docker-images`)
- ✅ Dark Tech Theme for UI (see `ai-react-ui`)
- ✅ 13-phase workflow completion (see `ai-development-workflow`)
