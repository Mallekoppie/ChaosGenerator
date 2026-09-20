# Docker & Kind Setup

The Chaos Generator runs in a local **kind** cluster. There is no native
process mode and no local monitoring stack any more: `make kind-up` builds the
images, loads them into the cluster and deploys the master, the agents, the
target variants and Prometheus/Grafana.

```bash
make kind-up     # create cluster + build/load images + deploy everything
make proxy       # web UI on 9001, Prometheus on 9091, Grafana on 3000
```

`make kind-up` and `make proxy` need to run on the **host** - kind spawns
privileged containers, which does not work inside the dev container. Everything
else (`make all`, `make test`, `make manifests`, `make docker-build`) works in
the container.

## Prerequisites

- `kind` and `kubectl` on the host
- podman or Docker on the host (rootless podman works; the Makefile sets
  `KIND_EXPERIMENTAL_PROVIDER=podman` when it detects podman)
- GNU Make, Go 1.26+ and `protoc` for building
- Flutter + Dart SDK only when building the web app or the master image outside
  the container (the master image compiles the Flutter app itself)

## Container Images

| Target | Dockerfile | Image | Ports |
|--------|-----------|-------|-------|
| `make docker-build-master` | `build/dockerfile-chaos-master` | `chaos-master:latest` | 9002 gRPC, 8080 web |
| `make docker-build-agent` | `build/dockerfile-chaos-agent` | `chaos-agent:latest` | - |
| `make docker-build-target-http` | `build/dockerfile-target-http` | `target-http:latest` | 8080, 9090 metrics |
| `make docker-build-target-grpc` | `build/dockerfile-target-grpc` | `target-grpc:latest` | 9000, 9090 metrics |

`make docker-build` builds all four. The master image compiles the Flutter web
app during the build (Flutter is not required on the host) and embeds it in the
binary. Tag the images for a registry with:

```bash
make docker-build REGISTRY=registry.example.com/ IMAGE_TAG=1.2.3
```

If you host the control panel under a sub-path, pass `BASE_HREF`:

```bash
make docker-build-master BASE_HREF=/chaos/
```

### The `localhost/` image prefix

Rootless podman stores unqualified image names under a `localhost/` prefix, and
so does the `docker` shim that distros such as Bazzite install. If the images were
loaded into the node as `localhost/chaos-master:latest` while the manifests asked
for `chaos-master:latest`, the kubelet would try to pull the bare name from Docker
Hub and every pod would sit in `ImagePullBackOff`.

The Makefile therefore derives `REGISTRY` from the engine itself
(`docker --version` reports `podman version ...` through the shim):

| Engine | Default `REGISTRY` |
|--------|--------------------|
| podman (directly or via the `docker` shim) | `localhost/` |
| Docker | empty |

Override it if you push to a real registry: `make kind-up REGISTRY=registry.example.com/`.

## Distroless Runtime Images

The runtime stage of all four images is
[`gcr.io/distroless/static-debian13:nonroot`](https://github.com/GoogleContainerTools/distroless),
so a running container holds only the static binary plus its runtime
dependencies - a CA bundle (`/etc/ssl/certs`), `/tmp`, `/etc/passwd` - and runs
as the **`nonroot` user (uid/gid 65532)**. There is no shell, no package manager
and no libc.

| Concern | Consequence |
|---------|-------------|
| Probes | Must be `httpGet`/`tcpSocket`/`grpc`. An `exec` probe that runs `/bin/sh` can never succeed. The master serves `/healthz`; the agent and the HTTP target serve `/health` (the agent's is on its metrics port); the gRPC target uses a `tcpSocket` probe. |
| Interactive access | `docker exec ... sh` is impossible. For a shell use the `:debug` variant, e.g. `podman run --rm -it --entrypoint=/busybox/sh gcr.io/distroless/static-debian13:debug`, or `kubectl debug -it <pod> --image=busybox:stable --target=<container>`. |
| Writable volumes | `/data` is created in the build stage already owned by uid 65532, so named volumes (`-v chaos-master-data:/data`) work unchanged. A bind mount of a root-owned host directory fails with `permission denied`: `chown 65532:65532 <dir>` or add `--user 0`. |
| CA certificates | Bundled, so outbound TLS (OIDC discovery, HTTPS targets with public CAs) keeps working without `apk add ca-certificates`. |
| `$HOME` / `/tmp` | The `nonroot` variant sets `WORKDIR /home/nonroot`, and `/tmp` exists with mode `01777`, so tools that assume either keep working. |

## TLS for the target variants

`make generate-certs` creates one self-signed certificate in `certs/target-http`
covering `localhost` and the in-cluster service names, and `make kind-secrets`
applies it as the `target-tls` Secret. The TLS and HTTP/2 targets mount it:

```bash
make generate-certs
make kind-secrets      # also run automatically by make kind-deploy
```

The agents skip certificate verification, so the certificate matters for clients
that do verify. Delete it with `make clean-certs` to force regeneration.

## Monitoring

Prometheus and Grafana are part of the cluster deployment, not a local stack:

| Service | Local URL | Login |
|---------|-----------|-------|
| Prometheus | http://localhost:9091 | - |
| Grafana | http://localhost:3000 | admin/admin |

```bash
make proxy                 # forward the web UI and monitoring
make proxy-monitoring      # forward only Prometheus and Grafana
```

Override the local ports if they are taken:

```bash
make proxy PROMETHEUS_PORT=9191 GRAFANA_PORT=3001
```

See [monitoring/README.md](monitoring/README.md) for the scrape configuration and
how to add dashboards.

## Ports

Local ports are only what `make proxy` forwards; the services themselves listen
on their container ports inside the cluster.

| Service | In cluster | Local (proxy) | Override |
|---------|-----------|---------------|----------|
| Master web UI / gRPC-Web | 8080 | 9001 | `WEB_UI_PORT` |
| Master gRPC | 9002 | 9002 (`make proxy-grpc`) | `GRPC_PORT` |
| Prometheus | 9090 | 9091 | `PROMETHEUS_PORT` |
| Grafana | 3000 | 3000 | `GRAFANA_PORT` |
| Agent metrics | 9091 | - | - |
| Targets (HTTP/TLS/HTTP2) | 8080, metrics 9090 | - | - |
| Target gRPC | 9000, metrics 9090 | - | - |

## Files and Directories

```
build/
  dockerfile-chaos-master       # master image (builds and embeds the Flutter app)
  dockerfile-chaos-agent        # agent image
  dockerfile-target-http        # HTTP / HTTPS / HTTP2 target image
  dockerfile-target-grpc        # gRPC target image
manifests/
  kind/                         # self-contained kind manifests
    server/                     # master Deployment, Service, PVC, Secret
    agent/                      # agent Deployment + headless metrics Service
    target/                     # HTTP, TLS, HTTP2 and gRPC targets
    monitoring/                 # Prometheus and Grafana
    kustomization.template.yaml # rendered by `make manifests`
  server/ agent/ target-http/   # registry-oriented manifests for real clusters
monitoring/
  grafana/provisioning/         # datasource and dashboards, generated into ConfigMaps
Makefile                        # build, test, image, cluster and proxy targets
```

## Cleanup

```bash
make kind-down        # remove the Chaos resources (keep the cluster)
make kind-delete      # delete the kind cluster
make kind-clean       # delete the cluster and manifests/generated
make clean            # binaries, generated code and web build output
make clean-certs      # generated TLS certificates
```

## Related

- [KUBERNETES.md](KUBERNETES.md) - deploying to a real cluster with your own registry
- [DOCKER-QUICK-REFERENCE.md](DOCKER-QUICK-REFERENCE.md) - command cheat sheet
- [manifests/kind/README.md](manifests/kind/README.md) - the kind manifest set
