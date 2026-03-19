package forge

import (
	"fmt"

	"github.com/usetheo/theo/forge/model"
)

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

func (c *ContainerNode) buildModel() model.ContainerModel {
	return model.ContainerModel{
		Name:      c.Name,
		Image:     c.Image,
		Command:   c.Command,
		Args:      c.Args,
		Env:       buildEnvVars(c.Env),
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
func (cs *ContainerSet) BuildTemplate() (model.TemplateModel, error) {
	if cs.Name == "" {
		return model.TemplateModel{}, fmt.Errorf("container set template name cannot be empty")
	}
	if len(cs.Containers) == 0 {
		return model.TemplateModel{}, fmt.Errorf("container set must have at least one container")
	}

	containers := make([]model.ContainerModel, len(cs.Containers))
	for i, c := range cs.Containers {
		containers[i] = c.buildModel()
	}

	var inputs *model.InputsModel
	if len(cs.Inputs) > 0 {
		inputs = &model.InputsModel{}
		for _, p := range cs.Inputs {
			m, err := p.AsInput()
			if err != nil {
				return model.TemplateModel{}, fmt.Errorf("container set %q input parameter %q: %w", cs.Name, p.Name, err)
			}
			inputs.Parameters = append(inputs.Parameters, m)
		}
	}

	var outputs *model.OutputsModel
	if len(cs.Outputs) > 0 {
		outputs = &model.OutputsModel{}
		for _, p := range cs.Outputs {
			m, err := p.AsOutput()
			if err != nil {
				return model.TemplateModel{}, fmt.Errorf("container set %q output parameter %q: %w", cs.Name, p.Name, err)
			}
			outputs.Parameters = append(outputs.Parameters, m)
		}
	}

	var mounts []model.VolumeMountModel
	for _, v := range cs.VolumeMounts {
		mounts = append(mounts, v.BuildVolumeMount())
	}

	return model.TemplateModel{
		Name:    cs.Name,
		Inputs:  inputs,
		Outputs: outputs,
		ContainerSet: &model.ContainerSetModel{
			Containers:   containers,
			VolumeMounts: mounts,
		},
		RetryStrategy: buildRetryStrategyModel(cs.RetryStrategy),
	}, nil
}

