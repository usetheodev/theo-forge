# Phase 4 — Deep Code Review
**Date:** 2026-05-20
**Reviewer:** chief-reviewer (code-reviewer agent)
**Files reviewed:** ~50 Go files across root package, client/, config/, expr/, model/, serialize/, validate/

---

## Static Analysis Results

| Tool | Result |
|------|--------|
| `go vet ./...` | CLEAN — zero issues |
| `go test ./... -count=1 -race` | ALL PASS — 2 packages tested, no data races |
| `gofmt -l .` | 11 files fail — 7 production files, 4 test files |
| `golangci-lint` | Not installed in environment |

**gofmt-failing production files:**
- workflow_template.go
- model/artifact.go, model/cron.go, model/dag.go, model/steps.go, model/template.go, model/workflow.go

---

## Phase 4 Findings Summary

**20 new findings registered (phase=4, category=code)**

| ID | Severity | Title |
|----|----------|-------|
| code-p4-verifyssl-dead | HIGH | VerifySSL field never wired into HTTP transport |
| code-p4-readme-broken-examples | HIGH | README quickstart contains two non-compiling examples |
| code-p4-interface-unsafe-fields | HIGH | Five model fields use interface{} losing type safety |
| code-p4-gofmt-violations | MEDIUM | gofmt violations in 7 production files |
| code-p4-container-script-dry | MEDIUM | Container and Script duplicate ~20 fields, drift visible |
| code-p4-configmapvolume-name-bug | MEDIUM | ConfigMapVolume cannot use a different name than volume |
| code-p4-template-ref-mutual-exclusion | MEDIUM | Task/Step allow Template+TemplateRef+Inline simultaneously |
| code-p4-missing-entrypoint-validation | MEDIUM | Workflow.validate() missing entrypoint check |
| code-p4-missing-image-validation | MEDIUM | Container/Script BuildTemplate() missing Image check |
| code-p4-http-timeout-field-mismatch | MEDIUM | HTTPTemplate.Timeout never populates HTTPModel.TimeoutSeconds |
| code-p4-client-empty-name-url | MEDIUM | Client methods accept empty name causing malformed URLs |
| code-p4-namelen-inconsistent | MEDIUM | NameLimit check missing in ClusterWorkflowTemplate and CronWorkflow |
| code-p4-globalconfig-exported-fields-race | MEDIUM | Exported GlobalConfig fields enable mutex bypass |
| code-p4-containerset-inline-io-build | MEDIUM | ContainerSet duplicates input/output building, skipping shared helpers |
| code-p4-dead-len-check | LOW | Dead len()==0 check after guaranteed non-empty loop |
| code-p4-err2-naming | LOW | err2 variable breaks codebase naming convention |
| code-p4-expr-G-broken | LOW | expr.G produces leading-dot malformed expressions |
| code-p4-doc-stepsexpr-nonexistent | LOW | doc.go references non-existent [StepsExpr] symbol |
| code-p4-backoff-factor-partial-normalization | LOW | Backoff.Factor normalization silently ignores float64 |
| code-p4-getinfo-getversion-untyped | LOW | GetInfo/GetVersion return untyped map[string]interface{} |

---

## Top 7 Most Impactful Issues

### 1. VerifySSL Dead Code — client/client.go:35-49
`WorkflowsService.VerifySSL` and `GlobalConfig.VerifySSL` are documented as controlling TLS verification. `NewWorkflowsService()` injects `&http.Client{Timeout: 30s}` with no custom `tls.Config`. No code ever reads `VerifySSL`. Users setting `VerifySSL=false` receive zero effect. The API contract is broken.

**Fix:** Wire into `http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !s.VerifySSL}}`.

### 2. README Non-Compiling Examples — README.md:187,205
- Line 187: `svc.ListWorkflows(ctx, nil)` — `ListWorkflows` takes `string`, not `nil`
- Line 205: `.Eq(expr.C("success"))` — `Expr.Eq()` is zero-argument, returning string. The correct method is `Equals(other Expr)`

Both will fail `go build`. These are first-contact examples for new users.

**Fix:** `svc.ListWorkflows(ctx, "")` and `.Equals(expr.C("success"))`.

### 3. interface{} Type-Unsafe Fields — model/workflow.go:118, model/template.go:176,182, model/retry.go:16,23
Five fields accept unknown runtime types with no compile-time validation:
- `PodDisruptionBudget.MinAvailable`, `MaxUnavailable` (int or string)
- `HTTPGetAction.Port`, `TCPSocketAction.Port` (int or named port string)
- `Backoff.Factor`, `RetryStrategyModel.Limit` (int, string, or float)

**Fix:** Define `IntOrString` type with custom JSON marshal, replacing all six fields.

### 4. Container/Script Field Duplication — container.go:10-280
~20 fields duplicated between `Container` and `Script` with no shared base type. Script already shows drift: missing `Parallelism`, `ArchiveLocation`, `InitContainers`, `Ports`, `SecurityContext`, `EnvFrom`, `ReadinessProbe`, `LivenessProbe` with no documentation.

**Fix:** Extract `BaseTemplate` embedded struct.

### 5. Missing Template/TemplateRef Mutual Exclusion — dag.go:118, steps.go:55
`Task.BuildDAGTask()` and `Step.BuildStep()` pass all three of `Template`, `TemplateRef`, `Inline` to the model with no validation. Argo requires exactly one. Setting multiple produces invalid YAML with a runtime API error.

**Fix:** Count non-zero fields; return error if count > 1.

### 6. Missing Image Validation — container.go:91,226
`Container.BuildTemplate()` and `Script.BuildTemplate()` do not validate empty `Image`. `ContainerModel.Image` has no `omitempty`, so `image: ""` is serialized and rejected at Argo API runtime with an opaque error.

**Fix:** Add `if c.Image == "" { return error }` at build time.

### 7. ConfigMapVolume Name Bug — volume.go:88-106
`ConfigMapVolume` has no `ConfigMapName` field; `BuildVolume()` always sets `ConfigMapVolumeModel.Name = v.Name`. This forces the ConfigMap object name to equal the volume mount name. `SecretVolume` correctly adds a separate `SecretName` field; `ConfigMapVolume` does not.

**Fix:** Add `ConfigMapName string` to `ConfigMapVolume` struct with fallback to `v.Name`.

---

## Concurrency Analysis

No data races detected by `go test -race`. The `config.GlobalConfig` uses `sync.RWMutex` consistently across all `Set*/Get*` methods and hook dispatch. However, exported struct fields (`Host`, `Token`, `Image`, etc.) allow callers to bypass the mutex entirely via direct field access. Under concurrent access, direct reads of `cfg.Image` while another goroutine calls `cfg.SetImage()` would be a data race. **Risk: MEDIUM** (requires concurrent use of the global singleton, which is the common case).

---

## Quality Score: 0.72 (PASS — threshold 0.70)
