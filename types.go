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
