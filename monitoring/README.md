# Monitoring Setup

This directory contains the configuration for local monitoring with Prometheus and Grafana.

## Quick Start

Start the local stack and monitoring together:

```bash
make run-full-stack
```

Or start the services and monitoring separately:

```bash
make start          # master, agents and targets in the background
make monitoring-up  # Prometheus + Grafana
```

## Components

### Prometheus
- **URL**: http://localhost:9091
- **Configuration**: `prometheus.yml`
- Scrapes metrics from all target-http services (ports 9090, 9093, 9094)

### Grafana
- **URL**: http://localhost:3000
- **Login**: admin/admin
- **Pre-configured**:
  - Prometheus datasource
  - Target HTTP Services dashboard

## Metrics Endpoints

| Service | Port | Metrics Port |
|---------|------|--------------|
| Target HTTP (non-TLS) | 8080 | 9090 |
| Target HTTP (TLS) | 8443 | 9093 |
| Target HTTP/2 (TLS) | 8444 | 9094 |

## Dashboards

### Target HTTP Services Dashboard
Pre-provisioned dashboard showing:
- Request rate per service
- Response times (p95, p99)
- Error rates
- Request distribution by protocol and TLS status

## Make Targets

```bash
# Start monitoring only
make monitoring-up

# Stop monitoring
make monitoring-down

# Restart monitoring
make monitoring-restart

# View logs
make monitoring-logs

# Start the local services plus monitoring
make run-full-stack

# Stop the local services (monitoring keeps running)
make stop
```

## Configuration Files

- `prometheus.yml` - Prometheus scrape configuration
- `grafana/provisioning/datasources/` - Grafana datasource configs
- `grafana/provisioning/dashboards/` - Dashboard definitions

## Customization

### Adding More Scrape Targets

Edit `prometheus.yml` to add more services:
```yaml
scrape_configs:
  - job_name: 'my-service'
    static_configs:
      - targets: ['host.docker.internal:9095']
        labels:
          service: 'my-service'
```

### Adding Dashboards

Place JSON dashboard files in `grafana/provisioning/dashboards/` and they will be automatically loaded.

## How it runs

`make monitoring-up` picks a mode automatically:

- **Docker** when a Docker daemon is reachable: runs `docker-compose-monitoring.yml`.
- **Native** otherwise (for example inside a dev container without Docker): downloads
  Prometheus and Grafana into `.monitoring/` and runs them as background processes.

Force a mode with `MONITORING_MODE=docker` or `MONITORING_MODE=local`.

```bash
make monitoring-install   # just download the binaries
make monitoring-clean     # remove .monitoring (binaries + data)
```

## Ports

| Service | Default | Override |
|---------|---------|----------|
| Prometheus | 9091 | `PROMETHEUS_PORT` |
| Grafana | 3000 | `GRAFANA_PORT` |

If a port is already taken the command reports `FAILED - see .run/<name>.log` and
suggests free ports:

```bash
make monitoring-up PROMETHEUS_PORT=9191 GRAFANA_PORT=3001
```

## Metric names

Target and agent metrics follow [METRICS-STANDARDIZATION.md](../METRICS-STANDARDIZATION.md):

- `chaos_target_*` with labels `target_type`, `use_case`, `method`, `status_code`
- `chaos_agent_*` with labels `target_protocol`, `use_case`, `method`, `status_code`, `connection_pooled`

The provisioned dashboards use these names. Metric vectors only appear in
Prometheus after a test has generated at least one request.

## Requirements

Docker with the Compose plugin (`docker compose`) or the standalone `docker-compose`
for the container path. The native path only needs `curl`, `tar` and a free port;
the Makefile picks whichever mode is available.
