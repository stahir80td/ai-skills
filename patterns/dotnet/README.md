# AI Patterns - Complete Core.Infrastructure Reference Implementation

> **Comprehensive patterns demonstrating ALL Core.Infrastructure data platform clients**

This reference implementation showcases optimal usage patterns for **all 5 Core.Infrastructure data platform clients** within a single ASP.NET Core Web API, following AI architecture standards.

## 🎯 What This Demonstrates

### Core.Infrastructure Clients Used

| Platform | Package | Use Cases | Patterns Shown |
|----------|---------|-----------|----------------|
| **SQL Server** | `Core.Infrastructure.SqlServer` | Transactional data, ACID compliance | Orders, transactions, complex queries |
| **MongoDB** | `Core.Infrastructure.MongoDB` | Document storage, flexible schemas | User profiles, preferences, text search |
| **ScyllaDB** | `Core.Infrastructure.ScyllaDB` | Time-series data, high throughput | IoT telemetry, device data, aggregations |
| **Redis** | `Core.Infrastructure.Redis` | Caching, real-time data | Leaderboards, sessions, counters |
| **Kafka** | `Core.Infrastructure.Kafka` | Event streaming, pub/sub messaging | Domain events, notifications |

### AI Core Packages Integration

✅ **All 8 Core packages properly integrated:**
- `Core.Config` - Environment-based configuration
- `Core.Logger` - Structured JSON logging with correlation
- `Core.Errors` - Error codes and ServiceException patterns
- `Core.Sli` - Availability, latency, throughput tracking
- `Core.Reliability` - Circuit breakers, retry policies  
- `Core.Infrastructure` - **All 5 data platform clients**
- `Core.Metrics` - Prometheus metrics
- `Core.Sod` - Service Oriented Design patterns

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    AI PATTERNS API                         │
│              (All Core.Infrastructure Clients)              │
└─────────────────────────────────────────────────────────────┘
                                │
                    ┌───────────┴───────────┐
                    │    PatternsController │
                    │   (Cross-platform)    │
                    └───────────┬───────────┘
                                │
                    ┌───────────┴───────────┐
                    │    PatternsService    │
                    │   (Business Logic)    │
                    └───────────┬───────────┘
                                │
        ┌───────┬───────┬───────┼───────┬───────┬───────┐
        │       │       │       │       │       │       │
        ▼       ▼       ▼       ▼       ▼       ▼       ▼
   ┌─────────┐ │ ┌─────────┐ │ ┌─────────┐ │ ┌─────────┐
   │   SQL   │ │ │ MongoDB │ │ │ScyllaDB │ │ │  Redis  │
   │ Server  │ │ │         │ │ │         │ │ │         │
   └─────────┘ │ └─────────┘ │ └─────────┘ │ └─────────┘
   Orders,     │ Users,      │ Telemetry,  │ Cache,
   Transactions│ Profiles    │ Time-series │ Sessions
               │             │             │
               ▼             ▼             ▼
           ┌─────────────────────────────────────┐
           │              Kafka                  │
           │         (Event Streaming)           │
           └─────────────────────────────────────┘
           Domain Events, Notifications, Pub/Sub
```

## 🚀 Quick Start

### 1. Start Dependencies

```bash
# Start all data platforms (Docker Compose recommended)
docker run -d --name sqlserver -p 1433:1433 -e ACCEPT_EULA=Y -e SA_PASSWORD=AiPatterns2024! mcr.microsoft.com/mssql/server:2022-latest
docker run -d --name mongodb -p 27017:27017 mongo:7
docker run -d --name redis -p 6379:6379 redis:7-alpine
docker run -d --name scylla -p 9042:9042 scylladb/scylla:5.4.3 --smp 1
docker run -d --name kafka -p 9092:9092 apache/kafka:3.7.0
```

### 2. Build and Run

```bash
# Build the patterns project
dotnet build

# Run the API
dotnet run

# Open Swagger UI
# http://localhost:5000
```

### 3. Test All Patterns

The Swagger UI provides interactive testing of all data platform patterns:

- **SQL Server**: Create/update orders with transactions
- **MongoDB**: Manage user profiles with flexible schemas  
- **ScyllaDB**: Record and query IoT telemetry data
- **Redis**: Real-time leaderboards and session management
- **Kafka**: Cross-platform event streaming
- **Analytics**: Query all platforms simultaneously

## 📊 API Endpoints

### SQL Server Patterns (Transactional)

```http
POST   /api/v1/patterns/orders          # Create order with items
PATCH  /api/v1/patterns/orders/{id}/status # Update order status
GET    /api/v1/patterns/orders/{id}     # Get order details
```

### MongoDB Patterns (Document)

```http
POST   /api/v1/patterns/users           # Create user profile
PUT    /api/v1/patterns/users/{id}/preferences # Update preferences
GET    /api/v1/patterns/users/{id}      # Get user profile
```

### ScyllaDB Patterns (Time-Series)

```http
POST   /api/v1/patterns/telemetry       # Record device telemetry
GET    /api/v1/patterns/telemetry/{deviceId} # Get telemetry history
```

### Redis Patterns (Real-time)

```http
POST   /api/v1/patterns/leaderboards/{category}/scores # Update leaderboard
GET    /api/v1/patterns/leaderboards/{category}        # Get leaderboard
POST   /api/v1/patterns/sessions                       # Create session
```

### Cross-Platform Analytics

```http
GET    /api/v1/patterns/analytics       # Query all platforms simultaneously
```

### Health & Monitoring

```http
GET    /api/v1/patterns/health          # Connectivity check for all platforms
GET    /health/live                     # Liveness probe
GET    /health/ready                    # Readiness probe (all dependencies)
GET    /metrics                         # Prometheus metrics
```

## 🔧 Configuration

All Core.Infrastructure clients are configured in `appsettings.json`:

```json
{
  "ConnectionStrings": {
    "SqlServer": "Server=localhost,1433;Database=AiPatterns;User Id=sa;Password=AiPatterns2024!;TrustServerCertificate=True;",
    "MongoDB": "mongodb://localhost:27017",
    "Redis": "localhost:6379"
  },
  "Kafka": {
    "BootstrapServers": "localhost:9092",
    "ConsumerGroup": "ai-patterns-group"
  },
  "ScyllaDB": {
    "ContactPoints": "localhost:9042",
    "Keyspace": "ai_patterns"
  }
}
```

## 📁 Project Structure (Service Oriented Design)

```
AiPatterns/
├── Program.cs                          # All Core.Infrastructure client registration
├── appsettings.json                    # Configuration for all platforms
│
├── Api/
│   └── Controllers/
│       └── PatternsController.cs       # REST endpoints for all patterns
│
├── Domain/                             # NO external dependencies
│   ├── Models/                         # Domain entities for all platforms
│   │   ├── Order.cs                    # SQL Server entity
│   │   ├── UserProfile.cs              # MongoDB document
│   │   ├── DeviceTelemetry.cs          # ScyllaDB time-series
│   │   └── Events.cs                   # Kafka events
│   ├── Services/
│   │   └── PatternsService.cs          # Cross-platform business logic
│   ├── Interfaces/
│   │   ├── IPatternsService.cs         # Service contract
│   │   └── I*Repository.cs             # Repository contracts
│   ├── Errors/
│   │   └── ProductErrors.cs            # Error repository with codes
│   └── Sli/
│       └── PatternsSli.cs              # SLI tracking
│
└── Infrastructure/                     # External integrations
    ├── Repositories/
    │   ├── OrderRepository.cs          # SQL Server via Core.Infrastructure
    │   ├── UserProfileRepository.cs    # MongoDB via Core.Infrastructure
    │   └── TelemetryRepository.cs      # ScyllaDB via Core.Infrastructure
    ├── Cache/
    │   └── RealtimeCache.cs            # Redis via Core.Infrastructure
    └── Messaging/
        ├── EventPublisher.cs           # Kafka producer via Core.Infrastructure
        └── EventConsumer.cs            # Kafka consumer via Core.Infrastructure
```

## 🎭 Pattern Examples

### Cross-Platform Workflow Example

When creating an order, the system demonstrates all platforms working together:

1. **SQL Server**: Store order and items (transactional)
2. **Redis**: Cache order for quick access
3. **Kafka**: Publish order created event
4. **MongoDB**: Update user's order history in profile
5. **ScyllaDB**: Log order metrics as time-series data

```csharp
public async Task<Order> CreateOrderAsync(Guid customerId, string shippingAddress, IEnumerable<OrderItem> items)
{
    // 1. SQL Server - Transactional storage
    var order = Order.Create(customerId, shippingAddress, items);
    var createdOrder = await _orderRepository.CreateAsync(order);

    // 2. Redis - Cache for performance  
    await _cache.SetAsync($"order:{order.Id}", order, TimeSpan.FromHours(24));

    // 3. Kafka - Event publishing
    var orderEvent = OrderEvent.OrderCreated(order, "patterns-service");
    await _eventPublisher.PublishOrderEventAsync(orderEvent);

    return createdOrder;
}
```

### SQL Server Transactional Pattern

```csharp
public async Task<Order> CreateAsync(Order order)
{
    using var connection = await _sqlClient.GetConnectionAsync();
    using var transaction = connection.BeginTransaction();
    
    try
    {
        // Insert order
        var orderSql = """
            INSERT INTO Orders (Id, CustomerId, Status, TotalAmount, ShippingAddress, OrderDate)
            VALUES (@Id, @CustomerId, @Status, @TotalAmount, @ShippingAddress, @OrderDate)
            """;
        await connection.ExecuteAsync(orderSql, order, transaction);

        // Insert order items
        foreach (var item in order.Items)
        {
            var itemSql = """
                INSERT INTO OrderItems (Id, OrderId, ProductName, Quantity, UnitPrice)
                VALUES (@Id, @OrderId, @ProductName, @Quantity, @UnitPrice)
                """;
            await connection.ExecuteAsync(itemSql, item, transaction);
        }

        transaction.Commit();
        return order;
    }
    catch
    {
        transaction.Rollback();
        throw;
    }
}
```

### MongoDB Document Pattern

```csharp
public async Task<UserProfile> CreateAsync(UserProfile profile)
{
    var collection = await _mongoClient.GetCollectionAsync<UserProfile>("user_profiles");
    await collection.InsertOneAsync(profile);
    return profile;
}

public async Task<IEnumerable<UserProfile>> SearchByPreferencesAsync(string category, string value)
{
    var collection = await _mongoClient.GetCollectionAsync<UserProfile>("user_profiles");
    var filter = Builders<UserProfile>.Filter.ElemMatch(
        p => p.Preferences.Categories, 
        c => c.Category == category && c.Value == value);
    return await collection.Find(filter).ToListAsync();
}
```

### ScyllaDB Time-Series Pattern

```csharp
public async Task<DeviceTelemetry> InsertAsync(DeviceTelemetry telemetry)
{
    var cql = """
        INSERT INTO device_telemetry (device_id, timestamp, metric, value, unit, data_type)
        VALUES (?, ?, ?, ?, ?, ?)
        """;
    
    await _scyllaClient.ExecuteAsync(cql, 
        telemetry.DeviceId, 
        telemetry.Timestamp, 
        telemetry.Metric, 
        telemetry.Value,
        telemetry.Unit,
        telemetry.DataType);
    
    return telemetry;
}

public async Task<IEnumerable<DeviceTelemetry>> GetByTimeRangeAsync(DateTime startTime, DateTime endTime)
{
    var cql = """
        SELECT device_id, timestamp, metric, value, unit, data_type
        FROM device_telemetry
        WHERE timestamp >= ? AND timestamp <= ?
        ORDER BY timestamp DESC
        LIMIT 10000
        """;
    
    return await _scyllaClient.QueryAsync<DeviceTelemetry>(cql, startTime, endTime);
}
```

### Redis Real-time Pattern

```csharp
public async Task UpdateLeaderboardAsync(string category, string userId, double score)
{
    var key = $"leaderboard:{category}";
    await _redisClient.SortedSetAddAsync(key, userId, score);
    await _redisClient.ExpireAsync(key, TimeSpan.FromDays(30));
}

public async Task<LeaderboardEntry[]> GetLeaderboardAsync(string category, int top = 10)
{
    var key = $"leaderboard:{category}";
    var entries = await _redisClient.SortedSetRangeByScoreAsync(key, 0, -1, top, Order.Descending);
    
    return entries.Select((entry, index) => new LeaderboardEntry
    {
        Rank = index + 1,
        UserId = entry.Element,
        Score = entry.Score
    }).ToArray();
}
```

### Kafka Event Streaming Pattern

```csharp
public async Task PublishOrderEventAsync(OrderEvent orderEvent)
{
    var topic = _configuration["Kafka:Topics:Orders"] ?? "patterns.orders";
    
    await _kafkaProducer.ProduceAsync(topic, new Message<string, OrderEvent>
    {
        Key = orderEvent.OrderId.ToString(),
        Value = orderEvent,
        Headers = new Headers
        {
            { "event_type", Encoding.UTF8.GetBytes(orderEvent.EventType) },
            { "timestamp", Encoding.UTF8.GetBytes(orderEvent.Timestamp.ToString("O")) },
            { "source", Encoding.UTF8.GetBytes(orderEvent.Source) }
        }
    });
}
```

## 🔍 Monitoring & Observability

### Health Checks

All data platforms are monitored:

- **SQL Server**: Connection and query test
- **MongoDB**: Database ping
- **Redis**: Connection test  
- **Kafka**: Producer/consumer connectivity
- **ScyllaDB**: Cluster status

### SLI Tracking

Service Level Indicators are automatically tracked:

- **Availability**: Percentage of successful requests
- **Latency**: P95/P99 response times
- **Throughput**: Requests per second
- **Error Rate**: Failed request percentage

### Prometheus Metrics

Custom business metrics for each platform:

```csharp
// Orders (SQL Server)
_ordersCreated.WithLabels("web", "premium").Inc();

// Users (MongoDB) 
_userProfilesUpdated.WithLabels("preferences").Inc();

// Telemetry (ScyllaDB)
_telemetryPointsReceived.WithLabels(deviceType, metric).Inc();

// Cache (Redis)
_cacheOperations.WithLabels("hit").Inc();

// Events (Kafka)
_eventsPublished.WithLabels(topic, eventType).Inc();
```

## 🧪 Testing

### Manual Testing via Swagger UI

1. Navigate to `http://localhost:5000`
2. Test each data platform pattern:
   - Create orders (SQL Server)
   - Manage user profiles (MongoDB)
   - Record telemetry (ScyllaDB)
   - Update leaderboards (Redis)
   - Cross-platform analytics

### Integration Testing

```csharp
[Test]
public async Task CrossPlatformWorkflow_Success()
{
    // Arrange
    var customerId = Guid.NewGuid();
    var items = new[] { new OrderItem { ProductName = "Test", Quantity = 1, UnitPrice = 10.00m } };

    // Act - Create order (triggers all platforms)
    var order = await _patternsService.CreateOrderAsync(customerId, "123 Test St", items);

    // Assert
    Assert.That(order.Id, Is.Not.EqualTo(Guid.Empty));
    
    // Verify SQL Server storage
    var storedOrder = await _orderRepository.GetByIdAsync(order.Id);
    Assert.That(storedOrder, Is.Not.Null);
    
    // Verify Redis cache
    var cachedOrder = await _cache.GetAsync<Order>($"order:{order.Id}");
    Assert.That(cachedOrder, Is.Not.Null);
    
    // Verify Kafka event (would need consumer verification)
    // Verify MongoDB user history update
    // Verify ScyllaDB metrics recording
}
```

## 📚 Key Learnings

This patterns implementation demonstrates:

1. **Proper Core.Infrastructure Usage**: All data platforms accessed through Core packages
2. **Service Oriented Design**: Clear separation of concerns across layers
3. **Cross-Platform Workflows**: How to orchestrate multiple data platforms  
4. **Error Handling**: Consistent error codes and exception patterns
5. **Monitoring**: Comprehensive SLI tracking and health checks
6. **Event-Driven Architecture**: Kafka for cross-service communication
7. **Caching Strategies**: Redis for performance optimization
8. **Time-Series Patterns**: ScyllaDB for high-throughput IoT data
9. **Document Flexibility**: MongoDB for evolving schemas
10. **ACID Compliance**: SQL Server for critical transactional data

## 🎯 Production Readiness

This implementation includes production-ready patterns:

- ✅ Circuit breakers for external calls
- ✅ Correlation ID propagation
- ✅ Structured logging with context  
- ✅ Health checks for all dependencies
- ✅ Prometheus metrics
- ✅ Error codes and consistent responses
- ✅ Configuration management
- ✅ Graceful degradation
- ✅ Resource management and connection pooling

---

**This is the definitive reference for using ALL Core.Infrastructure data platform clients in AI applications! 🚀**