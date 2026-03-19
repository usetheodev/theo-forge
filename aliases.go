package forge

import (
	"github.com/usetheo/theo/forge/client"
	"github.com/usetheo/theo/forge/expr"
	"github.com/usetheo/theo/forge/model"
)

// Type aliases for commonly used model types.
// These allow users to reference model types without importing the model package.

// --- Primitive/K8s types ---

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
type Histogram = model.Histogram
type ArchiveStrategy = model.ArchiveStrategy
type SecretKeySelector = model.SecretKeySelector
type TTLStrategy = model.TTLStrategy
type PodGC = model.PodGC
type ContinueOn = model.ContinueOn
type TemplateRef = model.TemplateRef
type ValueFrom = model.ValueFrom

// --- Serializable model types used by builders ---

type TemplateModel = model.TemplateModel
type InputsModel = model.InputsModel
type OutputsModel = model.OutputsModel
type MetadataModel = model.MetadataModel
type MetricsModel = model.MetricsModel
type ContainerModel = model.ContainerModel
type ScriptModel = model.ScriptModel
type SuspendModel = model.SuspendModel
type ResourceTplModel = model.ResourceTplModel
type RetryStrategyModel = model.RetryStrategyModel
type ParameterModel = model.ParameterModel
type ArtifactModel = model.ArtifactModel
type EnvVarModel = model.EnvVarModel
type EnvVarSource = model.EnvVarSource
type KeySelector = model.KeySelector
type FieldSelector = model.FieldSelector
type ResourceFieldSelector = model.ResourceFieldSelector
type VolumeMountModel = model.VolumeMountModel
type VolumeModel = model.VolumeModel
type WorkflowModel = model.WorkflowModel
type WorkflowSpec = model.WorkflowSpec
type WorkflowMetadata = model.WorkflowMetadata
type WorkflowTemplateModel = model.WorkflowTemplateModel
type ArgumentsModel = model.ArgumentsModel
type DAGModel = model.DAGModel
type DAGTaskModel = model.DAGTaskModel
type StepModel = model.StepModel
type ContainerSetModel = model.ContainerSetModel
type HTTPModel = model.HTTPModel
type HTTPHeader = model.HTTPHeader
type ImagePullSecret = model.ImagePullSecret
type PVCModel = model.PVCModel
type PVCSpec = model.PVCSpec
type PVCResources = model.PVCResources
type PVCResourceRequest = model.PVCResourceRequest
type EmptyDirVolumeModel = model.EmptyDirVolumeModel
type HostPathVolumeModel = model.HostPathVolumeModel
type ConfigMapVolumeModel = model.ConfigMapVolumeModel
type SecretVolumeModel = model.SecretVolumeModel
type PersistentVolumeClaimVolRef = model.PersistentVolumeClaimVolRef
type NFSVolumeModel = model.NFSVolumeModel
type DownwardAPIVolumeModel = model.DownwardAPIVolumeModel
type ProjectedVolumeModel = model.ProjectedVolumeModel
type S3ArtifactModel = model.S3ArtifactModel
type GCSArtifactModel = model.GCSArtifactModel
type HTTPArtifactModel = model.HTTPArtifactModel
type GitArtifactModel = model.GitArtifactModel
type RawArtifactModel = model.RawArtifactModel
type AzureArtifactModel = model.AzureArtifactModel
type OSSArtifactModel = model.OSSArtifactModel
type HDFSArtifactModel = model.HDFSArtifactModel
type CronWorkflowModel = model.CronWorkflowModel
type CronWorkflowSpec = model.CronWorkflowSpec

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

// --- Client types ---

type WorkflowsService = client.WorkflowsService
type APIError = client.APIError
type Buildable = client.Buildable
type HTTPClient = client.HTTPClient
type ListWorkflowsResponse = client.ListWorkflowsResponse

// --- Model error types ---

type InvalidType = model.InvalidType

// --- Expression types ---

type Expr = expr.Expr

// Re-export functions
var (
	ParseImagePullPolicy = model.ParseImagePullPolicy
	ParseWorkflowStatus  = model.ParseWorkflowStatus

	// Client constructors
	NewWorkflowsService = client.NewWorkflowsService

	// Expression constructors
	E                 = expr.E
	C                 = expr.C
	Tasks             = expr.Tasks
	StepsExpr  = expr.Steps
	Inputs     = expr.Inputs
	OutputsExpr = expr.Outputs
	Item              = expr.Item
	WorkflowExpr      = expr.Workflow
	ParamRef          = expr.ParamRef
	InputParam        = expr.InputParam
	TaskOutputParam   = expr.TaskOutputParam
	StepOutputParam   = expr.StepOutputParam
	TaskOutputResult  = expr.TaskOutputResult
	StepOutputResult  = expr.StepOutputResult
	WorkflowParam     = expr.WorkflowParam
	Concat            = expr.Concat
)

// Sprig expression namespace re-exported from expr.
var Sprig = expr.Sprig
