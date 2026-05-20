# Dependency Graph — theo-forge

**Date:** 2026-05-20
**Phase:** 3 (Architecture Review)

## Package Inventory

| Package | Path | Role | Lines |
|---|---|---|---|
| `forge` (root) | `.` | Builder DSL, public API surface | ~2000 (production), ~4000 (test) |
| `model` | `./model` | Pure data structs (serialization targets) | 1172 |
| `serialize` | `./serialize` | YAML/JSON I/O | ~154 |
| `validate` | `./validate` | CPU/memory unit validation | ~175 |
| `config` | `./config` | GlobalConfig singleton + hooks | ~163 |
| `expr` | `./expr` | Argo expression DSL | 317 |
| `client` | `./client` | REST client for Argo API | 296 |

## Import Graph (Directed, Runtime Only)

```
forge (root)
  ├── model          [direct]
  ├── serialize      [direct, in helpers.go + workflow.go + workflow_template.go]
  ├── config         [direct, in helpers.go]
  └── validate       [direct, in helpers.go]

config
  └── model          [direct]

client
  └── model          [direct]

serialize
  ├── model          [direct]
  └── sigs.k8s.io/yaml [external]

validate
  └── model          [direct]

expr
  (no internal imports — stdlib only)

model
  (no imports at all — pure structs)
```

## Dependency Matrix

| From \ To | model | serialize | config | validate | expr | client |
|---|---|---|---|---|---|---|
| forge (root) | YES | YES | YES | YES | — | — |
| config | YES | — | — | — | — | — |
| client | YES | — | — | — | — | — |
| serialize | YES | — | — | — | — | — |
| validate | YES | — | — | — | — | — |
| expr | — | — | — | — | — | — |

## Cycle Detection

**Result: NO CYCLES FOUND.**

All edges are unidirectional. `model` is a pure leaf with no imports. `expr` is a pure leaf with stdlib-only imports. All other packages (config, client, serialize, validate) depend only on `model`. The root `forge` package depends on all sub-packages but no sub-package imports `forge`.

## Layering Analysis

```
Layer 0 (Foundation):  model, expr
Layer 1 (Adapters):    config, serialize, validate, client
Layer 2 (Builder DSL): forge (root)
```

The layering is clean and consistent. No layering violations detected: no Layer 0 package imports Layer 1 or Layer 2. No Layer 1 package imports another Layer 1 package (no cross-adapter coupling). The root package aggregates all layers correctly.

## Notable Structural Observations

1. **helpers.go is the only file with 4 imports** — it pulls config, model, serialize, and validate simultaneously. This is the "integration hub" of the root package and by LOC/responsibility is a moderate SRP concern (build helpers + file I/O + validation proxies + config wiring).

2. **expr is fully decoupled** — imports only `fmt` and `strings`. No dependency on model types, no dependency on runtime state. This is architecturally correct for a pure DSL.

3. **client defines its own Buildable interface** — `client.Buildable` is defined locally in `client/client.go` rather than importing the root package interface. This is correct DIP practice (avoids circular import) but creates a second interface definition that mirrors `forge.Workflow.Build()` signature.

4. **workflow_template.go imports serialize directly** — 3 root-package files (workflow.go, workflow_template.go, helpers.go) import serialize. This is not a layering violation but means serialization concerns are spread across 3 builder files rather than being centralized.

5. **Two `globalConfig` variables exist** — `config.globalConfig` (the singleton) and `forge.globalConfig` (a package-level pointer captured at init). Both are separate variables; the forge-level one is frozen at program startup.
