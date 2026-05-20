---
type: project-rule
domain: dry
language: go
specializes: ~/.claude/CLAUDE.md §12 DRY
created_at: 2026-05-20
---

# DRY — theo-forge

> Cada conhecimento do sistema deve ter uma representação única e autoritativa.
> — `~/.claude/CLAUDE.md §12`

DRY is about **knowledge and business rules**, not lines of code. Two functions that look identical but represent different domain concepts should NOT be merged.

## What counts as duplication (REAL)

| Example in theo-forge | Why it's duplication |
|-----------------------|----------------------|
| Argo schema field defaults defined in `container.go` AND `script.go` separately | Same knowledge ("the Argo default for X is Y") in two places. |
| `Container` and `Script` redefining the same 20 fields with no shared base | Same data model expressed twice; drift already observed (`SecurityContext`, `InitContainers`, etc. missing on Script). |
| `buildInputs` / `buildOutputs` / `buildEnv` copy-pasted across builders | Removed in v0.2.1 via `build_helpers.go` — keep that discipline. |
| Validation logic for CPU/memory units inlined in builders | All validation lives in `validate/` — never inline. |
| Sentinel error strings spelled differently across packages | `ErrEmptyImage` (canonical) vs `errors.New("missing image")` (drift). Define once, reuse. |

## What is NOT duplication (FALSE positives)

| Example | Why it's NOT a violation |
|---------|--------------------------|
| `BuildTemplate()` methods on Container, Script, DAG, Steps — all start with validation, hook dispatch, return | Same *structure*, different *responsibilities*. Each handles its own template type. |
| Error wrapping format `fmt.Errorf("operation %q: %w", name, err)` repeated everywhere | Idiomatic Go pattern; not knowledge. |
| Test helper `ptrStr`, `ptrBool`, `ptrInt` already exist in `testutil_test.go` | Centralized; if a new test file copies them, THAT is duplication (use the existing helpers). |

## Rule of Three

> Only extract an abstraction when the same knowledge appears for the **third** time.

Premature abstraction (`interface Foo { Bar() }` with one implementer) is worse than mild duplication. Wait for the third concrete need.

## Application to theo-forge

### When proposing a new helper

Before adding to `helpers.go` or any new `*_helpers.go` file, the plan MUST search for existing helpers:

```bash
# Inventory current helpers
grep -nE 'func \w+\(' helpers.go build_helpers.go testutil_test.go 2>/dev/null
```

Cite the search in the ADR Rationale. New helpers that duplicate existing functionality are hard-capped at ≤70 by `/plan-confidence`.

### When refactoring duplicate fields

For the known Container/Script duplication (P3 in the review):

- Extract `BaseTemplate` embedded struct holding the ~20 shared fields.
- Container and Script each `BaseTemplate` plus their unique fields.
- `BuildTemplate()` for both calls a shared `buildBaseTemplate(b *BaseTemplate, out *model.TemplateModel) error` helper.
- Tests cover both types via table-driven cases.

### When centralizing constants

| Constant | Single source of truth |
|----------|------------------------|
| Argo CRD API version (`argoproj.io/v1alpha1`) | `model/types.go` |
| Default container image (if any) | `config.GlobalConfig` |
| HTTP client defaults (timeout, max body) | `client/client.go` |
| Resource unit limits | `validate/units.go` |

A magic string `"argoproj.io/v1alpha1"` appearing in a builder file is a DRY violation.

## How `/plan-confidence` checks DRY

- Plans introducing new utility files that duplicate existing helpers → ≤70 cap.
- ADRs creating abstractions for a single use case (no Rule of Three justification) → flagged as YAGNI violation.
- Plans removing duplication MUST cite the duplication source files in Evidence.

## Coordination with `/plan-improve`

When `/plan-improve` Phase B (LLM-driven) considers extracting an abstraction:

- It MUST search for existing implementations FIRST.
- If a candidate exists, the improvement is "reuse existing" not "extract new".
- If no candidate exists AND Rule of Three is met, propose extraction with ADR.
