# Security Review — Phase 6
**Date:** 2026-05-20
**Reviewer:** security-auditor
**Scope:** Go SDK (github.com/usetheodev/theo-forge) — not a deployed service

---

## Executive Summary

The SDK has no hardcoded secrets and no logging of sensitive data. The primary security
surface is the `client/` HTTP client and the `expr/`, `serialize/` packages. Seven new
security findings were identified. The most significant are:

1. Path traversal in `serialize.WorkflowToFile` via user-controlled `fileName`
2. Single-quote injection in `expr.C()` and all string-interpolation methods
3. URL path injection in `client.doRequest` via unsanitized namespace/name parameters
4. Broken contract on `VerifySSL` — the field is declared but never wired to `http.Transport`
5. Unbounded response body read — no `io.LimitReader` in `client.doRequest`
6. Exported `Token` field in both `client.WorkflowsService` and `config.GlobalConfig` — no
   Stringer-based redaction, meaning struct-dump logs will expose the token
7. `GlobalConfig.Token` stored as plain string in a public field with no protection

No hardcoded credentials were found (`git grep` scan clean). No logging of Authorization
headers was found. The `sigs.k8s.io/yaml v1.6.0` dependency uses `go.yaml.in/yaml/v3` which
has alias-expansion limits, so YAML bomb via `FromYAML` is mitigated by the library.

---

## Findings

### SEC-001 — Path Traversal in WorkflowToFile via User-Controlled fileName
**Severity:** HIGH
**File:** `serialize/serialize.go:133`
**OWASP:** A05:2021 Security Misconfiguration / CWE-22 Path Traversal

**Description:**
`WorkflowToFile(yamlStr, outputDir, fileName, ...)` calls `filepath.Join(absDir, fileName)`.
Go's `filepath.Join` calls `filepath.Clean` internally, which resolves `..` components.
A caller passing `fileName = "../../etc/cron.d/evil.yaml"` will write outside the intended
`outputDir`. There is no check that the resulting path remains under `absDir`.

**Verified:**
```go
absDir := "/tmp/output"
fileName := "../../../etc/cron.yaml"
path := filepath.Join(absDir, fileName)
// Result: /etc/cron.yaml — OUTSIDE absDir
```

**Attack scenario:** An SDK consumer that derives `fileName` from user input (e.g., workflow
name from an HTTP API) can be tricked into writing a YAML file to an arbitrary filesystem
location.

**Also affects:** The auto-derived `fileName` path: if `wfName = "../../etc/passwd"` and
`fileName` is left empty, the auto-derived name (`"../../etc/passwd.yaml"`) is equally
unvalidated before `filepath.Join`.

**Recommendation:**
After computing `path`, verify containment:
```go
rel, err := filepath.Rel(absDir, path)
if err != nil || strings.HasPrefix(rel, "..") {
    return "", fmt.Errorf("fileName escapes output directory")
}
```

---

### SEC-002 — Single-Quote Injection in expr.C() and String Methods
**Severity:** HIGH
**File:** `expr/expr.go:35`, `expr/expr.go:161–173`, `expr/expr.go:221–234`
**OWASP:** A03:2021 Injection / CWE-74 Neutralization Failure

**Description:**
`C(v interface{})` for string values formats the result as `'%s'` without escaping single
quotes in the input:
```go
case string:
    return Expr{repr: fmt.Sprintf("'%s'", val)}
```
If `val` contains a single quote (e.g., `"it's a run"`), the output `'it's a run'` is a
malformed Argo expression. Worse, a crafted value can escape the string literal and inject
arbitrary expression syntax: `C("' + malicious_expr + '")`.

The same pattern appears in all string-interpolation methods:
- `Contains(s string)` → `%s.contains('%s')`
- `Matches(pattern string)` → `%s.matches('%s')`
- `StartsWith(prefix string)` → `%s.startsWith('%s')`
- `EndsWith(suffix string)` → `%s.endsWith('%s')`
- `Sprig.Trim/Upper/Lower/Replace(s string)` → same unescaped `'%s'` pattern

**Attack scenario:** If an SDK consumer passes user-controlled strings to `C()` or string
methods (e.g., a user-supplied tag value being used as a constant in an Argo expression),
the injected expression is evaluated by the Argo engine, potentially bypassing conditional
execution guards or triggering unintended template execution.

**Recommendation:**
Escape single quotes in string inputs:
```go
case string:
    escaped := strings.ReplaceAll(val, `'`, `\'`)
    return Expr{repr: fmt.Sprintf("'%s'", escaped)}
```
Apply the same escaping to all string interpolation methods.

---

### SEC-003 — URL Path Injection via Unsanitized namespace/name Parameters
**Severity:** MEDIUM
**File:** `client/client.go:149`, `client/client.go:162`, `client/client.go:239–266`
**OWASP:** A03:2021 Injection / CWE-74

**Description:**
All client methods that take `namespace` and `name` parameters construct URL paths by simple
string concatenation:
```go
"/api/v1/workflows/" + namespace + "/" + name
```
Neither parameter is validated or URL-path-encoded. A `namespace` value of `"default/../admin"`
results in the path `/api/v1/workflows/default/../admin/my-wf`, which some HTTP servers will
normalize to `/api/v1/admin/my-wf`. A `name` value containing a query character (`?`) would
inject query parameters.

**Verified:** Go's `http.NewRequestWithContext` does not normalize the path before transmission;
the raw path is sent to the server.

**Attack scenario:** An SDK consumer that uses namespace/name values derived from external
sources (form inputs, CI parameters) could be manipulated to hit different API endpoints on
the Argo server, potentially accessing admin or system endpoints.

**Recommendation:**
Use `url.PathEscape` on the namespace and name segments:
```go
import "net/url"
path := "/api/v1/workflows/" + url.PathEscape(namespace) + "/" + url.PathEscape(name)
```

---

### SEC-004 — VerifySSL Field Not Wired to http.Transport (Broken Contract)
**Severity:** MEDIUM
**File:** `client/client.go:36–48`, `config/config.go:34–46`
**OWASP:** A02:2021 Cryptographic Failures / CWE-295

**Description:**
`WorkflowsService.VerifySSL` and `config.GlobalConfig.VerifySSL` are declared and default
to `true`, but `NewWorkflowsService` creates an `http.Client` with the default `http.Transport`
(no custom TLS config). The `VerifySSL` field value is read nowhere. The SDK's public API
promises TLS control that it does not deliver.

**Current state:**
- `VerifySSL = true` → TLS IS verified (Go default — correct but accidental)
- `VerifySSL = false` → TLS IS STILL verified (field is ignored — silent contract violation)

**Consequence:** A consumer who explicitly sets `VerifySSL = false` to communicate with a
self-signed certificate endpoint will receive an unexpected TLS error, and may work around it
by replacing the `HTTPClient` with an insecure transport — creating a worse security posture
than if the field were properly implemented with an opt-in warning.

**Recommendation:**
Either (a) implement the field correctly using `&tls.Config{InsecureSkipVerify: !s.VerifySSL}`
inside a custom `http.Transport`, with a `//nolint:gosec` comment and documentation warning,
or (b) remove the field from the public API and document that TLS verification always uses
system roots.

---

### SEC-005 — Unbounded Response Body Read in doRequest
**Severity:** MEDIUM
**File:** `client/client.go:91–93`
**OWASP:** A05:2021 Security Misconfiguration / CWE-400 Resource Exhaustion

**Description:**
`doRequest` reads the entire response body with `io.ReadAll(resp.Body)` without any size
limit. A malicious or misbehaving Argo server (including a MITM server if TLS verification
is disabled — see SEC-004) could return an arbitrarily large response body, causing unbounded
memory consumption in the SDK consumer's process.

```go
respBody, err := io.ReadAll(resp.Body)  // no limit
```

**Recommendation:**
Wrap the body reader with `io.LimitReader` before reading:
```go
const maxResponseBody = 32 * 1024 * 1024 // 32 MB
limited := io.LimitReader(resp.Body, maxResponseBody)
respBody, err := io.ReadAll(limited)
```
Consider exposing `MaxResponseBodyBytes` as a configurable field on `WorkflowsService`.

---

### SEC-006 — Exported Token Field Without Redaction
**Severity:** MEDIUM
**File:** `client/client.go:32`, `config/config.go:24`
**OWASP:** A09:2021 Security Logging and Monitoring Failures / CWE-532

**Description:**
`WorkflowsService.Token` and `GlobalConfig.Token` are public `string` fields. Neither type
implements `fmt.Stringer` or `MarshalJSON` to redact the token value. If a consumer logs
the struct (e.g., `log.Printf("%+v", svc)`), the Bearer token is exposed in plain text in
the log output.

This is especially relevant because:
- The `GlobalConfig` singleton may be logged during initialization/debug
- The `WorkflowsService` may be passed to middleware or error handlers that log structs
- Test frameworks that print struct values on failure will expose tokens

**Recommendation:**
Implement `String() string` on both types returning `[REDACTED]` for the Token field, or
change `Token` to a custom `SensitiveString` type with redaction in `String()`.

---

### SEC-007 — GlobalConfig.Token Accessible as Plain Public Field
**Severity:** LOW
**File:** `config/config.go:24`, `config/config.go:103–107`
**OWASP:** A07:2021 Identification and Authentication Failures / CWE-522

**Description:**
`GlobalConfig.Token` is a direct public `string` field, writable without going through
`SetToken()`. Any code with a pointer to the global config (obtainable via `config.GetGlobal()`)
can read or overwrite the token without any access control:
```go
config.GetGlobal().Token = "attacker_controlled_token"
```
While in-process mutation is expected for a Go library (this is not a deployed service),
the pattern means any dependency that receives a reference to the GlobalConfig can silently
replace the authentication credential.

**Recommendation:**
Make `Token` unexported and only accessible through `SetToken` / `GetToken` accessors with
mutex protection. `SetToken` is already implemented; direct field access bypasses the mutex.

---

## Secrets Scan Results

```
git grep -E '(AKIA[0-9A-Z]{16}|ghp_|sk-|-----BEGIN PRIVATE KEY-----|password\s*=\s*"[^"]*")'
```
Result: No hardcoded AWS keys, GitHub tokens, OpenAI keys, private keys, or plaintext
passwords found in the codebase.

Test files use `"test-token"` which is clearly a placeholder value.

---

## OWASP Coverage Summary

| OWASP 2021 | Category | Covered | Finding |
|---|---|---|---|
| A01 | Broken Access Control | Partial | SEC-007 (token field public) |
| A02 | Cryptographic Failures | Yes | SEC-004 (VerifySSL not wired) |
| A03 | Injection | Yes | SEC-001 (path traversal), SEC-002 (expr injection), SEC-003 (URL injection) |
| A04 | Insecure Design | N/A | Not applicable to SDK |
| A05 | Security Misconfiguration | Yes | SEC-001, SEC-005 (unbounded read) |
| A06 | Vulnerable/Outdated Components | Prior | gosec@master (Phase 5 INFRA findings) |
| A07 | Auth/Identity Failures | Yes | SEC-006, SEC-007 |
| A08 | Software/Data Integrity | Prior | Action SHA pinning (Phase 5) |
| A09 | Logging/Monitoring Failures | Yes | SEC-006 (token in logs) |
| A10 | SSRF | Partial | SEC-003 (URL path injection — adjacent) |

---

## Not Applicable (SDK Context)

- Rate limiting on login endpoints: N/A — SDK is not a server
- Session management: N/A — stateless Bearer token
- CSRF: N/A — no browser interaction
- SQL injection: N/A — no database
- Command injection via shell: The `Command` and `Args` fields on `Container` and `Script`
  are string slices passed directly to the Argo API as YAML fields. They are NOT executed
  locally by the SDK — the Argo engine executes them in a pod. No local shell is involved.
  This is NOT a command injection vulnerability in the SDK.
