# Docker & Kind - Quick Reference

## Container images

```bash
make docker-build              # all four images
make docker-build-master       # builds the Flutter web app inside the image
make docker-build-agent
make docker-build-target-http
make docker-build-target-grpc
```

To tag images for a registry:

```bash
make docker-build REGISTRY=registry.example.com/ IMAGE_TAG=1.2.3
```

`REGISTRY` defaults to `localhost/` when the engine is podman (including the
`docker` shim on Bazzite). Podman stores unqualified image names under
`localhost/`, and the kind manifests must use the same name that `kind load` puts
into the node, otherwise every pod fails with `ImagePullBackOff`. With real Docker
it defaults to empty.

The master image embeds the web control panel. Override the hosting path with
`make docker-build-master BASE_HREF=/chaos/` if needed.

The runtime images are distroless
(`gcr.io/distroless/static-debian13:nonroot`): no shell, no package manager, no
libc, running as uid 65532. Use `httpGet`/`tcpSocket` probes (the master serves
`/healthz`, the agent and the HTTP target serve `/health`) and the `:debug` tag
when you need a shell for troubleshooting - see
[DOCKER-SETUP.md](DOCKER-SETUP.md#distroless-runtime-images).

## Kind cluster

Everything runs in a kind cluster. There are no local run targets any more.

```bash
make manifests      # render manifests/kind -> manifests/generated (works in the container)
make kind-up        # create cluster + build/load images + deploy everything
make proxy          # web UI 9001, Prometheus 9091, Grafana 3000
make kind-status    # deployments, pods and services
make kind-logs      # master, agent, target and Prometheus logs
make kind-verify    # agents that registered with the master
make kind-restart   # roll the deployments after rebuilding images
make kind-down      # remove the Chaos resources (keep the cluster)
make kind-delete    # delete the kind cluster
make kind-clean     # delete the cluster and manifests/generated
```

`make kind-up` and the proxy targets must run on the **host**: kind spawns
privileged containers, which does not work inside the dev container. `make all`,
`make test`, `make manifests` and `make docker-build` all work in the container.

## Monitoring

Prometheus and Grafana run in the cluster and come up with `make kind-up`.

| Service | Local URL | Login |
|---------|-----------|-------|
| Prometheus | http://localhost:9091 | - |
| Grafana | http://localhost:3000 | admin/admin |

Override with `PROMETHEUS_PORT` / `GRAFANA_PORT`, or use `make proxy-monitoring` to
forward only these two. Prometheus scrapes the target variants and every agent pod
with the same job names and labels as before, so the provisioned dashboards keep
working.

## Access URLs

| Service | URL | Login |
|---------|-----|-------|
| Web control panel | http://localhost:9001 | register with the registration secret |
| Grafana | http://localhost:3000 | admin/admin |
| Prometheus | http://localhost:9091 | - |

`WEB_UI_PORT` overrides the web UI port, because 8080 is frequently already taken
on a workstation. Inside the cluster the master still listens on 8080.

## In-cluster ports

| Service | Ports | Purpose |
|---------|-------|---------|
| Master | 8080 / 9002 | Web UI + gRPC-Web / gRPC (agents) |
| Agent (2 replicas) | 9091 | Metrics |
| Target HTTP / TLS / HTTP2 | 8080, 9090 | Service / metrics |
| Target gRPC | 9000, 9090 | Service / metrics |
| Prometheus | 9090 | UI |
| Grafana | 3000 | UI |

## Grafana Dashboards

1. **Chaos Generator - Target HTTP Services** - request rate, latency and error
   rates per protocol and TLS variant
2. **Client vs Service - Load Testing Comparison** - client versus service view

## Troubleshooting

```bash
make kind-status                 # deployments, pods and services
make kind-logs                   # application logs
kubectl -n chaos-testing get pods
kubectl -n chaos-testing describe pod <pod>
make clean-certs                 # remove generated TLS certificates
```

| Symptom | Fix |
|---------|-----|
| Pods in `ImagePullBackOff` | The images were not loaded (`make kind-images`), or the pod's image name does not match the loaded image. Podman needs the `localhost/` prefix, which the Makefile derives automatically from the engine. |
| `make proxy` reports the stack is not deployed | Run `make kind-up` first. |
| A proxy port is already in use | `make proxy WEB_UI_PORT=9100 PROMETHEUS_PORT=9191 GRAFANA_PORT=3100` |

## Kubernetes (registry-based)

```bash
kubectl apply -k manifests/server/
kubectl apply -k manifests/agent/
```

See [KUBERNETES.md](KUBERNETES.md) for the full guide and
[manifests/kind/README.md](manifests/kind/README.md) for the local cluster.

## See Also

- [DOCKER-SETUP.md](DOCKER-SETUP.md) - container images and the distroless runtime
- [CONFIGURATION.md](CONFIGURATION.md) - configuration options
- [readme.md](readme.md) - project overview
