---
type: project-rule
domain: testing
language: go
specializes: ~/.claude/CLAUDE.md §7 Testes
created_at: 2026-05-20
---

# Testing — theo-forge

> Código sem teste é código que funciona por coincidência.
> — `~/.claude/CLAUDE.md §7`

## TDD is the default

RED → GREEN → REFACTOR.

```
RED:     Write a test that asserts the new behavior. Run it. Confirm it FAILS.
GREEN:   Write the minimal code to make the test pass. No more.
REFACTOR: Clean up duplication, naming, structure. Tests stay green.
```

Plans MUST include a `#### TDD` block per task listing every RED test by name and what it asserts. `/plan-confidence` hard-caps any **bug-fix task** without an explicit TDD block at ≤70.

## Pyramid (Go specialization)

```
                E2E (none here — SDK has no runtime)
              ──────────────────────────────────────
              Integration: client/ against httptest.NewServer
              Integration: serialize/ round-trip on 194 upstream Argo YAMLs
              ────────────────────────────────────────────────────────────
              Unit: builder validation, expr DSL, model JSON tags,
                    validate package, config hook dispatch
              ────────────────────────────────────────────────────────────
```

### Coverage targets

| Package | Minimum | Current (2026-05-20) |
|---------|---------|----------------------|
| root `forge` | 90% | 93.6% ✅ |
| `client` | 80% | 31% ❌ |
| `expr` | 80% | 0% ❌ |
| `config` | 90% | 0% ❌ |
| `serialize` | 90% | 0% ❌ |
| `validate` | 90% | 0% ❌ |
| `model` | 70% (via round-trip) | 0% direct; full coverage via integration ⚠️ |

`/plan-confidence` caps at ≤70 any plan that touches a `0%` package without adding tests for the touched code.

## Test layout

- Test file lives next to the file it tests: `container.go` ↔ `container_test.go`.
- Package convention: tests in the **same package** (`package forge`) to access unexported identifiers. Use `package forge_test` only when explicitly testing the public API surface (e.g., `examples_test.go`).
- Golden files in `testdata/` per Go convention. Use `-update` flag to regenerate, not `-update-golden`.

## Test naming

```go
// REQUIRED — describes behavior, not method
func TestContainer_BuildTemplate_RejectsEmptyImage(t *testing.T)
func TestWorkflowsService_CreateWorkflow_SetsBearerToken(t *testing.T)
func TestExprC_EscapesSingleQuotes(t *testing.T)

// FORBIDDEN — describes only mechanics
func TestBuild(t *testing.T)
func TestContainer1(t *testing.T)
```

Pattern: `Test<Type>_<Method>_<Scenario>` or `Test<Function>_<Scenario>`.

## Arrange-Act-Assert (mandatory)

```go
func TestContainer_BuildTemplate_RejectsEmptyImage(t *testing.T) {
    // Arrange
    c := &Container{Name: "x"}

    // Act
    _, err := c.BuildTemplate()

    // Assert
    if !errors.Is(err, ErrEmptyImage) {
        t.Fatalf("got %v, want ErrEmptyImage", err)
    }
}
```

Comments `// Arrange` / `// Act` / `// Assert` are optional but the three sections must be visually distinct.

## Table-driven tests

For builder validation and similar branching logic, **table-driven is required**:

```go
func TestContainer_BuildTemplate(t *testing.T) {
    tests := []struct {
        name    string
        in      *Container
        want    model.TemplateModel
        wantErr error
    }{
        {name: "minimum valid", in: &Container{Name: "x", Image: "alpine"}, want: ..., wantErr: nil},
        {name: "rejects empty image", in: &Container{Name: "x"}, wantErr: ErrEmptyImage},
        {name: "rejects empty name", in: &Container{Image: "alpine"}, wantErr: &ValidationError{Field: "Name"}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := tt.in.BuildTemplate()
            if tt.wantErr != nil {
                if !errors.Is(err, tt.wantErr) {
                    t.Fatalf("err = %v, want %v", err, tt.wantErr)
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if diff := cmp.Diff(tt.want, got); diff != "" {
                t.Errorf("mismatch (-want +got):\n%s", diff)
            }
        })
    }
}
```

## Regression-first for bug fixes

> A bug-fix task MUST include a test that REPRODUCES the bug (failing before the fix), the fix, and the same test passing after.
> — `~/.claude/CLAUDE.md §7`

Every fix from `review-output/` (path traversal, expr injection, NewConfig hook isolation, etc.) MUST be preceded by a failing test that demonstrates the defect on the current code.

`/plan-confidence` hard-caps any task fixing a finding from `review.db` that does NOT include a regression test at ≤70.

## Determinism

- No `time.Sleep` to wait for state. Use channels, contexts, or polling with timeout.
- No reliance on map iteration order.
- No randomized inputs without explicit seed (`math/rand` seeded; `crypto/rand` documented).
- Race detector (`go test -race`) MUST pass on every commit.
- Flaky tests are bugs. Fix or delete — never `t.Skip()`.

## Independence

- Tests MUST run in any order (`go test -shuffle=on`).
- No global state shared between tests (e.g., the `globalConfig` singleton is a known anti-pattern — tests calling `config.SetX()` MUST reset in `t.Cleanup`).
- No filesystem state leaked between tests. Use `t.TempDir()`.

## Mocks vs real implementations

| Scenario | Approach |
|----------|----------|
| HTTP client tests | `httptest.NewServer` with a handler asserting request shape. NEVER mock `*http.Client`. |
| YAML serialization | Real `sigs.k8s.io/yaml`. Round-trip against golden files. |
| Hook dispatch | Real `config.GlobalConfig` instance (call `config.NewConfig()`, register hook, build). |
| File I/O | `t.TempDir()`. Never `/tmp` or absolute paths. |

If a plan proposes mocking something other than a true external dependency (network, time, randomness), `/plan-confidence` flags it.

## What NOT to test

| Don't test | Why |
|------------|-----|
| Getters/setters that are trivial public fields | Tests provide no value beyond compilation. |
| `sigs.k8s.io/yaml` internals | It has its own tests. |
| Test for test-helpers (`ptrStr` et al.) | Unless the helper has non-trivial logic. |
| Code paths excluded from public API | Cover via the public API or delete. |

## CI gates (must pass)

Every PR:

```bash
gofmt -l .                      # empty
go vet ./...                    # clean
go test -race -count=1 ./...    # all pass
go test -coverprofile=cov.out ./... && go tool cover -func=cov.out  # meets package thresholds
golangci-lint run ./...         # clean (when wired)
```

`/plan-confidence` requires all five in the Global DoD.

## How `/plan-confidence` checks testing

- Bug-fix task without `#### TDD` block → ≤70 cap.
- Test name not matching `Test<Type>_<Method>_<Scenario>` → flag.
- Plan touching `expr`, `config`, `serialize`, or `validate` (0% coverage) without adding tests → ≤70 cap.
- Acceptance Criteria missing coverage check → flag.
