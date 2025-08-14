#!/bin/bash

# Test script for kube-prometheus-stack deployment
# This script validates that all monitoring components are properly deployed and accessible

set -e

NAMESPACE="core-banking"
NODE_IP="${1:-localhost}"

echo "========================================="
echo "Monitoring Stack Deployment Test"
echo "========================================="
echo ""

# Function to check if a deployment is ready
check_deployment() {
    local deployment=$1
    echo -n "Checking $deployment... "
    
    if kubectl get deployment -n $NAMESPACE $deployment &> /dev/null; then
        local ready=$(kubectl get deployment -n $NAMESPACE $deployment -o jsonpath='{.status.readyReplicas}')
        local desired=$(kubectl get deployment -n $NAMESPACE $deployment -o jsonpath='{.spec.replicas}')
        
        if [ "$ready" == "$desired" ] && [ "$ready" -gt 0 ]; then
            echo "✓ Ready ($ready/$desired)"
            return 0
        else
            echo "✗ Not ready ($ready/$desired)"
            return 1
        fi
    else
        echo "✗ Not found"
        return 1
    fi
}

# Function to check if a statefulset is ready
check_statefulset() {
    local sts=$1
    echo -n "Checking $sts... "
    
    if kubectl get statefulset -n $NAMESPACE $sts &> /dev/null; then
        local ready=$(kubectl get statefulset -n $NAMESPACE $sts -o jsonpath='{.status.readyReplicas}')
        local desired=$(kubectl get statefulset -n $NAMESPACE $sts -o jsonpath='{.spec.replicas}')
        
        if [ "$ready" == "$desired" ] && [ "$ready" -gt 0 ]; then
            echo "✓ Ready ($ready/$desired)"
            return 0
        else
            echo "✗ Not ready ($ready/$desired)"
            return 1
        fi
    else
        echo "✗ Not found"
        return 1
    fi
}

# Function to check if a daemonset is ready
check_daemonset() {
    local ds=$1
    echo -n "Checking $ds... "
    
    if kubectl get daemonset -n $NAMESPACE $ds &> /dev/null; then
        local ready=$(kubectl get daemonset -n $NAMESPACE $ds -o jsonpath='{.status.numberReady}')
        local desired=$(kubectl get daemonset -n $NAMESPACE $ds -o jsonpath='{.status.desiredNumberScheduled}')
        
        if [ "$ready" == "$desired" ] && [ "$ready" -gt 0 ]; then
            echo "✓ Ready ($ready/$desired)"
            return 0
        else
            echo "✗ Not ready ($ready/$desired)"
            return 1
        fi
    else
        echo "✗ Not found"
        return 1
    fi
}

# Function to test HTTP endpoint
test_endpoint() {
    local name=$1
    local port=$2
    local path=${3:-/}
    
    echo -n "Testing $name (http://$NODE_IP:$port$path)... "
    
    if curl -s -o /dev/null -w "%{http_code}" "http://$NODE_IP:$port$path" | grep -q "200\|301\|302"; then
        echo "✓ Accessible"
        return 0
    else
        echo "✗ Not accessible"
        return 1
    fi
}

echo "1. Checking Kubernetes Components"
echo "---------------------------------"

# Check core components
check_deployment "kube-prometheus-stack-grafana"
check_deployment "kube-prometheus-stack-kube-state-metrics"
check_deployment "kube-prometheus-stack-operator"
check_statefulset "prometheus-kube-prometheus-stack-prometheus"
check_statefulset "alertmanager-kube-prometheus-stack-alertmanager"
check_daemonset "kube-prometheus-stack-prometheus-node-exporter"

echo ""
echo "2. Checking Banking Application"
echo "-------------------------------"

check_deployment "bank-api"
check_deployment "dashboard"

echo ""
echo "3. Checking Service Endpoints"
echo "-----------------------------"

test_endpoint "Banking API" 30080 "/health"
test_endpoint "Dashboard" 30000 "/"
test_endpoint "Prometheus" 30090 "/-/ready"
test_endpoint "Grafana" 30030 "/api/health"

echo ""
echo "4. Checking Prometheus Targets"
echo "------------------------------"

# Check if Prometheus is scraping targets
echo -n "Fetching Prometheus targets... "
targets=$(curl -s "http://$NODE_IP:30090/api/v1/targets" 2>/dev/null | grep -o '"health":"up"' | wc -l)
if [ "$targets" -gt 0 ]; then
    echo "✓ $targets healthy targets"
else
    echo "✗ No healthy targets found"
fi

echo ""
echo "5. Checking Metrics Collection"
echo "------------------------------"

# Check for banking metrics
echo -n "Banking API metrics... "
if curl -s "http://$NODE_IP:30090/api/v1/query?query=banking_api_requests_total" | grep -q "success"; then
    echo "✓ Available"
else
    echo "✗ Not found"
fi

echo -n "Node metrics... "
if curl -s "http://$NODE_IP:30090/api/v1/query?query=node_cpu_seconds_total" | grep -q "success"; then
    echo "✓ Available"
else
    echo "✗ Not found"
fi

echo -n "Kubernetes metrics... "
if curl -s "http://$NODE_IP:30090/api/v1/query?query=kube_pod_info" | grep -q "success"; then
    echo "✓ Available"
else
    echo "✗ Not found"
fi

echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="

# Count all pods
total_pods=$(kubectl get pods -n $NAMESPACE --no-headers | wc -l)
running_pods=$(kubectl get pods -n $NAMESPACE --no-headers | grep "Running" | wc -l)

echo "Total Pods: $total_pods"
echo "Running Pods: $running_pods"

if [ "$running_pods" -eq "$total_pods" ] && [ "$total_pods" -gt 0 ]; then
    echo ""
    echo "✓ All monitoring components are healthy!"
    echo ""
    echo "Access points:"
    echo "  - Banking API: http://$NODE_IP:30080"
    echo "  - Dashboard: http://$NODE_IP:30000"
    echo "  - Prometheus: http://$NODE_IP:30090"
    echo "  - Grafana: http://$NODE_IP:30030 (admin/admin)"
    exit 0
else
    echo ""
    echo "✗ Some components are not ready. Check pod status:"
    kubectl get pods -n $NAMESPACE
    exit 1
fi