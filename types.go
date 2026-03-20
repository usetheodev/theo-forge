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

type ImagePullPolicy = model.ImagePullPolicy
type ResourceRequirements = model.ResourceRequirements
type ResourceList = model.ResourceList
type Toleration = model.Toleration
type ContainerPort = model.ContainerPort
type AccessMode = model.AccessMode
type WorkflowStatus = model.WorkflowStatus
type RetryPolicy = model.RetryPolicy
type Backoff = model.Backoff
type Metric = model.Metric
type Label = model.Label
type Counter = model.Counter
type Gauge = model.Gauge
type ArchiveStrategy = model.ArchiveStrategy
type TTLStrategy = model.TTLStrategy
type PodGC = model.PodGC
type ContinueOn = model.ContinueOn
type TemplateRef = model.TemplateRef
type ValueFrom = model.ValueFrom

// Re-export constants
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
