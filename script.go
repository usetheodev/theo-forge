package forge

import "fmt"

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
}

func (s *Script) GetName() string {
	return s.Name
}

func (s *Script) buildEnv() []EnvVarModel {
	if len(s.Env) == 0 {
		return nil
	}
	envs := make([]EnvVarModel, len(s.Env))
	for i, e := range s.Env {
		envs[i] = e.Build()
	}
	return envs
}

func (s *Script) buildVolumeMounts() []VolumeMountModel {
	if len(s.VolumeMounts) == 0 {
		return nil
	}
	mounts := make([]VolumeMountModel, len(s.VolumeMounts))
	for i, v := range s.VolumeMounts {
		mounts[i] = v.BuildVolumeMount()
	}
	return mounts
}

func (s *Script) buildInputs() (*InputsModel, error) {
	var params []ParameterModel
	for _, p := range s.Inputs {
		m, err := p.AsInput()
		if err != nil {
			return nil, fmt.Errorf("input parameter %q: %w", p.Name, err)
		}
		params = append(params, m)
	}
	var arts []ArtifactModel
	for _, a := range s.InputArtifacts {
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

func (s *Script) buildOutputs() (*OutputsModel, error) {
	var params []ParameterModel
	for _, p := range s.Outputs {
		m, err := p.AsOutput()
		if err != nil {
			return nil, fmt.Errorf("output parameter %q: %w", p.Name, err)
		}
		params = append(params, m)
	}
	var arts []ArtifactModel
	for _, a := range s.OutputArtifacts {
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

func (s *Script) buildMetadata() *MetadataModel {
	if len(s.Labels) == 0 && len(s.Annotations) == 0 {
		return nil
	}
	return &MetadataModel{Labels: s.Labels, Annotations: s.Annotations}
}

func (s *Script) buildMetrics() *MetricsModel {
	if len(s.Metrics) == 0 {
		return nil
	}
	return &MetricsModel{Prometheus: s.Metrics}
}

// BuildTemplate builds the Argo Template for this script.
func (s *Script) BuildTemplate() (TemplateModel, error) {
	if s.Name == "" {
		return TemplateModel{}, fmt.Errorf("script template name cannot be empty")
	}
	if s.Source == "" {
		return TemplateModel{}, fmt.Errorf("script source cannot be empty")
	}

	var rs *RetryStrategyModel
	if s.RetryStrategy != nil {
		m := s.RetryStrategy.Build()
		rs = &m
	}

	inputs, err := s.buildInputs()
	if err != nil {
		return TemplateModel{}, fmt.Errorf("script %q: %w", s.Name, err)
	}

	outputs, err := s.buildOutputs()
	if err != nil {
		return TemplateModel{}, fmt.Errorf("script %q: %w", s.Name, err)
	}

	return TemplateModel{
		Name: s.Name,
		Script: &ScriptModel{
			Image:           s.Image,
			Command:         s.Command,
			Args:            s.Args,
			Source:          s.Source,
			WorkingDir:      s.WorkingDir,
			Env:             s.buildEnv(),
			Resources:       s.Resources,
			VolumeMounts:    s.buildVolumeMounts(),
			ImagePullPolicy: string(s.ImagePullPolicy),
		},
		Inputs:                inputs,
		Outputs:               outputs,
		Metadata:              s.buildMetadata(),
		Timeout:               s.Timeout,
		ActiveDeadlineSeconds: s.ActiveDeadlineSeconds,
		RetryStrategy:         rs,
		NodeSelector:          s.NodeSelector,
		ServiceAccountName:    s.ServiceAccountName,
		Metrics:               s.buildMetrics(),
	}, nil
}
