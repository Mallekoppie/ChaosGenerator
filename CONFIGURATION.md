# Environment Variables Configuration Examples

## Local Development (Default)

Run the agent with defaults:
```bash
./bin/chaos-agent
```

Output will show:
```
Configuration loaded:
  Port: 9001
  Metrics Port: 9091
  Master Address: localhost:9002
  Version: 2.0.0
```

## Custom Configuration for Testing

### Linux
```bash
export CHAOS_MASTER_ADDRESS="localhost:9003"
export CHAOS_AGENT_VERSION="2.1.0"
./bin/chaos-agent
```

Or override for a single command without exporting:
```bash
CHAOS_MASTER_ADDRESS=localhost:9003 CHAOS_AGENT_VERSION=2.1.0 ./bin/chaos-agent
```

## Master Configuration

| Environment Variable | Default | Description |
|----------------------|---------|-------------|
| `CHAOS_MASTER_GRPC_ADDRESS` | `0.0.0.0:9002` | Native gRPC listener used by agents |
| `CHAOS_MASTER_WEB_ADDRESS` | `0.0.0.0:9001` | Web control panel + gRPC-Web listener |
| `CHAOS_MASTER_DB_PATH` | `chaos_master.db` | BoltDB file (`/data/...` in containers) |
| `CHAOS_JWT_SIGNING_KEY` | development value | Signs the local JWTs issued to the web UI |
| `CHAOS_REGISTER_SECRET` | `chaos-dev-secret` | Secret required to register the first user |
| `DEBUG` | unset | `true` serves gRPC-Web only (no embedded UI) |

Example: run the master with the web control panel on a different port:
```bash
CHAOS_MASTER_WEB_ADDRESS=0.0.0.0:9100 ./bin/chaos-master
```

## Kubernetes Deployment

### Using Direct Environment Variables
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chaos-agent
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: chaos-agent
        image: chaos-agent:latest
        env:
        - name: CHAOS_MASTER_ADDRESS
          value: "chaos-master-service:9002"
        - name: CHAOS_AGENT_VERSION
          value: "2.0.0"
```

### Using ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: chaos-agent-config
data:
  CHAOS_MASTER_ADDRESS: "chaos-master-service:9002"
  CHAOS_AGENT_VERSION: "2.0.0"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chaos-agent
spec:
  template:
    spec:
      containers:
      - name: chaos-agent
        image: chaos-agent:latest
        envFrom:
        - configMapRef:
            name: chaos-agent-config
```

### Using Secrets for Sensitive Data
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: chaos-agent-secret
stringData:
  CHAOS_MASTER_ADDRESS: "chaos-master-service.production:9002"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chaos-agent
spec:
  template:
    spec:
      containers:
      - name: chaos-agent
        image: chaos-agent:latest
        env:
        - name: CHAOS_MASTER_ADDRESS
          valueFrom:
            secretKeyRef:
              name: chaos-agent-secret
              key: CHAOS_MASTER_ADDRESS
```

## Docker Compose Example

```yaml
version: '3.8'
services:
  chaos-master:
    image: chaos-master:latest
    ports:
      - "9002:9002"
    volumes:
      - ./data:/data
  
  chaos-agent-1:
    image: chaos-agent:latest
    environment:
      CHAOS_MASTER_ADDRESS: "chaos-master:9002"
      CHAOS_AGENT_VERSION: "2.0.0"
    depends_on:
      - chaos-master
  
  chaos-agent-2:
    image: chaos-agent:latest
    environment:
      CHAOS_MASTER_ADDRESS: "chaos-master:9002"
      CHAOS_AGENT_VERSION: "2.0.0"
    depends_on:
      - chaos-master
```

## Testing Different Configurations

### Test with a custom master address
```bash
# Terminal 1 - start the master
./bin/chaos-master

# Terminal 2 - start an agent pointing at a specific master
CHAOS_MASTER_ADDRESS=192.168.1.100:9002 ./bin/chaos-agent
```

### Test with version tracking
```bash
CHAOS_AGENT_VERSION="dev-$(git rev-parse --short HEAD)" ./bin/chaos-agent
```

## Verification

After starting the agent, verify the configuration is loaded correctly by checking the logs:
```
Configuration loaded:
  Port: 9001
  Metrics Port: 9091
  Master Address: chaos-master-service:9002
  Version: 2.0.0
Chaos Agent starting...
Master address: chaos-master-service:9002
```
