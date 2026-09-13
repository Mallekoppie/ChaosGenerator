# Kubernetes Manifests

This directory contains Kubernetes manifests for deploying the Chaos Generator system.

## Directory Structure

```
manifests/
├── server/           # Chaos Master server manifests
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── pvc.yaml
│   └── kustomization.yaml
└── agent/            # Chaos Agent manifests
    ├── deployment.yaml
    ├── configmap.yaml
    ├── deployment-with-configmap.yaml
    └── kustomization.yaml
```

## Deployment

### Quick Start (Deploy Everything)

Deploy the entire stack:
```bash
# Deploy server first
kubectl apply -k manifests/server/

# Wait for server to be ready
kubectl wait --for=condition=available --timeout=60s deployment/chaos-master

# Deploy agents
kubectl apply -k manifests/agent/
```

### Deploy Server Only

```bash
kubectl apply -k manifests/server/
```

Or individually:
```bash
kubectl apply -f manifests/server/pvc.yaml
kubectl apply -f manifests/server/service.yaml
kubectl apply -f manifests/server/deployment.yaml
```

### Deploy Agents Only

**Option 1: Direct environment variables**
```bash
kubectl apply -f manifests/agent/deployment.yaml
```

**Option 2: Using ConfigMap**
```bash
kubectl apply -f manifests/agent/configmap.yaml
kubectl apply -f manifests/agent/deployment-with-configmap.yaml
```

**Option 3: Using Kustomize**
```bash
kubectl apply -k manifests/agent/
```

## Configuration

### Agent Configuration

Modify environment variables in:
- `manifests/agent/deployment.yaml` - for direct env vars
- `manifests/agent/configmap.yaml` - for ConfigMap-based config

Key environment variables:
- `CHAOS_MASTER_ADDRESS`: Address of the master service (default: `chaos-master-service:9002`)
- `CHAOS_AGENT_VERSION`: Agent version string
- `CHAOS_AGENT_PORT`: Agent port (not exposed, for future use)
- `CHAOS_AGENT_METRICS_PORT`: Metrics port

### Server Configuration

Modify `manifests/server/deployment.yaml` for:
- Resource limits
- Storage size (in `pvc.yaml`)
- Replicas (keep at 1 for now due to BoltDB)

## Scaling

### Scale Agents
```bash
# Scale to 5 agents
kubectl scale deployment/chaos-agent --replicas=5

# Scale to 10 agents
kubectl scale deployment/chaos-agent --replicas=10
```

### Server Scaling
**Note:** Currently, the server uses BoltDB which doesn't support multiple replicas. Keep replicas=1.

## Monitoring

### Check Status
```bash
# Server status
kubectl get pods -l app=chaos-master
kubectl logs -l app=chaos-master --tail=50

# Agent status
kubectl get pods -l app=chaos-agent
kubectl logs -l app=chaos-agent --tail=20
```

### Check Agent Connections
Check server logs to see connected agents:
```bash
kubectl logs -l app=chaos-master | grep "Agent connected"
```

## Cleanup

```bash
# Delete everything
kubectl delete -k manifests/agent/
kubectl delete -k manifests/server/
```

Or individually:
```bash
kubectl delete -f manifests/agent/
kubectl delete -f manifests/server/
```

## Customization

### Using Kustomize

Each directory has a `kustomization.yaml` that you can use as a base for your own customizations:

```yaml
# my-custom-kustomization.yaml
bases:
  - ../../manifests/agent/

namespace: my-chaos-namespace

replicas:
  - name: chaos-agent
    count: 10

configMapGenerator:
  - name: chaos-agent-config
    behavior: replace
    literals:
      - CHAOS_MASTER_ADDRESS=my-custom-master:9002
      - CHAOS_AGENT_VERSION=2.1.0
```

Apply with:
```bash
kubectl apply -k my-custom-kustomization.yaml
```

## Production Considerations

1. **Image Registry**: Update image references in deployment manifests to your registry
2. **Resources**: Adjust resource requests/limits based on your workload
3. **Storage**: Configure appropriate storage class and size for the master PVC
4. **Networking**: Consider using Ingress or LoadBalancer for external access
5. **Security**: Add pod security policies, network policies, and RBAC as needed
6. **Monitoring**: Set up Prometheus metrics collection
7. **High Availability**: For production, consider distributed database instead of BoltDB
