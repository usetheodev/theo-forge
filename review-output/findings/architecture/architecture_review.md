# Architecture Review — Phase 3
**Date:** 2026-05-20
**Reviewer:** Architecture Analyst (Chief Reviewer acting)
**Codebase:** theo-forge v0.4.0

## Executive Summary

The dependency graph is clean with no cycles and clear layering (model -> adapters -> root builder). The codebase is architecturally sound for its scope — a pure Go SDK with minimal external dependencies. However, three structural issues are significant enough to affect SDK usability and correctness: the globalConfig singleton freeze at init, the expr.Eq()/Equals() naming collision (already causing README errors), and the type-unsafe PodDisruptionBudget using interface{}. The Container/Script field duplication is a DRY violation at the architecture level that will create maintenance drift over time.

## Findings by Severity

### HIGH

| ID | Title | File | Line |
|---|---|---|---|
| arch-globalconfig-frozen-singleton | forge.globalConfig captured at package init prevents runtime config injection | helpers.go | 221 |

### MEDIUM

| ID | Title | File | Line |
|---|---|---|---|
| arch-helpers-srp-violation | helpers.go mixes 4 distinct concerns (build helpers, I/O, validation proxies, config wiring) | helpers.go | 1-234 |
| arch-client-buildable-interface-duplication | client.Buildable cannot accept WorkflowTemplate or CronWorkflow | client/client.go | 22-25 |
| arch-pdb-interface-type-unsafe | PodDisruptionBudget uses interface{} bypassing compile-time type safety | model/workflow.go | 117-120 |
| arch-container-script-field-duplication | Container and Script share ~20 identical fields with no shared base type | container.go | 10-280 |
| arch-expr-naming-collision | expr.Eq() returns rendered string; Equals() returns comparator Expr — same prefix, opposite semantics | expr/expr.go | 51-54 |

### LOW

| ID | Title | File | Line |
|---|---|---|---|
| arch-workflow-template-incomplete-api-surface | WorkflowTemplate exposes ~10 of Workflow's 32+ fields with no documented rationale | workflow_template.go | 11-32 |
| arch-hook-dispatch-not-composable | Hook system: no filtering by template type, no named removal, no error propagation | config/config.go | 110-131 |

## Dependency Graph Health

**Cycles:** NONE
**Layering violations:** NONE

Layer 0 (Foundation): `model` (no imports), `expr` (stdlib only)
Layer 1 (Adapters): `config` (→model), `serialize` (→model, →yaml), `validate` (→model), `client` (→model)
Layer 2 (Builder DSL): `forge` root (→model, →serialize, →config, →validate)

The layering is clean and consistent. No sub-package imports the root forge package (which would create cycles). `client` correctly avoids importing `forge` root by defining its own `Buildable` interface locally.

## Separation of Concerns Assessment

### Does the root builder mix domain types with serialization?
Partially YES. `workflow.go` and `workflow_template.go` import `serialize` directly to implement `ToYAML()`/`ToJSON()` methods on the builder structs. This is a minor concern — the builder owns the output format methods — but it means serialization knowledge is spread across 3 files (workflow.go, workflow_template.go, helpers.go). All serialization logic is delegated to the `serialize` package, so there is no actual mixing of serialization implementation.

### Is the client decoupled from the model?
YES, but with a gap. `client` imports only `model`, not `forge` root. The `Buildable` interface is defined locally using `model.WorkflowModel` as the return type. This is clean DIP. The gap is that `client.Buildable` only handles `*forge.Workflow`, not `*forge.WorkflowTemplate` or `*forge.CronWorkflow`.

### Does helpers.go violate SRP?
YES. See finding `arch-helpers-srp-violation`. Four distinct responsibilities in one file. Low urgency (internal functions, not public complexity), but should be refactored.

### Is expr a pure DSL?
YES. `expr` imports only `fmt` and `strings`. No dependency on model state, no runtime context, no config access. Architecturally ideal.

## Builder Pattern Consistency

| Builder | Pattern | Consistent? |
|---|---|---|
| Workflow | Struct with public fields, Build() method | Yes |
| WorkflowTemplate | Struct with public fields, Build() method | Yes |
| ClusterWorkflowTemplate | Struct with public fields, Build() method | Yes |
| CronWorkflow | Struct with public fields, Build() method | Yes |
| Container | Struct with public fields, BuildTemplate() method | Yes |
| Script | Struct with public fields, BuildTemplate() method | Yes |
| DAG | Struct with public fields + AddTask(), BuildTemplate() | Yes |
| Steps | Struct with public fields + AddSequentialStep()/AddParallelGroup(), BuildTemplate() | Yes |
| Task | Struct with public fields + Then()/OnSuccess()/OnFailure(), BuildDAGTask() | Yes |

**Assessment:** Builder pattern is highly consistent across all types. All builders use plain struct initialization (no functional options, no constructors) with a single Build/BuildTemplate method. This is a deliberate KISS decision that makes the API easy to learn. No inconsistency found.

**Functional options vs setters:** Neither is used — the SDK uses plain struct field assignment. This is the simplest possible approach and works well for a pure data-building context. The absence of functional options is a YAGNI-correct decision for this scope.

## OCP/DIP/ISP Assessment

### OCP
The `Templatable` interface (types.go line 12-15) is the core extension point. Adding a new template type requires implementing `BuildTemplate() (model.TemplateModel, error)` and `GetName() string`. No existing code needs to change. OCP is correctly applied here.

### DIP
- Root package depends on model interfaces via the Templatable interface (correct)
- Root package calls config.GetGlobal() through a package-level variable (violation for NewConfig use case)
- Client package defines Buildable locally (correct workaround for import cycle avoidance)
- Serialize package is injected as a dependency via direct import (acceptable for a library with no runtime variation)

### ISP
`Templatable` interface is minimal (2 methods). `HTTPClient` interface in client is minimal (1 method). Both respect ISP. `client.Buildable` is minimal (2 methods). No fat interfaces found.

## God Files / Large Files

| File | LOC | Concern |
|---|---|---|
| helpers_test.go | 1891 | Large test file — not a production concern |
| examples_test.go | 1770 | Large examples — not a production concern |
| workflow_test.go | 1216 | Large test file — not a production concern |
| workflow.go | 371 | Acceptable — Workflow struct has 32+ fields |
| workflow_template.go | 330 | Acceptable — contains Workflow, ClusterWorkflow, CronWorkflow builders |
| expr/expr.go | 317 | Acceptable — all methods are short, coherent |
| container.go | 280 | Moderate — contains Container + Script builders; field duplication concerns apply |
| helpers.go | 234 | Moderate — SRP violation (see finding arch-helpers-srp-violation) |

No single production file exceeds 500 LOC. The test files are large but test files do not violate SRP in the same way. The `workflow_template.go` packing three CRD types (WorkflowTemplate, ClusterWorkflowTemplate, CronWorkflow) into one file is a mild cohesion issue; ClusterWorkflowTemplate could be in a separate file.

## Summary

The architecture is fundamentally sound for a pure builder SDK. Layering is clean. No cycles. The main structural issues are:
1. The globalConfig singleton pattern that was never completed into a proper DI solution (arch-high-1)
2. The expr.Eq()/Equals() naming confusion that is actively causing documentation errors (arch-medium-5, already manifested as completeness-readme-expr-eq-mismatch)
3. The type-unsafe interface{} use in PodDisruptionBudget and HTTP port fields (arch-medium-4)
4. The Container/Script field duplication that will accumulate maintenance drift (arch-medium-5)
