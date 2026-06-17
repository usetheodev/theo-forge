# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/)
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- `NewRetryOnFailure(limit, initial, factor, maxDuration)` factory in `parameter.go` — preconfigured RetryStrategy with OnFailure policy, exponential Backoff, and skip-OOM expression (`asInt(lastRetry.exitCode) != 137`). Eliminates duplicated retry-strategy boilerplate in build pipelines (forge-contributions-abc-plan T1.1 / ADR-001).
- `ResourcesTiny()`, `ResourcesSmall()`, `ResourcesMedium()` factories in `types.go` — T-shirt-size `*ResourceRequirements` presets (50m/32Mi, 100m/128Mi, 500m/512Mi). Each call returns a freshly-allocated struct safe to mutate (forge-contributions-abc-plan T1.2 / ADR-002).
- 3 golden YAML fixtures in `testdata/resources-{tiny,small,medium}.golden.yaml` — wire-format snapshot of each preset; updated via `-update-golden` if defaults ever change (T1.2 / ADR-005).
- `model.IntOrString` (T4.3 / ADR-002) replaces `interface{}` in `PodDisruptionBudget.MinAvailable/MaxUnavailable`, `HTTPGetAction.Port`, `TCPSocketAction.Port`. Mirrors `k8s.io/apimachinery/pkg/util/intstr.IntOrString` semantics locally.
- `sharedTemplateFields` helper (T4.1) — Script gains `EnvFrom`, `Ports`, `SecurityContext`, `ReadinessProbe`, `LivenessProbe`, `InitContainers`, `Parallelism`, `ArchiveLocation` (previously Container-only).
- Composable hook system (T8.2): `RegisterNamedTemplateHook`, `RemoveTemplateHook`, `DispatchNamedTemplateHooks` (+ workflow counterparts) — named, removable, fallible hooks with error short-circuit.
- `ClusterWorkflowTemplateToFile` / `ClusterWorkflowTemplateFromFile` (T8.3) — dedicated round-trip helpers complete the parity matrix.
- `Set/GetServiceAccountName`, `Set/GetImagePullPolicy`, `Set/GetVerifySSL` on `config.GlobalConfig` (T4.5).
- Sentinel errors in `model/errors.go`: `ErrPathTraversal`, `ErrEmptyImage`, `ErrEntrypointMissing`, `ErrTemplateAmbiguous`, `ErrTemplateMissing`, `ErrInvalidName`, `ErrInvalidNamespace`, `ErrResponseTooLarge`, `ErrRedactedTokenLoaded`, `ErrInvalidTimeout` (#T2.1–T2.6, fix-all-review-findings plan).
- `expr.RawC(s string) Expr` — bypasses Argo escaping; SDK-internal use only. New companion to the now-escaped `expr.C` (#T2.2).
- Test-only `config.ResetGlobalConfigForTest(t)` + exported `config.TestLock` to serialize global-state mutation across tests (#T0.1, EC-1).
- `client.WorkflowsService` now implements `fmt.Stringer`, `MarshalJSON`, `UnmarshalJSON` with Token redaction (#T2.6, SEC-006, EC-6).
- `config.GlobalConfig` now implements the same redaction trio (#T2.6, SEC-007).
- CI gofmt gate (#T1.2).

### Changed
- License standardized to **Apache-2.0** (was MIT). Aligns all Theo open-core pillars under a single license — see root `CLAUDE.md` strategic review of 2026-05-14.
- License standardized to **Apache-2.0** (was MIT). Aligns all usetheo open-core pillars under a single license — see root `CLAUDE.md` strategic review of 2026-05-14.
- **BREAKING (behavior):** `expr.C` and string-interpolation methods (`Contains`, `Matches`, `StartsWith`, `EndsWith`, `Sprig.{Trim,Upper,Lower,Replace}`) now escape single quotes for safe Argo embedding. Callers that pre-escaped input must switch to `expr.RawC` to avoid double-escaping (#T2.2, SEC-002, EC-4).
- `client.WorkflowsService.VerifySSL` now actually controls TLS verification on the HTTP transport (was a dead field). `MinVersion` fixed to TLS 1.2. Externally-set `HTTPClient` is honored without overwrite (#T2.4, SEC-004, code-p4-verifyssl-dead, EC-5).
- `client` HTTP body reads bounded to 32 MiB to prevent memory exhaustion (`ErrResponseTooLarge`) (#T2.5, SEC-005).
- `client` URL-building methods (CreateWorkflow{FromModel}, GetWorkflow, DeleteWorkflow, ListWorkflows, LintWorkflow{FromModel}, StopWorkflow, TerminateWorkflow, SuspendWorkflow, ResumeWorkflow) now validate `name`/`namespace` against RFC1123 DNS subdomain rules and URL-escape segments (#T2.3, SEC-003).
- All 13 production Go files reformatted with `gofmt -w`; CI now fails on `gofmt -l .` non-empty output (#T1.2, code-p4-gofmt-violations).
- `doc.go` godoc references corrected: `[StepsExpr]` → `[expr.Steps]`, others namespaced for accurate pkg.go.dev rendering (#T1.3, code-p4-doc-stepsexpr-nonexistent).
- **BREAKING (T4.5):** `config.GlobalConfig` data fields (`Host`, `Token`, `Namespace`, `Image`, `ServiceAccountName`, `ImagePullPolicy`, `VerifySSL`) are now unexported. Use the `Set*`/`Get*` accessor methods instead of direct field access. Prevents mutex bypass and silent data races.
- helpers.go (310 LoC) split into 4 focused files per SRP (T4.2): `build_helpers.go`, `file_io_helpers.go`, `validation_helpers.go`, `config_helpers.go`.
- `ContainerSet.BuildTemplate()` reuses shared `buildInputsFromParams` / `buildOutputsFromParams` / `buildVolumeMountModels` helpers (T4.6 / code-p4-containerset-inline-io-build).
- expr.G + expr.Eq marked `Deprecated` (T4.4 / T4.10) — removal scheduled for v0.6.0.
- 191 godocs added (revive.exported re-enabled); zero lint warnings under golangci-lint v2 strict.
- Quality Gates Tier 2 / Strict in place: golangci-lint v2 + per-package coverage gate + govulncheck + osv-scanner; `make verify` single source of truth.

### Fixed
- All 67 deep-review findings remediated or marked wont_fix in `review-output/review.db`.

### Fixed
- **Security:** `serialize.WorkflowToFile` now rejects path traversal (`../`), absolute paths, and empty/dot names via the new `containedJoin` helper. Symlinks inside the output directory are resolved before the containment check (#T2.1, SEC-001, sec-001-path-traversal-confirmed, inv-006, EC-3).
- **Security:** `expr.C` and all string-interpolation methods escape single quotes; prevents Argo expression injection via attacker-controlled input (#T2.2, SEC-002, sec-002-expr-injection-confirmed).

## [0.4.0] - 2026-05-14

### Added
- `DefaultPodAffinityFor(w *Workflow) *model.Affinity` helper that builds the canonical podAffinity term used to co-locate every pod of a workflow on the same node (matches on `workflows.argoproj.io/workflow` with topology key `kubernetes.io/hostname`). Uses the literal `Name` when set; falls back to the Argo template variable `{{workflow.name}}` for `GenerateName`-only workflows (#11)
- `Workflow.DisableDefaultAffinity` opt-out field for the new default podAffinity injection (#11)

### Changed
- `Workflow.Build()` now injects `DefaultPodAffinityFor(w)` into `WorkflowSpec.Affinity` when the workflow declares `VolumeClaimTemplates` AND `Affinity` is nil AND `DisableDefaultAffinity` is false. This eliminates the PVC ReadWriteOnce Multi-Attach race that surfaces when DAG steps share an RWO PVC and Kubernetes schedules them on different nodes. Workflows without `VolumeClaimTemplates`, with a user-supplied `Affinity`, or with `DisableDefaultAffinity: true` are unaffected (#11)

## [0.3.1] - 2026-04-14

### Fixed
- Maintenance release between v0.3.0 and v0.4.0; restore tag in CHANGELOG for completeness (INFRA-007).

## [0.3.0] - 2026-04-13

### Added
- `SeccompProfile` and `Capabilities` fields on container `SecurityContext` for rootless BuildKit and hardened workloads (#F1.1)
- Complete `Affinity` model (`NodeAffinity`, `PodAffinity`, `PodAntiAffinity`) on `WorkflowSpec` and `TemplateModel` — mirrors k8s.io/api/core/v1 without importing the dependency (#F1.2)
- `WorkflowStatusDetail` and `NodeStatus` types for typed parsing of workflow status including per-node exit codes (#F1.3)
- `Shutdown` field on `WorkflowSpec` for typed workflow cancellation (#F1.3)
- `ExitCode` field on `OutputsModel` for reading per-node exit codes (#F1.3)
- `Status` field on `WorkflowModel` for deserializing workflow status from API/CRD responses (#F1.3)
- `ColocateByLabel` helper for PVC ReadWriteOnce pod co-location pattern (#F2.2)
- `ParseWorkflowStatusFromUnstructured` for extracting typed status from K8s unstructured objects (#F3.1)
- `AllPodNodesExitedZero` for detecting false failures from daemon sidecar termination (#F3.1)
- `StopWorkflow`, `TerminateWorkflow`, `SuspendWorkflow`, `ResumeWorkflow` client operations (#F4.1)

## [0.2.1] - 2026-03-22

### Added
- Full Argo Workflows feature coverage: all 198 upstream examples now round-trip through forge models
- Synchronization support (mutex, semaphore, DB-backed) at workflow and template levels
- Lifecycle hooks (`hooks:`) at workflow, template, step, and task levels
- Memoization (`memoize:`) for template output caching
- Inline templates for steps and DAG tasks
- `withSequence` for numeric fan-out in steps and DAG tasks
- `podSpecPatch` at workflow and template levels
- DNS configuration (`dnsConfig`, `dnsPolicy`) for workflow pods
- Pod disruption budget support
- Security context at pod and container levels
- Workflow template references (`workflowTemplateRef`)
- Artifact garbage collection (`artifactGC`) at workflow and artifact levels
- Artifact repository references (`artifactRepositoryRef`)
- Template defaults (`templateDefaults`) for workflow-level defaults
- CronWorkflow `schedules` (multiple), `when`, `workflowMetadata`, `stopStrategy`
- Data template type for data transformations
- Resource template `flags`, `setOwnerReference`, `mergeStrategy`, `manifestFrom`
- HTTP artifact `auth` (OAuth2, BasicAuth, ClientCert) and `headers`
- S3/GCS/Azure/OSS artifact secret key references
- HDFS artifact `force` field
- Container `securityContext`, `envFrom`, `readinessProbe`, `livenessProbe`, `lifecycle`, `dependencies`
- Sidecar `mirror`, `daemon`, `lifecycle`, `readinessProbe`
- Init containers and sidecars on Container and Script templates
- `Suspend` template inputs/outputs
- `WorkflowTemplateFromYAML` and `CronWorkflowFromYAML` deserializers in serialize package
- 28 hand-crafted programmatic builder tests matching Hera-generated YAML exactly
- Round-trip test framework verifying all 198 Hera upstream examples parse and re-serialize correctly
- New `serialize/` package: standalone functions for YAML/JSON serialization and file I/O (`WorkflowToYAML`, `WorkflowFromFile`, etc.)
- New `validate/` package: standalone resource unit validation (`BinaryUnit`, `DecimalUnit`, `ResourceRequirements`)
- New `config/` package: `GlobalConfig` singleton and hook management extracted from root
- `CronWorkflow` now lives in its own file (`cron_workflow.go`) separate from `workflow_template.go`
- Golden test framework with `-update-golden` flag for YAML output regression testing
- `GetNamespace()` method on `Workflow` for `Buildable` interface compliance
- `CreateWorkflow` and `LintWorkflow` convenience methods on `WorkflowsService` that accept `Buildable` interface
- `VolumeMounts` field on `model.ContainerSetModel` for container set volume support
- Comprehensive type aliases in `aliases.go` for `model/`, `client/`, and `expr/` packages
- Shared build helpers (`build_helpers.go`) centralizing input/output/env/volume/metadata/metrics building

### Changed
- Root package serialization methods (`ToYAML`, `ToJSON`, `ToDict`, `FromYAML`, `FromJSON`, `ToFile`, `FromFile`) now delegate to `serialize/` package
- Root package validation functions (`ValidateBinaryUnit`, `ConvertDecimalUnit`, etc.) now delegate to `validate/` package
- Root package `GlobalConfig`, `NewConfig`, `GetGlobalConfig` now delegate to `config/` package
- `resource_template.go` and `container_set.go` now use explicit `model.` prefix instead of type aliases
- `GlobalConfig` hooks are now wired into the `Build()` pipeline for `Workflow`, `WorkflowTemplate`, `ClusterWorkflowTemplate`, and `CronWorkflow`
- `FormatToken` on `WorkflowsService` is now exported (was `formatToken`)

### Fixed
- Project compilation: removed duplicate type definitions in root package that conflicted with `model/` aliases
- `ContainerSet.BuildTemplate()` now correctly assigns containers and volume mounts to the returned `TemplateModel`

### Removed
- **BREAKING:** ~70 internal type aliases from `aliases.go` (e.g., `forge.TemplateModel`, `forge.WorkflowModel`, `forge.ContainerModel`) — use `forge/model` package directly
- **BREAKING:** Client type re-exports (`forge.WorkflowsService`, `forge.APIError`, `forge.HTTPClient`) — use `forge/client` package directly
- **BREAKING:** Expression type/function re-exports (`forge.Expr`, `forge.E`, `forge.C`, `forge.InputParam`, etc.) — use `forge/expr` package directly
- **BREAKING:** Re-exported functions (`forge.ParseImagePullPolicy`, `forge.ParseWorkflowStatus`, `forge.NewWorkflowsService`) — use their original packages
- Duplicate type definitions from root package: `RetryPolicy`, `Backoff`, `RetryStrategyModel`, `AccessMode`, `ArchiveStrategy`, `SecretKeySelector`, all `*VolumeModel`, `*ArtifactModel`, `EnvVarModel`, `HTTPModel`, `HTTPHeader`, `ContainerSetModel` — these now live exclusively in `model/` and are re-exported via type aliases
- Copy-pasted `buildInputs`/`buildOutputs`/`buildEnv`/`buildVolumeMounts`/`buildMetadata`/`buildMetrics` methods from `Container`, `Script`, `DAG`, `Steps` — replaced by shared helpers
