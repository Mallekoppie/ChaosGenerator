# Docker Build Files

This directory contains Dockerfiles for building all components of the ChaosGenerator system.

## Available Dockerfiles

### 1. dockerfile-target-http
Builds the HTTP target service that implements all 11 use cases for load testing.

**Build:**
```bash
docker build -f build/dockerfile-target-http -t chaos-target-http:latest .
```

**Run:**
```bash
docker run -p 8080:8080 -p 9090:9090 \
  -e CHAOS_TARGET_PROTOCOL=http1 \
  -e CHAOS_TARGET_PORT=8080 \
  -e CHAOS_TARGET_METRICS_PORT=9090 \
  chaos-target-http:latest
```

### 2. dockerfile-chaos-master
Builds the Chaos Master server that coordinates agents and manages test executions.

**Build:**
```bash
docker build -f build/dockerfile-chaos-master -t chaos-master:latest .
```

**Run:**
```bash
docker run -p 9002:9002 \
  -v chaos-master-data:/data \
  -e CHAOS_MASTER_PORT=9002 \
  chaos-master:latest
```

### 3. dockerfile-chaos-agent
Builds the Chaos Agent that executes tests against target services.

**Build:**
```bash
docker build -f build/dockerfile-chaos-agent -t chaos-agent:latest .
```

**Run:**
```bash
docker run -p 9091:9091 \
  -e CHAOS_MASTER_ADDRESS=chaos-master:9002 \
  -e CHAOS_AGENT_PORT=8081 \
  -e CHAOS_AGENT_METRICS_PORT=9091 \
  -e CHAOS_AGENT_VERSION=1.0.0 \
  chaos-agent:latest
```

### 4. dockerfile-target-grpc
Placeholder for the gRPC target service (to be implemented).

## Building All Images

From the project root:

```bash
# Build target-http
docker build -f build/dockerfile-target-http -t chaos-target-http:latest .

# Build chaos-master
docker build -f build/dockerfile-chaos-master -t chaos-master:latest .

# Build chaos-agent
docker build -f build/dockerfile-chaos-agent -t chaos-agent:latest .
```

## Docker Compose

For local testing, you can use Docker Compose (create docker-compose.yml in project root):

```yaml
version: '3.8'

services:
  chaos-master:
    image: chaos-master:latest
    ports:
      - "9002:9002"
    volumes:
      - chaos-data:/data
    environment:
      - CHAOS_MASTER_PORT=9002

  chaos-agent:
    image: chaos-agent:latest
    depends_on:
      - chaos-master
    environment:
      - CHAOS_MASTER_ADDRESS=chaos-master:9002
      - CHAOS_AGENT_VERSION=1.0.0
    deploy:
      replicas: 3

  target-http:
    image: chaos-target-http:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - CHAOS_TARGET_PROTOCOL=http1

volumes:
  chaos-data:
```

Start with:
```bash
docker-compose up -d
```

## Kubernetes Deployment

For Kubernetes deployment, use the manifests in the `manifests/` directory.
