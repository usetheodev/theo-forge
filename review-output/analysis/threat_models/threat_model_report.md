# Threat Model Report — theo-forge SDK
**Date:** 2026-05-20
**Methodology:** STRIDE per flow (Spoofing, Tampering, Repudiation, Information Disclosure,
Denial of Service, Elevation of Privilege)
**Scope:** Five critical flows registered in Phase 1

---

## Trust Boundaries

```
[SDK Consumer Code]
        |
        | (in-process calls)
        v
[theo-forge SDK]  ──── serialize/  ──→ [Filesystem]
        |               expr/
        |               validate/
        |               config/ (GlobalConfig singleton)
        |
        | (HTTPS)
        v
[Argo Workflows Server API]
        |
        | (k8s API)
        v
[Kubernetes Cluster]
```

Key trust boundaries:
- TB1: Consumer code → SDK (in-process; SDK trusts consumer inputs)
- TB2: SDK client → Argo API (network; Bearer token authenticates SDK)
- TB3: SDK → Filesystem (WorkflowToFile / WorkflowFromFile)
- TB4: Argo API → Kubernetes (out of scope for this SDK review)

---

## Flow 1: build-workflow-to-yaml

**Description:** Consumer builds a `Workflow` struct, calls `Build()` + `ToYAML()` or `ToFile()`.
No network involvement. Pure in-memory → file/string transformation.

**Assets:** Generated YAML content, target filesystem location (if ToFile is used).

### STRIDE Analysis

| Threat | Vector | Existing Controls | Missing Controls | Risk |
|---|---|---|---|---|
| **Tampering** | GlobalConfig hooks mutate TemplateModel before serialization; any registered hook can arbitrarily alter the workflow | Hooks are registered explicitly in consumer code | No hook sandboxing; hook functions have write access to entire model | MEDIUM |
| **Tampering** | `PodSpecPatch` accepts arbitrary JSON/YAML string injected into template — no validation | None | No JSON/YAML schema validation on PodSpecPatch | MEDIUM |
| **Information Disclosure** | If consumer calls `ToFile()` with a workflow Name derived from user input, path traversal can write YAML to arbitrary filesystem locations | None | No containment check after `filepath.Join` | HIGH |
| **Denial of Service** | Consumer constructs a DAG with circular dependency expressions (Depends string) | None — Depends is a raw string | No cycle detection on raw Depends strings | LOW |

**Missing Controls:**
- Path containment check in `WorkflowToFile` (SEC-001)
- PodSpecPatch schema validation
- Hook result validation (hooks run as black boxes)

---

## Flow 2: client-submit-lifecycle

**Description:** Consumer creates a `WorkflowsService`, calls `CreateWorkflow()` or lifecycle
methods (`Stop`, `Terminate`, `Suspend`, `Resume`). Involves: Build() → JSON marshal →
HTTPS POST/PUT to Argo API with Bearer token.

**Assets:** Bearer token, workflow model (intellectual property), Argo server availability.

### STRIDE Analysis

| Threat | Vector | Existing Controls | Missing Controls | Risk |
|---|---|---|---|---|
| **Spoofing** | MITM between SDK and Argo server — attacker intercepts the Bearer token | Go default TLS verification (accidental — VerifySSL field is not wired) | VerifySSL is not wired to http.Transport; consumer cannot enable InsecureSkipVerify but also cannot confirm verification is actually enforced | MEDIUM |
| **Tampering** | MITM modifies workflow submitted or response received | TLS (accidental, see above) | No response signing or integrity check | MEDIUM |
| **Repudiation** | No request-level logging in the SDK; caller cannot audit what was submitted | None | No audit trail of submitted workflow payloads | LOW |
| **Information Disclosure** | Bearer token exposed in struct-dump logs | None | No Stringer/MarshalJSON redaction on Token field | MEDIUM |
| **Information Disclosure** | URL path injection via unescaped namespace/name → could hit unintended API endpoints | None | No url.PathEscape on namespace/name | MEDIUM |
| **Denial of Service** | Malicious Argo server returns unbounded response body; SDK reads all into memory | 30s HTTP timeout (prevents slow-loris on request body) | No io.LimitReader on response body; memory exhaustion possible | MEDIUM |
| **Elevation of Privilege** | Consumer sets Host to internal k8s service address (SSRF-as-a-design) | None — by design, consumers control Host | No allowlist for Host; if SDK is used in a multi-tenant pipeline, one tenant could target another tenant's Argo instance | LOW (by-design for SDK) |

**Toxic Combination TC-1:** VerifySSL not wired (SEC-004) + Unbounded response body (SEC-005)
→ If a consumer creates a custom `HTTPClient` with `InsecureSkipVerify=true` (workaround for
SEC-004), a MITM attacker can return an arbitrarily large response body, causing OOM in the
consumer's process. Together these two findings create a DoS vector that neither finding alone
enables.

---

## Flow 3: workflowtemplate-roundtrip

**Description:** Consumer builds a `WorkflowTemplate`, serializes to YAML, submits via
kubectl/ArgoCD, later round-trips back via `FromYAML()`.

**Assets:** Workflow template definition, secret references embedded in templates.

### STRIDE Analysis

| Threat | Vector | Existing Controls | Missing Controls | Risk |
|---|---|---|---|---|
| **Tampering** | `FromYAML` deserializes untrusted YAML from file/network; malformed YAML could corrupt model state | sigs.k8s.io/yaml v1.6.0 uses go.yaml.in/yaml/v3 (alias limits) | No size limit on input string; no schema validation on deserialized model | LOW |
| **Information Disclosure** | `SecretEnv.SecretName` / `SecretVolume.SecretName` accept arbitrary k8s secret names without namespace validation — a secret reference can be crafted to reference secrets outside the intended namespace | None | No validation that secret names follow k8s naming conventions; no namespace-scoped prefix enforcement | MEDIUM |
| **Tampering** | `Container.Hooks` and `DAGTask.Hooks` accept arbitrary `Expression` strings (LifecycleHook.Expression) — these are Argo expressions evaluated by the Argo engine; no validation by SDK | None | No expression syntax validation before serialization | LOW |

---

## Flow 4: dag-dependency-resolution

**Description:** Consumer creates `Task` nodes, chains them with `.Then()`, `.OnSuccess()`,
`.OnError()` or raw `Depends` strings. SDK serializes the dependency graph into YAML.

**Assets:** Correctness of workflow execution order; if dependencies are wrong, unexpected
template execution occurs.

### STRIDE Analysis

| Threat | Vector | Existing Controls | Missing Controls | Risk |
|---|---|---|---|---|
| **Tampering** | Raw `Depends` string accepts arbitrary content including operators that may confuse the Argo parser; SDK passes it verbatim | None | No validation of Depends string against known operators and task names | LOW |
| **Tampering** | Task `When` field accepts arbitrary Argo expression — same injection risk as LifecycleHook.Expression | None | No expression validation | LOW |
| **Elevation of Privilege** | If a consumer uses `expr.C(userControlledString)` to build a When condition, single-quote injection can alter condition logic — bypassing intended execution guards | None | No escaping in C() or string methods (SEC-002) | HIGH |

**Toxic Combination TC-2:** expr.C() single-quote injection (SEC-002) + Task.When field →
An SDK consumer who builds workflow conditions from user-controlled inputs (e.g., environment
name, branch name in CI context) can have the `When` condition manipulated by a crafted input
value, causing unintended tasks to execute in the Argo engine. The SDK generates the malformed
YAML, Argo evaluates it, and a wrong branch executes in the cluster. This is the highest-risk
combination because it crosses the SDK→Argo trust boundary with attacker-influenced content.

---

## Flow 5: global-config-hook-dispatch

**Description:** Consumer registers hooks via `config.GlobalConfig.RegisterTemplateHook()` or
`RegisterWorkflowHook()`. On every `Build()` call, `DispatchWorkflowHooks` and
`DispatchTemplateHooks` invoke all hooks with a pointer to the model.

**Assets:** Workflow/template model integrity; GlobalConfig state (Host, Token, Namespace).

### STRIDE Analysis

| Threat | Vector | Existing Controls | Missing Controls | Risk |
|---|---|---|---|---|
| **Tampering** | Hook functions receive `*model.TemplateModel` — a hook can zero out fields, inject arbitrary values, or modify security-relevant fields (ServiceAccountName, SecurityContext, Tolerations) | Hooks are registered by consumer code (first-party) | In multi-library scenarios (SDK used alongside other libs that also call RegisterTemplateHook), a third-party hook can tamper with templates built by the consumer | MEDIUM |
| **Information Disclosure** | GlobalConfig.Token is a public field — any code in the same process with access to the singleton can read it | None | No access control; in a shared-library context, Token is effectively global state | MEDIUM |
| **Repudiation** | Hooks modify models before serialization but after the consumer's explicit Build() call — the consumer cannot easily distinguish their template from the hook-modified version in logs | None | No before/after hook logging, no hook audit trail | LOW |
| **Denial of Service** | A panic inside a hook propagates up through DispatchTemplateHooks to the Build() call — if hooks are from third-party code, a misbehaving hook can crash the consumer | None | No recover() wrapper around hook dispatch | LOW |

---

## Toxic Combinations Summary

### TC-1: VerifySSL Contract Violation + Unbounded Response Body
**Findings:** SEC-004 + SEC-005
**Combined risk:** MEDIUM → HIGH (escalation when InsecureSkipVerify workaround is applied)
**Mechanism:** Consumer sets up InsecureSkipVerify transport as workaround for SEC-004.
MITM attacker intercepts TLS and returns unbounded response. SDK reads all bytes into memory
(SEC-005). Result: OOM crash in consumer process.
**Mitigation:** Fix both findings together — implement VerifySSL correctly AND add LimitReader.

### TC-2: expr.C() Single-Quote Injection + Task.When / Condition Fields
**Findings:** SEC-002 + unvalidated When/expression fields
**Combined risk:** HIGH (cross-trust-boundary exploit)
**Mechanism:** Consumer derives a string constant from user input (e.g., git branch name
`main' || 'true`) and passes it to `C()`. The resulting expression `'main' || 'true'` in a
`When` field causes a conditional task to execute unconditionally in the Argo engine. This
bypasses workflow execution guards without the consumer's knowledge.
**Mitigation:** Fix SEC-002 (escape single quotes in C() and all string methods).

### TC-3: Path Traversal + Workflow Name From User Input
**Findings:** SEC-001 + no workflow name validation beyond length
**Combined risk:** HIGH
**Mechanism:** Consumer accepts workflow name from API input (e.g., `wfName = "../../cron.d/evil"`).
Consumer calls `workflow.ToFile(outputDir, "")` — fileName is auto-derived from wfName.
Result: YAML written to `/etc/cron.d/evil.yaml` instead of `outputDir/evil.yaml`.
**Mitigation:** Fix SEC-001 (containment check) AND validate workflow names against k8s
naming conventions (alphanumeric + hyphens, no path separators).

---

## Risk Heat Map

```
            Likelihood
            Low     Medium  High
Severity
High    | TC-3   | TC-2   | -
Medium  | TC-1   | SEC-003| SEC-004
Low     | SEC-007| SEC-005| -
```

---

## Recommendations Priority Order

1. **Fix SEC-001** (path traversal) — containment check is a two-line fix, high impact
2. **Fix SEC-002** (single-quote injection) — escaping in C() and string methods, medium effort
3. **Fix SEC-003** (URL path injection) — add url.PathEscape to all path params, low effort
4. **Fix SEC-004** (VerifySSL not wired) — decide: implement properly or remove the field
5. **Fix SEC-005** (unbounded response) — wrap with io.LimitReader, one-line fix
6. **Fix SEC-006** (token redaction) — implement String() on WorkflowsService and GlobalConfig
7. **Fix SEC-007** (Token field exported) — make unexported, add GetToken() accessor
