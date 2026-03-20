package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/model"
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

// --- Retry ---

// RetryStrategy configures retry behavior for templates.
type RetryStrategy struct {
	Limit       *int
	RetryPolicy RetryPolicy
	Backoff     *Backoff
	Expression  string
}

// Build converts RetryStrategy to its serializable model.
func (r RetryStrategy) Build() model.RetryStrategyModel {
	var limit interface{}
	if r.Limit != nil {
		limit = fmt.Sprintf("%d", *r.Limit)
	}
	var backoff *model.Backoff
	if r.Backoff != nil {
		b := *r.Backoff
		// Normalize factor to string for Argo compatibility
		if factor, ok := b.Factor.(*int); ok && factor != nil {
			b.Factor = fmt.Sprintf("%d", *factor)
		} else if factor, ok := b.Factor.(int); ok {
			b.Factor = fmt.Sprintf("%d", factor)
		}
		backoff = &b
	}
	return model.RetryStrategyModel{
		Limit:       limit,
		RetryPolicy: string(r.RetryPolicy),
		Backoff:     backoff,
		Expression:  r.Expression,
	}
}

// --- User Container ---

// UserContainer represents a sidecar or init container in a template.
type UserContainer struct {
	// Name is the container name.
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
	// Resources defines CPU/memory.
	Resources *ResourceRequirements
	// VolumeMounts are the volume mounts.
	VolumeMounts []VolumeBuilder
	// Ports exposed by the container.
	Ports []ContainerPort
	// Mirror enables mirroring volume mounts from the main container.
	Mirror *bool
	// SecurityContext for the container.
	SecurityContext *model.SecurityContext
	// Lifecycle defines actions for container lifecycle events.
	Lifecycle *model.Lifecycle
	// ReadinessProbe for the container.
	ReadinessProbe *model.Probe
	// Daemon marks this sidecar as a daemon.
	Daemon *bool
}

// Build creates the serializable ContainerModel.
func (uc *UserContainer) Build() model.ContainerModel {
	return model.ContainerModel{
		Name:            uc.Name,
		Image:           uc.Image,
		Command:         uc.Command,
		Args:            uc.Args,
		WorkingDir:      uc.WorkingDir,
		ImagePullPolicy: string(uc.ImagePullPolicy),
		Env:             buildEnvVars(uc.Env),
		Resources:       uc.Resources,
		VolumeMounts:    buildVolumeMountModels(uc.VolumeMounts),
		Ports:           uc.Ports,
		SecurityContext: uc.SecurityContext,
		Mirror:          uc.Mirror,
		Lifecycle:       uc.Lifecycle,
		ReadinessProbe:  uc.ReadinessProbe,
	}
}
