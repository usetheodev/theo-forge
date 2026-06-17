# Edge Case Review — fix-all-review-findings-plan

Data: 2026-05-20
Tasks analisadas: 41 (10 phases)
Edge cases encontrados: 21 (MUST FIX: 7, SHOULD TEST: 9, DOCUMENT: 5)

Scope: pragmatic edge-case sweep across boundaries (INPUTS, STATE, I/O, CONCURRENCY, INTEGRATION). The plan already covers every finding in `review-output/review.db`; this report flags risks the planned fixes do NOT yet address.

---

## MUST FIX

### EC-1: `config.GlobalConfig` mutation races during concurrent `go test ./...`
- **Task afetada:** T0.1 (`resetGlobalConfigForTest`) + T3.1 (resolver)
- **Família:** Concurrency / State
- **Cenário:** `go test ./...` runs packages in parallel by default. `resetGlobalConfigForTest` calls `config.SetGlobalConfig(NewConfig())` in `t.Cleanup`, but `helpers.go` (after T3.1) still uses `config.GetGlobal()` as fallback. Two parallel test binaries sharing the package singleton AND the runtime hook map will see torn writes / racing dispatchers.
- **Impacto:** Flaky tests; `-race` flag intermittently fires; hook test in pkg A leaks state into pkg B.
- **Fix sugerido:** In T0.1 add `t.Parallel()` ban in the helper godoc AND have `resetGlobalConfigForTest` acquire a process-wide `sync.Mutex` exposed as `config.TestLock` (test-only). Tests touching global state must `defer config.TestLock.Unlock()`. Plan must require: ANY task touching `GlobalConfig` uses this lock.

### EC-2: README example blocks contaminate each other via package-level state
- **Task afetada:** T1.1 (`docs/readme_examples_test.go`)
- **Família:** State / Integration
- **Cenário:** If two extracted blocks both call `forge.SetImage("x")` (package singleton mutation), execution order of `TestReadme_*` determines which test sees what. Worse: a block mutates `config.GlobalConfig` and a later block asserts the default → fails non-deterministically.
- **Impacto:** README test green locally, red in CI under `-shuffle=on`; or vice versa.
- **Fix sugerido:** Each `TestReadme_*` function MUST start with `resetGlobalConfigForTest(t)` (helper from T0.1). Add this as an explicit acceptance criterion in T1.1.

### EC-3: `containedJoin` accepts symlinked `dir` whose target escapes
- **Task afetada:** T2.1 (containedJoin)
- **Família:** Boundary / Permission
- **Cenário:** Caller passes `dir = "/var/tmp/safe"` where `/var/tmp/safe -> /etc`. `filepath.Abs` does not resolve symlinks; `filepath.Rel(absDir, joined)` returns a clean relative path. Final write goes to `/etc/filename`.
- **Impacto:** Path-traversal mitigation bypassed via attacker-controlled directory. Closes finding only partially.
- **Fix sugerido:** After `filepath.Abs`, call `filepath.EvalSymlinks(absDir)` and use the resolved root for the `Rel` check. Document that this rejects paths under deleted/missing dirs (acceptable for `ToFile`).

### EC-4: `argoEscape` silently double-escapes consumer-pre-escaped strings
- **Task afetada:** T2.2 (argoEscape)
- **Família:** Format / Backward-compat
- **Cenário:** A user upgrading from v0.4.x has `expr.C("it''s")` in their code (manually escaped). New version produces `'it''''s'`, which Argo parses as `it''s` literally — surprise behavior change, no error.
- **Impacto:** Silent semantic drift; existing workflows break in production after SDK upgrade.
- **Fix sugerido:** Add a CHANGELOG-flagged BREAKING note AND an `expr.RawC()` migration example. Add a `TestExprC_DoubleEscapesPreEscapedInput` test that LOCKS this behavior so consumers see the failure-mode contract.

### EC-5: `VerifySSL=true` regression breaks consumers who relied on the bug
- **Task afetada:** T2.4 (Wire VerifySSL)
- **Família:** Format / Backward-compat
- **Cenário:** Today the field is dead — any user with self-signed Argo whose code sets `VerifySSL=true` is actually "working" because Go's stdlib default verifies and FAILS. So they must already have `VerifySSL=false` OR a custom transport. BUT: any user who passes a custom `*http.Client` via a hypothetical field (or who set `InsecureSkipVerify` upstream) will now have their transport overwritten by the lazy `httpClient()` builder.
- **Impacto:** Custom transports silently replaced; mTLS / corporate proxy configs break.
- **Fix sugerido:** `httpClient()` MUST honor a non-nil `s.client` set externally (the deep-dive snippet already does `if s.client != nil { return s.client }` — make this an explicit, tested AC). Add `TestWorkflowsService_HonorsExternalHTTPClient`.

### EC-6: `json.Marshal(svc)` returning `Token: "***"` causes data loss on round-trip
- **Task afetada:** T2.6 (token redaction)
- **Família:** Format / State
- **Cenário:** Consumers commonly serialize service config to YAML/JSON for persistence (cache, debugging dump, config sync). With the new `MarshalJSON`, the persisted file contains `"token":"***"`. On next process start, `json.Unmarshal` loads `***` as the literal token → auth fails silently OR succeeds with a server that accepts any string.
- **Impacto:** Data loss; auth outage; or worse — silent privilege escalation.
- **Fix sugerido:** Implement `MarshalJSON` redaction, but add `UnmarshalJSON` that REJECTS the literal `"***"` with `ErrRedactedTokenLoaded` (new sentinel). Document explicitly: "for persistence, use `SetToken` after load; never round-trip the struct."

### EC-7: Dependabot auto-merge on SHA-pin bump can ship broken CI
- **Task afetada:** T5.1, T5.2 (SHA-pin + Dependabot)
- **Família:** Integration / State
- **Cenário:** Plan adds `dependabot.yml`. If no review policy is set, default GitHub repo settings may allow Dependabot PRs to be auto-merged once required checks pass. A malicious or buggy SHA bump (e.g., `gosec` regression) flips green and lands on `main`.
- **Impacto:** Supply-chain mitigation undone by automation; the very risk the SHA pin defends against re-enters via the bump pipeline.
- **Fix sugerido:** T5.2 AC must include: "branch protection on `main` requires 1 human reviewer AND CODEOWNERS approval on `.github/**` changes; Dependabot PRs disabled from auto-merge for `github-actions` ecosystem." Add explicit task T5.2.b to configure branch protection via `gh api`.

---

## SHOULD TEST

### EC-8: `containedJoin` on Windows / case-insensitive filesystems
- **Task afetada:** T2.1
- **Teste sugerido:** `TestContainedJoin_CaseInsensitiveEscape` — assert that on `GOOS=windows` or APFS, `dir="/tmp/Safe"` + `name="../safe/escape"` is correctly classified (use `filepath.EvalSymlinks` semantics). Add a build-tagged test for `windows` if CI allows.

### EC-9: `argoEscape` and Unicode normalization
- **Task afetada:** T2.2
- **Teste sugerido:** `TestArgoEscape_UnicodeApostrophe` — assert that U+2019 (right single quotation mark `’`) is NOT escaped (it's not the ASCII `'` Argo cares about), and that combining marks pass through unchanged. Documents the contract: only ASCII `'` is escaped.

### EC-10: URL-escaped k8s names that Argo's API server rejects
- **Task afetada:** T2.3
- **Teste sugerido:** `TestClient_EscapedNameMatchesArgoSpec` — table test confirming that valid k8s names (per regex in `validateK8sName`) round-trip through `url.PathEscape` to bytes Argo accepts (i.e., no `.` becomes `%2E`). `url.PathEscape` already preserves `.` and `-`; lock this with an explicit assertion.

### EC-11: `LimitReader` on legitimate huge workflow log responses
- **Task afetada:** T2.5
- **Teste sugerido:** `TestClient_RejectsTrulyOversizedBody` AND `TestClient_AcceptsRealisticArgoListResponse` — second test should pull the largest fixture from `testdata/` (or a 5-MB synthetic listing) and confirm 32 MiB is comfortable. If any production response is known >32 MiB (e.g., logs metadata), DOCUMENT the constant rationale in client.go and surface in DoD.

### EC-12: `BaseTemplate` field-shadowing ambiguity in embedding
- **Task afetada:** T4.1
- **Teste sugerido:** `TestBaseTemplate_NoShadowing_Container` AND `_Script` — use `go vet` + a unit test that constructs `Container{BaseTemplate:BaseTemplate{Name:"a"}, /* no top-level Name */}` and asserts `c.Name == "a"`. Add a build-time linter or doc note: "Container and Script must NOT redeclare any BaseTemplate field."

### EC-13: `IntOrString` migration — compile-time hint for `int(50)` callers
- **Task afetada:** T4.3
- **Teste sugerido:** N/A (compile-time). Instead, add an `examples_test.go` entry `ExampleIntOrStringFromInt` that pkg.go.dev surfaces. AC update: CHANGELOG migration note shows BEFORE/AFTER code side-by-side.

### EC-14: `Logger` panic kills the HTTP request
- **Task afetada:** T6.1
- **Teste sugerido:** `TestWorkflowsService_LoggerPanic_DoesNotBreakRequest` — inject a logger whose `Debug` panics; assert the HTTP request still completes and panic is recovered + logged via stderr fallback. Implementation: wrap every logger call in `defer func() { recover() }()`.

### EC-15: `Task.Then` "always parens" produces noisy YAML on common case
- **Task afetada:** T3.9
- **Teste sugerido:** `TestTaskThen_ParensOnSingleDep_RoundTrip` — assert `a.Then(b)` where `a.Depends="x"` produces `(x) && b` and that Argo still accepts it. If readability becomes a complaint, fall back to the detection variant.

### EC-16: `HTTPTemplate.Timeout` ambiguity — "30" vs "30s"
- **Task afetada:** T3.10
- **Teste sugerido:** `TestHTTPTemplate_Timeout_RejectsBareNumber` — `time.ParseDuration("30")` returns an error; lock that bare ints are rejected with `ErrInvalidTimeout` and the error message tells the user to use `"30s"`.

---

## DOCUMENT

### EC-17: `T3.3` mutual exclusion breaks users who set both fields by accident today
- **Risco aceito:** Pre-fix, Argo silently uses one. Post-fix, SDK rejects with `ErrTemplateAmbiguous`. This is correct behavior; flag in CHANGELOG as BREAKING with explicit example. No code change beyond CHANGELOG entry already planned.

### EC-18: `T4.5` unexporting `Token` breaks `cfg.Token` direct reads
- **Risco aceito:** T4.5 already flags BREAKING in CHANGELOG. Plan should ADD a brief migration shim section in CHANGELOG showing `cfg.Token` → `cfg.GetToken()`. No code shim required (would defeat the purpose of unexporting).

### EC-19: `T2.5` `maxResponseBodyBytes` is a hard-coded constant
- **Risco aceito:** 32 MiB is reasonable today; future workloads may need tuning. Document in client godoc: "If 32 MiB is insufficient, file an issue with the response size observed." Do NOT add a config option yet (YAGNI).

### EC-20: `expr.RawC` is a foot-gun
- **Risco aceito:** Doc-warned in ADR-005. Trust the godoc warning; do not add runtime restrictions. If misuse becomes a pattern, revisit in v0.6.0.

### EC-21: Phase 0 ordering assumes no two phases write to the same `*_test.go` simultaneously
- **Risco aceito:** Plan §"Parallelism notes" says Phases 1, 2, 5, 6 can run in parallel. They write to disjoint files per the file inventories. Acceptable; add a one-line caveat in the plan that if any two phase PRs touch the same `_test.go` (e.g., `client/client_test.go` for T2.3, T2.4, T2.5, T2.6, T3.7, T6.1, T7.6 — actually a lot of overlap!), they must merge sequentially. **NB: this overlap is real and worth flagging.**

---

## Resumo

| Task | Edges encontrados | MUST FIX | SHOULD TEST | DOCUMENT |
|------|-------------------|----------|-------------|----------|
| T0.1 | 1 | 1 | 0 | 0 |
| T1.1 | 1 | 1 | 0 | 0 |
| T2.1 | 2 | 1 | 1 | 0 |
| T2.2 | 2 | 1 | 1 | 0 |
| T2.3 | 1 | 0 | 1 | 0 |
| T2.4 | 1 | 1 | 0 | 0 |
| T2.5 | 2 | 0 | 1 | 1 |
| T2.6 | 1 | 1 | 0 | 0 |
| T3.1 | 1 (overlaps T0.1) | (counted in T0.1) | 0 | 0 |
| T3.3 | 1 | 0 | 0 | 1 |
| T3.9 | 1 | 0 | 1 | 0 |
| T3.10 | 1 | 0 | 1 | 0 |
| T4.1 | 1 | 0 | 1 | 0 |
| T4.3 | 1 | 0 | 1 | 0 |
| T4.5 | 1 | 0 | 0 | 1 |
| T5.1+T5.2 | 1 | 1 | 0 | 0 |
| T6.1 | 1 | 0 | 1 | 0 |
| Plan-wide (Phase 0 / parallelism) | 2 | 0 | 0 | 2 |
| **TOTAL** | **21** | **7** | **9** | **5** |

**Veredicto:** PLANO PRECISA DE AJUSTE

Reason: 7 MUST FIX items represent real-world failure modes the planned fixes do not yet cover. None require new modules or architecture — all are 1-3 line additions or new test cases. Plan v1.1 should integrate EC-1 through EC-7 as sub-tasks under the affected T-numbers, then re-run `/plan-confidence`.
