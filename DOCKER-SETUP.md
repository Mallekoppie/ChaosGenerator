# Docker-Based Full Stack Setup

This guide describes how to run the Chaos Generator system in Docker with full monitoring capabilities.

## Overview

The Docker-based setup includes:
- **Chaos Master**: gRPC server for test orchestration
- **2 Chaos Agents**: Load generation agents
- **3 Target HTTP Services**: Test targets (HTTP, HTTP+TLS, HTTP/2+TLS)
- **Prometheus**: Metrics collection and storage
- **Grafana**: Metrics visualization with pre-configured dashboards

## Prerequisites

- Docker Desktop for Windows
- OpenSSL (for certificate generation)
- Make
- Go 1.21+ (for building binaries)

## Quick Start

### 1. Start the Full Stack

```powershell
make docker-up
```

This single command will:
1. Generate TLS certificates (if needed)
2. Build all Docker images
3. Start all services
4. Configure Prometheus scraping
5. Provision Grafana dashboards

### 2. Access the Services

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin/admin |
| Prometheus | http://localhost:9091 | - |
| Target HTTP | http://localhost:8080 | - |
| Target HTTPS | https://localhost:8443 | - |
| Target HTTP/2 | https://localhost:8444 | - |
| Chaos Master | localhost:9002 | - |

### 3. Run Load Tests

From your terminal (not in Docker), run the client:

```powershell
make run-client
```

The client will connect to the Chaos Master running in Docker and orchestrate tests across the agents.

### 4. View Metrics

Open Grafana at http://localhost:3000 (admin/admin) and navigate to the dashboards:

1. **Chaos Agents Dashboard**: Agent-level metrics
2. **Chaos Target Services Dashboard**: Target service metrics
3. **Client Perspective**: Load testing from client's view
4. **Service Perspective**: Load testing from service's view

### 5. Stop the Stack

```powershell
make stop-full-stack
```

## Architecture

### Network Configuration

All services run on the `chaos-network` Docker bridge network:

```
┌─────────────────────────────────────────────────┐
│                chaos-network                    │
│                                                 │
│  ┌──────────────┐         ┌──────────────┐    │
│  │ Chaos Master │◄────────┤ Agent 1      │    │
│  │  :9002       │         │  metrics:9096│    │
│  └──────┬───────┘         └──────────────┘    │
│         │                                       │
│         │                 ┌──────────────┐    │
│         └─────────────────┤ Agent 2      │    │
│                           │  metrics:9097│    │
│                           └──────────────┘    │
│                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌────────┐
│  │ Target HTTP  │  │ Target HTTPS │  │ HTTP/2 │
│  │  :8080/:9090 │  │  :8443/:9093 │  │:8444/  │
│  └──────────────┘  └──────────────┘  │:9094   │
│                                       └────────┘
│                                                 │
│  ┌──────────────┐         ┌──────────────┐    │
│  │ Prometheus   │◄────────┤ Grafana      │    │
│  │  :9090       │         │  :3000       │    │
│  └──────────────┘         └──────────────┘    │
└─────────────────────────────────────────────────┘
         │                           │
         └───────────────┬───────────┘
                    Host Machine
              (localhost:9091, :3000)
```

### Port Mappings

| Container | Internal Port | Host Port | Purpose |
|-----------|---------------|-----------|---------|
| chaos-master | 9002 | 9002 | gRPC API |
| chaos-agent-1 | 9096 | 9096 | Metrics |
| chaos-agent-2 | 9097 | 9097 | Metrics |
| target-http | 8080 | 8080 | HTTP Service |
| target-http | 9090 | 9090 | HTTP Metrics |
| target-http-tls | 8443 | 8443 | HTTPS Service |
| target-http-tls | 9093 | 9093 | HTTPS Metrics |
| target-http2 | 8444 | 8444 | HTTP/2 Service |
| target-http2 | 9094 | 9094 | HTTP/2 Metrics |
| prometheus | 9090 | 9091 | Prometheus UI |
| grafana | 3000 | 3000 | Grafana UI |

### Persistent Storage

The following data persists across container restarts:

- **Chaos Master Database**: `/data/chaos-master.db` (Docker volume: `chaos-master-data`)
- **Prometheus Data**: `/prometheus` (Docker volume: `prometheus-data`)
- **Grafana Data**: `/var/lib/grafana` (Docker volume: `grafana-data`)

## Available Make Targets

### Docker Stack Management

```powershell
# Build Docker images
make docker-build

# Start full stack (build + up)
make docker-up

# Stop all containers
make docker-down

# Stop and clean up (preserves volumes)
make stop-full-stack

# Restart all containers
make docker-restart

# View logs for all services
make docker-logs
```

### Client Operations

```powershell
# Build and run client (connects to Docker master)
make run-client
```

### Local Development (Non-Docker)

```powershell
# Run services locally (Windows processes)
make run-test-setup

# Stop local services
make stop-all
```

## Grafana Dashboards

### 1. Chaos Agents Dashboard

**Purpose**: Monitor individual agent performance and activity

**Key Panels**:
- Request rate per agent
- Request duration percentiles (p95, p99)
- Active test executions
- Active simulated users
- Error rates by type
- Connection time metrics
- Response time / TTFB

### 2. Chaos Target Services Dashboard

**Purpose**: Monitor target service performance

**Key Panels**:
- Request rate per service
- Request duration percentiles (p50, p95, p99)
- Total requests per second (gauge)
- Traffic distribution by service
- Error rates
- Processing time metrics
- Traffic distribution by use case

### 3. Client Perspective - Load Testing

**Purpose**: Analyze system performance from the client's viewpoint

**Key Panels**:
- Client load (requests sent per second)
- End-to-end latency observed by clients
- Latency breakdown (connection, TTFB, processing)
- Error rate (client-observed)
- Success rate (client-observed)
- Active simulated users
- Response status codes
- Errors by type

**Use Case**: Understanding the user experience and client-side performance characteristics during load tests.

### 4. Service Perspective - Load Testing

**Purpose**: Analyze system performance from the server's viewpoint

**Key Panels**:
- Service load (requests received per second)
- Request duration observed by services
- Processing breakdown (connection, processing)
- Error rate (service-observed)
- Success rate (service-observed)
- Total service load
- Response status codes
- Traffic distribution by use case
- Latency comparison by instance

**Use Case**: Understanding server-side performance, bottlenecks, and capacity during load tests.

## Metrics Available

### Agent Metrics (Scraped from :9096, :9097)

- `chaos_agent_requests_total` - Total requests made
- `chaos_agent_request_duration_seconds` - Request duration histogram
- `chaos_agent_errors_total` - Total errors encountered
- `chaos_agent_connection_time_seconds` - Connection time histogram
- `chaos_agent_response_time_seconds` - Time to first byte histogram
- `chaos_agent_processing_time_seconds` - Total processing time histogram
- `chaos_agent_success_total` - Successful requests counter
- `chaos_agent_active_tests` - Currently active test executions
- `chaos_agent_active_users` - Active simulated users per test

**Labels**: `use_case`, `method`, `status_code`, `connection_pooled`, `test_execution_id`

### Target Service Metrics (Scraped from :9090, :9093, :9094)

- `chaos_target_http_requests_total` - Total requests received
- `chaos_target_http_request_duration_seconds` - Request duration histogram
- `chaos_target_http_connection_time_seconds` - Connection time histogram
- `chaos_target_http_processing_time_seconds` - Processing time histogram
- `chaos_target_http_errors_total` - Total errors
- `chaos_target_http_success_total` - Successful requests counter

**Labels**: `use_case`, `method`, `status_code`, `error_type`

## Prometheus Configuration

Prometheus is configured to scrape metrics every **30 seconds** from:

- `target-http:9090` - HTTP service (non-TLS)
- `target-http-tls:9093` - HTTPS service
- `target-http2:9094` - HTTP/2 service
- `chaos-agent-1:9096` - Agent 1 metrics
- `chaos-agent-2:9097` - Agent 2 metrics

Configuration file: `monitoring/prometheus-docker.yml`

## Troubleshooting

### Services not starting

Check container logs:
```powershell
make docker-logs
```

View specific service logs:
```powershell
docker-compose -f docker-compose-full-stack.yml logs chaos-master
docker-compose -f docker-compose-full-stack.yml logs chaos-agent-1
```

### Client cannot connect to master

Ensure the master is running:
```powershell
docker ps | findstr chaos-master
```

Test connectivity:
```powershell
curl http://localhost:9002
```

### Prometheus not scraping metrics

1. Check Prometheus targets: http://localhost:9091/targets
2. All targets should show as "UP"
3. If targets are down, check that services are running

### Grafana dashboards not showing data

1. Verify Prometheus datasource is configured: http://localhost:3000/datasources
2. Check that Prometheus has data: http://localhost:9091/graph
3. Ensure metrics are being collected (run some load tests)

### TLS certificate issues

Regenerate certificates:
```powershell
rmdir /s /q certs\target-http
make generate-certs
make docker-up
```

## Load Testing Workflow

### 1. Start the Environment

```powershell
make docker-up
```

### 2. Prepare Test Scenarios

Modify test parameters in the client or use the provided use cases (UC1-UC11).

### 3. Execute Load Tests

```powershell
make run-client
```

### 4. Monitor in Real-Time

- Open Grafana: http://localhost:3000
- View "Client Perspective" dashboard during test execution
- Monitor "Service Perspective" dashboard for server performance

### 5. Analyze Results

Compare metrics between perspectives:
- **Client-observed latency** vs **Service-observed latency**
- **Request rates** (sent vs received)
- **Error rates** from both sides

### 6. Capture Evidence

Take screenshots of:
- Client Perspective dashboard
- Service Perspective dashboard
- Specific panels showing anomalies or interesting patterns
- Prometheus queries for detailed analysis

### 7. Stop and Clean Up

```powershell
make stop-full-stack
```

## Advanced Configuration

### Scaling Agents

To add more agents, edit `docker-compose-full-stack.yml`:

```yaml
chaos-agent-3:
  build:
    context: .
    dockerfile: build/dockerfile-chaos-agent
  container_name: chaos-agent-3
  networks:
    - chaos-network
  ports:
    - "9098:9098"
  environment:
    - CHAOS_MASTER_ADDRESS=chaos-master:9002
    - CHAOS_AGENT_METRICS_PORT=9098
```

Add to Prometheus configuration in `monitoring/prometheus-docker.yml`.

### Custom Dashboards

1. Create dashboard in Grafana UI
2. Export JSON (Dashboard Settings → JSON Model)
3. Save to `monitoring/grafana-docker/provisioning/dashboards/`
4. Restart Grafana: `docker-compose -f docker-compose-full-stack.yml restart grafana`

### Data Retention

Edit Prometheus retention in `docker-compose-full-stack.yml`:

```yaml
command:
  - '--storage.tsdb.retention.time=30d'  # Change from 7d to 30d
```

## Files and Directories

```
.
├── docker-compose-full-stack.yml     # Main Docker Compose file
├── build/
│   ├── dockerfile-chaos-master       # Master Dockerfile
│   ├── dockerfile-chaos-agent        # Agent Dockerfile
│   └── dockerfile-target-http        # Target service Dockerfile
├── monitoring/
│   ├── prometheus-docker.yml         # Prometheus config for Docker
│   └── grafana-docker/
│       └── provisioning/
│           ├── datasources/          # Grafana datasource config
│           └── dashboards/           # Grafana dashboard JSONs
├── certs/
│   └── target-http/                  # TLS certificates
└── Makefile                          # Build and run targets
```

## Comparison: Docker vs Local Setup

| Feature | Docker Setup | Local Setup |
|---------|-------------|-------------|
| Startup | `make docker-up` | `make run-test-setup` |
| Networking | Docker network | localhost |
| Monitoring | Included | Requires separate setup |
| Isolation | Full isolation | Shared resources |
| Cleanup | `make stop-full-stack` | `make stop-all` |
| Portability | Highly portable | OS-dependent |
| Resource Usage | Higher (containers) | Lower (native processes) |
| Best For | Production-like testing | Quick development |

## Best Practices

1. **Always generate certificates before first run**: `make generate-certs`
2. **Monitor resource usage**: Docker Desktop → Settings → Resources
3. **Regular cleanup**: Remove unused volumes with `docker volume prune`
4. **Take screenshots during tests**: Evidence for documentation
5. **Export Grafana dashboards**: Backup custom configurations
6. **Check logs regularly**: `make docker-logs` for troubleshooting
7. **Test incrementally**: Start with small load, increase gradually

## Next Steps

- Review the [CONFIGURATION.md](CONFIGURATION.md) for detailed configuration options
- See [KUBERNETES.md](KUBERNETES.md) for Kubernetes deployment
- Check [readme.md](readme.md) for overall project documentation
