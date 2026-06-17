# CRD Parity Matrix — theo-forge

This document records which `WorkflowSpec` fields each builder type
exposes. (T8.1 / arch-workflow-template-incomplete-api-surface)

Argo Workflows defines four CRD types that share the underlying
`WorkflowSpec`. theo-forge models each separately as a Go struct.
Some `WorkflowSpec` fields are intentionally omitted from the smaller
template types where they would be no-ops at runtime.

## Legend

- ✅ — field is exposed on the builder
- ⬜ — field is intentionally omitted (rationale below)
- ⚠️ — field is missing and SHOULD be added (tracked for a future plan)

## Field Matrix

| WorkflowSpec field | `Workflow` | `WorkflowTemplate` | `ClusterWorkflowTemplate` | `CronWorkflow` |
|---|---|---|---|---|
| Entrypoint | ✅ | ✅ | ✅ | ✅ |
| Templates | ✅ | ✅ | ✅ | ✅ |
| Arguments | ✅ | ✅ | ✅ | ✅ |
| Volumes | ✅ | ✅ | ⬜ (cluster-wide PVCs are unusual) | ✅ |
| VolumeClaimTemplates | ✅ | ⬜ (re-usable templates do not own PVCs) | ⬜ | ⬜ |
| ServiceAccountName | ✅ | ✅ | ✅ | ✅ |
| Parallelism | ✅ | ⚠️ | ⚠️ | ⚠️ |
| ActiveDeadlineSeconds | ✅ | ⬜ | ⬜ | ⬜ |
| NodeSelector | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Tolerations | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Suspend | ✅ | ⬜ | ⬜ | ✅ (different field: `Suspend *bool`) |
| HostNetwork | ✅ | ⚠️ | ⚠️ | ⚠️ |
| TTLStrategy | ✅ | ⬜ (templates do not run) | ⬜ | ⬜ |
| PodGC | ✅ | ⬜ | ⬜ | ⬜ |
| Priority | ✅ | ⬜ | ⬜ | ⬜ |
| OnExit | ✅ | ⬜ | ⬜ | ⬜ |
| Metrics | ✅ | ⬜ | ⬜ | ⬜ |
| ArchiveLogs | ✅ | ⬜ | ⬜ | ⬜ |
| RetryStrategy | ✅ | ⬜ (set on templates instead) | ⬜ | ⬜ |
| ImagePullSecrets | ✅ | ⚠️ | ⚠️ | ⚠️ |
| PodSpecPatch | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Synchronization | ✅ | ⬜ | ⬜ | ⬜ |
| Hooks (lifecycle) | ✅ | ⬜ | ⬜ | ⬜ |
| DNSConfig / DNSPolicy | ✅ | ⚠️ | ⚠️ | ⚠️ |
| PodDisruptionBudget | ✅ | ⬜ | ⬜ | ⬜ |
| PodMetadata | ✅ | ⬜ | ⬜ | ⬜ |
| SecurityContext | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Affinity | ✅ | ⚠️ | ⚠️ | ⚠️ |
| AutomountServiceAccountToken | ✅ | ⚠️ | ⚠️ | ⚠️ |
| WorkflowTemplateRef | ✅ | ⬜ (would be self-referential) | ⬜ | ⬜ |
| ArtifactGC / ArtifactRepositoryRef | ✅ | ⬜ | ⬜ | ⬜ |
| TemplateDefaults | ✅ | ⚠️ | ⚠️ | ⚠️ |

## Rationale for omissions

- **VolumeClaimTemplates on WorkflowTemplate**: PVCs are workflow-instance
  scoped; a reusable template should not own them. Use the
  `VolumeClaimTemplates` field on the consuming `Workflow`.
- **TTLStrategy / PodGC / OnExit on templates**: these govern workflow-instance
  lifecycle; templates do not execute on their own.
- **Synchronization / Hooks on templates**: per-template synchronization
  exists at the `TemplateModel` level (Container/Script/DAG/Steps), not
  the WorkflowTemplate root.

## "⚠️" backlog

The 12 fields marked ⚠️ are tracked as a backlog enhancement (not
P1 for v0.5.0). They can be added in a future patch release without
breaking changes (every field is additive).

When adding any of them: mirror the field exactly as on `Workflow`,
update this matrix, and add a builder test that round-trips the field
through YAML.
