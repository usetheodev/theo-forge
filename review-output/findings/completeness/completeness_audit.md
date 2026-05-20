# Completeness Audit — theo-forge
**Phase:** 2 | **Date:** 2026-05-20 | **Auditor:** completeness-auditor (chief-reviewer)

## Summary

| Severity | Count |
|---|---|
| High | 2 |
| Medium | 5 |
| Low | 3 |
| **Total** | **10** |

No TODO/FIXME/HACK/XXX/STUB markers found in any production Go source file. `go vet ./...` passes clean. All tests pass (`go test ./...`).

---

## Findings

### F1: GlobalConfig scalar fields documented as applied defaults but never consumed by Build() [HIGH]

**ID:** completeness-globalconfig-dead-scalars
**File:** config/config.go:17-62 | helpers.go:119,221 | workflow.go:322

`config.GlobalConfig` is documented as holding "default values applied to all workflows and templates." `llms.txt` explicitly demonstrates:

```go
cfg := forge.GetGlobalConfig()
cfg.SetImage("python:3.11")    // documented as "default image for scripts"
cfg.SetNamespace("argo")       // documented as "default namespace"
```

However, `grep -rn 'globalConfig\.' --include="*.go"` (excluding tests) returns exactly TWO call sites:
- `helpers.go:119` — `globalConfig.DispatchTemplateHooks(&m)`
- `workflow.go:322` — `globalConfig.DispatchWorkflowHooks(&wfModel)`

`GlobalConfig.GetImage()`, `GetNamespace()`, `ServiceAccountName`, and `ImagePullPolicy` are **never called in any production Build() path**. Scalar fields are set but silently ignored.

**Impact:** Callers following llms.txt documentation silently achieve nothing. `cfg.SetImage()` has no effect unless the caller also writes a `RegisterTemplateHook` to apply it manually. This is the most prominently documented config feature and it is a complete no-op.

**Recommendation:** Either (a) apply scalar defaults during `buildTemplateModels()` (e.g., if container image is empty, use `globalConfig.GetImage()`), or (b) remove scalar fields entirely and update docs to direct users to hooks only. Option (b) is simpler and avoids implicit magic.

---

### F2: NewConfig() hook isolation gap [HIGH]

**ID:** completeness-h2-newconfig (stored as empty-id in DB, see finding with title "NewConfig() hook isolation gap")
**File:** helpers.go:119-221 | workflow.go:322

`NewConfig()` is documented as the way to create isolated `GlobalConfig` instances for dependency injection and concurrent builds. However:

```go
// helpers.go:221 — frozen at package init
var globalConfig = config.GetGlobal()

// helpers.go:119 — always dispatches to the singleton
globalConfig.DispatchTemplateHooks(&m)

// workflow.go:322
globalConfig.DispatchWorkflowHooks(&wfModel)
```

Any `NewConfig()` instance is **structurally incapable** of having its hooks dispatched by `Build()`. The package-level `globalConfig` pointer cannot be overridden by callers.

The tests in `workflow_test.go` (lines 914–988) test hook dispatch by calling `cfg.DispatchTemplateHooks(tpl)` **directly** — not via `Workflow.Build()`. This means the tests validate the hook dispatch mechanism in isolation but do not catch the integration gap.

**Impact:** Silent correctness failure. Callers using `NewConfig()` for hook isolation cannot achieve it. Concurrent builds using isolated configs to prevent hook cross-contamination also fail silently.

**Recommendation:** Add `Workflow.BuildWithConfig(cfg *config.GlobalConfig)` that substitutes the provided config for `globalConfig`, or add a `Config *config.GlobalConfig` field to `Workflow` that Build() uses when non-nil.

---

### F3: WorkflowPreBuildHook not dispatched for WorkflowTemplate/CronWorkflow despite CHANGELOG claim [MEDIUM]

**ID:** completeness-workflow-hook-not-fired-for-wftemplate
**File:** workflow_template.go:44-98

CHANGELOG v0.3.0 states: "GlobalConfig hooks are now wired into the Build() pipeline for Workflow, WorkflowTemplate, ClusterWorkflowTemplate, and CronWorkflow."

Template hooks (`func(*model.TemplateModel)`) ARE fired for all types via `buildTemplateModels()`.

However, `WorkflowPreBuildHook` (`func(*model.WorkflowModel)`) is dispatched **only in `Workflow.Build()`** (workflow.go:322). `WorkflowTemplate.Build()` and `CronWorkflow.Build()` do not call `globalConfig.DispatchWorkflowHooks()`.

Since `WorkflowPreBuildHook` is typed `func(*model.WorkflowModel)`, it cannot directly apply to `WorkflowTemplateModel` or `CronWorkflowModel` — but workflow hooks registered by operators expecting to inject labels on all CRD types will silently not fire.

**Recommendation:** Either add `DispatchWorkflowHooks` calls in `WorkflowTemplate.Build()` and `CronWorkflow.Build()` using the shared spec fields, or document clearly that workflow hooks apply only to `Workflow` type.

---

### F4: README ListWorkflows example passes nil to a string parameter — will not compile [MEDIUM]

**ID:** completeness-readme-listworkflows-nil
**File:** README.md:188 | client/client.go:176

```go
// README shows:
workflows, err := svc.ListWorkflows(ctx, nil)

// Actual signature (client/client.go:176):
func (s *WorkflowsService) ListWorkflows(ctx context.Context, namespace string) ([]model.WorkflowModel, error)
```

`nil` cannot be assigned to a `string` in Go. This code will not compile.

**Recommendation:** Update README.md to `svc.ListWorkflows(ctx, "")` or `svc.ListWorkflows(ctx, "my-namespace")`.

---

### F5: README expr builder example calls .Eq() with argument but Eq() takes no arguments [MEDIUM]

**ID:** completeness-readme-expr-eq-mismatch
**File:** README.md:205 | expr/expr.go:52,88

```go
// README shows:
cond := expr.Steps("validate").Attr("outputs.result").Eq(expr.C("success"))

// Actual Eq() (expr/expr.go:52):
func (e Expr) Eq() string  // zero arguments, returns {{=...}} wrapper string

// Correct equality method (expr/expr.go:88):
func (e Expr) Equals(other Expr) Expr
```

This will not compile. `Eq()` takes no arguments and returns a string (not `Expr`), so chaining `.Eq(expr.C("success"))` is a type error.

**Recommendation:** Update README to `.Equals(expr.C("success"))`. Add a clarifying doc comment to `Eq()` explaining it renders `{{=expr}}` syntax.

---

### F6: Task.Then() operator precedence — no parentheses on existing Depends [MEDIUM]

**ID:** completeness-h4-then-precedence
**File:** dag.go:85-92

```go
func (t *Task) Then(other *Task) *Task {
    if other.Depends == "" {
        other.Depends = t.Name
    } else {
        other.Depends = other.Depends + " " + string(OperatorAnd) + " " + t.Name
        // Missing: parenthesize existing expression first
    }
    return other
}
```

For pure `Then()` chains this is correct. But if `other.Depends` already contains a complex expression (e.g., set via `OnSuccess()`, `OnFailure()`, or manual assignment containing `||`), the `&&` appended by `Then()` has higher precedence than `||`, producing silent incorrect dependency logic.

Note: `Or()` does produce parenthesized strings `(A || B)`, so `Or()`-then-`Then()` chains are safe. The risk is in manually constructed `Depends` strings mixed with `Then()`.

**Recommendation:** Parenthesize existing `Depends` before appending: `other.Depends = "(" + other.Depends + ") && " + t.Name`. Add a test with mixed AND/OR graph.

---

### F7: Five sub-packages have zero test files [MEDIUM]

**ID:** completeness-missing-package-tests
**File:** expr/, model/, serialize/, validate/, config/

```
go test ./... output:
? github.com/usetheodev/theo-forge/config    [no test files]
? github.com/usetheodev/theo-forge/expr      [no test files]
? github.com/usetheodev/theo-forge/model     [no test files]
? github.com/usetheodev/theo-forge/serialize [no test files]
? github.com/usetheodev/theo-forge/validate  [no test files]
```

These packages contain critical logic: expression DSL building, resource unit validation, YAML/JSON serialization, config hook dispatch. Zero isolated test coverage exists.

The root package tests incidentally exercise some of this via integration, but no package-level unit tests exist.

**Recommendation:** Add unit test files per package. Priority order: `validate` (business logic with numeric comparisons), `expr` (DSL formatting), `serialize` (round-trip correctness for WorkflowTemplate/CronWorkflow), `config` (hook dispatch integration via Build(), not just direct dispatch calls).

---

### F8: validate.ResourceRequirements never called during Build() [LOW]

**ID:** completeness-h7-resource-validation-not-in-build
**File:** container.go:91-153

`validate.ResourceRequirements(r)` validates CPU/memory strings and request-vs-limit comparisons. It is exposed as `forge.ValidateResourceRequirements()` but is **never called by `Container.BuildTemplate()`, `Script.BuildTemplate()`, or `Workflow.Build()`**.

Invalid resource strings (e.g., `CPU: "bad-value"`) pass silently through `Build()` and will fail only at Argo controller admission time after `kubectl apply`.

**Recommendation:** Call `validate.ResourceRequirements(*c.Resources)` inside `Container.BuildTemplate()` and `Script.BuildTemplate()` when `c.Resources != nil`. The error should propagate to the caller.

---

### F9: WorkflowsService.VerifySSL is a no-op [LOW]

**ID:** completeness-h1-verifyssl
**File:** client/client.go:35-49

`VerifySSL bool` is set to `true` by default and documented as controlling TLS verification. The `http.Client` is constructed with no custom `Transport`, so the field is never read and has zero effect.

`GlobalConfig.VerifySSL` (config/config.go:33-34) has the same issue — stored but never consumed anywhere in any Build() path or client construction.

Security outcome is accidentally correct (TLS always verified), but the API contract is broken.

**Recommendation:** Either implement `http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !s.VerifySSL}}`, or remove the field and document that TLS is always verified.

---

### F10: serialize package asymmetric — WorkflowTemplate missing JSON and file functions [LOW]

**ID:** completeness-h6-clusterwftemplate-serialize
**File:** serialize/serialize.go:66-108

The serialize package has asymmetric coverage:
- `Workflow`: YAML, JSON, Dict, File (all four)
- `CronWorkflow`: YAML, JSON, no File, no FromJSON
- `WorkflowTemplate`: YAML, FromYAML — no JSON, no FromJSON, no File

`ClusterWorkflowTemplate` has no dedicated functions (must use `WorkflowTemplateToYAML` with correct `Kind`).

**Recommendation:** Add `WorkflowTemplateToJSON`, `WorkflowTemplateFromJSON`, `WorkflowTemplateFromFile`, `CronWorkflowFromFile`. Low priority but breaks API symmetry expectations.

---

## Promises vs. Implementation Traceability

| README/llms.txt Claim | Status |
|---|---|
| `cfg.SetImage("python:3.11")` — default image for scripts | NOT IMPLEMENTED — field stored, never applied |
| `cfg.SetNamespace("argo")` — default namespace | NOT IMPLEMENTED — field stored, never applied |
| `NewConfig()` for isolated configuration with hooks | NOT IMPLEMENTED — hooks on NewConfig() never fired |
| `svc.ListWorkflows(ctx, nil)` — REST client example | WRONG — nil not assignable to string |
| `.Eq(expr.C("success"))` — expression builder example | WRONG — Eq() takes no arguments |
| Hooks wired for WorkflowTemplate/CronWorkflow (CHANGELOG) | PARTIAL — template hooks yes, workflow hooks no |
| "Catch mistakes before kubectl apply" (README value prop) | PARTIAL — ResourceRequirements validation not integrated |
| All serialize functions for all CRD types | PARTIAL — WorkflowTemplate missing JSON/File functions |

---

## Test Coverage Gaps

| Package | Test Files | Coverage |
|---|---|---|
| github.com/usetheodev/theo-forge | 16 test files | High |
| client | 1 test file | Moderate |
| config | 0 test files | None |
| expr | 0 test files | None |
| model | 0 test files | None |
| serialize | 0 test files | None |
| validate | 0 test files | None |
