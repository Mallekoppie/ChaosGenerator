# Introduction

Used to distribute tests to agents. The agents then make http requests to the remote target.

The main purpose is to determine the maximum theoretical through of a service. The tests are repeated and not randomized at this point in time.

## Project Structure

- `cmd/master` - Main entry point for the ChaosMaster service (native gRPC on `9002`, web + gRPC-Web on `8080`)
- `cmd/agent` - Main entry point for the ChaosAgent service
- `internal/master/` - Master service implementation
- `internal/agent/` - Agent service implementation
- `internal/contracts/` - Protocol buffer definitions for communication (master.proto, target.proto, auth.proto)
- `internal/master/web/` - gRPC-Web + embedded web app HTTP server
- `web/` - Flutter web control panel (talks to the master over gRPC-Web)
- `manifests/` - Kubernetes deployment manifests
  - `manifests/server/` - Master server Kubernetes manifests
  - `manifests/agent/` - Agent Kubernetes manifests
  - `manifests/target-http/` - Target HTTP/HTTPS/HTTP2 Kubernetes manifests
  - `manifests/kind/` - Self-contained manifests for a local kind cluster
  - `manifests/generated/` - Rendered manifests produced by `make manifests` (gitignored)

## Build

### Build all components

```bash
make all
```

### Build master only

```bash
make master
```

### Build agent only

```bash
make agent
```

### Generate protobuf files

```bash
make proto
```

Or manually:

```bash
# Generate agent proto
protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./internal/contracts/agent.proto

# Generate master proto
protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./internal/contracts/master.proto
```

### Clean build artifacts

```bash
make clean
```

## Running

### Run the stack in kind

The whole system runs in a local kind cluster. There are no local run targets:

```bash
make kind-up     # create the cluster, build/load the images, deploy everything
make proxy       # web UI, Prometheus and Grafana on your machine
```

`make kind-up` and the proxy targets must run on the **host**: kind spawns
privileged containers, which does not work inside the dev container. The dev
container remains the place to build, test and run `make manifests`.

| Local port | Forwards to | Override with |
|------------|-------------|---------------|
| 9001 | master web UI / gRPC-Web | `WEB_UI_PORT` |
| 9091 | Prometheus | `PROMETHEUS_PORT` |
| 3000 | Grafana (admin/admin) | `GRAFANA_PORT` |
| 9002 | master gRPC (only with `make proxy-grpc`) | `GRPC_PORT` |

The web UI port is a variable because 8080 is often already taken on a
workstation. Inside the cluster the master still listens on 8080.

### Run the test client

`make all` still builds `bin/chaos-client`, but there is no run target for it.
Forward the master gRPC port and point the client at it:

```bash
make proxy-grpc
CHAOS_MASTER_ADDRESS=localhost:9002 ./bin/chaos-client
```

The client defaults to `127.0.0.1:9002`, so with `make proxy-grpc` running, plain
`./bin/chaos-client` works without any environment variable.

## Web Control Panel

The master embeds a Flutter web control panel and exposes the same services over
**gRPC-Web**, so the whole system is gRPC end to end.

### Build

```bash
# Generate the Dart protobuf/gRPC-Web stubs (requires the Dart SDK + protoc_plugin)
make proto-dart

# Build the Flutter web app and embed it into the master binary
make buildweb
```

`make buildweb` runs `flutter build web --base-href=/` and copies the result into
`cmd/master/dist`, which is embedded with `//go:embed`. Override the base href
with `make buildweb BASE_HREF=/some/path/` if you host the app elsewhere.

The web app and the gRPC-Web endpoints share a single port (default
`0.0.0.0:9001`). Change it with the `CHAOS_MASTER_WEB_ADDRESS` environment
variable, for example:

```bash
CHAOS_MASTER_WEB_ADDRESS=127.0.0.1:9100 ./bin/chaos-master
```

If the port is already in use the master exits immediately with
`Unable to start the web server; set CHAOS_MASTER_WEB_ADDRESS to a free port`
rather than running without a web UI.

> Remember to rebuild (`make master`) after changing any of the defaults in
> `cmd/master/main.go` — the Flutter app and the defaults are compiled into the
> binary.

### Open the panel

```bash
make proxy       # then open http://localhost:9001/
```

The container image sets `CHAOS_MASTER_WEB_ADDRESS` to `0.0.0.0:8080`, so in the
cluster the panel listens on 8080 and `make proxy` maps it onto your local
`WEB_UI_PORT` (9001 by default).

The first user is created from the **Register** screen using the registration
secret. Read it from the cluster with:

```bash
kubectl -n chaos-testing get secret chaos-master-secrets \
  -o jsonpath='{.data.registerSecret}' | base64 -d
```

`CHAOS_REGISTER_SECRET` defaults to `chaos-dev-secret` for development, and the
JWT signing key is configured with `CHAOS_JWT_SIGNING_KEY`.

For Flutter work, run the app locally against the cluster:

```bash
make proto-dart
make proxy                                        # web UI + gRPC-Web on 9001
cd web && flutter run -d chrome --dart-define=CHAOS_MASTER_URL=http://localhost:9001
```

See [web/README.md](web/README.md) for details.

### Test metrics drilldown

Every agent reports cumulative per-execution metrics to the master over the same
`ConnectAgent` gRPC stream it already uses for heartbeats and command responses.
The master aggregates them in memory and the panel polls `GetTestMetrics` every
3 seconds, so `/tests/metrics/<execution-id>` (reachable by clicking an execution
id on the **Tests** screen) works without Prometheus or Grafana.

Each report carries client-observed throughput, a cumulative latency histogram,
the failure demux (4xx/5xx/timeouts/connection errors/resets plus per-status and
per-gRPC codes) and worker runtime telemetry (CPU, memory, scrape count). The
drilldown renders the request rate and status demux, the round-trip latency
envelope against an SLA ceiling, a Prometheus scraper-health panel, and a
per-agent worker table.

Failure classification is deliberately transport-aware:

- an HTTP response is only a success when the status is 2xx **and** the body was
  read to completion, so a restarted peer that truncates a response mid-body is
  reported as a failure rather than a clean success;
- refused, unreachable, premature-EOF and broken-pool connections are reported as
  connection errors; resets and broken pipes are reported as resets; neither is
  folded into the 4xx/5xx buckets;
- agents report one snapshot per second. If an agent stops reporting, its derived
  egress rate decays to zero after 5 seconds and it is flagged `stale` in the
  worker table and counted in `silentAgents`, while its cumulative counters stay
  intact. A silent agent therefore never keeps looking like it is generating
  load.

Metrics are aggregated in memory, and every finished run is also persisted to
BoltDB so it survives a master restart. The master keeps the most recent
`CHAOS_MASTER_HISTORY_LIMIT` (default `50`) finished runs; older runs are evicted
from both memory and the database.

### Test history and report export

The **Tests** screen shows a **Recent runs** panel listing finished executions
with their start/end times, duration and request totals. Click an execution id to
reopen the drilldown for a past run, or use the download action on a row to
export the run straight from history — useful when the live export was missed.

Two exports are available from the drilldown (and the Markdown one from history):

- **Export report** produces a formatted Markdown document intended for
  documentation: a run overview (use case executed, test start/end time, test
  duration, number of agents), fleet totals (throughput, latency percentiles,
  failure demux and payload volume), failures by code, and a per-agent metrics
  table with a totals row across every agent — including total requests per
  agent.
- **Export JSON** produces the raw, machine-readable metrics dump.

## Monitoring

Prometheus and Grafana run inside the cluster and come up with everything else:

```bash
make proxy       # Prometheus on 9091, Grafana on 3000
```

| Service | URL | Login |
|---------|-----|-------|
| Prometheus | http://localhost:9091 | - |
| Grafana | http://localhost:3000 | admin/admin |

Override the local ports with `PROMETHEUS_PORT` / `GRAFANA_PORT` if they are
taken. Prometheus scrapes the target variants and every agent pod using the same
job names and labels as before, so the provisioned dashboards keep working. See
[monitoring/README.md](monitoring/README.md) for details.

Prometheus is optional for the in-panel telemetry: the **Tests → metrics**
drilldown reads the metrics agents push over gRPC, so it keeps working in
deployments that run without the monitoring stack.

## Configuration

### Agent Configuration

Agents read configuration from environment variables with fallback to local development defaults:

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `CHAOS_MASTER_ADDRESS` | `localhost:9002` | Master server address |
| `CHAOS_AGENT_PORT` | `9001` | Agent port (for future use) |
| `CHAOS_AGENT_METRICS_PORT` | `9091` | Metrics port |
| `CHAOS_AGENT_VERSION` | `2.0.0` | Agent version |

#### Kubernetes Deployment
Set environment variables in your deployment:
```yaml
env:
- name: CHAOS_MASTER_ADDRESS
  value: "chaos-master-service:9002"
- name: CHAOS_AGENT_VERSION
  value: "2.0.0"
```

See `manifests/agent/` and `manifests/server/` for complete Kubernetes deployment examples.

For detailed Kubernetes deployment instructions, see [manifests/README.md](manifests/README.md).

## Kubernetes Deployment

Quick start with Kubernetes:

```bash
# Deploy master server
kubectl apply -k manifests/server/

# Deploy agents
kubectl apply -k manifests/agent/

# Scale agents
kubectl scale deployment/chaos-agent --replicas=10
```

📚 See [KUBERNETES.md](KUBERNETES.md) for a complete step-by-step guide.

### Local kind cluster

This is the supported way to run the stack (`kubectl apply -k` above expects
images in a registry; the kind flow loads locally built images instead):

```bash
# kind needs podman/docker, so run these on the host
make kind-up        # create cluster + build/load images + deploy everything
make proxy          # web UI (9001), Prometheus (9091), Grafana (3000)
make kind-status    # deployments, pods and services
make kind-logs      # master, agent and target logs
make kind-verify    # agents that registered with the master
make kind-restart   # roll the deployments after rebuilding images
make kind-down      # remove the Chaos resources, keep the cluster
make kind-clean     # delete the cluster and manifests/generated
```

`make manifests` renders `manifests/kind/` into `manifests/generated/` with the
image names and tags produced by `make docker-build`. Point a test at
`http://target-http.chaos-testing.svc.cluster.local:8080`, or at the TLS, HTTP/2
and gRPC variants documented in
[manifests/kind/README.md](manifests/kind/README.md).

## Architecture

### Design Overview

The system uses a client-server architecture where:

- **Master Server**: Runs a gRPC server on port 9002
  - Manages agent connections via bidirectional streaming
  - Tracks connected agents in real-time
  - Sends commands to agents through persistent connections
  - Stores configuration in BoltDB

- **Agents**: Connect to the master server
  - Establish persistent gRPC stream connections
  - Listen for commands from master
  - Execute load tests against target APIs
  - Send heartbeats and status updates

### Communication Flow

1. Agent starts and registers with master (gets unique ID)
2. Agent establishes bidirectional stream with master and sends a heartbeat
   immediately, so it appears online without waiting for the first periodic beat
3. Agent sends periodic heartbeats to maintain connection
4. Master sends commands down the stream (AddTests, StartTest, etc.)
5. Agent executes commands and sends responses back

### Agent Liveness

The agent list is a live view of the connected agents:

- An agent row is written on registration and removed when its stream ends, so
  the UI only ever shows agents that are reachable. Starting a test fans out to
  every connected agent (or the subset named in the agent selection).
- The master reaps any connection that stops heartbeating (30s, i.e. two missed
  10s beats), which covers pods that die without closing their socket.
- Agents reconnect with exponential backoff (1s to 30s), re-registering with the
  master each time. Restarting a master therefore repopulates the fleet on its
  own without restarting the agents.

### Benefits of This Architecture

- **Firewall Friendly**: Agents only make outbound connections
- **Cloud Native**: Agents can be behind NAT or in different clusters
- **Dynamic Scaling**: New agents automatically register and connect
- **Real-time Control**: Master knows which agents are online
- **No Agent Ports**: Agents don't expose any listening ports
