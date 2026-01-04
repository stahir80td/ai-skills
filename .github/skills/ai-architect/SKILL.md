````markdown
---
name: ai-architect
description: >
  MANDATORY architecture analysis skill for ALL new services. Evaluates use-cases,
  asks clarifying questions, and determines optimal data platform allocation before
  ANY code generation. Produces ARCHITECTURE.md with data flow diagrams and platform
  decisions. MUST run BEFORE scaffolding code. Use when user says "create service",
  "implement system", "build application", or provides system requirements.
---

# AI Architect Skill

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  MANDATORY ARCHITECTURE PHASE  ⚠️                               ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   STOP! Before writing ANY code, you MUST:                                               ║
║                                                                                          ║
║   1. UNDERSTAND the use-case (ask clarifying questions)                                  ║
║   2. DECIDE which data goes in which platform                                            ║
║   3. DOCUMENT the architecture in ARCHITECTURE.md                                        ║
║   4. GET USER CONFIRMATION before proceeding to code                                     ║
║                                                                                          ║
║   ❌ DO NOT generate code until architecture is approved!                                ║
║   ❌ DO NOT assume data platforms - ASK if unclear!                                      ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## Phase 0: Architecture Analysis (BEFORE Code Generation)

### Step 1: Gather Requirements

When a user provides system requirements, ASK these questions if not already answered:

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                     ARCHITECTURE DISCOVERY QUESTIONS                                     ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   DATA CHARACTERISTICS:                                                                  ║
║   □ What entities need ACID transactions? (→ SQL Server)                                 ║
║   □ What data has complex relationships/foreign keys? (→ SQL Server)                     ║
║   □ What data is hierarchical/nested/flexible schema? (→ MongoDB)                        ║
║   □ What data is time-series (sensor readings, metrics)? (→ ScyllaDB)                    ║
║   □ What events need to be immutably logged? (→ ScyllaDB)                                ║
║   □ What data needs sub-millisecond access? (→ Redis)                                    ║
║   □ What events need to be streamed to multiple consumers? (→ Kafka)                     ║
║                                                                                          ║
║   SCALE & PERFORMANCE:                                                                   ║
║   □ Expected read/write ratio?                                                           ║
║   □ Expected data volume per day?                                                        ║
║   □ Latency requirements (P99)?                                                          ║
║   □ Data retention requirements?                                                         ║
║                                                                                          ║
║   INTEGRATION:                                                                           ║
║   □ What external systems need to be notified of changes?                                ║
║   □ Are there downstream consumers that need real-time events?                           ║
║   □ Is there a need for event sourcing/audit trail?                                      ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### Step 2: Apply Data Platform Decision Matrix

Use this matrix to decide WHERE each data type belongs:

```
╔═══════════════════════════════════════════════════════════════════════════════════════════════════════╗
║                              DATA PLATFORM DECISION MATRIX                                            ║
╠═══════════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                       ║
║   DATA TYPE                        │ PRIMARY PLATFORM  │ SECONDARY (Cache)  │ EVENT STREAM           ║
║   ─────────────────────────────────┼───────────────────┼────────────────────┼────────────────────────║
║   User accounts, roles, auth       │ SQL Server        │ Redis (sessions)   │ Kafka (user.events)    ║
║   Orders, transactions             │ SQL Server        │ Redis (hot orders) │ Kafka (order.events)   ║
║   Products, inventory              │ SQL Server        │ Redis (catalog)    │ Kafka (inventory.*)    ║
║   Customers, contacts              │ SQL Server        │ Redis (lookup)     │ Kafka (customer.*)     ║
║   Payments, invoices               │ SQL Server        │ -                  │ Kafka (payment.*)      ║
║   ─────────────────────────────────┼───────────────────┼────────────────────┼────────────────────────║
║   Device profiles, configs         │ MongoDB           │ Redis (active)     │ Kafka (device.config)  ║
║   User preferences, settings       │ MongoDB           │ Redis (session)    │ -                      ║
║   Content, documents, files meta   │ MongoDB           │ Redis (hot docs)   │ -                      ║
║   Automation rules, workflows      │ MongoDB           │ -                  │ Kafka (rule.*)         ║
║   Feature flags, A/B configs       │ MongoDB           │ Redis (flags)      │ -                      ║
║   ─────────────────────────────────┼───────────────────┼────────────────────┼────────────────────────║
║   Sensor telemetry, metrics        │ ScyllaDB          │ Redis (latest)     │ Kafka (telemetry.*)    ║
║   Time-series events               │ ScyllaDB          │ -                  │ Kafka (events.*)       ║
║   Audit logs, compliance           │ ScyllaDB          │ -                  │ Kafka (audit.*)        ║
║   IoT device readings              │ ScyllaDB          │ Redis (latest)     │ Kafka (iot.*)          ║
║   Analytics aggregates             │ ScyllaDB          │ Redis (dashboard)  │ -                      ║
║   ─────────────────────────────────┼───────────────────┼────────────────────┼────────────────────────║
║   Session state                    │ Redis             │ -                  │ -                      ║
║   Rate limiting                    │ Redis             │ -                  │ -                      ║
║   Real-time counters               │ Redis             │ -                  │ -                      ║
║   Pub/Sub notifications            │ Redis             │ -                  │ -                      ║
║   Distributed locks                │ Redis             │ -                  │ -                      ║
║                                                                                                       ║
╚═══════════════════════════════════════════════════════════════════════════════════════════════════════╝
```

### Step 3: Define Kafka Topics

For each entity/event that needs streaming, define topics:

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                              KAFKA TOPIC NAMING CONVENTION                               ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   Pattern: {service-name}.{entity}.{action}                                              ║
║                                                                                          ║
║   DOMAIN EVENTS (State Changes):                                                         ║
║   ├── {service}.{entity}.created      # New entity created                               ║
║   ├── {service}.{entity}.updated      # Entity modified                                  ║
║   ├── {service}.{entity}.deleted      # Entity removed                                   ║
║   └── {service}.{entity}.{custom}     # Domain-specific events                           ║
║                                                                                          ║
║   COMMANDS (CQRS Pattern):                                                               ║
║   └── {service}.commands.{action}     # Request to perform action                        ║
║                                                                                          ║
║   NOTIFICATIONS:                                                                         ║
║   ├── {service}.notifications.email   # Email notifications                              ║
║   ├── {service}.notifications.push    # Push notifications                               ║
║   └── {service}.notifications.webhook # Webhook callbacks                                ║
║                                                                                          ║
║   INFRASTRUCTURE:                                                                        ║
║   └── {service}.dlq                   # Dead Letter Queue                                ║
║                                                                                          ║
║   Examples:                                                                              ║
║   • order-service.orders.created                                                         ║
║   • order-service.orders.status-changed                                                  ║
║   • order-service.payments.completed                                                     ║
║   • device-service.telemetry.readings                                                    ║
║   • user-service.users.registered                                                        ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### Step 4: Generate ARCHITECTURE.md

BEFORE generating any code, create this document:

```markdown
# {Service Name} - Architecture Document

## Overview

{Brief description of the service and its purpose}

## Data Platform Allocation

### SQL Server (Azure SQL MI)
**Purpose:** Transactional data with ACID requirements

| Entity | Table | Relationships | Indexes |
|--------|-------|---------------|---------|
| {Entity1} | {table_name} | FK to {other} | IX_{field} |
| {Entity2} | {table_name} | FK to {other} | IX_{field} |

**Rationale:** {Why SQL Server for this data}

### MongoDB
**Purpose:** Document/hierarchical data with flexible schema

| Collection | Document Type | Indexes |
|------------|---------------|---------|
| {collection} | {type} | {indexes} |

**Rationale:** {Why MongoDB for this data}

### ScyllaDB
**Purpose:** Time-series data, event store, high-throughput writes

| Table | Partition Key | Clustering Key | TTL |
|-------|---------------|----------------|-----|
| {table} | {partition} | {clustering} | {days} |

**Rationale:** {Why ScyllaDB for this data}

### Redis
**Purpose:** Caching, sessions, real-time data

| Key Pattern | Data Structure | TTL | Purpose |
|-------------|----------------|-----|---------|
| {pattern} | {type} | {ttl} | {purpose} |

**Rationale:** {Why Redis for this data}

### Kafka Topics
**Purpose:** Event streaming, async communication

| Topic | Partitions | Retention | Consumers |
|-------|------------|-----------|-----------|
| {topic} | {n} | {hours} | {services} |

**Rationale:** {Why these events need streaming}

## Data Flow Diagram

```
[Client] → [API Gateway] → [Service]
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         [SQL Server]    [MongoDB]       [Redis Cache]
              │               │               │
              └───────────────┼───────────────┘
                              ▼
                          [Kafka]
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         [Consumer1]    [Consumer2]     [ScyllaDB]
```

## Entity Relationship

{ERD or entity descriptions}

## API Endpoints

| Method | Endpoint | Description | Data Source |
|--------|----------|-------------|-------------|
| GET | /api/v1/{entities} | List all | SQL Server + Redis |
| POST | /api/v1/{entities} | Create | SQL Server → Kafka |
| GET | /api/v1/{entities}/{id} | Get one | Redis → SQL Server |
| PUT | /api/v1/{entities}/{id} | Update | SQL Server → Kafka |
| DELETE | /api/v1/{entities}/{id} | Delete | SQL Server → Kafka |

## Event Contracts

### {service}.{entity}.created
```json
{
  "eventId": "uuid",
  "eventType": "{entity}.created",
  "timestamp": "ISO8601",
  "data": {
    "id": "uuid",
    // entity fields
  },
  "metadata": {
    "correlationId": "uuid",
    "userId": "uuid"
  }
}
```

## Caching Strategy

| Data | Cache Key | TTL | Invalidation |
|------|-----------|-----|--------------|
| {data} | {pattern} | {ttl} | {strategy} |

## Estimated Scale

| Metric | Expected Value |
|--------|----------------|
| Requests/second | {n} |
| Data volume/day | {size} |
| Active users | {n} |
| Retention period | {days} |

---

**Architecture Approved:** [ ] Yes / [ ] No
**Date:** {date}
**Architect:** Copilot + User
```

---

## Platform Selection Guidelines

### When to Use SQL Server (Azure SQL MI)

```
✅ USE SQL SERVER WHEN:
├── Data requires ACID transactions
├── Complex relationships with foreign keys
├── Need for complex JOINs across tables
├── Referential integrity is critical
├── Financial/payment data
├── User accounts and authentication
├── Order management with line items
├── Inventory with stock transactions
└── Reporting with complex aggregations

❌ AVOID SQL SERVER WHEN:
├── Schema changes frequently
├── Data is deeply nested/hierarchical
├── Write volume exceeds 10K/sec
├── No relationships between records
└── Time-series data with TTL needs
```

### When to Use MongoDB

```
✅ USE MONGODB WHEN:
├── Flexible/evolving schema needed
├── Data is hierarchical/nested
├── Document-oriented data (JSON)
├── Device profiles with varying attributes
├── User preferences and settings
├── Content management (articles, posts)
├── Configuration and feature flags
├── Catalog data with attributes
└── Workflow/rule definitions

❌ AVOID MONGODB WHEN:
├── Need ACID transactions across documents
├── Heavy JOIN operations required
├── Data is purely relational
├── Financial ledger requiring consistency
└── Simple key-value lookups (use Redis)
```

### When to Use ScyllaDB

```
✅ USE SCYLLADB WHEN:
├── Time-series data (sensor readings)
├── High-throughput writes (>10K/sec)
├── Event sourcing / audit logs
├── IoT telemetry data
├── Analytics aggregates
├── Log data with TTL
├── Immutable event store
├── Wide-column data model fits
└── Need horizontal scaling

❌ AVOID SCYLLADB WHEN:
├── Need complex transactions
├── Frequent schema changes
├── Ad-hoc queries on any column
├── Small dataset (<1M rows)
└── Need for JOINs
```

### When to Use Redis

```
✅ USE REDIS WHEN:
├── Sub-millisecond latency required
├── Session state management
├── Caching frequently accessed data
├── Rate limiting / throttling
├── Real-time counters / leaderboards
├── Pub/Sub for real-time notifications
├── Distributed locks
├── Latest values (device readings)
└── Temporary data with TTL

❌ AVOID REDIS WHEN:
├── Data must survive restart (use DB)
├── Complex queries needed
├── Data relationships exist
├── Large objects (>1MB)
└── Need for transactions
```

### When to Use Kafka

```
✅ USE KAFKA WHEN:
├── Event-driven architecture
├── Multiple consumers need same events
├── Async processing required
├── Decoupling services
├── Event sourcing pattern
├── High-throughput streaming
├── Audit trail of all changes
├── Real-time analytics pipeline
└── Replay capability needed

❌ AVOID KAFKA WHEN:
├── Simple request-response needed
├── Low-volume (<100 msgs/sec)
├── No downstream consumers
├── Synchronous processing required
└── Point-to-point only (use queue)
```

---

## Example Architecture Analysis

### User Prompt:
> "Create an Order Management System with customers, products, orders, and inventory tracking"

### Architecture Decision:

```
ENTITY ANALYSIS:
────────────────────────────────────────────────────────────────

1. CUSTOMERS
   - Has relationships (orders)
   - ACID needed for account balance
   - Relatively static data
   → PRIMARY: SQL Server
   → CACHE: Redis (hot customer data)
   → EVENTS: Kafka (customer.created, customer.updated)

2. PRODUCTS
   - Has relationships (order items, inventory)
   - Catalog with attributes
   - Moderate update frequency
   → PRIMARY: SQL Server
   → CACHE: Redis (product catalog)
   → EVENTS: Kafka (product.created, product.updated, product.price-changed)

3. ORDERS
   - Complex relationships (customer, items, products)
   - ACID transactions critical
   - Status changes are events
   → PRIMARY: SQL Server
   → CACHE: Redis (recent orders, order status)
   → EVENTS: Kafka (order.created, order.status-changed, order.completed)

4. ORDER ITEMS
   - Junction table (order ↔ product)
   - Part of order transaction
   → PRIMARY: SQL Server (same transaction as order)
   → NO separate cache (loaded with order)
   → NO separate events (part of order events)

5. INVENTORY
   - High-frequency updates (stock changes)
   - Needs consistency for stock levels
   - History tracking needed
   → PRIMARY: SQL Server (current stock)
   → CACHE: Redis (available quantity)
   → EVENTS: Kafka (inventory.adjusted, inventory.low-stock)
   → HISTORY: ScyllaDB (stock movement history - time-series)

6. AUDIT LOG
   - Immutable event log
   - High-volume writes
   - Time-based queries
   → PRIMARY: ScyllaDB
   → EVENTS: All Kafka events → ScyllaDB consumer
```

### Resulting Kafka Topics:

```
order-management.customers.created
order-management.customers.updated
order-management.products.created
order-management.products.updated
order-management.products.price-changed
order-management.orders.created
order-management.orders.status-changed
order-management.orders.completed
order-management.orders.cancelled
order-management.inventory.adjusted
order-management.inventory.low-stock
order-management.notifications.email
order-management.dlq
```

---

## Confirmation Before Code Generation

After completing architecture analysis, ALWAYS ask:

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ARCHITECTURE CONFIRMATION                                         ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   I've analyzed your requirements and propose the following architecture:                ║
║                                                                                          ║
║   📊 SQL Server:                                                                         ║
║      • {entities}                                                                        ║
║                                                                                          ║
║   📁 MongoDB:                                                                            ║
║      • {collections}                                                                     ║
║                                                                                          ║
║   ⏱️ ScyllaDB:                                                                           ║
║      • {tables}                                                                          ║
║                                                                                          ║
║   🔴 Redis:                                                                              ║
║      • {cache patterns}                                                                  ║
║                                                                                          ║
║   📨 Kafka Topics:                                                                       ║
║      • {topics}                                                                          ║
║                                                                                          ║
║   Do you want me to:                                                                     ║
║   1. Proceed with this architecture and generate code?                                   ║
║   2. Modify the platform allocation?                                                     ║
║   3. Add/remove data platforms?                                                          ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

Only after user confirms, proceed to:
1. Generate ARCHITECTURE.md
2. Start code generation (Phase 1 of ai-development-workflow)

---

## Integration with Development Workflow

This skill runs as **Phase 0** BEFORE the 10-phase development workflow:

```
COMPLETE SERVICE GENERATION FLOW:
─────────────────────────────────────────────────────────

PHASE 0: Architecture (THIS SKILL)
    ├── Gather requirements
    ├── Ask clarifying questions
    ├── Analyze data characteristics
    ├── Allocate data to platforms
    ├── Define Kafka topics
    ├── Generate ARCHITECTURE.md
    └── Get user confirmation
           │
           ▼
PHASE 1-10: Development Workflow (ai-development-workflow)
    ├── Phase 1: Generate Code & Seed Files
    ├── Phase 2: Build Locally
    ├── Phase 3: Run Unit Tests
    ├── Phase 4: Build Docker Image
    ├── Phase 5: Start Containers
    ├── Phase 6: Seed ALL Databases
    ├── Phase 7: Test API Endpoints
    ├── Phase 8: Test Frontend
    ├── Phase 9: Create Helm Chart
    └── Phase 10: Final Status
```

````
