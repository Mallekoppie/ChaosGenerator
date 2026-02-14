# Introduction

Used to distribute tests to agents. The agents then make http requests to the remote target.

The main purpose is to determine the maximum theoretical through of a service. The tests are repeated and not randomized at this point in time.

## Project Structure

- `cmd/master` - Main entry point for the ChaosMaster service
- `cmd/agent` - Main entry point for the ChaosAgent service
- `ChaosMaster/` - Master service implementation
- `ChaosAgent/` - Agent service implementation
- `contract/` - Protocol buffer definitions for communication

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
protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./contract/chaos.proto
```

### Clean build artifacts

```bash
make clean
```