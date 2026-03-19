package model

// EnvVarModel is the serializable K8s EnvVar.
type EnvVarModel struct {
	Name      string        `json:"name" yaml:"name"`
	Value     string        `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

// EnvVarSource represents a source for an environment variable's value.
type EnvVarSource struct {
	SecretKeyRef     *KeySelector           `json:"secretKeyRef,omitempty" yaml:"secretKeyRef,omitempty"`
	ConfigMapKeyRef  *KeySelector           `json:"configMapKeyRef,omitempty" yaml:"configMapKeyRef,omitempty"`
	FieldRef         *FieldSelector         `json:"fieldRef,omitempty" yaml:"fieldRef,omitempty"`
	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty" yaml:"resourceFieldRef,omitempty"`
}

// KeySelector selects a key from a Secret or ConfigMap.
type KeySelector struct {
	Name     string `json:"name" yaml:"name"`
	Key      string `json:"key" yaml:"key"`
	Optional *bool  `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// FieldSelector selects a field from a pod (Downward API).
type FieldSelector struct {
	FieldPath  string `json:"fieldPath" yaml:"fieldPath"`
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

// ResourceFieldSelector selects a resource field from a container.
type ResourceFieldSelector struct {
	Resource      string `json:"resource" yaml:"resource"`
	ContainerName string `json:"containerName,omitempty" yaml:"containerName,omitempty"`
	Divisor       string `json:"divisor,omitempty" yaml:"divisor,omitempty"`
}
