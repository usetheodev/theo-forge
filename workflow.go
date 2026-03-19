package forge

import (
	"encoding/json"
	"fmt"

	"github.com/usetheo/theo/forge/model"
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
	Tolerations []model.Toleration
	// Suspend starts the workflow in a suspended state.
	Suspend *bool
	// HostNetwork enables host networking.
	HostNetwork *bool
	// TTLStrategy defines CRD retention.
	TTLStrategy *model.TTLStrategy
	// PodGC defines pod cleanup.
	PodGC *model.PodGC
	// Priority sets the workflow priority.
	Priority *int
	// OnExit is the exit handler template name.
	OnExit string
	// Metrics for the workflow.
	Metrics []model.Metric
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

func (w *Workflow) buildArguments() (*model.ArgumentsModel, error) {
	if len(w.Arguments) == 0 && len(w.ArgumentArtifacts) == 0 {
		return nil, nil
	}
	args := &model.ArgumentsModel{}
	for _, p := range w.Arguments {
		m, err := p.AsArgument()
		if err != nil {
			return nil, fmt.Errorf("argument %q: %w", p.Name, err)
		}
		args.Parameters = append(args.Parameters, m)
	}
	for _, a := range w.ArgumentArtifacts {
		m, err := a.Build()
		if err != nil {
			return nil, fmt.Errorf("argument artifact: %w", err)
		}
		args.Artifacts = append(args.Artifacts, m)
	}
	return args, nil
}

func (w *Workflow) buildVolumes() ([]model.VolumeModel, error) {
	if len(w.Volumes) == 0 {
		return nil, nil
	}
	vols := make([]model.VolumeModel, 0, len(w.Volumes))
	for _, v := range w.Volumes {
		m, err := v.BuildVolume()
		if err != nil {
			return nil, fmt.Errorf("volume: %w", err)
		}
		vols = append(vols, m)
	}
	if len(vols) == 0 {
		return nil, nil
	}
	return vols, nil
}

func (w *Workflow) buildVolumeClaimTemplates() ([]model.PVCModel, error) {
	if len(w.VolumeClaimTemplates) == 0 {
		return nil, nil
	}
	pvcs := make([]model.PVCModel, 0, len(w.VolumeClaimTemplates))
	for _, v := range w.VolumeClaimTemplates {
		m, err := v.BuildPVC()
		if err != nil {
			return nil, fmt.Errorf("volume claim template: %w", err)
		}
		pvcs = append(pvcs, m)
	}
	if len(pvcs) == 0 {
		return nil, nil
	}
	return pvcs, nil
}

func (w *Workflow) buildMetrics() *model.MetricsModel {
	if len(w.Metrics) == 0 {
		return nil
	}
	return &model.MetricsModel{Prometheus: w.Metrics}
}

func (w *Workflow) buildImagePullSecrets() []model.ImagePullSecret {
	if len(w.ImagePullSecrets) == 0 {
		return nil
	}
	secrets := make([]model.ImagePullSecret, len(w.ImagePullSecrets))
	for i, s := range w.ImagePullSecrets {
		secrets[i] = model.ImagePullSecret{Name: s}
	}
	return secrets
}

// GetNamespace returns the workflow namespace.
func (w *Workflow) GetNamespace() string {
	return w.Namespace
}

// Build converts the Workflow to its serializable model.
func (w *Workflow) Build() (model.WorkflowModel, error) {
	if err := w.validate(); err != nil {
		return model.WorkflowModel{}, err
	}

	apiVersion := w.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}
	kind := w.Kind
	if kind == "" {
		kind = DefaultKind
	}

	templates := make([]model.TemplateModel, 0, len(w.Templates))
	for _, t := range w.Templates {
		m, err := t.BuildTemplate()
		if err != nil {
			return model.WorkflowModel{}, fmt.Errorf("template %q: %w", t.GetName(), err)
		}
		templates = append(templates, m)
	}

	args, err := w.buildArguments()
	if err != nil {
		return model.WorkflowModel{}, err
	}

	vols, err := w.buildVolumes()
	if err != nil {
		return model.WorkflowModel{}, err
	}

	pvcs, err := w.buildVolumeClaimTemplates()
	if err != nil {
		return model.WorkflowModel{}, err
	}

	var rs *model.RetryStrategyModel
	if w.RetryStrategy != nil {
		m := w.RetryStrategy.Build()
		rs = &m
	}

	return model.WorkflowModel{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: model.WorkflowMetadata{
			Name:         w.Name,
			GenerateName: w.GenerateName,
			Namespace:    w.Namespace,
			Labels:       w.Labels,
			Annotations:  w.Annotations,
		},
		Spec: model.WorkflowSpec{
			Entrypoint:            w.Entrypoint,
			Templates:             templates,
			Arguments:             args,
			Volumes:               vols,
			VolumeClaimTemplates:  pvcs,
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
	m, err := w.Build()
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(m)
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
	m, err := w.Build()
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToYAML converts the workflow to a YAML string.
func (w *Workflow) ToYAML() (string, error) {
	m, err := w.Build()
	if err != nil {
		return "", err
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromYAML creates a WorkflowModel from a YAML string.
func FromYAML(yamlStr string) (model.WorkflowModel, error) {
	var m model.WorkflowModel
	if err := yaml.Unmarshal([]byte(yamlStr), &m); err != nil {
		return model.WorkflowModel{}, err
	}
	return m, nil
}

// FromJSON creates a WorkflowModel from a JSON string.
func FromJSON(jsonStr string) (model.WorkflowModel, error) {
	var m model.WorkflowModel
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return model.WorkflowModel{}, err
	}
	return m, nil
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
