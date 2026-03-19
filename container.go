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

	return model.TemplateModel{
		Name: c.Name,
		Container: &model.ContainerModel{
			Image:           c.Image,
			Command:         c.Command,
			Args:            c.Args,
			WorkingDir:      c.WorkingDir,
			Env:             buildEnvVars(c.Env),
			Resources:       c.Resources,
			VolumeMounts:    buildVolumeMountModels(c.VolumeMounts),
			ImagePullPolicy: string(c.ImagePullPolicy),
			Ports:           c.Ports,
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
	}, nil
}
