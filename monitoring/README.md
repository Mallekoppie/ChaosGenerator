# Monitoring Setup

This directory contains the configuration for local monitoring with Prometheus and Grafana.

## Quick Start

Start the full stack (services + monitoring):
```powershell
make run-full-stack
```

Or start monitoring separately:
```powershell
make monitoring-up
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

```powershell
# Start monitoring only
make monitoring-up

# Stop monitoring
make monitoring-down

# View logs
make monitoring-logs

# Start full stack (services + monitoring)
make run-full-stack

# Stop services (but not monitoring)
make stop-all
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

## Docker Requirements

Requires Docker Desktop for Windows to be running.
