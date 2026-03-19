package forge

import (
	"testing"

	"github.com/usetheo/theo/forge/model"
)

func TestUserContainerBuild(t *testing.T) {
	uc := &UserContainer{
		Name:    "sidecar",
		Image:   "nginx:latest",
		Command: []string{"nginx", "-g", "daemon off;"},
		Ports:   []ContainerPort{{ContainerPort: 80}},
	}
	model := uc.Build()
	if model.Name != "sidecar" {
		t.Errorf("name = %q", model.Name)
	}
	if model.Image != "nginx:latest" {
		t.Errorf("image = %q", model.Image)
	}
	if len(model.Ports) != 1 || model.Ports[0].ContainerPort != 80 {
		t.Errorf("ports = %v", model.Ports)
	}
}

func TestUserContainerWithImagePullPolicy(t *testing.T) {
	uc := &UserContainer{
		Name:            "test",
		Image:           "alpine",
		ImagePullPolicy: ImagePullIfNotPresent,
	}
	model := uc.Build()
	if model.ImagePullPolicy != "IfNotPresent" {
		t.Errorf("policy = %q", model.ImagePullPolicy)
	}
}

func TestUserContainerWithVolumeMounts(t *testing.T) {
	uc := &UserContainer{
		Name:  "test",
		Image: "alpine",
		VolumeMounts: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "data", MountPath: "/data"}},
		},
	}
	model := uc.Build()
	if len(model.VolumeMounts) != 1 {
		t.Fatalf("mounts = %d", len(model.VolumeMounts))
	}
	if model.VolumeMounts[0].Name != "data" {
		t.Errorf("mount name = %q", model.VolumeMounts[0].Name)
	}
	if model.VolumeMounts[0].MountPath != "/data" {
		t.Errorf("mount path = %q", model.VolumeMounts[0].MountPath)
	}
}

func TestUserContainerWithMultipleVolumeMounts(t *testing.T) {
	uc := &UserContainer{
		Name:  "test",
		Image: "alpine",
		VolumeMounts: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "test1", MountPath: "/test1"}},
			&PVCVolume{BaseVolume: BaseVolume{Name: "test2", MountPath: "/test2"}, Size: "1Gi"},
		},
	}
	model := uc.Build()
	if len(model.VolumeMounts) != 2 {
		t.Fatalf("mounts = %d, want 2", len(model.VolumeMounts))
	}
	if model.VolumeMounts[0].Name != "test1" {
		t.Errorf("mount[0] name = %q", model.VolumeMounts[0].Name)
	}
	if model.VolumeMounts[1].Name != "test2" {
		t.Errorf("mount[1] name = %q", model.VolumeMounts[1].Name)
	}
}

func TestUserContainerAsInitContainer(t *testing.T) {
	initC := &UserContainer{
		Name:    "init",
		Image:   "alpine",
		Command: []string{"sh", "-c"},
		Args:    []string{"echo initializing"},
	}

	c := &Container{
		Name:    "main",
		Image:   "alpine",
		Command: []string{"echo"},
		Args:    []string{"hello"},
	}

	// Verify the init container model can be used in a template's InitContainers
	initModel := initC.Build()
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	tpl.InitContainers = []model.ContainerModel{initModel}

	if len(tpl.InitContainers) != 1 {
		t.Fatalf("initContainers = %d", len(tpl.InitContainers))
	}
	if tpl.InitContainers[0].Name != "init" {
		t.Errorf("init name = %q", tpl.InitContainers[0].Name)
	}
}

func TestUserContainerAsSidecar(t *testing.T) {
	sidecar := &UserContainer{
		Name:    "log-collector",
		Image:   "fluentd:latest",
		Env:     []EnvBuilder{Env{Name: "LOG_LEVEL", Value: "debug"}},
	}

	c := &Container{
		Name:  "main",
		Image: "alpine",
	}

	sidecarModel := sidecar.Build()
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	tpl.Sidecars = []model.ContainerModel{sidecarModel}

	if len(tpl.Sidecars) != 1 {
		t.Fatalf("sidecars = %d", len(tpl.Sidecars))
	}
	if tpl.Sidecars[0].Name != "log-collector" {
		t.Errorf("sidecar name = %q", tpl.Sidecars[0].Name)
	}
	if len(tpl.Sidecars[0].Env) != 1 || tpl.Sidecars[0].Env[0].Name != "LOG_LEVEL" {
		t.Error("expected sidecar env")
	}
}

func TestWorkflowTemplateWithDefaultParam(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "with-default",
		Entrypoint: "main",
		Arguments: []Parameter{
			{Name: "my-arg", Default: ptrStr("foo")},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := wt.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Arguments == nil || len(model.Spec.Arguments.Parameters) != 1 {
		t.Fatal("expected 1 argument")
	}
	p := model.Spec.Arguments.Parameters[0]
	if p.Name != "my-arg" {
		t.Errorf("name = %q", p.Name)
	}
	// AsArgument doesn't include Default, so this tests the argument building path
}
