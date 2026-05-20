package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

// --- Interface ---

// Templatable is implemented by types that can build an Argo Template.
type Templatable interface {
	BuildTemplate() (model.TemplateModel, error)
	GetName() string
}

// --- Error types ---

// NodeNameConflict is returned when duplicate step/task names are detected.
type NodeNameConflict struct {
	Name string
}

func (e *NodeNameConflict) Error() string {
	return fmt.Sprintf("node name conflict: %q already exists in this context", e.Name)
}

// InvalidTemplateCall is returned when a template is called in an invalid context.
type InvalidTemplateCall struct {
	Name    string
	Context string
}

func (e *InvalidTemplateCall) Error() string {
	return fmt.Sprintf("template %q is not callable under a %s context", e.Name, e.Context)
}

// --- Type aliases for commonly used model types ---

// These allow users to reference frequently-used Kubernetes types
// without importing the model package directly.

// ImagePullPolicy is the type alias.
type ImagePullPolicy = model.ImagePullPolicy

// ResourceRequirements is the type.
//
// For common T-shirt-size presets used by build pipelines, see
// [ResourcesTiny], [ResourcesSmall], and [ResourcesMedium].
type ResourceRequirements = model.ResourceRequirements

// ResourceList is the type.
type ResourceList = model.ResourceList

// Toleration is the type.
type Toleration = model.Toleration

// ContainerPort is the type.
type ContainerPort = model.ContainerPort

// AccessMode is the type.
type AccessMode = model.AccessMode

// WorkflowStatus is the type.
type WorkflowStatus = model.WorkflowStatus

// RetryPolicy is the type.
type RetryPolicy = model.RetryPolicy

// Backoff is the type.
type Backoff = model.Backoff

// Metric is the type.
type Metric = model.Metric

// Label is the type.
type Label = model.Label

// Counter is the type.
type Counter = model.Counter

// Gauge is the type.
type Gauge = model.Gauge

// ArchiveStrategy is the type.
type ArchiveStrategy = model.ArchiveStrategy

// TTLStrategy is the type.
type TTLStrategy = model.TTLStrategy

// PodGC is the type.
type PodGC = model.PodGC

// ContinueOn is the type.
type ContinueOn = model.ContinueOn

// TemplateRef is the type.
type TemplateRef = model.TemplateRef

// ValueFrom is the type.
type ValueFrom = model.ValueFrom

// Re-export constants.
const (
	ImagePullAlways       = model.ImagePullAlways
	ImagePullNever        = model.ImagePullNever
	ImagePullIfNotPresent = model.ImagePullIfNotPresent

	ReadWriteOnce    = model.ReadWriteOnce
	ReadOnlyMany     = model.ReadOnlyMany
	ReadWriteMany    = model.ReadWriteMany
	ReadWriteOncePod = model.ReadWriteOncePod

	RetryAlways           = model.RetryAlways
	RetryOnFailure        = model.RetryOnFailure
	RetryOnError          = model.RetryOnError
	RetryOnTransientError = model.RetryOnTransientError

	WorkflowPending    = model.WorkflowPending
	WorkflowRunning    = model.WorkflowRunning
	WorkflowSucceeded  = model.WorkflowSucceeded
	WorkflowFailed     = model.WorkflowFailed
	WorkflowError      = model.WorkflowError
	WorkflowTerminated = model.WorkflowTerminated
)

// --- Resource presets (Phase 1 / T1.2) ---

// Common T-shirt-size resource presets for build-pipeline pods.
// Each factory returns a freshly-allocated *ResourceRequirements; the
// returned pointer is safe to mutate without affecting other callers.
//
// Pick the smallest preset that holds your workload — over-allocating
// CPU/memory in build clusters wastes scheduling capacity (Argo and
// k8s scheduler reserve based on `requests`, not on actual usage).
//
// For sizes outside Tiny/Small/Medium, construct a [ResourceRequirements]
// struct literal directly.

// ResourcesTiny returns a tiny preset: 50m/32Mi requests, 100m/64Mi limits.
// Use for almost-noop steps (echo, validation gates, manifest dump).
func ResourcesTiny() *ResourceRequirements {
	return &ResourceRequirements{
		Requests: ResourceList{CPU: "50m", Memory: "32Mi"},
		Limits:   ResourceList{CPU: "100m", Memory: "64Mi"},
	}
}

// ResourcesSmall returns a small preset: 100m/128Mi requests, 500m/256Mi limits.
// Use for short-lived CLI steps (aws s3, syft, cosign, lint).
func ResourcesSmall() *ResourceRequirements {
	return &ResourceRequirements{
		Requests: ResourceList{CPU: "100m", Memory: "128Mi"},
		Limits:   ResourceList{CPU: "500m", Memory: "256Mi"},
	}
}

// ResourcesMedium returns a medium preset: 500m/512Mi requests, 2000m/2Gi limits.
// Use for build steps (npm install, go build, BuildKit) that need
// burstable CPU and meaningful memory.
func ResourcesMedium() *ResourceRequirements {
	return &ResourceRequirements{
		Requests: ResourceList{CPU: "500m", Memory: "512Mi"},
		Limits:   ResourceList{CPU: "2000m", Memory: "2Gi"},
	}
}
