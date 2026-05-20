// Package e2e holds end-to-end tests for the theo-forge SDK.
//
// These tests submit workflows to a REAL Argo Workflows control plane
// (running on a local kind cluster bootstrapped by `make e2e-up`) and
// assert on observable side effects (workflow phase, pod placement,
// emitted outputs).
//
// Why a separate package + build tag:
//   - Setup is slow (~60s for kind + argo install) and requires Docker.
//     We do NOT want this in the default `go test ./...` path.
//   - Tests depend on a kubeconfig + bearer token written by
//     `scripts/e2e-up.sh`; they fail loudly if missing.
//
// Run locally:
//
//	make e2e        # bootstrap + run + teardown
//	make e2e-keep   # bootstrap + run, keep cluster for debugging
//	make e2e-down   # destroy cluster
//
// Run a single test:
//
//	make e2e-up
//	KUBECONFIG=$(pwd)/.e2e/kubeconfig \
//	ARGO_TOKEN_FILE=$(pwd)/.e2e/argo-token \
//	  go test -tags=e2e -v -run TestE2E_HelloWorld ./e2e/...
//
// See docs/E2E.md for the full guide.
//
//go:build e2e

package e2e
