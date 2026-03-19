package model

// WorkflowModel is the serializable Argo Workflow.
type WorkflowModel struct {
	APIVersion string           `json:"apiVersion" yaml:"apiVersion"`
	Kind       string           `json:"kind" yaml:"kind"`
	Metadata   WorkflowMetadata `json:"metadata" yaml:"metadata"`
	Spec       WorkflowSpec     `json:"spec" yaml:"spec"`
}

// WorkflowMetadata is the metadata for a workflow.
type WorkflowMetadata struct {
	Name         string            `json:"name,omitempty" yaml:"name,omitempty"`
	GenerateName string            `json:"generateName,omitempty" yaml:"generateName,omitempty"`
	Namespace    string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// WorkflowSpec is the spec for a workflow.
type WorkflowSpec struct {
	Entrypoint            string               `json:"entrypoint" yaml:"entrypoint"`
	Templates             []TemplateModel      `json:"templates" yaml:"templates"`
	Arguments             *ArgumentsModel      `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	Volumes               []VolumeModel        `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	VolumeClaimTemplates  []PVCModel           `json:"volumeClaimTemplates,omitempty" yaml:"volumeClaimTemplates,omitempty"`
	ServiceAccountName    string               `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	Parallelism           *int                 `json:"parallelism,omitempty" yaml:"parallelism,omitempty"`
	ActiveDeadlineSeconds *int                 `json:"activeDeadlineSeconds,omitempty" yaml:"activeDeadlineSeconds,omitempty"`
	NodeSelector          map[string]string    `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	Tolerations           []Toleration         `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	Suspend               *bool                `json:"suspend,omitempty" yaml:"suspend,omitempty"`
	HostNetwork           *bool                `json:"hostNetwork,omitempty" yaml:"hostNetwork,omitempty"`
	TTLStrategy           *TTLStrategy         `json:"ttlStrategy,omitempty" yaml:"ttlStrategy,omitempty"`
	PodGC                 *PodGC               `json:"podGC,omitempty" yaml:"podGC,omitempty"`
	Priority              *int                 `json:"priority,omitempty" yaml:"priority,omitempty"`
	OnExit                string               `json:"onExit,omitempty" yaml:"onExit,omitempty"`
	Metrics               *MetricsModel        `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	ArchiveLogs           *bool                `json:"archiveLogs,omitempty" yaml:"archiveLogs,omitempty"`
	RetryStrategy         *RetryStrategyModel  `json:"retryStrategy,omitempty" yaml:"retryStrategy,omitempty"`
	ImagePullSecrets      []ImagePullSecret    `json:"imagePullSecrets,omitempty" yaml:"imagePullSecrets,omitempty"`
}

// TTLStrategy defines how long the workflow CRD persists after completion.
type TTLStrategy struct {
	SecondsAfterCompletion *int `json:"secondsAfterCompletion,omitempty" yaml:"secondsAfterCompletion,omitempty"`
	SecondsAfterSuccess    *int `json:"secondsAfterSuccess,omitempty" yaml:"secondsAfterSuccess,omitempty"`
	SecondsAfterFailure    *int `json:"secondsAfterFailure,omitempty" yaml:"secondsAfterFailure,omitempty"`
}

// PodGC defines the pod garbage collection strategy.
type PodGC struct {
	Strategy string `json:"strategy" yaml:"strategy"`
}

// WorkflowTemplateModel is the serializable Argo WorkflowTemplate.
type WorkflowTemplateModel struct {
	APIVersion string           `json:"apiVersion" yaml:"apiVersion"`
	Kind       string           `json:"kind" yaml:"kind"`
	Metadata   WorkflowMetadata `json:"metadata" yaml:"metadata"`
	Spec       WorkflowSpec     `json:"spec" yaml:"spec"`
}
