package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

// Container represents an Argo Workflows container template.
type Container struct {
	// Name is the template name.
	Name string
	// Image is the Docker image.
	Image string
	// Command is the entrypoint.
	Command []string
	// Args are the command arguments.
	Args []string
	// WorkingDir is the working directory.
	WorkingDir string
	// ImagePullPolicy defines when to pull the image.
	ImagePullPolicy ImagePullPolicy
	// Env is the list of environment variables.
	Env []EnvBuilder
	// Resources defines CPU/memory requests and limits.
	Resources *ResourceRequirements
	// VolumeMounts are the volume mounts for the container.
	VolumeMounts []VolumeBuilder
	// Inputs are the template inputs.
	Inputs []Parameter
	// Outputs are the template outputs.
	Outputs []Parameter
	// InputArtifacts are the input artifacts.
	InputArtifacts []ArtifactBuilder
	// OutputArtifacts are the output artifacts.
	OutputArtifacts []ArtifactBuilder
	// Timeout is the template timeout.
	Timeout string
	// ActiveDeadlineSeconds kills the template after X seconds.
	ActiveDeadlineSeconds *int
	// RetryStrategy configures retry behavior.
	RetryStrategy *RetryStrategy
	// NodeSelector constrains pod scheduling.
	NodeSelector map[string]string
	// ServiceAccountName for the pod.
	ServiceAccountName string
	// Metadata for the template.
	Labels      map[string]string
	Annotations map[string]string
	// Metrics for the template.
	Metrics []Metric
	// Ports exposed by the container.
	Ports []ContainerPort
	// Daemon marks this container as a daemon.
	Daemon *bool
	// Memoize caches template outputs.
	Memoize *model.MemoizeModel
	// Synchronization configures synchronization constraints.
	Synchronization *model.SynchronizationModel
	// PodSpecPatch is a JSON/YAML patch for the pod spec.
	PodSpecPatch string
	// Hooks are lifecycle hooks.
	Hooks map[string]model.LifecycleHook
	// ArchiveLocation overrides the default artifact location.
	ArchiveLocation *model.ArtifactLocation
	// InitContainers are init containers for the pod.
	InitContainers []UserContainer
	// Sidecars are sidecar containers.
	Sidecars []UserContainer
	// Tolerations for pod scheduling.
	Tolerations []model.Toleration
	// Parallelism limits concurrent pods.
	Parallelism *int
	// SecurityContext for the container.
	SecurityContext *model.SecurityContext
	// Affinity defines scheduling constraints for the template pod.
	Affinity *model.Affinity
	// EnvFrom sources for env vars.
	EnvFrom []model.EnvFromSource
	// ReadinessProbe for the container.
	ReadinessProbe *model.Probe
	// LivenessProbe for the container.
	LivenessProbe *model.Probe
}

// GetName is the method.
func (c *Container) GetName() string {
	return c.Name
}

// BuildTemplate builds the Argo Template for this container.
// Note: Image presence is validated AFTER ApplyTemplateDefaults runs (so
// GlobalConfig.Image can supply a default); see validateBuiltTemplate.
func (c *Container) BuildTemplate() (model.TemplateModel, error) {
	if c.Name == "" {
		return model.TemplateModel{}, fmt.Errorf("container template name cannot be empty")
	}

	var out model.TemplateModel
	sidecars, err := assignSharedTemplateFields(&out, c.Name, sharedTemplateFields{
		Inputs:                c.Inputs,
		Outputs:               c.Outputs,
		InputArtifacts:        c.InputArtifacts,
		OutputArtifacts:       c.OutputArtifacts,
		Labels:                c.Labels,
		Annotations:           c.Annotations,
		Timeout:               c.Timeout,
		ActiveDeadlineSeconds: c.ActiveDeadlineSeconds,
		RetryStrategy:         c.RetryStrategy,
		NodeSelector:          c.NodeSelector,
		ServiceAccountName:    c.ServiceAccountName,
		Metrics:               c.Metrics,
		Daemon:                c.Daemon,
		Memoize:               c.Memoize,
		Synchronization:       c.Synchronization,
		PodSpecPatch:          c.PodSpecPatch,
		Hooks:                 c.Hooks,
		Sidecars:              c.Sidecars,
		Tolerations:           c.Tolerations,
		Affinity:              c.Affinity,
	})
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("container %q: %w", c.Name, err)
	}

	var initContainers []model.ContainerModel
	for i := range c.InitContainers {
		initContainers = append(initContainers, c.InitContainers[i].Build())
	}

	out.Container = &model.ContainerModel{
		Image:           c.Image,
		Command:         c.Command,
		Args:            c.Args,
		WorkingDir:      c.WorkingDir,
		Env:             buildEnvVars(c.Env),
		EnvFrom:         c.EnvFrom,
		Resources:       c.Resources,
		VolumeMounts:    buildVolumeMountModels(c.VolumeMounts),
		ImagePullPolicy: string(c.ImagePullPolicy),
		Ports:           c.Ports,
		SecurityContext: c.SecurityContext,
		ReadinessProbe:  c.ReadinessProbe,
		LivenessProbe:   c.LivenessProbe,
	}
	out.ArchiveLocation = c.ArchiveLocation
	out.InitContainers = initContainers
	out.Sidecars = sidecars
	out.Parallelism = c.Parallelism
	return out, nil
}

// --- Script ---

// Script represents an Argo Workflows script template.
type Script struct {
	// Name is the template name.
	Name string
	// Image is the Docker image.
	Image string
	// Command is the script interpreter (e.g., ["python"], ["bash"]).
	Command []string
	// Args are additional arguments.
	Args []string
	// Source is the script source code.
	Source string
	// WorkingDir is the working directory.
	WorkingDir string
	// ImagePullPolicy defines when to pull the image.
	ImagePullPolicy ImagePullPolicy
	// Env is the list of environment variables.
	Env []EnvBuilder
	// Resources defines CPU/memory requests and limits.
	Resources *ResourceRequirements
	// VolumeMounts are the volume mounts.
	VolumeMounts []VolumeBuilder
	// Inputs are the template inputs.
	Inputs []Parameter
	// Outputs are the template outputs.
	Outputs []Parameter
	// InputArtifacts are input artifacts.
	InputArtifacts []ArtifactBuilder
	// OutputArtifacts are output artifacts.
	OutputArtifacts []ArtifactBuilder
	// Timeout is the template timeout.
	Timeout string
	// ActiveDeadlineSeconds kills the template after X seconds.
	ActiveDeadlineSeconds *int
	// RetryStrategy configures retry behavior.
	RetryStrategy *RetryStrategy
	// NodeSelector constrains pod scheduling.
	NodeSelector map[string]string
	// ServiceAccountName for the pod.
	ServiceAccountName string
	// Labels for the template.
	Labels map[string]string
	// Annotations for the template.
	Annotations map[string]string
	// Metrics for the template.
	Metrics []Metric
	// Daemon marks this script as a daemon.
	Daemon *bool
	// Memoize caches template outputs.
	Memoize *model.MemoizeModel
	// Synchronization constraints.
	Synchronization *model.SynchronizationModel
	// PodSpecPatch is a JSON/YAML patch for the pod spec.
	PodSpecPatch string
	// Hooks are lifecycle hooks.
	Hooks map[string]model.LifecycleHook
	// Sidecars are sidecar containers.
	Sidecars []UserContainer
	// Tolerations for pod scheduling.
	Tolerations []model.Toleration
	// Affinity defines scheduling constraints for the template pod.
	Affinity *model.Affinity
	// SecurityContext for the container (T4.1: Script now mirrors Container).
	SecurityContext *model.SecurityContext
	// InitContainers are init containers for the pod (T4.1).
	InitContainers []UserContainer
	// EnvFrom sources for env vars (T4.1).
	EnvFrom []model.EnvFromSource
	// ReadinessProbe for the container (T4.1).
	ReadinessProbe *model.Probe
	// LivenessProbe for the container (T4.1).
	LivenessProbe *model.Probe
	// Ports exposed by the container (T4.1).
	Ports []ContainerPort
	// Parallelism limits concurrent pods (T4.1).
	Parallelism *int
	// Lifecycle defines container lifecycle hooks (T4.1).
	// Note: held as a model-level concept; for now this is a placeholder
	// mirroring Container's Hooks map. Kept distinct because container.Lifecycle
	// is a richer model in upstream Argo; see model.LifecycleHook documentation.
	Lifecycle map[string]model.LifecycleHook
}

// GetName is the method.
func (s *Script) GetName() string {
	return s.Name
}

// BuildTemplate builds the Argo Template for this script.
// Note: Image presence is validated AFTER ApplyTemplateDefaults runs (so
// GlobalConfig.Image can supply a default); see validateBuiltTemplate.
func (s *Script) BuildTemplate() (model.TemplateModel, error) {
	if s.Name == "" {
		return model.TemplateModel{}, fmt.Errorf("script template name cannot be empty")
	}
	if s.Source == "" {
		return model.TemplateModel{}, fmt.Errorf("script source cannot be empty")
	}

	var out model.TemplateModel
	sidecars, err := assignSharedTemplateFields(&out, s.Name, sharedTemplateFields{
		Inputs:                s.Inputs,
		Outputs:               s.Outputs,
		InputArtifacts:        s.InputArtifacts,
		OutputArtifacts:       s.OutputArtifacts,
		Labels:                s.Labels,
		Annotations:           s.Annotations,
		Timeout:               s.Timeout,
		ActiveDeadlineSeconds: s.ActiveDeadlineSeconds,
		RetryStrategy:         s.RetryStrategy,
		NodeSelector:          s.NodeSelector,
		ServiceAccountName:    s.ServiceAccountName,
		Metrics:               s.Metrics,
		Daemon:                s.Daemon,
		Memoize:               s.Memoize,
		Synchronization:       s.Synchronization,
		PodSpecPatch:          s.PodSpecPatch,
		Hooks:                 s.Hooks,
		Sidecars:              s.Sidecars,
		Tolerations:           s.Tolerations,
		Affinity:              s.Affinity,
	})
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("script %q: %w", s.Name, err)
	}

	// T4.1: Script-only initContainers (paralelo a Container) — Script gains
	// the field as part of this refactor.
	var initContainers []model.ContainerModel
	for i := range s.InitContainers {
		initContainers = append(initContainers, s.InitContainers[i].Build())
	}

	out.Script = &model.ScriptModel{
		Image:           s.Image,
		Command:         s.Command,
		Args:            s.Args,
		Source:          s.Source,
		WorkingDir:      s.WorkingDir,
		Env:             buildEnvVars(s.Env),
		EnvFrom:         s.EnvFrom,
		Resources:       s.Resources,
		VolumeMounts:    buildVolumeMountModels(s.VolumeMounts),
		ImagePullPolicy: string(s.ImagePullPolicy),
		Ports:           s.Ports,
		SecurityContext: s.SecurityContext,
		ReadinessProbe:  s.ReadinessProbe,
		LivenessProbe:   s.LivenessProbe,
	}
	out.InitContainers = initContainers
	out.Sidecars = sidecars
	out.Parallelism = s.Parallelism
	return out, nil
}
