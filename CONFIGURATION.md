# Environment Variables Configuration Examples

## Local Development (Default)

Run agent with defaults:
```bash
.\bin\chaos-agent.exe
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

### Windows PowerShell
```powershell
$env:CHAOS_MASTER_ADDRESS = "localhost:9003"
$env:CHAOS_AGENT_VERSION = "2.1.0"
.\bin\chaos-agent.exe
```

### Windows CMD
```cmd
set CHAOS_MASTER_ADDRESS=localhost:9003
set CHAOS_AGENT_VERSION=2.1.0
.\bin\chaos-agent.exe
```

### Linux/Mac
```bash
export CHAOS_MASTER_ADDRESS="localhost:9003"
export CHAOS_AGENT_VERSION="2.1.0"
./bin/chaos-agent
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

### Test with custom master address
```powershell
# Terminal 1 - Start master on custom port
# (requires master code changes to support custom port)
.\bin\chaos-master.exe

# Terminal 2 - Start agent connecting to master
$env:CHAOS_MASTER_ADDRESS = "192.168.1.100:9002"
.\bin\chaos-agent.exe
```

### Test with version tracking
```powershell
$env:CHAOS_AGENT_VERSION = "dev-$(git rev-parse --short HEAD)"
.\bin\chaos-agent.exe
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
