package forge

import (
	"encoding/json"
	"testing"
)

func TestAccessModeValues(t *testing.T) {
	tests := []struct {
		mode AccessMode
		want string
	}{
		{ReadWriteOnce, "ReadWriteOnce"},
		{ReadOnlyMany, "ReadOnlyMany"},
		{ReadWriteMany, "ReadWriteMany"},
		{ReadWriteOncePod, "ReadWriteOncePod"},
	}
	for _, tt := range tests {
		if string(tt.mode) != tt.want {
			t.Errorf("AccessMode(%q) = %q, want %q", tt.mode, string(tt.mode), tt.want)
		}
	}
}

func TestEmptyDirVolumeBuild(t *testing.T) {
	v := EmptyDirVolume{
		BaseVolume: BaseVolume{Name: "scratch", MountPath: "/tmp/scratch"},
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatal(err)
	}
	if vol.Name != "scratch" {
		t.Errorf("name = %q, want 'scratch'", vol.Name)
	}
	if vol.EmptyDir == nil {
		t.Fatal("expected emptyDir to be set")
	}

	mount := v.BuildVolumeMount()
	if mount.Name != "scratch" {
		t.Errorf("mount name = %q", mount.Name)
	}
	if mount.MountPath != "/tmp/scratch" {
		t.Errorf("mount path = %q", mount.MountPath)
	}
}

func TestEmptyDirVolumeWithMedium(t *testing.T) {
	v := EmptyDirVolume{
		BaseVolume: BaseVolume{Name: "mem", MountPath: "/dev/shm"},
		Medium:     "Memory",
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatal(err)
	}
	if vol.EmptyDir.Medium != "Memory" {
		t.Errorf("medium = %q, want 'Memory'", vol.EmptyDir.Medium)
	}
}

func TestConfigMapVolumeBuild(t *testing.T) {
	v := ConfigMapVolume{
		BaseVolume: BaseVolume{Name: "config", MountPath: "/etc/config"},
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatal(err)
	}
	if vol.ConfigMap == nil {
		t.Fatal("expected configMap to be set")
	}
	if vol.ConfigMap.Name != "config" {
		t.Errorf("configMap name = %q, want 'config'", vol.ConfigMap.Name)
	}
}

func TestSecretVolumeBuild(t *testing.T) {
	v := SecretVolume{
		BaseVolume: BaseVolume{Name: "creds", MountPath: "/etc/creds"},
		SecretName: "my-secret",
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatal(err)
	}
	if vol.Secret == nil {
		t.Fatal("expected secret to be set")
	}
	if vol.Secret.SecretName != "my-secret" {
		t.Errorf("secretName = %q, want 'my-secret'", vol.Secret.SecretName)
	}
}

func TestHostPathVolumeBuild(t *testing.T) {
	v := HostPathVolume{
		BaseVolume: BaseVolume{Name: "host", MountPath: "/host-data"},
		Path:       "/data",
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatal(err)
	}
	if vol.HostPath == nil {
		t.Fatal("expected hostPath to be set")
	}
	if vol.HostPath.Path != "/data" {
		t.Errorf("path = %q, want '/data'", vol.HostPath.Path)
	}
}

func TestExistingVolumeBuild(t *testing.T) {
	v := ExistingVolume{
		BaseVolume: BaseVolume{Name: "pvc", MountPath: "/data"},
		ClaimName:  "my-pvc",
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatal(err)
	}
	if vol.PersistentVolumeClaim == nil {
		t.Fatal("expected PVC to be set")
	}
	if vol.PersistentVolumeClaim.ClaimName != "my-pvc" {
		t.Errorf("claimName = %q", vol.PersistentVolumeClaim.ClaimName)
	}
}

func TestPVCVolumeBuild(t *testing.T) {
	v := PVCVolume{
		BaseVolume:       BaseVolume{Name: "dynamic", MountPath: "/data"},
		Size:             "10Gi",
		StorageClassName: "standard",
		AccessModes:      []AccessMode{ReadWriteOnce},
	}
	pvc, err := v.BuildPVC()
	if err != nil {
		t.Fatal(err)
	}
	if pvc.Metadata.Name != "dynamic" {
		t.Errorf("pvc name = %q", pvc.Metadata.Name)
	}
	if pvc.Spec.Resources.Requests.Storage != "10Gi" {
		t.Errorf("storage = %q, want '10Gi'", pvc.Spec.Resources.Requests.Storage)
	}
	if pvc.Spec.StorageClassName != "standard" {
		t.Errorf("storageClass = %q", pvc.Spec.StorageClassName)
	}
	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != "ReadWriteOnce" {
		t.Errorf("accessModes = %v", pvc.Spec.AccessModes)
	}
}

func TestVolumeMountReadOnly(t *testing.T) {
	v := EmptyDirVolume{
		BaseVolume: BaseVolume{Name: "ro", MountPath: "/ro", ReadOnly: true},
	}
	mount := v.BuildVolumeMount()
	if !mount.ReadOnly {
		t.Error("expected read-only mount")
	}
}

func TestVolumeMountSubPath(t *testing.T) {
	v := ConfigMapVolume{
		BaseVolume: BaseVolume{Name: "cfg", MountPath: "/cfg", SubPath: "app.conf"},
	}
	mount := v.BuildVolumeMount()
	if mount.SubPath != "app.conf" {
		t.Errorf("subPath = %q, want 'app.conf'", mount.SubPath)
	}
}

func TestVolumeNoNameFails(t *testing.T) {
	v := EmptyDirVolume{
		BaseVolume: BaseVolume{MountPath: "/tmp"},
	}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestVolumeModelJSON(t *testing.T) {
	v := SecretVolume{
		BaseVolume: BaseVolume{Name: "creds", MountPath: "/etc/creds"},
		SecretName: "my-secret",
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(vol)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["name"] != "creds" {
		t.Errorf("json name = %v", m["name"])
	}
}

func TestNFSVolumeBuild(t *testing.T) {
	v := NFSVolume{
		BaseVolume: BaseVolume{Name: "nfs", MountPath: "/nfs"},
		Server:     "nfs.example.com",
		Path:       "/exports/data",
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatal(err)
	}
	if vol.NFS == nil {
		t.Fatal("expected NFS to be set")
	}
	if vol.NFS.Server != "nfs.example.com" {
		t.Errorf("server = %q", vol.NFS.Server)
	}
}

// --- ProjectedVolume tests ---

func TestProjectedVolume_ServiceAccountToken(t *testing.T) {
	expSeconds := int64(600)
	v := ProjectedVolume{
		BaseVolume: BaseVolume{Name: "sigstore-token", MountPath: "/var/run/secrets/sigstore"},
		Sources: []VolumeProjection{
			{ServiceAccountToken: &ServiceAccountTokenProjection{
				Audience:          "sigstore",
				Path:              "token",
				ExpirationSeconds: &expSeconds,
			}},
		},
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatalf("BuildVolume failed: %v", err)
	}
	if vol.Name != "sigstore-token" {
		t.Errorf("name = %q, want sigstore-token", vol.Name)
	}
	if vol.Projected == nil {
		t.Fatal("expected Projected to be set")
	}
	if len(vol.Projected.Sources) != 1 {
		t.Fatalf("sources len = %d, want 1", len(vol.Projected.Sources))
	}
	sat := vol.Projected.Sources[0].ServiceAccountToken
	if sat == nil {
		t.Fatal("expected ServiceAccountToken source to be set")
	}
	if sat.Audience != "sigstore" {
		t.Errorf("audience = %q, want sigstore", sat.Audience)
	}
	if sat.Path != "token" {
		t.Errorf("path = %q, want token", sat.Path)
	}
	if sat.ExpirationSeconds == nil || *sat.ExpirationSeconds != 600 {
		t.Errorf("ExpirationSeconds = %v, want 600", sat.ExpirationSeconds)
	}
}

func TestProjectedVolume_ServiceAccountToken_PathRequired(t *testing.T) {
	v := ProjectedVolume{
		BaseVolume: BaseVolume{Name: "sigstore-token", MountPath: "/var/run/secrets/sigstore"},
		Sources: []VolumeProjection{
			{ServiceAccountToken: &ServiceAccountTokenProjection{Audience: "sigstore"}},
		},
	}
	if _, err := v.BuildVolume(); err == nil {
		t.Fatal("expected error: Path is required")
	}
}

func TestProjectedVolume_ExactlyOneSource(t *testing.T) {
	// Zero sources in projection → error
	v := ProjectedVolume{
		BaseVolume: BaseVolume{Name: "x", MountPath: "/x"},
		Sources:    []VolumeProjection{{}},
	}
	if _, err := v.BuildVolume(); err == nil {
		t.Fatal("expected error: exactly one source must be set (got 0)")
	}

	// Two sources in same projection → error
	v2 := ProjectedVolume{
		BaseVolume: BaseVolume{Name: "x", MountPath: "/x"},
		Sources: []VolumeProjection{{
			ServiceAccountToken: &ServiceAccountTokenProjection{Audience: "a", Path: "t"},
			ConfigMap:           &ConfigMapProjection{Name: "cm"},
		}},
	}
	if _, err := v2.BuildVolume(); err == nil {
		t.Fatal("expected error: exactly one source must be set (got 2)")
	}
}

func TestProjectedVolume_NameRequired(t *testing.T) {
	v := ProjectedVolume{
		Sources: []VolumeProjection{
			{ServiceAccountToken: &ServiceAccountTokenProjection{Audience: "sigstore", Path: "token"}},
		},
	}
	if _, err := v.BuildVolume(); err == nil {
		t.Fatal("expected error: volume name cannot be empty")
	}
}

func TestProjectedVolume_MultipleSources(t *testing.T) {
	v := ProjectedVolume{
		BaseVolume: BaseVolume{Name: "multi", MountPath: "/multi"},
		Sources: []VolumeProjection{
			{ServiceAccountToken: &ServiceAccountTokenProjection{Audience: "sigstore", Path: "token"}},
			{ConfigMap: &ConfigMapProjection{Name: "extra-config"}},
			{Secret: &SecretProjection{SecretName: "extra-secret"}},
		},
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatalf("BuildVolume failed: %v", err)
	}
	if len(vol.Projected.Sources) != 3 {
		t.Fatalf("sources len = %d, want 3", len(vol.Projected.Sources))
	}
	if vol.Projected.Sources[0].ServiceAccountToken == nil {
		t.Error("expected sources[0] to be SAT")
	}
	if vol.Projected.Sources[1].ConfigMap == nil {
		t.Error("expected sources[1] to be ConfigMap")
	}
	if vol.Projected.Sources[2].Secret == nil {
		t.Error("expected sources[2] to be Secret")
	}
}

func TestProjectedVolume_VolumeMount(t *testing.T) {
	v := ProjectedVolume{
		BaseVolume: BaseVolume{Name: "sigstore-token", MountPath: "/var/run/secrets/sigstore", ReadOnly: true},
		Sources: []VolumeProjection{
			{ServiceAccountToken: &ServiceAccountTokenProjection{Audience: "sigstore", Path: "token"}},
		},
	}
	vm := v.BuildVolumeMount()
	if vm.Name != "sigstore-token" {
		t.Errorf("mount name = %q", vm.Name)
	}
	if vm.MountPath != "/var/run/secrets/sigstore" {
		t.Errorf("mount path = %q", vm.MountPath)
	}
	if !vm.ReadOnly {
		t.Error("expected ReadOnly=true")
	}
}
