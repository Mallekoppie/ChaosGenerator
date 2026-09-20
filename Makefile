# Chaos Generator
#
# Build:
#   make                        Build every component (master, agent, client, targets)
#   make buildweb               Build the Flutter web app and embed it into the master
#   make docker-build           Build all container images
#
# Run (foreground):
#   make run-master             Master: native gRPC + web control panel
#   make run-agent              A single agent
#   make run-client             Interactive gRPC client (needs a running master)
#   make run-target-http        Plain HTTP target
#   make run-target-http-tls    HTTPS target
#   make run-target-http-http2  HTTPS/HTTP2 target
#   make run-target-grpc        gRPC target
#   make run-target-grpc-tls    gRPC target over TLS
#
# Run (background dev stack):
#   make start                  master + agents + target services
#   make status                 What is running and on which ports
#   make logs                   Follow the background stack logs
#   make stop                   Stop the background stack
#   make run-full-stack         start + Prometheus/Grafana
#
# Kubernetes (local kind cluster):
#   make manifests              Render manifests/kind into manifests/generated
#   make kind-up                Create the cluster, load the images, deploy (host only)
#   make kind-host              Run 'make kind-up' on the host (for distrobox sessions)
#   make kind-status | kind-logs | kind-verify | kind-port-forward
#   make kind-down              Remove the Chaos resources from the cluster
#   make kind-delete            Delete the kind cluster
#   make kind-clean             Delete the kind cluster and the generated manifests
#
# Housekeeping:
#   make clean                  Remove binaries, generated code and web build output
#   make test                   Go test suite + Flutter tests
#   make lint                   go vet + flutter analyze
#   make help                   Show this help

SHELL := /bin/bash

# --- Layout -----------------------------------------------------------------
BIN_DIR       := bin
PROTO_DIR     := internal/contracts
WEB_DIR       := web
DIST_DIR      := cmd/master/dist
RUN_DIR       := .run
CERT_DIR      := certs/target-http

# --- Tooling ----------------------------------------------------------------
GO            ?= go
PROTOC        ?= protoc
PROTOC_DART   ?= protoc-gen-dart
FLUTTER       ?= flutter
DOCKER        ?= docker

PROTOC_FLAGS  := --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

# --- Ports and addresses (override on the command line) ---------------------
MASTER_GRPC_ADDR     ?= 0.0.0.0:9002
MASTER_WEB_ADDR      ?= 0.0.0.0:9001
AGENT_COUNT          ?= 2
# Agent metrics ports are AGENT_METRICS_BASE+1 .. AGENT_METRICS_BASE+AGENT_COUNT
AGENT_METRICS_BASE   ?= 9095
TARGET_HTTP_PORT     ?= 8081
TARGET_HTTP_TLS_PORT ?= 8443
TARGET_HTTP_H2_PORT  ?= 8444
TARGET_GRPC_PORT     ?= 9000

MASTER_GRPC_PORT  := $(lastword $(subst :, ,$(MASTER_GRPC_ADDR)))
MASTER_WEB_PORT   := $(lastword $(subst :, ,$(MASTER_WEB_ADDR)))
AGENT_MASTER_ADDR := 127.0.0.1:$(MASTER_GRPC_PORT)

# --- Web / containers -------------------------------------------------------
# Base href used by `flutter build web`; "/" works for local runs and
# `kubectl port-forward`.
BASE_HREF       ?= /
REGISTRY        ?=
IMAGE_TAG       ?= latest
MONITORING_COMPOSE := docker-compose-monitoring.yml
# Prefer the docker compose plugin but fall back to the standalone binary.
DOCKER_COMPOSE  ?= $(shell $(DOCKER) compose version >/dev/null 2>&1 && echo "$(DOCKER) compose" || echo docker-compose)

COMPONENTS := master agent client target-http target-grpc

.DEFAULT_GOAL := all

# --- Help -------------------------------------------------------------------
.PHONY: help
help:
	@echo "Chaos Generator"
	@echo
	@echo "Build:"
	@echo "  make                     Build every component"
	@echo "  make buildweb            Build the Flutter web app into $(DIST_DIR)"
	@echo "  make proto | proto-dart  Generate Go / Dart stubs"
	@echo "  make tools               Install the protoc plugins"
	@echo "  make docker-build        Build all container images"
	@echo
	@echo "Run (foreground):"
	@echo "  make run-master          gRPC $(MASTER_GRPC_ADDR) + web http://localhost:$(MASTER_WEB_PORT)"
	@echo "  make run-master-dev      gRPC-Web only, for 'flutter run'"
	@echo "  make run-agent | run-client"
	@echo "  make run-target-http | run-target-http-tls | run-target-http-http2"
	@echo "  make run-target-grpc | run-target-grpc-tls"
	@echo
	@echo "Run (background):"
	@echo "  make start               master + $(AGENT_COUNT) agent(s) + targets"
	@echo "  make status | logs | stop | restart"
	@echo "  make run-full-stack      start + Prometheus/Grafana"
	@echo
	@echo "Monitoring:"
	@echo "  make monitoring-up       Docker if available, otherwise native binaries"
	@echo "  make monitoring-down | monitoring-logs | monitoring-restart"
	@echo "  make monitoring-install  Download Prometheus/Grafana into .monitoring"
	@echo
	@echo "Kubernetes (local kind):"
	@echo "  make manifests           Render $(MANIFEST_DIR) into $(GENERATED_DIR)"
	@echo "  make kind-up             kind cluster + images + deploy (run on the host)"
	@echo "  make kind-host           Run 'make kind-up' on the host via $(DISTROBOX_EXEC)"
	@echo "  make kind-status | kind-logs | kind-verify | kind-port-forward"
	@echo "  make kind-down | kind-delete | kind-clean"
	@echo
	@echo "Housekeeping:"
	@echo "  make clean | test | lint"
	@echo
	@echo "Override ports, for example:"
	@echo "  make start MASTER_WEB_ADDR=0.0.0.0:9100 TARGET_HTTP_PORT=8085"

# --- Build ------------------------------------------------------------------
define go-build
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(1) ./cmd/$(2)
endef

.PHONY: all $(COMPONENTS)
all: $(COMPONENTS)

master: proto
	$(call go-build,chaos-master,master)

agent: proto
	$(call go-build,chaos-agent,agent)

client: proto
	$(call go-build,chaos-client,client)

target-http: proto
	$(call go-build,target-http,target-http)

target-grpc: proto
	$(call go-build,target-grpc,target-grpc)

# --- Protobuf ---------------------------------------------------------------
.PHONY: proto
proto: $(PROTO_DIR)/master.pb.go $(PROTO_DIR)/target.pb.go $(PROTO_DIR)/auth.pb.go

# Each protoc invocation produces both the message and the gRPC stubs.
$(PROTO_DIR)/master.pb.go $(PROTO_DIR)/master_grpc.pb.go &: $(PROTO_DIR)/master.proto
	$(PROTOC) $(PROTOC_FLAGS) $<

$(PROTO_DIR)/target.pb.go $(PROTO_DIR)/target_grpc.pb.go &: $(PROTO_DIR)/target.proto
	$(PROTOC) $(PROTOC_FLAGS) $<

$(PROTO_DIR)/auth.pb.go $(PROTO_DIR)/auth_grpc.pb.go &: $(PROTO_DIR)/auth.proto
	$(PROTOC) $(PROTOC_FLAGS) $<

.PHONY: tools
tools:
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	dart pub global activate protoc_plugin 25.1.0
	@echo "Installed protoc-gen-go, protoc-gen-go-grpc and protoc-gen-dart."
	@echo "Add $$($(GO) env GOPATH)/bin and the pub cache bin directory to your PATH."

# --- Flutter web app --------------------------------------------------------
.PHONY: proto-dart buildweb web
proto-dart:
	@command -v $(PROTOC_DART) >/dev/null 2>&1 || { \
		echo "$(PROTOC_DART) not found. Run 'make tools' first."; exit 1; }
	@mkdir -p $(WEB_DIR)/lib/src/generated
	$(PROTOC) -I $(PROTO_DIR) --dart_out=grpc:$(WEB_DIR)/lib/src/generated $(PROTO_DIR)/master.proto $(PROTO_DIR)/auth.proto
	@echo "Generated Dart stubs in $(WEB_DIR)/lib/src/generated"

buildweb: proto-dart
	@rm -rf $(DIST_DIR)
	@mkdir -p $(DIST_DIR)
	cd $(WEB_DIR) && $(FLUTTER) pub get && $(FLUTTER) build web --release --base-href=$(BASE_HREF)
	cp -r $(WEB_DIR)/build/web/. $(DIST_DIR)/
	@touch $(DIST_DIR)/.gitkeep
	@echo "Embedded web app written to $(DIST_DIR)"

web: buildweb

# --- Clean ------------------------------------------------------------------
.PHONY: clean clean-web clean-proto clean-certs
clean: clean-web clean-proto clean-certs
	rm -rf $(BIN_DIR) $(RUN_DIR)
	@echo "Removed binaries and generated output."

clean-web:
	rm -rf $(WEB_DIR)/build $(WEB_DIR)/.dart_tool
	rm -f $(WEB_DIR)/lib/src/generated/*.pb.dart \
	      $(WEB_DIR)/lib/src/generated/*.pbgrpc.dart \
	      $(WEB_DIR)/lib/src/generated/*.pbenum.dart \
	      $(WEB_DIR)/lib/src/generated/*.pbjson.dart
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
	touch $(DIST_DIR)/.gitkeep

clean-proto:
	rm -f $(PROTO_DIR)/*.pb.go

clean-certs:
	rm -rf certs

# --- TLS certificates -------------------------------------------------------
.PHONY: generate-certs
generate-certs:
	@mkdir -p $(CERT_DIR)
	@if [ ! -f $(CERT_DIR)/tls.crt ]; then \
		echo "Generating a self-signed certificate for localhost"; \
		openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
			-keyout $(CERT_DIR)/tls.key -out $(CERT_DIR)/tls.crt \
			-subj "/CN=localhost" \
			-addext "subjectAltName=DNS:localhost,DNS:target-http,IP:127.0.0.1"; \
	fi

# --- Run (foreground) -------------------------------------------------------
.PHONY: run-master run-master-dev run-agent run-client run-target-http \
        run-target-http-tls run-target-http-http2 run-target-grpc run-target-grpc-tls

run-master: master
	CHAOS_MASTER_GRPC_ADDRESS=$(MASTER_GRPC_ADDR) \
	CHAOS_MASTER_WEB_ADDRESS=$(MASTER_WEB_ADDR) \
	$(BIN_DIR)/chaos-master

# Master with gRPC-Web only (no embedded UI) so `flutter run` can serve the app.
run-master-dev: master
	DEBUG=true \
	CHAOS_MASTER_GRPC_ADDRESS=$(MASTER_GRPC_ADDR) \
	CHAOS_MASTER_WEB_ADDRESS=$(MASTER_WEB_ADDR) \
	$(BIN_DIR)/chaos-master

run-agent: agent
	$(BIN_DIR)/chaos-agent

run-client: client
	$(BIN_DIR)/chaos-client

run-target-http: target-http
	CHAOS_TARGET_PORT=$(TARGET_HTTP_PORT) $(BIN_DIR)/target-http

run-target-http-tls: target-http generate-certs
	CHAOS_TARGET_PORT=$(TARGET_HTTP_TLS_PORT) \
	CHAOS_TARGET_CERT_FILE=$(CERT_DIR)/tls.crt \
	CHAOS_TARGET_KEY_FILE=$(CERT_DIR)/tls.key \
	$(BIN_DIR)/target-http

run-target-http-http2: target-http generate-certs
	CHAOS_TARGET_PORT=$(TARGET_HTTP_H2_PORT) \
	CHAOS_TARGET_PROTOCOL=http2 \
	CHAOS_TARGET_CERT_FILE=$(CERT_DIR)/tls.crt \
	CHAOS_TARGET_KEY_FILE=$(CERT_DIR)/tls.key \
	$(BIN_DIR)/target-http

run-target-grpc: target-grpc
	CHAOS_TARGET_PORT=$(TARGET_GRPC_PORT) $(BIN_DIR)/target-grpc

run-target-grpc-tls: target-grpc generate-certs
	CHAOS_TARGET_PORT=$(TARGET_GRPC_PORT) \
	CHAOS_TARGET_CERT_FILE=$(CERT_DIR)/tls.crt \
	CHAOS_TARGET_KEY_FILE=$(CERT_DIR)/tls.key \
	$(BIN_DIR)/target-grpc

# --- Background dev stack ---------------------------------------------------
.PHONY: start stop restart status logs run-full-stack

start: all generate-certs
	@if ls $(RUN_DIR)/target-*.pid >/dev/null 2>&1; then \
		echo "A stack is already running - stopping it first..."; \
		$(MAKE) --no-print-directory stop; \
	fi
	@mkdir -p $(RUN_DIR)
	@for f in $(RUN_DIR)/*.pid; do \
		[ -e "$$f" ] || continue; \
		case "$$(basename "$$f")" in prometheus.pid|grafana.pid) continue;; esac; \
		rm -f "$$f"; \
	done
	@echo "Starting target services..."
	@nohup env CHAOS_TARGET_PORT=$(TARGET_HTTP_PORT) $(BIN_DIR)/target-http > $(RUN_DIR)/target-http.log 2>&1 & echo $$! > $(RUN_DIR)/target-http.pid
	@nohup env CHAOS_TARGET_PORT=$(TARGET_HTTP_TLS_PORT) CHAOS_TARGET_METRICS_PORT=9093 \
		CHAOS_TARGET_CERT_FILE=$(CERT_DIR)/tls.crt CHAOS_TARGET_KEY_FILE=$(CERT_DIR)/tls.key \
		$(BIN_DIR)/target-http > $(RUN_DIR)/target-http-tls.log 2>&1 & echo $$! > $(RUN_DIR)/target-http-tls.pid
	@nohup env CHAOS_TARGET_PORT=$(TARGET_HTTP_H2_PORT) CHAOS_TARGET_METRICS_PORT=9094 CHAOS_TARGET_PROTOCOL=http2 \
		CHAOS_TARGET_CERT_FILE=$(CERT_DIR)/tls.crt CHAOS_TARGET_KEY_FILE=$(CERT_DIR)/tls.key \
		$(BIN_DIR)/target-http > $(RUN_DIR)/target-http-http2.log 2>&1 & echo $$! > $(RUN_DIR)/target-http-http2.pid
	@nohup env CHAOS_TARGET_PORT=$(TARGET_GRPC_PORT) CHAOS_TARGET_METRICS_PORT=9095 \
		CHAOS_TARGET_CERT_FILE=$(CERT_DIR)/tls.crt CHAOS_TARGET_KEY_FILE=$(CERT_DIR)/tls.key \
		$(BIN_DIR)/target-grpc > $(RUN_DIR)/target-grpc.log 2>&1 & echo $$! > $(RUN_DIR)/target-grpc.pid
	@echo "Starting master (gRPC $(MASTER_GRPC_ADDR), web $(MASTER_WEB_ADDR))..."
	@nohup env CHAOS_MASTER_GRPC_ADDRESS=$(MASTER_GRPC_ADDR) CHAOS_MASTER_WEB_ADDRESS=$(MASTER_WEB_ADDR) \
		$(BIN_DIR)/chaos-master > $(RUN_DIR)/master.log 2>&1 & echo $$! > $(RUN_DIR)/master.pid
	@sleep 2
	@echo "Starting $(AGENT_COUNT) agent(s)..."
	@i=1; while [ $$i -le $(AGENT_COUNT) ]; do \
		metrics=$$(( $(AGENT_METRICS_BASE) + $$i )); \
		nohup env CHAOS_AGENT_METRICS_PORT=$$metrics CHAOS_MASTER_ADDRESS=$(AGENT_MASTER_ADDR) \
			$(BIN_DIR)/chaos-agent > $(RUN_DIR)/agent-$$i.log 2>&1 & \
		echo $$! > $(RUN_DIR)/agent-$$i.pid; \
		echo "  agent-$$i (metrics port $$metrics)"; \
		i=$$((i + 1)); \
	done
	@sleep 1
	@echo
	@echo "Service status:"
	@for f in $(RUN_DIR)/*.pid; do \
		name=$$(basename "$$f" .pid); \
		pid=$$(cat "$$f"); \
		if kill -0 "$$pid" 2>/dev/null; then \
			echo "  $$name: running (pid $$pid)"; \
		else \
			echo "  $$name: FAILED - see $(RUN_DIR)/$$name.log"; \
		fi; \
	done
	@echo
	@echo "Web UI       http://localhost:$(MASTER_WEB_PORT)"
	@echo "Master gRPC  $(MASTER_GRPC_ADDR)"
	@echo "Targets      http://localhost:$(TARGET_HTTP_PORT)  https://localhost:$(TARGET_HTTP_TLS_PORT)  https://localhost:$(TARGET_HTTP_H2_PORT)  grpc://localhost:$(TARGET_GRPC_PORT)"
	@echo "Monitoring   make monitoring-up   (Prometheus :9091, Grafana :3000)"
	@echo
	@echo "Logs: make logs    Status: make status    Stop: make stop"

stop:
	@if [ -d $(RUN_DIR) ]; then \
		for f in $(RUN_DIR)/*.pid; do \
			[ -e "$$f" ] || continue; \
			name=$$(basename "$$f" .pid); \
			case "$$name" in prometheus|grafana) continue;; esac; \
			pid=$$(cat "$$f"); \
			if kill -0 "$$pid" 2>/dev/null; then \
				kill "$$pid" 2>/dev/null && echo "stopped $$name (pid $$pid)"; \
			fi; \
			rm -f "$$f"; \
		done; \
	fi
	@pkill -f "[b]in/chaos-" 2>/dev/null || true
	@pkill -f "[b]in/target-" 2>/dev/null || true
	@echo "All Chaos processes stopped."

restart: stop
	@$(MAKE) --no-print-directory start

status:
	@echo "Master   gRPC $(MASTER_GRPC_ADDR)   web http://localhost:$(MASTER_WEB_PORT)"
	@echo
	@if pgrep -af "[b]in/(chaos|target)-" >/dev/null 2>&1; then \
		echo "Processes:"; pgrep -af "[b]in/(chaos|target)-"; \
	else \
		echo "No Chaos processes are running (make start)"; \
	fi
	@echo
	@echo "Listening ports:"
	@ss -ltn 2>/dev/null | grep -E ':(9000|9001|9002|8080|8443|8444|9090|9091|9093|9094|9095|9096|9097)\b' || echo "  (none of the Chaos ports are listening)"

logs:
	@if [ -d $(RUN_DIR) ]; then \
		tail -n 20 -f $(RUN_DIR)/*.log; \
	else \
		echo "No background stack is running (make start)"; \
	fi

run-full-stack: start
	@$(MAKE) --no-print-directory monitoring-up
	@echo
	@echo "Full stack running:"
	@echo "  Web UI:     http://localhost:$(MASTER_WEB_PORT)"
	@echo "  Prometheus: http://localhost:$(PROMETHEUS_PORT)"
	@echo "  Grafana:    http://localhost:$(GRAFANA_PORT) (admin/admin)"

# --- Monitoring -------------------------------------------------------------
# `monitoring-up` uses Docker Compose when a Docker daemon is reachable, and
# otherwise downloads and runs Prometheus and Grafana natively. Set
# MONITORING_MODE=docker or MONITORING_MODE=local to force one of them.
MONITORING_MODE     ?= auto
MONITORING_DIR      ?= $(CURDIR)/.monitoring
PROMETHEUS_VERSION  ?= 3.14.0
GRAFANA_VERSION     ?= 13.2.1
MONITORING_ARCH     ?= linux-amd64

# Remember the ports used by the previous run so `monitoring-restart` keeps
# them. Values passed on the command line still take precedence.
-include $(MONITORING_DIR)/monitoring.env

PROMETHEUS_PORT     ?= 9091
GRAFANA_PORT        ?= 3000
PROMETHEUS_LOCAL_CONFIG := monitoring/prometheus-local.yml
PROMETHEUS_DIR      := $(MONITORING_DIR)/prometheus-$(PROMETHEUS_VERSION).$(MONITORING_ARCH)
GRAFANA_DIR         := $(MONITORING_DIR)/grafana

.PHONY: monitoring-up monitoring-down monitoring-restart monitoring-logs monitoring-clean \
        monitoring-docker-up monitoring-docker-down monitoring-docker-logs \
        monitoring-local-up monitoring-local-down monitoring-local-logs monitoring-install

monitoring-up:
	@if [ "$(MONITORING_MODE)" = "local" ] || { [ "$(MONITORING_MODE)" = "auto" ] && ! $(DOCKER) info >/dev/null 2>&1; }; then \
		$(MAKE) --no-print-directory monitoring-local-up; \
	else \
		$(MAKE) --no-print-directory monitoring-docker-up; \
	fi

monitoring-down:
	@if pgrep -f "[.]monitoring/(prometheus|grafana)/" >/dev/null 2>&1 || \
	   [ -f $(RUN_DIR)/prometheus.pid ] || [ -f $(RUN_DIR)/grafana.pid ]; then \
		$(MAKE) --no-print-directory monitoring-local-down; \
	else \
		$(MAKE) --no-print-directory monitoring-docker-down; \
	fi

monitoring-restart: monitoring-down monitoring-up

monitoring-logs:
	@if pgrep -f "[.]monitoring/(prometheus|grafana)/" >/dev/null 2>&1 || \
	   [ -f $(RUN_DIR)/prometheus.pid ] || [ -f $(RUN_DIR)/grafana.pid ]; then \
		tail -n 20 -f $(RUN_DIR)/prometheus.log $(RUN_DIR)/grafana.log; \
	else \
		$(MAKE) --no-print-directory monitoring-docker-logs; \
	fi

monitoring-clean:
	rm -rf $(MONITORING_DIR)
	@echo "Removed $(MONITORING_DIR)"

# --- Monitoring via Docker Compose ---
monitoring-docker-up:
	@command -v $(DOCKER) >/dev/null 2>&1 || { echo "$(DOCKER) not found"; exit 1; }
	@$(DOCKER) info >/dev/null 2>&1 || { \
		echo "No Docker daemon is reachable; run 'make monitoring-up' to use the native binaries."; exit 1; }
	$(DOCKER_COMPOSE) -f $(MONITORING_COMPOSE) up -d
	@echo "Prometheus: http://localhost:$(PROMETHEUS_PORT)"
	@echo "Grafana:    http://localhost:$(GRAFANA_PORT) (admin/admin)"

monitoring-docker-down:
	$(DOCKER_COMPOSE) -f $(MONITORING_COMPOSE) down

monitoring-docker-logs:
	$(DOCKER_COMPOSE) -f $(MONITORING_COMPOSE) logs -f

# --- Monitoring natively (no Docker required) ---
monitoring-install:
	@mkdir -p $(MONITORING_DIR)/data/prometheus $(MONITORING_DIR)/data/grafana
	@if [ ! -x $(PROMETHEUS_DIR)/prometheus ]; then \
		echo "Downloading Prometheus $(PROMETHEUS_VERSION)..."; \
		curl -fsSL "https://github.com/prometheus/prometheus/releases/download/v$(PROMETHEUS_VERSION)/prometheus-$(PROMETHEUS_VERSION).$(MONITORING_ARCH).tar.gz" \
			| tar -xz -C $(MONITORING_DIR); \
	fi
	@if [ ! -x $(GRAFANA_DIR)/bin/grafana ]; then \
		echo "Downloading Grafana $(GRAFANA_VERSION) (large download, please wait)..."; \
		mkdir -p $(MONITORING_DIR)/tmp; \
		curl -fsSL "https://dl.grafana.com/oss/release/grafana-$(GRAFANA_VERSION).$(MONITORING_ARCH).tar.gz" \
			| tar -xz -C $(MONITORING_DIR)/tmp; \
		rm -rf $(GRAFANA_DIR); \
		mv $(MONITORING_DIR)/tmp/grafana* $(GRAFANA_DIR); \
		rm -rf $(MONITORING_DIR)/tmp; \
	fi

monitoring-local-up: monitoring-install
	@mkdir -p $(RUN_DIR)
	@printf 'PROMETHEUS_PORT := %s\nGRAFANA_PORT := %s\n' '$(PROMETHEUS_PORT)' '$(GRAFANA_PORT)' > $(MONITORING_DIR)/monitoring.env
	@$(MAKE) --no-print-directory monitoring-local-provisioning
	@echo "Starting Prometheus on :$(PROMETHEUS_PORT)..."
	@nohup $(PROMETHEUS_DIR)/prometheus \
		--config.file=$(PROMETHEUS_LOCAL_CONFIG) \
		--storage.tsdb.path=$(MONITORING_DIR)/data/prometheus \
		--web.listen-address=0.0.0.0:$(PROMETHEUS_PORT) \
		> $(RUN_DIR)/prometheus.log 2>&1 & echo $$! > $(RUN_DIR)/prometheus.pid
	@echo "Starting Grafana on :$(GRAFANA_PORT)..."
	@nohup env GF_SECURITY_ADMIN_USER=admin GF_SECURITY_ADMIN_PASSWORD=admin GF_USERS_ALLOW_SIGN_UP=false \
		$(GRAFANA_DIR)/bin/grafana server \
		--homepath $(GRAFANA_DIR) \
		cfg:default.paths.data=$(MONITORING_DIR)/data/grafana \
		cfg:default.paths.logs=$(MONITORING_DIR)/data/grafana/logs \
		cfg:default.paths.provisioning=$(MONITORING_DIR)/provisioning \
		cfg:default.server.http_port=$(GRAFANA_PORT) \
		> $(RUN_DIR)/grafana.log 2>&1 & echo $$! > $(RUN_DIR)/grafana.pid
	@sleep 4
	@failed=0; \
	for name in prometheus grafana; do \
		pid=$$(cat $(RUN_DIR)/$$name.pid 2>/dev/null); \
		if kill -0 "$$pid" 2>/dev/null; then \
			echo "  $$name: running (pid $$pid)"; \
		else \
			echo "  $$name: FAILED - see $(RUN_DIR)/$$name.log"; failed=1; \
		fi; \
	done; \
	if [ "$$failed" = "1" ]; then \
		echo; \
		echo "A port may already be in use. Try different ports, for example:"; \
		echo "  make monitoring-up PROMETHEUS_PORT=9191 GRAFANA_PORT=3001"; \
		exit 1; \
	fi
	@echo
	@echo "Prometheus: http://localhost:$(PROMETHEUS_PORT)"
	@echo "Grafana:    http://localhost:$(GRAFANA_PORT) (admin/admin)"
	@echo "Logs: make monitoring-logs"

monitoring-local-down:
	@for name in grafana prometheus; do \
		if [ -f $(RUN_DIR)/$$name.pid ]; then \
			pid=$$(cat $(RUN_DIR)/$$name.pid); \
			if kill -0 "$$pid" 2>/dev/null; then \
				kill "$$pid" 2>/dev/null && echo "stopped $$name (pid $$pid)"; \
			fi; \
			rm -f $(RUN_DIR)/$$name.pid; \
		fi; \
	done
	@pkill -f "[.]monitoring/prometheus-" 2>/dev/null || true
	@pkill -f "[.]monitoring/grafana/" 2>/dev/null || true
	@echo "Monitoring stopped."

monitoring-local-logs:
	@tail -n 20 -f $(RUN_DIR)/prometheus.log $(RUN_DIR)/grafana.log

# Grafana provisioning for native runs: reuse the committed dashboards but point
# the datasource at the host Prometheus port.
monitoring-local-provisioning:
	@mkdir -p $(MONITORING_DIR)/provisioning/datasources $(MONITORING_DIR)/provisioning/dashboards
	@printf '%s\n' \
		'apiVersion: 1' \
		'' \
		'datasources:' \
		'  - name: Prometheus' \
		'    type: prometheus' \
		'    access: proxy' \
		'    url: http://localhost:$(PROMETHEUS_PORT)' \
		'    isDefault: true' \
		'    editable: true' \
		> $(MONITORING_DIR)/provisioning/datasources/prometheus.yml
	@printf '%s\n' \
		'apiVersion: 1' \
		'' \
		'providers:' \
		'  - name: Chaos Generator Dashboards' \
		'    orgId: 1' \
		"    folder: ''" \
		'    type: file' \
		'    disableDeletion: false' \
		'    updateIntervalSeconds: 10' \
		'    allowUiUpdates: true' \
		'    options:' \
		"      path: '$(CURDIR)/monitoring/grafana/provisioning/dashboards'" \
		> $(MONITORING_DIR)/provisioning/dashboards/dashboard.yml

# --- Container images -------------------------------------------------------
.PHONY: docker-build docker-build-master docker-build-agent \
        docker-build-target-http docker-build-target-grpc
docker-build: docker-build-master docker-build-agent docker-build-target-http docker-build-target-grpc

docker-build-master:
	$(DOCKER) build -f build/dockerfile-chaos-master --build-arg FLUTTER_BASE_HREF=$(BASE_HREF) \
		-t $(REGISTRY)chaos-master:$(IMAGE_TAG) .

docker-build-agent:
	$(DOCKER) build -f build/dockerfile-chaos-agent -t $(REGISTRY)chaos-agent:$(IMAGE_TAG) .

docker-build-target-http:
	$(DOCKER) build -f build/dockerfile-target-http -t $(REGISTRY)target-http:$(IMAGE_TAG) .

docker-build-target-grpc:
	$(DOCKER) build -f build/dockerfile-target-grpc -t $(REGISTRY)target-grpc:$(IMAGE_TAG) .

# --- Local kind cluster -----------------------------------------------------
# `make manifests` renders manifests/kind into manifests/generated using the
# same image names and tags that `make docker-build` produces. The `kind-*`
# targets then create a throwaway cluster, load those images into it and deploy
# the master, the agents and the target service.
#
# kind drives podman (or docker), which cannot run inside the distrobox
# container this repo is usually edited in. Run the kind targets on the host:
#   make kind-host
# or directly:
#   distrobox-host-exec bash -lc 'cd <repo> && make kind-up'
KIND             ?= kind
KUBECTL          ?= kubectl
KIND_CLUSTER     ?= chaos
KIND_NAMESPACE   := chaos-testing
MANIFEST_DIR     := manifests/kind
GENERATED_DIR    := manifests/generated
CONTAINER_ENGINE ?= $(shell command -v docker >/dev/null 2>&1 && echo docker || echo podman)
DISTROBOX_EXEC   ?= distrobox-host-exec
KIND_IMAGES      := $(REGISTRY)chaos-master:$(IMAGE_TAG) $(REGISTRY)chaos-agent:$(IMAGE_TAG) $(REGISTRY)target-http:$(IMAGE_TAG)
# kind needs the experimental flag to drive podman instead of docker.
ifeq ($(CONTAINER_ENGINE),podman)
KIND_ENV         := KIND_EXPERIMENTAL_PROVIDER=podman
else
KIND_ENV         :=
endif

.PHONY: manifests kind-preflight kind-create kind-images kind-deploy kind-verify \
        kind-up kind-host kind-status kind-logs kind-port-forward kind-down \
        kind-delete kind-clean

# Render the kind manifests with the local image names/tags. No cluster and no
# container engine required, so this also works inside the dev container.
manifests:
	@rm -rf $(GENERATED_DIR)
	@mkdir -p $(GENERATED_DIR)/server $(GENERATED_DIR)/agent $(GENERATED_DIR)/target
	@cp $(MANIFEST_DIR)/namespace.yaml $(GENERATED_DIR)/namespace.yaml
	@cp $(MANIFEST_DIR)/server/*.yaml $(GENERATED_DIR)/server/
	@cp $(MANIFEST_DIR)/agent/*.yaml $(GENERATED_DIR)/agent/
	@cp $(MANIFEST_DIR)/target/*.yaml $(GENERATED_DIR)/target/
	@printf '%s\n' \
		'apiVersion: kustomize.config.k8s.io/v1beta1' \
		'kind: Kustomization' \
		'' \
		'namespace: $(KIND_NAMESPACE)' \
		'' \
		'resources:' \
		'  - namespace.yaml' \
		'  - server' \
		'  - agent' \
		'  - target' \
		'' \
		'images:' \
		'  - name: chaos-master' \
		'    newName: $(REGISTRY)chaos-master' \
		'    newTag: $(IMAGE_TAG)' \
		'  - name: chaos-agent' \
		'    newName: $(REGISTRY)chaos-agent' \
		'    newTag: $(IMAGE_TAG)' \
		'  - name: target-http' \
		'    newName: $(REGISTRY)target-http' \
		'    newTag: $(IMAGE_TAG)' \
		> $(GENERATED_DIR)/kustomization.yaml
	@$(KUBECTL) kustomize $(GENERATED_DIR) > $(GENERATED_DIR)/kind-all.yaml
	@$(KUBECTL) kustomize $(GENERATED_DIR)/server > $(GENERATED_DIR)/kind-server.yaml
	@$(KUBECTL) kustomize $(GENERATED_DIR)/agent > $(GENERATED_DIR)/kind-agent.yaml
	@$(KUBECTL) kustomize $(GENERATED_DIR)/target > $(GENERATED_DIR)/kind-target.yaml
	@echo "Generated kind manifests in $(GENERATED_DIR):"
	@ls -1 $(GENERATED_DIR)/kind-*.yaml | sed 's|^|  |'

# Refuse to run from inside distrobox: kind needs the host's container engine.
kind-preflight:
	@command -v $(KIND) >/dev/null 2>&1 || { \
		echo "$(KIND) not found - install it on the host (https://kind.sigs.k8s.io/)."; exit 1; }
	@command -v $(KUBECTL) >/dev/null 2>&1 || { \
		echo "$(KUBECTL) not found - install it on the host."; exit 1; }
	@command -v $(CONTAINER_ENGINE) >/dev/null 2>&1 || { \
		echo "$(CONTAINER_ENGINE) not found - install it on the host, or pass CONTAINER_ENGINE=docker."; exit 1; }
	@if [ -e /run/.containerenv ] || [ -n "$$CONTAINER_ID" ]; then \
		echo "$(KIND) cannot run inside distrobox ($${CONTAINER_ID:-container})."; \
		echo; \
		echo "Run the kind targets on the host instead:"; \
		echo "  $(DISTROBOX_EXEC) bash -lc 'cd $(CURDIR) && make kind-up'"; \
		echo; \
		echo "or from here: make kind-host"; \
		exit 1; \
	fi
	@$(CONTAINER_ENGINE) info >/dev/null 2>&1 || { \
		echo "$(CONTAINER_ENGINE) is not reachable - start it on the host (podman: systemctl --user start podman.socket)."; exit 1; }

kind-create: kind-preflight
	@if $(KIND) get clusters 2>/dev/null | grep -qx '$(KIND_CLUSTER)'; then \
		echo "kind cluster '$(KIND_CLUSTER)' already exists."; \
	else \
		echo "Creating kind cluster '$(KIND_CLUSTER)' with $(CONTAINER_ENGINE)..."; \
		$(KIND_ENV) $(KIND) create cluster --name $(KIND_CLUSTER); \
	fi

# Build the images kind deploys and load them into the node, so the pods never
# need a registry.
kind-images: kind-preflight
	@echo "Building images with $(CONTAINER_ENGINE)..."
	@$(MAKE) --no-print-directory docker-build-master docker-build-agent \
		docker-build-target-http DOCKER=$(CONTAINER_ENGINE)
	@for image in $(KIND_IMAGES); do \
		echo "Loading $$image into kind cluster '$(KIND_CLUSTER)'..."; \
		if [ "$(CONTAINER_ENGINE)" = "docker" ]; then \
			$(KIND_ENV) $(KIND) load docker-image --name $(KIND_CLUSTER) $$image; \
		else \
			tmp=$$(mktemp -d); \
			$(CONTAINER_ENGINE) save --format=docker-archive -o $$tmp/image.tar $$image; \
			$(KIND_ENV) $(KIND) load image-archive --name $(KIND_CLUSTER) $$tmp/image.tar; \
			rm -rf $$tmp; \
		fi; \
	done
	@echo "Images loaded into kind cluster '$(KIND_CLUSTER)'."

kind-deploy: manifests kind-create
	@$(KUBECTL) apply -f $(GENERATED_DIR)/kind-all.yaml
	@for name in chaos-master chaos-agent target-http; do \
		echo "Waiting for deployment/$$name..."; \
		$(KUBECTL) -n $(KIND_NAMESPACE) rollout status deployment/$$name --timeout=180s; \
	done
	@$(MAKE) --no-print-directory kind-status

kind-verify:
	@echo "Agents connected to the master:"
	@$(KUBECTL) -n $(KIND_NAMESPACE) logs -l app=chaos-master --tail=200 2>/dev/null \
		| grep "Agent connected" || echo "  (none yet)"

kind-up: kind-create kind-images kind-deploy
	@echo
	@echo "kind cluster '$(KIND_CLUSTER)' is running (namespace $(KIND_NAMESPACE))."
	@echo "  Web UI      make kind-port-forward   then open http://localhost:8080/"
	@echo "  Target URL  http://target-http.$(KIND_NAMESPACE).svc.cluster.local:8080"
	@echo "  Register    $(KUBECTL) -n $(KIND_NAMESPACE) get secret chaos-master-secrets -o jsonpath='{.data.registerSecret}' | base64 -d"
	@echo "  Status      make kind-status    Logs: make kind-logs"
	@echo "  Teardown    make kind-down | kind-delete | kind-clean"

# Convenience for a distrobox session: run the host's make with the host's tools.
kind-host:
	@command -v $(DISTROBOX_EXEC) >/dev/null 2>&1 || { \
		echo "$(DISTROBOX_EXEC) not found - run 'make kind-up' in a host terminal instead."; exit 1; }
	@echo "Running 'make kind-up' on the host via $(DISTROBOX_EXEC)..."
	@$(DISTROBOX_EXEC) bash -lc 'cd "$(CURDIR)" && make kind-up KIND_CLUSTER=$(KIND_CLUSTER) REGISTRY=$(REGISTRY) IMAGE_TAG=$(IMAGE_TAG)'

kind-status:
	@$(KUBECTL) -n $(KIND_NAMESPACE) get deployments,pods,svc,pvc 2>/dev/null \
		|| echo "Cannot reach kind cluster '$(KIND_CLUSTER)' - run 'make kind-up'."

kind-logs:
	@$(KUBECTL) -n $(KIND_NAMESPACE) logs -l app=chaos-master --tail=50 --prefix
	@$(KUBECTL) -n $(KIND_NAMESPACE) logs -l app=chaos-agent --tail=50 --prefix
	@$(KUBECTL) -n $(KIND_NAMESPACE) logs -l app=target-http --tail=50 --prefix

kind-port-forward:
	@echo "Master web control panel: http://localhost:8080/ (Ctrl-C to stop)"
	@$(KUBECTL) -n $(KIND_NAMESPACE) port-forward svc/chaos-master-service 8080:8080

kind-down:
	@if [ -f $(GENERATED_DIR)/kind-all.yaml ]; then \
		$(KUBECTL) delete -f $(GENERATED_DIR)/kind-all.yaml --ignore-not-found; \
	else \
		$(KUBECTL) delete namespace $(KIND_NAMESPACE) --ignore-not-found; \
	fi
	@echo "Removed the Chaos resources (kind cluster '$(KIND_CLUSTER)' is still running)."

kind-delete:
	@$(KIND) delete cluster --name $(KIND_CLUSTER) || true

kind-clean: kind-delete
	@rm -rf $(GENERATED_DIR)
	@echo "Removed kind cluster '$(KIND_CLUSTER)' and $(GENERATED_DIR)."

# --- Tests and linting ------------------------------------------------------
.PHONY: test test-go test-web lint lint-go lint-web
test: test-go test-web

test-go:
	$(GO) test ./...

test-web:
	@cd $(WEB_DIR) && $(FLUTTER) test

lint: lint-go lint-web

lint-go:
	$(GO) vet ./...

lint-web:
	@cd $(WEB_DIR) && $(FLUTTER) analyze
