package model

import "fmt"

// ImagePullPolicy defines when to pull the image.
type ImagePullPolicy string

const (
	ImagePullAlways       ImagePullPolicy = "Always"
	ImagePullNever        ImagePullPolicy = "Never"
	ImagePullIfNotPresent ImagePullPolicy = "IfNotPresent"
)

// ParseImagePullPolicy normalizes a string to an ImagePullPolicy.
func ParseImagePullPolicy(s string) (ImagePullPolicy, error) {
	switch s {
	case "Always", "always":
		return ImagePullAlways, nil
	case "Never", "never":
		return ImagePullNever, nil
	case "IfNotPresent", "ifNotPresent", "if_not_present":
		return ImagePullIfNotPresent, nil
	default:
		return "", &InvalidType{Expected: "Always|Never|IfNotPresent", Got: s}
	}
}

// ResourceRequirements specifies CPU/memory requests and limits.
type ResourceRequirements struct {
	Requests ResourceList `json:"requests,omitempty" yaml:"requests,omitempty"`
	Limits   ResourceList `json:"limits,omitempty" yaml:"limits,omitempty"`
}

// ResourceList is a map of resource name to quantity.
type ResourceList struct {
	CPU              string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory           string `json:"memory,omitempty" yaml:"memory,omitempty"`
	EphemeralStorage string `json:"ephemeral-storage,omitempty" yaml:"ephemeral-storage,omitempty"`
}

// Toleration is a K8s toleration.
type Toleration struct {
	Key      string `json:"key,omitempty" yaml:"key,omitempty"`
	Operator string `json:"operator,omitempty" yaml:"operator,omitempty"`
	Value    string `json:"value,omitempty" yaml:"value,omitempty"`
	Effect   string `json:"effect,omitempty" yaml:"effect,omitempty"`
}

// ContainerPort represents a network port in a container.
type ContainerPort struct {
	Name          string `json:"name,omitempty" yaml:"name,omitempty"`
	ContainerPort int32  `json:"containerPort" yaml:"containerPort"`
	Protocol      string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
}

// ImagePullSecret references a K8s secret for pulling images.
type ImagePullSecret struct {
	Name string `json:"name" yaml:"name"`
}

// AccessMode defines the access mode for a PersistentVolumeClaim.
type AccessMode string

const (
	ReadWriteOnce    AccessMode = "ReadWriteOnce"
	ReadOnlyMany     AccessMode = "ReadOnlyMany"
	ReadWriteMany    AccessMode = "ReadWriteMany"
	ReadWriteOncePod AccessMode = "ReadWriteOncePod"
)

// SecretKeySelector references a key in a K8s Secret.
type SecretKeySelector struct {
	Name string `json:"name" yaml:"name"`
	Key  string `json:"key" yaml:"key"`
}

// ArchiveStrategy describes how to archive an artifact.
type ArchiveStrategy struct {
	None *struct{} `json:"none,omitempty" yaml:"none,omitempty"`
	Tar  *struct {
		CompressionLevel *int32 `json:"compressionLevel,omitempty" yaml:"compressionLevel,omitempty"`
	} `json:"tar,omitempty" yaml:"tar,omitempty"`
	Zip *struct{} `json:"zip,omitempty" yaml:"zip,omitempty"`
}

// InvalidType is returned when a wrong type is submitted to a context.
type InvalidType struct {
	Expected string
	Got      string
}

func (e *InvalidType) Error() string {
	return fmt.Sprintf("invalid type: expected %s, got %s", e.Expected, e.Got)
}
