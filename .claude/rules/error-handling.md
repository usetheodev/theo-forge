---
type: project-rule
domain: error-handling
language: go
specializes: ~/.claude/CLAUDE.md §8 Error Handling (Fail-Fast)
created_at: 2026-05-20
---

# Error Handling — theo-forge

> Falhe alto, falhe cedo, falhe claro.
> — `~/.claude/CLAUDE.md §8`

For a Go SDK, error handling is the most visible part of the API contract. Sloppy error handling forces consumers to debug the SDK itself.

## Hierarchy

```
1. VALIDATE on entry        — reject invalid inputs at the builder/client boundary
2. FAIL fast                — return as soon as something is wrong
3. FAIL clear               — wrapped with %w, specific context, actionable
4. FAIL high                — propagate up to whoever can decide what to do
5. LOG with context          — when finally caught at top-level, log fields, not concatenated strings
6. RECOVER selectively      — retry/backoff only where it makes sense (idempotent ops)
```

## Required patterns

### Wrap with context using `%w`

```go
// REQUIRED
return fmt.Errorf("build container %q: %w", c.Name, err)

// FORBIDDEN — loses wrap chain, breaks errors.Is/As
return fmt.Errorf("build container: %v", err)

// FORBIDDEN — no context for caller
return err
```

Exception: returning `err` unchanged is acceptable at the **bottom** of a stack frame (e.g., a thin wrapper around `yaml.Marshal`) when adding context would be noise.

### Sentinel errors for known categories

```go
// model/errors.go
var (
    ErrEmptyImage          = errors.New("forge: container image is required")
    ErrTemplateAmbiguous   = errors.New("forge: exactly one of Template/TemplateRef/Inline must be set")
    ErrPathTraversal       = errors.New("forge: file name escapes output directory")
    ErrInvalidExpression   = errors.New("forge: invalid Argo expression")
)
```

Callers use `errors.Is(err, ErrEmptyImage)` to handle specific cases.

### Typed errors when callers need fields

```go
type ValidationError struct {
    Field string
    Value any
    Want  string
    Cause error
}

func (e *ValidationError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("validation failed for %s=%v (want %s): %v", e.Field, e.Value, e.Want, e.Cause)
    }
    return fmt.Sprintf("validation failed for %s=%v (want %s)", e.Field, e.Value, e.Want)
}

func (e *ValidationError) Unwrap() error { return e.Cause }
```

## Forbidden patterns

### Silent swallow

```go
// FORBIDDEN — error vanishes
result, _ := buildSomething()

// FORBIDDEN — error logged, never returned
if err := doWork(); err != nil {
    log.Println(err)
}

// FORBIDDEN — converts error to nil sentinel
if err != nil {
    return nil
}
```

### Catch-all panic recovery

```go
// FORBIDDEN — hides programming errors
defer func() {
    if r := recover(); r != nil {
        // swallow
    }
}()
```

`recover()` is acceptable ONLY in server-style code that must keep running (not applicable to this SDK). For an SDK, panics escape — they signal programmer error in the consumer code.

### Magic return values

```go
// FORBIDDEN — caller cannot distinguish "not found" from "empty"
func GetWorkflow(name string) string {
    return ""  // means error
}

// REQUIRED
func GetWorkflow(name string) (string, error) {
    return "", ErrNotFound
}
```

### Bare error strings without context

```go
// FORBIDDEN
return errors.New("error")
return errors.New("failed")

// REQUIRED
return errors.New("forge: workflow name must match k8s naming rules")
```

All public-package errors MUST be prefixed with `forge:` or the sub-package name (`client:`, `expr:`, `serialize:`) for diagnostic clarity.

## Validation at boundaries

Every `Build*()` and every client method MUST validate its inputs FIRST. Validation errors return `ErrXxx` sentinels or `*ValidationError`.

```go
func (c *Container) BuildTemplate() (model.TemplateModel, error) {
    if c.Image == "" {
        return model.TemplateModel{}, ErrEmptyImage
    }
    if c.Name == "" {
        return model.TemplateModel{}, &ValidationError{Field: "Name", Want: "non-empty"}
    }
    // ...
}
```

## HTTP error handling

`client/` requires:

```go
resp, err := c.http.Do(req)
if err != nil {
    return nil, fmt.Errorf("argo: %s %s: %w", req.Method, req.URL.Path, err)
}
defer resp.Body.Close()

body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes))
if err != nil {
    return nil, fmt.Errorf("argo: read response: %w", err)
}

if resp.StatusCode >= 400 {
    return nil, &APIError{Status: resp.StatusCode, Body: body, Method: req.Method, Path: req.URL.Path}
}
```

- `defer resp.Body.Close()` MUST appear immediately after the err check.
- Body reads MUST be bounded (see SEC-005 in review).
- Non-2xx responses MUST be converted to typed `*APIError` for caller inspection.

## Test requirements

Every error path MUST be tested:

```go
func TestContainer_BuildTemplate_EmptyImage(t *testing.T) {
    c := &Container{Name: "x"}
    _, err := c.BuildTemplate()
    if !errors.Is(err, ErrEmptyImage) {
        t.Fatalf("got %v, want ErrEmptyImage", err)
    }
}
```

See [testing.md](testing.md).

## How `/plan-confidence` checks error handling

- Plans introducing a function returning `(T, error)` MUST list at least one error-path test in `#### TDD`.
- Plans modifying existing error paths MUST preserve `errors.Is/As` compatibility (cite in Consequences).
- Pattern `_ = err` or `if err != nil { return nil }` mentioned positively in plan → ≤70 cap.
- HTTP client changes without `LimitReader` and `defer Close` → ≤70 cap.
