package forge

import "fmt"

// ValueFrom describes a location in which to obtain the value to a parameter.
type ValueFrom struct {
	// Path is a file path to read the value from.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Expression is an expression to evaluate.
	Expression string `json:"expression,omitempty" yaml:"expression,omitempty"`
	// JSONPath is a JSONPath expression to evaluate against the resource.
	JSONPath string `json:"jsonPath,omitempty" yaml:"jsonPath,omitempty"`
	// Parameter is a reference to another parameter.
	Parameter string `json:"parameter,omitempty" yaml:"parameter,omitempty"`
	// Default is the default value if the source cannot be resolved.
	Default *string `json:"default,omitempty" yaml:"default,omitempty"`
}

// Parameter represents a workflow parameter with name, value, default, and enum.
type Parameter struct {
	// Name is the parameter name.
	Name string
	// Description documents the parameter.
	Description string
	// Enum restricts the allowed values.
	Enum []string
	// GlobalName makes this parameter available globally.
	GlobalName string
	// Value is the literal value.
	Value *string
	// Default is the default value when not provided.
	Default *string
	// ValueFrom specifies where to obtain the value.
	ValueFrom *ValueFrom
}

// ParameterModel is the serializable representation of a Parameter,
// matching the Argo Workflows API schema.
type ParameterModel struct {
	Name        string     `json:"name" yaml:"name"`
	Value       *string    `json:"value,omitempty" yaml:"value,omitempty"`
	Default     *string    `json:"default,omitempty" yaml:"default,omitempty"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Enum        []string   `json:"enum,omitempty" yaml:"enum,omitempty"`
	GlobalName  string     `json:"globalName,omitempty" yaml:"globalName,omitempty"`
	ValueFrom   *ValueFrom `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

// String returns the string representation of the parameter value.
func (p Parameter) String() (string, error) {
	if p.Value == nil {
		return "", fmt.Errorf("cannot represent Parameter as string: value is not set")
	}
	return *p.Value, nil
}

// WithName returns a copy of the parameter with the given name.
func (p Parameter) WithName(name string) Parameter {
	cp := p
	cp.Name = name
	return cp
}

func (p Parameter) validateName() error {
	if p.Name == "" {
		return fmt.Errorf("name cannot be empty when used")
	}
	return nil
}

// AsInput formats the parameter as an input parameter for a template.
func (p Parameter) AsInput() (ParameterModel, error) {
	if err := p.validateName(); err != nil {
		return ParameterModel{}, err
	}
	return ParameterModel{
		Name:        p.Name,
		Value:       p.Value,
		Default:     p.Default,
		Description: p.Description,
		Enum:        p.Enum,
		GlobalName:  p.GlobalName,
		ValueFrom:   p.ValueFrom,
	}, nil
}

// AsArgument formats the parameter as an argument (excludes default).
func (p Parameter) AsArgument() (ParameterModel, error) {
	if err := p.validateName(); err != nil {
		return ParameterModel{}, err
	}
	return ParameterModel{
		Name:      p.Name,
		Value:     p.Value,
		ValueFrom: p.ValueFrom,
	}, nil
}

// AsOutput formats the parameter as an output parameter.
func (p Parameter) AsOutput() (ParameterModel, error) {
	if err := p.validateName(); err != nil {
		return ParameterModel{}, err
	}
	return ParameterModel{
		Name:       p.Name,
		Value:      p.Value,
		ValueFrom:  p.ValueFrom,
		GlobalName: p.GlobalName,
	}, nil
}
