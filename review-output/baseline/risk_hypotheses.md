# Risk Hypotheses — theo-forge

Generated: Phase 1, Iteration 1 (2026-05-20)

These are hypotheses formed during baseline mapping, ranked by likelihood and impact.
Each will be validated or dismissed in subsequent phases.

---

## H1: VerifySSL Field is Dead Code (HIGH likelihood, MEDIUM impact)

**Hypothesis**: `client.WorkflowsService.VerifySSL` is a struct field that is documented and set to `true` by default in `NewWorkflowsService()`, but it is NEVER applied to the underlying `http.Client`. The HTTP client is created as `&http.Client{Timeout: 30 * time.Second}` with no TLS configuration.

**Evidence**: `client/client.go` lines 42-50 — `NewWorkflowsService` creates an `http.Client` without a custom `Transport`. `VerifySSL` is set but never read.

**Impact**: Users who set `svc.VerifySSL = false` believing it disables TLS verification get no effect — the client always uses the default TLS verification. This is actually the SAFER behavior, but the API is deceptive and the field misleads users into thinking they can disable verification when they cannot.

**Severity estimate**: LOW (security is accidentally correct, but API contract is broken)

---

## H2: NewConfig() Hook Isolation Gap (HIGH likelihood, MEDIUM impact)

**Hypothesis**: `forge.NewConfig()` creates an isolated `GlobalConfig` instance, but hooks registered on it are NEVER applied during `Workflow.Build()`. The root package's `helpers.go:buildTemplateModels()` calls `globalConfig.DispatchTemplateHooks()` where `globalConfig` is the package-level variable pointing to `config.GetGlobal()`, not any user-created isolated instance.

**Evidence**: `helpers.go` line 119 (`globalConfig.DispatchTemplateHooks(&m)`) and line 323 in `workflow.go` (`globalConfig.DispatchWorkflowHooks(&wfModel)`). The `globalConfig` var is assigned at package init from `config.GetGlobal()`.

**Impact**: Consumers who follow the documented "use NewConfig() for isolated configuration" pattern and register hooks on isolated configs will find those hooks silently not applied. This is a correctness bug for the hook use case.

**Severity estimate**: MEDIUM (silent functional failure)

---

## H3: Expr String Injection via Single-Quote Strings (MEDIUM likelihood, LOW-MEDIUM impact)

**Hypothesis**: The `expr` package builds Argo expression strings by embedding user-provided strings inside single quotes without escaping. Methods like `Contains(s string)`, `Matches(pattern string)`, `StartsWith(prefix string)`, `EndsWith(suffix string)`, and all `Sprig.*` functions format strings as `'<user-value>'`. A string containing a single quote would break the expression syntax.

**Evidence**: `expr/expr.go` lines 161-174 (Contains, Matches, StartsWith, EndsWith) and lines 221-235 (Sprig.*). No escaping present.

**Impact**: For a pure YAML builder SDK with no server execution, this cannot cause server-side injection. However, it can produce malformed Argo expression syntax that fails at workflow runtime. The risk is template generation correctness, not security injection.

**Severity estimate**: LOW (no server, but produces broken YAML on special chars)

---

## H4: Task.Then() Operator Precedence with Complex Graphs (MEDIUM likelihood, MEDIUM impact)

**Hypothesis**: `Task.Then(other)` appends ` && other.Name` to `other.Depends` without parenthesizing. For simple linear chains this is correct, but for mixed AND/OR dependency graphs assembled with `Then()` and manual `Depends` strings, operator precedence may not produce the intended graph.

**Evidence**: `dag.go` lines 85-92. `Then()` does: `other.Depends = other.Depends + " " + "&&" + " " + t.Name`. No parentheses. The `Or()` method returns a formatted string `"(A || B)"` but it must be manually assigned to `Depends` — it is not integrated with `Then()`.

**Impact**: Silent incorrect DAG structure (wrong execution order or wrong skip conditions) in complex workflows.

**Severity estimate**: MEDIUM (correctness issue, no runtime error, hard to detect)

---

## H5: GlobalConfig Singleton Race Condition in Concurrent Tests (MEDIUM likelihood, LOW impact)

**Hypothesis**: Tests that use the global singleton config (via `GetGlobalConfig()`) and run in parallel may observe shared state mutations from other tests. The `config.GlobalConfig` uses `sync.RWMutex` for field access, but hooks registered by one test persist until `Reset()` or `ClearHooks()` is called.

**Evidence**: `config/config.go` — the `globalConfig` singleton with its `templateHooks` and `workflowHooks` slices. The test `TestNewConfigIsIndependent` correctly uses `NewConfig()` but does NOT interact with `globalConfig`. Tests that mutate the global config and fail before `defer cfg.Reset()` will leak state.

**Impact**: Flaky tests (not production issue). Not a security concern.

**Severity estimate**: LOW (test reliability issue)

---

## H6: Missing ClusterWorkflowTemplate Serialize Functions (LOW likelihood, LOW impact)

**Hypothesis**: The `serialize` package provides `WorkflowTemplateToYAML`, `WorkflowTemplateFromYAML`, `CronWorkflowToYAML`, `CronWorkflowFromYAML` but does NOT provide equivalent functions for `ClusterWorkflowTemplate`. Users must use `WorkflowTemplateToYAML` with a `model.WorkflowTemplateModel` constructed with `Kind: "ClusterWorkflowTemplate"` — this works but is not discoverable.

**Evidence**: `serialize/serialize.go` — no ClusterWorkflowTemplate functions. `workflow_template.go` in root package likely handles this at the Build() level.

**Impact**: Minor API completeness gap, not a defect.

**Severity estimate**: LOW (completeness gap)

---

## H7: validate.ResourceRequirements Does Not Call Build() (MEDIUM likelihood, LOW impact)

**Hypothesis**: `validate.ResourceRequirements(r model.ResourceRequirements)` validates string values but is not automatically called by `Container.BuildTemplate()`, `Workflow.Build()`, or any other Build method. Users must call it explicitly. Invalid resource strings in a workflow will only fail at `sigs.k8s.io/yaml` marshal time (unlikely to fail) or at Argo controller admission time (post-submit).

**Evidence**: Searching for `ValidateResourceRequirements` calls in container.go or other builder files — expected to find none.

**Impact**: Late detection of invalid resource strings (at Argo admission, not at SDK build time). Weakens the "catch mistakes before kubectl apply" value proposition.

**Severity estimate**: LOW-MEDIUM (usability/correctness gap in validation coverage)

---

## H8: PodDisruptionBudget Uses interface{} (LOW likelihood, LOW impact)

**Hypothesis**: `model.PodDisruptionBudget.MinAvailable` and `MaxUnavailable` are typed as `interface{}`. This mirrors K8s IntOrString pattern but loses type safety at the Go level — any value (including maps, slices) could be set.

**Evidence**: `model/workflow.go` lines 116-119.

**Impact**: Users can set nonsensical values that produce invalid YAML without compile-time or build-time errors.

**Severity estimate**: LOW (niche type, unlikely to affect most users)

---

## Summary Table

| ID | Hypothesis | Severity | Phase to validate |
|---|---|---|---|
| H1 | VerifySSL dead field | LOW | Phase 4 (code review) |
| H2 | NewConfig() hook isolation gap | MEDIUM | Phase 4 (code review) |
| H3 | Expr string injection | LOW | Phase 4 (code review) |
| H4 | Task.Then() operator precedence | MEDIUM | Phase 4 (code review) |
| H5 | GlobalConfig race in tests | LOW | Phase 4 (code review) |
| H6 | Missing ClusterWorkflowTemplate serialize | LOW | Phase 2 (completeness) |
| H7 | ResourceRequirements not validated in Build() | LOW-MEDIUM | Phase 2 (completeness) |
| H8 | PodDisruptionBudget interface{} | LOW | Phase 3 (architecture) |
