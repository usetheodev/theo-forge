# Architecture Map — theo-forge

Generated: Phase 1, Iteration 1 (2026-05-20)

## Project Identity

- **Module**: `github.com/usetheodev/theo-forge`
- **Language**: Go 1.25
- **Type**: Library SDK (no server, no runtime)
- **Purpose**: Type-safe programmatic builder for Argo Workflows CRDs in Go. Equivalent of Hera (Python SDK).
- **Output artifact**: Argo Workflows YAML/JSON (CRD manifests), NOT a running service.
- **Current version**: v0.4.0

## High-Level Architecture

```
Consumer code
     |
     v
[root package: forge]          <-- Primary API surface
  Workflow / WorkflowTemplate / ClusterWorkflowTemplate / CronWorkflow
  Container / Script / DAG / Steps / ResourceTemplate / HTTPTemplate / Suspend / ContainerSet
  Parameter / Artifact / Volume / Env / RetryStrategy
     |
     |-- Build() / ToYAML() / ToJSON() / ToDict()
     |
     v
[forge/model]                  <-- Data contract layer (serializable structs)
  WorkflowModel / WorkflowSpec / TemplateModel / DAGModel / StepsModel / etc.
     |
     v
[forge/serialize]              <-- I/O layer
  yaml.Marshal / yaml.Unmarshal (via sigs.k8s.io/yaml)
  WorkflowToYAML / WorkflowFromYAML / WorkflowToFile / etc.
     |
     v
  YAML / JSON string (Argo Workflow CRD)

[forge/config]                 <-- Global config & hooks (singleton + isolated)
  GlobalConfig: hooks, defaults
  Applied during Build() via DispatchTemplateHooks / DispatchWorkflowHooks

[forge/expr]                   <-- Expression DSL
  Fluent builders for {{...}} and {{=...}} Argo template expressions

[forge/validate]               <-- Resource validation
  BinaryUnit / DecimalUnit / ResourceRequirements validation

[forge/client]                 <-- Argo REST client (optional)
  WorkflowsService: CRUD + lifecycle ops against Argo server API
```

## Package Dependency Graph

```
root-package (forge)
  |- imports: model, serialize, config, validate
  |- re-exports: config.PreBuildHook, config.GlobalConfig, validate.* (as facade)

forge/client
  |- imports: model (for WorkflowModel, request/response types)
  |- interfaces: HTTPClient (injectable), Buildable (accepts any forge type)

forge/serialize
  |- imports: model, sigs.k8s.io/yaml

forge/validate
  |- imports: model (for ResourceRequirements)

forge/expr
  |- imports: none from forge (standalone string-builder DSL)

forge/config
  |- imports: model (for TemplateModel, WorkflowModel hook signatures)

forge/model
  |- imports: none from forge (pure data structs, no internal deps)
```

## Layer Architecture

The codebase follows a strict layering:

| Layer | Package | Role |
|---|---|---|
| API / Builder | forge (root) | User-facing struct constructors and Build() methods |
| Data Contract | forge/model | Serializable structs with JSON/YAML tags |
| Serialization | forge/serialize | Marshal/unmarshal + file I/O |
| Config/Hooks | forge/config | Global defaults and pre-build transformation hooks |
| Expression DSL | forge/expr | Type-safe Argo expression string builders |
| Validation | forge/validate | K8s resource unit validation |
| Client | forge/client | Optional REST client for Argo API |

## Key Design Decisions

1. **No k8s.io imports**: The model package mirrors the Argo CRD schema without importing `k8s.io/api/core/v1`. This keeps the dependency footprint minimal.

2. **Builder pattern with value types**: The root package uses concrete structs (not builders/fluent-chain), keeping construction explicit and readable.

3. **Templatable interface**: All template types implement `BuildTemplate() (model.TemplateModel, error)` and `GetName() string`. This allows heterogeneous `[]Templatable` slices.

4. **Hook system**: `config.GlobalConfig` allows registering pre-build hooks that transform every `TemplateModel` and `WorkflowModel` — useful for org-wide defaults (resources, labels, etc.).

5. **Default PodAffinity injection**: `Workflow.Build()` auto-injects a podAffinity co-location term when `VolumeClaimTemplates` is non-empty and no `Affinity` is set, solving the RWO PVC Multi-Attach race. Opt-out via `DisableDefaultAffinity: true`.

6. **Facade pattern in root**: Root package re-exports config/validate functions so consumers only need one import in most cases.

## Entry Points

All critical paths start from one of:
- `Workflow.Build()` / `Workflow.ToYAML()` / `Workflow.ToJSON()` / `Workflow.ToDict()`
- `WorkflowTemplate.Build()` / `WorkflowTemplate.ToYAML()`
- `CronWorkflow.Build()` / `CronWorkflow.ToYAML()`
- `client.WorkflowsService.CreateWorkflow()` / `LintWorkflow()`
- `serialize.WorkflowFromYAML()` / `serialize.WorkflowFromFile()`

## Test Infrastructure

- Golden files in `testdata/*.golden.yaml` for YAML regression testing
- Round-trip tests for all 198 upstream Argo Workflows examples (`testdata/examples/`)
- `-update-golden` flag for intentional baseline updates
- Race detection: `go test -race ./...`
- CI: GitHub Actions with go test, go vet, golangci-lint, govulncheck, gosec
