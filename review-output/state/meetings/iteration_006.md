# Meeting Minutes — Phase 6, Iteration 1 (Global Iteration 6)
**Date:** 2026-05-20

## Status
- Phase: 6/7 (Security + Threat Modeling)
- Findings entering phase: 50 (9 high, 25 medium, 16 low)
- New findings this iteration: 7 (2 high, 4 medium, 1 low)
- Findings after phase: 57 (11 high, 29 medium, 17 low)
- Previous iteration summary: Phase 5 covered infrastructure (CI/CD), supply chain (go.mod,
  action pinning), and data quality (golden files). 12 new findings registered.

## Agent Reports

### Security Auditor

**Secrets scan:** Clean. No AKIA/ghp_/sk- patterns, no private keys, no hardcoded passwords
in production code. Test files use obviously-fake "test-token" placeholders.

**YAML injection / expr package:**
- `expr.C(string)` wraps values in single quotes with NO escaping — `C("it's")` produces the
  malformed expression `'it's'`. All string methods (Contains, Matches, StartsWith, EndsWith,
  Sprig.*) share the same unescaped `'%s'` pattern. This is SEC-002 (HIGH).
- `expr.E(s)` takes a raw string — by design, no escaping intended (it's a "raw expression"
  constructor). No issue, correctly documented as raw.
- Argo template interpolation (`{{...}}`) does not go through the SDK — strings are passed
  as literal YAML fields. No template injection at SDK level.

**Path traversal:**
- `serialize.WorkflowToFile` has a confirmed path traversal: `filepath.Join(absDir, fileName)`
  resolves `..` components (verified via Go test). There is NO containment check. This is
  SEC-001 (HIGH). The auto-derived `fileName` from `wfName` is equally unvalidated.

**HTTP client security:**
- No `tls.Config` is set anywhere — `VerifySSL` field is declared in two places (client,
  config) but is never read. TLS is verified by Go default (accidental). SEC-004 (MEDIUM).
- No `io.LimitReader` — response body read is unbounded. SEC-005 (MEDIUM).
- No `url.PathEscape` on namespace/name parameters — path injection confirmed. SEC-003 (MEDIUM).
- Response body is properly closed via `defer resp.Body.Close()` — no socket leak.
- 30-second timeout is set on the `http.Client` — DoS via slow-loris on requests is mitigated.
- Context is threaded through all requests via `http.NewRequestWithContext` — good.

**Sensitive logging:**
- No logging code exists in any SDK package (client/, config/, serialize/, expr/). Zero fmt.Print
  or log.* calls in production code. No Authorization header logging.
- However, exported Token fields (client.WorkflowsService.Token, config.GlobalConfig.Token)
  have no Stringer redaction — struct-dump logging by the consumer will expose the token.
  SEC-006 (MEDIUM), SEC-007 (LOW).

**Token / identity:**
- Token loaded as plain string via constructor or public field — no env-var helper, no file
  loader, no rotation mechanism. Acceptable for an SDK (consumer's responsibility).
- GlobalConfig.Token is a public field — bypasses mutex protection when written directly.
  SEC-007 (LOW).

**Secret references in YAML:**
- SecretEnv and SecretVolume accept arbitrary secret names and keys with no k8s naming
  validation. Cross-namespace secret references cannot be made via this SDK (k8s RBAC enforces
  that), but malformed secret names would cause k8s API errors at submit time rather than
  build time. No new finding warranted — this is already captured in Phase 4 completeness work.

### Threat Modeler

**5 flows modeled:**
1. build-workflow-to-yaml — path traversal (TC-3), hook tampering, PodSpecPatch injection
2. client-submit-lifecycle — MITM/VerifySSL (TC-1), unbounded body, token disclosure, URL injection
3. workflowtemplate-roundtrip — secret name validation gap, LifecycleHook.Expression injection
4. dag-dependency-resolution — expr.C() injection + Task.When (TC-2, highest combined risk)
5. global-config-hook-dispatch — hook tampering, token as global state, panic propagation

**3 toxic combinations identified:**
- TC-1: SEC-004 + SEC-005 → MITM + OOM escalation (MEDIUM→HIGH when workaround applied)
- TC-2: SEC-002 + When field → single-quote injection bypasses execution guards (HIGH)
- TC-3: SEC-001 + workflow name from user input → arbitrary file write (HIGH)

## Strategic Assessment

**What worked well:**
- The targeted investigation approach was highly effective — starting with the known
  "VerifySSL not wired" finding and tracing all related surfaces (token, TLS, body reads)
  in one session produced a cohesive set of HTTP security findings.
- The path traversal finding was confirmed with a Go test (not assumed) — increases confidence.
- Single-quote injection in expr is a genuine finding — the SDK doc shows expr.C() being used
  with string values, and the lack of escaping is real.

**What didn't work:**
- Could not run gosec dynamically against the codebase — static reading only.
- No SBOM analysis was possible (go list -m -u all unavailable for vulnerability check).

**Key insight:**
The SDK has a clean secret hygiene record but has a cluster of injection/traversal issues in
the "path building" layer — URL paths in client, filesystem paths in serialize, and expression
string interpolation in expr. These share a common root cause: raw user-controlled strings are
used in construction without encoding.

**Cross-cutting concern:**
TC-2 (expr injection + When field) is the most important combined finding because it crosses
the SDK→Argo trust boundary. The expr package is designed for SDK consumers to build safe
expressions, but the lack of escaping makes it unsafe when inputs come from external sources.

## Decisions

1. **Register 7 new security findings** (SEC-001 through SEC-007) — DONE
2. **Register 3 toxic combinations** as separate threat model entries — DONE
3. **Severity calibration:**
   - SEC-001 (path traversal): HIGH — can write to arbitrary filesystem locations
   - SEC-002 (expr injection): HIGH — crosses trust boundary into Argo engine evaluation
   - SEC-003 (URL injection): MEDIUM — requires namespace/name from untrusted source
   - SEC-004 (VerifySSL): MEDIUM — broken contract; current behavior is secure but misleading
   - SEC-005 (unbounded body): MEDIUM — DoS potential, esp. in combination with TC-1
   - SEC-006 (token redaction): MEDIUM — common logging exposure pattern
   - SEC-007 (Token field public): LOW — in-process, by-design for Go SDK
4. **Phase 6 is complete** — all 5 flows modeled, all OWASP top-10 categories evaluated
5. **Advance to Phase 7** — dogfooding and E2E validation

## Loop-Back Assessment
- Should we loop back? NO
- Rationale: Phase 6 found 7 new findings (2 HIGH), which is a solid yield but not a signal
  of a large blind spot. The supply chain issues (gosec@master, action pinning) were already
  captured in Phase 5. No major subsystem was missed. The findings are well-distributed across
  the client, expr, and serialize packages. Phase 7 should validate the top-priority findings
  (SEC-001, SEC-002, SEC-003) and assess whether they are reproducible and fixable.

## Task Assignments for Phase 7

- **Dogfood Tester:** Attempt to reproduce SEC-001 (path traversal) and SEC-002 (single-quote
  injection) using the SDK's public API. Confirm reproducibility and assess fixability.
- **Test Auditor:** Check if any existing tests cover the SEC-001 and SEC-002 scenarios.
  If not, note the testing gap as an additional finding.
- **Observability Reviewer:** Assess whether the lack of logging in the SDK (good for security)
  creates observability blind spots for consumers who need to audit what was submitted.

## Next Meeting
- Expected at: Phase 7, Iteration 1 (Global Iteration 7)
- Focus: Dogfooding validation of security findings; E2E assessment of top recommendations;
  final loop-back decision based on Phase 7 results
