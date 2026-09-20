# Local kind Manifests

Self-contained manifests for testing the Chaos Generator in a local
[kind](https://kind.sigs.k8s.io/) cluster. They deploy the master, the agents and
one plain HTTP/1.1 target into a single namespace and deliberately avoid
everything a vanilla kind cluster does not provide:

- no `ServiceMonitor` (the Prometheus Operator CRDs are not installed),
- no TLS or HTTP/2 target deployments (no certificates to manage),
- local image names (`chaos-master:latest`, `chaos-agent:latest`,
  `target-http:latest`) instead of a `your-registry/...` placeholder,
- `imagePullPolicy: IfNotPresent` on every container, so the images loaded with
  `kind load` are used instead of being pulled from Docker Hub (a `:latest` tag
  would otherwise default to `Always`),
- distroless images that run as uid 65532 with no shell, so the agent's probes
  call `/health` over HTTP instead of running `pgrep` through a shell.

The registry/production oriented manifests live in `../server`, `../agent` and
`../target-http`.

## Contents

| Path | Resources |
|------|-----------|
| `namespace.yaml` | `chaos-testing` |
| `server/` | master Deployment (9002 gRPC, 8080 web), Service, PVC, Secret |
| `agent/` | agent Deployment (2 replicas, connects to `chaos-master-service:9002`) |
| `target/` | target-http Deployment + Service (8080 HTTP, 9090 metrics) |

## Usage

`make manifests` renders these directories into `manifests/generated/`
(gitignored) using the image names and tags produced by `make docker-build`, and
`make kind-up` deploys them:

```bash
make kind-up             # create cluster + build/load images + deploy
make kind-status         # deployments, pods, services and the PVC
make kind-verify         # agents that registered with the master
make kind-logs           # master, agent and target logs
make kind-port-forward   # web control panel on http://localhost:8080/
```

kind drives podman (or docker), which cannot run inside the distrobox container
this repository is usually edited in, so the `kind-*` targets must run on the
host. From inside distrobox:

```bash
make kind-host
# or
distrobox-host-exec bash -lc 'cd <repo> && make kind-up'
```

### Deploying by hand

```bash
make manifests
kubectl apply -f manifests/generated/kind-all.yaml
kubectl apply -f manifests/generated/kind-server.yaml   # master only
kubectl apply -f manifests/generated/kind-agent.yaml    # agents only
kubectl apply -f manifests/generated/kind-target.yaml   # target only
```

Use a registry or a different tag with:

```bash
make manifests REGISTRY=registry.example.com/ IMAGE_TAG=1.2.3
make kind-up KIND_CLUSTER=chaos2
```

## Accessing the deployment

| What | How |
|------|-----|
| Web control panel | `make kind-port-forward`, then open <http://localhost:8080/> |
| Target under test (cluster DNS) | `http://target-http.chaos-testing.svc.cluster.local:8080` |
| Target metrics (from the host) | `kubectl -n chaos-testing port-forward svc/target-http 9090:9090` |
| Registration secret | see below (`chaos-dev-secret` by default) |

```bash
kubectl -n chaos-testing get secret chaos-master-secrets \
  -o jsonpath='{.data.registerSecret}' | base64 -d
```

Use the in-cluster target URL when configuring a test in the panel - the agents
resolve `target-http.chaos-testing.svc.cluster.local` through cluster DNS.

## Teardown

```bash
make kind-down      # delete the Chaos resources, keep the cluster
make kind-delete    # delete the kind cluster
make kind-clean     # delete the cluster and manifests/generated
```

## Troubleshooting

| Symptom | Cause / fix |
|---------|-------------|
| Pods stuck in `ImagePullBackOff` | The images were not loaded (`make kind-images`) or `imagePullPolicy` is no longer `IfNotPresent`. |
| `kind create cluster` fails | kind needs a container engine: start podman on the host (`systemctl --user start podman.socket`). The Makefile sets `KIND_EXPERIMENTAL_PROVIDER=podman` for you. |
| `kubectl` cannot reach the cluster | `kind create cluster` writes `~/.kube/config`. Check `KUBECONFIG`, or create the cluster with `kind create cluster --kubeconfig ~/.kube/kind-config` and export that path. |
| The master PVC stays `Pending` | kind's default `standard` StorageClass (local-path provisioner) is missing. Drop `server/pvc.yaml`, set a `storageClassName` that exists, or mount an `emptyDir`. |
| Agent pods restart repeatedly | They cannot reach the master: check `make kind-logs` and that `chaos-master-service:9002` resolves (master pod Ready). A restart or two right after the first `make kind-up` is expected - `chaos-agent` is applied before `chaos-master` (kustomize sorts by name), so the agent exits until the master accepts registrations. Re-run `make kind-deploy` if the rollout timed out. |
| Master exits on startup | Web port already in use, or the BoltDB path is not writable - see `make kind-logs`. |
| Master logs `permission denied` opening `/data/chaos_master.db` | The distroless image runs as uid 65532 while the provisioned volume is owned by root. kind's local-path provisioner normally hands out a `0777` directory, which works. If it does not, add a root initContainer (`busybox`, `chown -R 65532:65532 /data`) or set `securityContext.runAsUser: 0` on the master. |
| `kubectl exec -it <pod> -- sh` fails | The images are distroless: there is no shell. Use `kubectl debug -it <pod> --image=busybox:stable --target=<container>` or build the `:debug` variant of the image. |
| Agent probes fail with `exec: "/bin/sh": stat ... no such file or directory` | The probes were reverted to `exec`/`pgrep`, which cannot work in distroless. Use `httpGet` on `/health` (metrics port 9091). |
