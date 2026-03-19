# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/)
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Golden test framework with `-update-golden` flag for YAML output regression testing
- `GetNamespace()` method on `Workflow` for `Buildable` interface compliance
- `CreateWorkflow` and `LintWorkflow` convenience methods on `WorkflowsService` that accept `Buildable` interface
- `VolumeMounts` field on `model.ContainerSetModel` for container set volume support
- Comprehensive type aliases in `aliases.go` for `model/`, `client/`, and `expr/` packages
- Shared build helpers (`build_helpers.go`) centralizing input/output/env/volume/metadata/metrics building

### Changed
- `GlobalConfig` hooks are now wired into the `Build()` pipeline for `Workflow`, `WorkflowTemplate`, `ClusterWorkflowTemplate`, and `CronWorkflow`
- `FormatToken` on `WorkflowsService` is now exported (was `formatToken`)

### Fixed
- Project compilation: removed duplicate type definitions in root package that conflicted with `model/` aliases
- `ContainerSet.BuildTemplate()` now correctly assigns containers and volume mounts to the returned `TemplateModel`

### Removed
- Duplicate type definitions from root package: `RetryPolicy`, `Backoff`, `RetryStrategyModel`, `AccessMode`, `ArchiveStrategy`, `SecretKeySelector`, all `*VolumeModel`, `*ArtifactModel`, `EnvVarModel`, `HTTPModel`, `HTTPHeader`, `ContainerSetModel` — these now live exclusively in `model/` and are re-exported via type aliases
- Copy-pasted `buildInputs`/`buildOutputs`/`buildEnv`/`buildVolumeMounts`/`buildMetadata`/`buildMetrics` methods from `Container`, `Script`, `DAG`, `Steps` — replaced by shared helpers
