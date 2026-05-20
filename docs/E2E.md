# E2E Tests — theo-forge

End-to-end tests submit workflows to a **real Argo Workflows control plane**
running on a local [kind](https://kind.sigs.k8s.io/) cluster and assert on
observable side effects (workflow phase, pod placement, emitted outputs).

This is the **L3** (gold-standard) fidelity level — the highest available
without paying for a hosted cluster.

## Why L3 and not mocks

The httptest-based "integration" tests in `client/*_test.go` mock the
Argo REST API: handler shape is fixed, status codes are hand-coded, and
nothing exercises the controller, scheduler, or kubelet. They catch
**client serialization bugs** but cannot catch:

- Schema drift between Argo versions (a field renamed in v3.6 still
  serializes happily on our side; the cluster rejects it).
- Admission webhook rejection (workflow-controller validating semantics
  the SDK ignores).
- Controller normalization (default values applied that change what
  the consumer sees).
- Actual execution (pods scheduled, args reach the container, output
  parameters populated).

E2E tests close all four gaps.

## Versions (pinned)

See [`scripts/e2e-versions.sh`](../scripts/e2e-versions.sh):

| Component | Version | Why this pin |
|-----------|---------|--------------|
| Argo Workflows | `v4.0.3` | Matches the appVersion shipped by argo-helm chart 1.0.5 — the version Theo Cloud installs in production (`infra/argocd-bootstrap/apps/build-argo-workflows.yaml`). Keeping E2E aligned with prod prevents v3↔v4 schema drift surprises. |
| kind | `v0.24.0` | Last release before kind started auto-bumping to k8s 1.32 (which has CRD changes we have not validated). |
| kind node image | `kindest/node:v1.31.0` | k8s 1.31 — matches Argo v3.5 support matrix. |

Bumping any of these is intentional and the round-trip suite
(`TestRoundTripTestdataExamples`) MUST still pass.

## Running

```bash
# Full cycle: kind up → install Argo → run e2e → kind down
make e2e

# Keep cluster around for debugging
make e2e-keep
# ... iterate ...
make e2e-down

# Just the cluster (no tests)
make e2e-up

# Recreate from scratch (rare; only when state is suspicious)
make e2e-up-fresh
```

## Prerequisites

- `docker` (any 24+).
- `kubectl` on PATH.
- `kind` on PATH (or `go` available — `e2e-up.sh` will `go install`
  the pinned version into `$GOPATH/bin`).
- A free port `2746` on the host (override with `ARGO_PORT=NNNN`).

## Architecture

```
host                                        kind cluster (k8s 1.31)
┌───────────────────┐                       ┌─────────────────────────┐
│ go test -tags=e2e │                       │  ns: argo               │
│   helloworld_test │  port-forward :2746   │  ┌───────────────────┐  │
│   diamond_test    │ ────────────────────► │  │ argo-server       │  │
│   hooks_test      │                       │  │ workflow-ctrl     │  │
│   pvc_aff_test    │                       │  │ minio (artifacts) │  │
│   lifecycle_test  │ ◄──── kubectl ──────► │  └───────────────────┘  │
└───────────────────┘    (apply, get,       │                         │
                          watch, delete)    │  + ServiceAccount       │
        ▲                                   │    theo-forge-e2e       │
        │                                   │    (cluster-admin RBAC) │
        │                                   └─────────────────────────┘
   ARGO_TOKEN_FILE                                 ▲
   KUBECONFIG                                      │
        │                                          │
        └─── written by scripts/e2e-up.sh ─────────┘
```

The bearer token lives at `.e2e/argo-token` (mode 0600, .gitignored).
Tests read it via `os.ReadFile`; you can re-mint it with
`kubectl -n argo create token theo-forge-e2e --duration=24h`.

## Test inventory

| File | Test | Asserts | Closes |
|------|------|---------|--------|
| `helloworld_test.go` | `TestE2E_HelloWorld` | Happy path: Build → Submit → Succeeded | smoke / regression for the README quickstart |
| `dag_diamond_test.go` | `TestE2E_DAGDiamond` | 4-task diamond DAG completes, all task nodes present | Task.Then() emission + DAG.AddTasks ordering (T3.9) |
| `hooks_isolation_test.go` | `TestE2E_NewConfigHookIsolation` | Hook on `NewConfig()` fires AND label survives to the cluster manifest | ADR-001 / val-004 (the most important v0.5.0 fix) |
| `pvc_affinity_test.go` | `TestE2E_PVCDefaultPodAffinity` | `Workflow.Build` injects podAffinity when PVC present + label survives to cluster | v0.4.0 PR #12 feature |
| `pvc_affinity_test.go` | `TestE2E_DisableDefaultAffinity` | Opt-out via `DisableDefaultAffinity=true` is honored | same feature, negative case |
| `client_lifecycle_test.go` | `TestE2E_ClientLifecycle` | Lint → Create → Get → List → Suspend → Resume → Stop → Delete → 404 | T4.11 client API end-to-end |
| `client_lifecycle_test.go` | `TestE2E_ClientNameValidation` | Malformed names rejected with `model.ErrInvalidName` | T2.3 / SEC-003 |

Each test:

1. Generates a unique workflow name (`uniqueName(prefix)`) to avoid
   collisions on rerun / parallel.
2. Registers a `t.Cleanup` that `kubectl delete wf` on exit (best-effort).
3. Submits + waits with `defaultWaitTimeout` (180s).
4. On failure: dumps the full cluster manifest into the test output.

## Adding a new E2E test

Two-file pattern, with the build tag at top:

```go
//go:build e2e

package e2e

import (
    "context"
    "testing"
    forge "github.com/usetheodev/theo-forge"
)

func TestE2E_MyFeature(t *testing.T) {
    name := uniqueName("my-feature")
    cleanupWorkflow(t, name)

    w := &forge.Workflow{ /* ... */ }
    wf, err := w.Build()
    if err != nil { t.Fatalf("Build: %v", err) }

    svc := argoClient(t)
    if _, err := svc.CreateWorkflowFromModel(context.Background(), wf, ""); err != nil {
        t.Fatalf("Create: %v", err)
    }
    final := waitWorkflowSucceeded(t, name, defaultWaitTimeout)
    // ... assert on final ...
}
```

Helpers available from `harness_test.go`:

- `requireEnv(t)` — checks KUBECONFIG + ARGO_TOKEN_FILE.
- `argoClient(t)` — `*client.WorkflowsService` bound to the live server.
- `kubectl(t, args...)` — runs kubectl against the e2e kubeconfig.
- `uniqueName(prefix)` — collision-free name.
- `cleanupWorkflow(t, name)` — registers post-test deletion.
- `waitWorkflowSucceeded(t, name, timeout) model.WorkflowModel`.
- `dumpWorkflow(t, name) string` — full cluster YAML for assertion messages.

## CI

Currently **local-only** (`make e2e`). To enable in GitHub Actions:

```yaml
- uses: helm/kind-action@<sha>  # see scripts/e2e-versions.sh for the pin
  with:
    cluster_name: theo-forge-e2e
    node_image: kindest/node:v1.31.0
- name: install argo + run e2e
  run: |
    kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.5.10/quick-start-minimal.yaml
    kubectl -n argo wait --for=condition=Available --timeout=300s deployment --all
    # ... mint token, port-forward, run go test -tags=e2e ...
```

Expected runtime: **~5-7 min** on a fresh runner (cluster bootstrap
+ image pulls dominate). Subsequent runs in the same job: **~90 s**.

## Troubleshooting

- **"E2E environment not ready: KUBECONFIG not set"** — you forgot
  `make e2e-up` (or it failed). Inspect `.e2e/port-forward.log`.
- **Workflow stuck Pending** — the kind node likely cannot pull
  `alpine:3.18`. Check `kubectl -n argo describe pod <name>`.
- **`x509: certificate signed by unknown authority`** — the port-forward
  is plaintext but a test built the client with `VerifySSL=true`. Use
  `argoClient(t)` which sets `VerifySSL=false` for the test env.
- **`make e2e` succeeds but `make verify` fails on coverage** — E2E tests
  use a separate build tag and are excluded from coverage profiles by
  design. The two suites are independent.
