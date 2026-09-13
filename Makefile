.PHONY: all master agent client target-http target-grpc clean proto tools run-server run-client run-both run-master run-agent run-target-http run-target-grpc run-target-grpc-tls run-target-http-tls run-target-http-http2 run-target-http-notls run-test-setup stop-all generate-certs monitoring-up monitoring-down monitoring-logs monitoring-restart run-full-stack docker-build docker-up docker-down docker-logs docker-full-stack stop-full-stack docker-restart

# Build all components
all: master agent client target-http

# Protobuf toolchain (override on the command line, e.g. PROTOC=/path/to/protoc)
PROTOC ?= protoc
PROTO_DIR := internal/contracts
PROTO_FLAGS := --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

# Generate protobuf files (only regenerated when the .proto file is newer).
# Prerequisites: protoc on PATH plus the Go plugins installed via `make tools`.
proto: $(PROTO_DIR)/master.pb.go $(PROTO_DIR)/target.pb.go

# Grouped targets: each protoc invocation produces both the message and gRPC files.
$(PROTO_DIR)/master.pb.go $(PROTO_DIR)/master_grpc.pb.go &: $(PROTO_DIR)/master.proto
	$(PROTOC) $(PROTO_FLAGS) $<

$(PROTO_DIR)/target.pb.go $(PROTO_DIR)/target_grpc.pb.go &: $(PROTO_DIR)/target.proto
	$(PROTOC) $(PROTO_FLAGS) $<

# Install the Go protobuf compiler plugins used by `make proto`.
# The versions are pinned to match the generated code already checked into the repo.
tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	@echo Installed protoc-gen-go and protoc-gen-go-grpc.
	@echo Make sure the Go bin directory is on your PATH: run 'go env GOPATH' and add its bin subdirectory.

# Build the master binary
master: proto
	go build -o bin/chaos-master ./cmd/master

# Build the agent binary
agent: proto
	go build -o bin/chaos-agent ./cmd/agent

# Build the client binary
client: proto
	go build -o bin/chaos-client ./cmd/client

# Build the target-http binary
target-http: proto
	go build -o bin/target-http ./cmd/target-http

# Build the target-grpc binary
target-grpc: proto
	go build -o bin/target-grpc ./cmd/target-grpc

# Run the master server
run-server: master
	@echo Starting ChaosMaster server...
	.\bin\chaos-master

# Run the client
run-client: client
	@echo Starting ChaosMaster client...
	.\bin\chaos-client

# Run both server and client (server in background, client in foreground)
run-both: master client
	@echo Starting ChaosMaster server in background...
	@powershell -Command "Start-Process -FilePath '.\bin\chaos-master' -WindowStyle Normal"
	@timeout /t 2 /nobreak >nul
	@echo Starting ChaosMaster client...
	.\bin\chaos-client

# Clean built binaries and generated files
clean:
ifeq ($(OS),Windows_NT)
	if exist bin rmdir /s /q bin
	del /q internal\contracts\*.pb.go 2>nul || exit 0
else
	rm -rf bin/
	rm -f internal/contracts/*.pb.go
endif

# Run the master server (foreground)
run-master: master
	@echo Starting ChaosMaster server...
	.\bin\chaos-master

# Run a single agent (foreground)
run-agent: agent
	@echo Starting Chaos Agent...
	.\bin\chaos-agent

# Run the target-http service (foreground)
run-target-http: target-http
	@echo Starting Target HTTP Service...
	.\bin\target-http

# Run the target-http service with TLS (foreground)
run-target-http-tls: target-http
	@echo Starting Target HTTP Service with TLS...
	@if not exist certs\target-http mkdir certs\target-http
	@if not exist certs\target-http\tls.crt $(MAKE) generate-certs
	@powershell -Command "$$env:CHAOS_TARGET_PORT=8443; $$env:CHAOS_TARGET_CERT_FILE='certs\target-http\tls.crt'; $$env:CHAOS_TARGET_KEY_FILE='certs\target-http\tls.key'; & '.\bin\target-http'"

# Run the target-http service with HTTP/2 and TLS (foreground)
run-target-http-http2: target-http
	@echo Starting Target HTTP/2 Service with TLS...
	@if not exist certs\target-http mkdir certs\target-http
	@if not exist certs\target-http\tls.crt $(MAKE) generate-certs
	@powershell -Command "$$env:CHAOS_TARGET_PORT=8444; $$env:CHAOS_TARGET_PROTOCOL='http2'; $$env:CHAOS_TARGET_CERT_FILE='certs\target-http\tls.crt'; $$env:CHAOS_TARGET_KEY_FILE='certs\target-http\tls.key'; & '.\bin\target-http'"

# Run the target-http service without TLS (foreground)
run-target-http-notls: target-http
	@echo Starting Target HTTP Service without TLS...
	.\bin\target-http

# Run the target-grpc service (foreground)
run-target-grpc: target-grpc
	@echo Starting Target gRPC Service...
	.\bin\target-grpc

# Run the target-grpc service with TLS (foreground)
run-target-grpc-tls: target-grpc
	@echo Starting Target gRPC Service with TLS...
	@if not exist certs\target-http mkdir certs\target-http
	@if not exist certs\target-http\tls.crt $(MAKE) generate-certs
	@powershell -Command "$$env:CHAOS_TARGET_PORT=9000; $$env:CHAOS_TARGET_CERT_FILE='certs\target-http\tls.crt'; $$env:CHAOS_TARGET_KEY_FILE='certs\target-http\tls.key'; & '.\bin\target-grpc'"

# Generate TLS certificates for target-http service
generate-certs:
	@echo Generating TLS certificates for target-http...
	@if not exist certs\target-http mkdir certs\target-http
	@powershell -Command "& { \
		$$cert = New-SelfSignedCertificate -DnsName 'localhost', 'target-http' -CertStoreLocation 'cert:\CurrentUser\My'; \
		$$pwd = ConvertTo-SecureString -String 'password' -Force -AsPlainText; \
		$$path = 'cert:\CurrentUser\My\' + $$cert.Thumbprint; \
		Export-PfxCertificate -Cert $$path -FilePath 'certs\target-http\cert.pfx' -Password $$pwd; \
		$$cert | Remove-Item; \
	}"
	@openssl pkcs12 -in certs\target-http\cert.pfx -nocerts -out certs\target-http\tls.key -nodes -passin pass:password
	@openssl pkcs12 -in certs\target-http\cert.pfx -clcerts -nokeys -out certs\target-http\tls.crt -passin pass:password
	@echo Certificates generated in certs\target-http\
	@echo   - tls.crt (certificate)
	@echo   - tls.key (private key)

# Run test setup: master + 2 agents + target-http in background
run-test-setup: master agent target-http
	@echo === Starting Test Environment ===
	@if not exist certs\target-http mkdir certs\target-http
	@if not exist certs\target-http\tls.crt $(MAKE) generate-certs
	@echo Starting Target HTTP Service (non-TLS) on port 8080 in background...
	@powershell -Command "Start-Process -FilePath '.\bin\target-http' -WindowStyle Normal"
	@timeout /t 2 /nobreak >nul
	@echo Starting Target HTTP Service (TLS) on port 8443 in background...
	@powershell -NoProfile -Command "$$env:CHAOS_TARGET_PORT=8443; $$env:CHAOS_TARGET_METRICS_PORT=9093; $$env:CHAOS_TARGET_CERT_FILE='certs\target-http\tls.crt'; $$env:CHAOS_TARGET_KEY_FILE='certs\target-http\tls.key'; Start-Process -FilePath '.\bin\target-http'"
	@timeout /t 2 /nobreak >nul
	@echo Starting Target HTTP/2 Service (TLS) on port 8444 in background...
	@powershell -NoProfile -Command "$$env:CHAOS_TARGET_PORT=8444; $$env:CHAOS_TARGET_METRICS_PORT=9094; $$env:CHAOS_TARGET_PROTOCOL='http2'; $$env:CHAOS_TARGET_CERT_FILE='certs\target-http\tls.crt'; $$env:CHAOS_TARGET_KEY_FILE='certs\target-http\tls.key'; Start-Process -FilePath '.\bin\target-http'"
	@timeout /t 2 /nobreak >nul
	@echo Starting ChaosMaster server in background...
	@powershell -Command "Start-Process -FilePath '.\bin\chaos-master' -WindowStyle Normal"
	@timeout /t 3 /nobreak >nul
	@echo Starting Agent 1 in background...
	@powershell -Command "Start-Process -FilePath '.\bin\chaos-agent' -WindowStyle Normal"
	@timeout /t 2 /nobreak >nul
	@echo Starting Agent 2 in background...
	@powershell -Command "Start-Process -FilePath '.\bin\chaos-agent' -WindowStyle Normal"
	@echo.
	@echo === Test Environment Running ===
	@echo Target HTTP (non-TLS): Running on http://localhost:8080
	@echo Target HTTP (TLS): Running on https://localhost:8443
	@echo Target HTTP/2 (TLS): Running on https://localhost:8444
	@echo Master: Running on localhost:9002
	@echo Agent 1: Connected
	@echo Agent 2: Connected
	@echo.
	@echo Use 'make stop-all' to stop all processes
	@echo Press Ctrl+C to exit this terminal

# Stop all chaos processes
stop-all:
	@echo Stopping all Chaos processes...
	@powershell -Command "Get-Process | Where-Object {$$_.ProcessName -like 'chaos-*' -or $$_.ProcessName -like 'target-*'} | Stop-Process -Force" 2>nul || echo No processes found
	@echo All processes stopped.

# Start Prometheus and Grafana monitoring stack
monitoring-up:
	@echo Starting Prometheus and Grafana...
	@docker-compose -f docker-compose-monitoring.yml up -d
	@echo.
	@echo === Monitoring Stack Started ===
	@echo Prometheus: http://localhost:9091
	@echo Grafana: http://localhost:3000 (admin/admin)
	@echo.

# Stop Prometheus and Grafana monitoring stack
monitoring-down:
	@echo Stopping monitoring stack...
	@docker-compose -f docker-compose-monitoring.yml down
	@echo Monitoring stack stopped.

# Restart Prometheus and Grafana monitoring stack
monitoring-restart:
	@echo Restarting monitoring stack...
	@docker-compose -f docker-compose-monitoring.yml down
	@docker-compose -f docker-compose-monitoring.yml up -d
	@echo Monitoring stack restarted.
	@echo Prometheus: http://localhost:9091
	@echo Grafana: http://localhost:3000 (admin/admin)

# View monitoring stack logs
monitoring-logs:
	@docker-compose -f docker-compose-monitoring.yml logs -f

# Start full stack: test environment + monitoring
run-full-stack: master agent target-http target-grpc
	@echo === Starting Full Chaos Generator Stack ===
	@if not exist certs\target-http mkdir certs\target-http
	@if not exist certs\target-http\tls.crt $(MAKE) generate-certs
	@echo.
	@echo [1/4] Starting monitoring stack...
	@docker-compose -f docker-compose-monitoring.yml up -d
	@timeout /t 5 /nobreak >nul
	@echo.
	@echo [2/4] Starting target services...
	@echo   - Target HTTP (non-TLS) on port 8080
	@powershell -Command "Start-Process -FilePath '.\bin\target-http' -WindowStyle Normal"
	@timeout /t 2 /nobreak >nul
	@echo   - Target HTTP (TLS) on port 8443
	@powershell -NoProfile -Command "$$env:CHAOS_TARGET_PORT=8443; $$env:CHAOS_TARGET_METRICS_PORT=9093; $$env:CHAOS_TARGET_CERT_FILE='certs\target-http\tls.crt'; $$env:CHAOS_TARGET_KEY_FILE='certs\target-http\tls.key'; Start-Process -FilePath '.\bin\target-http'"
	@timeout /t 2 /nobreak >nul
	@echo   - Target HTTP/2 (TLS) on port 8444
	@powershell -NoProfile -Command "$$env:CHAOS_TARGET_PORT=8444; $$env:CHAOS_TARGET_METRICS_PORT=9094; $$env:CHAOS_TARGET_PROTOCOL='http2'; $$env:CHAOS_TARGET_CERT_FILE='certs\target-http\tls.crt'; $$env:CHAOS_TARGET_KEY_FILE='certs\target-http\tls.key'; Start-Process -FilePath '.\bin\target-http'"
	@timeout /t 2 /nobreak >nul
	@echo   - Target gRPC (TLS) on port 9000
	@powershell -NoProfile -Command "$$env:CHAOS_TARGET_PORT=9000; $$env:CHAOS_TARGET_METRICS_PORT=9095; $$env:CHAOS_TARGET_CERT_FILE='certs\target-http\tls.crt'; $$env:CHAOS_TARGET_KEY_FILE='certs\target-http\tls.key'; Start-Process -FilePath '.\bin\target-grpc' -WindowStyle Normal"
	@timeout /t 2 /nobreak >nul
	@echo.
	@echo [3/4] Starting Chaos Master...
	@powershell -Command "Start-Process -FilePath '.\bin\chaos-master' -WindowStyle Normal"
	@timeout /t 3 /nobreak >nul
	@echo.
	@echo [4/4] Starting Chaos Agents...
	@echo   - Agent 1 (metrics on port 9096)
	@powershell -NoProfile -Command "$$env:CHAOS_AGENT_METRICS_PORT=9096; Start-Process -FilePath '.\bin\chaos-agent' -WindowStyle Normal"
	@timeout /t 2 /nobreak >nul
	@echo   - Agent 2 (metrics on port 9097)
	@powershell -NoProfile -Command "$$env:CHAOS_AGENT_METRICS_PORT=9097; Start-Process -FilePath '.\bin\chaos-agent' -WindowStyle Normal"
	@timeout /t 2 /nobreak >nul
	@echo.
	@echo ================================================================
	@echo === Full Stack Running ===
	@echo ================================================================
	@echo.
	@echo Services:
	@echo   Target HTTP (non-TLS):  http://localhost:8080 (metrics: 9090)
	@echo   Target HTTP (TLS):      https://localhost:8443 (metrics: 9093)
	@echo   Target HTTP/2 (TLS):    https://localhost:8444 (metrics: 9094)
	@echo   Target gRPC (TLS):      localhost:9000 (metrics: 9095)
	@echo   Master:                 localhost:9002
	@echo   Agent 1:                Connected (metrics: 9096)
	@echo   Agent 2:                Connected (metrics: 9097)
	@echo.
	@echo Monitoring:
	@echo   Prometheus:             http://localhost:9091
	@echo   Grafana:                http://localhost:3000 (admin/admin)
	@echo.
	@echo Commands:
	@echo   Stop services:          make stop-all
	@echo   Stop monitoring:        make monitoring-down
	@echo   View monitoring logs:   make monitoring-logs
	@echo.
	@echo ================================================================
