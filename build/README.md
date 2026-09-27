# Docker Build File

This directory contains the single Dockerfile that builds the Chaos Generator
image.

> You normally do not run `docker build` by hand: `make docker-build` builds the
> image and `make kind-up` builds, loads and deploys it into a local kind cluster.
> The `docker run` examples below are for inspecting one component, not for
> running the stack - the supported runtime is the kind cluster described in
> [../DOCKER-SETUP.md](../DOCKER-SETUP.md).

## dockerfile

One multi-stage build produces one image, `chaos-generator`, that carries every
binary plus the Flutter web control panel embedded in the master:

| Binary | Select it with | Ports |
|--------|----------------|-------|
| `chaos-master` | `/chaos-master` | 9002 gRPC, 8080 web |
| `chaos-agent` | `/chaos-agent` | 9091 metrics |
| `target-http` | `/target-http` | 8080 / 8443, 9090 metrics |
| `target-grpc` | `/target-grpc` | 9000, 9090 metrics |

A container picks its binary with `command:` in Kubernetes, or with a positional
argument to `docker run`. Everything else - ports, TLS, the master address - is
configuration through environment variables, so the manifests stay declarative.

The runtime stage is
[`gcr.io/distroless/static-debian13:nonroot`](https://github.com/GoogleContainerTools/distroless):
no shell, no package manager, no libc, running as uid 65532. Probes must use
`httpGet`/`tcpSocket`, and `docker exec ... sh` is impossible - see
[../DOCKER-SETUP.md](../DOCKER-SETUP.md#distroless-runtime-images).

The build compiles the Flutter app first and embeds it into `chaos-master`, so
Flutter is not needed on the host and a Go-only change still gets the cached
Flutter layer from BuildKit.

### Build

```bash
make docker-build                       # tags chaos-generator:latest
make docker-build IMAGE_TAG=1.2.3
make docker-build REGISTRY=registry.example.com/ IMAGE_TAG=1.2.3
make docker-build BASE_HREF=/chaos/     # web app hosted under a sub-path
```

By hand:

```bash
docker build -f build/dockerfile -t chaos-generator:latest .
```

### Run one component

Passing a path after the image name replaces the default `CMD ["/chaos-master"]`:

```bash
# master (the default, so no argument needed)
docker run --rm -p 9002:9002 -p 8080:8080 \
  -v chaos-master-data:/data \
  -e CHAOS_MASTER_WEB_ADDRESS=0.0.0.0:8080 \
  -e CHAOS_MASTER_DB_PATH=/data/chaos_master.db \
  chaos-generator:latest

# agent
docker run --rm -p 9091:9091 \
  -e CHAOS_MASTER_ADDRESS=chaos-master:9002 \
  -e CHAOS_AGENT_METRICS_PORT=9091 \
  chaos-generator:latest /chaos-agent

# HTTP / TLS / HTTP2 target
docker run --rm -p 8080:8080 -p 9090:9090 \
  -e CHAOS_TARGET_PROTOCOL=http1 \
  -e CHAOS_TARGET_PORT=8080 \
  -e CHAOS_TARGET_METRICS_PORT=9090 \
  chaos-generator:latest /target-http

# gRPC target
docker run --rm -p 9000:9000 -p 9090:9090 \
  -e CHAOS_TARGET_PORT=9000 \
  -e CHAOS_TARGET_METRICS_PORT=9090 \
  chaos-generator:latest /target-grpc
```

### Publish

```bash
docker login ghcr.io                 # or: podman login ghcr.io
make docker-push REGISTRY=ghcr.io/mallekoppie/ IMAGE_TAG=v0.1.0
```

Pushing requires a **classic** personal access token with the `write:packages`
scope - GitHub Packages does not accept fine-grained tokens. A token missing the
scope fails with `permission_denied: The token provided does not match expected
scopes`. Run `podman logout ghcr.io` before logging in again so the old token is
dropped.

`REGISTRY` has to name a real registry for `docker-push`; the `localhost/`
default only exists for `make kind-load`. `docker-push` builds and tags the image
for that registry first (incremental, so it is cheap), then pushes it. Make the
GHCR package public if the cluster should pull it without an imagePullSecret.

Podman works exactly the same way: `DOCKER` and `CONTAINER_ENGINE` auto-detect
the engine, including through the `docker` shim that Bazzite installs, so
`make docker-build`, `make docker-push` and the kind targets need no overrides.

## Kubernetes Deployment

For Kubernetes deployment, use the manifests in the `../manifests/` directory.
Each Deployment sets `command:` to the binary it needs, so all of them reference
the same image.
