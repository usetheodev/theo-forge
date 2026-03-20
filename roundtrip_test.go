package forge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/expr"
	"github.com/usetheodev/theo-forge/model"
	"github.com/usetheodev/theo-forge/serialize"
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

// --- Golden tests (consolidated from golden_test.go) ---

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
	yamlOut, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "simple_container", yamlOut)
}

func TestGoldenDiamondDAG(t *testing.T) {
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
	_ = dag.AddTasks(A, B, C, D)

	w := &Workflow{
		GenerateName: "diamond-",
		Namespace:    "argo",
		Entrypoint:   "diamond",
		Templates:    []Templatable{echoTpl, dag},
	}

	yamlOut, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "diamond_dag", yamlOut)
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
	yamlOut, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "script_workflow", yamlOut)
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
	yamlOut, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "steps_workflow", yamlOut)
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
	yamlOut, err := wt.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "workflow_template", yamlOut)
}

func TestGoldenCronWorkflow(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "nightly-build",
		Namespace:  "ci",
		Schedule:   "0 2 * * *",
		Timezone:   "America/Sao_Paulo",
		Entrypoint: "build",
		Templates: []Templatable{
			&Container{
				Name:    "build",
				Image:   "golang:1.22",
				Command: []string{"make", "build"},
			},
		},
	}
	yamlOut, err := cw.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "cron_workflow", yamlOut)
}

// --- Round-trip example tests (using testdata/examples/) ---

func TestRoundTripTestdataExamples(t *testing.T) {
	examplesDir := "testdata/examples"

	var tested int
	err := filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		// Get relative name for test identification
		relPath, _ := filepath.Rel(examplesDir, path)
		name := strings.TrimSuffix(relPath, ".yaml")

		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			yamlStr := string(data)

			kind := detectKind(yamlStr)

			switch kind {
			case "Workflow":
				roundTripWorkflow(t, name, yamlStr)
			case "WorkflowTemplate", "ClusterWorkflowTemplate":
				roundTripWorkflowTemplate(t, name, yamlStr)
			case "CronWorkflow":
				roundTripCronWorkflow(t, name, yamlStr)
			default:
				t.Skipf("unknown kind %q in %s", kind, name)
			}
		})
		tested++
		return nil
	})
	if err != nil {
		t.Fatalf("walk testdata/examples: %v", err)
	}

	if tested == 0 {
		t.Fatal("no examples found in testdata/examples/")
	}
	t.Logf("Round-trip tested %d testdata examples", tested)
}

// TestRoundTripWorkflowBuilder verifies that workflows built programmatically
// produce valid models that round-trip cleanly.
func TestRoundTripWorkflowBuilder(t *testing.T) {
	builders := map[string]func() *Workflow{
		"hello-world": buildHelloWorld,
		"steps":       buildSteps,
		"dag-diamond": buildDagDiamond,
	}

	for name, builder := range builders {
		t.Run(name, func(t *testing.T) {
			w := builder()

			// Build to model
			m, err := w.Build()
			if err != nil {
				t.Fatalf("build model %s: %v", name, err)
			}

			// Serialize
			yaml1, err := serialize.WorkflowToYAML(m)
			if err != nil {
				t.Fatalf("serialize %s: %v", name, err)
			}

			// Parse back
			m2, err := serialize.WorkflowFromYAML(yaml1)
			if err != nil {
				t.Fatalf("parse back %s: %v", name, err)
			}

			// Re-serialize
			yaml2, err := serialize.WorkflowToYAML(m2)
			if err != nil {
				t.Fatalf("re-serialize %s: %v", name, err)
			}

			// Should be identical
			if yaml1 != yaml2 {
				t.Errorf("round-trip not stable for %s", name)
			}
		})
	}
}

// Ensure CronWorkflowFromYAML exists
func init() {
	// Verify serialize package has the necessary functions
	_ = serialize.WorkflowFromYAML
	_ = serialize.WorkflowToYAML
	_ = serialize.WorkflowTemplateFromYAML
	_ = serialize.WorkflowTemplateToYAML
	_ = serialize.CronWorkflowFromYAML
	_ = serialize.CronWorkflowToYAML
}

// Additional test to verify we can programmatically create a workflow using model types
// directly for any feature.
func TestModelDirectConstruction(t *testing.T) {
	// Build a workflow with synchronization, hooks, memoize, etc.
	m := model.WorkflowModel{
		APIVersion: DefaultAPIVersion,
		Kind:       "Workflow",
		Metadata:   model.WorkflowMetadata{GenerateName: "feature-rich-"},
		Spec: model.WorkflowSpec{
			Entrypoint: "main",
			Synchronization: &model.SynchronizationModel{
				Mutex: &model.MutexModel{Name: "test-mutex"},
			},
			Hooks: map[string]model.LifecycleHook{
				"exit": {Template: "exit-handler"},
			},
			PodSpecPatch: `{"containers":[{"name":"main","resources":{"limits":{"cpu":"1"}}}]}`,
			Templates: []model.TemplateModel{
				{
					Name: "main",
					Container: &model.ContainerModel{
						Image:   "alpine:3.18",
						Command: []string{"echo", "hello"},
					},
					Memoize: &model.MemoizeModel{
						Key:    "{{inputs.parameters.msg}}",
						MaxAge: "1h",
						Cache: &model.CacheModel{
							ConfigMap: &model.ConfigMapKeyRef{
								Name: "my-cache",
								Key:  "data",
							},
						},
					},
					Daemon: ptrBool(true),
					Synchronization: &model.SynchronizationModel{
						Semaphore: &model.SemaphoreModel{
							ConfigMapKeyRef: &model.ConfigMapKeyRef{
								Name: "semaphore-config",
								Key:  "workflow",
							},
						},
					},
				},
				{
					Name: "exit-handler",
					Container: &model.ContainerModel{
						Image:   "alpine:3.18",
						Command: []string{"echo", "done"},
					},
				},
			},
		},
	}

	yamlStr, err := serialize.WorkflowToYAML(m)
	if err != nil {
		t.Fatal(err)
	}

	// Verify key features are in the YAML
	checks := []string{
		"synchronization:", "mutex:", "test-mutex",
		"hooks:", "exit:", "exit-handler",
		"podSpecPatch:", "memoize:", "daemon:",
		"semaphore:", "configMapKeyRef:",
	}
	for _, check := range checks {
		if !strings.Contains(yamlStr, check) {
			t.Errorf("YAML missing %q", check)
		}
	}
}
