# System Invariants — theo-forge SDK
**Defined:** Phase 7, Global Iteration 7
**Date:** 2026-05-20

## Overview

6 system invariants were defined and validated during Phase 7. 5 are validated. 1 is violated and maps directly to security finding SEC-001.

---

## Invariant 1 — Build() output always contains required Argo schema fields

**Description:** Every successful `Build()` call must produce a `WorkflowModel` with:
- `apiVersion: argoproj.io/v1alpha1`
- `kind: Workflow`
- Non-empty `spec.entrypoint`
- At least one `spec.template`

**Status: VALIDATED**

**Evidence:**
- `workflow.Build()` sets `DefaultAPIVersion` and `DefaultKind` unconditionally at `workflow.go:234`
- `Workflow.validate()` at `workflow.go:116` enforces `Entrypoint != ""`
- All 194 testdata round-trip examples parse and re-serialize correctly
- Direct code run confirms: APIVersion, Kind, Entrypoint, Templates all present

---

## Invariant 2 — globalConfig hooks dispatch exactly once per Build() per registration

**Description:** Every hook registered via `RegisterTemplateHook` or `RegisterWorkflowHook` on the global singleton must be called exactly once per `Build()` call.

**Status: VALIDATED**

**Evidence:**
- Atomic counter test: 3 `DispatchWorkflowHooks` calls produce exactly 3 invocations
- `config.go:133` iterates hook slice under `mu.RLock` — no duplication possible
- Order is guaranteed by slice append order at registration time

**Caveat:** This invariant applies only to the global config singleton. Hooks registered on `NewConfig()` isolated instances are never dispatched (see VAL-004 finding). The invariant is vacuously true for isolated configs because their hooks are never called at all.

---

## Invariant 3 — Golden files reflect current Build() serialization output

**Description:** All `TestGolden*` tests must pass without the `-update-golden` flag.

**Status: VALIDATED**

**Evidence:**
- All 6 golden tests pass: `TestGoldenSimpleContainer`, `TestGoldenDiamondDAG`, `TestGoldenScriptWorkflow`, `TestGoldenStepsWorkflow`, `TestGoldenWorkflowTemplate`, `TestGoldenCronWorkflow`
- Golden files in `testdata/*.yaml` are byte-for-byte current
- Golden comparison in `goldenTest()` uses exact string match (stricter than semantic)

---

## Invariant 4 — Serialize package round-trips all 194 upstream Argo example YAMLs

**Description:** `TestRoundTripTestdataExamples` must pass across all YAML files in `testdata/examples/`.

**Status: VALIDATED**

**Evidence:**
- 194 files walk completed; all pass
- Covers Workflow, WorkflowTemplate, ClusterWorkflowTemplate, and CronWorkflow kinds
- Unknown kind values are skipped with `t.Skip` (not failures)
- Semantic comparison via `assertSemantic()` handles key ordering and array normalization

---

## Invariant 5 — README minimal quickstart compiles to valid Argo YAML

**Description:** The hello-world `Workflow` + `Container` example at the top of `README.md` must compile and produce well-formed output.

**Status: VALIDATED**

**Evidence:**
- Direct `go run` confirms: compiles successfully
- Output contains correct `apiVersion`, `kind`, `metadata.generateName`, `spec.entrypoint`, and `spec.templates`
- YAML is parseable and semantically equivalent to the documented output

**Caveat:** This invariant applies to the quickstart only. The diamond DAG example on `README.md:108` and the REST client example on `README.md:187` fail to compile (findings VAL-001 and VAL-002).

---

## Invariant 6 — ToFile() output is always confined to the specified output directory

**Description:** `Workflow.ToFile(outputDir, name)` and `serialize.WorkflowToFile()` must not write outside `outputDir`. Path traversal via the `name` parameter must be rejected.

**Status: VIOLATED**

**Evidence:**
```
w.ToFile("/tmp/safe-dir", "../escaped")
→ writes to /tmp/escaped.yaml
→ TRAVERSAL: file written outside expected dir!
→ File confirmed to exist at traversal path
```

**Root cause:** `serialize.go:133` — `path := filepath.Join(absDir, fileName)`. `filepath.Join` resolves `..` components without a containment assertion. No check that `path` remains within `absDir`.

**Fix:**
```go
path := filepath.Join(absDir, fileName)
// Containment check
if !strings.HasPrefix(path, absDir+string(os.PathSeparator)) {
    return "", fmt.Errorf("filename %q would write outside output directory", fileName)
}
```

**Severity:** HIGH (SEC-001). Exploitable when `name` parameter comes from user-controlled input.

---

## Summary Table

| # | Invariant | Status | Evidence |
|---|---|---|---|
| 1 | Build() output has required Argo schema fields | VALIDATED | Code analysis + 194 round-trips |
| 2 | globalConfig hooks dispatch exactly once per Build() | VALIDATED | Atomic counter test |
| 3 | Golden files match current Build() output | VALIDATED | 6 golden tests pass |
| 4 | 194 testdata examples round-trip cleanly | VALIDATED | TestRoundTripTestdataExamples passes |
| 5 | README quickstart compiles to valid YAML | VALIDATED | go run success |
| 6 | ToFile() confines output to target directory | **VIOLATED** | Path traversal to /tmp/escaped.yaml confirmed |
