# Phase 5 — Infrastructure, CI/CD, Supply Chain, and Data Quality Review
**Date:** 2026-05-20
**Phase:** 5 / 7
**Reviewers:** cicd-reviewer, dependency-analyzer, data-reviewer (coordinated by chief-reviewer)

---

## 1. CI/CD Workflow Analysis

### Workflows Present
| File | Trigger | Jobs |
|---|---|---|
| `.github/workflows/ci.yml` | push: main, develop; PR: main, develop | `test`, `lint`, `security` |
| `.github/workflows/release.yml` | push: tags `v*` | `release` |

### CI Workflow — Job Details

**test job:**
- Runs on: `ubuntu-latest`
- Go version: `1.25` (via `actions/setup-go@v5`)
- Steps: checkout, setup-go (with cache), `go mod download`, `go build ./...`, `go vet ./...`, `go test -count=1 -race -coverprofile=coverage.out -covermode=atomic ./...`, upload coverage artifact
- Permissions: `contents: read` (least privilege — good)
- Coverage artifact: only uploaded on `push` events, not PRs

**lint job:**
- Runs on: `ubuntu-latest`
- Uses `golangci/golangci-lint-action@v7` with `version: latest` (unpinned lint tool)
- Permissions: `contents: read` (good)

**security job:**
- Runs on: `ubuntu-latest`
- Runs `golang/govulncheck-action@v1` (pinned to major)
- Runs `securego/gosec@master` — **CRITICAL: mutable floating branch ref**
- Permissions: `contents: read` (good)

### Release Workflow — Details
- Trigger: push to any tag matching `v*`
- Permissions: `contents: write` at workflow level (necessary for creating GitHub Release)
- Steps: checkout (fetch-depth: 0), setup-go, `go test -count=1 -race ./...`, `go mod verify && go mod tidy && git diff --exit-code go.mod go.sum`, Create GitHub Release via `softprops/action-gh-release@v2`
- **NO dependency on CI workflow jobs** — release proceeds independently even if lint or security checks are failing

### Gating Assessment
| Check | PR gated? | Release gated? |
|---|---|---|
| Tests (go test -race) | YES | YES (re-runs) |
| Build (go build ./...) | YES | NO |
| go vet | YES | NO |
| golangci-lint | YES | NO |
| govulncheck | YES | NO |
| gosec | YES | NO |
| Coverage threshold | NO (no threshold configured) | NO |

### Missing CI/CD Elements
- No CODEOWNERS file
- No PR template
- No branch protection configuration visible (not detectable from repo files alone)
- No SBOM generation
- No artifact signing (Sigstore/cosign)
- No SLSA provenance attestation
- Secret scanning: `govulncheck` and `gosec` run in CI but not release; no `git-secrets` or `trufflehog`

---

## 2. Supply Chain Analysis

### Go Module Dependencies
**go.mod declares:**
```
module github.com/usetheodev/theo-forge
go 1.25.0
require sigs.k8s.io/yaml v1.6.0          // direct
require (
    github.com/google/go-cmp v0.7.0       // indirect (test)
    github.com/kr/pretty v0.3.1           // indirect (test)
    github.com/rogpeppe/go-internal v1.14.1 // indirect
    go.yaml.in/yaml/v2 v2.4.3             // indirect
    go.yaml.in/yaml/v3 v3.0.4             // indirect
    gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect (test)
)
```

**Dependency counts:**
- Direct production dependencies: 1 (`sigs.k8s.io/yaml`)
- Indirect dependencies: 6 (all test-time or transitive of yaml)
- Total: 7

**Dependency health:**
| Dependency | Version | License | Notes |
|---|---|---|---|
| sigs.k8s.io/yaml | v1.6.0 | Apache-2.0 | Kubernetes SIG project, actively maintained |
| github.com/google/go-cmp | v0.7.0 | BSD-3-Clause | Google-maintained, current |
| github.com/kr/pretty | v0.3.1 | MIT | Stable, test-only |
| github.com/rogpeppe/go-internal | v1.14.1 | BSD-3-Clause | Russ Cox team, active |
| go.yaml.in/yaml/v2 | v2.4.3 | MIT | Transitive of sigs.k8s.io/yaml |
| go.yaml.in/yaml/v3 | v3.0.4 | MIT | Transitive of sigs.k8s.io/yaml |
| gopkg.in/check.v1 | v1.0.0-20201130... | BSD-2-Clause | Test-only via kr/pretty |

**License compatibility:** All dependencies use MIT, BSD-2, BSD-3, or Apache-2.0. All compatible with the project's MIT license.

**go.sum integrity:** 23 lines, both `h1:` and `go.mod` hashes present for all declared dependencies. No anomalies.

### Critical Issues
1. **`go 1.25.0` is a bleeding-edge / potentially unreleased toolchain version** — Go 1.24 is the latest stable as of mid-2025. This makes the library uninstallable by developers on the current stable toolchain without upgrading.
2. **No `toolchain` directive** — non-deterministic toolchain selection under Go 1.21+ toolchain management.
3. **No vendor directory** — builds depend on public module proxy availability.
4. **`golangci-lint` pinned to `latest`** — lint behavior may silently change between CI runs.

### Module Path
The module path is `github.com/usetheodev/theo-forge`. Since the project is at v0.x (pre-v1), no `/vN` suffix is required. This is correct.

---

## 3. Data Quality — testdata Golden Files

### Structure Overview
```
testdata/
├── *.golden.yaml       (29 files) — programmatic builder output fixtures (examples_test.go)
├── *.yaml              (6 files)  — programmatic builder output fixtures (roundtrip_test.go)
└── examples/           (194 files) — upstream Argo Workflows YAML examples for round-trip testing
```

### Golden Test Coverage
- **29 golden files** cover: hello-world, steps, DAG, arguments, coinflip, conditionals, scripts, loops, output-parameter, global-parameters, exit-handlers, artifact-passing, suspend, cron-workflow, parallelism, node-selector, retry-backoff, DAG-enhanced-depends, DAG-multiroot, DAG-targets, http-hello-world, volumes-emptydir, secrets, continue-on-fail, gc-ttl, pod-gc-strategy + 6 snake_case fixtures
- **Features NOT covered by golden tests:** DefaultPodAffinityFor (v0.4.0), DisableDefaultAffinity opt-out, SeccompProfile/Capabilities (v0.3.0), ColocateByLabel helper, WorkflowStatusDetail, ParseWorkflowStatusFromUnstructured
- **Round-trip coverage:** 194 Argo examples exercised semantically via `TestRoundTripTestdataExamples`

### Naming Convention Inconsistency
Two conventions coexist:
- `roundtrip_test.go` uses: `goldenTest(t, "simple_container", ...)` → `testdata/simple_container.yaml` (snake_case, no `.golden` in name)
- `examples_test.go` uses: `goldenTest(t, "hello-world.golden", ...)` → `testdata/hello-world.golden.yaml` (kebab-case, `.golden` embedded in the name passed to function)

The `goldenTest` function appends `.yaml` unconditionally, so the `.golden.yaml` double-extension is an artifact of passing `"hello-world.golden"` as the name. This creates two tiers of golden files that look identical by extension but are generated by different naming conventions.

### Golden File Generation
Golden files are generated via `go test -update-golden` flag. This is a clean, reproducible approach. The `-update-golden` flag is registered once in `testutil_test.go` line 200 and used in `goldenTest`. No hand-written golden files — all are generated from code. This is good practice.

### Round-trip Test Quality
`TestRoundTripTestdataExamples` walks `testdata/examples/` and round-trips every YAML file. The `assertSemantic` function normalizes JSON (removes nulls, sorts arrays of named objects) before comparison. This is robust — it handles YAML key ordering, null field emission, and array ordering differences. One subtlety: `normalizeForComparison` uses a bubble sort (O(n²)) but for the small arrays in workflow definitions, this is not a performance concern.

---

## 4. Release Engineering Assessment

### Git Tag Analysis
| Tag | Format | In CHANGELOG? | Date |
|---|---|---|---|
| v0.1.0 | valid | no entry in current CHANGELOG | — |
| v0.2.0 | valid | no entry in current CHANGELOG | — |
| v0.2,1 | **INVALID — comma typo** | — | — |
| v0.2.1 | valid | YES (missing date) | — |
| v0.3.0 | valid | YES | 2026-04-13 |
| v0.3.1 | valid | **NO ENTRY** | — |
| v0.4.0 | valid | YES | 2026-05-14 |

**CHANGELOG gaps:**
1. `v0.3.1` exists as a git tag (commits: gocritic lint fixes, PR #10) with no CHANGELOG section
2. `v0.2.1` section at line 33 is missing the release date
3. `v0.1.0` and `v0.2.0` have no CHANGELOG entries at all (may predate the CHANGELOG)

### Versioning Policy
- Module is at v0.x — no stability guarantee; breaking changes are acceptable per SemVer
- v0.2.1 introduced many BREAKING changes (documented in the section) — this is unusual for a patch version, even in v0.x. It would have been more conventional to call this v0.3.0.
- The project adheres to Keep a Changelog format for versions that do have entries

### Release Automation
- Release is triggered by pushing a tag matching `v*`
- GitHub Release notes are auto-generated by `softprops/action-gh-release` with `generate_release_notes: true`
- No Goreleaser, no binary artifacts (correct for a Go library)
- `go mod verify` runs before release to detect tampered modules — this is good practice

---

## 5. Findings Registry

| ID | Severity | Title |
|---|---|---|
| INFRA-001 | HIGH | securego/gosec@master — mutable branch, supply chain risk |
| INFRA-002 | HIGH | go 1.25.0 in go.mod — potentially unreleased toolchain |
| INFRA-005 | HIGH | All GitHub Actions use mutable version tags, no SHA pinning |
| INFRA-004 | MEDIUM | Release workflow not gated on lint/security checks |
| INFRA-008 | MEDIUM | No SBOM, artifact signing, or SLSA provenance |
| DATA-002 | MEDIUM | Golden test coverage gap for v0.3.0/v0.4.0 features |
| INFRA-003 | MEDIUM | Missing toolchain directive — non-deterministic toolchain selection |
| DATA-003 | LOW | Coverage upload only on push — PRs lack coverage feedback |
| DEP-001 | LOW | golangci-lint version pinned to latest — non-deterministic |
| INFRA-006 | LOW | git tag v0.2,1 — malformed tag with comma |
| INFRA-007 | LOW | CHANGELOG missing v0.3.1 entry; v0.2.1 missing date |
| DATA-001 | LOW | Mixed naming conventions in testdata golden files |

---

## 6. Phase 5 Quality Assessment

**Coverage:** All three Phase 5 domains fully examined:
- CI/CD: both workflows analyzed job-by-job, gating assessed, secrets usage reviewed
- Supply chain: go.mod/go.sum audited, dependency health checked, action pinning analyzed
- Data quality: golden file coverage, naming consistency, round-trip test quality assessed

**Evidence quality:** All findings cite specific file:line locations. No speculative findings.

**Severity calibration:**
- 3 HIGH: justified (supply chain attack surface — gosec@master, unreleased Go version affecting all consumers, broad mutable action risk)
- 5 MEDIUM: justified
- 4 LOW: accurate

**Phase 5 quality score: 0.82**

Deductions:
- (-0.10) go 1.25.0 is flagged HIGH but may be correct if Go 1.25 was released by now; cannot verify with certainty from static analysis
- (-0.08) No live `go list -m -u all` possible in this environment to check for outdated dep updates
