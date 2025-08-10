# Deployment Guide

## Overview

This guide covers deployment strategies for the Core Banking Lab, from local development to production-ready infrastructure with Docker, Kubernetes, and cloud platforms.

## Local Development

### **Prerequisites**

```bash
# Required software
Go 1.23+
Docker & Docker Compose
Git
Make (optional)

# Verify installation
go version
docker --version
docker-compose --version
```

### **Quick Start**

```bash
# Clone repository
git clone https://github.com/yourusername/core-banking-lab.git
cd core-banking-lab

# Run API server
go run src/main.go

# Or with Docker Compose (full stack)
docker-compose up --build
```

### **Development Configuration**

```bash
# .env.development
export SERVER_PORT=8080
export SERVER_HOST=localhost
export LOG_LEVEL=debug
export LOG_FORMAT=text
export CORS_ALLOWED_ORIGINS="http://localhost:3000,http://localhost:5173"
export RATE_LIMIT_REQUESTS_PER_MINUTE=1000
```

## Docker Deployment

### **Single Container (API Only)**

**Dockerfile.api**:
```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY src/ src/
RUN CGO_ENABLED=0 GOOS=linux go build -o bank-api src/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/bank-api .
EXPOSE 8080

CMD ["./bank-api"]
```

**Build & Run**:
```bash
# Build image
docker build -f Dockerfile.api -t core-banking-api .

# Run container
docker run -p 8080:8080 \
  -e CORS_ALLOWED_ORIGINS="http://localhost:3000" \
  -e LOG_LEVEL=info \
  core-banking-api
```

### **Multi-Container Stack**

**docker-compose.yml**:
```yaml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile.api
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=info
      - LOG_FORMAT=json
      - CORS_ALLOWED_ORIGINS=http://localhost:5173
      - RATE_LIMIT_REQUESTS_PER_MINUTE=100
    networks:
      - banking-network
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  dashboard:
    build:
      context: ./dev/dashboard
      dockerfile: Dockerfile
    ports:
      - "5173:5173"
    depends_on:
      - api
    networks:
      - banking-network

  simulator:
    build:
      context: ./cmd/dashboard
      dockerfile: Dockerfile
    depends_on:
      - api
    networks:
      - banking-network
    environment:
      - API_URL=http://api:8080

networks:
  banking-network:
    driver: bridge
```

**Deploy Stack**:
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Scale API instances
docker-compose up -d --scale api=3

# Stop stack
docker-compose down
```

## Kubernetes Deployment

### **Namespace & ConfigMap**

**k8s/namespace.yaml**:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: banking-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: banking-config
  namespace: banking-system
data:
  LOG_LEVEL: "info"
  LOG_FORMAT: "json"
  SERVER_PORT: "8080"
  CORS_ALLOWED_ORIGINS: "https://banking-dashboard.example.com"
  RATE_LIMIT_REQUESTS_PER_MINUTE: "100"
```

### **API Deployment**

**k8s/api-deployment.yaml**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: banking-api
  namespace: banking-system
  labels:
    app: banking-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: banking-api
  template:
    metadata:
      labels:
        app: banking-api
    spec:
      containers:
      - name: banking-api
        image: core-banking-api:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: banking-config
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: banking-api-service
  namespace: banking-system
spec:
  selector:
    app: banking-api
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### **Ingress Configuration**

**k8s/ingress.yaml**:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: banking-ingress
  namespace: banking-system
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.banking.example.com
    secretName: banking-tls
  rules:
  - host: api.banking.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: banking-api-service
            port:
              number: 80
```

### **Deploy to Kubernetes**

```bash
# Apply all configurations
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n banking-system
kubectl get services -n banking-system

# View logs
kubectl logs -f deployment/banking-api -n banking-system

# Scale deployment
kubectl scale deployment banking-api --replicas=5 -n banking-system

# Update deployment (rolling update)
kubectl set image deployment/banking-api banking-api=core-banking-api:v2.0 -n banking-system
```

## Production Configuration

### **Environment Variables**

**Production Security Settings**:
```bash
# Security
export CORS_ALLOWED_ORIGINS="https://secure-banking.example.com"
export RATE_LIMIT_REQUESTS_PER_MINUTE=50
export LOG_LEVEL=warn
export LOG_FORMAT=json

# Performance
export GOMAXPROCS=4
export GOGC=100

# Monitoring
export ENABLE_METRICS=true
export METRICS_PORT=9090
```

### **Secrets Management**

**Kubernetes Secrets**:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: banking-secrets
  namespace: banking-system
type: Opaque
data:
  jwt-secret: <base64-encoded-secret>
  database-password: <base64-encoded-password>
```

**Usage in Deployment**:
```yaml
env:
- name: JWT_SECRET
  valueFrom:
    secretKeyRef:
      name: banking-secrets
      key: jwt-secret
```

### **Resource Requirements**

**Production Sizing**:
```yaml
resources:
  requests:
    cpu: 500m      # 0.5 CPU core
    memory: 512Mi  # 512 MB RAM
  limits:
    cpu: 2000m     # 2 CPU cores  
    memory: 2Gi    # 2 GB RAM
```

**Horizontal Pod Autoscaler**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: banking-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: banking-api
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Database Integration

### **PostgreSQL Setup**

**docker-compose.prod.yml**:
```yaml
version: '3.8'

services:
  api:
    build: 
      context: .
      dockerfile: Dockerfile.api
    environment:
      - DATABASE_URL=postgres://banking:password@postgres:5432/banking_db?sslmode=disable
    depends_on:
      - postgres

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=banking_db
      - POSTGRES_USER=banking
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./db/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

### **Database Migrations**

**db/migrations/001_initial_schema.sql**:
```sql
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    owner VARCHAR(100) NOT NULL,
    balance INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    from_account_id INTEGER REFERENCES accounts(id),
    to_account_id INTEGER REFERENCES accounts(id),
    amount INTEGER NOT NULL,
    transaction_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accounts_owner ON accounts(owner);
CREATE INDEX idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX idx_transactions_to_account ON transactions(to_account_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
```

## Monitoring & Observability

### **Prometheus Configuration**

**prometheus.yml**:
```yaml
global:
  scrape_interval: 15s

scrape_configs:
- job_name: 'banking-api'
  static_configs:
  - targets: ['banking-api-service:9090']
  scrape_interval: 5s
  metrics_path: /metrics
```

### **Grafana Dashboard**

**Key Metrics to Monitor**:
- Request rate (requests/second)
- Response time percentiles (P50, P95, P99)
- Error rate percentage
- Active connections
- CPU and memory utilization
- Database connection pool status

### **Logging Pipeline**

**Filebeat Configuration**:
```yaml
filebeat.inputs:
- type: container
  paths:
    - '/var/lib/docker/containers/*/*.log'
  processors:
  - add_kubernetes_metadata: ~

output.elasticsearch:
  hosts: ["elasticsearch:9200"]

setup.kibana:
  host: "kibana:5601"
```

## Load Balancing

### **NGINX Configuration**

**nginx.conf**:
```nginx
upstream banking_api {
    least_conn;
    server api-1:8080 max_fails=3 fail_timeout=30s;
    server api-2:8080 max_fails=3 fail_timeout=30s;
    server api-3:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name api.banking.example.com;
    
    location / {
        proxy_pass http://banking_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        
        # Rate limiting
        limit_req zone=api_limit burst=20 nodelay;
    }
}

# Rate limiting configuration
http {
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
}
```

## CI/CD Pipeline

### **GitHub Actions**

**.github/workflows/deploy.yml**:
```yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.23
    
    - name: Run tests
      run: |
        go test -race ./...
        go test -cover ./...

  build-and-deploy:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Build Docker image
      run: |
        docker build -f Dockerfile.api -t ${{ secrets.REGISTRY }}/banking-api:${{ github.sha }} .
        docker push ${{ secrets.REGISTRY }}/banking-api:${{ github.sha }}
    
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/banking-api banking-api=${{ secrets.REGISTRY }}/banking-api:${{ github.sha }} -n banking-system
        kubectl rollout status deployment/banking-api -n banking-system
```

## Security Hardening

### **Production Security Checklist**

- ✅ TLS/SSL encryption (HTTPS only)
- ✅ Rate limiting configuration
- ✅ CORS policy restriction
- ✅ Input validation and sanitization
- ✅ Secrets management (not in environment variables)
- ✅ Network policies (Kubernetes)
- ✅ Container security scanning
- ✅ Regular security updates
- ✅ Audit logging enabled
- ✅ Monitoring and alerting configured

### **Network Security**

**Kubernetes Network Policy**:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: banking-api-netpol
  namespace: banking-system
spec:
  podSelector:
    matchLabels:
      app: banking-api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: nginx-ingress
    ports:
    - protocol: TCP
      port: 8080
```

## Disaster Recovery

### **Backup Strategy**

**Database Backups**:
```bash
# Automated daily backups
kubectl create cronjob postgres-backup \
  --image=postgres:15 \
  --schedule="0 2 * * *" \
  -- pg_dump -h postgres-service -U banking banking_db > /backups/backup-$(date +%Y%m%d).sql
```

**Application State**:
- Stateless application design (no local state to backup)
- Configuration stored in version control
- Container images stored in registry with tags

### **Recovery Procedures**

1. **Database Recovery**: Restore from backup
2. **Application Recovery**: Deploy from container registry
3. **Configuration Recovery**: Apply Kubernetes manifests
4. **Validation**: Run health checks and integration tests

This deployment guide provides a comprehensive path from development to production-ready infrastructure with proper security, monitoring, and operational practices.