#!/usr/bin/env bash
# Pinned versions for the E2E suite. Sourced by every other e2e-*.sh script.
# Updating any value here is intentional and MUST be reflected in
# docs/E2E.md AND testdata/examples/VERSION.
#
# When bumping ARGO_VERSION:
#   1. Verify the install manifest still applies cleanly to kind v0.24+.
#   2. Re-run `make e2e` locally; fix any field-name drift.
#   3. Update testdata/examples/VERSION if the upstream examples were also bumped.

set -euo pipefail

# Argo Workflows — quickstart manifest (controller + server + minio + UI).
# https://github.com/argoproj/argo-workflows/releases
#
# Pinned to v4.0.3 because that is the appVersion shipped by argo-helm
# chart 1.0.5, which Theo Cloud installs in production via ArgoCD
# (infra/argocd-bootstrap/apps/build-argo-workflows.yaml).
# Bumping requires re-running `make e2e` to catch CRD/API drift.
export ARGO_VERSION="${ARGO_VERSION:-v4.0.3}"
export ARGO_MANIFEST_URL="https://github.com/argoproj/argo-workflows/releases/download/${ARGO_VERSION}/quick-start-minimal.yaml"

# kind — Kubernetes-in-Docker. Pinned to a release with stable Kubernetes 1.31 support.
export KIND_VERSION="${KIND_VERSION:-v0.24.0}"
export KIND_NODE_IMAGE="${KIND_NODE_IMAGE:-kindest/node:v1.31.0@sha256:53df588e04085fd41ae12de0c3fe4c72f7013bba32a20e7325357a1ac94ba865}"

# kind cluster name — kept short to avoid hitting Linux interface-name limits.
export KIND_CLUSTER="${KIND_CLUSTER:-theo-forge-e2e}"

# Argo namespace inside the cluster.
export ARGO_NAMESPACE="${ARGO_NAMESPACE:-argo}"

# Port that argo-server is forwarded to on the host. Tests connect to
# http://localhost:${ARGO_PORT}. Override if it clashes with something local.
export ARGO_PORT="${ARGO_PORT:-2746}"

# Token file written by e2e-up.sh; sourced by tests via $ARGO_TOKEN_FILE.
export ARGO_TOKEN_FILE="${ARGO_TOKEN_FILE:-$(pwd)/.e2e/argo-token}"

# kubeconfig file written by kind; tests use this instead of ~/.kube/config.
export KUBECONFIG_E2E="${KUBECONFIG_E2E:-$(pwd)/.e2e/kubeconfig}"

# Timeout for waiting argo-server to become ready (seconds).
export ARGO_READY_TIMEOUT="${ARGO_READY_TIMEOUT:-300}"
