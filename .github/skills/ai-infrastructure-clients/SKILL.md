---
name: ai-infrastructure-clients
description: >
  Patterns for using AI Core infrastructure clients: Redis, Kafka, MongoDB, SQL Server, ScyllaDB.
  Use when adding data access, caching, or messaging to .NET or Go services.
  Ensures developers use Core.Infrastructure wrappers, not raw packages.
  Covers repository patterns, caching strategies, and event publishing.
---

# AI Infrastructure Clients

## Client Overview

| Client | Data Platform | Use Case |
|--------|--------------|----------|
| `ISqlServerClient` / `sqlserver.Client` | SQL Server | Transactional data: orders, payments |
| `IMongoClient` / `mongodb.Client` | MongoDB | Documents: user profiles, settings |
| `ScyllaDBClient` / `scylladb.Session` | ScyllaDB | Time-series: telemetry, metrics |
| `IRedisClient` / `redis.Client` | Redis | Caching, sessions, counters |
| `IKafkaProducer` / `kafka.Producer` | Kafka | Event publishing, async messaging |

---

## .NET Patterns

### SQL Server Repository

```csharp
using Core.Infrastructure;
using Core.Logger;

public class OrderRepository
{
    private readonly ISqlServerClient _sql;
    private readonly ServiceLogger _logger;

    public OrderRepository(ISqlServerClient sql, ServiceLogger logger)
    {
        _sql = sql;
        _logger = logger;
    }

    public async Task<Order?> GetByIdAsync(Guid id)
    {
        _logger.Information("Fetching order {OrderId}", id);
        
        return await _sql.QuerySingleOrDefaultAsync<Order>(
            "SELECT * FROM Orders WHERE Id = @Id", 
            new { Id = id });
    }

    public async Task CreateAsync(Order order)
    {
        await _sql.ExecuteAsync(
            @"INSERT INTO Orders (Id, CustomerId, Total, Status, CreatedAt)
              VALUES (@Id, @CustomerId, @Total, @Status, @CreatedAt)",
            order);
    }

    public async Task<IEnumerable<Order>> GetByCustomerAsync(Guid customerId)
    {
        return await _sql.QueryAsync<Order>(
            "SELECT * FROM Orders WHERE CustomerId = @CustomerId ORDER BY CreatedAt DESC",
            new { CustomerId = customerId });
    }
}
```

### Redis Cache

```csharp
using Core.Infrastructure;
using Core.Logger;

public class OrderCache
{
    private readonly IRedisClient _redis;
    private readonly ServiceLogger _logger;
    private readonly TimeSpan _defaultTtl = TimeSpan.FromMinutes(5);

    public OrderCache(IRedisClient redis, ServiceLogger logger)
    {
        _redis = redis;
        _logger = logger;
    }

    public async Task<Order?> GetAsync(Guid orderId)
    {
        var cached = await _redis.GetAsync<Order>($"order:{orderId}");
        
        if (cached != null)
            _logger.Information("Cache hit for order {OrderId}", orderId);
        
        return cached;
    }

    public async Task SetAsync(Order order)
    {
        await _redis.SetAsync($"order:{order.Id}", order, _defaultTtl);
        _logger.Information("Cached order {OrderId}", order.Id);
    }

    public async Task InvalidateAsync(Guid orderId)
    {
        await _redis.DeleteAsync($"order:{orderId}");
        _logger.Information("Invalidated cache for order {OrderId}", orderId);
    }
}
```

### Kafka Event Publisher

```csharp
using Core.Infrastructure.Kafka;
using Core.Logger;
using System.Text.Json;

public class OrderEventPublisher
{
    private readonly IKafkaProducer _kafka;
    private readonly ServiceLogger _logger;

    public OrderEventPublisher(IKafkaProducer kafka, ServiceLogger logger)
    {
        _kafka = kafka;
        _logger = logger;
    }

    public async Task PublishOrderCreatedAsync(Order order, string correlationId)
    {
        var message = new KafkaMessage<string>
        {
            Key = order.Id.ToString(),
            Value = JsonSerializer.Serialize(new
            {
                EventType = "order.created",
                OrderId = order.Id,
                CustomerId = order.CustomerId,
                Total = order.Total,
                Timestamp = DateTime.UtcNow
            }),
            Headers = new Dictionary<string, string>
            {
                { "correlation_id", correlationId },
                { "event_type", "order.created" }
            }
        };

        await _kafka.ProduceAsync("orders.events", message);
        _logger.Information("Published order.created event for {OrderId}", order.Id);
    }
}
```

### MongoDB Repository

```csharp
using Core.Infrastructure.MongoDB;
using Core.Logger;

public class UserProfileRepository
{
    private readonly Core.Infrastructure.MongoDB.IMongoClient _mongo;
    private readonly ServiceLogger _logger;
    private const string CollectionName = "user_profiles";

    public UserProfileRepository(
        Core.Infrastructure.MongoDB.IMongoClient mongo, 
        ServiceLogger logger)
    {
        _mongo = mongo;
        _logger = logger;
    }

    public async Task<UserProfile?> GetByIdAsync(Guid userId)
    {
        return await _mongo.FindOneAsync<UserProfile>(
            CollectionName,
            p => p.UserId == userId);
    }

    public async Task UpsertAsync(UserProfile profile)
    {
        await _mongo.ReplaceOneAsync(
            CollectionName,
            p => p.UserId == profile.UserId,
            profile,
            upsert: true);
    }
}
```

---

## Go Patterns

### SQL Server Repository

```go
import (
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/sqlserver"
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
    "go.uber.org/zap"
)

type OrderRepository struct {
    db     *sqlserver.Client
    logger *logger.Logger
}

func NewOrderRepository(db *sqlserver.Client, log *logger.Logger) *OrderRepository {
    return &OrderRepository{db: db, logger: log}
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*Order, error) {
    r.logger.Info("Fetching order", zap.String("order_id", id))
    
    var order Order
    err := r.db.QueryRow(ctx, 
        "SELECT id, customer_id, total, status FROM orders WHERE id = @p1", 
        id).Scan(&order.ID, &order.CustomerID, &order.Total, &order.Status)
    
    if err != nil {
        return nil, err
    }
    return &order, nil
}

func (r *OrderRepository) Create(ctx context.Context, order *Order) error {
    _, err := r.db.Exec(ctx,
        `INSERT INTO orders (id, customer_id, total, status, created_at)
         VALUES (@p1, @p2, @p3, @p4, @p5)`,
        order.ID, order.CustomerID, order.Total, order.Status, order.CreatedAt)
    return err
}
```

### Redis Cache

```go
import (
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/redis"
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
    "go.uber.org/zap"
    "time"
)

type OrderCache struct {
    redis  *redis.Client
    logger *logger.Logger
    ttl    time.Duration
}

func NewOrderCache(redis *redis.Client, log *logger.Logger) *OrderCache {
    return &OrderCache{
        redis:  redis,
        logger: log,
        ttl:    5 * time.Minute,
    }
}

func (c *OrderCache) Get(ctx context.Context, orderID string) (*Order, error) {
    var order Order
    err := c.redis.Get(ctx, "order:"+orderID, &order)
    
    if err == nil {
        c.logger.Info("Cache hit", zap.String("order_id", orderID))
    }
    return &order, err
}

func (c *OrderCache) Set(ctx context.Context, order *Order) error {
    err := c.redis.Set(ctx, "order:"+order.ID, order, c.ttl)
    if err == nil {
        c.logger.Info("Cached order", zap.String("order_id", order.ID))
    }
    return err
}
```

### Kafka Event Publisher

```go
import (
    "github.com/your-github-org/ai-scaffolder/core/go/infrastructure/kafka"
    "github.com/your-github-org/ai-scaffolder/core/go/logger"
    "go.uber.org/zap"
    "encoding/json"
)

type OrderEventPublisher struct {
    producer kafka.Producer
    logger   *logger.Logger
}

func NewOrderEventPublisher(producer kafka.Producer, log *logger.Logger) *OrderEventPublisher {
    return &OrderEventPublisher{producer: producer, logger: log}
}

func (p *OrderEventPublisher) PublishOrderCreated(ctx context.Context, order *Order) error {
    correlationID, _ := ctx.Value(logger.CorrelationIDKey).(string)
    
    payload, _ := json.Marshal(map[string]interface{}{
        "event_type":  "order.created",
        "order_id":    order.ID,
        "customer_id": order.CustomerID,
        "total":       order.Total,
        "timestamp":   time.Now().UTC(),
    })
    
    headers := map[string]string{
        "correlation_id": correlationID,
        "event_type":     "order.created",
    }
    
    err := p.producer.SendMessage(ctx, "orders.events", order.ID, payload, headers)
    if err == nil {
        p.logger.Info("Published order.created", zap.String("order_id", order.ID))
    }
    return err
}
```

---

## DI Registration

### .NET Program.cs

```csharp
// SQL Server
builder.Services.AddSingleton<ISqlServerClient>(sp =>
    new SqlServerClient(new SqlServerConfig 
    { 
        ConnectionString = builder.Configuration.GetConnectionString("SqlServer")! 
    }));

// Redis
builder.Services.AddSingleton<IRedisClient>(sp =>
    new RedisClient(new RedisConfig 
    { 
        ConnectionString = builder.Configuration.GetConnectionString("Redis")! 
    }));

// Kafka
builder.Services.AddSingleton<IKafkaProducer>(sp =>
    new KafkaProducer(new KafkaConfig 
    { 
        BootstrapServers = builder.Configuration["Kafka:BootstrapServers"]! 
    }));

// MongoDB
builder.Services.AddSingleton<Core.Infrastructure.MongoDB.IMongoClient>(sp =>
    new Core.Infrastructure.MongoDB.MongoClient(new MongoConfig
    {
        ConnectionString = builder.Configuration.GetConnectionString("MongoDB")!,
        DatabaseName = builder.Configuration["MongoDB:Database"]!
    }));
```

### Go main.go

```go
// SQL Server
sqlClient, _ := sqlserver.NewClient(sqlserver.ClientConfig{
    Server:   cfg.SQLServer.Server,
    Database: cfg.SQLServer.Database,
    User:     cfg.SQLServer.User,
    Password: cfg.SQLServer.Password,
    Logger:   log,
})

// Redis
redisClient, _ := redis.NewClient(redis.ClientConfig{
    Host:   cfg.Redis.Host,
    Port:   cfg.Redis.Port,
    Logger: log,
})

// Kafka
kafkaProducer, _ := kafka.NewProducer(kafka.ProducerConfig{
    Brokers: cfg.Kafka.Brokers,
    Logger:  log,
})

// MongoDB
mongoClient, _ := mongodb.NewClient(mongodb.ClientConfig{
    ConnectionURI: cfg.MongoDB.ConnectionURI,
    Database:      cfg.MongoDB.Database,
    Logger:        log,
})
```

---

## Connection String Formats

| Platform | Format |
|----------|--------|
| SQL Server | `Server=host,1433;Database=db;User Id=sa;Password=pass;TrustServerCertificate=True;` |
| Redis | `host:6379,abortConnect=false` |
| MongoDB | `mongodb://host:27017/database` |
| Kafka | `host:9092` (bootstrap servers) |
| ScyllaDB | `host:9042` (comma-separated hosts) |
