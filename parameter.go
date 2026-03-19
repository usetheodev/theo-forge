package forge

import (
	"fmt"

	"github.com/usetheo/theo/forge/model"
)

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
	ValueFrom *model.ValueFrom
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
func (p Parameter) AsInput() (model.ParameterModel, error) {
	if err := p.validateName(); err != nil {
		return model.ParameterModel{}, err
	}
	return model.ParameterModel{
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
func (p Parameter) AsArgument() (model.ParameterModel, error) {
	if err := p.validateName(); err != nil {
		return model.ParameterModel{}, err
	}
	return model.ParameterModel{
		Name:      p.Name,
		Value:     p.Value,
		ValueFrom: p.ValueFrom,
	}, nil
}

// AsOutput formats the parameter as an output parameter.
func (p Parameter) AsOutput() (model.ParameterModel, error) {
	if err := p.validateName(); err != nil {
		return model.ParameterModel{}, err
	}
	return model.ParameterModel{
		Name:       p.Name,
		Value:      p.Value,
		ValueFrom:  p.ValueFrom,
		GlobalName: p.GlobalName,
	}, nil
}
