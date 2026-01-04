---
name: ai-unit-testing
description: Unit testing skill for .NET, Go, Python, TypeScript, and React applications. Enforces 80% code coverage minimum. Provides patterns for test organization, mocking, assertions, and coverage reporting. Run tests iteratively until all pass. Use after code quality (Phase 2.5) and before Docker build (Phase 4).
---

# AI Unit Testing Skill

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  UNIT TESTING IS MANDATORY  ⚠️                                  ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   ALL services MUST have unit tests with ≥80% code coverage!                             ║
║                                                                                          ║
║   This phase occurs AFTER:                                                               ║
║     ✅ Phase 2.5: Code Quality (format & lint)                                           ║
║                                                                                          ║
║   This phase occurs BEFORE:                                                              ║
║     ➡️  Phase 4: Docker Build                                                            ║
║                                                                                          ║
║   ❌ DO NOT proceed to Docker build until ALL tests pass!                                ║
║   ❌ DO NOT proceed with <80% coverage without explicit approval!                        ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## Coverage Requirements

| Language | Minimum Coverage | Tool |
|----------|------------------|------|
| **.NET** | 80% | `coverlet` + `dotnet test` |
| **Go** | 80% | `go test -cover` |
| **Python** | 80% | `pytest-cov` |
| **TypeScript** | 80% | `vitest --coverage` or `jest --coverage` |
| **React** | 80% | `vitest --coverage` (components + hooks) |

---

## .NET Unit Testing

### Project Setup

```xml
<!-- {ServiceName}.Tests.csproj -->
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>enable</Nullable>
    <IsPackable>false</IsPackable>
    <IsTestProject>true</IsTestProject>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Microsoft.NET.Test.Sdk" Version="17.9.0" />
    <PackageReference Include="xunit" Version="2.7.0" />
    <PackageReference Include="xunit.runner.visualstudio" Version="2.5.7">
      <PrivateAssets>all</PrivateAssets>
      <IncludeAssets>runtime; build; native; contentfiles; analyzers</IncludeAssets>
    </PackageReference>
    <PackageReference Include="Moq" Version="4.20.70" />
    <PackageReference Include="FluentAssertions" Version="6.12.0" />
    <PackageReference Include="coverlet.collector" Version="6.0.2">
      <PrivateAssets>all</PrivateAssets>
      <IncludeAssets>runtime; build; native; contentfiles; analyzers</IncludeAssets>
    </PackageReference>
    <PackageReference Include="coverlet.msbuild" Version="6.0.2">
      <PrivateAssets>all</PrivateAssets>
      <IncludeAssets>runtime; build; native; contentfiles; analyzers</IncludeAssets>
    </PackageReference>
  </ItemGroup>

  <ItemGroup>
    <ProjectReference Include="..\{ServiceName}.Api\{ServiceName}.Api.csproj" />
  </ItemGroup>
</Project>
```

### Test Structure

```
{ServiceName}.Tests/
├── Controllers/
│   └── {Entity}ControllerTests.cs
├── Services/
│   └── {Entity}ServiceTests.cs
├── Repositories/
│   └── {Entity}RepositoryTests.cs
├── Validators/
│   └── {Entity}ValidatorTests.cs
├── Helpers/
│   └── TestDataFactory.cs
└── {ServiceName}.Tests.csproj
```

### Controller Test Example

```csharp
using FluentAssertions;
using Microsoft.AspNetCore.Mvc;
using Moq;
using Xunit;

namespace OrderManagement.Tests.Controllers;

public class OrdersControllerTests
{
    private readonly Mock<IOrderService> _mockService;
    private readonly Mock<IServiceLogger> _mockLogger;
    private readonly OrdersController _controller;

    public OrdersControllerTests()
    {
        _mockService = new Mock<IOrderService>();
        _mockLogger = new Mock<IServiceLogger>();
        _controller = new OrdersController(_mockService.Object, _mockLogger.Object);
    }

    [Fact]
    public async Task GetById_WithValidId_ReturnsOkWithOrder()
    {
        // Arrange
        var orderId = Guid.NewGuid();
        var expectedOrder = new OrderDto { Id = orderId, Status = "Pending" };
        _mockService.Setup(s => s.GetByIdAsync(orderId))
            .ReturnsAsync(expectedOrder);

        // Act
        var result = await _controller.GetById(orderId);

        // Assert
        var okResult = result.Result.Should().BeOfType<OkObjectResult>().Subject;
        var order = okResult.Value.Should().BeOfType<OrderDto>().Subject;
        order.Id.Should().Be(orderId);
    }

    [Fact]
    public async Task GetById_WithInvalidId_ReturnsNotFound()
    {
        // Arrange
        var orderId = Guid.NewGuid();
        _mockService.Setup(s => s.GetByIdAsync(orderId))
            .ReturnsAsync((OrderDto?)null);

        // Act
        var result = await _controller.GetById(orderId);

        // Assert
        result.Result.Should().BeOfType<NotFoundResult>();
    }

    [Fact]
    public async Task Create_WithValidOrder_ReturnsCreatedAtAction()
    {
        // Arrange
        var createDto = new CreateOrderDto { CustomerId = Guid.NewGuid() };
        var createdOrder = new OrderDto { Id = Guid.NewGuid(), Status = "Pending" };
        _mockService.Setup(s => s.CreateAsync(createDto))
            .ReturnsAsync(createdOrder);

        // Act
        var result = await _controller.Create(createDto);

        // Assert
        var createdResult = result.Result.Should().BeOfType<CreatedAtActionResult>().Subject;
        createdResult.ActionName.Should().Be(nameof(OrdersController.GetById));
    }

    [Fact]
    public async Task Create_WithInvalidModel_ReturnsBadRequest()
    {
        // Arrange
        _controller.ModelState.AddModelError("CustomerId", "Required");

        // Act
        var result = await _controller.Create(new CreateOrderDto());

        // Assert
        result.Result.Should().BeOfType<BadRequestObjectResult>();
    }

    [Fact]
    public async Task Delete_WithValidId_ReturnsNoContent()
    {
        // Arrange
        var orderId = Guid.NewGuid();
        _mockService.Setup(s => s.DeleteAsync(orderId))
            .ReturnsAsync(true);

        // Act
        var result = await _controller.Delete(orderId);

        // Assert
        result.Should().BeOfType<NoContentResult>();
    }

    [Fact]
    public async Task Delete_WithInvalidId_ReturnsNotFound()
    {
        // Arrange
        var orderId = Guid.NewGuid();
        _mockService.Setup(s => s.DeleteAsync(orderId))
            .ReturnsAsync(false);

        // Act
        var result = await _controller.Delete(orderId);

        // Assert
        result.Should().BeOfType<NotFoundResult>();
    }
}
```

### Service Test Example

```csharp
using FluentAssertions;
using Moq;
using Xunit;

namespace OrderManagement.Tests.Services;

public class OrderServiceTests
{
    private readonly Mock<IOrderRepository> _mockRepo;
    private readonly Mock<IKafkaClient> _mockKafka;
    private readonly Mock<IServiceLogger> _mockLogger;
    private readonly OrderService _service;

    public OrderServiceTests()
    {
        _mockRepo = new Mock<IOrderRepository>();
        _mockKafka = new Mock<IKafkaClient>();
        _mockLogger = new Mock<IServiceLogger>();
        _service = new OrderService(_mockRepo.Object, _mockKafka.Object, _mockLogger.Object);
    }

    [Fact]
    public async Task CreateAsync_WithValidOrder_PublishesEvent()
    {
        // Arrange
        var createDto = new CreateOrderDto
        {
            CustomerId = Guid.NewGuid(),
            Items = new List<OrderItemDto>
            {
                new() { ProductId = Guid.NewGuid(), Quantity = 2, UnitPrice = 10.00m }
            }
        };

        _mockRepo.Setup(r => r.CreateAsync(It.IsAny<Order>()))
            .ReturnsAsync((Order o) => o);

        // Act
        var result = await _service.CreateAsync(createDto);

        // Assert
        result.Should().NotBeNull();
        result.Status.Should().Be("Pending");
        result.TotalAmount.Should().Be(20.00m);

        _mockKafka.Verify(k => k.ProduceAsync(
            "order-management.order.created",
            It.IsAny<string>(),
            It.IsAny<object>()),
            Times.Once);
    }

    [Fact]
    public async Task GetByIdAsync_WhenNotFound_ReturnsNull()
    {
        // Arrange
        var orderId = Guid.NewGuid();
        _mockRepo.Setup(r => r.GetByIdAsync(orderId))
            .ReturnsAsync((Order?)null);

        // Act
        var result = await _service.GetByIdAsync(orderId);

        // Assert
        result.Should().BeNull();
    }

    [Theory]
    [InlineData("Pending", "Confirmed", true)]
    [InlineData("Confirmed", "Shipped", true)]
    [InlineData("Shipped", "Delivered", true)]
    [InlineData("Delivered", "Pending", false)]  // Invalid transition
    [InlineData("Cancelled", "Confirmed", false)] // Invalid transition
    public async Task UpdateStatusAsync_ValidatesTransition(
        string fromStatus, string toStatus, bool shouldSucceed)
    {
        // Arrange
        var orderId = Guid.NewGuid();
        var order = new Order { Id = orderId, Status = fromStatus };
        _mockRepo.Setup(r => r.GetByIdAsync(orderId)).ReturnsAsync(order);

        // Act
        var result = await _service.UpdateStatusAsync(orderId, toStatus);

        // Assert
        if (shouldSucceed)
        {
            result.Should().NotBeNull();
            result!.Status.Should().Be(toStatus);
        }
        else
        {
            result.Should().BeNull();
        }
    }
}
```

### Run Tests with Coverage

```powershell
# Run tests with coverage collection
dotnet test --collect:"XPlat Code Coverage" --results-directory ./coverage

# Generate coverage report (requires reportgenerator)
dotnet tool install -g dotnet-reportgenerator-globaltool
reportgenerator -reports:./coverage/**/coverage.cobertura.xml -targetdir:./coverage/report -reporttypes:Html

# Check coverage meets threshold (80%)
$coverage = [xml](Get-Content ./coverage/**/coverage.cobertura.xml)
$lineRate = [double]$coverage.coverage.'line-rate' * 100
Write-Host "Line Coverage: $lineRate%"
if ($lineRate -lt 80) {
    Write-Host "❌ Coverage below 80% threshold!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Coverage meets 80% threshold" -ForegroundColor Green
```

---

## Go Unit Testing

### Project Setup

```
{service-name}/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   ├── orders.go
│   │   └── orders_test.go       # Tests alongside source
│   ├── services/
│   │   ├── order_service.go
│   │   └── order_service_test.go
│   └── repository/
│       ├── order_repo.go
│       └── order_repo_test.go
├── go.mod
└── go.sum
```

### Dependencies

```go
// go.mod
require (
    github.com/stretchr/testify v1.9.0
    go.uber.org/mock v0.4.0
)
```

### Generate Mocks

```bash
# Install mockgen
go install go.uber.org/mock/mockgen@latest

# Generate mocks for interfaces
mockgen -source=internal/services/interfaces.go -destination=internal/mocks/services_mock.go -package=mocks
mockgen -source=internal/repository/interfaces.go -destination=internal/mocks/repository_mock.go -package=mocks
```

### Handler Test Example

```go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/mock/gomock"

    "order-service/internal/handlers"
    "order-service/internal/mocks"
    "order-service/internal/models"
)

func setupTest(t *testing.T) (*gomock.Controller, *mocks.MockOrderService, *gin.Engine) {
    ctrl := gomock.NewController(t)
    mockService := mocks.NewMockOrderService(ctrl)

    gin.SetMode(gin.TestMode)
    router := gin.New()

    handler := handlers.NewOrderHandler(mockService)
    api := router.Group("/api/v1")
    api.GET("/orders/:id", handler.GetByID)
    api.POST("/orders", handler.Create)
    api.PUT("/orders/:id", handler.Update)
    api.DELETE("/orders/:id", handler.Delete)

    return ctrl, mockService, router
}

func TestGetByID_Success(t *testing.T) {
    ctrl, mockService, router := setupTest(t)
    defer ctrl.Finish()

    orderID := uuid.New()
    expectedOrder := &models.Order{
        ID:     orderID,
        Status: "Pending",
    }

    mockService.EXPECT().
        GetByID(gomock.Any(), orderID).
        Return(expectedOrder, nil)

    req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+orderID.String(), nil)
    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var response models.Order
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    require.NoError(t, err)
    assert.Equal(t, orderID, response.ID)
}

func TestGetByID_NotFound(t *testing.T) {
    ctrl, mockService, router := setupTest(t)
    defer ctrl.Finish()

    orderID := uuid.New()

    mockService.EXPECT().
        GetByID(gomock.Any(), orderID).
        Return(nil, nil)

    req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+orderID.String(), nil)
    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCreate_Success(t *testing.T) {
    ctrl, mockService, router := setupTest(t)
    defer ctrl.Finish()

    createDTO := models.CreateOrderDTO{
        CustomerID: uuid.New(),
        Items: []models.OrderItemDTO{
            {ProductID: uuid.New(), Quantity: 2, UnitPrice: 10.00},
        },
    }

    createdOrder := &models.Order{
        ID:          uuid.New(),
        CustomerID:  createDTO.CustomerID,
        Status:      "Pending",
        TotalAmount: 20.00,
    }

    mockService.EXPECT().
        Create(gomock.Any(), gomock.Any()).
        Return(createdOrder, nil)

    body, _ := json.Marshal(createDTO)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreate_InvalidPayload(t *testing.T) {
    _, _, router := setupTest(t)

    req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader([]byte("invalid")))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusBadRequest, rec.Code)
}
```

### Service Test Example

```go
package services_test

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/mock/gomock"

    "order-service/internal/mocks"
    "order-service/internal/models"
    "order-service/internal/services"
)

func TestOrderService_Create(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockOrderRepository(ctrl)
    mockKafka := mocks.NewMockKafkaClient(ctrl)
    mockLogger := mocks.NewMockLogger(ctrl)

    service := services.NewOrderService(mockRepo, mockKafka, mockLogger)

    t.Run("creates order with correct total", func(t *testing.T) {
        ctx := context.Background()
        dto := &models.CreateOrderDTO{
            CustomerID: uuid.New(),
            Items: []models.OrderItemDTO{
                {ProductID: uuid.New(), Quantity: 2, UnitPrice: 10.00},
                {ProductID: uuid.New(), Quantity: 1, UnitPrice: 5.00},
            },
        }

        mockRepo.EXPECT().
            Create(ctx, gomock.Any()).
            DoAndReturn(func(_ context.Context, order *models.Order) (*models.Order, error) {
                return order, nil
            })

        mockKafka.EXPECT().
            Produce(gomock.Any(), "order-management.order.created", gomock.Any(), gomock.Any()).
            Return(nil)

        mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

        result, err := service.Create(ctx, dto)

        require.NoError(t, err)
        assert.Equal(t, 25.00, result.TotalAmount) // 2*10 + 1*5
        assert.Equal(t, "Pending", result.Status)
    })
}

func TestOrderService_UpdateStatus(t *testing.T) {
    testCases := []struct {
        name          string
        fromStatus    string
        toStatus      string
        shouldSucceed bool
    }{
        {"pending to confirmed", "Pending", "Confirmed", true},
        {"confirmed to shipped", "Confirmed", "Shipped", true},
        {"shipped to delivered", "Shipped", "Delivered", true},
        {"delivered to pending - invalid", "Delivered", "Pending", false},
        {"cancelled to confirmed - invalid", "Cancelled", "Confirmed", false},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockRepo := mocks.NewMockOrderRepository(ctrl)
            mockKafka := mocks.NewMockKafkaClient(ctrl)
            mockLogger := mocks.NewMockLogger(ctrl)

            service := services.NewOrderService(mockRepo, mockKafka, mockLogger)

            orderID := uuid.New()
            order := &models.Order{ID: orderID, Status: tc.fromStatus}

            mockRepo.EXPECT().
                GetByID(gomock.Any(), orderID).
                Return(order, nil)

            if tc.shouldSucceed {
                mockRepo.EXPECT().
                    Update(gomock.Any(), gomock.Any()).
                    DoAndReturn(func(_ context.Context, o *models.Order) (*models.Order, error) {
                        return o, nil
                    })
                mockKafka.EXPECT().
                    Produce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
                    Return(nil)
            }

            mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
            mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

            result, err := service.UpdateStatus(context.Background(), orderID, tc.toStatus)

            if tc.shouldSucceed {
                require.NoError(t, err)
                assert.Equal(t, tc.toStatus, result.Status)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

### Run Tests with Coverage

```bash
# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out -o coverage.html

# Check coverage percentage
go tool cover -func=coverage.out | grep total

# Enforce 80% threshold
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "❌ Coverage ${COVERAGE}% is below 80% threshold!"
    exit 1
fi
echo "✅ Coverage ${COVERAGE}% meets 80% threshold"
```

---

## Python Unit Testing

### Project Setup

```
{service-name}/
├── src/
│   └── {service_name}/
│       ├── __init__.py
│       ├── handlers/
│       ├── services/
│       └── repository/
├── tests/
│   ├── __init__.py
│   ├── conftest.py
│   ├── test_handlers.py
│   ├── test_services.py
│   └── test_repository.py
├── pyproject.toml
└── requirements-dev.txt
```

### Dependencies (requirements-dev.txt)

```
pytest>=8.0.0
pytest-cov>=4.1.0
pytest-asyncio>=0.23.0
pytest-mock>=3.12.0
httpx>=0.27.0
```

### pyproject.toml

```toml
[tool.pytest.ini_options]
testpaths = ["tests"]
asyncio_mode = "auto"
addopts = "-v --cov=src --cov-report=html --cov-report=term-missing --cov-fail-under=80"

[tool.coverage.run]
source = ["src"]
omit = ["*/tests/*", "*/__init__.py"]

[tool.coverage.report]
exclude_lines = [
    "pragma: no cover",
    "def __repr__",
    "raise NotImplementedError",
    "if TYPE_CHECKING:",
]
```

### conftest.py (Fixtures)

```python
import pytest
from unittest.mock import AsyncMock, MagicMock
from uuid import uuid4

@pytest.fixture
def mock_order_repo():
    repo = AsyncMock()
    return repo

@pytest.fixture
def mock_kafka_client():
    kafka = AsyncMock()
    return kafka

@pytest.fixture
def mock_logger():
    logger = MagicMock()
    return logger

@pytest.fixture
def sample_order():
    return {
        "id": str(uuid4()),
        "customer_id": str(uuid4()),
        "status": "Pending",
        "total_amount": 100.00,
        "items": [
            {"product_id": str(uuid4()), "quantity": 2, "unit_price": 50.00}
        ]
    }
```

### Service Test Example

```python
import pytest
from unittest.mock import AsyncMock
from uuid import uuid4

from order_service.services.order_service import OrderService
from order_service.models import CreateOrderDTO, OrderItemDTO

class TestOrderService:
    @pytest.fixture
    def service(self, mock_order_repo, mock_kafka_client, mock_logger):
        return OrderService(
            repository=mock_order_repo,
            kafka_client=mock_kafka_client,
            logger=mock_logger
        )

    async def test_create_order_calculates_total(
        self, service, mock_order_repo, mock_kafka_client
    ):
        # Arrange
        dto = CreateOrderDTO(
            customer_id=uuid4(),
            items=[
                OrderItemDTO(product_id=uuid4(), quantity=2, unit_price=10.00),
                OrderItemDTO(product_id=uuid4(), quantity=1, unit_price=5.00),
            ]
        )

        mock_order_repo.create.return_value = {"id": str(uuid4()), **dto.dict()}

        # Act
        result = await service.create(dto)

        # Assert
        assert result["total_amount"] == 25.00  # 2*10 + 1*5
        assert result["status"] == "Pending"
        mock_kafka_client.produce.assert_called_once()

    async def test_get_by_id_returns_none_when_not_found(
        self, service, mock_order_repo
    ):
        # Arrange
        mock_order_repo.get_by_id.return_value = None

        # Act
        result = await service.get_by_id(uuid4())

        # Assert
        assert result is None

    @pytest.mark.parametrize(
        "from_status,to_status,should_succeed",
        [
            ("Pending", "Confirmed", True),
            ("Confirmed", "Shipped", True),
            ("Shipped", "Delivered", True),
            ("Delivered", "Pending", False),
            ("Cancelled", "Confirmed", False),
        ]
    )
    async def test_update_status_validates_transitions(
        self, service, mock_order_repo, from_status, to_status, should_succeed
    ):
        # Arrange
        order_id = uuid4()
        mock_order_repo.get_by_id.return_value = {
            "id": str(order_id),
            "status": from_status
        }

        # Act
        if should_succeed:
            result = await service.update_status(order_id, to_status)
            assert result["status"] == to_status
        else:
            with pytest.raises(ValueError):
                await service.update_status(order_id, to_status)
```

### Handler Test Example (FastAPI)

```python
import pytest
from httpx import AsyncClient
from fastapi import FastAPI
from unittest.mock import AsyncMock
from uuid import uuid4

from order_service.handlers.orders import router
from order_service.dependencies import get_order_service

@pytest.fixture
def mock_service():
    return AsyncMock()

@pytest.fixture
def app(mock_service):
    app = FastAPI()
    app.include_router(router, prefix="/api/v1")
    app.dependency_overrides[get_order_service] = lambda: mock_service
    return app

@pytest.fixture
async def client(app):
    async with AsyncClient(app=app, base_url="http://test") as ac:
        yield ac

class TestOrdersHandler:
    async def test_get_by_id_returns_order(self, client, mock_service, sample_order):
        mock_service.get_by_id.return_value = sample_order

        response = await client.get(f"/api/v1/orders/{sample_order['id']}")

        assert response.status_code == 200
        assert response.json()["id"] == sample_order["id"]

    async def test_get_by_id_returns_404_when_not_found(self, client, mock_service):
        mock_service.get_by_id.return_value = None

        response = await client.get(f"/api/v1/orders/{uuid4()}")

        assert response.status_code == 404

    async def test_create_returns_201(self, client, mock_service, sample_order):
        mock_service.create.return_value = sample_order

        response = await client.post(
            "/api/v1/orders",
            json={
                "customer_id": str(uuid4()),
                "items": [{"product_id": str(uuid4()), "quantity": 1, "unit_price": 10.00}]
            }
        )

        assert response.status_code == 201

    async def test_create_returns_422_with_invalid_payload(self, client):
        response = await client.post("/api/v1/orders", json={})

        assert response.status_code == 422
```

### Run Tests with Coverage

```bash
# Run tests with coverage
pytest --cov=src --cov-report=html --cov-report=term-missing --cov-fail-under=80

# Just check coverage percentage
pytest --cov=src --cov-report=term-missing | grep TOTAL
```

---

## TypeScript Unit Testing

### Project Setup (Vitest)

```
{service-name}/
├── src/
│   ├── services/
│   │   ├── orderService.ts
│   │   └── orderService.test.ts
│   ├── handlers/
│   └── repository/
├── vitest.config.ts
├── package.json
└── tsconfig.json
```

### Dependencies (package.json)

```json
{
  "devDependencies": {
    "vitest": "^1.3.0",
    "@vitest/coverage-v8": "^1.3.0",
    "msw": "^2.2.0"
  },
  "scripts": {
    "test": "vitest",
    "test:coverage": "vitest run --coverage",
    "test:ci": "vitest run --coverage --coverage.thresholds.lines=80"
  }
}
```

### vitest.config.ts

```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      exclude: ['**/*.test.ts', '**/__mocks__/**', '**/types/**'],
      thresholds: {
        lines: 80,
        branches: 80,
        functions: 80,
        statements: 80,
      },
    },
  },
});
```

### Service Test Example

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { OrderService } from './orderService';
import { OrderRepository } from '../repository/orderRepository';
import { KafkaClient } from '../infrastructure/kafkaClient';

vi.mock('../repository/orderRepository');
vi.mock('../infrastructure/kafkaClient');

describe('OrderService', () => {
  let service: OrderService;
  let mockRepo: vi.Mocked<OrderRepository>;
  let mockKafka: vi.Mocked<KafkaClient>;

  beforeEach(() => {
    mockRepo = new OrderRepository() as vi.Mocked<OrderRepository>;
    mockKafka = new KafkaClient() as vi.Mocked<KafkaClient>;
    service = new OrderService(mockRepo, mockKafka);
    vi.clearAllMocks();
  });

  describe('create', () => {
    it('should calculate total correctly', async () => {
      const dto = {
        customerId: 'customer-1',
        items: [
          { productId: 'prod-1', quantity: 2, unitPrice: 10.0 },
          { productId: 'prod-2', quantity: 1, unitPrice: 5.0 },
        ],
      };

      mockRepo.create.mockResolvedValue({
        id: 'order-1',
        ...dto,
        status: 'Pending',
        totalAmount: 25.0,
      });

      const result = await service.create(dto);

      expect(result.totalAmount).toBe(25.0);
      expect(result.status).toBe('Pending');
      expect(mockKafka.produce).toHaveBeenCalledWith(
        'order-management.order.created',
        expect.any(String),
        expect.any(Object)
      );
    });
  });

  describe('updateStatus', () => {
    it.each([
      ['Pending', 'Confirmed', true],
      ['Confirmed', 'Shipped', true],
      ['Shipped', 'Delivered', true],
      ['Delivered', 'Pending', false],
      ['Cancelled', 'Confirmed', false],
    ])(
      'from %s to %s should %s',
      async (fromStatus, toStatus, shouldSucceed) => {
        const orderId = 'order-1';
        mockRepo.getById.mockResolvedValue({
          id: orderId,
          status: fromStatus,
        });

        if (shouldSucceed) {
          mockRepo.update.mockResolvedValue({
            id: orderId,
            status: toStatus,
          });

          const result = await service.updateStatus(orderId, toStatus);
          expect(result.status).toBe(toStatus);
        } else {
          await expect(
            service.updateStatus(orderId, toStatus)
          ).rejects.toThrow();
        }
      }
    );
  });
});
```

### Run Tests with Coverage

```bash
# Run tests with coverage
npm run test:coverage

# Run with CI thresholds
npm run test:ci
```

---

## React Component Testing

### Project Setup (Vitest + React Testing Library)

```json
{
  "devDependencies": {
    "vitest": "^1.3.0",
    "@vitest/coverage-v8": "^1.3.0",
    "@testing-library/react": "^14.2.0",
    "@testing-library/jest-dom": "^6.4.0",
    "@testing-library/user-event": "^14.5.0",
    "jsdom": "^24.0.0",
    "msw": "^2.2.0"
  }
}
```

### vitest.config.ts (React)

```typescript
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      exclude: [
        '**/*.test.tsx',
        '**/__mocks__/**',
        '**/types/**',
        'src/test/**',
      ],
      thresholds: {
        lines: 80,
        branches: 80,
        functions: 80,
        statements: 80,
      },
    },
  },
});
```

### Test Setup (src/test/setup.ts)

```typescript
import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

afterEach(() => {
  cleanup();
});
```

### Component Test Example

```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { OrderList } from './OrderList';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

// Mock API module
vi.mock('../api/orders', () => ({
  useOrders: vi.fn(),
  useDeleteOrder: vi.fn(),
}));

import { useOrders, useDeleteOrder } from '../api/orders';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('OrderList', () => {
  it('renders loading state initially', () => {
    vi.mocked(useOrders).mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    } as any);

    render(<OrderList />, { wrapper: createWrapper() });

    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('renders orders when data is loaded', async () => {
    vi.mocked(useOrders).mockReturnValue({
      data: [
        { id: '1', customerName: 'John Doe', status: 'Pending', totalAmount: 100 },
        { id: '2', customerName: 'Jane Doe', status: 'Confirmed', totalAmount: 200 },
      ],
      isLoading: false,
      error: null,
    } as any);

    render(<OrderList />, { wrapper: createWrapper() });

    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
  });

  it('renders error state when fetch fails', () => {
    vi.mocked(useOrders).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('Failed to fetch'),
    } as any);

    render(<OrderList />, { wrapper: createWrapper() });

    expect(screen.getByText(/error/i)).toBeInTheDocument();
  });

  it('calls delete when delete button is clicked', async () => {
    const deleteMutate = vi.fn();
    vi.mocked(useOrders).mockReturnValue({
      data: [{ id: '1', customerName: 'John Doe', status: 'Pending', totalAmount: 100 }],
      isLoading: false,
      error: null,
    } as any);
    vi.mocked(useDeleteOrder).mockReturnValue({
      mutate: deleteMutate,
    } as any);

    const user = userEvent.setup();
    render(<OrderList />, { wrapper: createWrapper() });

    const deleteButton = screen.getByRole('button', { name: /delete/i });
    await user.click(deleteButton);

    // Confirm in modal
    const confirmButton = screen.getByRole('button', { name: /confirm/i });
    await user.click(confirmButton);

    await waitFor(() => {
      expect(deleteMutate).toHaveBeenCalledWith('1');
    });
  });
});
```

### Hook Test Example

```typescript
import { describe, it, expect, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useOrderCalculation } from './useOrderCalculation';

describe('useOrderCalculation', () => {
  it('calculates total correctly', () => {
    const items = [
      { quantity: 2, unitPrice: 10.0 },
      { quantity: 1, unitPrice: 5.0 },
    ];

    const { result } = renderHook(() => useOrderCalculation(items));

    expect(result.current.subtotal).toBe(25.0);
    expect(result.current.tax).toBe(2.5); // 10% tax
    expect(result.current.total).toBe(27.5);
  });

  it('returns zero for empty items', () => {
    const { result } = renderHook(() => useOrderCalculation([]));

    expect(result.current.subtotal).toBe(0);
    expect(result.current.total).toBe(0);
  });

  it('recalculates when items change', () => {
    const { result, rerender } = renderHook(
      ({ items }) => useOrderCalculation(items),
      { initialProps: { items: [{ quantity: 1, unitPrice: 10.0 }] } }
    );

    expect(result.current.subtotal).toBe(10.0);

    rerender({ items: [{ quantity: 2, unitPrice: 10.0 }] });

    expect(result.current.subtotal).toBe(20.0);
  });
});
```

### Run React Tests with Coverage

```bash
# Run tests with coverage
npm run test:coverage

# Watch mode during development
npm run test
```

---

## Test Execution Workflow

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  TEST ITERATION LOOP  ⚠️                                        ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   while (tests_fail OR coverage < 80%) {                                                 ║
║       1. Run tests with coverage                                                         ║
║       2. Read failure messages carefully                                                 ║
║       3. Fix test OR source code                                                         ║
║       4. If coverage < 80% → Add more tests                                              ║
║       5. Repeat until ALL pass AND coverage >= 80%                                       ║
║   }                                                                                      ║
║                                                                                          ║
║   ❌ DO NOT PROCEED TO DOCKER BUILD UNTIL ALL TESTS PASS!                                ║
║   ❌ DO NOT PROCEED WITH <80% COVERAGE WITHOUT EXPLICIT APPROVAL!                        ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

### Taskfile Tasks

```yaml
# Taskfile.yml
version: '3'

tasks:
  test:dotnet:
    desc: Run .NET tests with coverage
    dir: '{{.SERVICE_PATH}}'
    cmds:
      - dotnet test --collect:"XPlat Code Coverage" --results-directory ./coverage
      - |
        $coverage = [xml](Get-Content ./coverage/**/coverage.cobertura.xml)
        $lineRate = [double]$coverage.coverage.'line-rate' * 100
        Write-Host "Coverage: $lineRate%"
        if ($lineRate -lt 80) { exit 1 }

  test:go:
    desc: Run Go tests with coverage
    dir: '{{.SERVICE_PATH}}'
    cmds:
      - go test -v -coverprofile=coverage.out ./...
      - |
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
        echo "Coverage: ${COVERAGE}%"
        if (( $(echo "$COVERAGE < 80" | bc -l) )); then exit 1; fi

  test:python:
    desc: Run Python tests with coverage
    dir: '{{.SERVICE_PATH}}'
    cmds:
      - pytest --cov=src --cov-report=term-missing --cov-fail-under=80

  test:typescript:
    desc: Run TypeScript/React tests with coverage
    dir: '{{.SERVICE_PATH}}'
    cmds:
      - npm run test:ci

  test:all:
    desc: Run all tests
    cmds:
      - task: test:dotnet
      - task: test:go
      - task: test:python
      - task: test:typescript
```

---

## Phase 3 Checklist

```
PHASE 3 UNIT TESTING CHECKLIST:

□ Test project/files created with proper structure
□ All dependencies installed (xunit, testify, pytest, vitest)
□ Mocks generated for interfaces
□ Controller/Handler tests written
  □ Success cases (200, 201, 204)
  □ Not found cases (404)
  □ Validation errors (400, 422)
□ Service tests written
  □ Business logic tested
  □ Edge cases covered
  □ State transitions validated
□ Repository tests written (if applicable)
□ Coverage report generated
□ Coverage >= 80%
□ All tests passing (0 failures)
```

---

## Common Test Patterns

### AAA Pattern (Arrange-Act-Assert)

```csharp
[Fact]
public async Task MethodName_Scenario_ExpectedBehavior()
{
    // Arrange - Set up test data and mocks
    var input = CreateTestInput();
    _mockDependency.Setup(x => x.Method()).Returns(expected);

    // Act - Execute the method under test
    var result = await _sut.MethodUnderTest(input);

    // Assert - Verify the result
    result.Should().NotBeNull();
    result.Property.Should().Be(expected);
}
```

### Table-Driven Tests (Go)

```go
func TestStatusTransitions(t *testing.T) {
    tests := []struct {
        name     string
        from     string
        to       string
        wantErr  bool
    }{
        {"pending to confirmed", "Pending", "Confirmed", false},
        {"invalid transition", "Delivered", "Pending", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.UpdateStatus(tt.from, tt.to)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Parametrized Tests (Python)

```python
@pytest.mark.parametrize("input,expected", [
    (10, 100),
    (0, 0),
    (-5, 25),
])
def test_square(input, expected):
    assert square(input) == expected
```

### Test Data Factories

```typescript
// factories/orderFactory.ts
export const createOrder = (overrides?: Partial<Order>): Order => ({
  id: faker.string.uuid(),
  customerId: faker.string.uuid(),
  status: 'Pending',
  totalAmount: faker.number.float({ min: 10, max: 1000 }),
  items: [],
  createdAt: new Date().toISOString(),
  ...overrides,
});
```
