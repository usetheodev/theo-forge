---
type: project-rule
domain: clean-code
language: go
specializes: ~/.claude/CLAUDE.md §5 Nomenclatura
created_at: 2026-05-20
---

# Clean Code — theo-forge

Specialized for Go. Project-specific override of `~/.claude/skills/plan-confidence/defaults/clean-code.md`.

## Naming

Follow [golang-conventions.md §Naming](golang-conventions.md). Additional project-specific rules:

| Concept | Required name | Rationale |
|---------|---------------|-----------|
| The builder root types | `Workflow`, `WorkflowTemplate`, `CronWorkflow`, `ClusterWorkflowTemplate` | Mirror Argo CRD names exactly. |
| Builder produces model type | `Build() (model.XModel, error)` | Always returns `model.XModel`, never re-exports. |
| Template builders | `BuildTemplate() (model.TemplateModel, error)` | Standard contract for `Templatable`. |
| Field setters | `Set<Field>` (rare — public fields preferred for builders) | Reserve setters for fields needing validation. |
| Adders (append to slice) | `Add<Singular>(item T)` | `AddTask`, not `AddTasks`. |
| Conditional builders | `When<Cond>(c Expr)` | Argo terminology. |

Avoid Go anti-patterns:

- `Get<Field>()` only when the getter does work (mutex, lazy init). Plain accessors should be exported fields.
- `Helpers` / `Utils` / `Manager` — these are dumping-ground names. Be specific (`expression.go`, `pod_affinity.go`).
- One-letter receivers are fine (`w *Workflow`), but `me`, `this`, `self` are forbidden.

## Function size

- **Default budget: ≤ 50 lines.** Pipeline functions without branching may exceed if reading top-to-bottom requires no jumps.
- Functions > 50 lines without justification → `/code-audit` complexity flag.
- Cyclomatic complexity ≤ 10. Higher requires ADR or refactor.

## Comments

- **Default: no comments.** Well-named identifiers explain what.
- Write a comment only when the **WHY** is non-obvious: a hidden Argo behavior, a workaround for a known upstream bug, an invariant that the type system cannot express.
- Public identifiers MUST have a doc comment (Go convention; enforced by `revive`/`golint`).
- Forbidden:
  - Commented-out code (use git).
  - Comments restating the function name in prose.
  - Stale comments (review during PR; delete if no longer accurate).

### TODO markers

`TODO` / `FIXME` / `HACK` / `XXX` MUST include at minimum:

```go
// TODO(@usetheodev, 2026-06-30): replace interface{} with IntOrString once #45 lands.
```

A naked `TODO` is rot. `/plan-confidence` caps at ≤70 any plan that introduces a naked TODO.

## Dead code

- Unused imports / variables / functions: DELETE. Don't ship "might need later" code.
- Empty exported functions / fields that are documented but never read (e.g., the original `VerifySSL`, `GlobalConfig.ServiceAccountName`): either WIRE or DELETE. No middle ground.
- "Generated files" are exempt — but theo-forge has none today.

## Magic numbers

- Constants for numbers used > once OR with semantic meaning. `const maxResponseBodyBytes = 32 << 20`, not `32 * 1024 * 1024`.
- Time durations as typed constants: `const defaultClientTimeout = 30 * time.Second`.

## Validation at boundaries

Per `~/.claude/CLAUDE.md §8`: validate at system boundaries (controllers, consumers, handlers). For an SDK, "system boundaries" are:

- Builder constructors and `Build*()` methods (validate before producing model).
- Client request methods (validate name, namespace, payload before HTTP).
- Serialize entry points (`ToFile`, `FromYAML` — validate path containment, source size).
- Public functional options (validate values inside the option function).

Internal helper functions trust their inputs.

## Examples

| Bad | Good | Why |
|-----|------|-----|
| `func (w *Workflow) Build() *model.WorkflowModel` | `func (w *Workflow) Build() (model.WorkflowModel, error)` | Errors are explicit, by value not pointer (no nil-check ambiguity). |
| `if c.Image == "" { return nil, nil }` | `if c.Image == "" { return model.TemplateModel{}, ErrEmptyImage }` | Silent failure vs explicit, with sentinel. |
| `data := getData()` (returns `interface{}`) | `data, err := getTypedData()` (returns `MyType`) | Type-safety preferred. |
| Method `func (s *Script) Build() {...}` with no error | `func (s *Script) BuildTemplate() (model.TemplateModel, error) {...}` | Consistent contract. |

## How `/plan-confidence` checks this

- Tasks with `Files to edit` listing files > 500 LoC after the edit → size flag.
- Functions implied > 50 lines (judged from task description) → complexity flag.
- Naked TODO mentioned in plan body → ≤70 cap.
- Acceptance Criteria must include at least one of: complexity check, coverage check, size check, lint check.
