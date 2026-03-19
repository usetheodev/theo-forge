package forge

import (
	"fmt"

	"github.com/usetheo/theo/forge/model"
)

// buildInputsFromParams builds an InputsModel from parameters and artifacts.
func buildInputsFromParams(params []Parameter, artifacts []ArtifactBuilder) (*model.InputsModel, error) {
	var ps []model.ParameterModel
	for _, p := range params {
		m, err := p.AsInput()
		if err != nil {
			return nil, fmt.Errorf("input parameter %q: %w", p.Name, err)
		}
		ps = append(ps, m)
	}
	var arts []model.ArtifactModel
	for _, a := range artifacts {
		m, err := a.Build()
		if err != nil {
			return nil, fmt.Errorf("input artifact: %w", err)
		}
		arts = append(arts, m)
	}
	if len(ps) == 0 && len(arts) == 0 {
		return nil, nil
	}
	return &model.InputsModel{Parameters: ps, Artifacts: arts}, nil
}

// buildOutputsFromParams builds an OutputsModel from parameters and artifacts.
func buildOutputsFromParams(params []Parameter, artifacts []ArtifactBuilder) (*model.OutputsModel, error) {
	var ps []model.ParameterModel
	for _, p := range params {
		m, err := p.AsOutput()
		if err != nil {
			return nil, fmt.Errorf("output parameter %q: %w", p.Name, err)
		}
		ps = append(ps, m)
	}
	var arts []model.ArtifactModel
	for _, a := range artifacts {
		m, err := a.Build()
		if err != nil {
			return nil, fmt.Errorf("output artifact: %w", err)
		}
		arts = append(arts, m)
	}
	if len(ps) == 0 && len(arts) == 0 {
		return nil, nil
	}
	return &model.OutputsModel{Parameters: ps, Artifacts: arts}, nil
}

// buildEnvVars converts a slice of EnvBuilder to serializable models.
func buildEnvVars(envs []EnvBuilder) []model.EnvVarModel {
	if len(envs) == 0 {
		return nil
	}
	result := make([]model.EnvVarModel, len(envs))
	for i, e := range envs {
		result[i] = e.Build()
	}
	return result
}

// buildVolumeMountModels converts a slice of VolumeBuilder to serializable mount models.
func buildVolumeMountModels(volumes []VolumeBuilder) []model.VolumeMountModel {
	if len(volumes) == 0 {
		return nil
	}
	result := make([]model.VolumeMountModel, len(volumes))
	for i, v := range volumes {
		result[i] = v.BuildVolumeMount()
	}
	return result
}

// buildMetadataModel creates a MetadataModel from labels and annotations.
func buildMetadataModel(labels, annotations map[string]string) *model.MetadataModel {
	if len(labels) == 0 && len(annotations) == 0 {
		return nil
	}
	return &model.MetadataModel{Labels: labels, Annotations: annotations}
}

// buildMetricsModel creates a MetricsModel from a slice of metrics.
func buildMetricsModel(metrics []model.Metric) *model.MetricsModel {
	if len(metrics) == 0 {
		return nil
	}
	return &model.MetricsModel{Prometheus: metrics}
}

// buildRetryStrategyModel converts a RetryStrategy to its model, or nil.
func buildRetryStrategyModel(rs *RetryStrategy) *model.RetryStrategyModel {
	if rs == nil {
		return nil
	}
	m := rs.Build()
	return &m
}
