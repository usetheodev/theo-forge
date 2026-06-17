package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/config"
	"github.com/usetheodev/theo-forge/model"
	"github.com/usetheodev/theo-forge/serialize"
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
	// PodSpecPatch is a JSON/YAML patch applied to the pod spec.
	PodSpecPatch string
	// Synchronization configures synchronization constraints.
	Synchronization *model.SynchronizationModel
	// Hooks are lifecycle hooks.
	Hooks map[string]model.LifecycleHook
	// DNSConfig specifies DNS parameters of the pod.
	DNSConfig *model.DNSConfig
	// DNSPolicy sets DNS policy for the pod.
	DNSPolicy string
	// PodDisruptionBudget configures PDB for workflow pods.
	PodDisruptionBudget *model.PodDisruptionBudget
	// PodMetadata sets metadata on workflow pods.
	PodMetadata *model.MetadataModel
	// SecurityContext holds pod-level security attributes.
	SecurityContext *model.PodSecurityContext
	// Affinity defines scheduling constraints for workflow pods.
	// When nil and VolumeClaimTemplates is non-empty (and
	// DisableDefaultAffinity is false), Build() injects a default
	// podAffinity that co-locates all workflow pods on the same node.
	// See DefaultPodAffinityFor.
	Affinity *model.Affinity
	// DisableDefaultAffinity opts out of the default podAffinity
	// injection performed by Build() when VolumeClaimTemplates is
	// non-empty. Has no effect when Affinity is already set or when
	// VolumeClaimTemplates is empty.
	DisableDefaultAffinity bool
	// AutomountServiceAccountToken controls SA token mounting.
	AutomountServiceAccountToken *bool
	// WorkflowTemplateRef references a WorkflowTemplate instead of inline templates.
	WorkflowTemplateRef *model.WorkflowTemplateRef
	// ArtifactGC defines artifact garbage collection strategy.
	ArtifactGC *model.ArtifactGCStrategy
	// ArtifactRepositoryRef references an artifact repository config.
	ArtifactRepositoryRef *model.ArtifactRepositoryRef
	// TemplateDefaults defines default values for all templates.
	TemplateDefaults *model.TemplateDefaults
	// Config is the optional GlobalConfig instance used during Build().
	// When nil, Build() falls back to config.GetGlobal() (package singleton).
	// Set via WithConfig() to enable hook isolation per T3.1 / ADR-001.
	Config *config.GlobalConfig
}

// WithConfig attaches a GlobalConfig instance to this Workflow. Build()
// dispatches hooks and applies scalar defaults through cfg rather than the
// package singleton. Returns w for chaining. (T3.1 / ADR-001).
func (w *Workflow) WithConfig(cfg *config.GlobalConfig) *Workflow {
	w.Config = cfg
	return w
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
	// T3.6: Entrypoint is required whenever Templates are defined.
	// A workflow with templates but no entrypoint produces invalid YAML
	// (Argo runtime rejects). WorkflowTemplateRef workflows are exempt
	// because the referenced WorkflowTemplate provides the entrypoint.
	if len(w.Templates) > 0 && w.Entrypoint == "" && w.WorkflowTemplateRef == nil {
		return fmt.Errorf("workflow %q: %w", w.effectiveName(), model.ErrEntrypointMissing)
	}
	return nil
}

// effectiveName returns Name or GenerateName for error messages.
func (w *Workflow) effectiveName() string {
	if w.Name != "" {
		return w.Name
	}
	return w.GenerateName
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
	// T4.7: the post-loop `len(vols)==0` check is unreachable — we just
	// returned early when w.Volumes was empty and otherwise appended one
	// entry per input. Removed (code-p4-dead-len-check).
	return vols, nil
}

func (w *Workflow) buildVolumeClaimTemplates() ([]model.PVCModel, error) {
	if len(w.VolumeClaimTemplates) == 0 {
		return nil, nil
	}
	pvcs := make([]model.PVCModel, 0, len(w.VolumeClaimTemplates))
	for i := range w.VolumeClaimTemplates {
		m, err := w.VolumeClaimTemplates[i].BuildPVC()
		if err != nil {
			return nil, fmt.Errorf("volume claim template: %w", err)
		}
		pvcs = append(pvcs, m)
	}
	// T4.7: same — post-loop len(pvcs)==0 was unreachable.
	return pvcs, nil
}

// resolveAffinity returns the affinity that should land on
// WorkflowSpec.Affinity. Precedence:
//
//  1. User-supplied Affinity wins (no injection, full control).
//  2. Explicit opt-out via DisableDefaultAffinity returns nil.
//  3. No VolumeClaimTemplates means no shared RWO PVC to protect —
//     default affinity adds nothing, so it is skipped (workflows that
//     legitimately parallelize across nodes are unaffected).
//  4. Otherwise inject the default podAffinity from
//     DefaultPodAffinityFor.
func resolveAffinity(w *Workflow) *model.Affinity {
	if w.Affinity != nil {
		return w.Affinity
	}
	if w.DisableDefaultAffinity {
		return nil
	}
	if len(w.VolumeClaimTemplates) == 0 {
		return nil
	}
	return DefaultPodAffinityFor(w)
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
//
// Workflow builder to the serializable WorkflowModel. The function is a
// straight-line pipeline (no branching beyond err checks); splitting it
// would just rename variables. The BaseTemplate extraction (T4.1, v0.6.0)
// will collapse several blocks once it lands.
//
//nolint:funlen // Build is the canonical translation from the user-facing
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

	cfg := resolveConfig(w.Config)
	templates, err := buildTemplateModels(w.Templates, cfg)
	if err != nil {
		return model.WorkflowModel{}, err
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

	wfModel := model.WorkflowModel{
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
			Entrypoint:                   w.Entrypoint,
			Templates:                    templates,
			Arguments:                    args,
			Volumes:                      vols,
			VolumeClaimTemplates:         pvcs,
			ServiceAccountName:           w.ServiceAccountName,
			Parallelism:                  w.Parallelism,
			ActiveDeadlineSeconds:        w.ActiveDeadlineSeconds,
			NodeSelector:                 w.NodeSelector,
			Tolerations:                  w.Tolerations,
			Suspend:                      w.Suspend,
			HostNetwork:                  w.HostNetwork,
			TTLStrategy:                  w.TTLStrategy,
			PodGC:                        w.PodGC,
			Priority:                     w.Priority,
			OnExit:                       w.OnExit,
			Metrics:                      w.buildMetrics(),
			ArchiveLogs:                  w.ArchiveLogs,
			RetryStrategy:                rs,
			ImagePullSecrets:             w.buildImagePullSecrets(),
			PodSpecPatch:                 w.PodSpecPatch,
			Synchronization:              w.Synchronization,
			Hooks:                        w.Hooks,
			DNSConfig:                    w.DNSConfig,
			DNSPolicy:                    w.DNSPolicy,
			PodDisruptionBudget:          w.PodDisruptionBudget,
			PodMetadata:                  w.PodMetadata,
			SecurityContext:              w.SecurityContext,
			Affinity:                     resolveAffinity(w),
			AutomountServiceAccountToken: w.AutomountServiceAccountToken,
			WorkflowTemplateRef:          w.WorkflowTemplateRef,
			ArtifactGC:                   w.ArtifactGC,
			ArtifactRepositoryRef:        w.ArtifactRepositoryRef,
			TemplateDefaults:             w.TemplateDefaults,
		},
	}

	cfg.DispatchWorkflowHooks(&wfModel)
	// T8.2: named workflow hooks fire after anonymous; errors abort.
	if err := cfg.DispatchNamedWorkflowHooks(&wfModel); err != nil {
		return model.WorkflowModel{}, fmt.Errorf("workflow %q: %w", w.effectiveName(), err)
	}
	return wfModel, nil
}

// ToDict converts the workflow to a map (via JSON round-trip).
func (w *Workflow) ToDict() (map[string]interface{}, error) {
	m, err := w.Build()
	if err != nil {
		return nil, err
	}
	return serialize.WorkflowToDict(m)
}

// ToJSON converts the workflow to a JSON string.
func (w *Workflow) ToJSON() (string, error) {
	m, err := w.Build()
	if err != nil {
		return "", err
	}
	return serialize.WorkflowToJSON(m)
}

// ToYAML converts the workflow to a YAML string.
func (w *Workflow) ToYAML() (string, error) {
	m, err := w.Build()
	if err != nil {
		return "", err
	}
	return serialize.WorkflowToYAML(m)
}

// FromYAML creates a WorkflowModel from a YAML string.
func FromYAML(yamlStr string) (model.WorkflowModel, error) {
	return serialize.WorkflowFromYAML(yamlStr)
}

// FromJSON creates a WorkflowModel from a JSON string.
func FromJSON(jsonStr string) (model.WorkflowModel, error) {
	return serialize.WorkflowFromJSON(jsonStr)
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
