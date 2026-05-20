# Component Inventory — theo-forge

Generated: Phase 1, Iteration 1 (2026-05-20)

## Components (7 total)

### 1. root-builder (forge root package)

| Field | Value |
|---|---|
| Path | `/` (package `forge`) |
| Type | library |
| Files | workflow.go, dag.go, steps.go, container.go, templates.go, types.go, affinity.go, artifact.go, env.go, volume.go, parameter.go, status.go, helpers.go, workflow_template.go, helpers_test.go + *_test.go |
| Key types | Workflow, WorkflowTemplate, ClusterWorkflowTemplate, CronWorkflow, DAG, Steps, Container, Script, ResourceTemplate, HTTPTemplate, Suspend, ContainerSet, Task, Step, Parameter, Artifact types, Volume types, Env types |
| Interfaces | Templatable, VolumeBuilder, ArtifactBuilder, EnvBuilder |
| Complexity | HIGH — largest package, most types, all Build() methods |

**Critical paths:**
- `Workflow.validate()` — name/generateName length and presence checks
- `Workflow.Build()` — orchestrates all sub-builders and hook dispatch
- `DAG.AddTask()` — nodeNames map duplicate check
- `Task.Then()` — Depends string mutation (no parentheses added, raw concatenation)
- `resolveAffinity()` — DefaultPodAffinity injection logic

### 2. pkg-model (forge/model)

| Field | Value |
|---|---|
| Path | `model/` |
| Type | module |
| Files | workflow.go, template.go, dag.go, steps.go, affinity.go, artifact.go, env.go, volume.go, parameter.go, retry.go, status.go, metrics.go, cron.go, types.go, doc.go |
| Key types | WorkflowModel, WorkflowSpec, TemplateModel, DAGModel, DAGTaskModel, StepModel, ContainerModel, ScriptModel, ContainerSetModel, all *Model types |
| Interfaces | None — pure data structs |
| Complexity | MEDIUM — many types, zero logic |

**Notes:**
- All fields use `json:"...,omitempty"` + `yaml:"...,omitempty"` tags — critical for clean YAML output
- `PodDisruptionBudget.MinAvailable/MaxUnavailable` use `interface{}` — can accept int or string (intOrString pattern)
- `SynchronizationModel` has both singular (legacy) and plural (v3.6+) mutex/semaphore fields — potential serialization ambiguity

### 3. pkg-client (forge/client)

| Field | Value |
|---|---|
| Path | `client/` |
| Type | library |
| Files | client.go, client_test.go, doc.go |
| Key types | WorkflowsService, APIError, WorkflowCreateRequest, ListWorkflowsResponse |
| Interfaces | HTTPClient (injectable), Buildable (accepts forge types) |
| Complexity | LOW-MEDIUM |

**Operations:**
- CreateWorkflow / CreateWorkflowFromModel
- GetWorkflow, ListWorkflows, DeleteWorkflow
- LintWorkflow / LintWorkflowFromModel
- StopWorkflow, TerminateWorkflow, SuspendWorkflow, ResumeWorkflow
- GetInfo, GetVersion

**Notes:**
- `VerifySSL` field exists on struct but is never used — `NewWorkflowsService` does NOT configure TLS verification
- Default HTTP timeout: 30 seconds
- Token auto-prefixed with "Bearer " if not already prefixed

### 4. pkg-serialize (forge/serialize)

| Field | Value |
|---|---|
| Path | `serialize/` |
| Type | module |
| Files | serialize.go |
| Key functions | WorkflowToYAML, WorkflowFromYAML, WorkflowToJSON, WorkflowFromJSON, WorkflowToDict, WorkflowToFile, WorkflowFromFile, WorkflowTemplateToYAML, CronWorkflowToYAML, etc. |
| Complexity | LOW |

**Notes:**
- `WorkflowToDict` uses JSON marshal then unmarshal into `map[string]interface{}` — double allocation, but functionally correct
- `WorkflowFromFile` uses `filepath.Clean` for path sanitization and has `#nosec G304` comment
- File creation: `os.MkdirAll(absDir, 0o750)` + `os.WriteFile(path, data, 0o600)` — reasonable permissions
- Missing: `ClusterWorkflowTemplateToYAML` / `ClusterWorkflowTemplateFromYAML` functions

### 5. pkg-validate (forge/validate)

| Field | Value |
|---|---|
| Path | `validate/` |
| Type | module |
| Files | units.go |
| Key functions | BinaryUnit, DecimalUnit, ConvertBinaryUnit, ConvertDecimalUnit, ResourceRequirements |
| Complexity | LOW |

**Notes:**
- Regex patterns compiled at package init (package-level vars) — correct approach
- `binaryUnitRe` matches values without suffix (plain bytes) via `(Ki|Mi|Gi|Ti|Pi|Ei)?$` — optional suffix is correct per K8s spec
- CPU "plain" values (e.g., `"2"`) pass DecimalUnit validation — correct
- `ResourceRequirements` only validates CPU, Memory, EphemeralStorage — does not cover custom resources

### 6. pkg-expr (forge/expr)

| Field | Value |
|---|---|
| Path | `expr/` |
| Type | module |
| Files | expr.go, doc.go |
| Key types | Expr (value type) |
| Key functions | E(), C(), Tasks(), Steps(), Inputs(), Outputs(), Item(), Workflow(), InputParam(), TaskOutputParam(), etc. |
| Complexity | LOW |

**Notes:**
- `Expr` is a value type (not pointer) — immutable by design, all methods return new Expr
- `G = Expr{repr: ""}` — global expression root with empty string. `G.Attr("x")` produces `.x` (leading dot) — potential bug if misused
- String constants (single-quoted) in `C()` for strings — correct for Argo expression syntax
- `OrExpr()` naming is inconsistent (not just `Or()`) — likely collision avoidance with `expr.Or` reserved word
- No escaping of user-provided strings in `Contains()`, `Matches()`, `StartsWith()`, `EndsWith()`, `Sprig.*` — injection risk if string values contain single quotes

### 7. pkg-config (forge/config)

| Field | Value |
|---|---|
| Path | `config/` |
| Type | config |
| Files | config.go |
| Key types | GlobalConfig, PreBuildHook, WorkflowPreBuildHook |
| Key functions | GetGlobal(), New(), SetImage(), SetNamespace(), RegisterTemplateHook(), DispatchTemplateHooks(), etc. |
| Complexity | LOW |

**Notes:**
- Package-level `globalConfig` singleton initialized with `python:3.11` as default image — this is a non-obvious default that could confuse Go users
- `Reset()` is on the instance, not package-level — callers must hold a reference
- Root package exposes `globalConfig = config.GetGlobal()` as a package-level var — both packages reference the same singleton pointer. This is correct but the coupling is tight.

## File Count Summary

| Package | Go Files | Test Files |
|---|---|---|
| forge (root) | ~18 | ~12 |
| model | 14 | 0 |
| client | 2 | 1 |
| serialize | 1 | 0 |
| validate | 1 | 0 |
| expr | 2 | 0 |
| config | 1 | 0 |
| **Total** | **~39** | **~13** |

## External Dependencies

| Dependency | Version | Usage |
|---|---|---|
| `sigs.k8s.io/yaml` | v1.6.0 | YAML marshal/unmarshal in serialize package |
| `github.com/google/go-cmp` | v0.7.0 | Indirect (via yaml dep) |
| `go.yaml.in/yaml/v2`, `v3` | various | Indirect (via sigs.k8s.io/yaml) |

**Notable**: No k8s.io/client-go, no k8s.io/api — this is intentional and keeps the dependency tree very lean.
