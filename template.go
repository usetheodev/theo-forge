package forge

// ImagePullPolicy defines when to pull the image.
type ImagePullPolicy string

const (
	ImagePullAlways       ImagePullPolicy = "Always"
	ImagePullNever        ImagePullPolicy = "Never"
	ImagePullIfNotPresent ImagePullPolicy = "IfNotPresent"
)

// ParseImagePullPolicy normalizes a string to an ImagePullPolicy.
func ParseImagePullPolicy(s string) (ImagePullPolicy, error) {
	switch s {
	case "Always", "always":
		return ImagePullAlways, nil
	case "Never", "never":
		return ImagePullNever, nil
	case "IfNotPresent", "ifNotPresent", "if_not_present":
		return ImagePullIfNotPresent, nil
	default:
		return "", &InvalidType{Expected: "Always|Never|IfNotPresent", Got: s}
	}
}

// ResourceRequirements specifies CPU/memory requests and limits.
type ResourceRequirements struct {
	Requests ResourceList `json:"requests,omitempty" yaml:"requests,omitempty"`
	Limits   ResourceList `json:"limits,omitempty" yaml:"limits,omitempty"`
}

// ResourceList is a map of resource name to quantity.
type ResourceList struct {
	CPU              string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory           string `json:"memory,omitempty" yaml:"memory,omitempty"`
	EphemeralStorage string `json:"ephemeral-storage,omitempty" yaml:"ephemeral-storage,omitempty"`
}

// Templatable is implemented by types that can build an Argo Template.
type Templatable interface {
	BuildTemplate() (TemplateModel, error)
	GetName() string
}

// TemplateModel is the serializable Argo Workflows Template.
type TemplateModel struct {
	Name               string              `json:"name" yaml:"name"`
	Container          *ContainerModel     `json:"container,omitempty" yaml:"container,omitempty"`
	Script             *ScriptModel        `json:"script,omitempty" yaml:"script,omitempty"`
	DAG                *DAGModel           `json:"dag,omitempty" yaml:"dag,omitempty"`
	Steps              [][]StepModel       `json:"steps,omitempty" yaml:"steps,omitempty"`
	Resource           *ResourceTplModel   `json:"resource,omitempty" yaml:"resource,omitempty"`
	Suspend            *SuspendModel       `json:"suspend,omitempty" yaml:"suspend,omitempty"`
	HTTP               *HTTPModel          `json:"http,omitempty" yaml:"http,omitempty"`
	ContainerSet       *ContainerSetModel  `json:"containerSet,omitempty" yaml:"containerSet,omitempty"`
	Inputs             *InputsModel        `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs            *OutputsModel       `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Metadata           *MetadataModel      `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Timeout            string              `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	ActiveDeadlineSeconds *int             `json:"activeDeadlineSeconds,omitempty" yaml:"activeDeadlineSeconds,omitempty"`
	RetryStrategy      *RetryStrategyModel `json:"retryStrategy,omitempty" yaml:"retryStrategy,omitempty"`
	Parallelism        *int                `json:"parallelism,omitempty" yaml:"parallelism,omitempty"`
	FailFast           *bool               `json:"failFast,omitempty" yaml:"failFast,omitempty"`
	ServiceAccountName string              `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	Volumes            []VolumeModel       `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Metrics            *MetricsModel       `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	NodeSelector       map[string]string   `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	Tolerations        []Toleration        `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	InitContainers     []ContainerModel    `json:"initContainers,omitempty" yaml:"initContainers,omitempty"`
	Sidecars           []ContainerModel    `json:"sidecars,omitempty" yaml:"sidecars,omitempty"`
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

// Toleration is a K8s toleration.
type Toleration struct {
	Key      string `json:"key,omitempty" yaml:"key,omitempty"`
	Operator string `json:"operator,omitempty" yaml:"operator,omitempty"`
	Value    string `json:"value,omitempty" yaml:"value,omitempty"`
	Effect   string `json:"effect,omitempty" yaml:"effect,omitempty"`
}

// SuspendModel suspends workflow execution.
type SuspendModel struct {
	Duration string `json:"duration,omitempty" yaml:"duration,omitempty"`
}

// ResourceTplModel is a resource template (create/apply K8s resources).
type ResourceTplModel struct {
	Action           string `json:"action" yaml:"action"`
	Manifest         string `json:"manifest" yaml:"manifest"`
	SuccessCondition string `json:"successCondition,omitempty" yaml:"successCondition,omitempty"`
	FailureCondition string `json:"failureCondition,omitempty" yaml:"failureCondition,omitempty"`
}

// ContainerModel is the serializable Argo Container.
type ContainerModel struct {
	Name            string               `json:"name,omitempty" yaml:"name,omitempty"`
	Image           string               `json:"image" yaml:"image"`
	Command         []string             `json:"command,omitempty" yaml:"command,omitempty"`
	Args            []string             `json:"args,omitempty" yaml:"args,omitempty"`
	WorkingDir      string               `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	Env             []EnvVarModel        `json:"env,omitempty" yaml:"env,omitempty"`
	Resources       *ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
	VolumeMounts    []VolumeMountModel   `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
	ImagePullPolicy string               `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	Ports           []ContainerPort      `json:"ports,omitempty" yaml:"ports,omitempty"`
}

// ContainerPort represents a network port in a container.
type ContainerPort struct {
	Name          string `json:"name,omitempty" yaml:"name,omitempty"`
	ContainerPort int32  `json:"containerPort" yaml:"containerPort"`
	Protocol      string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
}

// ScriptModel is the serializable Argo Script template.
type ScriptModel struct {
	Image           string               `json:"image" yaml:"image"`
	Command         []string             `json:"command,omitempty" yaml:"command,omitempty"`
	Args            []string             `json:"args,omitempty" yaml:"args,omitempty"`
	Source          string               `json:"source" yaml:"source"`
	WorkingDir      string               `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	Env             []EnvVarModel        `json:"env,omitempty" yaml:"env,omitempty"`
	Resources       *ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
	VolumeMounts    []VolumeMountModel   `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
	ImagePullPolicy string               `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
}
