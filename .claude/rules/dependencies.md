---
type: project-rule
domain: dependencies
language: go
specializes: ~/.claude/CLAUDE.md §9 Não Reinvente a Roda
created_at: 2026-05-20
---

# Dependencies — theo-forge

> Antes de escrever qualquer código, pergunte: "alguém já resolveu isso?"
> — `~/.claude/CLAUDE.md §9`

theo-forge currently has **one** direct production dependency: `sigs.k8s.io/yaml`. That minimalism is intentional and worth defending.

## Current dependency graph

```
sigs.k8s.io/yaml v1.6.0              ← canonical k8s YAML (JSON-compatible)
  └── go.yaml.in/yaml/v2 v2.4.3
  └── go.yaml.in/yaml/v3 v3.0.4
```

Plus test/dev dependencies:

- `github.com/google/go-cmp` — diff in test output.
- `github.com/kr/pretty`, `github.com/rogpeppe/go-internal` (indirect)
- `gopkg.in/check.v1` (test-only)

## When adding a new dependency is allowed

The plan MUST satisfy ALL of:

1. **Real present need**, not anticipation (YAGNI alignment).
2. **No suitable existing dependency** already in the graph that would solve it.
3. **No stdlib equivalent** that solves it acceptably (Go stdlib is unusually rich — `net/http`, `encoding/json`, `crypto/*`, `context`, `sync`).
4. **Author reputation**: prefer `github.com/golang/*`, `sigs.k8s.io/*`, `golang.org/x/*`, well-known maintainers. Avoid `github.com/randomUser/*` unless audited.
5. **License**: MIT, Apache-2.0, BSD-3. NEVER GPL/AGPL.
6. **Maintenance**: release within last 12 months OR explicit "stable" status (e.g., `sigs.k8s.io/yaml` ships rarely but is k8s-backed).
7. **Transitive count**: prefer dependencies that pull 0-5 transitive deps. Anything pulling 20+ requires ADR justification.
8. **Security**: no known unpatched CVEs (check with `govulncheck`).

Plans adding a dependency MUST cite this checklist in the ADR Rationale and confirm each item.

## When reimplementing is allowed

Per CLAUDE.md §9, building from scratch is acceptable when:

- The problem is genuinely specific to theo-forge (e.g., the Argo expression DSL in `expr/`).
- Existing options have unacceptable risks (license, abandonment, security).
- The abstraction is so thin the dependency costs more than the code (e.g., a 10-line helper).

`expr/` is an example: there is no Go library for Argo's `{{ }}` and `{{= }}` templating, so we implement it ourselves.

## Forbidden patterns

- **Don't write your own**: JSON/YAML parser, JWT parser, UUID generator, HTTP client retry logic, struct→map reflector, k8s schema types.
- **Don't import**: anything that pulls `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go` (massive transitive trees; theo-forge mirrors the schema deliberately to stay light).
- **Don't fork**: prefer upstream + minor PRs.

## Tools / Actions specifically forbidden

| Tool | Status | Rationale |
|------|--------|-----------|
| `aquasecurity/trivy-action` (any version) | ❌ Forbidden | Compromised in March 2026 supply-chain attack — 75 tags force-pushed with malicious payload that stole CI/CD secrets. Even with the patched v0.35.0+, the historical exposure makes the action unsafe in our threat model. Use the `trivy` CLI binary (SHA-pinned) if you must, OR prefer `govulncheck` + `osv-scanner` + `nancy` (the gate suite in `docs/QUALITY-GATES.md`). |
| `trivy` CLI < v0.69.3 | ❌ Forbidden | Affected by the same incident. |
| GitHub Actions on `@master` / `@main` | ❌ Forbidden | Mutable refs are an active attack surface (see `INFRA-001` in `review-output/`). Pin to a SHA or, at minimum, a release tag and let Dependabot bump. |

See [docs/QUALITY-GATES.md §Tools NOT Used](../../docs/QUALITY-GATES.md) for the full avoid-list with rationale.

## When evaluating a candidate

Quick eval commands:

```bash
# Maintenance: latest release date and frequency
gh release list -R <owner>/<repo> | head -5

# Transitive deps it brings
go mod download <module>@<version>
go mod why -m <module>

# Vulnerabilities
govulncheck -mod=mod ./...

# License
go-licenses report <module>
```

## How `/plan-confidence` checks dependencies

- Plans adding a new entry to `go.mod` MUST include an ADR with the 8-item checklist above. Missing checklist → ≤70 cap.
- Plans removing a dependency MUST cite migration tests in TDD.
- Plans bumping a dependency major version MUST cite the changelog and breaking-change impact in Consequences.

## Specific decisions worth preserving

| Decision | Why | Where written |
|----------|-----|---------------|
| Mirror Argo CRD schema in `model/` (no `k8s.io/*` deps) | Keeps SDK lightweight (~5 MB binary impact vs hundreds). | [architecture.md](architecture.md) |
| Use `sigs.k8s.io/yaml` (not `gopkg.in/yaml.v3`) | Canonical k8s YAML semantics — fields, ordering, types match `kubectl` output. | go.mod |
| Single `Logger` interface for `client/` (NOT `slog`, NOT `logrus`) | Lets consumers inject their own logger; SDK stays log-library-agnostic. | Future plan (val-006 fix) |
