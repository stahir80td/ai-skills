````markdown
---
name: ai-data-seeding
description: >
  MANDATORY seed data generation for ALL data platforms used by AI services.
  Use when scaffolding services that require SQL Server, MongoDB, ScyllaDB, Kafka, or Redis.
  Generate seed scripts for EVERY data platform specified in system requirements.
  Includes schema creation, sample data, indexes, and initialization scripts.
---

# AI Data Platform Seeding

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  MANDATORY DATA SEEDING  ⚠️                                     ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   When a service uses ANY data platform, you MUST generate seed scripts for it!          ║
║                                                                                          ║
║   ✅ SQL Server  →  scripts/seed-sqlserver.sql                                           ║
║   ✅ MongoDB     →  scripts/seed-mongodb.js                                              ║
║   ✅ ScyllaDB    →  scripts/seed-scylladb.cql                                            ║
║   ✅ Kafka       →  scripts/seed-kafka.ps1                                               ║
║   ✅ Redis       →  scripts/seed-redis.ps1                                               ║
║                                                                                          ║
║   Also generate: scripts/seed-all.ps1 (master script that runs ALL seeds)                ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## Master Seed Script (ALWAYS GENERATE)

### scripts/seed-all.ps1

```powershell
#!/usr/bin/env pwsh
# =============================================================================
# Master Seed Script - Seeds ALL data platforms
# =============================================================================

param(
    [string]$ComposeProject = "{service-name}",
    [switch]$Force
)

$ErrorActionPreference = "Stop"

Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           AI Data Platform Seeding                          ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

# Wait for containers to be healthy
function Wait-ForContainer {
    param([string]$ContainerName, [int]$TimeoutSeconds = 120)
    
    Write-Host "⏳ Waiting for $ContainerName to be healthy..." -ForegroundColor Yellow
    $elapsed = 0
    while ($elapsed -lt $TimeoutSeconds) {
        $status = docker inspect --format='{{.State.Health.Status}}' $ContainerName 2>$null
        if ($status -eq "healthy" -or $null -eq $status) {
            # If no health check, try to connect
            Write-Host "✅ $ContainerName is ready" -ForegroundColor Green
            return $true
        }
        Start-Sleep -Seconds 5
        $elapsed += 5
    }
    Write-Host "❌ Timeout waiting for $ContainerName" -ForegroundColor Red
    return $false
}

# =============================================================================
# SQL SERVER SEEDING
# =============================================================================
if (docker ps --format '{{.Names}}' | Select-String "$ComposeProject.*sqlserver") {
    Write-Host "`n📊 Seeding SQL Server..." -ForegroundColor Cyan
    Wait-ForContainer "$ComposeProject-sqlserver-1"
    
    docker exec -i "$ComposeProject-sqlserver-1" /opt/mssql-tools18/bin/sqlcmd `
        -S localhost -U sa -P "YourStrong!Password" -C `
        -i /scripts/seed-sqlserver.sql
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ SQL Server seeded successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ SQL Server seeding failed" -ForegroundColor Red
    }
}

# =============================================================================
# MONGODB SEEDING
# =============================================================================
if (docker ps --format '{{.Names}}' | Select-String "$ComposeProject.*mongodb") {
    Write-Host "`n🍃 Seeding MongoDB..." -ForegroundColor Cyan
    Wait-ForContainer "$ComposeProject-mongodb-1"
    
    docker exec -i "$ComposeProject-mongodb-1" mongosh --quiet < scripts/seed-mongodb.js
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ MongoDB seeded successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ MongoDB seeding failed" -ForegroundColor Red
    }
}

# =============================================================================
# SCYLLADB SEEDING
# =============================================================================
if (docker ps --format '{{.Names}}' | Select-String "$ComposeProject.*scylladb") {
    Write-Host "`n🔷 Seeding ScyllaDB..." -ForegroundColor Cyan
    # ScyllaDB takes longer to initialize
    Start-Sleep -Seconds 30
    Wait-ForContainer "$ComposeProject-scylladb-1" 180
    
    docker exec -i "$ComposeProject-scylladb-1" cqlsh -f /scripts/seed-scylladb.cql
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ ScyllaDB seeded successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ ScyllaDB seeding failed" -ForegroundColor Red
    }
}

# =============================================================================
# KAFKA TOPICS CREATION
# =============================================================================
if (docker ps --format '{{.Names}}' | Select-String "$ComposeProject.*kafka") {
    Write-Host "`n📨 Creating Kafka topics..." -ForegroundColor Cyan
    Wait-ForContainer "$ComposeProject-kafka-1"
    
    & "$PSScriptRoot/seed-kafka.ps1" -ComposeProject $ComposeProject
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Kafka topics created successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ Kafka topic creation failed" -ForegroundColor Red
    }
}

# =============================================================================
# REDIS INITIALIZATION
# =============================================================================
if (docker ps --format '{{.Names}}' | Select-String "$ComposeProject.*redis") {
    Write-Host "`n🔴 Initializing Redis..." -ForegroundColor Cyan
    Wait-ForContainer "$ComposeProject-redis-1"
    
    & "$PSScriptRoot/seed-redis.ps1" -ComposeProject $ComposeProject
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Redis initialized successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ Redis initialization failed" -ForegroundColor Red
    }
}

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║           ✅ All Data Platforms Seeded!                      ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Green
```

---

## SQL Server Seed Script

### scripts/seed-sqlserver.sql

```sql
-- =============================================================================
-- {ServiceName} Database Initialization Script
-- SQL Server 2022
-- =============================================================================

-- Create database if not exists
IF NOT EXISTS (SELECT * FROM sys.databases WHERE name = '{DatabaseName}')
BEGIN
    CREATE DATABASE [{DatabaseName}];
END
GO

USE [{DatabaseName}];
GO

-- =============================================================================
-- TABLE CREATION
-- =============================================================================

-- Example: Customers table
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Customers')
BEGIN
    CREATE TABLE [dbo].[Customers] (
        [Id] UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
        [Name] NVARCHAR(255) NOT NULL,
        [Email] NVARCHAR(255) NOT NULL UNIQUE,
        [Phone] NVARCHAR(50),
        [Address] NVARCHAR(500),
        [CreatedAt] DATETIME2 NOT NULL DEFAULT GETUTCDATE(),
        [UpdatedAt] DATETIME2 NOT NULL DEFAULT GETUTCDATE()
    );
    
    -- Indexes for common queries
    CREATE INDEX IX_Customers_Email ON [dbo].[Customers]([Email]);
    CREATE INDEX IX_Customers_Name ON [dbo].[Customers]([Name]);
    
    PRINT 'Created Customers table';
END
GO

-- Example: Products table
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Products')
BEGIN
    CREATE TABLE [dbo].[Products] (
        [Id] UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
        [Name] NVARCHAR(255) NOT NULL,
        [Description] NVARCHAR(1000),
        [Price] DECIMAL(18, 2) NOT NULL,
        [StockQuantity] INT NOT NULL DEFAULT 0,
        [Category] NVARCHAR(100),
        [CreatedAt] DATETIME2 NOT NULL DEFAULT GETUTCDATE(),
        [UpdatedAt] DATETIME2 NOT NULL DEFAULT GETUTCDATE()
    );
    
    CREATE INDEX IX_Products_Category ON [dbo].[Products]([Category]);
    CREATE INDEX IX_Products_Name ON [dbo].[Products]([Name]);
    
    PRINT 'Created Products table';
END
GO

-- Example: Orders table with foreign keys
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Orders')
BEGIN
    CREATE TABLE [dbo].[Orders] (
        [Id] UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
        [CustomerId] UNIQUEIDENTIFIER NOT NULL,
        [Status] NVARCHAR(50) NOT NULL DEFAULT 'Pending',
        [TotalAmount] DECIMAL(18, 2) NOT NULL DEFAULT 0,
        [CreatedAt] DATETIME2 NOT NULL DEFAULT GETUTCDATE(),
        [UpdatedAt] DATETIME2 NOT NULL DEFAULT GETUTCDATE(),
        CONSTRAINT FK_Orders_Customers FOREIGN KEY ([CustomerId]) 
            REFERENCES [dbo].[Customers]([Id])
    );
    
    CREATE INDEX IX_Orders_CustomerId ON [dbo].[Orders]([CustomerId]);
    CREATE INDEX IX_Orders_Status ON [dbo].[Orders]([Status]);
    CREATE INDEX IX_Orders_CreatedAt ON [dbo].[Orders]([CreatedAt] DESC);
    
    PRINT 'Created Orders table';
END
GO

-- Example: OrderItems junction table
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'OrderItems')
BEGIN
    CREATE TABLE [dbo].[OrderItems] (
        [Id] UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
        [OrderId] UNIQUEIDENTIFIER NOT NULL,
        [ProductId] UNIQUEIDENTIFIER NOT NULL,
        [Quantity] INT NOT NULL,
        [UnitPrice] DECIMAL(18, 2) NOT NULL,
        CONSTRAINT FK_OrderItems_Orders FOREIGN KEY ([OrderId]) 
            REFERENCES [dbo].[Orders]([Id]) ON DELETE CASCADE,
        CONSTRAINT FK_OrderItems_Products FOREIGN KEY ([ProductId]) 
            REFERENCES [dbo].[Products]([Id])
    );
    
    CREATE INDEX IX_OrderItems_OrderId ON [dbo].[OrderItems]([OrderId]);
    CREATE INDEX IX_OrderItems_ProductId ON [dbo].[OrderItems]([ProductId]);
    
    PRINT 'Created OrderItems table';
END
GO

-- =============================================================================
-- SEED DATA
-- =============================================================================

-- Only seed if tables are empty
IF NOT EXISTS (SELECT TOP 1 1 FROM [dbo].[Customers])
BEGIN
    INSERT INTO [dbo].[Customers] ([Id], [Name], [Email], [Phone], [Address])
    VALUES 
        ('11111111-1111-1111-1111-111111111111', 'John Doe', 'john.doe@example.com', '+1-555-0101', '123 Main St, New York, NY'),
        ('22222222-2222-2222-2222-222222222222', 'Jane Smith', 'jane.smith@example.com', '+1-555-0102', '456 Oak Ave, Los Angeles, CA'),
        ('33333333-3333-3333-3333-333333333333', 'Bob Wilson', 'bob.wilson@example.com', '+1-555-0103', '789 Pine Rd, Chicago, IL'),
        ('44444444-4444-4444-4444-444444444444', 'Alice Brown', 'alice.brown@example.com', '+1-555-0104', '321 Elm St, Houston, TX'),
        ('55555555-5555-5555-5555-555555555555', 'Charlie Davis', 'charlie.davis@example.com', '+1-555-0105', '654 Maple Dr, Phoenix, AZ');
    
    PRINT 'Seeded Customers table';
END
GO

IF NOT EXISTS (SELECT TOP 1 1 FROM [dbo].[Products])
BEGIN
    INSERT INTO [dbo].[Products] ([Id], [Name], [Description], [Price], [StockQuantity], [Category])
    VALUES 
        ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Laptop Pro 15', 'High-performance laptop', 1299.99, 50, 'Electronics'),
        ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Wireless Mouse', 'Ergonomic wireless mouse', 49.99, 200, 'Electronics'),
        ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'USB-C Hub', 'Multi-port USB-C hub', 79.99, 150, 'Electronics'),
        ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Mechanical Keyboard', 'RGB mechanical keyboard', 129.99, 100, 'Electronics'),
        ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Monitor 27"', '27-inch 4K monitor', 449.99, 30, 'Electronics');
    
    PRINT 'Seeded Products table';
END
GO

PRINT '========================================';
PRINT 'SQL Server seeding complete!';
PRINT '========================================';
GO
```

---

## MongoDB Seed Script

### scripts/seed-mongodb.js

```javascript
// =============================================================================
// {ServiceName} MongoDB Initialization Script
// =============================================================================

// Switch to database (creates if not exists)
use {database_name};

print("╔══════════════════════════════════════════════════════════════╗");
print("║           MongoDB Seeding - {ServiceName}                    ║");
print("╚══════════════════════════════════════════════════════════════╝");

// =============================================================================
// COLLECTION: user_profiles
// =============================================================================
print("\n📁 Creating user_profiles collection...");

db.createCollection("user_profiles", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["userId", "email", "createdAt"],
            properties: {
                userId: { bsonType: "string", description: "User ID (UUID)" },
                email: { bsonType: "string", description: "Email address" },
                preferences: { bsonType: "object", description: "User preferences" },
                metadata: { bsonType: "object", description: "Additional metadata" },
                createdAt: { bsonType: "date", description: "Creation timestamp" },
                updatedAt: { bsonType: "date", description: "Last update timestamp" }
            }
        }
    }
});

// Create indexes
db.user_profiles.createIndex({ "userId": 1 }, { unique: true });
db.user_profiles.createIndex({ "email": 1 }, { unique: true });
db.user_profiles.createIndex({ "createdAt": -1 });

// Seed data
if (db.user_profiles.countDocuments() === 0) {
    db.user_profiles.insertMany([
        {
            userId: "11111111-1111-1111-1111-111111111111",
            email: "john.doe@example.com",
            preferences: {
                theme: "dark",
                notifications: { email: true, push: true },
                language: "en-US"
            },
            metadata: {
                lastLogin: new Date(),
                loginCount: 42
            },
            createdAt: new Date(),
            updatedAt: new Date()
        },
        {
            userId: "22222222-2222-2222-2222-222222222222",
            email: "jane.smith@example.com",
            preferences: {
                theme: "light",
                notifications: { email: true, push: false },
                language: "en-US"
            },
            metadata: {
                lastLogin: new Date(),
                loginCount: 15
            },
            createdAt: new Date(),
            updatedAt: new Date()
        }
    ]);
    print("✅ Seeded user_profiles collection");
}

// =============================================================================
// COLLECTION: audit_logs
// =============================================================================
print("\n📁 Creating audit_logs collection...");

db.createCollection("audit_logs", {
    capped: true,
    size: 104857600,  // 100MB max
    max: 100000       // 100k documents max
});

// Create indexes for querying
db.audit_logs.createIndex({ "timestamp": -1 });
db.audit_logs.createIndex({ "userId": 1, "timestamp": -1 });
db.audit_logs.createIndex({ "action": 1, "timestamp": -1 });
db.audit_logs.createIndex({ "correlationId": 1 });

// Seed sample audit logs
if (db.audit_logs.countDocuments() === 0) {
    db.audit_logs.insertMany([
        {
            userId: "11111111-1111-1111-1111-111111111111",
            action: "user.login",
            resource: "auth",
            correlationId: "corr-001",
            details: { ip: "192.168.1.1", userAgent: "Mozilla/5.0" },
            timestamp: new Date()
        },
        {
            userId: "22222222-2222-2222-2222-222222222222",
            action: "order.created",
            resource: "orders",
            correlationId: "corr-002",
            details: { orderId: "order-123", total: 299.99 },
            timestamp: new Date()
        }
    ]);
    print("✅ Seeded audit_logs collection");
}

// =============================================================================
// COLLECTION: settings
// =============================================================================
print("\n📁 Creating settings collection...");

db.createCollection("settings");
db.settings.createIndex({ "key": 1 }, { unique: true });

if (db.settings.countDocuments() === 0) {
    db.settings.insertMany([
        { key: "app.version", value: "1.0.0", updatedAt: new Date() },
        { key: "feature.darkMode", value: true, updatedAt: new Date() },
        { key: "feature.analytics", value: true, updatedAt: new Date() },
        { key: "maintenance.mode", value: false, updatedAt: new Date() }
    ]);
    print("✅ Seeded settings collection");
}

print("\n╔══════════════════════════════════════════════════════════════╗");
print("║           ✅ MongoDB Seeding Complete!                       ║");
print("╚══════════════════════════════════════════════════════════════╝");
```

---

## ScyllaDB Seed Script

### scripts/seed-scylladb.cql

```cql
-- =============================================================================
-- {ServiceName} ScyllaDB Initialization Script
-- =============================================================================

-- Create keyspace with replication
CREATE KEYSPACE IF NOT EXISTS {keyspace_name}
WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
};

USE {keyspace_name};

-- =============================================================================
-- TABLE: telemetry (Time-Series Data)
-- =============================================================================
-- Optimized for time-series queries with partition by device and clustering by timestamp

CREATE TABLE IF NOT EXISTS telemetry (
    device_id UUID,
    timestamp TIMESTAMP,
    metric_name TEXT,
    metric_value DOUBLE,
    tags MAP<TEXT, TEXT>,
    PRIMARY KEY ((device_id), timestamp, metric_name)
) WITH CLUSTERING ORDER BY (timestamp DESC, metric_name ASC)
  AND default_time_to_live = 604800  -- 7 days TTL
  AND compaction = {
    'class': 'TimeWindowCompactionStrategy',
    'compaction_window_unit': 'DAYS',
    'compaction_window_size': 1
  };

-- Secondary index for querying by metric name
CREATE INDEX IF NOT EXISTS ON telemetry (metric_name);

-- =============================================================================
-- TABLE: events (Event Sourcing)
-- =============================================================================

CREATE TABLE IF NOT EXISTS events (
    aggregate_id UUID,
    event_id TIMEUUID,
    event_type TEXT,
    event_data TEXT,  -- JSON payload
    metadata MAP<TEXT, TEXT>,
    created_at TIMESTAMP,
    PRIMARY KEY ((aggregate_id), event_id)
) WITH CLUSTERING ORDER BY (event_id ASC);

CREATE INDEX IF NOT EXISTS ON events (event_type);

-- =============================================================================
-- TABLE: metrics_aggregates (Pre-computed Aggregations)
-- =============================================================================

CREATE TABLE IF NOT EXISTS metrics_aggregates (
    metric_name TEXT,
    time_bucket TIMESTAMP,  -- Hourly buckets
    device_id UUID,
    count COUNTER,
    PRIMARY KEY ((metric_name, time_bucket), device_id)
);

-- =============================================================================
-- TABLE: device_registry
-- =============================================================================

CREATE TABLE IF NOT EXISTS device_registry (
    device_id UUID PRIMARY KEY,
    device_name TEXT,
    device_type TEXT,
    location TEXT,
    metadata MAP<TEXT, TEXT>,
    last_seen TIMESTAMP,
    created_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS ON device_registry (device_type);
CREATE INDEX IF NOT EXISTS ON device_registry (location);

-- =============================================================================
-- SEED DATA
-- =============================================================================

-- Sample devices
INSERT INTO device_registry (device_id, device_name, device_type, location, metadata, last_seen, created_at)
VALUES (
    11111111-1111-1111-1111-111111111111,
    'Temperature Sensor A1',
    'temperature_sensor',
    'Building A - Floor 1',
    {'firmware': 'v2.1.0', 'calibrated': 'true'},
    toTimestamp(now()),
    toTimestamp(now())
);

INSERT INTO device_registry (device_id, device_name, device_type, location, metadata, last_seen, created_at)
VALUES (
    22222222-2222-2222-2222-222222222222,
    'Humidity Sensor B2',
    'humidity_sensor',
    'Building B - Floor 2',
    {'firmware': 'v2.0.5', 'calibrated': 'true'},
    toTimestamp(now()),
    toTimestamp(now())
);

-- Sample telemetry data
INSERT INTO telemetry (device_id, timestamp, metric_name, metric_value, tags)
VALUES (
    11111111-1111-1111-1111-111111111111,
    toTimestamp(now()),
    'temperature',
    23.5,
    {'unit': 'celsius', 'quality': 'good'}
);

INSERT INTO telemetry (device_id, timestamp, metric_name, metric_value, tags)
VALUES (
    22222222-2222-2222-2222-222222222222,
    toTimestamp(now()),
    'humidity',
    65.2,
    {'unit': 'percent', 'quality': 'good'}
);

-- Sample events
INSERT INTO events (aggregate_id, event_id, event_type, event_data, metadata, created_at)
VALUES (
    11111111-1111-1111-1111-111111111111,
    now(),
    'device.registered',
    '{"deviceName": "Temperature Sensor A1", "deviceType": "temperature_sensor"}',
    {'correlationId': 'init-001', 'source': 'seed-script'},
    toTimestamp(now())
);
```

---

## Kafka Seed Script

### scripts/seed-kafka.ps1

```powershell
#!/usr/bin/env pwsh
# =============================================================================
# Kafka Topic Creation Script
# =============================================================================

param(
    [string]$ComposeProject = "{service-name}",
    [int]$Partitions = 3,
    [int]$ReplicationFactor = 1
)

$ErrorActionPreference = "Stop"

Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           Kafka Topic Creation                               ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

# Define topics based on domain events
$topics = @(
    # Entity events
    "{service-name}.orders.created",
    "{service-name}.orders.updated",
    "{service-name}.orders.cancelled",
    "{service-name}.orders.completed",
    
    # Notification events
    "{service-name}.notifications.email",
    "{service-name}.notifications.push",
    
    # Analytics events
    "{service-name}.analytics.pageviews",
    "{service-name}.analytics.events",
    
    # Dead letter queue
    "{service-name}.dlq",
    
    # Commands (CQRS pattern)
    "{service-name}.commands.process-order",
    "{service-name}.commands.send-notification"
)

$kafkaContainer = "$ComposeProject-kafka-1"

# Wait for Kafka to be ready
Write-Host "`n⏳ Waiting for Kafka to be ready..." -ForegroundColor Yellow
$ready = $false
$attempts = 0
$maxAttempts = 30

while (-not $ready -and $attempts -lt $maxAttempts) {
    $result = docker exec $kafkaContainer /opt/kafka/bin/kafka-topics.sh `
        --bootstrap-server localhost:9092 --list 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        $ready = $true
        Write-Host "✅ Kafka is ready" -ForegroundColor Green
    } else {
        $attempts++
        Start-Sleep -Seconds 2
    }
}

if (-not $ready) {
    Write-Host "❌ Kafka is not ready after $maxAttempts attempts" -ForegroundColor Red
    exit 1
}

# Create each topic
Write-Host "`n📨 Creating topics..." -ForegroundColor Cyan

foreach ($topic in $topics) {
    Write-Host "  Creating topic: $topic" -ForegroundColor Gray
    
    docker exec $kafkaContainer /opt/kafka/bin/kafka-topics.sh `
        --bootstrap-server localhost:9092 `
        --create `
        --topic $topic `
        --partitions $Partitions `
        --replication-factor $ReplicationFactor `
        --if-not-exists 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "    ✅ Created: $topic" -ForegroundColor Green
    } else {
        Write-Host "    ⚠️ Topic may already exist: $topic" -ForegroundColor Yellow
    }
}

# List all topics
Write-Host "`n📋 Current topics:" -ForegroundColor Cyan
docker exec $kafkaContainer /opt/kafka/bin/kafka-topics.sh `
    --bootstrap-server localhost:9092 --list

# Describe topics with partition info
Write-Host "`n📊 Topic details:" -ForegroundColor Cyan
foreach ($topic in $topics) {
    docker exec $kafkaContainer /opt/kafka/bin/kafka-topics.sh `
        --bootstrap-server localhost:9092 `
        --describe --topic $topic 2>&1 | Select-Object -First 3
}

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║           ✅ Kafka Topics Created Successfully!              ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Green
```

---

## Redis Seed Script

### scripts/seed-redis.ps1

```powershell
#!/usr/bin/env pwsh
# =============================================================================
# Redis Initialization Script
# =============================================================================

param(
    [string]$ComposeProject = "{service-name}"
)

$ErrorActionPreference = "Stop"

Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           Redis Initialization                               ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

$redisContainer = "$ComposeProject-redis-1"

# Test connectivity
Write-Host "`n🔴 Testing Redis connectivity..." -ForegroundColor Cyan
$ping = docker exec $redisContainer redis-cli ping
if ($ping -ne "PONG") {
    Write-Host "❌ Redis is not responding" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Redis is responding" -ForegroundColor Green

# =============================================================================
# Initialize Redis structures
# =============================================================================

Write-Host "`n📦 Setting up Redis data structures..." -ForegroundColor Cyan

# Configuration values (Hash)
Write-Host "  Setting up configuration hash..." -ForegroundColor Gray
docker exec $redisContainer redis-cli HSET "config:app" `
    "version" "1.0.0" `
    "environment" "development" `
    "maintenance_mode" "false" `
    "cache_ttl_seconds" "300"

# Feature flags (Hash)
Write-Host "  Setting up feature flags..." -ForegroundColor Gray
docker exec $redisContainer redis-cli HSET "features" `
    "dark_mode" "true" `
    "analytics" "true" `
    "new_checkout" "false" `
    "beta_features" "false"

# Rate limiting counters (will be created on-demand, but set up keys pattern)
Write-Host "  Setting up rate limit keys..." -ForegroundColor Gray
docker exec $redisContainer redis-cli SET "ratelimit:global:requests" "0" EX 60

# Session template (example structure)
Write-Host "  Creating session template..." -ForegroundColor Gray
docker exec $redisContainer redis-cli HSET "session:template" `
    "user_id" "" `
    "created_at" "" `
    "expires_at" "" `
    "ip_address" "" `
    "user_agent" ""

# Cache some initial data
Write-Host "  Pre-warming cache..." -ForegroundColor Gray

# Product categories (Set)
docker exec $redisContainer redis-cli SADD "cache:categories" `
    "Electronics" "Clothing" "Books" "Home" "Sports"

# Popular product IDs (Sorted Set for ranking)
docker exec $redisContainer redis-cli ZADD "cache:popular_products" `
    100 "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" `
    85 "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" `
    72 "cccccccc-cccc-cccc-cccc-cccccccccccc"

# Queue structures (Lists)
Write-Host "  Setting up queue structures..." -ForegroundColor Gray
docker exec $redisContainer redis-cli LPUSH "queue:notifications" "init"
docker exec $redisContainer redis-cli LPOP "queue:notifications"  # Remove init value

docker exec $redisContainer redis-cli LPUSH "queue:emails" "init"
docker exec $redisContainer redis-cli LPOP "queue:emails"

# Pub/Sub channels documentation (just a marker key)
Write-Host "  Documenting Pub/Sub channels..." -ForegroundColor Gray
docker exec $redisContainer redis-cli SADD "meta:pubsub_channels" `
    "events:orders" `
    "events:inventory" `
    "events:notifications" `
    "broadcast:system"

# =============================================================================
# Verify setup
# =============================================================================

Write-Host "`n📋 Verifying Redis setup..." -ForegroundColor Cyan

# Check keys
Write-Host "  Checking created keys..." -ForegroundColor Gray
$keys = docker exec $redisContainer redis-cli KEYS "*"
Write-Host "  Found keys: $($keys -join ', ')" -ForegroundColor Gray

# Memory info
Write-Host "  Memory usage:" -ForegroundColor Gray
docker exec $redisContainer redis-cli INFO memory | Select-String "used_memory_human"

# DB size
$dbsize = docker exec $redisContainer redis-cli DBSIZE
Write-Host "  Database size: $dbsize" -ForegroundColor Gray

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║           ✅ Redis Initialized Successfully!                 ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Green

# Print useful commands
Write-Host "`n📌 Useful Redis commands:" -ForegroundColor Yellow
Write-Host "  docker exec -it $redisContainer redis-cli" -ForegroundColor Gray
Write-Host "  docker exec $redisContainer redis-cli MONITOR  # Watch all commands" -ForegroundColor Gray
Write-Host "  docker exec $redisContainer redis-cli INFO    # Server info" -ForegroundColor Gray
```

---

## Docker Compose Volume Mounts

Ensure docker-compose.yml mounts scripts directory:

```yaml
services:
  sqlserver:
    image: mcr.microsoft.com/mssql/server:2022-latest
    volumes:
      - ./scripts:/scripts:ro
    # ... other config
    
  mongodb:
    image: mongo:7
    volumes:
      - ./scripts:/scripts:ro
    # ... other config
    
  scylladb:
    image: scylladb/scylla:5.4
    volumes:
      - ./scripts:/scripts:ro
    # ... other config
```

---

## Seed Script Checklist

When generating seed scripts, ensure:

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                          SEED SCRIPT CHECKLIST                                           ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   SQL Server:                                                                            ║
║   □ Database creation (IF NOT EXISTS)                                                    ║
║   □ All tables with proper data types                                                    ║
║   □ Primary keys and foreign keys                                                        ║
║   □ Indexes for common queries                                                           ║
║   □ Sample data (5-10 rows per table)                                                    ║
║   □ Idempotent (can run multiple times)                                                  ║
║                                                                                          ║
║   MongoDB:                                                                               ║
║   □ Database/collection creation                                                         ║
║   □ Schema validation (optional but recommended)                                         ║
║   □ Indexes for common queries                                                           ║
║   □ Sample documents                                                                     ║
║   □ TTL indexes if needed                                                                ║
║                                                                                          ║
║   ScyllaDB:                                                                              ║
║   □ Keyspace with replication strategy                                                   ║
║   □ Tables with proper partition/clustering keys                                         ║
║   □ TTL configuration for time-series data                                               ║
║   □ Secondary indexes if needed                                                          ║
║   □ Sample data                                                                          ║
║                                                                                          ║
║   Kafka:                                                                                 ║
║   □ All domain event topics                                                              ║
║   □ DLQ (Dead Letter Queue) topic                                                        ║
║   □ Proper partition count (typically 3+)                                                ║
║   □ Topic naming convention: {service}.{entity}.{action}                                 ║
║                                                                                          ║
║   Redis:                                                                                 ║
║   □ Configuration values                                                                 ║
║   □ Feature flags                                                                        ║
║   □ Cache structure templates                                                            ║
║   □ Queue structures (if using Redis queues)                                             ║
║   □ Pub/Sub channel documentation                                                        ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## Execution Order

Always seed in this order (respects dependencies):

1. **SQL Server** - Primary transactional data (other platforms may reference these IDs)
2. **MongoDB** - Document data (may reference SQL Server IDs)
3. **ScyllaDB** - Time-series/event data (may reference entity IDs)
4. **Kafka** - Topics must exist before producers start
5. **Redis** - Configuration and cache warm-up (last, as it caches data from other DBs)

````
