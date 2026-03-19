package forge

import "github.com/usetheo/theo/forge/model"

// Type aliases for commonly used model types.
// These allow users to reference frequently-used Kubernetes types
// without importing the model package directly.

// --- Primitive/K8s types (user-facing) ---

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
