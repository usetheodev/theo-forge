package forge

import (
	"encoding/json"
	"testing"

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

	model, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}

	if model.APIVersion != DefaultAPIVersion {
		t.Errorf("apiVersion = %q, want %q", model.APIVersion, DefaultAPIVersion)
	}
	if model.Kind != DefaultKind {
		t.Errorf("kind = %q, want %q", model.Kind, DefaultKind)
	}
	if model.Metadata.Name != "roundtrip-yaml" {
		t.Errorf("name = %q", model.Metadata.Name)
	}
	if model.Metadata.Namespace != "default" {
		t.Errorf("namespace = %q", model.Metadata.Namespace)
	}
	if model.Metadata.Labels["app"] != "test" {
		t.Errorf("label app = %q", model.Metadata.Labels["app"])
	}
	if model.Metadata.Annotations["note"] != "roundtrip" {
		t.Errorf("annotation note = %q", model.Metadata.Annotations["note"])
	}
	if model.Spec.Entrypoint != "main" {
		t.Errorf("entrypoint = %q", model.Spec.Entrypoint)
	}
	if len(model.Spec.Templates) != 1 {
		t.Fatalf("templates = %d", len(model.Spec.Templates))
	}
	if model.Spec.Templates[0].Container == nil {
		t.Fatal("expected container template")
	}
	if model.Spec.Templates[0].Container.Image != "alpine:3.18" {
		t.Errorf("image = %q", model.Spec.Templates[0].Container.Image)
	}
	if model.Spec.Arguments == nil || len(model.Spec.Arguments.Parameters) != 1 {
		t.Fatalf("expected 1 argument parameter")
	}
	if model.Spec.Arguments.Parameters[0].Name != "greeting" {
		t.Errorf("argument name = %q", model.Spec.Arguments.Parameters[0].Name)
	}
	if model.Spec.Arguments.Parameters[0].Value == nil || *model.Spec.Arguments.Parameters[0].Value != "world" {
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

	model, err := FromJSON(jsonStr)
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}

	if model.Metadata.Name != "roundtrip-json" {
		t.Errorf("name = %q", model.Metadata.Name)
	}
	if len(model.Spec.Templates) != 1 {
		t.Fatalf("templates = %d", len(model.Spec.Templates))
	}
	if model.Spec.Templates[0].Script == nil {
		t.Fatal("expected script template")
	}
	if model.Spec.Templates[0].Script.Source != "print('hello')" {
		t.Errorf("source = %q", model.Spec.Templates[0].Script.Source)
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

	model, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}

	if len(model.Spec.Templates) != 2 {
		t.Fatalf("templates = %d", len(model.Spec.Templates))
	}

	var dagModel *DAGModel
	for _, tpl := range model.Spec.Templates {
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

	taskMap := make(map[string]DAGTaskModel)
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

	model, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}

	var stepsGroups [][]StepModel
	for _, tpl := range model.Spec.Templates {
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

	var model WorkflowTemplateModel
	if err := yaml.Unmarshal([]byte(yamlStr), &model); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if model.Kind != "WorkflowTemplate" {
		t.Errorf("kind = %q", model.Kind)
	}
	if model.Metadata.Name != "my-template" {
		t.Errorf("name = %q", model.Metadata.Name)
	}
	if model.Metadata.Namespace != "argo" {
		t.Errorf("namespace = %q", model.Metadata.Namespace)
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

	var model CronWorkflowModel
	if err := json.Unmarshal([]byte(jsonStr), &model); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if model.Kind != "CronWorkflow" {
		t.Errorf("kind = %q", model.Kind)
	}
	if model.Spec.Schedule != "0 0 * * *" {
		t.Errorf("schedule = %q", model.Spec.Schedule)
	}
	if model.Spec.Timezone != "UTC" {
		t.Errorf("timezone = %q", model.Spec.Timezone)
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

	model, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}

	c := model.Spec.Templates[0].Container
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
