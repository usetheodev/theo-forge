package forge

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

const (
	// NameLimit is the maximum length of a workflow name.
	NameLimit = 63
	// DefaultAPIVersion is the default Argo Workflows API version.
	DefaultAPIVersion = "argoproj.io/v1alpha1"
	// DefaultKind is the default resource kind.
	DefaultKind = "Workflow"
)

// WorkflowModel is the serializable Argo Workflow.
type WorkflowModel struct {
	APIVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Kind       string            `json:"kind" yaml:"kind"`
	Metadata   WorkflowMetadata  `json:"metadata" yaml:"metadata"`
	Spec       WorkflowSpec      `json:"spec" yaml:"spec"`
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
	Entrypoint          string               `json:"entrypoint" yaml:"entrypoint"`
	Templates           []TemplateModel      `json:"templates" yaml:"templates"`
	Arguments           *ArgumentsModel      `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	Volumes             []VolumeModel        `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	VolumeClaimTemplates []PVCModel          `json:"volumeClaimTemplates,omitempty" yaml:"volumeClaimTemplates,omitempty"`
	ServiceAccountName  string               `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	Parallelism         *int                 `json:"parallelism,omitempty" yaml:"parallelism,omitempty"`
	ActiveDeadlineSeconds *int               `json:"activeDeadlineSeconds,omitempty" yaml:"activeDeadlineSeconds,omitempty"`
	NodeSelector        map[string]string    `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	Tolerations         []Toleration         `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	Suspend             *bool                `json:"suspend,omitempty" yaml:"suspend,omitempty"`
	HostNetwork         *bool                `json:"hostNetwork,omitempty" yaml:"hostNetwork,omitempty"`
	TTLStrategy         *TTLStrategy         `json:"ttlStrategy,omitempty" yaml:"ttlStrategy,omitempty"`
	PodGC               *PodGC               `json:"podGC,omitempty" yaml:"podGC,omitempty"`
	Priority            *int                 `json:"priority,omitempty" yaml:"priority,omitempty"`
	OnExit              string               `json:"onExit,omitempty" yaml:"onExit,omitempty"`
	Metrics             *MetricsModel        `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	ArchiveLogs         *bool                `json:"archiveLogs,omitempty" yaml:"archiveLogs,omitempty"`
	RetryStrategy       *RetryStrategyModel  `json:"retryStrategy,omitempty" yaml:"retryStrategy,omitempty"`
	ImagePullSecrets    []ImagePullSecret    `json:"imagePullSecrets,omitempty" yaml:"imagePullSecrets,omitempty"`
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

// ImagePullSecret references a K8s secret for pulling images.
type ImagePullSecret struct {
	Name string `json:"name" yaml:"name"`
}

// Workflow represents an Argo Workflow.
type Workflow struct {
	// Name is the workflow name (max 63 chars).
	Name string
	// GenerateName is the name prefix for auto-generation.
	GenerateName string
	// Namespace is the K8s namespace.
	Namespace string
	// APIVersion is the API version (default: argoproj.io/v1alpha1).
	APIVersion string
	// Kind is the resource kind (default: Workflow).
	Kind string
	// Entrypoint is the starting template name.
	Entrypoint string
	// Templates are the workflow templates.
	Templates []Templatable
	// Arguments are the workflow-level arguments.
	Arguments []Parameter
	// ArgumentArtifacts are workflow-level artifact arguments.
	ArgumentArtifacts []ArtifactBuilder
	// Volumes are the workflow-level volumes.
	Volumes []VolumeBuilder
	// VolumeClaimTemplates are PVCs for dynamic provisioning.
	VolumeClaimTemplates []PVCVolume
	// Labels for the workflow.
	Labels map[string]string
	// Annotations for the workflow.
	Annotations map[string]string
	// ServiceAccountName for the workflow.
	ServiceAccountName string
	// Parallelism limits the max concurrent pods.
	Parallelism *int
	// ActiveDeadlineSeconds kills the workflow after X seconds.
	ActiveDeadlineSeconds *int
	// NodeSelector constrains pod scheduling.
	NodeSelector map[string]string
	// Tolerations for pod scheduling.
	Tolerations []Toleration
	// Suspend starts the workflow in a suspended state.
	Suspend *bool
	// HostNetwork enables host networking.
	HostNetwork *bool
	// TTLStrategy defines CRD retention.
	TTLStrategy *TTLStrategy
	// PodGC defines pod cleanup.
	PodGC *PodGC
	// Priority sets the workflow priority.
	Priority *int
	// OnExit is the exit handler template name.
	OnExit string
	// Metrics for the workflow.
	Metrics []Metric
	// ArchiveLogs enables log archiving.
	ArchiveLogs *bool
	// RetryStrategy is the workflow-level retry strategy.
	RetryStrategy *RetryStrategy
	// ImagePullSecrets are secrets for pulling images.
	ImagePullSecrets []string
}

func (w *Workflow) validate() error {
	if w.Name != "" && len(w.Name) > NameLimit {
		return fmt.Errorf("name must be no more than %d characters", NameLimit)
	}
	if w.GenerateName != "" && len(w.GenerateName) > NameLimit {
		return fmt.Errorf("generateName must be no more than %d characters", NameLimit)
	}
	if w.Name == "" && w.GenerateName == "" {
		return fmt.Errorf("either name or generateName must be set")
	}
	return nil
}

func (w *Workflow) buildArguments() *ArgumentsModel {
	if len(w.Arguments) == 0 && len(w.ArgumentArtifacts) == 0 {
		return nil
	}
	args := &ArgumentsModel{}
	for _, p := range w.Arguments {
		m, err := p.AsArgument()
		if err != nil {
			continue
		}
		args.Parameters = append(args.Parameters, m)
	}
	for _, a := range w.ArgumentArtifacts {
		m, err := a.Build()
		if err != nil {
			continue
		}
		args.Artifacts = append(args.Artifacts, m)
	}
	return args
}

func (w *Workflow) buildVolumes() []VolumeModel {
	if len(w.Volumes) == 0 {
		return nil
	}
	vols := make([]VolumeModel, 0, len(w.Volumes))
	for _, v := range w.Volumes {
		m, err := v.BuildVolume()
		if err != nil {
			continue
		}
		vols = append(vols, m)
	}
	if len(vols) == 0 {
		return nil
	}
	return vols
}

func (w *Workflow) buildVolumeClaimTemplates() []PVCModel {
	if len(w.VolumeClaimTemplates) == 0 {
		return nil
	}
	pvcs := make([]PVCModel, 0, len(w.VolumeClaimTemplates))
	for _, v := range w.VolumeClaimTemplates {
		m, err := v.BuildPVC()
		if err != nil {
			continue
		}
		pvcs = append(pvcs, m)
	}
	if len(pvcs) == 0 {
		return nil
	}
	return pvcs
}

func (w *Workflow) buildMetrics() *MetricsModel {
	if len(w.Metrics) == 0 {
		return nil
	}
	return &MetricsModel{Prometheus: w.Metrics}
}

func (w *Workflow) buildImagePullSecrets() []ImagePullSecret {
	if len(w.ImagePullSecrets) == 0 {
		return nil
	}
	secrets := make([]ImagePullSecret, len(w.ImagePullSecrets))
	for i, s := range w.ImagePullSecrets {
		secrets[i] = ImagePullSecret{Name: s}
	}
	return secrets
}

// Build converts the Workflow to its serializable model.
func (w *Workflow) Build() (WorkflowModel, error) {
	if err := w.validate(); err != nil {
		return WorkflowModel{}, err
	}

	apiVersion := w.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}
	kind := w.Kind
	if kind == "" {
		kind = DefaultKind
	}

	templates := make([]TemplateModel, 0, len(w.Templates))
	for _, t := range w.Templates {
		m, err := t.BuildTemplate()
		if err != nil {
			return WorkflowModel{}, fmt.Errorf("template %q: %w", t.GetName(), err)
		}
		templates = append(templates, m)
	}

	var rs *RetryStrategyModel
	if w.RetryStrategy != nil {
		m := w.RetryStrategy.Build()
		rs = &m
	}

	return WorkflowModel{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: WorkflowMetadata{
			Name:         w.Name,
			GenerateName: w.GenerateName,
			Namespace:    w.Namespace,
			Labels:       w.Labels,
			Annotations:  w.Annotations,
		},
		Spec: WorkflowSpec{
			Entrypoint:            w.Entrypoint,
			Templates:             templates,
			Arguments:             w.buildArguments(),
			Volumes:               w.buildVolumes(),
			VolumeClaimTemplates:  w.buildVolumeClaimTemplates(),
			ServiceAccountName:    w.ServiceAccountName,
			Parallelism:           w.Parallelism,
			ActiveDeadlineSeconds: w.ActiveDeadlineSeconds,
			NodeSelector:          w.NodeSelector,
			Tolerations:           w.Tolerations,
			Suspend:               w.Suspend,
			HostNetwork:           w.HostNetwork,
			TTLStrategy:           w.TTLStrategy,
			PodGC:                 w.PodGC,
			Priority:              w.Priority,
			OnExit:                w.OnExit,
			Metrics:               w.buildMetrics(),
			ArchiveLogs:           w.ArchiveLogs,
			RetryStrategy:         rs,
			ImagePullSecrets:      w.buildImagePullSecrets(),
		},
	}, nil
}

// ToDict converts the workflow to a map (via JSON round-trip).
func (w *Workflow) ToDict() (map[string]interface{}, error) {
	model, err := w.Build()
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ToJSON converts the workflow to a JSON string.
func (w *Workflow) ToJSON() (string, error) {
	model, err := w.Build()
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToYAML converts the workflow to a YAML string.
func (w *Workflow) ToYAML() (string, error) {
	model, err := w.Build()
	if err != nil {
		return "", err
	}
	data, err := yaml.Marshal(model)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromYAML creates a WorkflowModel from a YAML string.
func FromYAML(yamlStr string) (WorkflowModel, error) {
	var model WorkflowModel
	if err := yaml.Unmarshal([]byte(yamlStr), &model); err != nil {
		return WorkflowModel{}, err
	}
	return model, nil
}

// FromJSON creates a WorkflowModel from a JSON string.
func FromJSON(jsonStr string) (WorkflowModel, error) {
	var model WorkflowModel
	if err := json.Unmarshal([]byte(jsonStr), &model); err != nil {
		return WorkflowModel{}, err
	}
	return model, nil
}

// GetParameter retrieves a parameter from the workflow arguments by name.
func (w *Workflow) GetParameter(name string) (Parameter, error) {
	for _, p := range w.Arguments {
		if p.Name == name {
			return p, nil
		}
	}
	return Parameter{}, fmt.Errorf("parameter %q not found in workflow arguments", name)
}
