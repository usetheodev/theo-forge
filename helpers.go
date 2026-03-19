package forge

import (
	"fmt"

	"github.com/usetheo/theo/forge/config"
	"github.com/usetheo/theo/forge/model"
	"github.com/usetheo/theo/forge/serialize"
	"github.com/usetheo/theo/forge/validate"
)

// --- Build helpers ---

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

// buildTemplateModels builds and hooks all Templatable entries, returning the slice of TemplateModel.
func buildTemplateModels(templates []Templatable) ([]model.TemplateModel, error) {
	result := make([]model.TemplateModel, 0, len(templates))
	for _, t := range templates {
		m, err := t.BuildTemplate()
		if err != nil {
			return nil, fmt.Errorf("template %q: %w", t.GetName(), err)
		}
		globalConfig.DispatchTemplateHooks(&m)
		result = append(result, m)
	}
	return result, nil
}

// BuildArguments is a helper to build ArgumentsModel from a mix of parameters and artifacts.
func BuildArguments(params []Parameter, artifacts []ArtifactBuilder) (*model.ArgumentsModel, error) {
	if len(params) == 0 && len(artifacts) == 0 {
		return nil, nil
	}
	args := &model.ArgumentsModel{}
	for _, p := range params {
		m, err := p.AsArgument()
		if err != nil {
			return nil, err
		}
		args.Parameters = append(args.Parameters, m)
	}
	for _, a := range artifacts {
		m, err := a.Build()
		if err != nil {
			return nil, err
		}
		args.Artifacts = append(args.Artifacts, m)
	}
	return args, nil
}

// BuildArgumentsFromMap builds arguments from a map of name->value pairs.
// Values are converted to string Parameter arguments.
func BuildArgumentsFromMap(params map[string]string) *model.ArgumentsModel {
	if len(params) == 0 {
		return nil
	}
	args := &model.ArgumentsModel{}
	for k, v := range params {
		val := v
		args.Parameters = append(args.Parameters, model.ParameterModel{
			Name:  k,
			Value: &val,
		})
	}
	return args
}

// --- File I/O ---

// ToFile writes the workflow YAML to a file.
// If name is empty, the workflow name is used as the filename.
func (w *Workflow) ToFile(outputDir string, name string) (string, error) {
	yamlStr, err := w.ToYAML()
	if err != nil {
		return "", err
	}
	return serialize.WorkflowToFile(yamlStr, outputDir, name, w.Name, w.GenerateName)
}

// FromFile reads a WorkflowModel from a YAML file.
func FromFile(path string) (model.WorkflowModel, error) {
	return serialize.WorkflowFromFile(path)
}

// --- Unit validation ---

// ValidateBinaryUnit validates a binary resource unit (memory: Ki, Mi, Gi, Ti, Pi, Ei).
func ValidateBinaryUnit(s string) error {
	return validate.BinaryUnit(s)
}

// ValidateDecimalUnit validates a decimal resource unit (CPU: m, k, M, G, T, P, E).
func ValidateDecimalUnit(s string) error {
	return validate.DecimalUnit(s)
}

// ConvertBinaryUnit converts a binary unit string to its numeric value in base units (bytes).
func ConvertBinaryUnit(s string) (float64, error) {
	return validate.ConvertBinaryUnit(s)
}

// ConvertDecimalUnit converts a decimal unit string to its numeric value in base units.
func ConvertDecimalUnit(s string) (float64, error) {
	return validate.ConvertDecimalUnit(s)
}

// ValidateResourceRequirements checks that requests don't exceed limits and values are positive.
func ValidateResourceRequirements(r ResourceRequirements) error {
	return validate.ResourceRequirements(r)
}

// --- Global configuration ---

// PreBuildHook is a function that transforms a TemplateModel before submission.
type PreBuildHook = config.PreBuildHook

// WorkflowPreBuildHook is a function that transforms a WorkflowModel before submission.
type WorkflowPreBuildHook = config.WorkflowPreBuildHook

// GlobalConfig holds default values applied to all workflows and templates.
type GlobalConfig = config.GlobalConfig

// globalConfig is the package-level reference to the global singleton.
var globalConfig = config.GetGlobal()

// NewConfig creates an independent GlobalConfig instance for dependency injection.
// Use this instead of GetGlobalConfig when you need isolated configuration
// (e.g., in tests or when building workflows with different settings concurrently).
func NewConfig() *config.GlobalConfig {
	return config.New()
}

// GetGlobalConfig returns the global configuration singleton.
// For isolated configuration (tests, concurrent builds), use NewConfig() instead.
func GetGlobalConfig() *config.GlobalConfig {
	return config.GetGlobal()
}
