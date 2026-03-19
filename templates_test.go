package forge

import (
	"testing"
)

// --- ContainerSet tests (consolidated from container_set_test.go) ---

func TestContainerSetBuild(t *testing.T) {
	cs := &ContainerSet{
		Name: "setup-and-run",
		Containers: []ContainerNode{
			{Name: "setup", Image: "alpine", Command: []string{"sh", "-c"}, Args: []string{"echo setup"}},
			{Name: "run", Image: "python:3.11", Command: []string{"python"}, Args: []string{"-c", "print('run')"}},
		},
	}
	tpl, err := cs.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "setup-and-run" {
		t.Errorf("name = %q", tpl.Name)
	}
}

func TestContainerSetNoNameFails(t *testing.T) {
	cs := &ContainerSet{
		Containers: []ContainerNode{{Name: "c", Image: "alpine"}},
	}
	_, err := cs.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestContainerSetNoContainersFails(t *testing.T) {
	cs := &ContainerSet{Name: "empty"}
	_, err := cs.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for no containers")
	}
}

func TestContainerSetWithInputs(t *testing.T) {
	cs := &ContainerSet{
		Name: "with-inputs",
		Containers: []ContainerNode{
			{Name: "main", Image: "alpine"},
		},
		Inputs: []Parameter{{Name: "msg", Value: ptrStr("hello")}},
	}
	tpl, err := cs.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Parameters) != 1 {
		t.Fatal("expected 1 input")
	}
}

func TestContainerSetWithEnv(t *testing.T) {
	cs := &ContainerSet{
		Name: "with-env",
		Containers: []ContainerNode{
			{
				Name:  "main",
				Image: "alpine",
				Env:   []EnvBuilder{Env{Name: "FOO", Value: "bar"}},
			},
		},
	}
	_, err := cs.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
}

func TestContainerSetInWorkflow(t *testing.T) {
	w := &Workflow{
		Name:       "cs-workflow",
		Entrypoint: "main",
		Templates: []Templatable{
			&ContainerSet{
				Name: "main",
				Containers: []ContainerNode{
					{Name: "a", Image: "alpine", Command: []string{"echo"}, Args: []string{"hello"}},
					{Name: "b", Image: "alpine", Command: []string{"echo"}, Args: []string{"world"}, Dependencies: []string{"a"}},
				},
			},
		},
	}
	_, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
}

// --- BuildArguments helpers ---

func TestBuildArguments(t *testing.T) {
	args, err := BuildArguments(
		[]Parameter{{Name: "msg", Value: ptrStr("hello")}},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if args == nil || len(args.Parameters) != 1 {
		t.Fatal("expected 1 parameter")
	}
	if args.Parameters[0].Name != "msg" {
		t.Errorf("name = %q", args.Parameters[0].Name)
	}
}

func TestBuildArgumentsEmpty(t *testing.T) {
	args, err := BuildArguments(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if args != nil {
		t.Error("expected nil for empty arguments")
	}
}

func TestBuildArgumentsMixed(t *testing.T) {
	args, err := BuildArguments(
		[]Parameter{{Name: "msg", Value: ptrStr("hello")}},
		[]ArtifactBuilder{&Artifact{Name: "data", Path: "/tmp/data"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(args.Parameters) != 1 {
		t.Errorf("params = %d", len(args.Parameters))
	}
	if len(args.Artifacts) != 1 {
		t.Errorf("artifacts = %d", len(args.Artifacts))
	}
}

func TestBuildArgumentsFromMap(t *testing.T) {
	args := BuildArgumentsFromMap(map[string]string{
		"msg":   "hello",
		"count": "3",
	})
	if args == nil {
		t.Fatal("expected non-nil arguments")
	}
	if len(args.Parameters) != 2 {
		t.Fatalf("params = %d, want 2", len(args.Parameters))
	}
	// Verify all params have values
	for _, p := range args.Parameters {
		if p.Value == nil {
			t.Errorf("param %q has nil value", p.Name)
		}
	}
}

func TestBuildArgumentsFromMapEmpty(t *testing.T) {
	args := BuildArgumentsFromMap(nil)
	if args != nil {
		t.Error("expected nil for empty map")
	}
}

func TestBuildArgumentsFromMapDictKeyBecomesName(t *testing.T) {
	args := BuildArgumentsFromMap(map[string]string{"a-key": "a-value"})
	if args.Parameters[0].Name != "a-key" {
		t.Errorf("name = %q, want 'a-key'", args.Parameters[0].Name)
	}
	if *args.Parameters[0].Value != "a-value" {
		t.Errorf("value = %q, want 'a-value'", *args.Parameters[0].Value)
	}
}

// --- Resource Template tests (consolidated from resource_template_test.go) ---

func TestResourceTemplateBuild(t *testing.T) {
	r := &ResourceTemplate{
		Name:   "create-configmap",
		Action: "create",
		Manifest: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key: value`,
		SuccessCondition: "status.phase == Active",
	}
	tpl, err := r.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "create-configmap" {
		t.Errorf("name = %q", tpl.Name)
	}
	if tpl.Resource == nil {
		t.Fatal("expected resource to be set")
	}
	if tpl.Resource.Action != "create" {
		t.Errorf("action = %q", tpl.Resource.Action)
	}
	if tpl.Resource.SuccessCondition != "status.phase == Active" {
		t.Errorf("successCondition = %q", tpl.Resource.SuccessCondition)
	}
}

func TestResourceTemplateNoNameFails(t *testing.T) {
	r := &ResourceTemplate{Action: "create", Manifest: "test"}
	_, err := r.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestResourceTemplateNoActionFails(t *testing.T) {
	r := &ResourceTemplate{Name: "test", Manifest: "test"}
	_, err := r.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty action")
	}
}

func TestResourceTemplateNoManifestFails(t *testing.T) {
	r := &ResourceTemplate{Name: "test", Action: "create"}
	_, err := r.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty manifest")
	}
}

func TestResourceTemplateWithInputsOutputs(t *testing.T) {
	r := &ResourceTemplate{
		Name:     "with-io",
		Action:   "create",
		Manifest: "apiVersion: v1\nkind: ConfigMap",
		Inputs:   []Parameter{{Name: "name", Value: ptrStr("test")}},
		Outputs:  []Parameter{{Name: "uid", ValueFrom: &ValueFrom{JSONPath: "{.metadata.uid}"}}},
	}
	tpl, err := r.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Parameters) != 1 {
		t.Fatal("expected 1 input")
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Parameters) != 1 {
		t.Fatal("expected 1 output")
	}
}

// --- Suspend Template ---

func TestSuspendTemplateBuild(t *testing.T) {
	s := &Suspend{Name: "wait", Duration: "30s"}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "wait" {
		t.Errorf("name = %q", tpl.Name)
	}
	if tpl.Suspend == nil {
		t.Fatal("expected suspend to be set")
	}
	if tpl.Suspend.Duration != "30s" {
		t.Errorf("duration = %q", tpl.Suspend.Duration)
	}
}

func TestSuspendTemplateNoDuration(t *testing.T) {
	s := &Suspend{Name: "manual-approval"}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Suspend == nil {
		t.Fatal("expected suspend to be set")
	}
	if tpl.Suspend.Duration != "" {
		t.Errorf("duration should be empty for manual approval, got %q", tpl.Suspend.Duration)
	}
}

func TestSuspendTemplateNoNameFails(t *testing.T) {
	s := &Suspend{Duration: "10s"}
	_, err := s.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

// --- HTTP Template ---

func TestHTTPTemplateBuild(t *testing.T) {
	h := &HTTPTemplate{
		Name:   "health-check",
		URL:    "https://api.example.com/health",
		Method: "GET",
		Headers: map[string]string{
			"Accept": "application/json",
		},
		SuccessCondition: "response.statusCode == 200",
	}
	tpl, err := h.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "health-check" {
		t.Errorf("name = %q", tpl.Name)
	}
	if tpl.HTTP == nil {
		t.Fatal("expected HTTP to be set")
	}
	if tpl.HTTP.URL != "https://api.example.com/health" {
		t.Errorf("url = %q", tpl.HTTP.URL)
	}
	if tpl.HTTP.Method != "GET" {
		t.Errorf("method = %q", tpl.HTTP.Method)
	}
	if tpl.HTTP.SuccessCondition != "response.statusCode == 200" {
		t.Errorf("successCondition = %q", tpl.HTTP.SuccessCondition)
	}
	if len(tpl.HTTP.Headers) != 1 {
		t.Fatalf("headers = %d", len(tpl.HTTP.Headers))
	}
}

func TestHTTPTemplateNoNameFails(t *testing.T) {
	h := &HTTPTemplate{URL: "https://example.com"}
	_, err := h.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestHTTPTemplateNoURLFails(t *testing.T) {
	h := &HTTPTemplate{Name: "test"}
	_, err := h.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestHTTPTemplateWithBody(t *testing.T) {
	h := &HTTPTemplate{
		Name:   "post-data",
		URL:    "https://api.example.com/data",
		Method: "POST",
		Body:   `{"key": "value"}`,
	}
	tpl, err := h.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.HTTP.Body != `{"key": "value"}` {
		t.Errorf("body = %q", tpl.HTTP.Body)
	}
}

func TestHTTPTemplateInWorkflow(t *testing.T) {
	w := &Workflow{
		Name:       "http-workflow",
		Entrypoint: "main",
		Templates: []Templatable{
			&HTTPTemplate{
				Name:   "main",
				URL:    "https://example.com/api",
				Method: "GET",
			},
		},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Spec.Templates) != 1 {
		t.Fatalf("templates = %d", len(model.Spec.Templates))
	}
	if model.Spec.Templates[0].HTTP == nil {
		t.Error("expected HTTP template")
	}
}

func TestSuspendInWorkflow(t *testing.T) {
	w := &Workflow{
		Name:       "suspend-workflow",
		Entrypoint: "main",
		Templates: []Templatable{
			&Suspend{Name: "main", Duration: "5m"},
		},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Templates[0].Suspend == nil {
		t.Error("expected suspend template")
	}
}

func TestResourceInWorkflow(t *testing.T) {
	w := &Workflow{
		Name:       "resource-workflow",
		Entrypoint: "main",
		Templates: []Templatable{
			&ResourceTemplate{
				Name:     "main",
				Action:   "create",
				Manifest: "apiVersion: v1\nkind: ConfigMap",
			},
		},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Templates[0].Resource == nil {
		t.Error("expected resource template")
	}
}
