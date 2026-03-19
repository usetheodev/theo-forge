package forge

import "github.com/usetheo/theo/forge/model"

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
