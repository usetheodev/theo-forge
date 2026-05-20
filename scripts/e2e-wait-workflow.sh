#!/usr/bin/env bash
# Block until an Argo workflow reaches a terminal phase OR the timeout fires.
#
# Usage:
#   ./scripts/e2e-wait-workflow.sh <workflow-name> [namespace] [timeout-seconds]
#
# Exit codes:
#   0 — workflow Succeeded.
#   1 — workflow Failed / Error / Terminated.
#   2 — timeout.
#   3 — invalid args.

set -euo pipefail

# shellcheck disable=SC1091
source "$(dirname "$0")/e2e-versions.sh"

name="${1:-}"
ns="${2:-${ARGO_NAMESPACE}}"
timeout="${3:-180}"

[[ -n "${name}" ]] || { echo "usage: $0 <workflow-name> [ns] [timeout]" >&2; exit 3; }

export KUBECONFIG="${KUBECONFIG_E2E}"

deadline=$(( $(date +%s) + timeout ))
last_phase=""

while [[ $(date +%s) -lt "${deadline}" ]]; do
  phase="$(kubectl -n "${ns}" get wf "${name}" -o jsonpath='{.status.phase}' 2>/dev/null || true)"
  if [[ "${phase}" != "${last_phase}" ]]; then
    printf '  workflow %s: %s\n' "${name}" "${phase:-<not yet scheduled>}" >&2
    last_phase="${phase}"
  fi
  case "${phase}" in
    Succeeded) exit 0 ;;
    Failed|Error) exit 1 ;;
    Terminated) exit 1 ;;
  esac
  sleep 2
done

echo "TIMEOUT after ${timeout}s waiting for workflow ${name}" >&2
kubectl -n "${ns}" describe wf "${name}" >&2 || true
exit 2
