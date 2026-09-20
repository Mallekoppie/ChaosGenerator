# Kubernetes Deployment Quick Start

## Prerequisites
- Kubernetes cluster (v1.19+)
- kubectl configured
- Docker images built and pushed to your registry

## Local kind cluster (fastest way to test)

If you only want to try the manifests locally, the `manifests/kind/` set deploys
the master, the agents, the target variants (HTTP/1.1, TLS, HTTP/2, gRPC) and the
monitoring stack into a throwaway [kind](https://kind.sigs.k8s.io/) cluster with
images that are loaded straight into the node - no registry needed:

```bash
make kind-up      # create cluster + build/load images + deploy everything
make proxy        # web UI on 9001, Prometheus on 9091, Grafana on 3000
make kind-status  # deployments, pods and services
make kind-logs    # master, agent, target and Prometheus logs
make kind-verify  # agents that registered with the master
make kind-down    # delete the Chaos resources (keep the cluster)
make kind-clean   # delete the cluster and manifests/generated
```

kind needs podman (or docker), so it cannot run inside the dev container this
repository is usually edited in - run these targets in a host terminal.
`make manifests` renders `manifests/kind/` into `manifests/generated/` with the
images `make docker-build` produces. Configure tests against
`http://target-http.chaos-testing.svc.cluster.local:8080`, or against the TLS,
HTTP/2 and gRPC service names.

See [manifests/kind/README.md](manifests/kind/README.md) for details.

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

## Step 5: Access the Web Control Panel

The master serves a Flutter web control panel (over gRPC-Web) on port 8080
in-cluster. Forward it to your machine:

```bash
# Keep this running in its own terminal
kubectl port-forward svc/chaos-master-service 9001:8080
```

Then open <http://localhost:9001/>. When you are running the kind cluster,
`make proxy` forwards the web UI and the monitoring stack for you.

The first time you open the panel there are no users, so use **Register** and
supply the registration secret:

```bash
kubectl get secret chaos-master-secrets -o jsonpath='{.data.registerSecret}' | base64 -d
```

Change the `jwtSigningKey` and `registerSecret` values in
`manifests/server/secret.yaml` before deploying to a real cluster.

The panel talks to the master over **gRPC-Web** using the same protobuf
contract (`internal/contracts`) as the native gRPC API on port 9002, so every
call is authenticated with the JWT returned by the `Auth` service.

## Step 6: Scale Agents

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

### Probes fail / cannot get a shell in a pod

The images are distroless - no shell, no package manager, running as uid 65532 -
so `kubectl exec ... -- sh` and `exec` probes such as `pgrep` cannot work. Use
`httpGet`/`tcpSocket` probes (the master serves `/healthz`, the agent and the
HTTP target serve `/health`), and for debugging either build the `:debug`
variant or attach an ephemeral container:

```bash
kubectl debug -it <pod> --image=busybox:stable --target=<container>
```

## Next Steps

See [manifests/README.md](manifests/README.md) for detailed configuration options and production considerations.
