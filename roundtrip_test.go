package forge

import (
	"encoding/json"
	"testing"

	"github.com/usetheo/theo/forge/model"
	"sigs.k8s.io/yaml"
)

func TestRoundTripYAML(t *testing.T) {
	val := "world"
	w := &Workflow{
		Name:       "roundtrip-yaml",
		Namespace:  "default",
		Entrypoint: "main",
		Arguments: []Parameter{
			{Name: "greeting", Value: &val},
		},
		Templates: []Templatable{
			&Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"hello"},
			},
		},
		Labels:      map[string]string{"app": "test"},
		Annotations: map[string]string{"note": "roundtrip"},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}

	wf, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}

	if wf.APIVersion != DefaultAPIVersion {
		t.Errorf("apiVersion = %q, want %q", wf.APIVersion, DefaultAPIVersion)
	}
	if wf.Kind != DefaultKind {
		t.Errorf("kind = %q, want %q", wf.Kind, DefaultKind)
	}
	if wf.Metadata.Name != "roundtrip-yaml" {
		t.Errorf("name = %q", wf.Metadata.Name)
	}
	if wf.Metadata.Namespace != "default" {
		t.Errorf("namespace = %q", wf.Metadata.Namespace)
	}
	if wf.Metadata.Labels["app"] != "test" {
		t.Errorf("label app = %q", wf.Metadata.Labels["app"])
	}
	if wf.Metadata.Annotations["note"] != "roundtrip" {
		t.Errorf("annotation note = %q", wf.Metadata.Annotations["note"])
	}
	if wf.Spec.Entrypoint != "main" {
		t.Errorf("entrypoint = %q", wf.Spec.Entrypoint)
	}
	if len(wf.Spec.Templates) != 1 {
		t.Fatalf("templates = %d", len(wf.Spec.Templates))
	}
	if wf.Spec.Templates[0].Container == nil {
		t.Fatal("expected container template")
	}
	if wf.Spec.Templates[0].Container.Image != "alpine:3.18" {
		t.Errorf("image = %q", wf.Spec.Templates[0].Container.Image)
	}
	if wf.Spec.Arguments == nil || len(wf.Spec.Arguments.Parameters) != 1 {
		t.Fatalf("expected 1 argument parameter")
	}
	if wf.Spec.Arguments.Parameters[0].Name != "greeting" {
		t.Errorf("argument name = %q", wf.Spec.Arguments.Parameters[0].Name)
	}
	if wf.Spec.Arguments.Parameters[0].Value == nil || *wf.Spec.Arguments.Parameters[0].Value != "world" {
		t.Errorf("argument value mismatch")
	}
}

func TestRoundTripJSON(t *testing.T) {
	w := &Workflow{
		Name:       "roundtrip-json",
		Entrypoint: "run",
		Templates: []Templatable{
			&Script{
				Name:    "run",
				Image:   "python:3.11",
				Command: []string{"python"},
				Source:  "print('hello')",
			},
		},
	}

	jsonStr, err := w.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	wf, err := FromJSON(jsonStr)
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}

	if wf.Metadata.Name != "roundtrip-json" {
		t.Errorf("name = %q", wf.Metadata.Name)
	}
	if len(wf.Spec.Templates) != 1 {
		t.Fatalf("templates = %d", len(wf.Spec.Templates))
	}
	if wf.Spec.Templates[0].Script == nil {
		t.Fatal("expected script template")
	}
	if wf.Spec.Templates[0].Script.Source != "print('hello')" {
		t.Errorf("source = %q", wf.Spec.Templates[0].Script.Source)
	}
}

func TestRoundTripDAG(t *testing.T) {
	dag := &DAG{Name: "pipeline"}
	taskA := &Task{Name: "a", Template: "echo"}
	taskB := &Task{Name: "b", Template: "echo"}
	taskA.Then(taskB)

	if err := dag.AddTasks(taskA, taskB); err != nil {
		t.Fatal(err)
	}

	w := &Workflow{
		Name:       "dag-roundtrip",
		Entrypoint: "pipeline",
		Templates: []Templatable{
			dag,
			&Container{Name: "echo", Image: "alpine", Command: []string{"echo", "hi"}},
		},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}

	wf, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}

	if len(wf.Spec.Templates) != 2 {
		t.Fatalf("templates = %d", len(wf.Spec.Templates))
	}

	var dagModel *model.DAGModel
	for _, tpl := range wf.Spec.Templates {
		if tpl.Name == "pipeline" {
			dagModel = tpl.DAG
		}
	}
	if dagModel == nil {
		t.Fatal("expected DAG template")
	}
	if len(dagModel.Tasks) != 2 {
		t.Fatalf("dag tasks = %d", len(dagModel.Tasks))
	}

	taskMap := make(map[string]model.DAGTaskModel)
	for _, task := range dagModel.Tasks {
		taskMap[task.Name] = task
	}
	if taskMap["b"].Depends != "a" {
		t.Errorf("task b depends = %q, want %q", taskMap["b"].Depends, "a")
	}
}

func TestRoundTripSteps(t *testing.T) {
	steps := &Steps{Name: "sequential"}
	if err := steps.AddSequentialStep(&Step{Name: "s1", Template: "echo"}); err != nil {
		t.Fatal(err)
	}
	if err := steps.AddSequentialStep(&Step{Name: "s2", Template: "echo"}); err != nil {
		t.Fatal(err)
	}

	w := &Workflow{
		Name:       "steps-roundtrip",
		Entrypoint: "sequential",
		Templates: []Templatable{
			steps,
			&Container{Name: "echo", Image: "alpine", Command: []string{"echo"}},
		},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}

	wf, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}

	var stepsGroups [][]model.StepModel
	for _, tpl := range wf.Spec.Templates {
		if tpl.Name == "sequential" {
			stepsGroups = tpl.Steps
		}
	}
	if len(stepsGroups) != 2 {
		t.Fatalf("step groups = %d, want 2", len(stepsGroups))
	}
	if stepsGroups[0][0].Name != "s1" {
		t.Errorf("step 0 = %q", stepsGroups[0][0].Name)
	}
	if stepsGroups[1][0].Name != "s2" {
		t.Errorf("step 1 = %q", stepsGroups[1][0].Name)
	}
}

func TestRoundTripWorkflowTemplate(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "my-template",
		Namespace:  "argo",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine", Command: []string{"echo", "hello"}},
		},
	}

	yamlStr, err := wt.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}

	var wfModel model.WorkflowTemplateModel
	if err := yaml.Unmarshal([]byte(yamlStr), &wfModel); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if wfModel.Kind != "WorkflowTemplate" {
		t.Errorf("kind = %q", wfModel.Kind)
	}
	if wfModel.Metadata.Name != "my-template" {
		t.Errorf("name = %q", wfModel.Metadata.Name)
	}
	if wfModel.Metadata.Namespace != "argo" {
		t.Errorf("namespace = %q", wfModel.Metadata.Namespace)
	}
}

func TestRoundTripCronWorkflow(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "daily-job",
		Schedule:   "0 0 * * *",
		Timezone:   "UTC",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine", Command: []string{"echo", "cron"}},
		},
	}

	jsonStr, err := cw.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	var cwModel model.CronWorkflowModel
	if err := json.Unmarshal([]byte(jsonStr), &cwModel); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if cwModel.Kind != "CronWorkflow" {
		t.Errorf("kind = %q", cwModel.Kind)
	}
	if cwModel.Spec.Schedule != "0 0 * * *" {
		t.Errorf("schedule = %q", cwModel.Spec.Schedule)
	}
	if cwModel.Spec.Timezone != "UTC" {
		t.Errorf("timezone = %q", cwModel.Spec.Timezone)
	}
}

func TestRoundTripPreservesContainerFields(t *testing.T) {
	w := &Workflow{
		Name:       "fields-test",
		Entrypoint: "worker",
		Templates: []Templatable{
			&Container{
				Name:       "worker",
				Image:      "node:20",
				Command:    []string{"node"},
				Args:       []string{"index.js"},
				WorkingDir: "/app",
				Resources: &ResourceRequirements{
					Requests: ResourceList{CPU: "100m", Memory: "128Mi"},
					Limits:   ResourceList{CPU: "500m", Memory: "512Mi"},
				},
			},
		},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}

	wf, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}

	c := wf.Spec.Templates[0].Container
	if c.WorkingDir != "/app" {
		t.Errorf("workingDir = %q", c.WorkingDir)
	}
	if c.Resources == nil {
		t.Fatal("resources is nil")
	}
	if c.Resources.Requests.CPU != "100m" {
		t.Errorf("cpu request = %q", c.Resources.Requests.CPU)
	}
	if c.Resources.Limits.Memory != "512Mi" {
		t.Errorf("memory limit = %q", c.Resources.Limits.Memory)
	}
}

func TestBuildErrorPropagatesFromArguments(t *testing.T) {
	w := &Workflow{
		Name:       "error-test",
		Entrypoint: "main",
		Arguments: []Parameter{
			{Name: "", Value: strPtr("val")}, // empty name
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	_, err := w.Build()
	if err == nil {
		t.Fatal("expected error for argument with empty name")
	}
}

func TestBuildErrorPropagatesFromInputs(t *testing.T) {
	w := &Workflow{
		Name:       "error-test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{
				Name:   "main",
				Image:  "alpine",
				Inputs: []Parameter{{Name: ""}}, // empty name
			},
		},
	}
	_, err := w.Build()
	if err == nil {
		t.Fatal("expected error for input parameter with empty name")
	}
}

func TestBuildErrorPropagatesFromOutputs(t *testing.T) {
	w := &Workflow{
		Name:       "error-test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{
				Name:    "main",
				Image:   "alpine",
				Outputs: []Parameter{{Name: ""}}, // empty name
			},
		},
	}
	_, err := w.Build()
	if err == nil {
		t.Fatal("expected error for output parameter with empty name")
	}
}

func TestNewConfigIsIndependent(t *testing.T) {
	cfg1 := NewConfig()
	cfg2 := NewConfig()

	cfg1.SetImage("custom:latest")

	if cfg2.GetImage() == "custom:latest" {
		t.Error("NewConfig instances should be independent")
	}
	if cfg2.GetImage() != "python:3.11" {
		t.Errorf("default image = %q", cfg2.GetImage())
	}
}

func TestNewConfigDoesNotAffectGlobal(t *testing.T) {
	global := GetGlobalConfig()
	originalImage := global.GetImage()
	defer global.SetImage(originalImage)

	cfg := NewConfig()
	cfg.SetImage("isolated:1.0")

	if global.GetImage() == "isolated:1.0" {
		t.Error("NewConfig should not affect global config")
	}
}

func strPtr(s string) *string {
	return &s
}
