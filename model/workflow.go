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
	Entrypoint            string               `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	Templates             []TemplateModel      `json:"templates,omitempty" yaml:"templates,omitempty"`
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
	PodSpecPatch          string               `json:"podSpecPatch,omitempty" yaml:"podSpecPatch,omitempty"`
	Synchronization       *SynchronizationModel `json:"synchronization,omitempty" yaml:"synchronization,omitempty"`
	Hooks                 map[string]LifecycleHook `json:"hooks,omitempty" yaml:"hooks,omitempty"`
	DNSConfig             *DNSConfig           `json:"dnsConfig,omitempty" yaml:"dnsConfig,omitempty"`
	DNSPolicy             string               `json:"dnsPolicy,omitempty" yaml:"dnsPolicy,omitempty"`
	PodDisruptionBudget   *PodDisruptionBudget `json:"podDisruptionBudget,omitempty" yaml:"podDisruptionBudget,omitempty"`
	PodMetadata           *MetadataModel       `json:"podMetadata,omitempty" yaml:"podMetadata,omitempty"`
	SecurityContext       *PodSecurityContext   `json:"securityContext,omitempty" yaml:"securityContext,omitempty"`
	AutomountServiceAccountToken *bool          `json:"automountServiceAccountToken,omitempty" yaml:"automountServiceAccountToken,omitempty"`
	WorkflowMetadata      *WorkflowLevelMetadata `json:"workflowMetadata,omitempty" yaml:"workflowMetadata,omitempty"`
	WorkflowTemplateRef   *WorkflowTemplateRef `json:"workflowTemplateRef,omitempty" yaml:"workflowTemplateRef,omitempty"`
	ArtifactGC            *ArtifactGCStrategy  `json:"artifactGC,omitempty" yaml:"artifactGC,omitempty"`
	ArtifactRepositoryRef *ArtifactRepositoryRef `json:"artifactRepositoryRef,omitempty" yaml:"artifactRepositoryRef,omitempty"`
	TemplateDefaults      *TemplateDefaults    `json:"templateDefaults,omitempty" yaml:"templateDefaults,omitempty"`
}

// SynchronizationModel defines synchronization constraints.
type SynchronizationModel struct {
	// Legacy singular fields (Argo Workflows < 3.6)
	Mutex     *MutexModel     `json:"mutex,omitempty" yaml:"mutex,omitempty"`
	Semaphore *SemaphoreModel `json:"semaphore,omitempty" yaml:"semaphore,omitempty"`
	// New plural fields (Argo Workflows >= 3.6)
	Mutexes    []MutexModel    `json:"mutexes,omitempty" yaml:"mutexes,omitempty"`
	Semaphores []SemaphoreModel `json:"semaphores,omitempty" yaml:"semaphores,omitempty"`
}

// MutexModel is a mutual exclusion lock.
type MutexModel struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Database  *bool  `json:"database,omitempty" yaml:"database,omitempty"`
}

// SemaphoreModel is a counting semaphore.
type SemaphoreModel struct {
	ConfigMapKeyRef *ConfigMapKeyRef    `json:"configMapKeyRef,omitempty" yaml:"configMapKeyRef,omitempty"`
	Database        *SemaphoreDBRef     `json:"database,omitempty" yaml:"database,omitempty"`
}

// SemaphoreDBRef references a database-backed semaphore.
type SemaphoreDBRef struct {
	Key string `json:"key" yaml:"key"`
}

// ConfigMapKeyRef references a key in a ConfigMap.
type ConfigMapKeyRef struct {
	Name string `json:"name" yaml:"name"`
	Key  string `json:"key,omitempty" yaml:"key,omitempty"`
}

// LifecycleHook defines a hook that runs at a lifecycle event.
type LifecycleHook struct {
	Template   string          `json:"template,omitempty" yaml:"template,omitempty"`
	TemplateRef *TemplateRef   `json:"templateRef,omitempty" yaml:"templateRef,omitempty"`
	Arguments  *ArgumentsModel `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	Expression string          `json:"expression,omitempty" yaml:"expression,omitempty"`
}

// DNSConfig specifies the DNS parameters of a pod.
type DNSConfig struct {
	Nameservers []string          `json:"nameservers,omitempty" yaml:"nameservers,omitempty"`
	Searches    []string          `json:"searches,omitempty" yaml:"searches,omitempty"`
	Options     []DNSConfigOption `json:"options,omitempty" yaml:"options,omitempty"`
}

// DNSConfigOption is a DNS resolver option.
type DNSConfigOption struct {
	Name  string  `json:"name" yaml:"name"`
	Value *string `json:"value,omitempty" yaml:"value,omitempty"`
}

// PodDisruptionBudget configures the PDB for workflow pods.
type PodDisruptionBudget struct {
	MinAvailable   interface{} `json:"minAvailable,omitempty" yaml:"minAvailable,omitempty"`
	MaxUnavailable interface{} `json:"maxUnavailable,omitempty" yaml:"maxUnavailable,omitempty"`
}

// PodSecurityContext holds pod-level security attributes.
type PodSecurityContext struct {
	RunAsUser          *int64 `json:"runAsUser,omitempty" yaml:"runAsUser,omitempty"`
	RunAsGroup         *int64 `json:"runAsGroup,omitempty" yaml:"runAsGroup,omitempty"`
	RunAsNonRoot       *bool  `json:"runAsNonRoot,omitempty" yaml:"runAsNonRoot,omitempty"`
	FSGroup            *int64 `json:"fsGroup,omitempty" yaml:"fsGroup,omitempty"`
	SupplementalGroups []int64 `json:"supplementalGroups,omitempty" yaml:"supplementalGroups,omitempty"`
}

// WorkflowTemplateRef references a WorkflowTemplate by name.
type WorkflowTemplateRef struct {
	Name            string `json:"name" yaml:"name"`
	ClusterScope    bool   `json:"clusterScope,omitempty" yaml:"clusterScope,omitempty"`
}

// WorkflowLevelMetadata defines metadata applied to workflow pods at the spec level.
type WorkflowLevelMetadata struct {
	Labels      map[string]string       `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string       `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	LabelsFrom  map[string]LabelValueFrom `json:"labelsFrom,omitempty" yaml:"labelsFrom,omitempty"`
}

// LabelValueFrom describes how to set a label value from an expression.
type LabelValueFrom struct {
	Expression string `json:"expression" yaml:"expression"`
}

// ArtifactGCStrategy defines artifact garbage collection at the workflow level.
type ArtifactGCStrategy struct {
	Strategy              string              `json:"strategy,omitempty" yaml:"strategy,omitempty"`
	ServiceAccountName    string              `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	PodSpecPatch          string              `json:"podSpecPatch,omitempty" yaml:"podSpecPatch,omitempty"`
	ForceFinalizerRemoval bool                `json:"forceFinalizerRemoval,omitempty" yaml:"forceFinalizerRemoval,omitempty"`
}

// ArtifactRepositoryRef references an artifact repository config.
type ArtifactRepositoryRef struct {
	ConfigMap string `json:"configMap,omitempty" yaml:"configMap,omitempty"`
	Key       string `json:"key,omitempty" yaml:"key,omitempty"`
}

// TemplateDefaults defines default values applied to all templates.
type TemplateDefaults struct {
	Timeout               string               `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	RetryStrategy         *RetryStrategyModel  `json:"retryStrategy,omitempty" yaml:"retryStrategy,omitempty"`
	ActiveDeadlineSeconds *int                 `json:"activeDeadlineSeconds,omitempty" yaml:"activeDeadlineSeconds,omitempty"`
	ServiceAccountName    string               `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	Metadata              *MetadataModel       `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// TTLStrategy defines how long the workflow CRD persists after completion.
type TTLStrategy struct {
	SecondsAfterCompletion *int `json:"secondsAfterCompletion,omitempty" yaml:"secondsAfterCompletion,omitempty"`
	SecondsAfterSuccess    *int `json:"secondsAfterSuccess,omitempty" yaml:"secondsAfterSuccess,omitempty"`
	SecondsAfterFailure    *int `json:"secondsAfterFailure,omitempty" yaml:"secondsAfterFailure,omitempty"`
}

// PodGC defines the pod garbage collection strategy.
type PodGC struct {
	Strategy        string               `json:"strategy" yaml:"strategy"`
	LabelSelector   *LabelSelector       `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
	DeleteDelayDuration string           `json:"deleteDelayDuration,omitempty" yaml:"deleteDelayDuration,omitempty"`
}

// LabelSelector is a K8s label selector.
type LabelSelector struct {
	MatchLabels      map[string]string        `json:"matchLabels,omitempty" yaml:"matchLabels,omitempty"`
	MatchExpressions []LabelSelectorRequirement `json:"matchExpressions,omitempty" yaml:"matchExpressions,omitempty"`
}

// LabelSelectorRequirement is a single requirement on a label.
type LabelSelectorRequirement struct {
	Key      string   `json:"key" yaml:"key"`
	Operator string   `json:"operator" yaml:"operator"`
	Values   []string `json:"values,omitempty" yaml:"values,omitempty"`
}

// WorkflowTemplateModel is the serializable Argo WorkflowTemplate.
type WorkflowTemplateModel struct {
	APIVersion string           `json:"apiVersion" yaml:"apiVersion"`
	Kind       string           `json:"kind" yaml:"kind"`
	Metadata   WorkflowMetadata `json:"metadata" yaml:"metadata"`
	Spec       WorkflowSpec     `json:"spec" yaml:"spec"`
}
