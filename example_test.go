package forge

import (
	"strings"
	"testing"

	"github.com/usetheo/theo/forge/expr"
)

// TestExampleDiamondDAG builds a complete diamond DAG workflow and validates YAML.
func TestExampleDiamondDAG(t *testing.T) {
	echoTpl := &Container{
		Name:    "echo",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{expr.InputParam("msg")},
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
	dag.AddTasks(A, B, C, D)

	w := &Workflow{
		GenerateName: "diamond-",
		Namespace:    "argo",
		Entrypoint:   "diamond",
		Templates:    []Templatable{echoTpl, dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"apiVersion: argoproj.io/v1alpha1",
		"kind: Workflow",
		"generateName: diamond-",
		"namespace: argo",
		"entrypoint: diamond",
		"name: echo",
		"image: alpine:3.18",
		"name: diamond",
		"name: A",
		"name: B",
		"name: C",
		"name: D",
		"depends: A",
	}
	for _, s := range expected {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q", s)
		}
	}
}

// TestExampleCoinflip builds a coinflip workflow with conditionals.
func TestExampleCoinflip(t *testing.T) {
	flip := &Script{
		Name:    "flip-coin",
		Image:   "python:3.11-alpine",
		Command: []string{"python"},
		Source: `import random
result = "heads" if random.randint(0, 1) == 0 else "tails"
print(result)`,
	}

	heads := &Container{
		Name:    "heads",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{"it was heads"},
	}

	tails := &Container{
		Name:    "tails",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{"it was tails"},
	}

	steps := &Steps{Name: "coinflip"}
	steps.AddSequentialStep(&Step{Name: "flip", Template: "flip-coin"})
	steps.AddParallelGroup(
		&Step{Name: "heads", Template: "heads", When: "{{steps.flip.outputs.result}} == heads"},
		&Step{Name: "tails", Template: "tails", When: "{{steps.flip.outputs.result}} == tails"},
	)

	w := &Workflow{
		GenerateName: "coinflip-",
		Entrypoint:   "coinflip",
		Templates:    []Templatable{flip, heads, tails, steps},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{"flip-coin", "heads", "tails", "coinflip", "when:"} {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q", s)
		}
	}
}

// TestExampleParameterPassing builds a workflow that passes outputs between steps.
func TestExampleParameterPassing(t *testing.T) {
	generate := &Script{
		Name:    "generate",
		Image:   "alpine:3.18",
		Command: []string{"sh", "-c"},
		Source:  `echo "42" > /tmp/result`,
		Outputs: []Parameter{{Name: "result", ValueFrom: &ValueFrom{Path: "/tmp/result"}}},
	}

	consume := &Container{
		Name:    "consume",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{expr.InputParam("msg")},
		Inputs:  []Parameter{{Name: "msg"}},
	}

	dag := &DAG{Name: "main"}
	genTask := &Task{Name: "generate", Template: "generate"}
	consumeTask := &Task{
		Name:     "consume",
		Template: "consume",
		Arguments: []Parameter{
			{Name: "msg", Value: ptrStr(expr.TaskOutputParam("generate", "result"))},
		},
	}
	genTask.Then(consumeTask)
	dag.AddTasks(genTask, consumeTask)

	w := &Workflow{
		GenerateName: "param-passing-",
		Entrypoint:   "main",
		Templates:    []Templatable{generate, consume, dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "{{tasks.generate.outputs.parameters.result}}") {
		t.Error("YAML missing task output reference")
	}
}

// TestExampleArtifactPassing builds a workflow with artifact passing.
func TestExampleArtifactPassing(t *testing.T) {
	generate := &Script{
		Name:    "generate",
		Image:   "alpine:3.18",
		Command: []string{"sh", "-c"},
		Source:  `echo "hello world" > /tmp/output.txt`,
		OutputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "output-file", Path: "/tmp/output.txt"},
		},
	}

	consume := &Container{
		Name:    "consume",
		Image:   "alpine:3.18",
		Command: []string{"cat"},
		Args:    []string{"/tmp/input.txt"},
	}

	w := &Workflow{
		GenerateName: "artifacts-",
		Entrypoint:   "main",
		Templates: []Templatable{
			generate,
			consume,
			&DAG{
				Name: "main",
				Tasks: []*Task{
					{Name: "gen", Template: "generate"},
					{Name: "use", Template: "consume", Depends: "gen"},
				},
			},
		},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "output-file") {
		t.Error("YAML missing artifact name")
	}
}

// TestExampleCronWorkflow builds a scheduled workflow.
func TestExampleCronWorkflow(t *testing.T) {
	cw := &CronWorkflow{
		Name:              "hourly-cleanup",
		Namespace:         "ops",
		Schedule:          "0 * * * *",
		Timezone:          "UTC",
		ConcurrencyPolicy: "Forbid",
		Entrypoint:        "cleanup",
		Templates: []Templatable{
			&Container{
				Name:    "cleanup",
				Image:   "alpine:3.18",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo 'cleaning up...'"},
			},
		},
	}

	y, err := cw.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{
		"kind: CronWorkflow",
		"schedule: 0 * * * *",
		"timezone: UTC",
		"concurrencyPolicy: Forbid",
		"name: hourly-cleanup",
		"namespace: ops",
	} {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q", s)
		}
	}
}

// TestExampleWorkflowTemplateRef builds a workflow that references a WorkflowTemplate.
func TestExampleWorkflowTemplateRef(t *testing.T) {
	// First, define the reusable template
	wt := &WorkflowTemplate{
		Name:       "echo-template",
		Namespace:  "default",
		Entrypoint: "echo",
		Templates: []Templatable{
			&Container{
				Name:    "echo",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{expr.InputParam("msg")},
				Inputs:  []Parameter{{Name: "msg"}},
			},
		},
	}

	wtYAML, err := wt.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(wtYAML, "kind: WorkflowTemplate") {
		t.Error("missing kind in WorkflowTemplate YAML")
	}

	// Then, use it in a workflow via templateRef
	dag := &DAG{Name: "main"}
	dag.AddTask(&Task{
		Name: "call-echo",
		TemplateRef: &TemplateRef{
			Name:     "echo-template",
			Template: "echo",
		},
		Arguments: []Parameter{{Name: "msg", Value: ptrStr("Hello from ref!")}},
	})

	w := &Workflow{
		Name:       "use-template-ref",
		Entrypoint: "main",
		Templates:  []Templatable{dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "templateRef:") {
		t.Error("YAML missing templateRef")
	}
	if !strings.Contains(y, "name: echo-template") {
		t.Error("YAML missing template ref name")
	}
}

// TestExampleWithVolumesAndSecrets builds a complete workflow with volumes.
func TestExampleWithVolumesAndSecrets(t *testing.T) {
	w := &Workflow{
		Name:       "with-volumes",
		Entrypoint: "main",
		Volumes: []VolumeBuilder{
			&SecretVolume{BaseVolume: BaseVolume{Name: "creds", MountPath: "/etc/creds"}, SecretName: "app-creds"},
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "workspace", MountPath: "/workspace"}},
		},
		Templates: []Templatable{
			&Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"sh", "-c"},
				Args:    []string{"cat /etc/creds/password && ls /workspace"},
				Env: []EnvBuilder{
					SecretEnv{Name: "DB_PASS", SecretName: "db-creds", SecretKey: "password"},
				},
				VolumeMounts: []VolumeBuilder{
					&SecretVolume{BaseVolume: BaseVolume{Name: "creds", MountPath: "/etc/creds"}},
					&EmptyDirVolume{BaseVolume: BaseVolume{Name: "workspace", MountPath: "/workspace"}},
				},
			},
		},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{"secretName: app-creds", "emptyDir:", "mountPath: /etc/creds", "mountPath: /workspace"} {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q", s)
		}
	}
}
