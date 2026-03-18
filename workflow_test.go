package forge

import (
	"strings"
	"testing"
)

func TestWorkflowNameValidation(t *testing.T) {
	tests := []struct {
		name string
		wf   Workflow
		err  bool
	}{
		{"valid name", Workflow{Name: "my-workflow", Entrypoint: "main"}, false},
		{"valid generate name", Workflow{GenerateName: "my-wf-", Entrypoint: "main"}, false},
		{"no name", Workflow{Entrypoint: "main"}, true},
		{"name too long", Workflow{Name: strings.Repeat("a", NameLimit+1), Entrypoint: "main"}, true},
		{"generate name too long", Workflow{GenerateName: strings.Repeat("a", NameLimit+1), Entrypoint: "main"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.wf.Build()
			if tt.err && err == nil {
				t.Fatal("expected error")
			}
			if !tt.err && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWorkflowBuild(t *testing.T) {
	w := &Workflow{
		Name:       "test-workflow",
		Namespace:  "default",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{
				Name:    "main",
				Image:   "alpine",
				Command: []string{"echo"},
				Args:    []string{"hello"},
			},
		},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.APIVersion != DefaultAPIVersion {
		t.Errorf("apiVersion = %q", model.APIVersion)
	}
	if model.Kind != DefaultKind {
		t.Errorf("kind = %q", model.Kind)
	}
	if model.Metadata.Name != "test-workflow" {
		t.Errorf("name = %q", model.Metadata.Name)
	}
	if model.Metadata.Namespace != "default" {
		t.Errorf("namespace = %q", model.Metadata.Namespace)
	}
	if model.Spec.Entrypoint != "main" {
		t.Errorf("entrypoint = %q", model.Spec.Entrypoint)
	}
	if len(model.Spec.Templates) != 1 {
		t.Fatalf("templates = %d, want 1", len(model.Spec.Templates))
	}
}

func TestWorkflowWithDAG(t *testing.T) {
	dag := &DAG{Name: "main"}
	a := &Task{Name: "A", Template: "echo"}
	b := &Task{Name: "B", Template: "echo"}
	a.Then(b)
	dag.AddTasks(a, b)

	w := &Workflow{
		Name:       "dag-workflow",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "echo", Image: "alpine", Command: []string{"echo"}, Args: []string{"hello"}},
			dag,
		},
	}

	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Spec.Templates) != 2 {
		t.Fatalf("templates = %d, want 2", len(model.Spec.Templates))
	}
	// First should be container, second should be DAG
	if model.Spec.Templates[0].Container == nil {
		t.Error("expected template[0] to be container")
	}
	if model.Spec.Templates[1].DAG == nil {
		t.Error("expected template[1] to be DAG")
	}
}

func TestWorkflowWithArguments(t *testing.T) {
	w := &Workflow{
		Name:       "args-workflow",
		Entrypoint: "main",
		Arguments: []Parameter{
			{Name: "msg", Value: ptrStr("hello")},
			{Name: "count", Value: ptrStr("3")},
		},
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine"},
		},
	}

	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Arguments == nil {
		t.Fatal("expected arguments")
	}
	if len(model.Spec.Arguments.Parameters) != 2 {
		t.Fatalf("args = %d, want 2", len(model.Spec.Arguments.Parameters))
	}
}

func TestWorkflowReassignArguments(t *testing.T) {
	w := &Workflow{
		Name:       "reassign-args",
		Entrypoint: "main",
		Arguments:  []Parameter{{Name: "msg", Value: ptrStr("hello")}},
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	// Reassign
	w.Arguments = []Parameter{{Name: "msg", Value: ptrStr("world")}}

	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Arguments == nil || len(model.Spec.Arguments.Parameters) != 1 {
		t.Fatal("expected 1 argument after reassign")
	}
	if *model.Spec.Arguments.Parameters[0].Value != "world" {
		t.Errorf("value = %q, want 'world'", *model.Spec.Arguments.Parameters[0].Value)
	}
}

func TestWorkflowGetParameter(t *testing.T) {
	w := &Workflow{
		Name:       "get-param",
		Entrypoint: "main",
		Arguments: []Parameter{
			{Name: "msg", Value: ptrStr("hello")},
			{Name: "count", Value: ptrStr("3")},
		},
	}

	p, err := w.GetParameter("msg")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "msg" {
		t.Errorf("name = %q", p.Name)
	}
	if *p.Value != "hello" {
		t.Errorf("value = %q", *p.Value)
	}

	_, err = w.GetParameter("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing parameter")
	}
}

func TestWorkflowToYAML(t *testing.T) {
	w := &Workflow{
		Name:       "yaml-test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine", Command: []string{"echo"}, Args: []string{"hello"}},
		},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(yamlStr, "apiVersion: argoproj.io/v1alpha1") {
		t.Error("missing apiVersion in YAML")
	}
	if !strings.Contains(yamlStr, "kind: Workflow") {
		t.Error("missing kind in YAML")
	}
	if !strings.Contains(yamlStr, "name: yaml-test") {
		t.Error("missing name in YAML")
	}
	if !strings.Contains(yamlStr, "entrypoint: main") {
		t.Error("missing entrypoint in YAML")
	}
	if !strings.Contains(yamlStr, "image: alpine") {
		t.Error("missing image in YAML")
	}
}

func TestWorkflowToJSON(t *testing.T) {
	w := &Workflow{
		Name:       "json-test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine"},
		},
	}

	jsonStr, err := w.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(jsonStr, `"apiVersion": "argoproj.io/v1alpha1"`) {
		t.Error("missing apiVersion in JSON")
	}
	if !strings.Contains(jsonStr, `"kind": "Workflow"`) {
		t.Error("missing kind in JSON")
	}
}

func TestWorkflowYAMLRoundTrip(t *testing.T) {
	w := &Workflow{
		Name:       "round-trip",
		Namespace:  "argo",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine", Command: []string{"echo"}, Args: []string{"hello"}},
		},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	model, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatal(err)
	}
	if model.Metadata.Name != "round-trip" {
		t.Errorf("name = %q after round-trip", model.Metadata.Name)
	}
	if model.Metadata.Namespace != "argo" {
		t.Errorf("namespace = %q after round-trip", model.Metadata.Namespace)
	}
	if model.Spec.Entrypoint != "main" {
		t.Errorf("entrypoint = %q after round-trip", model.Spec.Entrypoint)
	}
}

func TestWorkflowToDict(t *testing.T) {
	w := &Workflow{
		Name:       "dict-test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine"},
		},
	}

	d, err := w.ToDict()
	if err != nil {
		t.Fatal(err)
	}
	if d["apiVersion"] != DefaultAPIVersion {
		t.Errorf("apiVersion = %v", d["apiVersion"])
	}
	meta, ok := d["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("expected metadata to be a map")
	}
	if meta["name"] != "dict-test" {
		t.Errorf("name = %v", meta["name"])
	}
}

func TestWorkflowWithLabels(t *testing.T) {
	w := &Workflow{
		Name:       "labeled",
		Entrypoint: "main",
		Labels:     map[string]string{"app": "test", "team": "backend"},
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Metadata.Labels["app"] != "test" {
		t.Errorf("label app = %q", model.Metadata.Labels["app"])
	}
}

func TestWorkflowWithVolumes(t *testing.T) {
	w := &Workflow{
		Name:       "with-volumes",
		Entrypoint: "main",
		Volumes: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "scratch", MountPath: "/tmp"}},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Spec.Volumes) != 1 {
		t.Fatalf("volumes = %d, want 1", len(model.Spec.Volumes))
	}
	if model.Spec.Volumes[0].EmptyDir == nil {
		t.Error("expected emptyDir volume")
	}
}

func TestWorkflowWithTTLAndPodGC(t *testing.T) {
	secs := 3600
	w := &Workflow{
		Name:       "with-gc",
		Entrypoint: "main",
		TTLStrategy: &TTLStrategy{SecondsAfterCompletion: &secs},
		PodGC:      &PodGC{Strategy: "OnPodCompletion"},
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.TTLStrategy == nil {
		t.Fatal("expected TTL strategy")
	}
	if *model.Spec.TTLStrategy.SecondsAfterCompletion != 3600 {
		t.Errorf("ttl = %d", *model.Spec.TTLStrategy.SecondsAfterCompletion)
	}
	if model.Spec.PodGC == nil || model.Spec.PodGC.Strategy != "OnPodCompletion" {
		t.Error("expected pod GC strategy")
	}
}

func TestWorkflowWithImagePullSecrets(t *testing.T) {
	w := &Workflow{
		Name:             "with-secrets",
		Entrypoint:       "main",
		ImagePullSecrets: []string{"my-registry-key"},
		Templates:        []Templatable{&Container{Name: "main", Image: "private.registry/app:v1"}},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Spec.ImagePullSecrets) != 1 {
		t.Fatalf("secrets = %d", len(model.Spec.ImagePullSecrets))
	}
	if model.Spec.ImagePullSecrets[0].Name != "my-registry-key" {
		t.Errorf("secret name = %q", model.Spec.ImagePullSecrets[0].Name)
	}
}

func TestWorkflowWithSteps(t *testing.T) {
	steps := &Steps{Name: "main"}
	steps.AddSequentialStep(&Step{Name: "echo", Template: "echo-tpl"})

	w := &Workflow{
		Name:       "steps-workflow",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "echo-tpl", Image: "alpine", Command: []string{"echo"}, Args: []string{"hello"}},
			steps,
		},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Spec.Templates) != 2 {
		t.Fatalf("templates = %d", len(model.Spec.Templates))
	}
	if model.Spec.Templates[1].Steps == nil {
		t.Error("expected steps template")
	}
}

func TestWorkflowComplexDiamond(t *testing.T) {
	// Build a complete diamond DAG workflow
	echoTpl := &Container{
		Name:    "echo",
		Image:   "alpine",
		Command: []string{"echo"},
		Args:    []string{"{{inputs.parameters.msg}}"},
		Inputs:  []Parameter{{Name: "msg"}},
	}

	dag := &DAG{Name: "diamond"}
	a := &Task{Name: "A", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("A")}}}
	b := &Task{Name: "B", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("B")}}}
	c := &Task{Name: "C", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("C")}}}
	d := &Task{Name: "D", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("D")}}}

	a.Then(b)
	a.Then(c)
	b.Then(d)
	c.Then(d)
	dag.AddTasks(a, b, c, d)

	w := &Workflow{
		Name:       "diamond-workflow",
		Namespace:  "argo",
		Entrypoint: "diamond",
		Templates:  []Templatable{echoTpl, dag},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	// Verify the YAML contains expected structure
	for _, expected := range []string{
		"apiVersion: argoproj.io/v1alpha1",
		"kind: Workflow",
		"name: diamond-workflow",
		"namespace: argo",
		"entrypoint: diamond",
		"image: alpine",
	} {
		if !strings.Contains(yamlStr, expected) {
			t.Errorf("YAML missing: %q", expected)
		}
	}
}

func TestWorkflowGenerateName(t *testing.T) {
	w := &Workflow{
		GenerateName: "my-wf-",
		Entrypoint:   "main",
		Templates:    []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Metadata.GenerateName != "my-wf-" {
		t.Errorf("generateName = %q", model.Metadata.GenerateName)
	}
	if model.Metadata.Name != "" {
		t.Errorf("name should be empty, got %q", model.Metadata.Name)
	}
}
