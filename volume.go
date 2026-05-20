package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

// VolumeBuilder is implemented by types that can build a VolumeModel.
type VolumeBuilder interface {
	BuildVolume() (model.VolumeModel, error)
	BuildVolumeMount() model.VolumeMountModel
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
func (v BaseVolume) BuildVolumeMount() model.VolumeMountModel {
	return model.VolumeMountModel{
		Name:             v.Name,
		MountPath:        v.MountPath,
		ReadOnly:         v.ReadOnly,
		SubPath:          v.SubPath,
		SubPathExpr:      v.SubPathExpr,
		MountPropagation: v.MountPropagation,
	}
}

// --- EmptyDir ---

// EmptyDirVolume is the type.
type EmptyDirVolume struct {
	BaseVolume
	Medium    string
	SizeLimit string
}

// BuildVolume is the method.
func (v EmptyDirVolume) BuildVolume() (model.VolumeModel, error) {
	if err := v.validate(); err != nil {
		return model.VolumeModel{}, err
	}
	return model.VolumeModel{
		Name: v.Name,
		EmptyDir: &model.EmptyDirVolumeModel{
			Medium:    v.Medium,
			SizeLimit: v.SizeLimit,
		},
	}, nil
}

// --- HostPath ---

// HostPathVolume is the type.
type HostPathVolume struct {
	BaseVolume
	Path string
	Type string
}

// BuildVolume is the method.
func (v HostPathVolume) BuildVolume() (model.VolumeModel, error) {
	if err := v.validate(); err != nil {
		return model.VolumeModel{}, err
	}
	return model.VolumeModel{
		Name: v.Name,
		HostPath: &model.HostPathVolumeModel{
			Path: v.Path,
			Type: v.Type,
		},
	}, nil
}

// --- ConfigMap ---

// ConfigMapVolume is the type.
type ConfigMapVolume struct {
	BaseVolume
	// ConfigMapName is the name of the ConfigMap to mount. When empty,
	// BaseVolume.Name is used as a fallback (preserves pre-T3.4 behavior).
	// (T3.4 / code-p4-configmapvolume-name-bug — closes the bug where a
	// mount could not reference a ConfigMap with a different name than the
	// volume mount).
	ConfigMapName string
	DefaultMode   *int32
	Optional      *bool
}

// BuildVolume is the method.
func (v ConfigMapVolume) BuildVolume() (model.VolumeModel, error) {
	if err := v.validate(); err != nil {
		return model.VolumeModel{}, err
	}
	cmName := v.ConfigMapName
	if cmName == "" {
		cmName = v.Name
	}
	return model.VolumeModel{
		Name: v.Name,
		ConfigMap: &model.ConfigMapVolumeModel{
			Name:        cmName,
			DefaultMode: v.DefaultMode,
			Optional:    v.Optional,
		},
	}, nil
}

// --- Secret ---

// SecretVolume is the type.
type SecretVolume struct {
	BaseVolume
	SecretName  string
	DefaultMode *int32
	Optional    *bool
}

// BuildVolume is the method.
func (v SecretVolume) BuildVolume() (model.VolumeModel, error) {
	if err := v.validate(); err != nil {
		return model.VolumeModel{}, err
	}
	name := v.SecretName
	if name == "" {
		name = v.Name
	}
	return model.VolumeModel{
		Name: v.Name,
		Secret: &model.SecretVolumeModel{
			SecretName:  name,
			DefaultMode: v.DefaultMode,
			Optional:    v.Optional,
		},
	}, nil
}

// --- ExistingVolume (references existing PVC) ---

// ExistingVolume is the type.
type ExistingVolume struct {
	BaseVolume
	ClaimName string
}

// BuildVolume is the method.
func (v ExistingVolume) BuildVolume() (model.VolumeModel, error) {
	if err := v.validate(); err != nil {
		return model.VolumeModel{}, err
	}
	return model.VolumeModel{
		Name: v.Name,
		PersistentVolumeClaim: &model.PersistentVolumeClaimVolRef{
			ClaimName: v.ClaimName,
			ReadOnly:  v.ReadOnly,
		},
	}, nil
}

// --- PVCVolume (dynamic provisioning) ---

// PVCVolume is the type.
type PVCVolume struct {
	BaseVolume
	Size             string
	StorageClassName string
	AccessModes      []AccessMode
}

// BuildVolume is the method.
func (v PVCVolume) BuildVolume() (model.VolumeModel, error) {
	if err := v.validate(); err != nil {
		return model.VolumeModel{}, err
	}
	return model.VolumeModel{
		Name: v.Name,
		PersistentVolumeClaim: &model.PersistentVolumeClaimVolRef{
			ClaimName: v.Name,
		},
	}, nil
}

// BuildPVC is the method.
func (v PVCVolume) BuildPVC() (model.PVCModel, error) {
	if err := v.validate(); err != nil {
		return model.PVCModel{}, err
	}
	modes := make([]string, len(v.AccessModes))
	for i, m := range v.AccessModes {
		modes[i] = string(m)
	}
	if len(modes) == 0 {
		modes = []string{string(ReadWriteOnce)}
	}
	return model.PVCModel{
		Metadata: model.PVCMetadata{Name: v.Name},
		Spec: model.PVCSpec{
			AccessModes:      modes,
			StorageClassName: v.StorageClassName,
			Resources: model.PVCResources{
				Requests: model.PVCResourceRequest{Storage: v.Size},
			},
		},
	}, nil
}

// --- NFS ---

// NFSVolume is the type.
type NFSVolume struct {
	BaseVolume
	Server string
	Path   string
}

// BuildVolume is the method.
func (v NFSVolume) BuildVolume() (model.VolumeModel, error) {
	if err := v.validate(); err != nil {
		return model.VolumeModel{}, err
	}
	return model.VolumeModel{
		Name: v.Name,
		NFS: &model.NFSVolumeModel{
			Server: v.Server,
			Path:   v.Path,
		},
	}, nil
}
