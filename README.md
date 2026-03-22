<p align="center">
  <strong>Forge</strong><br>
  <em>Build Argo Workflows in Go. No YAML required.</em>
</p>

<p align="center">
  <a href="https://github.com/usetheodev/theo-forge/actions/workflows/ci.yml"><img src="https://github.com/usetheodev/theo-forge/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/usetheodev/theo-forge"><img src="https://pkg.go.dev/badge/github.com/usetheodev/theo-forge.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/usetheodev/theo-forge"><img src="https://goreportcard.com/badge/github.com/usetheodev/theo-forge" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://github.com/usetheodev/theo-forge/releases"><img src="https://img.shields.io/github/v/release/usetheodev/theo-forge?include_prereleases&sort=semver" alt="Release"></a>
</p>

---

Forge is a **type-safe Go SDK** for building [Argo Workflows](https://argoproj.github.io/workflows/) programmatically — the Go equivalent of [Hera](https://github.com/argoproj-labs/hera) (Python).

Define workflows as Go structs, get compile-time safety, and let Forge handle the YAML serialization. All **198 upstream Argo Workflows examples** round-trip through Forge models.

## Why Forge?

| Without Forge | With Forge |
|---|---|
| Hand-write YAML, pray for valid indentation | Type-safe Go structs with compile-time checks |
| Copy-paste templates between files | Compose and reuse templates as Go values |
| String-based parameter references | Expression builder with autocomplete |
| Discover errors at submit time | Catch mistakes before `kubectl apply` |
| No programmatic workflow generation | Generate workflows dynamically from code |

## Install

```bash
go get github.com/usetheodev/theo-forge
```

Requires **Go 1.25+**.

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    forge "github.com/usetheodev/theo-forge"
)

func main() {
    w := &forge.Workflow{
        GenerateName: "hello-",
        Entrypoint:   "main",
        Templates: []forge.Templatable{
            &forge.Container{
                Name:    "main",
                Image:   "alpine:3.18",
                Command: []string{"echo"},
                Args:    []string{"hello world"},
            },
        },
    }

    yaml, err := w.ToYAML()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(yaml)
}
```

Output:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: hello-
spec:
  entrypoint: main
  templates:
    - name: main
      container:
        image: alpine:3.18
        command:
          - echo
        args:
          - hello world
```

## Examples

### Diamond DAG

Build complex dependency graphs with a fluent API:

```go
echoTpl := &forge.Container{
    Name:    "echo",
    Image:   "alpine:3.18",
    Command: []string{"echo"},
    Args:    []string{forge.InputParam("msg")},
    Inputs:  []forge.Parameter{{Name: "msg"}},
}

dag := &forge.DAG{Name: "diamond"}

A := &forge.Task{Name: "A", Template: "echo", Arguments: []forge.Parameter{{Name: "msg", Value: ptr("Task A")}}}
B := &forge.Task{Name: "B", Template: "echo", Arguments: []forge.Parameter{{Name: "msg", Value: ptr("Task B")}}}
C := &forge.Task{Name: "C", Template: "echo", Arguments: []forge.Parameter{{Name: "msg", Value: ptr("Task C")}}}
D := &forge.Task{Name: "D", Template: "echo", Arguments: []forge.Parameter{{Name: "msg", Value: ptr("Task D")}}}

A.Then(B)   // A → B
A.Then(C)   // A → C
B.Then(D)   // B → D
C.Then(D)   // C → D
dag.AddTasks(A, B, C, D)

w := &forge.Workflow{
    GenerateName: "diamond-",
    Entrypoint:   "diamond",
    Templates:    []forge.Templatable{echoTpl, dag},
}
```

```
     A
    / \
   B   C
    \ /
     D
```

### Conditional Logic (Coinflip)

```go
flip := &forge.Script{
    Name:    "flip-coin",
    Image:   "python:3.11-alpine",
    Command: []string{"python"},
    Source:  `import random; print("heads" if random.randint(0,1) == 0 else "tails")`,
}

heads := &forge.Container{
    Name: "heads", Image: "alpine:3.18",
    Command: []string{"echo"}, Args: []string{"it was heads"},
}

tails := &forge.Container{
    Name: "tails", Image: "alpine:3.18",
    Command: []string{"echo"}, Args: []string{"it was tails"},
}

dag := &forge.DAG{Name: "coinflip"}
flipTask := &forge.Task{Name: "flip", Template: "flip-coin"}
headsTask := &forge.Task{Name: "heads", Template: "heads", When: `{{tasks.flip.outputs.result}} == "heads"`}
tailsTask := &forge.Task{Name: "tails", Template: "tails", When: `{{tasks.flip.outputs.result}} == "tails"`}

flipTask.Then(headsTask)
flipTask.Then(tailsTask)
dag.AddTasks(flipTask, headsTask, tailsTask)
```

### REST Client

Submit, list, and lint workflows against a running Argo server:

```go
import "github.com/usetheodev/theo-forge/client"

svc := client.NewWorkflowsService(
    "https://argo.example.com",
    "my-token",
    "default",
)

// Submit a workflow
result, err := svc.CreateWorkflow(ctx, w)

// List workflows
workflows, err := svc.ListWorkflows(ctx, nil)

// Lint before submitting
linted, err := svc.LintWorkflow(ctx, w)
```

### Expression Builder

Build Argo expressions with type safety instead of raw strings:

```go
import "github.com/usetheodev/theo-forge/expr"

// Reference task outputs
ref := expr.Tasks("my-task").Attr("outputs.result")
fmt.Println(ref.Tmpl()) // {{tasks.my-task.outputs.result}}

// Build conditionals
cond := expr.Steps("validate").Attr("outputs.result").Eq(expr.C("success"))
```

## Features

### Template Types

| Type | Description |
|------|-------------|
| `Container` | Docker container execution |
| `Script` | Inline scripts (Python, Bash, etc.) |
| `DAG` | Directed acyclic graph with `Task` nodes |
| `Steps` | Sequential/parallel step groups |
| `ResourceTemplate` | Kubernetes resource create/apply |
| `HTTPTemplate` | HTTP requests |
| `Suspend` | Pause workflow execution |
| `ContainerSet` | Multiple containers in a single pod |

### Reusable Templates

| Type | Description |
|------|-------------|
| `WorkflowTemplate` | Namespace-scoped reusable templates |
| `ClusterWorkflowTemplate` | Cluster-scoped reusable templates |
| `CronWorkflow` | Scheduled workflow execution |

### Data Flow

- **Parameters** — Named inputs/outputs with defaults and value references
- **Artifacts** — S3, GCS, HTTP, Git, Raw, Azure, OSS, HDFS
- **Environment variables** — Literals, Secrets, ConfigMaps

### And More

- Retry strategies with backoff
- Timeouts and active deadlines
- Resource requests/limits
- Node selectors and tolerations
- Volume mounting (EmptyDir, Secret, ConfigMap, PVC, etc.)
- Synchronization (mutex, semaphore)
- Lifecycle hooks and memoization
- Parallelism limits and TTL strategies
- Metrics and gauges
- OnExit handlers
- Pod and container security contexts
- Artifact garbage collection

## Packages

| Package | Description |
|---------|-------------|
| `theo-forge` | Core builder API — main types and fluent constructors |
| `theo-forge/model` | Serializable types matching the Argo Workflows API schema |
| `theo-forge/expr` | Expression DSL for conditionals and parameter references |
| `theo-forge/client` | REST client for the Argo Workflows API |
| `theo-forge/serialize` | YAML/JSON serialization and file I/O helpers |
| `theo-forge/validate` | Resource unit validation (binary/decimal units) |
| `theo-forge/config` | Global configuration and hook management |

## Running Tests

```bash
go test ./...

# With race detection
go test -race ./...

# Update golden files after intentional changes
go test ./... -update-golden
```

## Contributing

Contributions are welcome! Here's how to get started:

1. **Fork** the repository
2. **Create a branch** from `develop` (`git checkout -b feat/my-feature develop`)
3. **Write tests first** — we follow TDD strictly
4. **Make your changes** and ensure all tests pass (`go test -race ./...`)
5. **Lint** your code (`golangci-lint run`)
6. **Open a Pull Request** against `develop`

Please keep PRs focused — one feature or fix per PR.

## License

[MIT License](LICENSE) — use it however you want.
