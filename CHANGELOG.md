# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/)
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

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
