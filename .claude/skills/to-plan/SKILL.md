---
name: to-plan
description: Turn the current conversation context into an implementation plan and save it to .claude/knowledge-base/plans/. Use when user wants to create a plan from the current context.
---

This skill takes the current conversation context and codebase understanding and produces a detailed implementation plan. Do NOT interview the user — synthesize what you already know.

## Process

### Step 0 — MANDATORY: Read `.claude/rules/` BEFORE anything else

Every plan SHALL comply 100% with the architecture rules and project patterns in `.claude/rules/`. This is the FIRST thing the agent does — before exploring code, before drafting the plan, before writing ADRs.

```bash
ls .claude/rules/ 2>/dev/null | head -40
```

Read each `.md` file in `.claude/rules/`. Internalize:

- **architecture-related rules** (e.g., `architecture.md`, `domain-boundary.md`, `architecture-contract.yaml`) — dependency direction, layering, allowed imports
- **testing rules** (e.g., `testing.md`) — TDD, coverage, fixture conventions
- **language conventions** (e.g., `rust-conventions.md`, `frontend.md`) — naming, error handling, idioms
- **golden rules** (e.g., `dogfood-golden-rule.md`, `plan-confidence-golden-rule.md`) — inviolable constraints
- **integration / doc / governance rules** (e.g., `integration-first.md`, `auto-docs.md`, `learning-system.md`) — workflow contracts
- **allowlists with sunset** (e.g., `complexity-allowlist.txt`, `unwrap-allowlist.txt`, `size-allowlist.txt`) — known-debt entries

**Fallback (if `.claude/rules/` does NOT exist or is empty):** read defaults bundled at `.claude/skills/plan-confidence/defaults/` (SOLID, DRY, Clean Code, LoC ~500, testing). These defaults are FALLBACK ONLY — project rules win when present.

The plan SHALL:

- Cite at least one rule by filename (e.g., `architecture.md`) or principle (SOLID/DRY/KISS/YAGNI) in the ADR Rationale section.
- Respect file size budgets in `Files to edit` (default 500 LoC, see project's `size-allowlist.txt` or defaults).
- Honor crate dependency direction if `architecture-contract.yaml` exists.
- Include tests per `testing.md` (or defaults).
- Include Global DoD entries that reference quality gates (lint, complexity, size, code-audit).

If the agent does not honor Step 0, `/plan-confidence` will deduct via the `architecture_compliance` sub-report and may apply `soft_floor_low_architecture_compliance` (cap 89).

### Step 1 — Explore the repo

After reading the rules, explore the codebase to understand the current state, patterns, and architecture in light of those rules.

### Step 2 — Architecture Snapshot (BEFORE)

Run `/architecture-docs {domain}` for the affected domain(s). This saves the current-state C4 docs to `.claude/knowledge-base/architecture/{domain}/`. This is the baseline before the plan changes anything.

### Step 3 — Identify the modules

You will need to build or modify. Actively look for opportunities to extract deep modules (lots of functionality behind a simple, testable interface that rarely changes). Check with the user that these modules match their expectations and which modules need tests.

**Compliance discipline:** for each module decision, justify against the rules read in Step 0. Example: "We split `foo_service.rs` into 3 files because it would exceed 500 LoC per `size-allowlist.txt`" or "We use trait `Foo` over concrete `FooImpl` per DIP (per `architecture.md`)".

### Step 4 — Write the plan

Use the template below and save to `.claude/knowledge-base/plans/{slug}-plan.md`. The slug should be kebab-case derived from the plan title.

The plan MUST include:

- ADR Rationale citing project rules or principles (Step 0).
- Global DoD with quality-gate entries (lint, complexity, size).
- File size budget mention (default 500 LoC OR project-specific from `size-allowlist.txt`).

## Plan Template

Every plan MUST follow this structure. Each section is mandatory unless marked (optional).

<plan-template>

# Plan: {Title}

> **Version 1.0** — one-paragraph executive summary explaining what this plan does, why it matters, and what the expected outcome is.

## Context

What exists today, what's broken or missing, and what evidence (data, logs, user reports, benchmark results) motivates this work. Include links to issues, PRs, or ADRs when available.

## Objective

One clear sentence: what does "done" look like? Then a short list of specific, measurable goals.

## ADRs

Architecture Decision Records for this plan. Each decision gets:
- **ID** (D1, D2, ...) for cross-referencing in tasks
- **Decision** — what was decided
- **Rationale** — why this approach over alternatives
- **Consequences** — what this enables and what it constrains

## Dependency Graph

ASCII diagram showing phase dependencies. Example:

```
Phase 0 ──▶ Phase 1 ──▶ Phase 2 ──▶ Phase 3
                │                       │
                │                       ▼
                │               Phase 3.1
                │
                ▼
          Phase 4 (parallel)
```

Annotate which phases can run in parallel vs which are sequential blockers.

---

## Phase N: {Title}

**Objective:** one sentence describing what this phase achieves.

### T{N}.{M} — {Task Title}

#### Objective
What this specific task accomplishes.

#### Evidence
Data, logs, or observations that justify this task. Why it's needed NOW, not later.

#### Files to edit
```
path/to/file.rs — what changes and why
path/to/other.rs — what changes and why
```

#### Deep file dependency analysis
For each file listed above, explain:
- What the file does today
- How this task changes it
- What downstream files depend on this change

#### Deep Dives
Technical details for non-obvious aspects:
- Data structures: exact fields, types, derive macros
- Algorithms: step-by-step logic
- Invariants: what MUST be true before and after
- Edge cases: empty inputs, zero values, missing fields, backward compat

#### Tasks
Numbered checklist of atomic implementation steps:
1. Step one
2. Step two
3. ...

#### TDD
Strict RED-GREEN-REFACTOR cycle. List every test FIRST:

```
RED:     test_name() — what it asserts (this test MUST fail before implementation)
RED:     test_name_2() — what it asserts
GREEN:   Implement the minimal code to make all RED tests pass
REFACTOR: What cleanup is expected (or "None expected")
VERIFY:  cargo test -p {crate} / pytest {path}
```

#### Acceptance Criteria
Bulleted list of observable, verifiable conditions:
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Pass: /code-audit complexity check (cyclomatic complexity <= 10)
- [ ] Pass: /code-audit coverage check (coverage >= 90%)
- [ ] Pass: /code-audit lint check (zero warnings)
- [ ] Pass: /code-audit size check (file <= 500 lines)

#### DoD (Definition of Done)
- [ ] All tasks completed and validated
- [ ] All tests passing
- [ ] Zero clippy warnings
- [ ] cargo test / pytest green
- [ ] code-audit checks passing

---

(Repeat `### T{N}.{M}` for each task in the phase)

(Repeat `## Phase N` for each phase)

---

## Coverage Matrix

Table mapping original gaps/requirements to tasks:

| # | Gap / Requirement | Task(s) | Resolution |
|---|---|---|---|
| 1 | Description | T{N}.{M} | How it's resolved |

**Coverage: X/Y gaps covered (Z%)**

## Global Definition of Done

- [ ] All phases completed
- [ ] All tests passing (Rust + Python where applicable)
- [ ] Zero clippy/lint warnings
- [ ] Backward compatibility preserved
- [ ] code-audit checks passing across all modified crates
- [ ] Plan-specific criteria (list them)
- [ ] **Dogfood QA PASS** — `/dogfood full` health score >= 70, zero CRITICAL issues
- [ ] **Runtime-metric proof** — for every task whose DoD references a runtime counter (hit rate, miss rate, fold count, etc.), the metric MUST be observed non-zero in a real workload against the wired binary path (smoke suite or dedicated scenario), not just verified to compile. Lesson from `sota-context-engineering` (commit `e67d134`): two prior dogfoods accepted "code exists + tests pass" and missed two real wiring bugs that only fired in benchmark workloads. See `.claude/rules/integration-first.md` §"Runtime-Metric Acceptance".

## Final Phase: Dogfood QA (MANDATORY)

> This phase runs AFTER all implementation phases are complete. The plan is NOT done until dogfood passes.

**Objective:** Validate that the implemented changes work as a real user would experience them, not just as unit tests assert.

### Execution

Run `/dogfood full`. Always full. No shortcuts.

### Acceptance Criteria

- [ ] Health score >= 70/100
- [ ] Zero CRITICAL issues introduced by this plan's changes
- [ ] Zero HIGH issues in commands/features modified by this plan
- [ ] Any pre-existing issues documented (not caused by this plan)

### If Dogfood Fails

1. Identify which issues are caused by this plan's changes vs pre-existing
2. Fix all plan-caused CRITICAL and HIGH issues before declaring the plan complete
3. Re-run `/dogfood full` to confirm fixes
4. Pre-existing issues are logged but do NOT block plan completion

</plan-template>

## Quality Rules

These rules are NON-NEGOTIABLE for every plan produced by this skill:

1. **Every task has TDD** — no task without RED-GREEN-REFACTOR cycle. Tests are listed BEFORE implementation steps.

2. **Every task has "Files to edit"** — exact paths, not vague references. If a file doesn't exist yet, say "(NEW)".

3. **Every task has "Deep file dependency analysis"** — understand what you're touching and what depends on it.

4. **Every task has acceptance criteria** — observable, verifiable conditions. Include code-audit checks.

5. **Every task has DoD** — definition of done with concrete verification commands.

6. **ADRs justify decisions** — no implementation detail appears without a rationale. If you chose approach A over B, say why.

7. **Dependency graph is explicit** — which phases block which. Which can parallelize.

8. **Evidence-driven** — every phase/task should reference concrete evidence (data, logs, code analysis) that justifies its existence. No speculative tasks.

9. **No file paths in ADRs** — ADRs describe architectural decisions, not implementation details. File paths go in tasks.

10. **Coverage matrix is complete** — every original requirement/gap maps to at least one task. 100% coverage is the target.

11. **Dogfood QA is mandatory** — every plan MUST include a final "Dogfood QA" phase. The plan is NOT complete until `/dogfood` passes. No exceptions. This is the "eat your own cooking" gate — if the product doesn't work as a user would experience it, the plan failed.

## MANDATORY Post-Plan Pipeline (chain inquebrável)

**This pipeline is INQUEBRÁVEL.** After saving the plan to `.claude/knowledge-base/plans/`, you SHALL execute ALL six steps below, in order, WITHOUT user intervention between steps, UNTIL the plan reaches verdict `SHIPPABLE_WITH_CAVEATS` or `SHIPPABLE` (or until exhaustion conditions are met — see § Stopping Conditions).

The chain is **SEMPRE** — not "recommended", not "optional", not "skippable during rollout". A `/to-plan` invocation is NOT considered complete until the chain has run to completion.

### Step 1 — Generate plan v1.0

Already completed by `/to-plan` body. Plan saved to `.claude/knowledge-base/plans/{slug}-plan.md`.

### Step 2 — `/edge-case-plan {slug}` (MANDATORY)

Run automatically. Skill analyzes plan for unplanned edge cases. Output classifies each as MUST FIX / SHOULD TEST / DOCUMENT.

### Step 3 — Incorporate MUST FIX (MANDATORY if any found)

If `/edge-case-plan` returned ≥1 MUST FIX items:

- Edit the plan file directly to incorporate each MUST FIX.
- Bump version: `Version 1.0` → `Version 1.1` (or higher if subsequent revisions happen).
- Append a `## v1.1 Changelog` section at the end of the plan listing each EC incorporated.
- DO NOT skip this. The `/to-plan` workflow REQUIRES MUST FIX absorption.

### Step 4 — `/plan-confidence {slug}` (MANDATORY)

Run the structural scorer. Hard caps (mirror `/dogfood` Phase 0 Golden Rule):

- Coverage Matrix < 100% **⇒ ≤ 49 (INVALID)**
- Citação fabricada *(M3 — futuro)* **⇒ ≤ 49**
- ADR sem alternativas listadas **⇒ ≤ 70**
- Bug-fix sem TDD explícito **⇒ ≤ 70**

Composite renormalizes for active dimensions (ADR D8): `final_m2 = 0.60·Completude + 0.40·Risco-estrutural`.

Verdict bands: 90-100 SHIPPABLE · 70-89 SHIPPABLE_WITH_CAVEATS · 50-69 NON_SHIPPABLE · 0-49 INVALID.

### Step 5 — `/plan-improve {slug}` (MANDATORY if verdict < SHIPPABLE_WITH_CAVEATS)

If Step 4 returned verdict ∈ {`NON_SHIPPABLE`, `INVALID`}, automatically invoke:

```
/plan-improve {slug}
```

This activates the ralph-loop-style auto-improvement loop:

- Phase A (deterministic, $0): weak imperatives → must, loopholes removed, TDD template injected in bug-fix tasks
- Phase B (LLM-driven): ADR alternatives written from context, OR `<!-- TODO -->` comments left for ambiguous cases (NUNCA fabrica)
- Loop iterates with default `max-iterations=20` until verdict reaches `SHIPPABLE_WITH_CAVEATS`+ OR no-improvement-detected OR max iterations reached

### Step 6 — `/plan-confidence {slug}` re-score (MANDATORY)

After `/plan-improve` finishes, re-run `/plan-confidence` to record the FINAL verdict. The re-score is part of the audit trail and proves the loop did what it claimed.

## Pipeline contract (the chain MUST be honored)

```
/to-plan
  → /edge-case-plan
    → incorporate MUST FIX (v1.0 → v1.1)
      → /plan-confidence
        → if verdict < SHIPPABLE_WITH_CAVEATS:
            → /plan-improve
              → /plan-confidence (re-score)
        → if verdict ≥ SHIPPABLE_WITH_CAVEATS: DONE
```

This chain is **inquebrável**. The `/to-plan` invocation is NOT complete until:

1. Plan is saved.
2. Edge case review ran and MUST FIX items are absorbed.
3. `/plan-confidence` has been invoked at least once (post-MUST-FIX).
4. If verdict was below target, `/plan-improve` ran AND `/plan-confidence` re-ran.
5. Final verdict is reported to the user.

## Stopping conditions

The chain STOPS (and reports outcome) when ANY of these are TRUE:

| Condition | Outcome |
|---|---|
| `/plan-confidence` verdict ∈ {SHIPPABLE_WITH_CAVEATS, SHIPPABLE} | SUCCESS — chain complete, ready for implementation |
| `/plan-improve` reached `--max-iterations` (default 20) | EXHAUSTED — report remaining issues honestly to user |
| `/plan-improve` detected no-improvement for 2 consecutive iterations | STUCK — report which issues blocked progress |
| Hard cap that can ONLY be fixed by human (e.g., unmapped Coverage Matrix gap requires creating new task) | NEEDS_HUMAN — leave TODO comments, report exit |

If the chain exits without reaching `SHIPPABLE_WITH_CAVEATS`+, the agent MUST present the user:

- Final verdict + score
- List of issues that could not be auto-fixed
- Recommended next steps (e.g., "Add task for unmapped gap #5; the loop cannot create tasks intelligently")
- Do NOT claim the plan is ready for implementation if verdict < SHIPPABLE_WITH_CAVEATS

## What the user sees

A single invocation of `/to-plan` produces:

1. **The plan file** at `.claude/knowledge-base/plans/{slug}-plan.md` (final version after MUST FIX + auto-improvement)
2. **An audit trail** showing what each step did:
   - `/edge-case-plan` found N MUST FIX (list them)
   - MUST FIX incorporated (v1.0 → v1.1)
   - `/plan-confidence` initial verdict: X (score N)
   - `/plan-improve` applied M changes across K iterations
   - `/plan-confidence` final verdict: Y (score N+Δ)
3. **The next instruction** — typically "Plan is ready for implementation" OR "Plan needs human intervention before implementation: [list]"

The user invokes `/to-plan` ONCE and receives a FULLY-VALIDATED plan with quantitative score.

### Why this chain is inquebrável

- **Edge cases caught BEFORE implementation** save days of rework.
- **Quantitative score** removes ambiguity ("é um bom plano?") with a concrete band + motivos.
- **Auto-improvement** addresses the predictable 80% of issues (weak imperatives, loopholes, missing TDD) so humans focus on the 20% that need judgment.
- **Re-score after improvement** is the proof the loop worked — without it, the "improved" verdict is unverified.

### Reference docs

- Skill: `.claude/skills/plan-confidence/SKILL.md`
- Skill: `.claude/skills/plan-improve/SKILL.md`
- Golden rule: `.claude/rules/plan-confidence-golden-rule.md`
- SOTA report: `.claude/knowledge-base/reviews/plan-confidence-sota-2026-05-17.md`
- ADR D6 (chain optional) **SUPERSEDED** by this section — chain is now MANDATORY, not optional.

## Post-Implementation: Cross-Validation (BEFORE dogfood)

After implementing all phases, BEFORE running `/dogfood`, run the cross-validation:

```
/cross-validation {slug}
```

This is the **most rigorous gate** in the pipeline. It reads the plan line by line and cross-references every task, ADR, TDD cycle, acceptance criterion, and DoD item against the actual code. Divergences are classified by severity (BLOCKER/CRITICAL/MAJOR/MINOR/INFO).

- **APROVADO** → proceed to `/dogfood`
- **REPROVADO** → fix divergences, then re-run `/cross-validation {slug}`
- **APROVADO COM RESSALVAS** → fix CRITICALs, then proceed to `/dogfood`

Report is saved to `.claude/knowledge-base/reviews/cross-validation/{slug}-xval-{YYYY-MM-DD}.md`.

## Post-Implementation: Architecture Diff (AFTER)

When the plan implementation is COMPLETE (all phases done, cross-validation passed, dogfood passed), run `/architecture-docs {domain}` again but output to the **diff** directory:

```
.claude/knowledge-base/architecture/{domain}/diff/
├── system-context.md
├── container-diagram.md
├── component-*.md
└── deep-dive.md
```

This captures the NEW architecture state after the changes. Then **ask the user**:

> "A implementação alterou a arquitetura do domínio `{domain}`. Os novos diagramas estão em `.claude/knowledge-base/architecture/{domain}/diff/`. Posso substituir os documentos principais em `.claude/knowledge-base/architecture/{domain}/` com a versão atualizada?"

- If **YES** → replace the main docs with the diff version, then delete the `diff/` directory.
- If **NO** → keep the diff for reference, do not touch the main docs.

This ensures architecture docs are always accurate and changes are explicitly approved.
