# Introduction

Used to distribute tests to agents. The agents then make http requests to the remote target.

The main purpose is to determine the maximum theoretical through of a service. The tests are repeated and not randomized at this point in time.

## Project Structure

- `cmd/master` - Main entry point for the ChaosMaster service
- `cmd/agent` - Main entry point for the ChaosAgent service
- `internal/master/` - Master service implementation
- `internal/agent/` - Agent service implementation
- `internal/contracts/` - Protocol buffer definitions for communication (agent.proto, master.proto)
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

#### Run complete test environment

```bash
make run-test-setup
```

This starts the master server and 2 agents in separate windows for testing.

#### Stop all running processes

```bash
make stop-all
```

Stops all chaos-master and chaos-agent processes.

### Run the test client

```bash
make run-client
```

This builds and starts the interactive test client that connects to the master server.

### Run both server and client together

```bash
make run-both
```

This starts the server in a separate window and then launches the client in the foreground, allowing you to interact with the server immediately.

**Note**: When using `make run-both`, the server runs in a separate window. To stop the server, close its window or use Task Manager.

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
