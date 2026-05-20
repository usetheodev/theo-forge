---
type: project-rule
domain: language-conventions
language: go
references:
  - https://go.dev/doc/effective_go
  - https://github.com/uber-go/guide
  - https://github.com/golang/go/wiki/CodeReviewComments
created_at: 2026-05-20
---

# Go Conventions — theo-forge

Go has idiomatic patterns the community has converged on. theo-forge follows them. This file lists the **non-negotiable** ones for plans and reviews.

## Formatting

- `gofmt -l .` MUST return empty on every commit. CI gate.
- `goimports -l .` MUST return empty (import grouping: stdlib / external / internal, separated by blank line).
- No tabs-vs-spaces debate — `gofmt` settles it.

`/plan-confidence` caps at ≤70 any plan DoD that does not include a `gofmt -l .` check.

## Naming

| Element | Rule | Bad | Good |
|---------|------|-----|------|
| Exported identifiers | UpperCamelCase, no underscores, no abbreviations | `BUILD_WORKFLOW`, `Http_Client` | `BuildWorkflow`, `HTTPClient` |
| Acronyms in names | All-caps (`HTTP`, `URL`, `ID`, `API`) | `HttpClient`, `Url`, `Id` | `HTTPClient`, `URL`, `ID` |
| Receivers | 1-2 letter abbreviation of the type | `func (this *Workflow)` | `func (w *Workflow)` |
| Errors | `Err` prefix for sentinel; `Error` suffix for types | `NotFound` | `ErrNotFound`, `ValidationError` |
| Interfaces | `-er` suffix for single-method | `BuildInterface` | `Builder`, `Templater` |
| Booleans | `is*`, `has*`, `can*` predicates | `valid bool` | `isValid bool` |
| Constructors | `New<Type>` returns `*Type` or `Type` | `MakeWorkflow` | `NewWorkflow` |

Package names: lowercase, single word, no underscores, no plural. `client`, `model`, `expr` — never `clients`, `workflow_model`, `expressions`.

## Errors

### Wrapping (Go 1.13+)

```go
// REQUIRED
return fmt.Errorf("build container %q: %w", c.Name, err)

// FORBIDDEN
return errors.New("error")              // no context
return fmt.Errorf("error: %v", err)     // loses wrap chain
return err                              // unless deliberate at boundary
```

### Sentinel errors

```go
// REQUIRED for known error categories
var ErrEmptyImage = errors.New("forge: container image is required")
```

### Custom error types when callers need to inspect

```go
type ValidationError struct {
    Field string
    Value any
    Cause error
}

func (e *ValidationError) Error() string { return ... }
func (e *ValidationError) Unwrap() error { return e.Cause }
```

### Forbidden

- `panic()` outside `init()` or genuinely unrecoverable conditions.
- `recover()` to swallow programming errors silently.
- `if err != nil { return nil }` — see [error-handling.md](error-handling.md).
- `_ = someCall()` when `someCall` returns an error.

## Context

- Any function performing I/O (HTTP, file, time) MUST accept `ctx context.Context` as the FIRST parameter.
- `context.Background()` only at the top level (`main()`, test) or in long-running services. Library code never invents a context.
- Never store `context.Context` in a struct field. Pass it.
- `ctx, cancel := context.WithTimeout(parent, d)` MUST be paired with `defer cancel()`.

## Concurrency

- Shared mutable state needs explicit synchronization (`sync.RWMutex`, `atomic`, channels). Document the invariant in a comment.
- Exported struct fields cannot be concurrency-safe by themselves — if a struct uses a mutex, the protected fields MUST be unexported.
- `go func() { ... }()` without a way to wait for completion is a goroutine leak. Use `sync.WaitGroup`, `errgroup.Group`, or a done-channel.
- Never `time.Sleep()` to wait for goroutines — synchronize properly.

## HTTP client

- Reuse a single `*http.Client` per service. Don't create one per request.
- Always set explicit timeouts: `&http.Client{Timeout: 30 * time.Second}` or per-request via context.
- Always `defer resp.Body.Close()` immediately after `err` check.
- Always limit body reads: `io.ReadAll(io.LimitReader(resp.Body, maxSize))`. theo-forge max: 32 MiB.
- TLS configuration is explicit: never silently rely on defaults. If a `VerifySSL` flag exists, it MUST be wired to `tls.Config`.

## Struct tags

- JSON: `json:"name,omitempty"` for optional fields. **YAML omitempty must match** when the struct serializes to both formats.
- YAML output goes through `sigs.k8s.io/yaml` (canonical JSON-compatible YAML) — design tags for JSON; YAML follows.
- `omitempty` on slices/maps means *nil* is omitted but *empty non-nil* is rendered. Pick deliberately.

## Generics (Go 1.18+)

- Use generics only when at least 2 concrete instantiations exist OR the alternative is `interface{}` + type assertion.
- Generic function name follows convention `Map[T, U]`, `Filter[T]`. Type parameters: single uppercase letter when meaning is obvious.

## Function design

- Public functions MUST have a doc comment starting with the function name: `// NewWorkflow returns ...`.
- Function size budget: ≤ 50 lines (see [clean-code.md](clean-code.md)).
- Number of return values: ≤ 3 (typically `(T, error)`). More signals a struct return is needed.
- Variadic only when callers genuinely pass 0-N; never for "one-or-many" (define an overload via slices).

## Imports

```go
import (
    "context"        // stdlib first
    "fmt"

    "sigs.k8s.io/yaml"  // external next, blank line above

    "github.com/usetheodev/theo-forge/model"  // internal last, blank line above
)
```

`goimports` enforces this.

## Forbidden patterns (caught by `golangci-lint`)

- Shadowed variable names (`govet`).
- Unused imports / variables (`unused`).
- Returning the same value from both branches of `if err != nil` (`gocritic`).
- `bytes.Compare(a, b) == 0` instead of `bytes.Equal(a, b)` (`gocritic`).
- `for _, v := range slice` when iterating large structs by value when only the pointer is needed (`rangeValCopy`).

## `golangci-lint` config

The project ships a `.golangci.yml` (v2 schema). Plans changing the lint config MUST include an ADR justifying the change.

`/plan-confidence` caps at ≤70 any plan that disables an existing linter without ADR justification.
