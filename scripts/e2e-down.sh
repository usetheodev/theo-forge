#!/usr/bin/env bash
# Tear down the E2E environment created by e2e-up.sh.
# Idempotent — safe to run when nothing exists.

set -euo pipefail

# shellcheck disable=SC1091
source "$(dirname "$0")/e2e-versions.sh"

log() { printf '==> %s\n' "$*" >&2; }

# Stop the background port-forwarder first; otherwise it logs noisy errors
# when the cluster goes away.
if [[ -f .e2e/port-forward.pid ]]; then
  pid="$(cat .e2e/port-forward.pid)"
  log "stopping port-forward (pid=${pid})"
  kill "${pid}" 2>/dev/null || true
  rm -f .e2e/port-forward.pid
fi
pkill -f "kubectl.*port-forward.*svc/argo-server" >/dev/null 2>&1 || true

if command -v kind >/dev/null 2>&1; then
  if kind get clusters 2>/dev/null | grep -qx "${KIND_CLUSTER}"; then
    log "deleting kind cluster ${KIND_CLUSTER}"
    kind delete cluster --name "${KIND_CLUSTER}"
  else
    log "no kind cluster named ${KIND_CLUSTER} — nothing to delete"
  fi
else
  log "kind not on PATH — skipping cluster deletion"
fi

rm -rf .e2e
log "teardown complete"
