# Kubernetes Deployment Quick Start

## Prerequisites
- Kubernetes cluster (v1.19+)
- kubectl configured
- Docker images built and pushed to your registry

## Step 1: Update Image References

Edit the manifest files to point to your container registry:
- `manifests/server/deployment.yaml` 
- `manifests/agent/deployment.yaml`

Change `your-registry/chaos-master:latest` and `your-registry/chaos-agent:latest` to your actual images.

## Step 2: Deploy Server

```bash
# Deploy the master server
kubectl apply -k manifests/server/

# Wait for it to be ready
kubectl wait --for=condition=available --timeout=60s deployment/chaos-master

# Verify it's running
kubectl get pods -l app=chaos-master
kubectl logs -l app=chaos-master
```

## Step 3: Deploy Agents

```bash
# Deploy agents (default: 3 replicas)
kubectl apply -k manifests/agent/

# Verify agents are running
kubectl get pods -l app=chaos-agent

# Check agent logs to see them connecting to master
kubectl logs -l app=chaos-agent --tail=20
```

## Step 4: Verify Connections

Check the master logs to see agents connecting:
```bash
kubectl logs -l app=chaos-master | grep "Agent connected"
```

You should see messages like:
```
Agent connected: agentId=<uuid>
Successfully registered with master. Agent ID: <uuid>
```

## Step 5: Scale Agents

```bash
# Scale to 10 agents
kubectl scale deployment/chaos-agent --replicas=10

# Verify
kubectl get pods -l app=chaos-agent
```

## Cleanup

```bash
kubectl delete -k manifests/agent/
kubectl delete -k manifests/server/
```

## Troubleshooting

### Agents not connecting?
```bash
# Check agent logs
kubectl logs -l app=chaos-agent --tail=50

# Check if service exists
kubectl get svc chaos-master-service

# Check if master is ready
kubectl get pods -l app=chaos-master
```

### Need to change configuration?
```bash
# Edit the configmap
kubectl edit configmap chaos-agent-config

# Restart agents to pick up changes
kubectl rollout restart deployment/chaos-agent
```

## Next Steps

See [manifests/README.md](manifests/README.md) for detailed configuration options and production considerations.
