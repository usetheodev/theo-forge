package forge

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
}

// Build creates the serializable ContainerModel.
func (uc *UserContainer) Build() ContainerModel {
	var envs []EnvVarModel
	for _, e := range uc.Env {
		envs = append(envs, e.Build())
	}

	var mounts []VolumeMountModel
	for _, v := range uc.VolumeMounts {
		mounts = append(mounts, v.BuildVolumeMount())
	}

	return ContainerModel{
		Name:            uc.Name,
		Image:           uc.Image,
		Command:         uc.Command,
		Args:            uc.Args,
		WorkingDir:      uc.WorkingDir,
		ImagePullPolicy: string(uc.ImagePullPolicy),
		Env:             envs,
		Resources:       uc.Resources,
		VolumeMounts:    mounts,
		Ports:           uc.Ports,
	}
}
