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

type EmptyDirVolume struct {
	BaseVolume
	Medium    string
	SizeLimit string
}

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

type HostPathVolume struct {
	BaseVolume
	Path string
	Type string
}

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

type ConfigMapVolume struct {
	BaseVolume
	DefaultMode *int32
	Optional    *bool
}

func (v ConfigMapVolume) BuildVolume() (model.VolumeModel, error) {
	if err := v.validate(); err != nil {
		return model.VolumeModel{}, err
	}
	return model.VolumeModel{
		Name: v.Name,
		ConfigMap: &model.ConfigMapVolumeModel{
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

type ExistingVolume struct {
	BaseVolume
	ClaimName string
}

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

type PVCVolume struct {
	BaseVolume
	Size             string
	StorageClassName string
	AccessModes      []AccessMode
}

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

type NFSVolume struct {
	BaseVolume
	Server string
	Path   string
}

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
