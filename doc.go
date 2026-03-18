// Package forge is a Go SDK for building and managing Argo Workflows.
//
// It provides a type-safe, builder-style API for constructing Argo Workflow
// definitions in Go, without writing YAML directly. It is the Go equivalent
// of the Hera Python SDK (https://github.com/argoproj-labs/hera).
//
// # Core Types
//
// The main types mirror the Argo Workflows API:
//
//   - [Workflow] — The top-level workflow definition
//   - [DAG] — Directed acyclic graph template with [Task] nodes
//   - [Steps] — Sequential/parallel step template with [Step] and [Parallel] groups
//   - [Container] — Docker container template
//   - [Script] — Script execution template (Python, Bash, etc.)
//   - [ResourceTemplate] — K8s resource create/apply template
//   - [HTTPTemplate] — HTTP request template
//   - [Suspend] — Pause execution template
//   - [ContainerSet] — Multiple containers in a single pod
//
// # Reusable Templates
//
//   - [WorkflowTemplate] — Namespace-scoped reusable templates
//   - [ClusterWorkflowTemplate] — Cluster-scoped reusable templates
//   - [CronWorkflow] — Scheduled workflow execution
//
// # Data Types
//
//   - [Parameter] — Named input/output parameters
//   - [Artifact] — Artifact storage (S3, GCS, HTTP, Git, Raw, Azure, OSS, HDFS)
//   - [Env], [SecretEnv], [ConfigMapEnv] — Environment variables
//   - Volume types: [EmptyDirVolume], [SecretVolume], [ConfigMapVolume], etc.
//
// # Expression Builder
//
// The [Expr] type and helper functions ([E], [C], [Tasks], [StepsExpr], [InputParam])
// provide a fluent API for building Argo expressions.
//
// # REST Client
//
// [WorkflowsService] provides a REST client for the Argo Workflows API
// (create, get, list, delete, lint workflows).
//
// # Example
//
//	w := &forge.Workflow{
//	    GenerateName: "hello-",
//	    Entrypoint:   "main",
//	    Templates: []forge.Templatable{
//	        &forge.Container{
//	            Name:    "main",
//	            Image:   "alpine:3.18",
//	            Command: []string{"echo"},
//	            Args:    []string{"hello world"},
//	        },
//	    },
//	}
//	yaml, _ := w.ToYAML()
package forge
