# Kubernetes Deployment Guide - Initial Tests

This document outlines the complete process for deploying the Core Banking Lab application to a local k3s cluster using k3d.

## Prerequisites

- Docker installed and running
- k3d installed (via Homebrew: `brew install k3d`)
- kubectl configured

## Step 1: Install k3d

```bash
# Install k3d via Homebrew (macOS)
brew install k3d

# Verify installation
k3d version
```

## Step 2: Create k3d Cluster

```bash
# Create cluster with port mappings for our services
k3d cluster create banking-cluster \
  --port "8080:30080@agent:0" \
  --port "3000:30000@agent:0" \
  --port "9090:30090@agent:0" \
  --agents 2

# Verify cluster is running
kubectl cluster-info
kubectl get nodes
```

Expected output:
```
NAME                           STATUS   ROLES                  AGE   VERSION
k3d-banking-cluster-agent-0    Ready    <none>                 36s   v1.31.5+k3s1
k3d-banking-cluster-agent-1    Ready    <none>                 36s   v1.31.5+k3s1
k3d-banking-cluster-server-0   Ready    control-plane,master   39s   v1.31.5+k3s1
```

## Step 3: Build Docker Images

```bash
# Build the banking API image
docker build -f Dockerfile.api -t banking-api:latest .

# Build the dashboard image
cd dev/dashboard
docker build -t banking-dashboard:latest .
cd ../..
```

## Step 4: Import Images to k3d Cluster

```bash
# Import API image into k3d cluster
k3d image import banking-api:latest -c banking-cluster

# Import dashboard image into k3d cluster
k3d image import banking-dashboard:latest -c banking-cluster
```

## Step 5: Deploy Kubernetes Manifests

```bash
# Deploy in order: namespace first, then applications
kubectl apply -f k8s/00-namespace.yaml
kubectl apply -f k8s/01-api-deployment.yaml
kubectl apply -f k8s/02-dashboard-deployment.yaml
kubectl apply -f k8s/03-monitoring-deployment.yaml
```

## Step 6: Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n banking

# Check services
kubectl get services -n banking

# Check pod logs
kubectl logs -n banking deployment/banking-api --tail=5
```

Expected pod status:
```
NAME                                READY   STATUS    RESTARTS   AGE
banking-api-7dcbcbbd7-sltg4         1/1     Running   0          41s
banking-api-7dcbcbbd7-vs4lf         1/1     Running   0          41s
banking-dashboard-c969b956c-wf6rf   1/1     Running   0          29s
grafana-697fd6c7dc-8lrj4            1/1     Running   0          20s
prometheus-77d97d7c94-tsfrt         1/1     Running   0          20s
```

Expected services:
```
NAME                        TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
banking-api-service         NodePort   10.43.159.146   <none>        8080:30080/TCP   44s
banking-dashboard-service   NodePort   10.43.39.148    <none>        5173:30000/TCP   32s
grafana-service             NodePort   10.43.37.25     <none>        3000:30030/TCP   23s
prometheus-service          NodePort   10.43.28.72     <none>        9090:30090/TCP   23s
```

## Step 7: Test Application

### Test API Endpoint
```bash
# Test API health
curl -s http://localhost:8080/prometheus | head -5

# Create test account
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"owner": "TestUser123", "initialDeposit": 1000}'

# Check account balance
curl http://localhost:8080/accounts/1/balance
```

### Test Dashboard
```bash
# Check dashboard is accessible
curl -s http://localhost:3000 | grep -i "title"
```

### Test Prometheus
```bash
# Check Prometheus targets
curl -s http://localhost:9090/api/v1/targets

# Query metrics
curl -s "http://localhost:9090/api/v1/query?query=http_requests_total"
```

### Test Grafana
```bash
# Port forward Grafana (since it uses different NodePort)
kubectl port-forward -n banking svc/grafana-service 3001:3000 &

# Test access
curl -s http://localhost:3001 | head -5

# Login: admin / admin123
```

## Access Points

After successful deployment:

- **Banking API**: http://localhost:8080
- **Dashboard**: http://localhost:3000
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (via port-forward)

## Troubleshooting

### Check Pod Status
```bash
kubectl describe pods -n banking
kubectl logs -n banking <pod-name>
```

### Check Service Configuration
```bash
kubectl get endpoints -n banking
kubectl describe service <service-name> -n banking
```

### Restart Deployment
```bash
kubectl rollout restart deployment/<deployment-name> -n banking
```

## Cleanup

```bash
# Delete all resources
kubectl delete namespace banking

# Delete k3d cluster
k3d cluster delete banking-cluster
```

## Notes

- All images use `imagePullPolicy: Never` for local development
- Resource limits are set for development environment
- For production deployment, replace with proper container registry images
- Port mappings are configured in k3d cluster creation for easy access

## Next Steps for AWS Deployment

1. Push images to container registry (ECR, Docker Hub)
2. Update image references in manifests
3. Change `imagePullPolicy: Never` to `imagePullPolicy: Always`
4. Deploy to EC2 k3s cluster using the same manifests