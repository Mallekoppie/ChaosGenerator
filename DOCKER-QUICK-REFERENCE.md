# Docker & Monitoring - Quick Reference

## Container images

```bash
make docker-build              # all images
make docker-build-master       # builds the Flutter web app inside the image
make docker-build-agent
make docker-build-target-http
make docker-build-target-grpc
```

To tag images for a registry:

```bash
make docker-build REGISTRY=registry.example.com/ IMAGE_TAG=1.2.3
```

The master image embeds the web control panel. Override the hosting path with
`make docker-build-master BASE_HREF=/chaos/` if needed.

## Local stack (native processes)

```bash
make start            # master + agents + targets in the background
make status           # what is running and on which ports
make logs             # follow the logs
make stop             # stop everything
```

## Monitoring

```bash
make monitoring-up        # Docker if available, otherwise native binaries
make monitoring-down
make monitoring-restart
make monitoring-logs
make monitoring-install   # download Prometheus/Grafana into .monitoring
make monitoring-clean
make run-full-stack       # local stack + monitoring
```

Prometheus defaults to port **9091** and Grafana to **3000**; override with
`PROMETHEUS_PORT` / `GRAFANA_PORT` if they are taken.

## Access URLs

| Service | URL | Login |
|---------|-----|-------|
| Web control panel | http://localhost:9001 | register with the registration secret |
| Grafana | http://localhost:3000 | admin/admin |
| Prometheus | http://localhost:9091 | - |
| Target HTTP | http://localhost:8080 | - |
| Target HTTPS | https://localhost:8443 | - |
| Target HTTP/2 | https://localhost:8444 | - |

## Port Reference

| Service | Port | Purpose |
|---------|------|---------|
| Master | 9002 | gRPC (agents) |
| Master | 9001 | Web UI + gRPC-Web |
| Agent 1 / 2 | 9096 / 9097 | Metrics |
| Target HTTP | 8080 / 9090 | Service / metrics |
| Target HTTPS | 8443 / 9093 | Service / metrics |
| Target HTTP/2 | 8444 / 9094 | Service / metrics |
| Target gRPC | 9000 / 9095 | Service / metrics |
| Prometheus | 9091 | UI |
| Grafana | 3000 | UI |

## Grafana Dashboards

1. **Chaos Agents Dashboard** - Agent metrics and activity
2. **Chaos Target Services Dashboard** - Service performance
3. **Client Perspective** - Load testing from client view
4. **Service Perspective** - Load testing from service view

## Troubleshooting

```bash
docker ps          # running containers
make status        # local processes and listening ports
make logs          # follow the background stack logs
make clean-certs   # remove generated TLS certificates
```

## Kubernetes

```bash
kubectl apply -k manifests/server/
kubectl apply -k manifests/agent/
kubectl port-forward svc/chaos-master-service 8080:8080
# open http://localhost:8080/
```

See [KUBERNETES.md](KUBERNETES.md) for the full guide.

## See Also

- [DOCKER-SETUP.md](DOCKER-SETUP.md) - Detailed container and monitoring documentation
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration options
- [readme.md](readme.md) - Project overview
