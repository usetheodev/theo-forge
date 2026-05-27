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

// --- Projected (with ServiceAccountToken support) ---

// ProjectedVolume is a volume that aggregates multiple sources (ServiceAccountToken,
// ConfigMap, Secret, DownwardAPI) into a single mount point.
//
// Primary use case in Theo: Sigstore cosign keyless OIDC signing.
// Each build worker pod mounts a SA token with audience="sigstore" so cosign
// can present it to Fulcio for short-lived signing certificate issuance,
// instead of using a long-lived static key.
//
// Example (SA token for Sigstore Fulcio):
//
//	forge.ProjectedVolume{
//	    BaseVolume: forge.BaseVolume{Name: "sigstore-token", MountPath: "/var/run/secrets/sigstore"},
//	    Sources: []forge.VolumeProjection{
//	        {ServiceAccountToken: &forge.ServiceAccountTokenProjection{
//	            Audience: "sigstore",
//	            Path:     "token",
//	            ExpirationSeconds: forge.Ptr(int64(600)),
//	        }},
//	    },
//	}
type ProjectedVolume struct {
	BaseVolume
	// DefaultMode is the optional default mode bits for files (octal).
	DefaultMode *int32
	// Sources is the list of volume projections aggregated into this mount.
	// Each entry MUST set exactly one of ServiceAccountToken/ConfigMap/Secret/DownwardAPI.
	Sources []VolumeProjection
}

// VolumeProjection is one source within a ProjectedVolume.
// Exactly one of ServiceAccountToken, ConfigMap, Secret, DownwardAPI must be set.
type VolumeProjection struct {
	ServiceAccountToken *ServiceAccountTokenProjection
	ConfigMap           *ConfigMapProjection
	Secret              *SecretProjection
	DownwardAPI         *DownwardAPIProjection
}

// ServiceAccountTokenProjection requests a SA token from kubelet, written to
// the volume at the given Path. Audience MUST match Fulcio's expected
// audience for Sigstore keyless (typically "sigstore").
type ServiceAccountTokenProjection struct {
	// Audience is the intended audience of the token. For Sigstore, "sigstore".
	Audience string
	// ExpirationSeconds is the requested duration of validity (min 600, default 3600).
	// Kubelet rotates the token if older than 80% of expiration.
	ExpirationSeconds *int64
	// Path is the file path relative to the mount where the token is written.
	Path string
}

// ConfigMapProjection projects a ConfigMap into the volume. Reuses
// ConfigMapVolumeModel — items/optional/defaultMode fields supported via
// underlying ConfigMapVolumeModel.
type ConfigMapProjection struct {
	Name        string
	DefaultMode *int32
	Optional    *bool
}

// SecretProjection projects a Secret into the volume.
type SecretProjection struct {
	SecretName  string
	DefaultMode *int32
	Optional    *bool
}

// DownwardAPIProjection projects downwardAPI fields into the volume.
type DownwardAPIProjection struct {
	DefaultMode *int32
}

// BuildVolume implements VolumeBuilder for ProjectedVolume.
func (v ProjectedVolume) BuildVolume() (model.VolumeModel, error) {
	if err := v.validate(); err != nil {
		return model.VolumeModel{}, err
	}
	sources := make([]model.VolumeProjectionModel, 0, len(v.Sources))
	for i, s := range v.Sources {
		count := 0
		if s.ServiceAccountToken != nil {
			count++
		}
		if s.ConfigMap != nil {
			count++
		}
		if s.Secret != nil {
			count++
		}
		if s.DownwardAPI != nil {
			count++
		}
		if count != 1 {
			return model.VolumeModel{}, fmt.Errorf("projected volume %q source[%d]: exactly one of ServiceAccountToken/ConfigMap/Secret/DownwardAPI must be set (got %d)", v.Name, i, count)
		}

		var proj model.VolumeProjectionModel
		if s.ServiceAccountToken != nil {
			if s.ServiceAccountToken.Path == "" {
				return model.VolumeModel{}, fmt.Errorf("projected volume %q source[%d] ServiceAccountToken: Path is required", v.Name, i)
			}
			proj.ServiceAccountToken = &model.ServiceAccountTokenProjectionModel{
				Audience:          s.ServiceAccountToken.Audience,
				ExpirationSeconds: s.ServiceAccountToken.ExpirationSeconds,
				Path:              s.ServiceAccountToken.Path,
			}
		}
		if s.ConfigMap != nil {
			proj.ConfigMap = &model.ConfigMapVolumeModel{
				Name:        s.ConfigMap.Name,
				DefaultMode: s.ConfigMap.DefaultMode,
				Optional:    s.ConfigMap.Optional,
			}
		}
		if s.Secret != nil {
			proj.Secret = &model.SecretVolumeModel{
				SecretName:  s.Secret.SecretName,
				DefaultMode: s.Secret.DefaultMode,
				Optional:    s.Secret.Optional,
			}
		}
		if s.DownwardAPI != nil {
			proj.DownwardAPI = &model.DownwardAPIVolumeModel{
				DefaultMode: s.DownwardAPI.DefaultMode,
			}
		}
		sources = append(sources, proj)
	}

	return model.VolumeModel{
		Name: v.Name,
		Projected: &model.ProjectedVolumeModel{
			DefaultMode: v.DefaultMode,
			Sources:     sources,
		},
	}, nil
}
