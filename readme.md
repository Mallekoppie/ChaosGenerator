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

### Local Development

#### Run the master server

```bash
make run-master
```

This builds and starts the ChaosMaster gRPC server on port 9002.

#### Run a single agent

```bash
make run-agent
```

This builds and starts a ChaosAgent that connects to the master at localhost:9002.

#### Run the full local stack in the background

```bash
make start
```

Builds everything, generates certificates and starts the master, `AGENT_COUNT`
(default 2) agents and the target services in the background. Each process logs
to `.run/<name>.log`, and the command prints a per-service status table and the
web UI URL.

```bash
make status    # what is running and on which ports
make logs      # follow the background logs
make stop      # stop everything
make restart   # stop then start again
```

Default ports (override on the command line, for example
`make start TARGET_HTTP_PORT=8085 MASTER_WEB_ADDR=0.0.0.0:9100`):

| Service | Default |
|---------|---------|
| Master gRPC | 9002 |
| Master web UI / gRPC-Web | 9001 |
| Target HTTP | 8080 |
| Target HTTPS | 8443 |
| Target HTTP/2 | 8444 |
| Target gRPC | 9000 |

### Run the test client

```bash
make run-client
```

This builds and starts the interactive test client. It needs a master that is
already running (`make start`, or `make run-master` in another terminal).

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

### Run locally

```bash
# Start the master (native gRPC on 9002, web + gRPC-Web on 9001)
make master
./bin/chaos-master
# open http://localhost:9001/
```

The first user is created from the **Register** screen using the registration
secret (`CHAOS_REGISTER_SECRET`, default `chaos-dev-secret` for development).
The JWT signing key is configured with `CHAOS_JWT_SIGNING_KEY`.

For Flutter hot reload during development, run the master with gRPC-Web only and
point the app at it:

```bash
make run-master-dev
cd web && flutter run -d chrome --dart-define=CHAOS_MASTER_URL=http://localhost:9001
```

### Kubernetes

```bash
kubectl port-forward svc/chaos-master-service 8080:8080
# open http://localhost:8080/
```

See [KUBERNETES.md](KUBERNETES.md) for details.

## Monitoring

Prometheus and Grafana come up with a single command:

```bash
make monitoring-up
```

- If a Docker daemon is reachable it runs `docker-compose-monitoring.yml`.
- Otherwise it downloads Prometheus and Grafana into `.monitoring/` and runs them
  as background processes (useful inside dev containers where Docker can't run).

| Service | URL | Login |
|---------|-----|-------|
| Prometheus | http://localhost:9091 | - |
| Grafana | http://localhost:3000 | admin/admin |

Override the ports with `PROMETHEUS_PORT` / `GRAFANA_PORT` if they are taken.
`make run-full-stack` starts the services and monitoring together; `make monitoring-down`
stops monitoring only. See [monitoring/README.md](monitoring/README.md) for details.

## Configuration

### Agent Configuration

Agents read configuration from environment variables with fallback to local development defaults:

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `CHAOS_MASTER_ADDRESS` | `localhost:9002` | Master server address |
| `CHAOS_AGENT_PORT` | `9001` | Agent port (for future use) |
| `CHAOS_AGENT_METRICS_PORT` | `9091` | Metrics port |
| `CHAOS_AGENT_VERSION` | `2.0.0` | Agent version |

#### Local Development
Simply run the agent - it will use default values:
```bash
make run-agent
```

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
2. Agent establishes bidirectional stream with master
3. Agent sends periodic heartbeats to maintain connection
4. Master sends commands down the stream (AddTests, StartTest, etc.)
5. Agent executes commands and sends responses back

### Benefits of This Architecture

- **Firewall Friendly**: Agents only make outbound connections
- **Cloud Native**: Agents can be behind NAT or in different clusters
- **Dynamic Scaling**: New agents automatically register and connect
- **Real-time Control**: Master knows which agents are online
- **No Agent Ports**: Agents don't expose any listening ports
