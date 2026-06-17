package forge

import "github.com/usetheodev/theo-forge/model"

// sharedTemplateFields captures the ~17 fields that Container and Script
// historically duplicated. Used internally by BuildTemplate methods to
// translate to TemplateModel via assignSharedTemplateFields.
//
// T4.1 / arch-container-script-field-duplication — we keep the public
// Container/Script flat-field shape (callers use `&Container{Name: ..., Image: ...}`)
// for backward compatibility, but the BUILD path goes through this helper
// to eliminate the duplicated translation logic. A future v1.0 may promote
// this to a real embedded BaseTemplate; documented in docs/CRD_PARITY.md.
type sharedTemplateFields struct {
	Inputs                []Parameter
	Outputs               []Parameter
	InputArtifacts        []ArtifactBuilder
	OutputArtifacts       []ArtifactBuilder
	Labels                map[string]string
	Annotations           map[string]string
	Timeout               string
	ActiveDeadlineSeconds *int
	RetryStrategy         *RetryStrategy
	NodeSelector          map[string]string
	ServiceAccountName    string
	Metrics               []Metric
	Daemon                *bool
	Memoize               *model.MemoizeModel
	Synchronization       *model.SynchronizationModel
	PodSpecPatch          string
	Hooks                 map[string]model.LifecycleHook
	Sidecars              []UserContainer
	Tolerations           []model.Toleration
	Affinity              *model.Affinity
}

// assignSharedTemplateFields fills the shared fields of `out` from `s`.
// `name` is the symbolic name used in error wrapping.
// Returns (sidecars, error) so the caller can attach sidecars to the right
// container set (Container vs Script have distinct sidecar plumbing).
func assignSharedTemplateFields(out *model.TemplateModel, name string, s sharedTemplateFields) ([]model.ContainerModel, error) {
	inputs, err := buildInputsFromParams(s.Inputs, s.InputArtifacts)
	if err != nil {
		return nil, err
	}
	outputs, err := buildOutputsFromParams(s.Outputs, s.OutputArtifacts)
	if err != nil {
		return nil, err
	}

	out.Name = name
	out.Inputs = inputs
	out.Outputs = outputs
	out.Metadata = buildMetadataModel(s.Labels, s.Annotations)
	out.Timeout = s.Timeout
	out.ActiveDeadlineSeconds = s.ActiveDeadlineSeconds
	out.RetryStrategy = buildRetryStrategyModel(s.RetryStrategy)
	out.NodeSelector = s.NodeSelector
	out.ServiceAccountName = s.ServiceAccountName
	out.Metrics = buildMetricsModel(s.Metrics)
	out.Daemon = s.Daemon
	out.Memoize = s.Memoize
	out.Synchronization = s.Synchronization
	out.PodSpecPatch = s.PodSpecPatch
	out.Hooks = s.Hooks
	out.Tolerations = s.Tolerations
	out.Affinity = s.Affinity

	var sidecars []model.ContainerModel
	for i := range s.Sidecars {
		sidecars = append(sidecars, s.Sidecars[i].Build())
	}
	return sidecars, nil
}
