package model

// TemplateModel is the serializable Argo Workflows Template.
type TemplateModel struct {
	Name                  string               `json:"name,omitempty" yaml:"name,omitempty"`
	Container             *ContainerModel      `json:"container,omitempty" yaml:"container,omitempty"`
	Script                *ScriptModel         `json:"script,omitempty" yaml:"script,omitempty"`
	DAG                   *DAGModel            `json:"dag,omitempty" yaml:"dag,omitempty"`
	Steps                 [][]StepModel        `json:"steps,omitempty" yaml:"steps,omitempty"`
	Resource              *ResourceTplModel    `json:"resource,omitempty" yaml:"resource,omitempty"`
	Suspend               *SuspendModel        `json:"suspend,omitempty" yaml:"suspend,omitempty"`
	HTTP                  *HTTPModel           `json:"http,omitempty" yaml:"http,omitempty"`
	ContainerSet          *ContainerSetModel   `json:"containerSet,omitempty" yaml:"containerSet,omitempty"`
	Data                  *DataModel           `json:"data,omitempty" yaml:"data,omitempty"`
	Inputs                *InputsModel         `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs               *OutputsModel        `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Metadata              *MetadataModel       `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Timeout               string               `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	ActiveDeadlineSeconds *int                 `json:"activeDeadlineSeconds,omitempty" yaml:"activeDeadlineSeconds,omitempty"`
	RetryStrategy         *RetryStrategyModel  `json:"retryStrategy,omitempty" yaml:"retryStrategy,omitempty"`
	Parallelism           *int                 `json:"parallelism,omitempty" yaml:"parallelism,omitempty"`
	FailFast              *bool                `json:"failFast,omitempty" yaml:"failFast,omitempty"`
	ServiceAccountName    string               `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	Volumes               []VolumeModel        `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Metrics               *MetricsModel        `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	NodeSelector          map[string]string    `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	Tolerations           []Toleration         `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	InitContainers        []ContainerModel     `json:"initContainers,omitempty" yaml:"initContainers,omitempty"`
	Sidecars              []ContainerModel     `json:"sidecars,omitempty" yaml:"sidecars,omitempty"`
	Daemon                *bool                `json:"daemon,omitempty" yaml:"daemon,omitempty"`
	Memoize               *MemoizeModel        `json:"memoize,omitempty" yaml:"memoize,omitempty"`
	Synchronization       *SynchronizationModel `json:"synchronization,omitempty" yaml:"synchronization,omitempty"`
	PodSpecPatch          string               `json:"podSpecPatch,omitempty" yaml:"podSpecPatch,omitempty"`
	Hooks                 map[string]LifecycleHook `json:"hooks,omitempty" yaml:"hooks,omitempty"`
	ArchiveLocation       *ArtifactLocation    `json:"archiveLocation,omitempty" yaml:"archiveLocation,omitempty"`
	Labels                map[string]string    `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations           map[string]string    `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// MemoizeModel caches template outputs.
type MemoizeModel struct {
	Key        string          `json:"key" yaml:"key"`
	MaxAge     string          `json:"maxAge" yaml:"maxAge"`
	Cache      *CacheModel     `json:"cache" yaml:"cache"`
}

// CacheModel references a config map for memoization cache.
type CacheModel struct {
	ConfigMap *ConfigMapKeyRef `json:"configMap" yaml:"configMap"`
}

// ArtifactLocation defines a default artifact repository location.
type ArtifactLocation struct {
	S3          *S3ArtifactModel    `json:"s3,omitempty" yaml:"s3,omitempty"`
	ArchiveLogs *bool               `json:"archiveLogs,omitempty" yaml:"archiveLogs,omitempty"`
}

// InputsModel is the serializable Argo Inputs.
type InputsModel struct {
	Parameters []ParameterModel `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Artifacts  []ArtifactModel  `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
}

// OutputsModel is the serializable Argo Outputs.
type OutputsModel struct {
	Parameters []ParameterModel `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Artifacts  []ArtifactModel  `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
}

// MetadataModel is metadata for a template.
type MetadataModel struct {
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// SuspendModel suspends workflow execution.
type SuspendModel struct {
	Duration string `json:"duration,omitempty" yaml:"duration,omitempty"`
}

// ResourceTplModel is a resource template (create/apply K8s resources).
type ResourceTplModel struct {
	Action            string   `json:"action" yaml:"action"`
	Manifest          string   `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	SuccessCondition  string   `json:"successCondition,omitempty" yaml:"successCondition,omitempty"`
	FailureCondition  string   `json:"failureCondition,omitempty" yaml:"failureCondition,omitempty"`
	Flags             []string `json:"flags,omitempty" yaml:"flags,omitempty"`
	SetOwnerReference *bool    `json:"setOwnerReference,omitempty" yaml:"setOwnerReference,omitempty"`
	MergeStrategy     string   `json:"mergeStrategy,omitempty" yaml:"mergeStrategy,omitempty"`
	ManifestFrom      *ManifestFrom `json:"manifestFrom,omitempty" yaml:"manifestFrom,omitempty"`
}

// ManifestFrom references a manifest from an artifact.
type ManifestFrom struct {
	Artifact *ArtifactModel `json:"artifact,omitempty" yaml:"artifact,omitempty"`
}

// ContainerModel is the serializable Argo Container.
type ContainerModel struct {
	Name            string                `json:"name,omitempty" yaml:"name,omitempty"`
	Image           string                `json:"image" yaml:"image"`
	Command         []string              `json:"command,omitempty" yaml:"command,omitempty"`
	Args            []string              `json:"args,omitempty" yaml:"args,omitempty"`
	WorkingDir      string                `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	Env             []EnvVarModel         `json:"env,omitempty" yaml:"env,omitempty"`
	EnvFrom         []EnvFromSource       `json:"envFrom,omitempty" yaml:"envFrom,omitempty"`
	Resources       *ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
	VolumeMounts    []VolumeMountModel    `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
	ImagePullPolicy string                `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	Ports           []ContainerPort       `json:"ports,omitempty" yaml:"ports,omitempty"`
	SecurityContext *SecurityContext       `json:"securityContext,omitempty" yaml:"securityContext,omitempty"`
	Stdin           *bool                 `json:"stdin,omitempty" yaml:"stdin,omitempty"`
	Mirror          *bool                 `json:"mirrorVolumeMounts,omitempty" yaml:"mirrorVolumeMounts,omitempty"`
	Lifecycle       *Lifecycle            `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
	ReadinessProbe  *Probe                `json:"readinessProbe,omitempty" yaml:"readinessProbe,omitempty"`
	LivenessProbe   *Probe                `json:"livenessProbe,omitempty" yaml:"livenessProbe,omitempty"`
	Dependencies    []string              `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
}

// SecurityContext holds security configuration for a container.
type SecurityContext struct {
	RunAsUser                *int64 `json:"runAsUser,omitempty" yaml:"runAsUser,omitempty"`
	RunAsGroup               *int64 `json:"runAsGroup,omitempty" yaml:"runAsGroup,omitempty"`
	RunAsNonRoot             *bool  `json:"runAsNonRoot,omitempty" yaml:"runAsNonRoot,omitempty"`
	Privileged               *bool  `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	ReadOnlyRootFilesystem   *bool  `json:"readOnlyRootFilesystem,omitempty" yaml:"readOnlyRootFilesystem,omitempty"`
	AllowPrivilegeEscalation *bool  `json:"allowPrivilegeEscalation,omitempty" yaml:"allowPrivilegeEscalation,omitempty"`
}

// Lifecycle describes actions the management system should take in response to container lifecycle events.
type Lifecycle struct {
	PreStop *LifecycleHandler `json:"preStop,omitempty" yaml:"preStop,omitempty"`
}

// LifecycleHandler defines a handler for a lifecycle event.
type LifecycleHandler struct {
	Exec *ExecAction `json:"exec,omitempty" yaml:"exec,omitempty"`
}

// ExecAction describes a "run in container" action.
type ExecAction struct {
	Command []string `json:"command" yaml:"command"`
}

// Probe describes a health check to be performed against a container.
type Probe struct {
	HTTPGet             *HTTPGetAction `json:"httpGet,omitempty" yaml:"httpGet,omitempty"`
	TCPSocket           *TCPSocketAction `json:"tcpSocket,omitempty" yaml:"tcpSocket,omitempty"`
	Exec                *ExecAction    `json:"exec,omitempty" yaml:"exec,omitempty"`
	InitialDelaySeconds *int           `json:"initialDelaySeconds,omitempty" yaml:"initialDelaySeconds,omitempty"`
	PeriodSeconds       *int           `json:"periodSeconds,omitempty" yaml:"periodSeconds,omitempty"`
	SuccessThreshold    *int           `json:"successThreshold,omitempty" yaml:"successThreshold,omitempty"`
	FailureThreshold    *int           `json:"failureThreshold,omitempty" yaml:"failureThreshold,omitempty"`
	TimeoutSeconds      *int           `json:"timeoutSeconds,omitempty" yaml:"timeoutSeconds,omitempty"`
}

// HTTPGetAction describes an action based on HTTP Get requests.
type HTTPGetAction struct {
	Path   string `json:"path,omitempty" yaml:"path,omitempty"`
	Port   interface{} `json:"port" yaml:"port"`
	Scheme string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
}

// TCPSocketAction describes an action based on opening a socket.
type TCPSocketAction struct {
	Port interface{} `json:"port" yaml:"port"`
}

// EnvFromSource represents a source for environment variables.
type EnvFromSource struct {
	ConfigMapRef *ConfigMapEnvSource `json:"configMapRef,omitempty" yaml:"configMapRef,omitempty"`
	SecretRef    *SecretEnvSource    `json:"secretRef,omitempty" yaml:"secretRef,omitempty"`
	Prefix       string              `json:"prefix,omitempty" yaml:"prefix,omitempty"`
}

// ConfigMapEnvSource selects a ConfigMap to populate the environment variables with.
type ConfigMapEnvSource struct {
	Name     string `json:"name" yaml:"name"`
	Optional *bool  `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// SecretEnvSource selects a Secret to populate the environment variables with.
type SecretEnvSource struct {
	Name     string `json:"name" yaml:"name"`
	Optional *bool  `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// ScriptModel is the serializable Argo Script template.
type ScriptModel struct {
	Image           string                `json:"image" yaml:"image"`
	Command         []string              `json:"command,omitempty" yaml:"command,omitempty"`
	Args            []string              `json:"args,omitempty" yaml:"args,omitempty"`
	Source          string                `json:"source" yaml:"source"`
	WorkingDir      string                `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	Env             []EnvVarModel         `json:"env,omitempty" yaml:"env,omitempty"`
	Resources       *ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
	VolumeMounts    []VolumeMountModel    `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
	ImagePullPolicy string                `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
}

// ContainerSetModel is the serializable Argo ContainerSet.
type ContainerSetModel struct {
	Containers   []ContainerModel  `json:"containers" yaml:"containers"`
	VolumeMounts []VolumeMountModel `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
}

// HTTPModel is the serializable Argo HTTP template.
type HTTPModel struct {
	URL              string       `json:"url" yaml:"url"`
	Method           string       `json:"method,omitempty" yaml:"method,omitempty"`
	Headers          []HTTPHeader `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body             string       `json:"body,omitempty" yaml:"body,omitempty"`
	SuccessCondition string       `json:"successCondition,omitempty" yaml:"successCondition,omitempty"`
	TimeoutSeconds   *int         `json:"timeoutSeconds,omitempty" yaml:"timeoutSeconds,omitempty"`
}

// HTTPHeader is a single HTTP header.
type HTTPHeader struct {
	Name      string `json:"name" yaml:"name"`
	Value     string `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom *HTTPHeaderSource `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

// HTTPHeaderSource describes a source for an HTTP header value.
type HTTPHeaderSource struct {
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty" yaml:"secretKeyRef,omitempty"`
}

// DataModel is the serializable Argo Data template.
type DataModel struct {
	Source         DataSource           `json:"source" yaml:"source"`
	Transformation []TransformationStep `json:"transformation" yaml:"transformation"`
}

// DataSource defines the data source for a data template.
type DataSource struct {
	ArtifactPaths *ArtifactModel `json:"artifactPaths,omitempty" yaml:"artifactPaths,omitempty"`
}

// TransformationStep defines a transformation step for a data template.
type TransformationStep struct {
	Expression string `json:"expression" yaml:"expression"`
}
