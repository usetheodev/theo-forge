package forge

import (
	"fmt"

	"github.com/usetheo/theo/forge/model"
)

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

// BuildTemplate builds the Argo Template for this script.
func (s *Script) BuildTemplate() (model.TemplateModel, error) {
	if s.Name == "" {
		return model.TemplateModel{}, fmt.Errorf("script template name cannot be empty")
	}
	if s.Source == "" {
		return model.TemplateModel{}, fmt.Errorf("script source cannot be empty")
	}

	inputs, err := buildInputsFromParams(s.Inputs, s.InputArtifacts)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("script %q: %w", s.Name, err)
	}

	outputs, err := buildOutputsFromParams(s.Outputs, s.OutputArtifacts)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("script %q: %w", s.Name, err)
	}

	return model.TemplateModel{
		Name: s.Name,
		Script: &model.ScriptModel{
			Image:           s.Image,
			Command:         s.Command,
			Args:            s.Args,
			Source:          s.Source,
			WorkingDir:      s.WorkingDir,
			Env:             buildEnvVars(s.Env),
			Resources:       s.Resources,
			VolumeMounts:    buildVolumeMountModels(s.VolumeMounts),
			ImagePullPolicy: string(s.ImagePullPolicy),
		},
		Inputs:                inputs,
		Outputs:               outputs,
		Metadata:              buildMetadataModel(s.Labels, s.Annotations),
		Timeout:               s.Timeout,
		ActiveDeadlineSeconds: s.ActiveDeadlineSeconds,
		RetryStrategy:         buildRetryStrategyModel(s.RetryStrategy),
		NodeSelector:          s.NodeSelector,
		ServiceAccountName:    s.ServiceAccountName,
		Metrics:               buildMetricsModel(s.Metrics),
	}, nil
}
