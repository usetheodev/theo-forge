#!/usr/bin/env bash
# Bootstrap an E2E environment: kind cluster + Argo Workflows install.
#
# Usage:
#   ./scripts/e2e-up.sh         # idempotent — reuses existing cluster
#   ./scripts/e2e-up.sh --fresh # tear down + recreate
#
# Side effects:
#   - Creates kind cluster named ${KIND_CLUSTER}.
#   - Writes kubeconfig to ${KUBECONFIG_E2E}.
#   - Installs Argo Workflows from the pinned manifest.
#   - Starts a background `kubectl port-forward` of argo-server on ${ARGO_PORT}.
#   - Writes a service-account bearer token to ${ARGO_TOKEN_FILE}.
#
# Exit codes:
#   0 — cluster ready, argo healthy, token + kubeconfig written.
#   1 — bootstrap failed; partial state may exist (call e2e-down.sh).

set -euo pipefail

# shellcheck disable=SC1091
source "$(dirname "$0")/e2e-versions.sh"

mkdir -p "$(dirname "${ARGO_TOKEN_FILE}")"

log() { printf '==> %s\n' "$*" >&2; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

ensure_kind() {
  if command -v kind >/dev/null 2>&1; then
    log "kind already on PATH: $(kind version)"
    return
  fi
  log "kind not found — installing ${KIND_VERSION} to \$GOPATH/bin"
  GOBIN=$(go env GOPATH)/bin go install "sigs.k8s.io/kind@${KIND_VERSION}"
  export PATH="$(go env GOPATH)/bin:${PATH}"
}

ensure_kubectl() {
  command -v kubectl >/dev/null 2>&1 || die "kubectl not found on PATH"
}

fresh=0
case "${1:-}" in
  --fresh) fresh=1 ;;
  '') ;;
  *) die "unknown arg: $1" ;;
esac

ensure_kind
ensure_kubectl

if [[ "${fresh}" == "1" ]]; then
  log "tearing down existing cluster (--fresh)"
  kind delete cluster --name "${KIND_CLUSTER}" >/dev/null 2>&1 || true
fi

if kind get clusters 2>/dev/null | grep -qx "${KIND_CLUSTER}"; then
  log "kind cluster ${KIND_CLUSTER} already exists — reusing"
else
  log "creating kind cluster ${KIND_CLUSTER} (image: ${KIND_NODE_IMAGE})"
  kind create cluster \
    --name "${KIND_CLUSTER}" \
    --image "${KIND_NODE_IMAGE}" \
    --kubeconfig "${KUBECONFIG_E2E}" \
    --wait 120s
fi

# Always emit a fresh kubeconfig snapshot for the tests to consume.
kind get kubeconfig --name "${KIND_CLUSTER}" >"${KUBECONFIG_E2E}"
export KUBECONFIG="${KUBECONFIG_E2E}"

# --- Install Argo Workflows ---

if kubectl get ns "${ARGO_NAMESPACE}" >/dev/null 2>&1; then
  log "namespace ${ARGO_NAMESPACE} exists — refreshing manifest"
else
  log "creating namespace ${ARGO_NAMESPACE}"
  kubectl create namespace "${ARGO_NAMESPACE}"
fi

log "applying Argo Workflows ${ARGO_VERSION} manifest"
# --server-side is required for v4.x: CRDs ship with >262KB of OpenAPI
# annotations which exceeds the client-side `last-applied-configuration`
# annotation limit. SSA stores the field map server-side instead.
# `--force-conflicts` lets re-apply replace conflicting fields owned by
# a previous run (idempotent for `make e2e-up`).
kubectl apply -n "${ARGO_NAMESPACE}" --server-side --force-conflicts -f "${ARGO_MANIFEST_URL}" >/dev/null

log "waiting for argo-server + workflow-controller to be Ready (${ARGO_READY_TIMEOUT}s)"
kubectl -n "${ARGO_NAMESPACE}" wait --for=condition=Available --timeout="${ARGO_READY_TIMEOUT}s" \
  deployment/argo-server deployment/workflow-controller

# --- Service account + bearer token for the REST client ---

if ! kubectl -n "${ARGO_NAMESPACE}" get sa theo-forge-e2e >/dev/null 2>&1; then
  log "creating ServiceAccount + ClusterRoleBinding for the e2e client"
  kubectl -n "${ARGO_NAMESPACE}" create sa theo-forge-e2e
  kubectl create clusterrolebinding theo-forge-e2e-admin \
    --clusterrole=cluster-admin \
    --serviceaccount="${ARGO_NAMESPACE}:theo-forge-e2e" >/dev/null
fi

# Kubernetes 1.24+ does not auto-issue tokens for SAs; mint one ourselves.
log "minting bearer token for ServiceAccount theo-forge-e2e"
kubectl -n "${ARGO_NAMESPACE}" create token theo-forge-e2e --duration=24h >"${ARGO_TOKEN_FILE}"
chmod 600 "${ARGO_TOKEN_FILE}"

# --- Port-forward argo-server for the REST client to talk to ---

# Kill stale forwarders pointing at the same port.
pkill -f "kubectl.*port-forward.*svc/argo-server" >/dev/null 2>&1 || true
sleep 1
log "starting background kubectl port-forward svc/argo-server :${ARGO_PORT}"
nohup kubectl -n "${ARGO_NAMESPACE}" port-forward svc/argo-server \
  "${ARGO_PORT}:2746" >.e2e/port-forward.log 2>&1 &
echo $! >.e2e/port-forward.pid

# Wait for the forwarder to actually accept connections.
for _ in $(seq 1 30); do
  if curl -ks "https://localhost:${ARGO_PORT}/" >/dev/null 2>&1 || \
     curl -s "http://localhost:${ARGO_PORT}/" >/dev/null 2>&1; then
    log "argo-server reachable on :${ARGO_PORT}"
    break
  fi
  sleep 1
done

log "E2E environment ready"
log "  cluster:    ${KIND_CLUSTER}"
log "  argo:       ${ARGO_VERSION} (ns=${ARGO_NAMESPACE})"
log "  kubeconfig: ${KUBECONFIG_E2E}"
log "  token:      ${ARGO_TOKEN_FILE}"
log "  server:     https://localhost:${ARGO_PORT}"
log "next: make e2e   (or: go test -tags=e2e ./e2e/...)"
