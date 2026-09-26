# Local kind Manifests

Self-contained manifests for testing the Chaos Generator in a local
[kind](https://kind.sigs.k8s.io/) cluster. This is the supported way to run the
whole system: the master, the agents, the target variants and the monitoring
stack all live in one namespace, and nothing runs on the host except `kubectl`.

They deliberately avoid everything a vanilla kind cluster does not provide:

- no `ServiceMonitor` (the Prometheus Operator CRDs are not installed), so
  Prometheus is deployed with a ConfigMap-based scrape config instead,
- local image names (`localhost/chaos-master:latest`, ...) instead of a
  `your-registry/...` placeholder. Podman stores unqualified names under
  `localhost/`, and the Makefile sets `REGISTRY=localhost/` to match, so the pods
  use the images loaded by `kind load` instead of pulling from Docker Hub,
- `imagePullPolicy: IfNotPresent` on every container, so the images loaded with
  `kind load` are used instead of being pulled from Docker Hub (a `:latest` tag
  would otherwise default to `Always`),
- distroless images that run as uid 65532 with no shell, so probes use
  `httpGet` / `tcpSocket` instead of running `pgrep` through a shell,
- `emptyDir` rather than PVCs for the Prometheus and Grafana data.

The registry/production oriented manifests live in `../server`, `../agent` and
`../target-http`.

## Contents

| Path | Resources |
|------|-----------|
| `namespace.yaml` | `chaos-testing` |
| `server/` | master Deployment (9002 gRPC, 8080 web), Service, PVC, Secret |
| `agent/` | agent Deployment (2 replicas, connects to `chaos-master-service:9002`) plus the headless metrics Service Prometheus discovers |
| `target/` | HTTP/1.1, TLS, HTTP/2 and gRPC targets, each with a Service and a metrics port |
| `monitoring/` | Prometheus and Grafana Deployments, Services and ConfigMaps |
| `kustomization.template.yaml` | Rendered into `generated/kustomization.yaml` |

## Usage

`make manifests` renders these directories into `manifests/generated/`
(gitignored) using the image names and tags produced by `make docker-build`, and
`make kind-up` deploys them:

```bash
make kind-up      # create cluster + build/load images + deploy everything
make proxy        # web UI on 9001, Prometheus on 9091, Grafana on 3000
make kind-status  # deployments, pods and services
make kind-verify  # agents that registered with the master
make kind-logs    # master, agent, target and Prometheus logs
make kind-down    # remove the Chaos resources (keep the cluster)
```

kind drives podman (or docker), which cannot run inside the dev container this
repository is usually edited in, so the `kind-*` targets and `make proxy` must
run on the host. The Makefile detects that and prints the host command to run
instead of failing obscurely. Everything else — `make all`, `make test`,
`make manifests`, `make docker-build` — works in the dev container.

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
| Web control panel | `make proxy`, then open <http://localhost:9001/> (override with `WEB_UI_PORT`) |
| Prometheus / Grafana | `make proxy` → <http://localhost:9091/> and <http://localhost:3000/> |
| Target under test (cluster DNS) | `http://target-http.chaos-testing.svc.cluster.local:8080` |
| TLS target | `https://target-http-tls.chaos-testing.svc.cluster.local` (self-signed certificate) |
| HTTP/2 target | `https://target-http2.chaos-testing.svc.cluster.local` |
| gRPC target | `target-grpc.chaos-testing.svc.cluster.local:9000` |
| Registration secret | see below (`chaos-dev-secret` by default) |

```bash
kubectl -n chaos-testing get secret chaos-master-secrets \
  -o jsonpath='{.data.registerSecret}' | base64 -d
```

Use the in-cluster URLs when configuring a test in the panel - the agents resolve
the service names through cluster DNS. The TLS and HTTP/2 targets serve the
self-signed certificate that `make generate-certs` creates and `make kind-secrets`
applies as the `target-tls` Secret.

## Teardown

```bash
make kind-down      # delete the Chaos resources, keep the cluster
make kind-delete    # delete the kind cluster
make kind-clean     # delete the cluster and manifests/generated
```

## Troubleshooting

| Symptom | Cause / fix |
|---------|-------------|
| Pods stuck in `ImagePullBackOff` | The images were not loaded (`make kind-images`), `imagePullPolicy` is no longer `IfNotPresent`, or the pod's image name does not match the loaded image. Podman stores unqualified names as `localhost/<name>`, so `REGISTRY` must stay `localhost/` on podman hosts - the Makefile derives it from `docker --version`, which also works through the podman `docker` shim. |
| `kind create cluster` fails | kind needs a container engine: start podman on the host (`systemctl --user start podman.socket`). The Makefile sets `KIND_EXPERIMENTAL_PROVIDER=podman` for you. |
| `kubectl` cannot reach the cluster | `kind create cluster` writes `~/.kube/config`. Check `KUBECONFIG`, or create the cluster with `kind create cluster --kubeconfig ~/.kube/kind-config` and export that path. |
| The master PVC stays `Pending` | kind's default `standard` StorageClass (local-path provisioner) is missing. Drop `server/pvc.yaml`, set a `storageClassName` that exists, or mount an `emptyDir`. |
| Agent pods restart repeatedly | They cannot reach the master: check `make kind-logs` and that `chaos-master-service:9002` resolves (master pod Ready). A restart or two right after the first `make kind-up` is expected - `chaos-agent` is applied before `chaos-master` (kustomize sorts by name), so the agent exits until the master accepts registrations. Re-run `make kind-deploy` if the rollout timed out. |
| Master exits on startup | Web port already in use, or the BoltDB path is not writable - see `make kind-logs`. |
| Master logs `permission denied` opening `/data/chaos_master.db` | The distroless image runs as uid 65532 while the provisioned volume is owned by root. kind's local-path provisioner normally hands out a `0777` directory, which works. If it does not, add a root initContainer (`busybox`, `chown -R 65532:65532 /data`) or set `securityContext.runAsUser: 0` on the master. |
| `kubectl exec -it <pod> -- sh` fails | The images are distroless: there is no shell. Use `kubectl debug -it <pod> --image=busybox:stable --target=<container>` or build the `:debug` variant of the image. |
| Agent probes fail with `exec: "/bin/sh": stat ... no such file or directory` | The probes were reverted to `exec`/`pgrep`, which cannot work in distroless. Use `httpGet` on `/health` (metrics port 9091). |
