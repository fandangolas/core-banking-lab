# Kubernetes Monitoring Stack

This project uses the **kube-prometheus-stack** Helm chart to provide comprehensive monitoring for both the Core Banking application and the Kubernetes cluster.

## Components Included

The kube-prometheus-stack provides:

- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization and dashboards
- **AlertManager**: Alert routing and notifications
- **Node Exporter**: Host-level metrics (CPU, memory, disk, network)
- **Kube State Metrics**: Kubernetes object state metrics
- **Prometheus Operator**: Manages Prometheus configuration via CRDs

## Metrics Collection

### Application Metrics
- Banking API custom metrics at `/prometheus` endpoint
- Request rates, response times, transaction counts
- Account totals and operation metrics

### Infrastructure Metrics
- Node-level metrics (CPU, memory, disk, network usage)
- Container and pod metrics via kubelet/cAdvisor
- Kubernetes object states (deployments, services, pods)
- API server, scheduler, and controller manager metrics

## Deployment

### Using Ansible (Recommended)

Deploy the complete stack with monitoring:

```bash
cd infra/ansible
ansible-playbook -i inventory/aws_ec2.yml playbooks/deploy-with-helm.yml \
  -e github_actor=YOUR_GITHUB_USER \
  -e github_token=YOUR_GITHUB_TOKEN
```

### Manual Helm Installation

1. Add the Prometheus community repository:
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

2. Install the chart with custom values:
```bash
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  -f k8s/helm-values/kube-prometheus-stack-values.yaml \
  -n core-banking \
  --create-namespace \
  --version 58.0.0
```

3. Deploy the banking application:
```bash
kubectl apply -f k8s/01-api-deployment.yaml
kubectl apply -f k8s/02-dashboard-deployment.yaml
```

4. Apply the banking dashboard ConfigMap:
```bash
kubectl apply -f k8s/grafana-banking-dashboard-configmap.yaml
```

## Access Points

After deployment, services are available at:

- **Banking API**: `http://<node-ip>:30080`
- **Dashboard**: `http://<node-ip>:30000`
- **Prometheus**: `http://<node-ip>:30090`
- **Grafana**: `http://<node-ip>:30030`
  - Default credentials: `admin` / `admin`

## Configuration

### Helm Values

The main configuration is in `k8s/helm-values/kube-prometheus-stack-values.yaml`. Key settings:

- **Retention**: 7 days of metrics retention
- **Storage**: 20GB for Prometheus, 5GB for Grafana
- **Resources**: Conservative resource requests for cloud environments
- **Scraping**: Configured to auto-discover banking API pods

### Custom Dashboards

Banking-specific dashboards are deployed via ConfigMaps with the label `grafana_dashboard: "1"`. These are automatically imported into Grafana.

To add new dashboards:
1. Export the dashboard JSON from Grafana
2. Create a ConfigMap with the dashboard content
3. Add the `grafana_dashboard: "1"` label

## Monitoring Queries

Useful Prometheus queries for the banking application:

```promql
# Request rate by endpoint
rate(banking_api_requests_total[5m])

# 95th percentile response time
histogram_quantile(0.95, rate(banking_api_request_duration_seconds_bucket[5m]))

# Total transactions per hour
increase(banking_transaction_total[1h])

# Active accounts
banking_total_accounts

# Node CPU usage
100 - (avg by (instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# Pod memory usage
sum(container_memory_usage_bytes{namespace="core-banking"}) by (pod)
```

## Troubleshooting

### Check component status:
```bash
kubectl get pods -n core-banking
kubectl get pvc -n core-banking
helm status kube-prometheus-stack -n core-banking
```

### View Prometheus targets:
```bash
kubectl port-forward -n core-banking svc/kube-prometheus-stack-prometheus 9090:9090
# Visit http://localhost:9090/targets
```

### Check logs:
```bash
kubectl logs -n core-banking -l app.kubernetes.io/name=prometheus
kubectl logs -n core-banking -l app.kubernetes.io/name=grafana
```

## Upgrading

To upgrade the monitoring stack:

```bash
helm upgrade kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  -f k8s/helm-values/kube-prometheus-stack-values.yaml \
  -n core-banking \
  --version <new-version>
```

## Uninstalling

To remove the monitoring stack:

```bash
helm uninstall kube-prometheus-stack -n core-banking
kubectl delete pvc -n core-banking -l app.kubernetes.io/instance=kube-prometheus-stack
```