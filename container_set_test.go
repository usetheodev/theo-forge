package forge

import (
	"testing"
)

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
