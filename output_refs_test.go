package forge

import "testing"

func TestTaskGetOutputParameter(t *testing.T) {
	task := &Task{Name: "generate", Template: "gen"}
	got := task.GetOutputParameter("result")
	want := "{{tasks.generate.outputs.parameters.result}}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTaskGetOutputResult(t *testing.T) {
	task := &Task{Name: "check", Template: "check"}
	got := task.GetOutputResult()
	want := "{{tasks.check.outputs.result}}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTaskGetOutputArtifact(t *testing.T) {
	task := &Task{Name: "build", Template: "build"}
	got := task.GetOutputArtifact("binary")
	want := "{{tasks.build.outputs.artifacts.binary}}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStepGetOutputParameter(t *testing.T) {
	step := &Step{Name: "generate", Template: "gen"}
	got := step.GetOutputParameter("result")
	want := "{{steps.generate.outputs.parameters.result}}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStepGetOutputResult(t *testing.T) {
	step := &Step{Name: "check", Template: "check"}
	got := step.GetOutputResult()
	want := "{{steps.check.outputs.result}}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStepGetOutputArtifact(t *testing.T) {
	step := &Step{Name: "build", Template: "build"}
	got := step.GetOutputArtifact("logs")
	want := "{{steps.build.outputs.artifacts.logs}}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Test using output refs in a workflow
func TestOutputRefsInWorkflow(t *testing.T) {
	dag := &DAG{Name: "pipeline"}
	gen := &Task{Name: "generate", Template: "gen-script"}
	consume := &Task{
		Name:     "consume",
		Template: "echo",
		Arguments: []Parameter{
			{Name: "msg", Value: ptrStr(gen.GetOutputParameter("data"))},
		},
	}
	gen.Then(consume)
	dag.AddTasks(gen, consume)

	w := &Workflow{
		Name:       "output-refs",
		Entrypoint: "pipeline",
		Templates: []Templatable{
			&Container{Name: "echo", Image: "alpine", Inputs: []Parameter{{Name: "msg"}}},
			&Script{Name: "gen-script", Image: "python:3.11", Command: []string{"python"}, Source: "print('42')"},
			dag,
		},
	}

	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}

	// Find the consume task in DAG and verify its argument
	dagTpl := model.Spec.Templates[2]
	if dagTpl.DAG == nil {
		t.Fatal("expected DAG template")
	}

	for _, task := range dagTpl.DAG.Tasks {
		if task.Name == "consume" {
			if task.Arguments == nil || len(task.Arguments.Parameters) != 1 {
				t.Fatal("expected 1 argument on consume task")
			}
			got := *task.Arguments.Parameters[0].Value
			want := "{{tasks.generate.outputs.parameters.data}}"
			if got != want {
				t.Errorf("arg value = %q, want %q", got, want)
			}
			return
		}
	}
	t.Fatal("consume task not found")
}
