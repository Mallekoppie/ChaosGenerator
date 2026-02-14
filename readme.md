# Introduction

Used to distribute tests to agents. The agents then make http requests to the remote target.

The main purpose is to determine the maximum theoretical through of a service. The tests are repeated and not randomized at this point in time.

## Project Structure

- `cmd/master` - Main entry point for the ChaosMaster service
- `cmd/agent` - Main entry point for the ChaosAgent service
- `internal/master/` - Master service implementation
- `internal/agent/` - Agent service implementation
- `internal/contracts/` - Protocol buffer definitions for communication (agent.proto, master.proto)

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

### Run the master server

```bash
make run-server
```

This builds and starts the ChaosMaster gRPC server on port 9002.

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
