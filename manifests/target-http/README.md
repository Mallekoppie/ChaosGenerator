# Target HTTP Service Kubernetes Manifests

This directory contains Kubernetes manifests for deploying the target-http service in three configurations:
1. HTTP/1.1 without TLS (plain HTTP)
2. HTTP/1.1 with TLS (HTTPS)
3. HTTP/2 with TLS (HTTP/2 requires TLS)

## Files

- `deployment.yaml` - HTTP/1.1 deployment (no TLS)
- `deployment-tls.yaml` - HTTP/1.1 deployment with TLS
- `deployment-http2.yaml` - HTTP/2 deployment with TLS
- `service.yaml` - Service definitions for all three deployments
- `servicemonitor.yaml` - Prometheus ServiceMonitor for metrics collection
- `secret-tls.yaml` - TLS secret template (requires certificate generation)
- `kustomization.yaml` - Kustomize configuration

## Prerequisites

1. Kubernetes cluster with kubectl configured
2. (Optional) Prometheus Operator for ServiceMonitor support
3. TLS certificates for HTTPS and HTTP/2 deployments

## Generating TLS Certificates

Before deploying TLS or HTTP/2 services, generate certificates:

```bash
# Generate certificates locally
make generate-certs

# Create Kubernetes secret
kubectl create secret tls target-http-tls \
  --cert=certs/target-http/tls.crt \
  --key=certs/target-http/tls.key \
  -n chaos-testing
```

## Deployment Options

### Option 1: Deploy all configurations using Kustomize

```bash
# Create namespace
kubectl create namespace chaos-testing

# Generate and apply TLS certificates
make generate-certs
kubectl create secret tls target-http-tls \
  --cert=certs/target-http/tls.crt \
  --key=certs/target-http/tls.key \
  -n chaos-testing

# Deploy all services
kubectl apply -k manifests/target-http/
```

### Option 2: Deploy individual configurations

#### HTTP/1.1 without TLS
```bash
kubectl apply -f manifests/target-http/deployment.yaml -n chaos-testing
kubectl apply -f manifests/target-http/service.yaml -n chaos-testing
```

#### HTTP/1.1 with TLS
```bash
# Create TLS secret first
kubectl create secret tls target-http-tls \
  --cert=certs/target-http/tls.crt \
  --key=certs/target-http/tls.key \
  -n chaos-testing

# Deploy
kubectl apply -f manifests/target-http/deployment-tls.yaml -n chaos-testing
kubectl apply -f manifests/target-http/service.yaml -n chaos-testing
```

#### HTTP/2 with TLS
```bash
# Create TLS secret first
kubectl create secret tls target-http-tls \
  --cert=certs/target-http/tls.crt \
  --key=certs/target-http/tls.key \
  -n chaos-testing

# Deploy
kubectl apply -f manifests/target-http/deployment-http2.yaml -n chaos-testing
kubectl apply -f manifests/target-http/service.yaml -n chaos-testing
```

## Accessing the Services

After deployment, the services are available at:

- HTTP/1.1 (no TLS): `http://target-http.chaos-testing.svc.cluster.local:8080`
- HTTP/1.1 with TLS: `https://target-http-tls.chaos-testing.svc.cluster.local:8443`
- HTTP/2 with TLS: `https://target-http2.chaos-testing.svc.cluster.local:8443`

Metrics endpoints:
- All services expose metrics at port 9090: `/metrics`

## Use Case Endpoints

All deployments expose the same endpoints:

| Use Case | Path | Method |
|----------|------|--------|
| UC1 - Health Check | `/health` | GET |
| UC2 - Status | `/api/v1/status` | GET |
| UC3 - Simple Query | `/api/v1/users/` | GET |
| UC4 - Create Resource | `/api/v1/users` | POST |
| UC5 - List Collection | `/api/v1/users?page=1` | GET |
| UC6 - Bulk Update | `/api/v1/users/bulk` | PATCH |
| UC7 - Report Generation | `/api/v1/reports/generate` | POST |
| UC8 - File Upload | `/api/v1/documents/upload` | POST |
| UC9 - Data Export | `/api/v1/transactions/export` | GET |
| UC10 - Network Saturation | `/api/v1/data/process` | POST |
| UC11 - Memory Pressure | `/api/v1/analytics/stream` | POST |

## Monitoring

If Prometheus Operator is installed, ServiceMonitors will automatically configure Prometheus to scrape metrics from all target-http deployments.

Check metrics:
```bash
# Port-forward to access metrics
kubectl port-forward svc/target-http 9090:9090 -n chaos-testing

# View metrics
curl http://localhost:9090/metrics
```

## Scaling

Scale deployments:
```bash
# Scale HTTP/1.1 deployment
kubectl scale deployment/target-http --replicas=5 -n chaos-testing

# Scale HTTP/1.1 TLS deployment
kubectl scale deployment/target-http-tls --replicas=5 -n chaos-testing

# Scale HTTP/2 deployment
kubectl scale deployment/target-http2 --replicas=5 -n chaos-testing
```

## Resource Configuration

Default resource limits:
- Requests: 256Mi RAM, 250m CPU
- Limits: 2Gi RAM, 2000m CPU

Adjust in deployment files as needed for your cluster.

## Cleanup

```bash
# Delete all resources
kubectl delete -k manifests/target-http/

# Or delete individually
kubectl delete deployment target-http target-http-tls target-http2 -n chaos-testing
kubectl delete service target-http target-http-tls target-http2 -n chaos-testing
kubectl delete secret target-http-tls -n chaos-testing
```
