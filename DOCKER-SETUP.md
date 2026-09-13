# Docker & Monitoring Setup

There are two ways to run the Chaos Generator stack on Linux:

1. **Native processes** - `make start` builds and runs the binaries directly (fastest for development).
2. **Container images** - `make docker-build` produces images for Kubernetes or Docker.

## Prerequisites

- Docker with the Compose plugin (`docker compose`) or the standalone `docker-compose`
- GNU Make
- Go 1.26+
- `protoc` (only when regenerating the stubs)
- Flutter + Dart SDK (only to build the web control panel or the master image locally)

## Container Images

| Target | Dockerfile | Image | Ports |
|--------|-----------|-------|-------|
| `make docker-build-master` | `build/dockerfile-chaos-master` | `chaos-master:latest` | 9002 gRPC, 8080 web |
| `make docker-build-agent` | `build/dockerfile-chaos-agent` | `chaos-agent:latest` | - |
| `make docker-build-target-http` | `build/dockerfile-target-http` | `target-http:latest` | 8080 |
| `make docker-build-target-grpc` | `build/dockerfile-target-grpc` | `target-grpc:latest` | 9000 |

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

## Running Containers

There is no long-lived Compose file for the application itself - deploy the
images with Kubernetes (see [KUBERNETES.md](KUBERNETES.md)) or run them
individually:

```bash
# Master: gRPC for agents plus the web control panel
docker run -d --name chaos-master \
  -p 9002:9002 -p 8080:8080 \
  -e CHAOS_JWT_SIGNING_KEY=change-me \
  -e CHAOS_REGISTER_SECRET=change-me \
  -v chaos-master-data:/data \
  chaos-master:latest

# Agent: connects out to the master
docker run -d --name chaos-agent-1 \
  --add-host=host.docker.internal:host-gateway \
  -e CHAOS_MASTER_ADDRESS=host.docker.internal:9002 \
  -e CHAOS_AGENT_METRICS_PORT=9096 \
  -p 9096:9096 \
  chaos-agent:latest

# Target service
docker run -d --name target-http -p 8080:8080 -p 9090:9090 target-http:latest
```

The containerised master serves the web control panel on **8080** (the image
sets `CHAOS_MASTER_WEB_ADDRESS=0.0.0.0:8080`), while the native default is 9001.
Open <http://localhost:8080/> and register the first user with the value of
`CHAOS_REGISTER_SECRET`.

## Monitoring

Prometheus and Grafana are started by `make monitoring-up`, which uses Docker
Compose when a Docker daemon is reachable and otherwise runs the native binaries
(Prometheus 3.14, Grafana 13.2) downloaded into `.monitoring/`.

```bash
make monitoring-up        # start Prometheus + Grafana
make monitoring-down      # stop them
make monitoring-restart
make monitoring-logs      # follow the logs
make monitoring-install   # only download the native binaries
make monitoring-clean     # remove .monitoring
```

| Service | URL | Login |
|---------|-----|-------|
| Prometheus | http://localhost:9091 | - |
| Grafana | http://localhost:3000 | admin/admin |

Override the ports if they are already in use:

```bash
make monitoring-up PROMETHEUS_PORT=9191 GRAFANA_PORT=3001
```

To run the native services and monitoring together:

```bash
make run-full-stack
```

## Local Stack Without Containers

```bash
make start     # master + AGENT_COUNT agents + targets in the background
make status    # per-process state and listening ports
make logs      # follow .run/*.log
make stop      # stop everything
```

`make start` prints a status table; a service that could not start (for example
because its port is already in use) is reported as `FAILED - see .run/<name>.log`.

## Port Reference

| Service | Port | Purpose |
|---------|------|---------|
| Chaos Master | 9002 | gRPC (agents) |
| Chaos Master | 9001 | Web UI + gRPC-Web (native); 8080 inside containers |
| Chaos Agent | 9096, 9097 | Metrics (one per agent) |
| Target HTTP | 8080 / 9090 | Service / metrics |
| Target HTTPS | 8443 / 9093 | Service / metrics |
| Target HTTP/2 | 8444 / 9094 | Service / metrics |
| Target gRPC | 9000 / 9095 | Service / metrics |
| Prometheus | 9091 | UI |
| Grafana | 3000 | UI |

## Files and Directories

```
build/
  dockerfile-chaos-master       # master image (builds and embeds the Flutter app)
  dockerfile-chaos-agent        # agent image
  dockerfile-target-http        # HTTP / HTTPS / HTTP2 target image
  dockerfile-target-grpc        # gRPC target image
docker-compose-monitoring.yml   # Prometheus + Grafana
monitoring/
  prometheus.yml                # scrape configuration
  grafana/provisioning/         # datasource and dashboard provisioning
manifests/
  server/                       # Kubernetes manifests (master)
  agent/                        # Kubernetes manifests (agents)
Makefile                        # build, run, clean and monitoring targets
```

## Cleanup

```bash
make clean            # binaries, generated code and web build output
make clean-certs      # generated TLS certificates
docker volume prune   # unused Docker volumes
```

## Related

- [KUBERNETES.md](KUBERNETES.md) - deploying the master and agents to Kubernetes
- [DOCKER-QUICK-REFERENCE.md](DOCKER-QUICK-REFERENCE.md) - command cheat sheet
- [CONFIGURATION.md](CONFIGURATION.md) - environment variables
- [readme.md](readme.md) - project overview
