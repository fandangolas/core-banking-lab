# Deployment Guide

**Multiple deployment options from local development to production-ready infrastructure.**

## Quick Start

### Local Development
```bash
# Clone and run
git clone https://github.com/fandangolas/core-banking-lab.git
cd core-banking-lab
go run src/main.go

# API available at http://localhost:8080
```

### Docker (Recommended)
```bash
# Full stack with dashboard
docker-compose up --build

# Services:
# - API: http://localhost:8080
# - Dashboard: http://localhost:5173
# - Load simulator: Automatic background transactions
```

### API Only Container
```bash
# Build and run API container
docker build -f Dockerfile.api -t banking-api .
docker run -p 8080:8080 \
  -e CORS_ALLOWED_ORIGINS="http://localhost:3000" \
  -e LOG_LEVEL=info \
  banking-api
```

## Production Configuration

### Environment Variables

**Security Settings:**
```bash
export CORS_ALLOWED_ORIGINS="https://yourdomain.com"
export RATE_LIMIT_REQUESTS_PER_MINUTE=50
export LOG_LEVEL=warn
export LOG_FORMAT=json
```

**Server Settings:**
```bash
export SERVER_PORT=8080
export SERVER_HOST=0.0.0.0
```

### Docker Compose Production

**docker-compose.prod.yml** example:
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
      - CORS_ALLOWED_ORIGINS=https://yourdomain.com
      - RATE_LIMIT_REQUESTS_PER_MINUTE=100
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--spider", "http://localhost:8080/health"]
      interval: 30s
      retries: 3

  dashboard:
    build: ./dev/dashboard
    ports:
      - "5173:5173"
    depends_on:
      - api
    restart: unless-stopped
```

## Kubernetes Deployment

### Basic Kubernetes Setup
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: banking-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: banking-api
  template:
    spec:
      containers:
      - name: banking-api
        image: banking-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: RATE_LIMIT_REQUESTS_PER_MINUTE
          value: "100"
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
---
apiVersion: v1
kind: Service
metadata:
  name: banking-api-service
spec:
  selector:
    app: banking-api
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### Deploy Commands
```bash
# Apply configuration
kubectl apply -f k8s-deployment.yaml

# Scale deployment
kubectl scale deployment banking-api --replicas=5

# Check status
kubectl get pods
kubectl logs -f deployment/banking-api
```

## Database Integration

### PostgreSQL with Docker
```yaml
# Add to docker-compose.yml
postgres:
  image: postgres:15-alpine
  environment:
    - POSTGRES_DB=banking_db
    - POSTGRES_USER=banking
    - POSTGRES_PASSWORD=password
  volumes:
    - postgres_data:/var/lib/postgresql/data
  ports:
    - "5432:5432"

# Update API environment
api:
  environment:
    - DATABASE_URL=postgres://banking:password@postgres:5432/banking_db
```

## Monitoring & Health Checks

### Built-in Health Endpoint
```bash
curl http://localhost:8080/health

# Response:
{
  "status": "healthy",
  "timestamp": "2025-08-10T02:54:47Z"
}
```

### Metrics Collection
```bash
curl http://localhost:8080/metrics

# Returns:
{
  "endpoints": {
    "POST /accounts/transfer": {"count": 445, "avg_duration_ms": 1.2}
  },
  "system": {
    "uptime_seconds": 3600,
    "goroutines": 15
  }
}
```

## Production Considerations

### **Security Hardening**
- Configure strict CORS origins (no wildcards)
- Set appropriate rate limits for your traffic
- Use JSON logging for structured log analysis  
- Keep error messages generic (no sensitive data)

### **Performance Tuning**
- Set `GOMAXPROCS` based on available CPU cores
- Configure appropriate resource limits in containers
- Use health checks for proper load balancing
- Monitor goroutine count and memory usage

### **Scalability**
- Stateless design enables horizontal scaling
- Each API instance can handle concurrent requests safely
- Database will be the primary bottleneck (plan accordingly)
- Consider load balancing for multiple instances

### **Container Optimization**
```dockerfile
# Multi-stage build for smaller images
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY src/ src/
RUN CGO_ENABLED=0 GOOS=linux go build -o bank-api src/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bank-api .
EXPOSE 8080
CMD ["./bank-api"]
```

## Testing Deployment

### Validate API Health
```bash
# Test core endpoints
curl -X POST http://localhost:8080/accounts -d '{"owner": "Test"}'
curl http://localhost:8080/accounts/1/balance
curl -X POST http://localhost:8080/accounts/1/deposit -d '{"amount": 1000}'

# Check metrics
curl http://localhost:8080/metrics

# Verify concurrent safety
go test -run TestConcurrentTransfer ./tests/integration/account/
```

This deployment setup provides a solid foundation for running the banking API from development to production scale.