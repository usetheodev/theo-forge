# Quality Gates — theo-forge

**Authority:** this document is the single contract for what "passing CI"
means. The Makefile mechanizes it; `.golangci.yml` and `.testcoverage.yml`
encode the specifics. `.github/workflows/ci.yml` calls `make verify`
without redefining gates.

> Quality gates are **strict automated checkpoints**. They enforce
> predefined rules (e.g., minimum test coverage, zero security
> vulnerabilities). If code fails to meet these metrics, it is blocked
> from merging or deploying.

This is the Tier 2 / Strict implementation chosen on 2026-05-20 (see
`AskUserQuestion` answers in session log).

## Gate Catalog

| Gate | Tool | Threshold | Make target | Fails the build? |
|------|------|-----------|-------------|------------------|
| **Formatting** | `gofmt` (stdlib) | `gofmt -l .` MUST be empty | `verify-fmt` | Yes |
| **Static analysis** | `go vet` (stdlib) | Clean | `verify-vet` | Yes |
| **Lint** | `golangci-lint` (≥v1.61.0, config v2 strict) | Zero warnings, [zero suppression](#zero-suppression) | `verify-lint` | Yes |
| **Build** | `go build` | All packages compile | `verify-build` | Yes |
| **Race detection** | `go test -race -count=1` | All tests pass under the race detector | `verify-race` | Yes |
| **Per-package coverage** | `vladopajic/go-test-coverage` ([.testcoverage.yml](../.testcoverage.yml)) | Per-package floors + total ≥80% | `verify-coverage` | Yes |
| **Stdlib + module vulns** | `govulncheck` (golang.org/x/vuln) | Zero known CVEs | `verify-vuln` | Yes |
| **OSV broad scan** | `osv-scanner` (Google OSV.dev DB) | Zero advisories | `verify-osv` | Yes |
| **Sonatype OSS Index** | `nancy` (Sonatype) | Zero advisories | `verify-nancy` | Yes |

All gates aggregated under one target:

```bash
make verify
```

## Complexity Thresholds (golangci-lint settings)

| Metric | Linter | Limit | Rationale |
|--------|--------|-------|-----------|
| Cyclomatic complexity | `gocyclo` | ≤ **12** | Go-pragmatic — McCabe's classic 10 is too tight given Go's explicit error handling. Confirmed by community survey. |
| Cognitive complexity | `gocognit` | ≤ **20** | Better signal for "code that's hard to read" — weights nesting. |
| Function length | `funlen` | ≤ 80 lines / 50 statements | `.claude/rules/clean-code.md §Function size`. |
| Nested if depth | `nestif` | ≤ 4 | Deeper nesting indicates extract-function is warranted. |
| Package average complexity | `cyclop` | avg ≤ 8 | Package-level smoke gauge for systemic complexity creep. |

## Coverage Thresholds (.testcoverage.yml)

Global floors:

- **Total project:** 80%
- **Every package:** 70% (override below per-package as we backfill)

Per-package overrides (current vs. target):

| Package | Current | Threshold | Target (rules) | Status |
|---------|---------|-----------|----------------|--------|
| `forge` (root) | 91.8% | 90% | ≥ 90 | ✅ |
| `expr` | 100% | 80% | ≥ 80 | ✅ |
| `config` | 82.8% | 80% | ≥ 90 | ⏳ raise to 90 next |
| `serialize` | 51.8% | 50% | ≥ 90 | ⏳ raise progressively |
| `validate` | 66.2% | 65% | ≥ 90 | ⏳ raise progressively |
| `client` | 70.5% | 70% | ≥ 80 | ⏳ raise to 80 once `template_ops.go` has tests |
| `model` | 35.7% | 30% | ≥ 70 | ⏳ unlocks more after `IntOrString` (v0.6.0) |
| `docs` | n/a | 0% | n/a | docs/ packages exist only as compile-tests for README |

The progressive floors are deliberate — raise them as Phase 7 backfill
rounds land. Increase a per-package threshold whenever coverage rises
by ≥ 5pp above the current floor.

## Zero Suppression

`.golangci.yml`:

```yaml
issues:
  max-issues-per-linter: 0
  max-same-issues: 0
```

Plus `exclusions.warn-unused: true` — surfaces stale `//nolint` comments
so they cannot rot in the codebase.

`//nolint:LINTER` is permitted ONLY with a rationale on the same line:

```go
// OK:
foo() //nolint:gocyclo // gateway switch over RPC handlers — extract-function would just rename
// NOT OK:
foo() //nolint
```

The `revive: whyNoLint` check enforces this.

## Tools NOT Used (and Why)

| Tool | Reason |
|------|--------|
| `trivy` (and `aquasecurity/trivy-action`) | **Avoid.** Compromised in [March 2026 supply-chain attack](https://www.paloaltonetworks.com/blog/cloud-security/trivy-supply-chain-attack/) — 75 tags hijacked in the official GitHub Action. Even though patched (≥v0.69.3 / trivy-action ≥v0.35.0), the historical exposure + our `govulncheck`+`osv-scanner`+`nancy` triple gives equivalent SCA coverage with smaller attack surface. Recorded in `.claude/rules/dependencies.md`. |
| `staticcheck` (standalone) | Already runs as the `staticcheck` linter inside `golangci-lint`. Running it separately wastes CI minutes. |
| `gosec` (standalone) | Same — runs as the `gosec` linter inside `golangci-lint`. |
| `gocyclo` / `gocognit` (standalone) | Same — both run inside `golangci-lint`. The CLI versions are useful locally for debugging a specific function. |

## Local Workflow

```bash
# One-time: install all tools into $GOPATH/bin.
make install-tools

# Before every commit: run the full gate locally.
make verify

# Faster iteration during development:
make verify-fmt      # gofmt only
make verify-lint     # golangci-lint only
make verify-race     # tests w/ -race
make verify-coverage # coverage thresholds
```

`make verify` is non-destructive (reads only) and exits non-zero on any
failure, suitable for git pre-push hooks if desired.

## CI Workflow

`.github/workflows/ci.yml` runs `make verify` on every push and pull
request. Branch protection (see `docs/REPO-SETTINGS.md`) requires the
`Quality Gates (make verify)` job to pass before merge.

## When to Adjust Thresholds

Raise (tighten):

- After a coverage backfill round increases a package by ≥ 5pp above
  the current floor.
- After a refactor reduces ciclomática across the codebase.

Lower (loosen) — REQUIRES an ADR:

- Demonstrate that the existing threshold blocks correct code (with a
  real PR showing the false-positive pattern).
- Open an issue, link the ADR in `.claude/knowledge-base/plans/`.
- Bump `.golangci.yml` and `.testcoverage.yml` in the same PR; include
  before/after metrics in the commit message.

Adding a new gate or removing an existing one ALWAYS requires an ADR.

## Sources & References

- [golangci-lint v2 announcement](https://ldez.github.io/blog/2025/03/23/golangci-lint-v2/)
- [Golangci-lint configuration reference](https://golangci-lint.run/docs/configuration/file/)
- [gocyclo](https://github.com/fzipp/gocyclo)
- [gocognit](https://github.com/uudashr/gocognit)
- [vladopajic/go-test-coverage](https://github.com/vladopajic/go-test-coverage)
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- [osv-scanner](https://github.com/google/osv-scanner)
- [nancy](https://github.com/sonatype-nexus-community/nancy)
- [Trivy supply chain attack — Palo Alto Networks](https://www.paloaltonetworks.com/blog/cloud-security/trivy-supply-chain-attack/)
- [Cognitive Complexity whitepaper — G. Ann Campbell, SonarSource](https://www.sonarsource.com/resources/cognitive-complexity/)
