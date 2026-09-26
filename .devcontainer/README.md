# Development container

A reproducible toolchain for developing **ChaosGenerator** in VS Code (or the
`devcontainer` CLI) on Bazzite / Fedora with **rootless Podman**.

## What's inside

| Tool | Version | Why |
|------|---------|-----|
| Go | 1.26 (base image) | builds `cmd/*` |
| protoc | 27.3 | `make proto` |
| protoc-gen-go / protoc-gen-go-grpc | 1.36.11 / 1.2.0 | Go stubs |
| Dart + Flutter | stable channel (`/opt/flutter`) | `make buildweb`, `make proto-dart` |
| protoc-gen-dart | 25.1.0 | Dart stubs |
| kubectl / kind / helm | v1.31.0 / v0.24.0 / v3.16.2 | manifests / kind flow |
| podman + `docker` shim | (client only) | drive the **host** engine |
| make, jq, openssl, curl, socat... | | Makefile targets |

## How it talks to the host engine

The container does **not** run its own container engine. Instead:

* the host's rootless podman socket is bind-mounted at `/run/host-podman.sock`;
* `DOCKER_HOST` / `CONTAINER_HOST` point at it;
* a `/usr/local/bin/docker` shim forwards to `podman`.

`runArgs` includes `--userns=keep-id` so the container user (uid 1000) matches the
host user that owns both the bind-mounted workspace and the socket. Without it the
socket mount fails with `permission denied`.

This lets you run `make docker-build`, `docker ps`, etc. **inside** the container
while the images are actually built by the host's podman.

> **kind still cannot run in-container.** kind needs to spawn containers with its
> own privileges, which does not work nested here. Run `make kind-up`, the other
> `kind-*` targets and `make proxy` in a host terminal; the Makefile detects the
> container and prints the host command rather than failing obscurely. Building,
> testing and `make manifests` all work here.

## First run

The `postCreateCommand` (`.devcontainer/post-create.sh`) runs `make proto`-style
plugin checks, `go mod download`, and `flutter pub get`. If `internal/contracts/*.pb.go`
are missing you can regenerate them at any time:

```bash
make proto        # Go stubs
make all          # build every component
make manifests    # render manifests/kind -> manifests/generated (no cluster needed)
make buildweb     # Flutter web app -> cmd/master/dist
```

The image bakes in `protoc-gen-go`, `protoc-gen-go-grpc` and `protoc-gen-dart`, so
`make tools` is only an escape hatch for a bare host.

## Common tasks

```bash
make all                       # build master, agent, client, targets
make test                      # go test + flutter test
make lint                      # go vet + flutter analyze
make fmt                       # format the Go and Dart sources
make docker-build              # build images via the host engine
make manifests                 # render manifests/kind -> manifests/generated
make host TARGET=proxy         # optional host passthrough (see HOST_RUNNER)
```

Everything runtime-related - `make kind-up`, the other `kind-*` targets and
`make proxy` - runs on the host, because kind cannot run nested here.

Forwarded ports: 9001 (web UI), 9091 (Prometheus), 3000 (Grafana). These match the
`make proxy` defaults; if you change `WEB_UI_PORT` / `PROMETHEUS_PORT` /
`GRAFANA_PORT`, update the `forwardPorts` list in `devcontainer.json` too.

## Regenerating the image

```bash
podman build --format docker \
  -f .devcontainer/Dockerfile \
  -t chaosgenerator-dev \
  .
```

Then in VS Code: **Dev Containers: Rebuild Container**.
