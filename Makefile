.PHONY: all master agent clean proto

# Build both master and agent
all: master agent

# Generate protobuf files
proto:
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./internal/contracts/agent.proto

# Build the master binary
master: proto
	go build -o bin/chaos-master.exe ./cmd/master

# Build the agent binary
agent: proto
	go build -o bin/chaos-agent.exe ./cmd/agent

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
