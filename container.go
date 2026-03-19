package forge

import "fmt"

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

func (c *Container) buildEnv() []EnvVarModel {
	if len(c.Env) == 0 {
		return nil
	}
	envs := make([]EnvVarModel, len(c.Env))
	for i, e := range c.Env {
		envs[i] = e.Build()
	}
	return envs
}

func (c *Container) buildVolumeMounts() []VolumeMountModel {
	if len(c.VolumeMounts) == 0 {
		return nil
	}
	mounts := make([]VolumeMountModel, len(c.VolumeMounts))
	for i, v := range c.VolumeMounts {
		mounts[i] = v.BuildVolumeMount()
	}
	return mounts
}

func (c *Container) buildInputs() (*InputsModel, error) {
	var params []ParameterModel
	for _, p := range c.Inputs {
		m, err := p.AsInput()
		if err != nil {
			return nil, fmt.Errorf("input parameter %q: %w", p.Name, err)
		}
		params = append(params, m)
	}
	var arts []ArtifactModel
	for _, a := range c.InputArtifacts {
		m, err := a.Build()
		if err != nil {
			return nil, fmt.Errorf("input artifact: %w", err)
		}
		arts = append(arts, m)
	}
	if len(params) == 0 && len(arts) == 0 {
		return nil, nil
	}
	return &InputsModel{Parameters: params, Artifacts: arts}, nil
}

func (c *Container) buildOutputs() (*OutputsModel, error) {
	var params []ParameterModel
	for _, p := range c.Outputs {
		m, err := p.AsOutput()
		if err != nil {
			return nil, fmt.Errorf("output parameter %q: %w", p.Name, err)
		}
		params = append(params, m)
	}
	var arts []ArtifactModel
	for _, a := range c.OutputArtifacts {
		m, err := a.Build()
		if err != nil {
			return nil, fmt.Errorf("output artifact: %w", err)
		}
		arts = append(arts, m)
	}
	if len(params) == 0 && len(arts) == 0 {
		return nil, nil
	}
	return &OutputsModel{Parameters: params, Artifacts: arts}, nil
}

func (c *Container) buildMetadata() *MetadataModel {
	if len(c.Labels) == 0 && len(c.Annotations) == 0 {
		return nil
	}
	return &MetadataModel{Labels: c.Labels, Annotations: c.Annotations}
}

func (c *Container) buildMetrics() *MetricsModel {
	if len(c.Metrics) == 0 {
		return nil
	}
	return &MetricsModel{Prometheus: c.Metrics}
}

// BuildTemplate builds the Argo Template for this container.
func (c *Container) BuildTemplate() (TemplateModel, error) {
	if c.Name == "" {
		return TemplateModel{}, fmt.Errorf("container template name cannot be empty")
	}

	var rs *RetryStrategyModel
	if c.RetryStrategy != nil {
		m := c.RetryStrategy.Build()
		rs = &m
	}

	inputs, err := c.buildInputs()
	if err != nil {
		return TemplateModel{}, fmt.Errorf("container %q: %w", c.Name, err)
	}

	outputs, err := c.buildOutputs()
	if err != nil {
		return TemplateModel{}, fmt.Errorf("container %q: %w", c.Name, err)
	}

	return TemplateModel{
		Name: c.Name,
		Container: &ContainerModel{
			Image:           c.Image,
			Command:         c.Command,
			Args:            c.Args,
			WorkingDir:      c.WorkingDir,
			Env:             c.buildEnv(),
			Resources:       c.Resources,
			VolumeMounts:    c.buildVolumeMounts(),
			ImagePullPolicy: string(c.ImagePullPolicy),
			Ports:           c.Ports,
		},
		Inputs:                inputs,
		Outputs:               outputs,
		Metadata:              c.buildMetadata(),
		Timeout:               c.Timeout,
		ActiveDeadlineSeconds: c.ActiveDeadlineSeconds,
		RetryStrategy:         rs,
		NodeSelector:          c.NodeSelector,
		ServiceAccountName:    c.ServiceAccountName,
		Metrics:               c.buildMetrics(),
	}, nil
}
