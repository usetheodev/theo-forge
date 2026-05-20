package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/config"
	"github.com/usetheodev/theo-forge/expr"
	"github.com/usetheodev/theo-forge/model"
	"github.com/usetheodev/theo-forge/validate"
)

// T4.2 — split from helpers.go per SRP. Contains:
//   - shared sub-model builders (inputs/outputs/env/volume/metadata/metrics/retry)
//   - the top-level Templatable pipeline (buildTemplateModels)
//   - Build-time invariants (validateBuiltTemplate)
//   - generic public helpers (BuildArguments, BuildArgumentsFromMap, Ptr, InputParam)

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
// cfg is the resolved config (see resolveConfig). Order of operations:
//  1. BuildTemplate() produces the raw model (cheap structural validation only).
//  2. cfg.ApplyTemplateDefaults overlays scalar defaults (Image, ImagePullPolicy, ServiceAccountName).
//  3. validateBuiltTemplate checks required fields that may have been supplied by step 2
//     (T3.5 ErrEmptyImage — Image must exist either explicitly OR via cfg default).
//  4. cfg.DispatchTemplateHooks fires consumer hooks last, after defaults + validation.
func buildTemplateModels(templates []Templatable, cfg *config.GlobalConfig) ([]model.TemplateModel, error) {
	result := make([]model.TemplateModel, 0, len(templates))
	for _, t := range templates {
		m, err := t.BuildTemplate()
		if err != nil {
			return nil, fmt.Errorf("template %q: %w", t.GetName(), err)
		}
		cfg.ApplyTemplateDefaults(&m)
		if err := validateBuiltTemplate(&m); err != nil {
			return nil, fmt.Errorf("template %q: %w", t.GetName(), err)
		}
		cfg.DispatchTemplateHooks(&m)
		// T8.2: named hooks fire after anonymous hooks; their errors abort the build.
		if err := cfg.DispatchNamedTemplateHooks(&m); err != nil {
			return nil, fmt.Errorf("template %q: %w", t.GetName(), err)
		}
		result = append(result, m)
	}
	return result, nil
}

// validateBuiltTemplate enforces field invariants that depend on
// GlobalConfig defaults having been applied first.
//   - Image presence on Container/Script (T3.5 / code-p4-missing-image-validation).
//   - Resource units on Container/Script (T3.12 / completeness-h7-resource-validation-not-in-build).
//
// type. Flattening would lose the early-return on missing Image.
//
//nolint:nestif // 2-deep nesting (type check + resource check) per receiver
func validateBuiltTemplate(t *model.TemplateModel) error {
	if t.Container != nil {
		if t.Container.Image == "" {
			return model.ErrEmptyImage
		}
		if t.Container.Resources != nil {
			if err := validate.ResourceRequirements(*t.Container.Resources); err != nil {
				return fmt.Errorf("container resources: %w", err)
			}
		}
	}
	if t.Script != nil {
		if t.Script.Image == "" {
			return model.ErrEmptyImage
		}
		if t.Script.Resources != nil {
			if err := validate.ResourceRequirements(*t.Script.Resources); err != nil {
				return fmt.Errorf("script resources: %w", err)
			}
		}
	}
	return nil
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

// Ptr returns a pointer to v. Use it when an SDK field expects `*T`
// (e.g., Parameter.Value, *int retry counts) and you have a literal value.
//
// Example:
//
//	forge.Parameter{Name: "msg", Value: forge.Ptr("hello")}
func Ptr[T any](v T) *T {
	return &v
}

// InputParam returns an Argo template reference to an input parameter:
// `"{{inputs.parameters.<name>}}"`. Convenience re-export of [expr.InputParam].
func InputParam(name string) string {
	return expr.InputParam(name)
}
