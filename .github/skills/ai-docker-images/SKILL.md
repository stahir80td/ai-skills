---
name: ai-docker-images
description: >
  Official Docker images for AI infrastructure components.
  Use when creating Dockerfile, docker-compose.yml, or Helm charts.
  CRITICAL: Only use official images, NO Bitnami (causes pull issues), NO Confluent (licensing).
  Covers Kafka KRaft mode setup without Zookeeper.
---

# AI Docker Images - Official Only

## ⚠️ CRITICAL RULES

### ✅ APPROVED Images ONLY

| Component | Official Image | Version |
|-----------|---------------|---------|
| SQL Server | `mcr.microsoft.com/mssql/server` | `2022-latest` |
| MongoDB | `mongo` | `7` |
| ScyllaDB | `scylladb/scylla` | `5.4` |
| Redis | `redis` | `7-alpine` |
| Kafka | `apache/kafka` | `3.7.0` |
| Go Builder | `golang` | `1.24-alpine` |
| .NET SDK | `mcr.microsoft.com/dotnet/sdk` | `8.0` |
| .NET Runtime | `mcr.microsoft.com/dotnet/aspnet` | `8.0` |
| Alpine | `alpine` | `3.19` |

### ❌ FORBIDDEN Images

| ❌ Never Use | Reason |
|-------------|--------|
| `bitnami/*` | Image pull failures, complex config |
| `wurstmeister/kafka` | Outdated, requires Zookeeper |
| `confluentinc/*` | Enterprise licensing issues |
| `bitnami/mongodb` | Use official `mongo` instead |
| `bitnami/redis` | Use official `redis` instead |
| `zookeeper` | Kafka KRaft mode doesn't need it |

---

## Kafka (Apache KRaft Mode - NO Zookeeper)

### docker-compose.yml

```yaml
kafka:
  image: apache/kafka:3.7.0
  ports:
    - "9092:9092"
  environment:
    - KAFKA_NODE_ID=1
    - KAFKA_PROCESS_ROLES=broker,controller
    - KAFKA_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093
    - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka:9092
    - KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER
    - KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
    - KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093
    - KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1
    - KAFKA_LOG_DIRS=/tmp/kraft-combined-logs
    - CLUSTER_ID=MkU3OEVBNTcwNTJENDM2Qk
  healthcheck:
    test: /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list || exit 1
    interval: 10s
    timeout: 5s
    retries: 10
    start_period: 30s
```

### Helm StatefulSet

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ .Release.Name }}-kafka
spec:
  serviceName: {{ .Release.Name }}-kafka
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Release.Name }}-kafka
  template:
    spec:
      containers:
        - name: kafka
          image: apache/kafka:3.7.0
          ports:
            - containerPort: 9092
              name: kafka
          env:
            - name: KAFKA_NODE_ID
              value: "1"
            - name: KAFKA_PROCESS_ROLES
              value: "broker,controller"
            - name: KAFKA_LISTENERS
              value: "PLAINTEXT://:9092,CONTROLLER://:9093"
            - name: KAFKA_ADVERTISED_LISTENERS
              value: "PLAINTEXT://{{ .Release.Name }}-kafka:9092"
            - name: KAFKA_CONTROLLER_LISTENER_NAMES
              value: "CONTROLLER"
            - name: KAFKA_LISTENER_SECURITY_PROTOCOL_MAP
              value: "CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT"
            - name: KAFKA_CONTROLLER_QUORUM_VOTERS
              value: "1@localhost:9093"
            - name: KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR
              value: "1"
            - name: CLUSTER_ID
              value: "MkU3OEVBNTcwNTJENDM2Qk"
```

---

## SQL Server

### docker-compose.yml

```yaml
sqlserver:
  image: mcr.microsoft.com/mssql/server:2022-latest
  ports:
    - "1433:1433"
  environment:
    - ACCEPT_EULA=Y
    - MSSQL_SA_PASSWORD=YourStrong!Passw0rd
    - MSSQL_PID=Developer
  healthcheck:
    test: /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "$$MSSQL_SA_PASSWORD" -Q "SELECT 1" || exit 1
    interval: 10s
    timeout: 5s
    retries: 10
```

---

## MongoDB

### docker-compose.yml

```yaml
mongodb:
  image: mongo:7
  ports:
    - "27017:27017"
  environment:
    - MONGO_INITDB_ROOT_USERNAME=admin
    - MONGO_INITDB_ROOT_PASSWORD=password
  volumes:
    - mongodb_data:/data/db
```

### For Atlas Local (K8s)

```yaml
# Use mongodb-atlas-local for local K8s deployments
mongodb:
  image: mongodb/mongodb-atlas-local:8.0.3
  ports:
    - "27017:27017"
```

---

## Redis

### docker-compose.yml

```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
  command: redis-server --appendonly yes
  healthcheck:
    test: ["CMD", "redis-cli", "ping"]
    interval: 10s
    timeout: 5s
    retries: 5
```

---

## ScyllaDB

### docker-compose.yml

```yaml
scylladb:
  image: scylladb/scylla:5.4
  ports:
    - "9042:9042"
  command: --smp 1 --memory 750M --overprovisioned 1
  healthcheck:
    test: ["CMD", "cqlsh", "-e", "describe cluster"]
    interval: 30s
    timeout: 10s
    retries: 10
    start_period: 60s
```

---

## Complete docker-compose.infra.yml

```yaml
version: '3.8'

services:
  sqlserver:
    image: mcr.microsoft.com/mssql/server:2022-latest
    ports:
      - "1433:1433"
    environment:
      - ACCEPT_EULA=Y
      - MSSQL_SA_PASSWORD=${SERVICE_NAME:-service}-local-2024!
      - MSSQL_PID=Developer
    healthcheck:
      test: /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "$$MSSQL_SA_PASSWORD" -Q "SELECT 1"
      interval: 10s
      timeout: 5s
      retries: 10

  mongodb:
    image: mongo:7
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes

  kafka:
    image: apache/kafka:3.7.0
    ports:
      - "9092:9092"
    environment:
      - KAFKA_NODE_ID=1
      - KAFKA_PROCESS_ROLES=broker,controller
      - KAFKA_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093
      - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka:9092
      - KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER
      - KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      - KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093
      - KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1
      - CLUSTER_ID=MkU3OEVBNTcwNTJENDM2Qk

  scylladb:
    image: scylladb/scylla:5.4
    ports:
      - "9042:9042"
    command: --smp 1 --memory 750M --overprovisioned 1

volumes:
  mongodb_data:
```

---

## .NET Dockerfile

```dockerfile
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
ARG GITHUB_TOKEN
ENV GITHUB_TOKEN=$GITHUB_TOKEN
WORKDIR /src
COPY NuGet.config .
COPY *.sln .
COPY src/ src/
RUN dotnet restore
RUN dotnet publish -c Release -o /app

FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime
WORKDIR /app
COPY --from=build /app .
EXPOSE 80
ENTRYPOINT ["dotnet", "MyService.Api.dll"]
```

---

## Go Dockerfile

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./cmd/service

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /service .
COPY config/config.yaml ./config/
EXPOSE 8080
CMD ["./service"]
```
