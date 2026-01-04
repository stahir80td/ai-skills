# Infrastructure Standards

> **Kubernetes, Helm, Docker, and deployment patterns for generated services**

This document defines infrastructure patterns for generated applications. All services follow these standards for consistent deployment.

---

## Container Strategy

### .dockerignore (Required for ALL projects)

Every generated project MUST include a `.dockerignore` file:

```
# Build artifacts
bin/
obj/
node_modules/
dist/
__pycache__/
*.pyc
.pytest_cache/

# IDE and editor
.git/
.gitignore
.vscode/
.idea/
*.swp
*.swo

# Documentation
*.md
docs/

# Environment and secrets
.env*
*.local

# Logs
*.log
logs/

# OS files
.DS_Store
Thumbs.db

# Test files
**/tests/
**/*_test.go
**/*_test.py
```

### Dockerfile Pattern (Go)

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build with optimizations
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o /app/server ./cmd/server

# Runtime stage
FROM scratch

# Copy timezone data and certs
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /app/server /server

# Non-root user
USER 65534

EXPOSE 8080

ENTRYPOINT ["/server"]
```

### Dockerfile Pattern (Python)

```dockerfile
# Build stage
FROM python:3.11-slim AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Create virtual environment
RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Runtime stage
FROM python:3.11-slim

WORKDIR /app

# Copy virtual environment
COPY --from=builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Copy application
COPY . .

# Non-root user
RUN useradd -r -u 1001 appuser
USER appuser

EXPOSE 8080

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
```

### Dockerfile Pattern (React/Vite)

```dockerfile
# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm install

# Build application
COPY . .
RUN npm run build

# Runtime stage
FROM nginx:alpine

# Copy custom nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Non-root user
RUN chown -R nginx:nginx /usr/share/nginx/html
USER nginx

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

**Note:** Use `npm install` instead of `npm ci` since generated projects won't have `package-lock.json`.

### nginx.conf for React SPA

```nginx
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /tmp/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    
    # Gzip compression
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
    gzip_min_length 1000;
    
    server {
        listen 80;
        server_name _;
        root /usr/share/nginx/html;
        index index.html;
        
        # SPA routing - fallback to index.html
        location / {
            try_files $uri $uri/ /index.html;
        }
        
        # API proxy (if needed)
        location /api/ {
            proxy_pass http://api-service:8080/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
        
        # Cache static assets
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
        
        # Health check
        location /health {
            return 200 'ok';
            add_header Content-Type text/plain;
        }
    }
}
```

---

## Helm Chart Structure

### Directory Layout

```
helm/{service-name}/
├── Chart.yaml
├── values.yaml
├── values-dev.yaml
├── values-prod.yaml
├── templates/
│   ├── _helpers.tpl
│   ├── configmap.yaml
│   ├── deployment.yaml
│   ├── hpa.yaml
│   ├── ingress.yaml
│   ├── pdb.yaml
│   ├── secret.yaml
│   ├── service.yaml
│   └── serviceaccount.yaml
└── charts/               # Dependencies
```

### Chart.yaml

```yaml
apiVersion: v2
name: order-service
description: Order management service
type: application
version: 0.1.0
appVersion: "1.0.0"

dependencies:
  - name: redis
    version: "17.x.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: redis.enabled
```

### values.yaml (Base)

```yaml
# Application
replicaCount: 2

image:
  repository: myacr.azurecr.io/order-service
  pullPolicy: IfNotPresent
  tag: ""  # Defaults to appVersion

imagePullSecrets:
  - name: acr-secret

# Service Account
serviceAccount:
  create: true
  annotations: {}
  name: ""

# Pod Security
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65534
  fsGroup: 65534

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL

# Service
service:
  type: ClusterIP
  port: 80
  targetPort: 8080

# Ingress
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: api.example.com
      paths:
        - path: /api/v1/orders
          pathType: Prefix
  tls:
    - secretName: api-tls
      hosts:
        - api.example.com

# Resources
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

# Autoscaling
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

# Pod Disruption Budget
podDisruptionBudget:
  enabled: true
  minAvailable: 1

# Probes
livenessProbe:
  httpGet:
    path: /health/live
    port: http
  initialDelaySeconds: 15
  periodSeconds: 20

readinessProbe:
  httpGet:
    path: /health/ready
    port: http
  initialDelaySeconds: 5
  periodSeconds: 10

# Environment Configuration
config:
  LOG_LEVEL: info
  HTTP_PORT: "8080"
  
# Secrets (reference Key Vault)
secrets:
  DATABASE_URL: ""
  REDIS_URL: ""
  KAFKA_BOOTSTRAP_SERVERS: ""

# External secrets (Azure Key Vault)
externalSecrets:
  enabled: true
  secretStoreRef:
    name: azure-keyvault
    kind: ClusterSecretStore
  target:
    name: order-service-secrets
  data:
    - secretKey: DATABASE_URL
      remoteRef:
        key: order-service-db-url
    - secretKey: REDIS_URL
      remoteRef:
        key: redis-url

# Redis subchart
redis:
  enabled: false  # Use external Redis
```

### values-dev.yaml

```yaml
replicaCount: 1

autoscaling:
  enabled: false

resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 256Mi

config:
  LOG_LEVEL: debug

ingress:
  hosts:
    - host: api.dev.example.com
      paths:
        - path: /api/v1/orders
          pathType: Prefix
```

### values-prod.yaml

```yaml
replicaCount: 3

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi

podDisruptionBudget:
  minAvailable: 2

config:
  LOG_LEVEL: warn
```

---

## Helm Templates

### deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "chart.fullname" . }}
  labels:
    {{- include "chart.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "chart.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      annotations:
        checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
      labels:
        {{- include "chart.selectorLabels" . | nindent 8 }}
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "chart.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.podSecurityContext | nindent 8 }}
      containers:
        - name: {{ .Chart.Name }}
          securityContext:
            {{- toYaml .Values.securityContext | nindent 12 }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ .Values.service.targetPort }}
              protocol: TCP
          envFrom:
            - configMapRef:
                name: {{ include "chart.fullname" . }}-config
            - secretRef:
                name: {{ include "chart.fullname" . }}-secrets
          livenessProbe:
            {{- toYaml .Values.livenessProbe | nindent 12 }}
          readinessProbe:
            {{- toYaml .Values.readinessProbe | nindent 12 }}
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
```

### hpa.yaml

```yaml
{{- if .Values.autoscaling.enabled }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "chart.fullname" . }}
  labels:
    {{- include "chart.labels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "chart.fullname" . }}
  minReplicas: {{ .Values.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.autoscaling.maxReplicas }}
  metrics:
    {{- if .Values.autoscaling.targetCPUUtilizationPercentage }}
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ .Values.autoscaling.targetCPUUtilizationPercentage }}
    {{- end }}
    {{- if .Values.autoscaling.targetMemoryUtilizationPercentage }}
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: {{ .Values.autoscaling.targetMemoryUtilizationPercentage }}
    {{- end }}
{{- end }}
```

### pdb.yaml

```yaml
{{- if .Values.podDisruptionBudget.enabled }}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "chart.fullname" . }}
  labels:
    {{- include "chart.labels" . | nindent 4 }}
spec:
  minAvailable: {{ .Values.podDisruptionBudget.minAvailable }}
  selector:
    matchLabels:
      {{- include "chart.selectorLabels" . | nindent 6 }}
{{- end }}
```

---

## Taskfile.yml

### Standard Task Definitions

```yaml
version: '3'

vars:
  SERVICE_NAME: order-service
  VERSION:
    sh: cat VERSION
  REGISTRY: myacr.azurecr.io
  IMAGE: '{{.REGISTRY}}/{{.SERVICE_NAME}}'

env:
  CGO_ENABLED: 0
  GOOS: linux
  GOARCH: amd64

tasks:
  # ==================== Development ====================
  
  dev:
    desc: Run service locally with hot reload
    cmds:
      - air -c .air.toml

  test:
    desc: Run unit tests
    cmds:
      - go test -v -race -coverprofile=coverage.out ./...

  test:coverage:
    desc: Run tests with coverage report
    cmds:
      - go test -v -race -coverprofile=coverage.out ./...
      - go tool cover -html=coverage.out -o coverage.html

  lint:
    desc: Run linter
    cmds:
      - golangci-lint run ./...

  fmt:
    desc: Format code
    cmds:
      - go fmt ./...
      - goimports -w .

  # ==================== Build ====================
  
  build:
    desc: Build binary
    cmds:
      - go build -ldflags="-w -s -X main.version={{.VERSION}}" -o bin/server ./cmd/server

  docker:build:
    desc: Build Docker image
    cmds:
      - docker build -t {{.IMAGE}}:{{.VERSION}} -t {{.IMAGE}}:latest .

  docker:push:
    desc: Push Docker image to registry
    deps: [docker:build]
    cmds:
      - docker push {{.IMAGE}}:{{.VERSION}}
      - docker push {{.IMAGE}}:latest

  # ==================== Deploy ====================
  
  deploy:dev:
    desc: Deploy to development
    cmds:
      - helm upgrade --install {{.SERVICE_NAME}} ./helm/{{.SERVICE_NAME}} 
        -f ./helm/{{.SERVICE_NAME}}/values-dev.yaml
        --set image.tag={{.VERSION}}
        -n dev

  deploy:prod:
    desc: Deploy to production
    cmds:
      - helm upgrade --install {{.SERVICE_NAME}} ./helm/{{.SERVICE_NAME}} 
        -f ./helm/{{.SERVICE_NAME}}/values-prod.yaml
        --set image.tag={{.VERSION}}
        -n prod

  # ==================== All-in-One ====================
  
  doall:
    desc: Lint, test, build, push, and deploy
    cmds:
      - task: lint
      - task: test
      - task: docker:push
      - task: deploy:dev

  # ==================== Database ====================
  
  db:migrate:
    desc: Run database migrations
    cmds:
      - go run ./cmd/migrate up

  db:seed:
    desc: Seed database with sample data
    cmds:
      - go run ./cmd/seed

  # ==================== Code Generation ====================
  
  generate:
    desc: Generate code (mocks, swagger, etc.)
    cmds:
      - go generate ./...
      - swag init -g cmd/server/main.go

  # ==================== Utilities ====================
  
  clean:
    desc: Clean build artifacts
    cmds:
      - rm -rf bin/ coverage.out coverage.html

  version:
    desc: Show current version
    cmds:
      - echo {{.VERSION}}

  version:bump:
    desc: Bump patch version
    cmds:
      - |
        current=$(cat VERSION)
        IFS='.' read -r major minor patch <<< "$current"
        patch=$((patch + 1))
        echo "$major.$minor.$patch" > VERSION
        echo "Bumped to $(cat VERSION)"
```

---

## Azure Infrastructure

### Key Vault Integration

```yaml
# External Secrets Operator - SecretStore
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: azure-keyvault
spec:
  provider:
    azurekv:
      authType: WorkloadIdentity
      vaultUrl: https://mykeyvault.vault.azure.net
      serviceAccountRef:
        name: external-secrets-sa
        namespace: external-secrets
```

### Azure AD Pod Identity (Workload Identity)

```yaml
# ServiceAccount with workload identity
apiVersion: v1
kind: ServiceAccount
metadata:
  name: order-service
  annotations:
    azure.workload.identity/client-id: <managed-identity-client-id>
    azure.workload.identity/tenant-id: <tenant-id>
```

---

## Monitoring Setup

### ServiceMonitor (Prometheus)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: {{ include "chart.fullname" . }}
  labels:
    {{- include "chart.labels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "chart.selectorLabels" . | nindent 6 }}
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
```

### Grafana Dashboard ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "chart.fullname" . }}-dashboard
  labels:
    grafana_dashboard: "1"
data:
  dashboard.json: |
    {
      "title": "{{ .Chart.Name }}",
      "panels": [
        {
          "title": "Request Rate",
          "type": "graph",
          "targets": [
            {
              "expr": "rate(http_requests_total{service=\"{{ include "chart.fullname" . }}\"}[5m])"
            }
          ]
        }
      ]
    }
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/deploy.yml
name: Build and Deploy

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  REGISTRY: myacr.azurecr.io
  IMAGE_NAME: order-service

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -v -race ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'push'
    steps:
      - uses: actions/checkout@v4
      
      - name: Login to ACR
        uses: azure/docker-login@v1
        with:
          login-server: ${{ env.REGISTRY }}
          username: ${{ secrets.ACR_USERNAME }}
          password: ${{ secrets.ACR_PASSWORD }}
      
      - name: Build and push
        run: |
          VERSION=$(cat VERSION)
          docker build -t $REGISTRY/$IMAGE_NAME:$VERSION .
          docker push $REGISTRY/$IMAGE_NAME:$VERSION

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      
      - name: Azure Login
        uses: azure/login@v1
        with:
          creds: ${{ secrets.AZURE_CREDENTIALS }}
      
      - name: Set AKS context
        uses: azure/aks-set-context@v3
        with:
          cluster-name: my-aks-cluster
          resource-group: my-resource-group
      
      - name: Deploy with Helm
        run: |
          VERSION=$(cat VERSION)
          helm upgrade --install order-service ./helm/order-service \
            -f ./helm/order-service/values-prod.yaml \
            --set image.tag=$VERSION \
            -n prod
```

---

## Directory Structure (Full Service)

```
{service-name}/
├── .github/
│   └── workflows/
│       └── deploy.yml
├── cmd/
│   ├── server/
│   │   └── main.go
│   ├── migrate/
│   │   └── main.go
│   └── seed/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── routes.go
│   ├── domain/
│   │   ├── models/
│   │   └── services/
│   └── infrastructure/
│       ├── database/
│       ├── kafka/
│       └── redis/
├── helm/
│   └── {service-name}/
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── values-dev.yaml
│       ├── values-prod.yaml
│       └── templates/
├── migrations/
│   ├── 001_initial.up.sql
│   └── 001_initial.down.sql
├── .air.toml
├── .gitignore
├── .golangci.yml
├── Dockerfile
├── go.mod
├── go.sum
├── Taskfile.yml
├── VERSION
└── README.md
```
