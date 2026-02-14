# Docker Full Stack - Quick Reference

## Single Command Setup

```powershell
# Start everything (builds images, starts all services + monitoring)
make docker-up

# Stop everything
make stop-full-stack
```

## Access URLs

| Service | URL | Login |
|---------|-----|-------|
| Grafana | http://localhost:3000 | admin/admin |
| Prometheus | http://localhost:9091 | - |
| Target HTTP | http://localhost:8080 | - |
| Target HTTPS | https://localhost:8443 | - |
| Target HTTP/2 | https://localhost:8444 | - |

## Run Load Tests

```powershell
# Run client (connects to Docker master)
make run-client
```

## Common Commands

```powershell
# View all logs
make docker-logs

# Restart services
make docker-restart

# Stop containers
make docker-down

# Rebuild images
make docker-build
```

## Grafana Dashboards

1. **Chaos Agents Dashboard** - Agent metrics and activity
2. **Chaos Target Services Dashboard** - Service performance
3. **Client Perspective** - Load testing from client view
4. **Service Perspective** - Load testing from service view

## Port Reference

| Service | Port | Purpose |
|---------|------|---------|
| Master | 9002 | gRPC |
| Agent 1 | 9096 | Metrics |
| Agent 2 | 9097 | Metrics |
| Target HTTP | 8080 | Service |
| Target HTTP | 9090 | Metrics |
| Target HTTPS | 8443 | Service |
| Target HTTPS | 9093 | Metrics |
| Target HTTP/2 | 8444 | Service |
| Target HTTP/2 | 9094 | Metrics |
| Prometheus | 9091 | UI |
| Grafana | 3000 | UI |

## Troubleshooting

```powershell
# Check running containers
docker ps

# View specific service logs
docker-compose -f docker-compose-full-stack.yml logs chaos-master

# Check Prometheus targets
# Open: http://localhost:9091/targets

# Regenerate certificates
rmdir /s /q certs\target-http
make generate-certs
```

## Load Testing Workflow

1. Start stack: `make docker-up`
2. Open Grafana: http://localhost:3000
3. Run tests: `make run-client`
4. Monitor dashboards (Client + Service perspectives)
5. Take screenshots for evidence
6. Stop: `make stop-full-stack`

## Architecture

```
Client (Your Terminal)
    ↓
Chaos Master (Docker) (:9002)
    ↓
Agents (Docker) (:9096, :9097)
    ↓
Targets (Docker) (:8080, :8443, :8444)
    ↓
Prometheus (Docker) (:9091)
    ↓
Grafana (Docker) (:3000)
```

## Metrics

### Agent Metrics
- Request rate, duration, errors
- Connection time, TTFB
- Active tests and users

### Target Metrics  
- Request rate, duration, errors
- Processing time
- Success/failure counts

## Files

- `docker-compose-full-stack.yml` - Main compose file
- `monitoring/prometheus-docker.yml` - Prometheus config
- `monitoring/grafana-docker/provisioning/` - Grafana config
- `DOCKER-SETUP.md` - Full documentation

## See Also

- [DOCKER-SETUP.md](DOCKER-SETUP.md) - Detailed documentation
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration options
- [readme.md](readme.md) - Project overview
