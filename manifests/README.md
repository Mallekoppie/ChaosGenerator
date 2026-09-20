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
├── agent/            # Chaos Agent manifests
│   ├── deployment.yaml
│   ├── configmap.yaml
│   ├── deployment-with-configmap.yaml
│   └── kustomization.yaml
├── target-http/      # Target HTTP/HTTPS/HTTP2 service manifests
└── kind/             # Self-contained manifests for a local kind cluster
    ├── namespace.yaml
    ├── server/       # master Deployment, Service, PVC, Secret
    ├── agent/        # agent Deployment
    └── target/       # target-http Deployment + Service
```

`make manifests` renders `manifests/kind/` into `manifests/generated/`
(gitignored) - see [kind/README.md](kind/README.md).

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

## Local kind Cluster

`manifests/kind/` is a self-contained set for a throwaway
[kind](https://kind.sigs.k8s.io/) cluster: the master, the agents and a plain
HTTP/1.1 target in one namespace, with no ServiceMonitor CRDs, no TLS and image
names that match what `make docker-build` produces.

```bash
make kind-up             # create cluster + build/load images + deploy
make kind-status         # deployments, pods, services and the PVC
make kind-verify         # agents that registered with the master
make kind-logs           # master, agent and target logs
make kind-port-forward   # web control panel on http://localhost:8080/
make kind-down           # delete the Chaos resources
make kind-delete         # delete the kind cluster
make kind-clean          # delete the cluster and manifests/generated
```

`make manifests` renders the set into `manifests/generated/` (`kind-all.yaml`
plus `kind-server.yaml`, `kind-agent.yaml` and `kind-target.yaml`) and honours
`REGISTRY=` / `IMAGE_TAG=` overrides:

```bash
make manifests REGISTRY=registry.example.com/ IMAGE_TAG=1.2.3
kubectl apply -f manifests/generated/kind-all.yaml
```

The `kind-*` targets drive podman (or docker), so they must run on the host -
from inside distrobox use `make kind-host`. See
[kind/README.md](kind/README.md) for details and troubleshooting.

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

The agent image is distroless (no shell), so the liveness/readiness probes hit
the `/health` endpoint of the metrics server over HTTP instead of running
`pgrep` through a shell.

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
