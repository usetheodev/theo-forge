package forge

import "fmt"

// ContainerSetModel is the serializable Argo ContainerSet.
type ContainerSetModel struct {
	Containers []ContainerModel `json:"containers" yaml:"containers"`
}

// ContainerNode is a container within a ContainerSet.
type ContainerNode struct {
	// Name is the container name.
	Name string
	// Image is the Docker image.
	Image string
	// Command is the entrypoint.
	Command []string
	// Args are the command arguments.
	Args []string
	// Env is the list of environment variables.
	Env []EnvBuilder
	// Resources defines CPU/memory.
	Resources *ResourceRequirements
	// Dependencies are container names that must complete first.
	Dependencies []string
}

func (c *ContainerNode) buildModel() ContainerModel {
	var envs []EnvVarModel
	for _, e := range c.Env {
		envs = append(envs, e.Build())
	}
	return ContainerModel{
		Name:      c.Name,
		Image:     c.Image,
		Command:   c.Command,
		Args:      c.Args,
		Env:       envs,
		Resources: c.Resources,
	}
}

// ContainerSet represents an Argo ContainerSet template — multiple containers in a single pod.
type ContainerSet struct {
	// Name is the template name.
	Name string
	// Containers are the containers in the set.
	Containers []ContainerNode
	// Inputs are the template inputs.
	Inputs []Parameter
	// Outputs are the template outputs.
	Outputs []Parameter
	// VolumeMounts are shared volume mounts for all containers.
	VolumeMounts []VolumeBuilder
	// RetryStrategy configures retry behavior.
	RetryStrategy *RetryStrategy
}

func (cs *ContainerSet) GetName() string {
	return cs.Name
}

// BuildTemplate builds the Argo Template for this ContainerSet.
func (cs *ContainerSet) BuildTemplate() (TemplateModel, error) {
	if cs.Name == "" {
		return TemplateModel{}, fmt.Errorf("container set template name cannot be empty")
	}
	if len(cs.Containers) == 0 {
		return TemplateModel{}, fmt.Errorf("container set must have at least one container")
	}

	containers := make([]ContainerModel, len(cs.Containers))
	for i, c := range cs.Containers {
		containers[i] = c.buildModel()
	}

	var inputs *InputsModel
	if len(cs.Inputs) > 0 {
		inputs = &InputsModel{}
		for _, p := range cs.Inputs {
			m, err := p.AsInput()
			if err != nil {
				continue
			}
			inputs.Parameters = append(inputs.Parameters, m)
		}
	}

	var outputs *OutputsModel
	if len(cs.Outputs) > 0 {
		outputs = &OutputsModel{}
		for _, p := range cs.Outputs {
			m, err := p.AsOutput()
			if err != nil {
				continue
			}
			outputs.Parameters = append(outputs.Parameters, m)
		}
	}

	var mounts []VolumeMountModel
	for _, v := range cs.VolumeMounts {
		mounts = append(mounts, v.BuildVolumeMount())
	}

	var rs *RetryStrategyModel
	if cs.RetryStrategy != nil {
		m := cs.RetryStrategy.Build()
		rs = &m
	}

	return TemplateModel{
		Name:          cs.Name,
		Inputs:        inputs,
		Outputs:       outputs,
		RetryStrategy: rs,
	}, nil
}

// BuildArguments is a helper to build ArgumentsModel from a mix of parameters and artifacts.
func BuildArguments(params []Parameter, artifacts []ArtifactBuilder) (*ArgumentsModel, error) {
	if len(params) == 0 && len(artifacts) == 0 {
		return nil, nil
	}
	args := &ArgumentsModel{}
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

// BuildArgumentsFromMap builds arguments from a map of name→value pairs.
// Values are converted to string Parameter arguments.
func BuildArgumentsFromMap(params map[string]string) *ArgumentsModel {
	if len(params) == 0 {
		return nil
	}
	args := &ArgumentsModel{}
	for k, v := range params {
		val := v
		args.Parameters = append(args.Parameters, ParameterModel{
			Name:  k,
			Value: &val,
		})
	}
	return args
}
