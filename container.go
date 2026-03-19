package forge

import (
	"fmt"

	"github.com/usetheo/theo/forge/model"
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
	// EnvFrom sources for env vars.
	EnvFrom []model.EnvFromSource
	// ReadinessProbe for the container.
	ReadinessProbe *model.Probe
	// LivenessProbe for the container.
	LivenessProbe *model.Probe
}

func (c *Container) GetName() string {
	return c.Name
}

// BuildTemplate builds the Argo Template for this container.
func (c *Container) BuildTemplate() (model.TemplateModel, error) {
	if c.Name == "" {
		return model.TemplateModel{}, fmt.Errorf("container template name cannot be empty")
	}

	inputs, err := buildInputsFromParams(c.Inputs, c.InputArtifacts)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("container %q: %w", c.Name, err)
	}

	outputs, err := buildOutputsFromParams(c.Outputs, c.OutputArtifacts)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("container %q: %w", c.Name, err)
	}

	var initContainers []model.ContainerModel
	for _, ic := range c.InitContainers {
		initContainers = append(initContainers, ic.Build())
	}
	var sidecars []model.ContainerModel
	for _, sc := range c.Sidecars {
		sidecars = append(sidecars, sc.Build())
	}

	return model.TemplateModel{
		Name: c.Name,
		Container: &model.ContainerModel{
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
		},
		Inputs:                inputs,
		Outputs:               outputs,
		Metadata:              buildMetadataModel(c.Labels, c.Annotations),
		Timeout:               c.Timeout,
		ActiveDeadlineSeconds: c.ActiveDeadlineSeconds,
		RetryStrategy:         buildRetryStrategyModel(c.RetryStrategy),
		NodeSelector:          c.NodeSelector,
		ServiceAccountName:    c.ServiceAccountName,
		Metrics:               buildMetricsModel(c.Metrics),
		Daemon:                c.Daemon,
		Memoize:               c.Memoize,
		Synchronization:       c.Synchronization,
		PodSpecPatch:          c.PodSpecPatch,
		Hooks:                 c.Hooks,
		ArchiveLocation:       c.ArchiveLocation,
		InitContainers:        initContainers,
		Sidecars:              sidecars,
		Tolerations:           c.Tolerations,
		Parallelism:           c.Parallelism,
	}, nil
}
