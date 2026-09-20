# Chaos Generator
#
# Everything runs in a local kind cluster: this Makefile builds the binaries and
# the container images, renders the manifests, drives the cluster and proxies the
# web UI and the monitoring stack back to your machine.
#
# The kind targets and the proxy need kubectl, kind and a container engine on the
# host. They refuse to run inside the dev container, which drives the host engine
# over a mounted socket but cannot run nested containers.
#
# Build:
#   make                        Build every component
#   make buildweb               Build the Flutter web app (also done by the image)
#   make proto                  Generate the Go stubs (proto-dart for Dart)
#   make tools                  Install the protoc plugins (dev container has them)
#   make docker-build           Build all container images
#
# Cluster:
#   make kind-up                Create the cluster, load images, deploy everything
#   make kind-deploy            Re-render and re-apply the manifests
#   make kind-status            Deployments, pods, services and PVCs
#   make kind-verify            Agents that registered with the master
#   make kind-logs              Master, agent and target logs
#   make kind-restart           Roll the deployments (pick up freshly loaded images)
#   make kind-down              Remove the Chaos resources, keep the cluster
#   make kind-delete            Delete the kind cluster
#   make kind-clean             Delete the cluster and the generated manifests
#
# Access:
#   make proxy                  Port-forward the web UI, Prometheus and Grafana
#   make proxy-web              Port-forward only the web UI
#   make proxy-monitoring       Port-forward only Prometheus and Grafana
#   make proxy-grpc             Port-forward the master gRPC port (for the client)
#   make host TARGET=proxy      Run an above target on the host (see HOST_RUNNER)
#
# Housekeeping:
#   make test                   Go test suite + Flutter tests
#   make lint                   go vet + flutter analyze
#   make fmt                    Format the Go and Dart sources
#   make tidy                   go mod tidy
#   make clean                  Remove binaries, generated code and web output
#   make env                    Show the resolved cluster and port configuration
#   make help                   Show this help
#
# Local ports (override on the command line, for example:
#   make proxy WEB_UI_PORT=9100 PROMETHEUS_PORT=9191):
#   web UI     9001
#   Prometheus 9091
#   Grafana    3000

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

.DEFAULT_GOAL := all

# --- Layout -----------------------------------------------------------------
BIN_DIR   := bin
PROTO_DIR := internal/contracts
WEB_DIR   := web
DIST_DIR  := cmd/master/dist
CERT_DIR  := certs/target-http

# --- Tooling ----------------------------------------------------------------
GO          ?= go
PROTOC      ?= protoc
PROTOC_DART ?= protoc-gen-dart
FLUTTER     ?= flutter
DART        ?= dart
DOCKER      ?= docker
KIND        ?= kind
KUBECTL     ?= kubectl

PROTOC_FLAGS := --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

# --- Components -------------------------------------------------------------
COMPONENTS := master agent client target-http target-grpc

COMPONENT_DIR_master      := master
COMPONENT_DIR_agent       := agent
COMPONENT_DIR_client      := client
COMPONENT_DIR_target-http := target-http
COMPONENT_DIR_target-grpc := target-grpc

BIN_NAME_master      := chaos-master
BIN_NAME_agent       := chaos-agent
BIN_NAME_client      := chaos-client
BIN_NAME_target-http := target-http
BIN_NAME_target-grpc := target-grpc

# Components that have a container image, the image tag they produce and the
# Dockerfile that builds them. The Dockerfiles are multi-stage, so building an
# image does not require `make all` first.
IMAGE_COMPONENTS := master agent target-http target-grpc

IMAGE_NAME_master      := chaos-master
IMAGE_NAME_agent       := chaos-agent
IMAGE_NAME_target-http := target-http
IMAGE_NAME_target-grpc := target-grpc

DOCKERFILE_master      := build/dockerfile-chaos-master
DOCKERFILE_agent       := build/dockerfile-chaos-agent
DOCKERFILE_target-http := build/dockerfile-target-http
DOCKERFILE_target-grpc := build/dockerfile-target-grpc

# --- Web / images -----------------------------------------------------------
BASE_HREF ?= /
IMAGE_TAG ?= latest
# REGISTRY gets a default further down, once the container engine is known.

# --- kind cluster -----------------------------------------------------------
KIND_CLUSTER     ?= chaos
KIND_NAMESPACE   := chaos-testing
MANIFEST_DIR     := manifests/kind
GENERATED_DIR    := manifests/generated
CONTAINER_ENGINE ?= $(shell command -v docker >/dev/null 2>&1 && echo docker || echo podman)

# kind needs the experimental flag to drive podman instead of docker. The engine
# is often podman reached through a `docker` shim (Bazzite, podman-docker), so ask
# the binary what it is rather than trusting its name.
ENGINE_IS_PODMAN := $(shell $(DOCKER) --version 2>/dev/null | grep -qi podman && echo yes || echo no)

ifeq ($(ENGINE_IS_PODMAN),yes)
KIND_ENV := KIND_EXPERIMENTAL_PROVIDER=podman
# Podman stores unqualified image names under a localhost/ prefix. Without that
# prefix here, the images loaded into the node would be called
# localhost/chaos-master:latest while the pods asked for chaos-master:latest, and
# every pod would fail with ImagePullBackOff trying to reach Docker Hub.
REGISTRY ?= localhost/
else
KIND_ENV :=
REGISTRY ?=
endif

KIND_IMAGES := $(foreach c,$(IMAGE_COMPONENTS),$(REGISTRY)$(IMAGE_NAME_$(c)):$(IMAGE_TAG))

# --- Local ports ------------------------------------------------------------
# 8080 is frequently taken on a workstation, so the web UI defaults to 9001.
WEB_UI_PORT     ?= 9001
PROMETHEUS_PORT ?= 9091
GRAFANA_PORT    ?= 3000
GRPC_PORT       ?= 9002

# Ports the services listen on inside the cluster.
CLUSTER_WEB_PORT        := 8080
CLUSTER_PROMETHEUS_PORT := 9090
CLUSTER_GRAFANA_PORT    := 3000
CLUSTER_GRPC_PORT       := 9002

# Optional passthrough for running targets on the host, for example
# HOST_RUNNER=flatpak-spawn --host. The dev container needs
# --talk-name=org.freedesktop.Flatpak in its runArgs for this to work.
HOST_RUNNER ?= flatpak-spawn --host

# --- Help -------------------------------------------------------------------
.PHONY: help
help:
	@awk 'NR == 1 { print; next } /^#/ { sub(/^# ?/, ""); print; next } /^[[:space:]]*$$/ { print ""; next } { exit }' $(firstword $(MAKEFILE_LIST))

.PHONY: env
env:
	@echo "cluster      kind cluster '$(KIND_CLUSTER)', namespace '$(KIND_NAMESPACE)'"
	@echo "engine       $(CONTAINER_ENGINE)"
	@echo "base href    $(BASE_HREF)"
	@echo "images       $(foreach i,$(KIND_IMAGES),$(i) )"
	@echo "proxy ports  web UI $(WEB_UI_PORT) -> $(CLUSTER_WEB_PORT)"
	@echo "             Prometheus $(PROMETHEUS_PORT) -> $(CLUSTER_PROMETHEUS_PORT)"
	@echo "             Grafana $(GRAFANA_PORT) -> $(CLUSTER_GRAFANA_PORT)"
	@echo "             gRPC $(GRPC_PORT) -> $(CLUSTER_GRPC_PORT) (make proxy-grpc)"

# --- Host guard -------------------------------------------------------------
# kind spawns privileged containers, which does not work nested inside the dev
# container. Targets that talk to the cluster therefore only run on the host.
define require_host
@if [ -e /.dockerenv ] || [ -e /run/.containerenv ] || [ -n "$${REMOTE_CONTAINERS:-}" ]; then \
	echo "'make $(1)' must run on the host: kind cannot run inside a container."; \
	echo; \
	echo "Open a host terminal and run:"; \
	echo "  cd $(CURDIR) && make $(1)"; \
	echo; \
	echo "The dev container is still fine for building, testing and 'make manifests'."; \
	exit 1; \
fi
endef

# --- Build ------------------------------------------------------------------
.PHONY: all $(COMPONENTS)
all: $(COMPONENTS)

# One rule for every component: the target name picks the cmd directory and the
# binary name. These stay .PHONY so a build always refreshes the binary.
$(COMPONENTS): proto
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(BIN_NAME_$@) ./cmd/$(COMPONENT_DIR_$@)

# --- Protobuf ---------------------------------------------------------------
.PHONY: proto proto-dart tools
proto: $(PROTO_DIR)/master.pb.go $(PROTO_DIR)/target.pb.go $(PROTO_DIR)/auth.pb.go

# Each protoc invocation produces both the message and the gRPC stubs.
$(PROTO_DIR)/master.pb.go $(PROTO_DIR)/master_grpc.pb.go &: $(PROTO_DIR)/master.proto
	$(PROTOC) $(PROTOC_FLAGS) $<

$(PROTO_DIR)/target.pb.go $(PROTO_DIR)/target_grpc.pb.go &: $(PROTO_DIR)/target.proto
	$(PROTOC) $(PROTOC_FLAGS) $<

$(PROTO_DIR)/auth.pb.go $(PROTO_DIR)/auth_grpc.pb.go &: $(PROTO_DIR)/auth.proto
	$(PROTOC) $(PROTOC_FLAGS) $<

# The dev container already ships all three plugins; this is for a bare host.
tools:
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	dart pub global activate protoc_plugin 25.1.0
	@echo "Installed protoc-gen-go, protoc-gen-go-grpc and protoc-gen-dart."

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

# --- Tests, lint and formatting ---------------------------------------------
.PHONY: test test-go test-web lint lint-go lint-web fmt fmt-check tidy ci

test: test-go test-web
test-go:
	$(GO) test ./...
test-web:
	cd $(WEB_DIR) && $(FLUTTER) test

lint: lint-go lint-web
lint-go:
	$(GO) vet ./...
lint-web:
	cd $(WEB_DIR) && $(FLUTTER) analyze

fmt:
	$(GO) fmt ./...
	cd $(WEB_DIR) && $(DART) format .

fmt-check:
	@unformatted=$$(gofmt -l cmd internal); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed for:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@echo "Go sources are formatted."
	cd $(WEB_DIR) && $(DART) format --output=none --set-exit-if-changed .

tidy:
	$(GO) mod tidy

ci: proto fmt-check lint test docker-build

# --- Container images -------------------------------------------------------
.PHONY: docker-build
docker-build: $(addprefix docker-build-,$(IMAGE_COMPONENTS))

# IMAGE_COMPONENTS are deliberately not .PHONY: a phony target is never matched
# by a pattern rule, which is what provides the recipe below.
docker-build-%:
	$(DOCKER) build -f $(DOCKERFILE_$*) $(BUILD_ARGS_$*) -t $(REGISTRY)$(IMAGE_NAME_$*):$(IMAGE_TAG) .

docker-build-master: BUILD_ARGS_master := --build-arg FLUTTER_BASE_HREF=$(BASE_HREF)

# --- Certificates -----------------------------------------------------------
# One self-signed certificate serves every TLS target. The SANs cover localhost
# and the in-cluster service names. The agents skip verification, so this is for
# clients that do verify.
.PHONY: generate-certs
generate-certs:
	@mkdir -p $(CERT_DIR)
	@if [ ! -f $(CERT_DIR)/tls.crt ]; then \
		echo "Generating a self-signed certificate in $(CERT_DIR)"; \
		openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
			-keyout $(CERT_DIR)/tls.key -out $(CERT_DIR)/tls.crt \
			-subj "/CN=chaos-target" \
			-addext "subjectAltName=DNS:localhost,DNS:target-http,DNS:target-http-tls,DNS:target-http2,DNS:target-grpc,DNS:*.$(KIND_NAMESPACE).svc.cluster.local,IP:127.0.0.1"; \
	fi

# --- Cluster secrets --------------------------------------------------------
.PHONY: kind-secrets
kind-secrets: generate-certs
	@$(KUBECTL) -n $(KIND_NAMESPACE) get namespace $(KIND_NAMESPACE) >/dev/null 2>&1 || { \
		echo "namespace $(KIND_NAMESPACE) does not exist - run 'make kind-up' first."; exit 1; }
	@$(KUBECTL) -n $(KIND_NAMESPACE) create secret tls target-tls \
		--cert=$(CERT_DIR)/tls.crt --key=$(CERT_DIR)/tls.key \
		--dry-run=client -o yaml | $(KUBECTL) -n $(KIND_NAMESPACE) apply -f -
	@echo "Applied secret $(KIND_NAMESPACE)/target-tls"

# --- Manifests --------------------------------------------------------------
.PHONY: manifests kind-preflight kind-create kind-images kind-deploy kind-verify \
        kind-up kind-status kind-logs kind-restart kind-down kind-delete kind-clean

# Render manifests/kind into manifests/generated using the image names and tags
# that `make docker-build` produces. Needs no cluster, so it also runs inside the
# dev container.
manifests:
	@rm -rf $(GENERATED_DIR)
	@mkdir -p $(GENERATED_DIR)/server $(GENERATED_DIR)/agent $(GENERATED_DIR)/target $(GENERATED_DIR)/monitoring
	@cp $(MANIFEST_DIR)/namespace.yaml $(GENERATED_DIR)/namespace.yaml
	@cp $(MANIFEST_DIR)/server/*.yaml $(GENERATED_DIR)/server/
	@cp $(MANIFEST_DIR)/agent/*.yaml $(GENERATED_DIR)/agent/
	@cp $(MANIFEST_DIR)/target/*.yaml $(GENERATED_DIR)/target/
	@cp $(MANIFEST_DIR)/monitoring/*.yaml $(GENERATED_DIR)/monitoring/
	@cp $(MANIFEST_DIR)/kustomization.template.yaml $(GENERATED_DIR)/kustomization.yaml
	@find $(GENERATED_DIR) -name '*.yaml' -print0 | xargs -0 sed -i \
		-e 's|__NAMESPACE__|$(KIND_NAMESPACE)|g' \
		-e 's|__REGISTRY__|$(REGISTRY)|g' \
		-e 's|__IMAGE_TAG__|$(IMAGE_TAG)|g'
	@# Grafana provisioning ConfigMaps generated from monitoring/, so the
	@# dashboards and the datasource stay single-sourced.
	@$(KUBECTL) create configmap grafana-datasources --dry-run=client -o yaml \
		--from-file=prometheus.yml=monitoring/grafana/provisioning/datasources/prometheus.yml \
		> $(GENERATED_DIR)/monitoring/grafana-datasources.yaml
	@$(KUBECTL) create configmap grafana-dashboards --dry-run=client -o yaml \
		--from-file=dashboard.yml=monitoring/grafana/provisioning/dashboards/dashboard.yml \
		--from-file=chaos-comparison.json=monitoring/grafana/provisioning/dashboards/chaos-comparison.json \
		--from-file=chaos-target-http.json=monitoring/grafana/provisioning/dashboards/chaos-target-http.json \
		> $(GENERATED_DIR)/monitoring/grafana-dashboards.yaml
	@$(KUBECTL) kustomize $(GENERATED_DIR) > $(GENERATED_DIR)/kind-all.yaml
	@$(KUBECTL) kustomize $(GENERATED_DIR)/server > $(GENERATED_DIR)/kind-server.yaml
	@$(KUBECTL) kustomize $(GENERATED_DIR)/agent > $(GENERATED_DIR)/kind-agent.yaml
	@$(KUBECTL) kustomize $(GENERATED_DIR)/target > $(GENERATED_DIR)/kind-target.yaml
	@$(KUBECTL) kustomize $(GENERATED_DIR)/monitoring > $(GENERATED_DIR)/kind-monitoring.yaml
	@echo "Generated kind manifests in $(GENERATED_DIR):"
	@ls -1 $(GENERATED_DIR)/kind-*.yaml | sed 's|^|  |'

kind-preflight:
	$(call require_host,kind-up)
	@command -v $(KIND) >/dev/null 2>&1 || { \
		echo "$(KIND) not found - install it on the host (https://kind.sigs.k8s.io/)."; exit 1; }
	@command -v $(KUBECTL) >/dev/null 2>&1 || { \
		echo "$(KUBECTL) not found - install it on the host."; exit 1; }
	@command -v $(CONTAINER_ENGINE) >/dev/null 2>&1 || { \
		echo "$(CONTAINER_ENGINE) not found - install it on the host, or pass CONTAINER_ENGINE=docker."; exit 1; }
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
	@$(MAKE) --no-print-directory docker-build DOCKER=$(CONTAINER_ENGINE)
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

# The master is applied first and allowed to become ready before the agents, so
# the agents do not crash-loop against a master that is not listening yet.
kind-deploy: manifests kind-create kind-secrets
	@echo "Deploying the master..."
	@$(KUBECTL) apply -f $(GENERATED_DIR)/kind-server.yaml
	@$(KUBECTL) -n $(KIND_NAMESPACE) rollout status deployment/chaos-master --timeout=180s
	@echo "Deploying the agents, the targets and the monitoring stack..."
	@$(KUBECTL) apply -f $(GENERATED_DIR)/kind-all.yaml
	@for d in chaos-agent target-http target-http-tls target-http2 target-grpc prometheus grafana; do \
		echo "Waiting for deployment/$$d..."; \
		$(KUBECTL) -n $(KIND_NAMESPACE) rollout status deployment/$$d --timeout=180s; \
	done
	@$(MAKE) --no-print-directory kind-status

kind-verify:
	@echo "Agents connected to the master:"
	@$(KUBECTL) -n $(KIND_NAMESPACE) logs -l app=chaos-master --tail=200 2>/dev/null \
		| grep "Agent connected" || echo "  (none yet)"

kind-up: kind-create kind-images kind-deploy
	@echo
	@echo "kind cluster '$(KIND_CLUSTER)' is running (namespace $(KIND_NAMESPACE))."
	@echo "  Access      make proxy"
	@echo "                web UI     http://localhost:$(WEB_UI_PORT)/"
	@echo "                Prometheus http://localhost:$(PROMETHEUS_PORT)/"
	@echo "                Grafana    http://localhost:$(GRAFANA_PORT)/ (admin/admin)"
	@echo "  Targets     http://target-http.$(KIND_NAMESPACE).svc.cluster.local:8080"
	@echo "              https://target-http-tls.$(KIND_NAMESPACE).svc.cluster.local"
	@echo "              https://target-http2.$(KIND_NAMESPACE).svc.cluster.local"
	@echo "              target-grpc.$(KIND_NAMESPACE).svc.cluster.local:9000"
	@echo "  Register    $(KUBECTL) -n $(KIND_NAMESPACE) get secret chaos-master-secrets -o jsonpath='{.data.registerSecret}' | base64 -d"
	@echo "  Status      make kind-status    Logs: make kind-logs"
	@echo "  Teardown    make kind-down | kind-delete | kind-clean"

kind-status:
	$(call require_host,kind-status)
	@$(KUBECTL) -n $(KIND_NAMESPACE) get deployments,pods,svc 2>/dev/null \
		|| echo "Cannot reach kind cluster '$(KIND_CLUSTER)' - run 'make kind-up'."

kind-logs:
	$(call require_host,kind-logs)
	@$(KUBECTL) -n $(KIND_NAMESPACE) logs -l app=chaos-master --tail=50 --prefix
	@$(KUBECTL) -n $(KIND_NAMESPACE) logs -l app=chaos-agent --tail=50 --prefix
	@$(KUBECTL) -n $(KIND_NAMESPACE) logs -l app=target-http --tail=50 --prefix
	@$(KUBECTL) -n $(KIND_NAMESPACE) logs -l app=prometheus --tail=50 --prefix

# Reload freshly built images without recreating the cluster.
kind-restart:
	$(call require_host,kind-restart)
	@$(KUBECTL) -n $(KIND_NAMESPACE) rollout restart \
		deployment/chaos-master deployment/chaos-agent deployment/target-http \
		deployment/target-http-tls deployment/target-http2 deployment/target-grpc \
		deployment/prometheus deployment/grafana
	@for d in chaos-master chaos-agent target-http target-http-tls target-http2 target-grpc prometheus grafana; do \
		$(KUBECTL) -n $(KIND_NAMESPACE) rollout status deployment/$$d --timeout=180s; \
	done
	@echo "Rolled out the Chaos stack."

kind-down:
	$(call require_host,kind-down)
	@if [ -f $(GENERATED_DIR)/kind-all.yaml ]; then \
		$(KUBECTL) delete -f $(GENERATED_DIR)/kind-all.yaml --ignore-not-found; \
	else \
		$(KUBECTL) delete namespace $(KIND_NAMESPACE) --ignore-not-found; \
	fi
	@echo "Removed the Chaos resources (kind cluster '$(KIND_CLUSTER)' is still running)."

kind-delete:
	$(call require_host,kind-delete)
	@$(KIND) delete cluster --name $(KIND_CLUSTER) || true

kind-clean: kind-delete
	@rm -rf $(GENERATED_DIR)
	@echo "Removed kind cluster '$(KIND_CLUSTER)' and $(GENERATED_DIR)."

# --- Access to the cluster --------------------------------------------------
# kubectl can only reach a kind cluster from the host, so the proxy targets run
# there too. Multiple port-forwards are started as background jobs and torn down
# together, which avoids depending on kubectl's multi-resource syntax.
.PHONY: cluster-preflight proxy proxy-web proxy-monitoring proxy-grpc kind-port-forward

cluster-preflight:
	$(call require_host,proxy)
	@command -v $(KUBECTL) >/dev/null 2>&1 || { \
		echo "$(KUBECTL) not found - install it on the host."; exit 1; }
	@$(KUBECTL) -n $(KIND_NAMESPACE) get service chaos-master-service >/dev/null 2>&1 || { \
		echo "Cannot reach the '$(KIND_NAMESPACE)' namespace in kind cluster '$(KIND_CLUSTER)'."; \
		echo "Create and deploy it first: make kind-up"; \
		exit 1; }

proxy: cluster-preflight
	@echo "Forwarding the Chaos stack (Ctrl-C to stop):"
	@echo "  web UI      http://localhost:$(WEB_UI_PORT)/"
	@echo "  Prometheus  http://localhost:$(PROMETHEUS_PORT)/"
	@echo "  Grafana     http://localhost:$(GRAFANA_PORT)/  (admin/admin)"
	@echo
	@pids=""; \
	forward() { $(KUBECTL) -n $(KIND_NAMESPACE) port-forward "$$@" >/dev/null & pids="$$pids $$!"; }; \
	trap 'kill $$pids 2>/dev/null || true' INT TERM; \
	forward svc/chaos-master-service $(WEB_UI_PORT):$(CLUSTER_WEB_PORT); \
	forward svc/prometheus $(PROMETHEUS_PORT):$(CLUSTER_PROMETHEUS_PORT); \
	forward svc/grafana $(GRAFANA_PORT):$(CLUSTER_GRAFANA_PORT); \
	sleep 2; \
	for p in $$pids; do \
		if ! kill -0 "$$p" 2>/dev/null; then \
			echo "A port-forward failed to start. Check that the stack is deployed (make kind-deploy)"; \
			echo "and that $(WEB_UI_PORT)/$(PROMETHEUS_PORT)/$(GRAFANA_PORT) are free on this machine."; \
			kill $$pids 2>/dev/null || true; \
			exit 1; \
		fi; \
	done; \
	wait || true; \
	kill $$pids 2>/dev/null || true

proxy-web: cluster-preflight
	@echo "web UI: http://localhost:$(WEB_UI_PORT)/ (Ctrl-C to stop)"
	@$(KUBECTL) -n $(KIND_NAMESPACE) port-forward svc/chaos-master-service $(WEB_UI_PORT):$(CLUSTER_WEB_PORT)

proxy-monitoring: cluster-preflight
	@echo "Prometheus: http://localhost:$(PROMETHEUS_PORT)/"
	@echo "Grafana:    http://localhost:$(GRAFANA_PORT)/ (admin/admin)"
	@pids=""; \
	forward() { $(KUBECTL) -n $(KIND_NAMESPACE) port-forward "$$@" >/dev/null & pids="$$pids $$!"; }; \
	trap 'kill $$pids 2>/dev/null || true' INT TERM; \
	forward svc/prometheus $(PROMETHEUS_PORT):$(CLUSTER_PROMETHEUS_PORT); \
	forward svc/grafana $(GRAFANA_PORT):$(CLUSTER_GRAFANA_PORT); \
	sleep 2; \
	for p in $$pids; do \
		if ! kill -0 "$$p" 2>/dev/null; then \
			echo "A port-forward failed to start. Check that the stack is deployed (make kind-deploy)"; \
			echo "and that $(PROMETHEUS_PORT)/$(GRAFANA_PORT) are free on this machine."; \
			kill $$pids 2>/dev/null || true; \
			exit 1; \
		fi; \
	done; \
	wait || true; \
	kill $$pids 2>/dev/null || true

proxy-grpc: cluster-preflight
	@echo "master gRPC: localhost:$(GRPC_PORT) (Ctrl-C to stop)"
	@echo "Point the client at it with:"
	@echo "  CHAOS_MASTER_ADDRESS=localhost:$(GRPC_PORT) ./bin/chaos-client"
	@$(KUBECTL) -n $(KIND_NAMESPACE) port-forward svc/chaos-master-service $(GRPC_PORT):$(CLUSTER_GRPC_PORT)

# Kept so older notes and links keep working.
kind-port-forward: proxy

# --- Run a target on the host ------------------------------------------------
# For sessions that can drive the host, for example
#   make host TARGET=proxy
.PHONY: host
host:
	@test -n "$(TARGET)" || { echo "usage: make host TARGET=proxy"; exit 1; }
	@command -v $(firstword $(HOST_RUNNER)) >/dev/null 2>&1 || { \
		echo "$(firstword $(HOST_RUNNER)) not found:"; \
		echo "  add \"--talk-name=org.freedesktop.Flatpak\" to runArgs in .devcontainer/devcontainer.json,"; \
		echo "  or run 'make $(TARGET)' in a host terminal."; \
		exit 1; }
	$(HOST_RUNNER) bash -lc 'cd "$(CURDIR)" && make $(TARGET)'

# --- Clean ------------------------------------------------------------------
.PHONY: clean clean-web clean-proto clean-certs
clean: clean-web clean-proto clean-certs
	rm -rf $(BIN_DIR)
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
