.PHONY: all master agent client target-http clean proto run-server run-client run-both run-master run-agent run-target-http run-target-http-tls run-target-http-notls run-test-setup stop-all generate-certs

# Build all components
all: master agent client target-http

# Generate protobuf files
proto:
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./internal/contracts/master.proto

# Build the master binary
master: proto
	go build -o bin/chaos-master.exe ./cmd/master

# Build the agent binary
agent: proto
	go build -o bin/chaos-agent.exe ./cmd/agent

# Build the client binary
client: proto
	go build -o bin/chaos-client.exe ./cmd/client

# Build the target-http binary
target-http: proto
	go build -o bin/target-http.exe ./cmd/target-http

# Run the master server
run-server: master
	@echo Starting ChaosMaster server...
	.\bin\chaos-master.exe

# Run the client
run-client: client
	@echo Starting ChaosMaster client...
	.\bin\chaos-client.exe

# Run both server and client (server in background, client in foreground)
run-both: master client
	@echo Starting ChaosMaster server in background...
	@powershell -Command "Start-Process -FilePath '.\bin\chaos-master.exe' -WindowStyle Normal"
	@timeout /t 2 /nobreak >nul
	@echo Starting ChaosMaster client...
	.\bin\chaos-client.exe

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
	.\bin\chaos-master.exe

# Run a single agent (foreground)
run-agent: agent
	@echo Starting Chaos Agent...
	.\bin\chaos-agent.exe

# Run the target-http service (foreground)
run-target-http: target-http
	@echo Starting Target HTTP Service...
	.\bin\target-http.exe

# Run the target-http service with TLS (foreground)
run-target-http-tls: target-http
	@echo Starting Target HTTP Service with TLS...
	@if not exist certs\target-http mkdir certs\target-http
	@if not exist certs\target-http\tls.crt $(MAKE) generate-certs
	set CHAOS_TARGET_CERT_FILE=certs\target-http\tls.crt && set CHAOS_TARGET_KEY_FILE=certs\target-http\tls.key && .\bin\target-http.exe

# Run the target-http service without TLS (foreground)
run-target-http-notls: target-http
	@echo Starting Target HTTP Service without TLS...
	.\bin\target-http.exe

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
	@echo Starting Target HTTP Service in background...
	@powershell -Command "Start-Process -FilePath '.\bin\target-http.exe' -WindowStyle Normal"
	@timeout /t 2 /nobreak >nul
	@echo Starting ChaosMaster server in background...
	@powershell -Command "Start-Process -FilePath '.\bin\chaos-master.exe' -WindowStyle Normal"
	@timeout /t 3 /nobreak >nul
	@echo Starting Agent 1 in background...
	@powershell -Command "Start-Process -FilePath '.\bin\chaos-agent.exe' -WindowStyle Normal"
	@timeout /t 2 /nobreak >nul
	@echo Starting Agent 2 in background...
	@powershell -Command "Start-Process -FilePath '.\bin\chaos-agent.exe' -WindowStyle Normal"
	@echo.
	@echo === Test Environment Running ===
	@echo Target HTTP: Running on localhost:8080
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
