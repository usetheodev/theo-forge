package model

// VolumeModel is the serializable representation of a K8s Volume.
type VolumeModel struct {
	Name                  string                       `json:"name" yaml:"name"`
	EmptyDir              *EmptyDirVolumeModel         `json:"emptyDir,omitempty" yaml:"emptyDir,omitempty"`
	HostPath              *HostPathVolumeModel         `json:"hostPath,omitempty" yaml:"hostPath,omitempty"`
	ConfigMap             *ConfigMapVolumeModel        `json:"configMap,omitempty" yaml:"configMap,omitempty"`
	Secret                *SecretVolumeModel           `json:"secret,omitempty" yaml:"secret,omitempty"`
	PersistentVolumeClaim *PersistentVolumeClaimVolRef `json:"persistentVolumeClaim,omitempty" yaml:"persistentVolumeClaim,omitempty"`
	NFS                   *NFSVolumeModel              `json:"nfs,omitempty" yaml:"nfs,omitempty"`
	DownwardAPI           *DownwardAPIVolumeModel      `json:"downwardAPI,omitempty" yaml:"downwardAPI,omitempty"`
	Projected             *ProjectedVolumeModel        `json:"projected,omitempty" yaml:"projected,omitempty"`
}

// EmptyDirVolumeModel is the serializable emptyDir volume source.
type EmptyDirVolumeModel struct {
	Medium    string `json:"medium,omitempty" yaml:"medium,omitempty"`
	SizeLimit string `json:"sizeLimit,omitempty" yaml:"sizeLimit,omitempty"`
}

// HostPathVolumeModel is the serializable hostPath volume source.
type HostPathVolumeModel struct {
	Path string `json:"path" yaml:"path"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// ConfigMapVolumeModel is the serializable configMap volume source.
type ConfigMapVolumeModel struct {
	Name        string `json:"name" yaml:"name"`
	DefaultMode *int32 `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
	Optional    *bool  `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// SecretVolumeModel is the serializable secret volume source.
type SecretVolumeModel struct {
	SecretName  string `json:"secretName" yaml:"secretName"`
	DefaultMode *int32 `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
	Optional    *bool  `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// PersistentVolumeClaimVolRef is the serializable PVC volume reference.
type PersistentVolumeClaimVolRef struct {
	ClaimName string `json:"claimName" yaml:"claimName"`
	ReadOnly  bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
}

// NFSVolumeModel is the serializable NFS volume source.
type NFSVolumeModel struct {
	Server   string `json:"server" yaml:"server"`
	Path     string `json:"path" yaml:"path"`
	ReadOnly bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
}

// DownwardAPIVolumeModel is the serializable downwardAPI volume source.
type DownwardAPIVolumeModel struct {
	DefaultMode *int32 `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
}

// ProjectedVolumeModel is the serializable projected volume source.
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
	Metadata PVCMetadata `json:"metadata" yaml:"metadata"`
	Spec     PVCSpec     `json:"spec" yaml:"spec"`
}

// PVCMetadata is the metadata for a PVC.
type PVCMetadata struct {
	Name string `json:"name" yaml:"name"`
}

// PVCSpec is the PVC specification.
type PVCSpec struct {
	AccessModes      []string     `json:"accessModes" yaml:"accessModes"`
	StorageClassName string       `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	Resources        PVCResources `json:"resources" yaml:"resources"`
}

// PVCResources is the PVC resource requirements.
type PVCResources struct {
	Requests PVCResourceRequest `json:"requests" yaml:"requests"`
}

// PVCResourceRequest is the PVC storage request.
type PVCResourceRequest struct {
	Storage string `json:"storage" yaml:"storage"`
}
