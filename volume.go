package forge

import "fmt"

// AccessMode defines the access mode for a PersistentVolumeClaim.
type AccessMode string

const (
	ReadWriteOnce    AccessMode = "ReadWriteOnce"
	ReadOnlyMany     AccessMode = "ReadOnlyMany"
	ReadWriteMany    AccessMode = "ReadWriteMany"
	ReadWriteOncePod AccessMode = "ReadWriteOncePod"
)

// VolumeModel is the serializable representation of a K8s Volume.
type VolumeModel struct {
	Name                  string                        `json:"name" yaml:"name"`
	EmptyDir              *EmptyDirVolumeModel          `json:"emptyDir,omitempty" yaml:"emptyDir,omitempty"`
	HostPath              *HostPathVolumeModel          `json:"hostPath,omitempty" yaml:"hostPath,omitempty"`
	ConfigMap             *ConfigMapVolumeModel         `json:"configMap,omitempty" yaml:"configMap,omitempty"`
	Secret                *SecretVolumeModel            `json:"secret,omitempty" yaml:"secret,omitempty"`
	PersistentVolumeClaim *PersistentVolumeClaimVolRef  `json:"persistentVolumeClaim,omitempty" yaml:"persistentVolumeClaim,omitempty"`
	NFS                   *NFSVolumeModel               `json:"nfs,omitempty" yaml:"nfs,omitempty"`
	DownwardAPI           *DownwardAPIVolumeModel       `json:"downwardAPI,omitempty" yaml:"downwardAPI,omitempty"`
	Projected             *ProjectedVolumeModel         `json:"projected,omitempty" yaml:"projected,omitempty"`
}

type EmptyDirVolumeModel struct {
	Medium    string `json:"medium,omitempty" yaml:"medium,omitempty"`
	SizeLimit string `json:"sizeLimit,omitempty" yaml:"sizeLimit,omitempty"`
}

type HostPathVolumeModel struct {
	Path string `json:"path" yaml:"path"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

type ConfigMapVolumeModel struct {
	Name        string `json:"name" yaml:"name"`
	DefaultMode *int32 `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
	Optional    *bool  `json:"optional,omitempty" yaml:"optional,omitempty"`
}

type SecretVolumeModel struct {
	SecretName  string `json:"secretName" yaml:"secretName"`
	DefaultMode *int32 `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
	Optional    *bool  `json:"optional,omitempty" yaml:"optional,omitempty"`
}

type PersistentVolumeClaimVolRef struct {
	ClaimName string `json:"claimName" yaml:"claimName"`
	ReadOnly  bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
}

type NFSVolumeModel struct {
	Server   string `json:"server" yaml:"server"`
	Path     string `json:"path" yaml:"path"`
	ReadOnly bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
}

type DownwardAPIVolumeModel struct {
	DefaultMode *int32 `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
}

type ProjectedVolumeModel struct {
	DefaultMode *int32 `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
}

// VolumeMountModel is the serializable representation of a K8s VolumeMount.
type VolumeMountModel struct {
	Name             string `json:"name" yaml:"name"`
	MountPath        string `json:"mountPath" yaml:"mountPath"`
	ReadOnly         bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	SubPath          string `json:"subPath,omitempty" yaml:"subPath,omitempty"`
	SubPathExpr      string `json:"subPathExpr,omitempty" yaml:"subPathExpr,omitempty"`
	MountPropagation string `json:"mountPropagation,omitempty" yaml:"mountPropagation,omitempty"`
}

// PVCModel is the serializable representation of a K8s PersistentVolumeClaim.
type PVCModel struct {
	Name string  `json:"name" yaml:"name"`
	Spec PVCSpec `json:"spec" yaml:"spec"`
}

type PVCSpec struct {
	AccessModes      []string        `json:"accessModes" yaml:"accessModes"`
	StorageClassName string          `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	Resources        PVCResources    `json:"resources" yaml:"resources"`
}

type PVCResources struct {
	Requests PVCResourceRequest `json:"requests" yaml:"requests"`
}

type PVCResourceRequest struct {
	Storage string `json:"storage" yaml:"storage"`
}

// VolumeBuilder is implemented by types that can build a VolumeModel.
type VolumeBuilder interface {
	BuildVolume() (VolumeModel, error)
	BuildVolumeMount() VolumeMountModel
}

// BaseVolume holds common volume mount fields.
type BaseVolume struct {
	Name             string
	MountPath        string
	ReadOnly         bool
	SubPath          string
	SubPathExpr      string
	MountPropagation string
}

func (v BaseVolume) validate() error {
	if v.Name == "" {
		return fmt.Errorf("volume name cannot be empty")
	}
	return nil
}

// BuildVolumeMount creates a VolumeMountModel from the base fields.
func (v BaseVolume) BuildVolumeMount() VolumeMountModel {
	return VolumeMountModel{
		Name:             v.Name,
		MountPath:        v.MountPath,
		ReadOnly:         v.ReadOnly,
		SubPath:          v.SubPath,
		SubPathExpr:      v.SubPathExpr,
		MountPropagation: v.MountPropagation,
	}
}

// --- EmptyDir ---

type EmptyDirVolume struct {
	BaseVolume
	Medium    string
	SizeLimit string
}

func (v EmptyDirVolume) BuildVolume() (VolumeModel, error) {
	if err := v.validate(); err != nil {
		return VolumeModel{}, err
	}
	return VolumeModel{
		Name: v.Name,
		EmptyDir: &EmptyDirVolumeModel{
			Medium:    v.Medium,
			SizeLimit: v.SizeLimit,
		},
	}, nil
}

// --- HostPath ---

type HostPathVolume struct {
	BaseVolume
	Path string
	Type string
}

func (v HostPathVolume) BuildVolume() (VolumeModel, error) {
	if err := v.validate(); err != nil {
		return VolumeModel{}, err
	}
	return VolumeModel{
		Name: v.Name,
		HostPath: &HostPathVolumeModel{
			Path: v.Path,
			Type: v.Type,
		},
	}, nil
}

// --- ConfigMap ---

type ConfigMapVolume struct {
	BaseVolume
	DefaultMode *int32
	Optional    *bool
}

func (v ConfigMapVolume) BuildVolume() (VolumeModel, error) {
	if err := v.validate(); err != nil {
		return VolumeModel{}, err
	}
	return VolumeModel{
		Name: v.Name,
		ConfigMap: &ConfigMapVolumeModel{
			Name:        v.Name,
			DefaultMode: v.DefaultMode,
			Optional:    v.Optional,
		},
	}, nil
}

// --- Secret ---

type SecretVolume struct {
	BaseVolume
	SecretName  string
	DefaultMode *int32
	Optional    *bool
}

func (v SecretVolume) BuildVolume() (VolumeModel, error) {
	if err := v.validate(); err != nil {
		return VolumeModel{}, err
	}
	name := v.SecretName
	if name == "" {
		name = v.Name
	}
	return VolumeModel{
		Name: v.Name,
		Secret: &SecretVolumeModel{
			SecretName:  name,
			DefaultMode: v.DefaultMode,
			Optional:    v.Optional,
		},
	}, nil
}

// --- ExistingVolume (references existing PVC) ---

type ExistingVolume struct {
	BaseVolume
	ClaimName string
}

func (v ExistingVolume) BuildVolume() (VolumeModel, error) {
	if err := v.validate(); err != nil {
		return VolumeModel{}, err
	}
	return VolumeModel{
		Name: v.Name,
		PersistentVolumeClaim: &PersistentVolumeClaimVolRef{
			ClaimName: v.ClaimName,
			ReadOnly:  v.ReadOnly,
		},
	}, nil
}

// --- PVCVolume (dynamic provisioning) ---

type PVCVolume struct {
	BaseVolume
	Size             string
	StorageClassName string
	AccessModes      []AccessMode
}

func (v PVCVolume) BuildVolume() (VolumeModel, error) {
	if err := v.validate(); err != nil {
		return VolumeModel{}, err
	}
	return VolumeModel{
		Name: v.Name,
		PersistentVolumeClaim: &PersistentVolumeClaimVolRef{
			ClaimName: v.Name,
		},
	}, nil
}

func (v PVCVolume) BuildPVC() (PVCModel, error) {
	if err := v.validate(); err != nil {
		return PVCModel{}, err
	}
	modes := make([]string, len(v.AccessModes))
	for i, m := range v.AccessModes {
		modes[i] = string(m)
	}
	if len(modes) == 0 {
		modes = []string{string(ReadWriteOnce)}
	}
	return PVCModel{
		Name: v.Name,
		Spec: PVCSpec{
			AccessModes:      modes,
			StorageClassName: v.StorageClassName,
			Resources: PVCResources{
				Requests: PVCResourceRequest{Storage: v.Size},
			},
		},
	}, nil
}

// --- NFS ---

type NFSVolume struct {
	BaseVolume
	Server string
	Path   string
}

func (v NFSVolume) BuildVolume() (VolumeModel, error) {
	if err := v.validate(); err != nil {
		return VolumeModel{}, err
	}
	return VolumeModel{
		Name: v.Name,
		NFS: &NFSVolumeModel{
			Server: v.Server,
			Path:   v.Path,
		},
	}, nil
}
