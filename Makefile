.PHONY: all master agent client clean proto run-server run-client run-both

# Build both master and agent
all: master agent client

# Generate protobuf files
proto:
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./internal/contracts/agent.proto
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
	del /q contract\*.pb.go 2>nul || exit 0
else
	rm -rf bin/
	rm -f internal/contracts/*.pb.go
	rm -f contract/*.pb.go
endif
