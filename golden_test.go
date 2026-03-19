package forge

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "update golden test files")

func goldenTest(t *testing.T, name string, got string) {
	t.Helper()

	goldenPath := filepath.Join("testdata", name+".yaml")

	if *updateGolden {
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden file: %v", err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file %s: %v (run with -update-golden to create)", goldenPath, err)
	}

	if got != string(expected) {
		t.Errorf("YAML output does not match golden file %s\n\nGot:\n%s\n\nExpected:\n%s", goldenPath, got, string(expected))
	}
}

func TestGoldenSimpleContainer(t *testing.T) {
	w := &Workflow{
		GenerateName: "hello-",
		Entrypoint:   "main",
		Templates: []Templatable{
			&Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"hello world"},
			},
		},
	}
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "simple_container", yaml)
}

func TestGoldenDiamondDAG(t *testing.T) {
	echoTpl := &Container{
		Name:    "echo",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{InputParam("msg")},
		Inputs:  []Parameter{{Name: "msg"}},
	}

	dag := &DAG{Name: "diamond"}
	A := &Task{Name: "A", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("Task A")}}}
	B := &Task{Name: "B", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("Task B")}}}
	C := &Task{Name: "C", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("Task C")}}}
	D := &Task{Name: "D", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("Task D")}}}

	A.Then(B)
	A.Then(C)
	B.Then(D)
	C.Then(D)
	_ = dag.AddTasks(A, B, C, D)

	w := &Workflow{
		GenerateName: "diamond-",
		Namespace:    "argo",
		Entrypoint:   "diamond",
		Templates:    []Templatable{echoTpl, dag},
	}

	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "diamond_dag", yaml)
}

func TestGoldenScriptWorkflow(t *testing.T) {
	w := &Workflow{
		GenerateName: "script-",
		Entrypoint:   "main",
		Templates: []Templatable{
			&Script{
				Name:    "main",
				Image:   "python:3.11-alpine",
				Command: []string{"python"},
				Source:  "print('hello from script')",
			},
		},
	}
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "script_workflow", yaml)
}

func TestGoldenStepsWorkflow(t *testing.T) {
	steps := &Steps{Name: "pipeline"}
	_ = steps.AddSequentialStep(&Step{Name: "build", Template: "build-tpl"})
	_ = steps.AddSequentialStep(&Step{Name: "test", Template: "test-tpl"})
	_ = steps.AddParallelGroup(
		&Step{Name: "deploy-a", Template: "deploy-tpl"},
		&Step{Name: "deploy-b", Template: "deploy-tpl"},
	)

	w := &Workflow{
		Name:       "steps-example",
		Entrypoint: "pipeline",
		Templates: []Templatable{
			steps,
			&Container{Name: "build-tpl", Image: "golang:1.22"},
			&Container{Name: "test-tpl", Image: "golang:1.22"},
			&Container{Name: "deploy-tpl", Image: "alpine:3.18"},
		},
	}
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "steps_workflow", yaml)
}

func TestGoldenWorkflowTemplate(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "shared-build",
		Namespace:  "default",
		Entrypoint: "build",
		Templates: []Templatable{
			&Container{
				Name:    "build",
				Image:   "golang:1.22",
				Command: []string{"go", "build", "./..."},
			},
		},
	}
	yaml, err := wt.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "workflow_template", yaml)
}

func TestGoldenCronWorkflow(t *testing.T) {
	cw := &CronWorkflow{
		Name:      "nightly-build",
		Namespace: "ci",
		Schedule:  "0 2 * * *",
		Timezone:  "America/Sao_Paulo",
		Entrypoint: "build",
		Templates: []Templatable{
			&Container{
				Name:    "build",
				Image:   "golang:1.22",
				Command: []string{"make", "build"},
			},
		},
	}
	yaml, err := cw.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "cron_workflow", yaml)
}
