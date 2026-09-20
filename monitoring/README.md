# Monitoring

Prometheus and Grafana run **inside the kind cluster**, alongside the master,
the agents and the targets. There is no local monitoring stack any more: the
compose file and the native binaries are gone.

```bash
make kind-up     # deploys the whole stack, monitoring included
make proxy       # forwards Prometheus (:9091) and Grafana (:3000) to your machine
```

| Service | Default URL | Login |
|---------|-------------|-------|
| Prometheus | http://localhost:9091 | - |
| Grafana | http://localhost:3000 | admin/admin |

Override the local ports on the command line:

```bash
make proxy PROMETHEUS_PORT=9191 GRAFANA_PORT=3001
```

## What gets deployed

`make manifests` renders `manifests/kind/monitoring/` into
`manifests/generated/`, and `make kind-deploy` applies it:

| Resource | Purpose |
|----------|---------|
| `prometheus.yaml` | Prometheus Deployment + Service (`prometheus:9090`) |
| `prometheus-configmap.yaml` | The scrape configuration |
| `grafana.yaml` | Grafana Deployment + Service (`grafana:3000`) |
| `grafana-datasources` ConfigMap | Generated from `grafana/provisioning/datasources/` |
| `grafana-dashboards` ConfigMap | Generated from `grafana/provisioning/dashboards/` |

The two Grafana ConfigMaps are **generated** by the `manifests` target with
`kubectl create configmap --dry-run=client`, so the dashboards in this directory
stay the single source of truth and are never duplicated as YAML.

Both Prometheus and Grafana use `emptyDir` storage: the data is disposable and a
PVC would stay `Pending` in a default kind cluster.

## Scrape configuration

The scrape config lives in `manifests/kind/monitoring/prometheus-configmap.yaml`
and targets cluster DNS names instead of the old `host.docker.internal:*` hosts:

| Job | Target | Labels |
|-----|--------|--------|
| `target-http-notls` | `target-http:9090` | `protocol=http1`, `tls=false` |
| `target-http-tls` | `target-http-tls:9090` | `protocol=http1`, `tls=true` |
| `target-http2-tls` | `target-http2:9090` | `protocol=http2`, `tls=true` |
| `target-grpc` | `target-grpc:9090` | `protocol=grpc` |
| `chaos-agents` | `chaos-agent-metrics:9091` (DNS discovery) | `service=chaos-agent` |

The job names and the `service` / `protocol` / `tls` labels are unchanged from
the previous host-based setup, so the existing dashboards keep working.

Agent pods are discovered through the headless `chaos-agent-metrics` Service with
`dns_sd_configs`, which scrapes every replica without needing RBAC for the
Kubernetes API.

## Dashboards

`grafana/provisioning/dashboards/` holds the dashboards that get provisioned:

| File | Dashboard |
|------|-----------|
| `chaos-target-http.json` | Chaos Generator - Target HTTP Services |
| `chaos-comparison.json` | Client vs Service - Load Testing Comparison |

To change a dashboard, edit the JSON here and run:

```bash
make manifests kind-deploy
```

Adding a new dashboard file also means adding it to the `grafana-dashboards`
`--from-file` list in the `manifests` target in the `Makefile`.

## Adding scrape targets

Add jobs to `prometheus-configmap.yaml` and re-apply:

```yaml
- job_name: 'my-service'
  static_configs:
    - targets: ['my-service.chaos-testing.svc.cluster.local:9090']
      labels:
        service: 'my-service'
```

```bash
make manifests kind-deploy
```

## Troubleshooting

| Symptom | Cause / fix |
|---------|-------------|
| Grafana is up but has no data | The datasource points at `http://prometheus:9090`; check the Prometheus targets page via `make proxy`. |
| A target shows as `down` | The target pod is not Ready. Check `make kind-status` and `make kind-logs`. |
| `chaos-agents` only lists some agents | The headless Service needs the `metrics` port name on the agent container. Scale-up takes one scrape interval to appear. |
| Dashboard changes have no effect | The ConfigMaps are generated during `make manifests`; re-run `make manifests kind-deploy`, then wait for the Grafana pod to restart. |
