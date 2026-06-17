---
type: project-rule
domain: architecture
specializes: ~/.claude/CLAUDE.md §13 SOLID
created_at: 2026-05-20
---

# Architecture — theo-forge

theo-forge is a **layered Go SDK** with strict unidirectional dependency flow. This document defines the package boundaries, builder pattern contract, and SOLID specialization for Go.

## Package layout (canonical)

```
Consumer code
    │
    ▼
forge (root) ── builder API: Workflow, DAG, Steps, Container, Script, …
    │
    ├── config/    — GlobalConfig + hook system
    ├── expr/      — expression DSL ({{ }} / {{= }} Argo syntax)
    ├── model/     — serializable structs mirroring Argo CRD schema
    ├── serialize/ — YAML/JSON I/O via sigs.k8s.io/yaml
    ├── validate/  — CPU/memory unit validation
    └── client/    — REST client for Argo API
```

### Dependency direction (INVIOLABLE)

```
Layer 0:  model, expr                   (zero internal deps; stdlib + sigs.k8s.io/yaml only)
Layer 1:  config, serialize, validate, client   (may import model only)
Layer 2:  forge (root)                  (may import all sub-packages)
```

- **No sub-package** may import the root `forge` package. Use a local interface (DIP) when the sub-package needs to consume a builder type — see `client.Buildable`.
- **No layer** may import upward.
- **model/** is a pure data layer: zero business logic, only structs + JSON/YAML tags.

`/plan-confidence` hard-caps any ADR proposing a violation at ≤70 unless the ADR includes an explicit `Consequences` section justifying it and a tracking issue for the refactor.

## SOLID specialization for Go

### SRP — Single Responsibility

- A file mixing four+ responsibilities (e.g., `helpers.go` mixing build helpers, file I/O, validation, and config wiring) is an SRP violation. Split.
- A task editing >5 unrelated files signals SRP violation in the *change* itself. Split the task.

### OCP — Open/Closed (Go idiom)

- New behavior added via **new builder types** or **new template hooks**, not by editing existing `Build()` switch statements.
- New CRD types (next to `Workflow`, `WorkflowTemplate`, `CronWorkflow`) implement the same `Build() (model.XModel, error)` contract.

### LSP — Liskov Substitution

- Every type that satisfies `Templatable` (or any future interface) MUST return a non-zero, valid `model.TemplateModel` from `BuildTemplate()` for any valid receiver state. No "this method is not implemented" stubs.
- Subtype tests (e.g., a test that calls `BuildTemplate()` against every `Templatable` implementer with a baseline fixture) are required when the interface gains a new implementer.

### ISP — Interface Segregation

- New interfaces with > 5 methods are smell. Justify in ADR or split.
- Prefer role interfaces (`Buildable`, `Templatable`, `Serializer`) over header interfaces (`SDK`).

### DIP — Dependency Inversion

- The `client` package defines the `Buildable` interface locally to avoid importing the root package. **Pattern to follow** when adding any cross-package consumer.
- Infrastructure (HTTP transport, YAML library, logger) is injected via interface or functional option, never imported as a singleton.

## Builder pattern (canonical contract)

Every builder MUST follow this shape:

```go
type Container struct { /* public fields */ }

func (c *Container) BuildTemplate() (model.TemplateModel, error) {
    // 1. Validate required fields → return wrapped error
    // 2. Translate from builder struct to model struct
    // 3. Dispatch template hooks (via injected config, NOT package singleton — see ADR-001)
    // 4. Return (model, nil)
}

func (w *Workflow) Build() (model.WorkflowModel, error) {
    // Same shape, top-level Workflow analogue
}
```

- `Build()` and `BuildTemplate()` MUST return `(T, error)`. Never `(T, nil)` when validation fails.
- Builders are **value-mutating receivers** (`*T`), never builder-result types (`b.With(...)` returning a new struct). Keeps API ergonomic for Go.

## Anti-patterns specifically forbidden

| Anti-pattern | Where it bit us | Rule |
|--------------|-----------------|------|
| Package-level singleton captured at init (`var globalConfig = config.GetGlobal()` in `helpers.go`) | Made `NewConfig()` hooks unreachable. | Thread config through `Build()` or via opt-in functional option. Never freeze at init. |
| `interface{}` for int-or-string fields (`PodDisruptionBudget.MinAvailable`, `HTTPGetAction.Port`) | Allows `bool`, `float64`, structs to serialize as garbage YAML. | Define a local `IntOrString` type with constructor functions; never `interface{}` for typed-union fields. |
| Duplicated struct fields across `Container` / `Script` with drift | 8 fields silently drifted between the two types. | Extract `BaseTemplate` embedded struct. |
| Public mutable token field (`WorkflowsService.Token string`) | Token leaks via `%+v` logging; mutex bypass via direct write. | Unexported field + redacted `String()` method + constructor. |

## ADR requirement

Every plan that introduces a new public package, breaks the dependency direction, or adds a new builder type MUST contain an ADR with:

- Decision
- Rationale citing this file (or `~/.claude/CLAUDE.md §13`)
- Alternatives considered (at least one)
- Consequences

ADRs missing alternatives are hard-capped at ≤70 by `/plan-confidence` (golden rule).
