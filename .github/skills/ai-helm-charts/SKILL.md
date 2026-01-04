---
name: ai-helm-charts
description: >
  Kubernetes Helm chart patterns for AI services.
  Use when creating deployment configurations, Helm charts, or Kubernetes manifests.
  Ensures self-contained charts with official Docker images, proper health checks, and ConfigMaps.
  NEVER use Bitnami subcharts - always deploy infrastructure in-chart.
---

# AI Helm Chart Patterns

## Chart Structure

```
helm/{service-name}/
├── Chart.yaml
├── values.yaml
├── values-local.yaml          # Local development overrides
├── templates/
│   ├── _helpers.tpl
│   ├── deployment.yaml        # API deployment
│   ├── service.yaml           # ClusterIP service
│   ├── configmap.yaml         # App configuration
│   ├── hpa.yaml              # Horizontal Pod Autoscaler
│   ├── ingress.yaml          # Ingress (optional)
│   ├── serviceaccount.yaml
│   │
│   ├── # Infrastructure (self-contained)
│   ├── sqlserver.yaml         # SQL Server StatefulSet
│   ├── mongodb.yaml           # MongoDB StatefulSet  
│   ├── redis.yaml             # Redis Deployment
│   ├── kafka.yaml             # Kafka StatefulSet (KRaft)
│   └── scylladb.yaml          # ScyllaDB StatefulSet
└── seed/
    ├── init.sql               # Database seed script
    └── kafka-topics.sh        # Topic creation script
```

---

## Chart.yaml

```yaml
apiVersion: v2
name: order-service
description: AI Order Service
type: application
version: 1.0.0
appVersion: "1.0.0"

# NO dependencies on Bitnami charts!
# All infrastructure is self-contained
```

---

## values.yaml

```yaml
# API Configuration
replicaCount: 2

image:
  repository: order-service
  tag: "latest"
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 80

resources:
  requests:
    cpu: 100m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi

# Health Checks
healthCheck:
  liveness:
    path: /health/live
    port: 80
    initialDelaySeconds: 30
    periodSeconds: 10
  readiness:
    path: /health/ready
    port: 80
    initialDelaySeconds: 5
    periodSeconds: 5

# Infrastructure Settings
sqlserver:
  enabled: true
  password: "order-service-local-2024!"

mongodb:
  enabled: true

redis:
  enabled: true

kafka:
  enabled: true

# Application Configuration
config:
  connectionStrings:
    sqlServer: "Server=order-service-sqlserver,1433;Database=OrderService;User Id=sa;Password=order-service-local-2024!;TrustServerCertificate=True;"
    redis: "order-service-redis:6379,abortConnect=false"
    mongodb: "mongodb://order-service-mongodb:27017/orderservice"
  kafka:
    bootstrapServers: "order-service-kafka:9092"
```

---

## templates/deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-api
  labels:
    app.kubernetes.io/name: {{ .Release.Name }}
    app.kubernetes.io/component: api
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ .Release.Name }}
      app.kubernetes.io/component: api
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ .Release.Name }}
        app.kubernetes.io/component: api
    spec:
      containers:
        - name: api
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - containerPort: 80
              name: http
          envFrom:
            - configMapRef:
                name: {{ .Release.Name }}-config
          livenessProbe:
            httpGet:
              path: {{ .Values.healthCheck.liveness.path }}
              port: {{ .Values.healthCheck.liveness.port }}
            initialDelaySeconds: {{ .Values.healthCheck.liveness.initialDelaySeconds }}
            periodSeconds: {{ .Values.healthCheck.liveness.periodSeconds }}
          readinessProbe:
            httpGet:
              path: {{ .Values.healthCheck.readiness.path }}
              port: {{ .Values.healthCheck.readiness.port }}
            initialDelaySeconds: {{ .Values.healthCheck.readiness.initialDelaySeconds }}
            periodSeconds: {{ .Values.healthCheck.readiness.periodSeconds }}
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
```

---

## templates/configmap.yaml

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-config
data:
  # Connection Strings
  ConnectionStrings__SqlServer: {{ .Values.config.connectionStrings.sqlServer | quote }}
  ConnectionStrings__Redis: {{ .Values.config.connectionStrings.redis | quote }}
  ConnectionStrings__MongoDB: {{ .Values.config.connectionStrings.mongodb | quote }}
  
  # Kafka
  Kafka__BootstrapServers: {{ .Values.config.kafka.bootstrapServers | quote }}
  
  # Service Info
  Service__Name: {{ .Release.Name | quote }}
  Service__Environment: {{ .Values.environment | default "development" | quote }}
```

---

## templates/sqlserver.yaml

```yaml
{{- if .Values.sqlserver.enabled }}
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ .Release.Name }}-sqlserver
spec:
  serviceName: {{ .Release.Name }}-sqlserver
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Release.Name }}-sqlserver
  template:
    metadata:
      labels:
        app: {{ .Release.Name }}-sqlserver
    spec:
      containers:
        - name: sqlserver
          image: mcr.microsoft.com/mssql/server:2022-latest
          ports:
            - containerPort: 1433
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_SA_PASSWORD
              value: {{ .Values.sqlserver.password | quote }}
            - name: MSSQL_PID
              value: "Developer"
          readinessProbe:
            exec:
              command:
                - /opt/mssql-tools/bin/sqlcmd
                - -S
                - localhost
                - -U
                - sa
                - -P
                - {{ .Values.sqlserver.password | quote }}
                - -Q
                - "SELECT 1"
            initialDelaySeconds: 30
            periodSeconds: 10
          volumeMounts:
            - name: data
              mountPath: /var/opt/mssql
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}-sqlserver
spec:
  selector:
    app: {{ .Release.Name }}-sqlserver
  ports:
    - port: 1433
      targetPort: 1433
  clusterIP: None
{{- end }}
```

---

## templates/redis.yaml

```yaml
{{- if .Values.redis.enabled }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Release.Name }}-redis
  template:
    metadata:
      labels:
        app: {{ .Release.Name }}-redis
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          ports:
            - containerPort: 6379
          command: ["redis-server", "--appendonly", "yes"]
          readinessProbe:
            exec:
              command: ["redis-cli", "ping"]
            initialDelaySeconds: 5
            periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}-redis
spec:
  selector:
    app: {{ .Release.Name }}-redis
  ports:
    - port: 6379
      targetPort: 6379
{{- end }}
```

---

## templates/kafka.yaml (KRaft Mode - NO Zookeeper)

```yaml
{{- if .Values.kafka.enabled }}
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
    metadata:
      labels:
        app: {{ .Release.Name }}-kafka
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
            - name: KAFKA_LOG_DIRS
              value: "/tmp/kraft-combined-logs"
            - name: CLUSTER_ID
              value: "MkU3OEVBNTcwNTJENDM2Qk"
          readinessProbe:
            exec:
              command:
                - /opt/kafka/bin/kafka-topics.sh
                - --bootstrap-server
                - localhost:9092
                - --list
            initialDelaySeconds: 30
            periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}-kafka
spec:
  selector:
    app: {{ .Release.Name }}-kafka
  ports:
    - port: 9092
      targetPort: 9092
  clusterIP: None
{{- end }}
```

---

## templates/mongodb.yaml

```yaml
{{- if .Values.mongodb.enabled }}
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ .Release.Name }}-mongodb
spec:
  serviceName: {{ .Release.Name }}-mongodb
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Release.Name }}-mongodb
  template:
    metadata:
      labels:
        app: {{ .Release.Name }}-mongodb
    spec:
      containers:
        - name: mongodb
          image: mongo:7
          ports:
            - containerPort: 27017
          readinessProbe:
            exec:
              command: ["mongosh", "--eval", "db.adminCommand('ping')"]
            initialDelaySeconds: 10
            periodSeconds: 10
          volumeMounts:
            - name: data
              mountPath: /data/db
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}-mongodb
spec:
  selector:
    app: {{ .Release.Name }}-mongodb
  ports:
    - port: 27017
      targetPort: 27017
  clusterIP: None
{{- end }}
```

---

## Deployment Commands

### Deploy to Local K8s

```powershell
# Create namespace
kubectl create namespace sandbox --dry-run=client -o yaml | kubectl apply -f -

# Deploy with Helm
helm upgrade --install order-service ./helm/order-service -f values-local.yaml -n sandbox

# Wait for pods
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=order-service -n sandbox --timeout=180s
```

### Port Forward

```powershell
# API
kubectl port-forward svc/order-service-api 8080:80 -n sandbox

# UI (if exists)
kubectl port-forward svc/order-service-ui 3000:80 -n sandbox
```

---

## ⚠️ CRITICAL: NO Bitnami

**NEVER use Bitnami subcharts:**

```yaml
# ❌ WRONG - Bitnami dependency
dependencies:
  - name: redis
    version: 17.x.x
    repository: https://charts.bitnami.com/bitnami

# ❌ WRONG - Bitnami image
image: bitnami/redis:7.0
```

**Always use official images with self-contained templates:**

```yaml
# ✅ CORRECT - Official image in own template
image: redis:7-alpine
image: apache/kafka:3.7.0
image: mcr.microsoft.com/mssql/server:2022-latest
image: mongo:7
```
