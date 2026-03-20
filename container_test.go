package forge

import (
	"encoding/json"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

func TestContainerBuildTemplate(t *testing.T) {
	c := &Container{
		Name:    "hello",
		Image:   "python:3.11",
		Command: []string{"python", "-c"},
		Args:    []string{"print('hello')"},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "hello" {
		t.Errorf("name = %q", tpl.Name)
	}
	if tpl.Container == nil {
		t.Fatal("expected container to be set")
	}
	if tpl.Container.Image != "python:3.11" {
		t.Errorf("image = %q", tpl.Container.Image)
	}
	if len(tpl.Container.Command) != 2 {
		t.Errorf("command len = %d", len(tpl.Container.Command))
	}
}

func TestContainerNoNameFails(t *testing.T) {
	c := &Container{Image: "alpine"}
	_, err := c.BuildTemplate()
	if err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestContainerWithEnv(t *testing.T) {
	c := &Container{
		Name:  "with-env",
		Image: "alpine",
		Env: []EnvBuilder{
			Env{Name: "FOO", Value: "bar"},
			SecretEnv{Name: "SECRET", SecretName: "my-secret", SecretKey: "key"},
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tpl.Container.Env) != 2 {
		t.Fatalf("env count = %d, want 2", len(tpl.Container.Env))
	}
	if tpl.Container.Env[0].Name != "FOO" {
		t.Errorf("env[0].name = %q", tpl.Container.Env[0].Name)
	}
	if tpl.Container.Env[1].ValueFrom == nil || tpl.Container.Env[1].ValueFrom.SecretKeyRef == nil {
		t.Error("expected env[1] to be a secret ref")
	}
}

func TestContainerWithResources(t *testing.T) {
	c := &Container{
		Name:  "with-resources",
		Image: "alpine",
		Resources: &ResourceRequirements{
			Requests: ResourceList{CPU: "100m", Memory: "256Mi"},
			Limits:   ResourceList{CPU: "500m", Memory: "1Gi"},
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Container.Resources == nil {
		t.Fatal("expected resources")
	}
	if tpl.Container.Resources.Requests.CPU != "100m" {
		t.Errorf("requests.cpu = %q", tpl.Container.Resources.Requests.CPU)
	}
	if tpl.Container.Resources.Limits.Memory != "1Gi" {
		t.Errorf("limits.memory = %q", tpl.Container.Resources.Limits.Memory)
	}
}

func TestContainerWithInputs(t *testing.T) {
	c := &Container{
		Name:  "with-inputs",
		Image: "alpine",
		Inputs: []Parameter{
			{Name: "msg", Value: ptrStr("hello")},
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Parameters) != 1 {
		t.Fatal("expected 1 input parameter")
	}
	if tpl.Inputs.Parameters[0].Name != "msg" {
		t.Errorf("input name = %q", tpl.Inputs.Parameters[0].Name)
	}
}

func TestContainerWithRetryStrategy(t *testing.T) {
	limit := 3
	c := &Container{
		Name:  "with-retry",
		Image: "alpine",
		RetryStrategy: &RetryStrategy{
			Limit:       &limit,
			RetryPolicy: RetryOnFailure,
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.RetryStrategy == nil {
		t.Fatal("expected retry strategy")
	}
	if tpl.RetryStrategy.Limit != "3" {
		t.Errorf("retry limit = %v, want \"3\"", tpl.RetryStrategy.Limit)
	}
}

func TestImagePullPolicyParsing(t *testing.T) {
	tests := []struct {
		input string
		want  ImagePullPolicy
		err   bool
	}{
		{"Always", ImagePullAlways, false},
		{"always", ImagePullAlways, false},
		{"Never", ImagePullNever, false},
		{"never", ImagePullNever, false},
		{"IfNotPresent", ImagePullIfNotPresent, false},
		{"ifNotPresent", ImagePullIfNotPresent, false},
		{"if_not_present", ImagePullIfNotPresent, false},
		{"invalid", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := model.ParseImagePullPolicy(tt.input)
			if tt.err && err == nil {
				t.Fatal("expected error")
			}
			if !tt.err && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContainerWithImagePullPolicy(t *testing.T) {
	tests := []struct {
		name   string
		policy ImagePullPolicy
		want   string
	}{
		{"always", ImagePullAlways, "Always"},
		{"never", ImagePullNever, "Never"},
		{"if-not-present", ImagePullIfNotPresent, "IfNotPresent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Container{
				Name:            "test",
				Image:           "alpine",
				ImagePullPolicy: tt.policy,
			}
			tpl, err := c.BuildTemplate()
			if err != nil {
				t.Fatal(err)
			}
			if tpl.Container.ImagePullPolicy != tt.want {
				t.Errorf("policy = %q, want %q", tpl.Container.ImagePullPolicy, tt.want)
			}
		})
	}
}

func TestContainerTemplateJSON(t *testing.T) {
	c := &Container{
		Name:    "json-test",
		Image:   "alpine",
		Command: []string{"echo"},
		Args:    []string{"hello"},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(tpl)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["name"] != "json-test" {
		t.Errorf("json name = %v", m["name"])
	}
	container, ok := m["container"].(map[string]interface{})
	if !ok {
		t.Fatal("expected container in json")
	}
	if container["image"] != "alpine" {
		t.Errorf("json image = %v", container["image"])
	}
}

func TestContainerWithVolumeMounts(t *testing.T) {
	c := &Container{
		Name:  "with-mounts",
		Image: "alpine",
		VolumeMounts: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "scratch", MountPath: "/tmp/scratch"}},
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tpl.Container.VolumeMounts) != 1 {
		t.Fatalf("volumeMounts count = %d, want 1", len(tpl.Container.VolumeMounts))
	}
	if tpl.Container.VolumeMounts[0].MountPath != "/tmp/scratch" {
		t.Errorf("mountPath = %q", tpl.Container.VolumeMounts[0].MountPath)
	}
}

func TestContainerWithMetadata(t *testing.T) {
	c := &Container{
		Name:   "with-meta",
		Image:  "alpine",
		Labels: map[string]string{"app": "test"},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Metadata == nil {
		t.Fatal("expected metadata")
	}
	if tpl.Metadata.Labels["app"] != "test" {
		t.Errorf("label = %q", tpl.Metadata.Labels["app"])
	}
}

// --- Script tests (consolidated from script_test.go) ---

func TestScriptBuildTemplate(t *testing.T) {
	s := &Script{
		Name:    "hello-script",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  `print("hello world")`,
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "hello-script" {
		t.Errorf("name = %q", tpl.Name)
	}
	if tpl.Script == nil {
		t.Fatal("expected script to be set")
	}
	if tpl.Script.Image != "python:3.11" {
		t.Errorf("image = %q", tpl.Script.Image)
	}
	if tpl.Script.Source != `print("hello world")` {
		t.Errorf("source = %q", tpl.Script.Source)
	}
}

func TestScriptNoNameFails(t *testing.T) {
	s := &Script{Image: "python:3.11", Source: "print()"}
	_, err := s.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestScriptNoSourceFails(t *testing.T) {
	s := &Script{Name: "test", Image: "python:3.11"}
	_, err := s.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestScriptWithBashCommand(t *testing.T) {
	s := &Script{
		Name:    "bash-script",
		Image:   "alpine",
		Command: []string{"sh", "-c"},
		Source:  `echo "hello"`,
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tpl.Script.Command) != 2 || tpl.Script.Command[0] != "sh" {
		t.Errorf("command = %v", tpl.Script.Command)
	}
}

func TestScriptWithEnv(t *testing.T) {
	s := &Script{
		Name:    "with-env",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "import os; print(os.environ['FOO'])",
		Env: []EnvBuilder{
			Env{Name: "FOO", Value: "bar"},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tpl.Script.Env) != 1 {
		t.Fatalf("env count = %d", len(tpl.Script.Env))
	}
	if tpl.Script.Env[0].Name != "FOO" {
		t.Errorf("env name = %q", tpl.Script.Env[0].Name)
	}
}

func TestScriptWithInputs(t *testing.T) {
	s := &Script{
		Name:    "with-inputs",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "print('{{inputs.parameters.msg}}')",
		Inputs: []Parameter{
			{Name: "msg", Default: ptrStr("hello")},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Parameters) != 1 {
		t.Fatal("expected 1 input parameter")
	}
}

func TestScriptWithOutputs(t *testing.T) {
	s := &Script{
		Name:    "with-outputs",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "open('/tmp/result', 'w').write('42')",
		Outputs: []Parameter{
			{Name: "result", ValueFrom: &ValueFrom{Path: "/tmp/result"}},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Parameters) != 1 {
		t.Fatal("expected 1 output parameter")
	}
}

func TestScriptWithResources(t *testing.T) {
	s := &Script{
		Name:    "with-resources",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "print('hello')",
		Resources: &ResourceRequirements{
			Requests: ResourceList{CPU: "100m", Memory: "128Mi"},
			Limits:   ResourceList{CPU: "500m", Memory: "512Mi"},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Script.Resources == nil {
		t.Fatal("expected resources")
	}
	if tpl.Script.Resources.Requests.CPU != "100m" {
		t.Errorf("cpu = %q", tpl.Script.Resources.Requests.CPU)
	}
}

func TestScriptWithRetry(t *testing.T) {
	limit := 5
	s := &Script{
		Name:    "with-retry",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "raise Exception('fail')",
		RetryStrategy: &RetryStrategy{
			Limit:       &limit,
			RetryPolicy: RetryAlways,
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.RetryStrategy == nil {
		t.Fatal("expected retry strategy")
	}
	if tpl.RetryStrategy.Limit != "5" {
		t.Errorf("limit = %v, want \"5\"", tpl.RetryStrategy.Limit)
	}
}

func TestScriptMultilineSource(t *testing.T) {
	source := `import json
import sys

data = json.loads(sys.argv[1])
print(f"Name: {data['name']}")
print(f"Value: {data['value']}")
`
	s := &Script{
		Name:    "multiline",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  source,
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Script.Source != source {
		t.Errorf("source mismatch")
	}
}
