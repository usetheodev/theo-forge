package forge

import (
	"fmt"

	"github.com/usetheo/theo/forge/model"
)

// ResourceTemplate creates/applies K8s resources via Argo.
type ResourceTemplate struct {
	// Name is the template name.
	Name string
	// Action is the operation (create, apply, patch, delete).
	Action string
	// Manifest is the K8s resource YAML manifest.
	Manifest string
	// SuccessCondition is a jsonpath condition for success.
	SuccessCondition string
	// FailureCondition is a jsonpath condition for failure.
	FailureCondition string
	// Inputs are the template inputs.
	Inputs []Parameter
	// Outputs are the template outputs.
	Outputs []Parameter
	// InputArtifacts are input artifacts.
	InputArtifacts []ArtifactBuilder
	// OutputArtifacts are output artifacts.
	OutputArtifacts []ArtifactBuilder
	// Flags are extra flags passed to kubectl.
	Flags []string
	// SetOwnerReference adds owner reference to the resource.
	SetOwnerReference *bool
	// MergeStrategy for patch operations.
	MergeStrategy string
	// Labels for the template.
	Labels map[string]string
	// Annotations for the template.
	Annotations map[string]string
}

func (r *ResourceTemplate) GetName() string {
	return r.Name
}

// BuildTemplate builds the Argo Template for this resource template.
func (r *ResourceTemplate) BuildTemplate() (model.TemplateModel, error) {
	if r.Name == "" {
		return model.TemplateModel{}, fmt.Errorf("resource template name cannot be empty")
	}
	if r.Action == "" {
		return model.TemplateModel{}, fmt.Errorf("resource template action cannot be empty")
	}
	if r.Manifest == "" {
		return model.TemplateModel{}, fmt.Errorf("resource template manifest cannot be empty")
	}

	inputs, err := buildInputsFromParams(r.Inputs, r.InputArtifacts)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("resource template %q: %w", r.Name, err)
	}

	outputs, err2 := buildOutputsFromParams(r.Outputs, r.OutputArtifacts)
	if err2 != nil {
		return model.TemplateModel{}, fmt.Errorf("resource template %q: %w", r.Name, err2)
	}

	return model.TemplateModel{
		Name: r.Name,
		Resource: &model.ResourceTplModel{
			Action:            r.Action,
			Manifest:          r.Manifest,
			SuccessCondition:  r.SuccessCondition,
			FailureCondition:  r.FailureCondition,
			Flags:             r.Flags,
			SetOwnerReference: r.SetOwnerReference,
			MergeStrategy:     r.MergeStrategy,
		},
		Inputs:   inputs,
		Outputs:  outputs,
		Metadata: buildMetadataModel(r.Labels, r.Annotations),
	}, nil
}

// Suspend represents a suspend template that pauses execution.
type Suspend struct {
	// Name is the template name.
	Name string
	// Duration is how long to suspend (e.g., "30s", "5m").
	Duration string
	// Inputs are the template inputs.
	Inputs []Parameter
	// Outputs are the template outputs.
	Outputs []Parameter
}

func (s *Suspend) GetName() string {
	return s.Name
}

// BuildTemplate builds the Argo Template for this suspend template.
func (s *Suspend) BuildTemplate() (model.TemplateModel, error) {
	if s.Name == "" {
		return model.TemplateModel{}, fmt.Errorf("suspend template name cannot be empty")
	}

	inputs, err := buildInputsFromParams(s.Inputs, nil)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("suspend %q: %w", s.Name, err)
	}

	outputs, err := buildOutputsFromParams(s.Outputs, nil)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("suspend %q: %w", s.Name, err)
	}

	return model.TemplateModel{
		Name:    s.Name,
		Suspend: &model.SuspendModel{Duration: s.Duration},
		Inputs:  inputs,
		Outputs: outputs,
	}, nil
}

// HTTPTemplate represents an HTTP request template.
type HTTPTemplate struct {
	// Name is the template name.
	Name string
	// URL is the request URL.
	URL string
	// Method is the HTTP method (GET, POST, etc.).
	Method string
	// Headers are the request headers.
	Headers map[string]string
	// Body is the request body.
	Body string
	// SuccessCondition is a condition for success.
	SuccessCondition string
	// Inputs are the template inputs.
	Inputs []Parameter
	// Outputs are the template outputs.
	Outputs []Parameter
	// Timeout is the request timeout.
	Timeout string
}

func (h *HTTPTemplate) GetName() string {
	return h.Name
}

// BuildTemplate builds the Argo Template for this HTTP template.
func (h *HTTPTemplate) BuildTemplate() (model.TemplateModel, error) {
	if h.Name == "" {
		return model.TemplateModel{}, fmt.Errorf("HTTP template name cannot be empty")
	}
	if h.URL == "" {
		return model.TemplateModel{}, fmt.Errorf("HTTP template URL cannot be empty")
	}

	headers := make([]model.HTTPHeader, 0, len(h.Headers))
	for k, v := range h.Headers {
		headers = append(headers, model.HTTPHeader{Name: k, Value: v})
	}

	var inputs *model.InputsModel
	if len(h.Inputs) > 0 {
		inputs = &model.InputsModel{}
		for _, p := range h.Inputs {
			m, err := p.AsInput()
			if err != nil {
				return model.TemplateModel{}, fmt.Errorf("HTTP template %q input parameter %q: %w", h.Name, p.Name, err)
			}
			inputs.Parameters = append(inputs.Parameters, m)
		}
	}

	var outputs *model.OutputsModel
	if len(h.Outputs) > 0 {
		outputs = &model.OutputsModel{}
		for _, p := range h.Outputs {
			m, err := p.AsOutput()
			if err != nil {
				return model.TemplateModel{}, fmt.Errorf("HTTP template %q output parameter %q: %w", h.Name, p.Name, err)
			}
			outputs.Parameters = append(outputs.Parameters, m)
		}
	}

	return model.TemplateModel{
		Name: h.Name,
		HTTP: &model.HTTPModel{
			URL:              h.URL,
			Method:           h.Method,
			Headers:          headers,
			Body:             h.Body,
			SuccessCondition: h.SuccessCondition,
		},
		Inputs:  inputs,
		Outputs: outputs,
		Timeout: h.Timeout,
	}, nil
}
