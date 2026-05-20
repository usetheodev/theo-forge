---
type: project-rules-index
project: theo-forge
language: go
authority: ~/.claude/CLAUDE.md (Princípios Inquebráveis v2.0) is the SUPERIOR authority. These project rules SPECIALIZE the global principles for this Go SDK and MUST NOT contradict them.
created_at: 2026-05-20
---

# theo-forge — Project Rules

theo-forge is a Go SDK for building Argo Workflows manifests. The rules in this directory specialize the global `~/.claude/CLAUDE.md` Princípios Inquebráveis for this codebase.

## Authority order (top wins)

1. `~/.claude/CLAUDE.md` — global Princípios Inquebráveis (v2.0). Universal, non-negotiable.
2. Files in this directory — project-specific specializations.
3. `.claude/skills/plan-confidence/defaults/` — fallback only (loaded when this directory is empty; ignored when these files exist).

## File index

| File | Scope | Hard caps in `/plan-confidence` |
|------|-------|---------------------------------|
| [architecture.md](architecture.md) | Layering, SOLID, package boundaries, builder pattern, hooks | SRP violation in ADR ⇒ ≤70 |
| [golang-conventions.md](golang-conventions.md) | gofmt, naming, error wrapping, context, panics, struct tags | `gofmt -l` non-empty in DoD ⇒ ≤70 |
| [clean-code.md](clean-code.md) | Naming, function size, comments, dead code | Naked TODO in committed code ⇒ ≤70 |
| [dry.md](dry.md) | Duplication of knowledge (not lines) | New utility duplicating existing helper ⇒ ≤70 |
| [error-handling.md](error-handling.md) | Fail-fast, `%w` wrapping, no silent swallowing | `_ = err` or empty err branch ⇒ ≤70 |
| [testing.md](testing.md) | TDD, table-driven, regression-first, pyramid | Bug-fix without RED test ⇒ ≤70 |
| [dependencies.md](dependencies.md) | Não reinventar a roda; dependency evaluation; trivy avoid-list | New stdlib reimplementation ⇒ ≤70 |
| [size-allowlist.txt](size-allowlist.txt) | File size budget (500 LoC default for Go) | File exceeds budget without ADR justification ⇒ flag |

## Quality Gates contract

For the rigorous quality-gate suite (golangci-lint v2 strict, per-package coverage, complexity, race detection, multi-scanner SCA), see [docs/QUALITY-GATES.md](../../docs/QUALITY-GATES.md). The Makefile target `make verify` mechanizes every gate; CI workflows delegate to it (single source of truth).

## What these rules are NOT

- A coding style nag list. Style is `gofmt` + `golangci-lint`. These rules describe *engineering decisions* a plan must defend.
- A wishlist. Every rule here has a concrete enforcement point in `/plan-confidence` or `/code-audit`.
- Mutable. Changes go through PR review like any other code change.

## Golden rule

> Código bom não é o que demonstra que somos inteligentes. É o que demonstra que respeitamos quem vai ler depois. — `~/.claude/CLAUDE.md`
