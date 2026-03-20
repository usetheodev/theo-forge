# Forge

Go SDK for building and managing [Argo Workflows](https://argoproj.github.io/workflows/) programmatically. Type-safe, builder-style API — no YAML by hand.

Forge is the Go equivalent of [Hera](https://github.com/argoproj-labs/hera) (Python SDK for Argo Workflows).

## Installation

```bash
go get github.com/usetheo/theo/forge
```

Requires **Go 1.25+**.

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/usetheo/theo/forge"
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

## Features

### Template Types

| Type | Description |
|------|-------------|
| `Container` | Docker container execution |
| `Script` | Inline script (Python, Bash, etc.) |
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

### Additional Capabilities

- Retry strategies with backoff
- Timeouts and active deadlines
- Resource requests/limits
- Node selectors and tolerations
- Volume mounting (EmptyDir, Secret, ConfigMap, PVC, etc.)
- Parallelism limits
- TTL strategies
- Metrics and gauges
- OnExit handlers

## Examples

### Diamond DAG

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

A.Then(B)
A.Then(C)
B.Then(D)
C.Then(D)
dag.AddTasks(A, B, C, D)

w := &forge.Workflow{
    GenerateName: "diamond-",
    Entrypoint:   "diamond",
    Templates:    []forge.Templatable{echoTpl, dag},
}

yaml, _ := w.ToYAML()
```

### Script with Conditionals (Coinflip)

```go
flip := &forge.Script{
    Name:    "flip-coin",
    Image:   "python:3.11-alpine",
    Command: []string{"python"},
    Source:  `import random; print("heads" if random.randint(0,1) == 0 else "tails")`,
}

heads := &forge.Container{
    Name:    "heads",
    Image:   "alpine:3.18",
    Command: []string{"echo"},
    Args:    []string{"it was heads"},
}

tails := &forge.Container{
    Name:    "tails",
    Image:   "alpine:3.18",
    Command: []string{"echo"},
    Args:    []string{"it was tails"},
}

dag := &forge.DAG{Name: "coinflip"}
flipTask := &forge.Task{Name: "flip", Template: "flip-coin"}
headsTask := &forge.Task{Name: "heads", Template: "heads", When: `{{tasks.flip.outputs.result}} == "heads"`}
tailsTask := &forge.Task{Name: "tails", Template: "tails", When: `{{tasks.flip.outputs.result}} == "tails"`}

flipTask.Then(headsTask)
flipTask.Then(tailsTask)
dag.AddTasks(flipTask, headsTask, tailsTask)

w := &forge.Workflow{
    GenerateName: "coinflip-",
    Entrypoint:   "coinflip",
    Templates:    []forge.Templatable{flip, heads, tails, dag},
}
```

### REST Client

```go
import "github.com/usetheo/theo/forge/client"

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

```go
import "github.com/usetheo/theo/forge/expr"

// Reference task outputs
ref := expr.Tasks("my-task").Attr("outputs.result")
fmt.Println(ref.Tmpl()) // {{tasks.my-task.outputs.result}}

// Build conditionals
cond := expr.Steps("validate").Attr("outputs.result").Eq(expr.C("success"))
```

## Packages

| Package | Description |
|---------|-------------|
| `forge` | Builder API — main types and fluent constructors |
| `forge/model` | Serializable types — wire format matching Argo Workflows API schema |
| `forge/expr` | Expression DSL for conditionals and parameter references |
| `forge/client` | REST client for the Argo Workflows API |

## Running Tests

```bash
go test ./...
```

## License

[Apache License 2.0](LICENSE)
