package forge

// EnvVarModel is the serializable K8s EnvVar.
type EnvVarModel struct {
	Name      string          `json:"name" yaml:"name"`
	Value     string          `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom *EnvVarSource   `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

// EnvVarSource represents a source for an environment variable's value.
type EnvVarSource struct {
	SecretKeyRef    *KeySelector    `json:"secretKeyRef,omitempty" yaml:"secretKeyRef,omitempty"`
	ConfigMapKeyRef *KeySelector    `json:"configMapKeyRef,omitempty" yaml:"configMapKeyRef,omitempty"`
	FieldRef        *FieldSelector  `json:"fieldRef,omitempty" yaml:"fieldRef,omitempty"`
	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty" yaml:"resourceFieldRef,omitempty"`
}

type KeySelector struct {
	Name     string `json:"name" yaml:"name"`
	Key      string `json:"key" yaml:"key"`
	Optional *bool  `json:"optional,omitempty" yaml:"optional,omitempty"`
}

type FieldSelector struct {
	FieldPath  string `json:"fieldPath" yaml:"fieldPath"`
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

type ResourceFieldSelector struct {
	Resource      string `json:"resource" yaml:"resource"`
	ContainerName string `json:"containerName,omitempty" yaml:"containerName,omitempty"`
	Divisor       string `json:"divisor,omitempty" yaml:"divisor,omitempty"`
}

// EnvBuilder is the interface for types that build EnvVarModel.
type EnvBuilder interface {
	Build() EnvVarModel
}

// Env is a plain environment variable with a literal value.
type Env struct {
	Name  string
	Value string
}

func (e Env) Build() EnvVarModel {
	return EnvVarModel{Name: e.Name, Value: e.Value}
}

// SecretEnv reads from a K8s Secret.
type SecretEnv struct {
	Name       string
	SecretName string
	SecretKey  string
	Optional   *bool
}

func (e SecretEnv) Build() EnvVarModel {
	return EnvVarModel{
		Name: e.Name,
		ValueFrom: &EnvVarSource{
			SecretKeyRef: &KeySelector{
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

func (e ConfigMapEnv) Build() EnvVarModel {
	return EnvVarModel{
		Name: e.Name,
		ValueFrom: &EnvVarSource{
			ConfigMapKeyRef: &KeySelector{
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

func (e FieldEnv) Build() EnvVarModel {
	return EnvVarModel{
		Name: e.Name,
		ValueFrom: &EnvVarSource{
			FieldRef: &FieldSelector{
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

func (e ResourceEnv) Build() EnvVarModel {
	return EnvVarModel{
		Name: e.Name,
		ValueFrom: &EnvVarSource{
			ResourceFieldRef: &ResourceFieldSelector{
				Resource:      e.Resource,
				ContainerName: e.ContainerName,
				Divisor:       e.Divisor,
			},
		},
	}
}
