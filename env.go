package forge

import "github.com/usetheodev/theo-forge/model"

// EnvBuilder is the interface for types that build EnvVarModel.
type EnvBuilder interface {
	Build() model.EnvVarModel
}

// Env is a plain environment variable with a literal value.
type Env struct {
	Name  string
	Value string
}

func (e Env) Build() model.EnvVarModel {
	v := e.Value
	return model.EnvVarModel{Name: e.Name, Value: &v}
}

// SecretEnv reads from a K8s Secret.
type SecretEnv struct {
	Name       string
	SecretName string
	SecretKey  string
	Optional   *bool
}

func (e SecretEnv) Build() model.EnvVarModel {
	return model.EnvVarModel{
		Name: e.Name,
		ValueFrom: &model.EnvVarSource{
			SecretKeyRef: &model.KeySelector{
				Name:     e.SecretName,
				Key:      e.SecretKey,
				Optional: e.Optional,
			},
		},
	}
}

// ConfigMapEnv reads from a K8s ConfigMap.
type ConfigMapEnv struct {
	Name          string
	ConfigMapName string
	ConfigMapKey  string
	Optional      *bool
}

func (e ConfigMapEnv) Build() model.EnvVarModel {
	return model.EnvVarModel{
		Name: e.Name,
		ValueFrom: &model.EnvVarSource{
			ConfigMapKeyRef: &model.KeySelector{
				Name:     e.ConfigMapName,
				Key:      e.ConfigMapKey,
				Optional: e.Optional,
			},
		},
	}
}

// FieldEnv reads from a pod field (Downward API).
type FieldEnv struct {
	Name       string
	FieldPath  string
	APIVersion string
}

func (e FieldEnv) Build() model.EnvVarModel {
	return model.EnvVarModel{
		Name: e.Name,
		ValueFrom: &model.EnvVarSource{
			FieldRef: &model.FieldSelector{
				FieldPath:  e.FieldPath,
				APIVersion: e.APIVersion,
			},
		},
	}
}

// ResourceEnv reads from resource constraints.
type ResourceEnv struct {
	Name          string
	Resource      string
	ContainerName string
	Divisor       string
}

func (e ResourceEnv) Build() model.EnvVarModel {
	return model.EnvVarModel{
		Name: e.Name,
		ValueFrom: &model.EnvVarSource{
			ResourceFieldRef: &model.ResourceFieldSelector{
				Resource:      e.Resource,
				ContainerName: e.ContainerName,
				Divisor:       e.Divisor,
			},
		},
	}
}
