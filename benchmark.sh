#!/bin/bash

echo "🚀 Performance Optimization Test"
echo "================================"

# Build optimized image
echo "📦 Building optimized Banking API..."
docker build -f Dockerfile.api.optimized -t banking-api:optimized .

# Import to k3d
echo "📤 Importing to k3d cluster..."
k3d image import banking-api:optimized --cluster core-banking

# Apply optimized deployment
echo "🔧 Deploying optimized version..."
kubectl apply -f k8s/01-banking-api-optimized.yaml

# Wait for rollout
echo "⏳ Waiting for deployment..."
kubectl rollout status deployment banking-api -n core-banking

# Show resource usage
echo "📊 Current resource allocation:"
kubectl top pods -n core-banking | grep banking-api

echo ""
echo "✅ Optimized deployment ready!"
echo "Expected improvements:"
echo "  • CPU cores: 1 → 8 (8x more parallelism)"
echo "  • GC frequency: 100% → 400% (4x less GC)"
echo "  • Memory limit: 512MB → 2GB (4x more heap)"
echo "  • Logging: JSON → text + error level only"
echo ""
echo "🎯 Expected results:"
echo "  • Throughput: 6K → 50K+ RPS"
echo "  • GC pause: 950ms → <50ms"
echo "  • CPU usage: 95% → 40-60%"
echo "  • P99 latency: 100ms → <10ms"