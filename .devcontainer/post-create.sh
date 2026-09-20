#!/usr/bin/env bash
# Runs once after the dev container is created (see devcontainer.json).
# The heavy toolchain is baked into the image; this only primes per-project
# state (Go modules, Flutter packages) and reports what is available.
# Every step is best-effort so a flaky network does not block opening the repo.
set -u

log()  { printf '\n==> %s\n' "$*"; }
warn() { printf 'WARNING: %s\n' "$*" >&2; }

cd /workspaces/ChaosGenerator 2>/dev/null || cd /workspaces 2>/dev/null || true

log "Toolchain"
printf '  go      : %s\n' "$(go version 2>/dev/null | awk '{print $3}')"
printf '  protoc  : %s\n' "$(protoc --version 2>/dev/null)"
printf '  dart    : %s\n' "$(dart --version 2>&1 | head -1)"
printf '  kubectl : %s\n' "$(kubectl version --client 2>/dev/null | head -1)"
printf '  kind    : %s\n' "$(kind version 2>/dev/null)"
for p in protoc-gen-go protoc-gen-go-grpc protoc-gen-dart; do
    if command -v "$p" >/dev/null 2>&1; then
        printf '  %-8s: %s\n' "${p#protoc-gen-}" "$(command -v "$p")"
    else
        warn "$p not found on PATH"
    fi
done

log "Host container engine"
if command -v docker >/dev/null 2>&1; then
    if docker ps >/dev/null 2>&1; then
        echo "  host engine reachable via ${DOCKER_HOST:-<unset>}"
    else
        warn "cannot reach the host engine - is the host podman.socket running?"
        warn "  host fix: systemctl --user start podman.socket"
    fi
else
    warn "docker/podman client not found"
fi

log "Downloading Go modules"
go mod download && echo "  ok" || warn "go mod download failed (offline?)"

log "Generating protobuf stubs (make proto)"
if [ -f Makefile ]; then
    make proto >/dev/null 2>&1 && echo "  ok" || warn "make proto failed - run it manually"
fi

log "Flutter packages (web/)"
if [ -f web/pubspec.yaml ] && command -v flutter >/dev/null 2>&1; then
    (cd web && flutter pub get) && echo "  ok" || warn "flutter pub get failed"
fi

log "Verifying the project builds"
if [ -f go.mod ]; then
    go build ./... && echo "  BUILD OK" || warn "go build ./... failed - inspect output above"
fi

log "post-create finished."
cat <<'EOF'

Next steps:
  make all          # build every component
  make test         # go test + flutter test
  make manifests    # render manifests/kind (no cluster required)
  make docker-build # build images with the HOST engine

Runtime targets (make kind-up, make proxy, ...) must run on the HOST, because
kind cannot spawn containers inside this container. Run them in a host terminal.

EOF
