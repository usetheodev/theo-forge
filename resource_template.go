package forge

import "fmt"

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
}

func (r *ResourceTemplate) GetName() string {
	return r.Name
}

// BuildTemplate builds the Argo Template for this resource template.
func (r *ResourceTemplate) BuildTemplate() (TemplateModel, error) {
	if r.Name == "" {
		return TemplateModel{}, fmt.Errorf("resource template name cannot be empty")
	}
	if r.Action == "" {
		return TemplateModel{}, fmt.Errorf("resource template action cannot be empty")
	}
	if r.Manifest == "" {
		return TemplateModel{}, fmt.Errorf("resource template manifest cannot be empty")
	}

	var inputs *InputsModel
	if len(r.Inputs) > 0 {
		inputs = &InputsModel{}
		for _, p := range r.Inputs {
			m, err := p.AsInput()
			if err != nil {
				continue
			}
			inputs.Parameters = append(inputs.Parameters, m)
		}
	}

	var outputs *OutputsModel
	if len(r.Outputs) > 0 {
		outputs = &OutputsModel{}
		for _, p := range r.Outputs {
			m, err := p.AsOutput()
			if err != nil {
				continue
			}
			outputs.Parameters = append(outputs.Parameters, m)
		}
	}

	return TemplateModel{
		Name: r.Name,
		Resource: &ResourceTplModel{
			Action:           r.Action,
			Manifest:         r.Manifest,
			SuccessCondition: r.SuccessCondition,
			FailureCondition: r.FailureCondition,
		},
		Inputs:  inputs,
		Outputs: outputs,
	}, nil
}

// Suspend represents a suspend template that pauses execution.
type Suspend struct {
	// Name is the template name.
	Name string
	// Duration is how long to suspend (e.g., "30s", "5m").
	Duration string
}

func (s *Suspend) GetName() string {
	return s.Name
}

// BuildTemplate builds the Argo Template for this suspend template.
func (s *Suspend) BuildTemplate() (TemplateModel, error) {
	if s.Name == "" {
		return TemplateModel{}, fmt.Errorf("suspend template name cannot be empty")
	}
	return TemplateModel{
		Name:    s.Name,
		Suspend: &SuspendModel{Duration: s.Duration},
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

// HTTPModel is the serializable Argo HTTP template.
type HTTPModel struct {
	URL              string            `json:"url" yaml:"url"`
	Method           string            `json:"method,omitempty" yaml:"method,omitempty"`
	Headers          []HTTPHeader      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body             string            `json:"body,omitempty" yaml:"body,omitempty"`
	SuccessCondition string            `json:"successCondition,omitempty" yaml:"successCondition,omitempty"`
	TimeoutSeconds   *int              `json:"timeoutSeconds,omitempty" yaml:"timeoutSeconds,omitempty"`
}

// HTTPHeader is a single HTTP header.
type HTTPHeader struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

func (h *HTTPTemplate) GetName() string {
	return h.Name
}

// BuildTemplate builds the Argo Template for this HTTP template.
func (h *HTTPTemplate) BuildTemplate() (TemplateModel, error) {
	if h.Name == "" {
		return TemplateModel{}, fmt.Errorf("HTTP template name cannot be empty")
	}
	if h.URL == "" {
		return TemplateModel{}, fmt.Errorf("HTTP template URL cannot be empty")
	}

	headers := make([]HTTPHeader, 0, len(h.Headers))
	for k, v := range h.Headers {
		headers = append(headers, HTTPHeader{Name: k, Value: v})
	}

	var inputs *InputsModel
	if len(h.Inputs) > 0 {
		inputs = &InputsModel{}
		for _, p := range h.Inputs {
			m, err := p.AsInput()
			if err != nil {
				continue
			}
			inputs.Parameters = append(inputs.Parameters, m)
		}
	}

	var outputs *OutputsModel
	if len(h.Outputs) > 0 {
		outputs = &OutputsModel{}
		for _, p := range h.Outputs {
			m, err := p.AsOutput()
			if err != nil {
				continue
			}
			outputs.Parameters = append(outputs.Parameters, m)
		}
	}

	return TemplateModel{
		Name:    h.Name,
		HTTP: &HTTPModel{
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
