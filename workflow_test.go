package forge

import (
	"strings"
	"testing"

	"github.com/usetheo/theo/forge/model"
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

// --- WorkflowTemplate tests (consolidated from workflow_template_test.go) ---

func TestWorkflowTemplateBuild(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "my-template",
		Namespace:  "default",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine", Command: []string{"echo"}, Args: []string{"hello"}},
		},
	}
	model, err := wt.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Kind != "WorkflowTemplate" {
		t.Errorf("kind = %q", model.Kind)
	}
	if model.APIVersion != DefaultAPIVersion {
		t.Errorf("apiVersion = %q", model.APIVersion)
	}
	if model.Metadata.Name != "my-template" {
		t.Errorf("name = %q", model.Metadata.Name)
	}
	if model.Metadata.Namespace != "default" {
		t.Errorf("namespace = %q", model.Metadata.Namespace)
	}
	if len(model.Spec.Templates) != 1 {
		t.Fatalf("templates = %d", len(model.Spec.Templates))
	}
}

func TestWorkflowTemplateNoNameFails(t *testing.T) {
	wt := &WorkflowTemplate{Entrypoint: "main"}
	_, err := wt.Build()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestWorkflowTemplateToYAML(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "yaml-test",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	y, err := wt.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(y, "kind: WorkflowTemplate") {
		t.Error("missing kind in YAML")
	}
}

func TestWorkflowTemplateWithArguments(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "with-args",
		Entrypoint: "main",
		Arguments:  []Parameter{{Name: "msg", Value: ptrStr("hello")}},
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := wt.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Arguments == nil || len(model.Spec.Arguments.Parameters) != 1 {
		t.Fatal("expected 1 argument")
	}
}

// --- ClusterWorkflowTemplate ---

func TestClusterWorkflowTemplateBuild(t *testing.T) {
	cwt := &ClusterWorkflowTemplate{
		Name:       "cluster-template",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine"},
		},
	}
	model, err := cwt.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Kind != "ClusterWorkflowTemplate" {
		t.Errorf("kind = %q", model.Kind)
	}
	if model.Metadata.Name != "cluster-template" {
		t.Errorf("name = %q", model.Metadata.Name)
	}
	// Cluster-scope: no namespace
	if model.Metadata.Namespace != "" {
		t.Errorf("namespace should be empty for cluster-scope, got %q", model.Metadata.Namespace)
	}
}

func TestClusterWorkflowTemplateNoNameFails(t *testing.T) {
	cwt := &ClusterWorkflowTemplate{Entrypoint: "main"}
	_, err := cwt.Build()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestClusterWorkflowTemplateToYAML(t *testing.T) {
	cwt := &ClusterWorkflowTemplate{
		Name:       "yaml-cwt",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	y, err := cwt.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(y, "kind: ClusterWorkflowTemplate") {
		t.Error("missing kind in YAML")
	}
}

// --- CronWorkflow ---

func TestCronWorkflowBuild(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "daily-job",
		Namespace:  "default",
		Schedule:   "0 0 * * *",
		Timezone:   "America/Sao_Paulo",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine", Command: []string{"echo"}, Args: []string{"daily"}},
		},
	}
	model, err := cw.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Kind != "CronWorkflow" {
		t.Errorf("kind = %q", model.Kind)
	}
	if model.Spec.Schedule != "0 0 * * *" {
		t.Errorf("schedule = %q", model.Spec.Schedule)
	}
	if model.Spec.Timezone != "America/Sao_Paulo" {
		t.Errorf("timezone = %q", model.Spec.Timezone)
	}
	if model.Spec.WorkflowSpec.Entrypoint != "main" {
		t.Errorf("entrypoint = %q", model.Spec.WorkflowSpec.Entrypoint)
	}
}

func TestCronWorkflowNoNameFails(t *testing.T) {
	cw := &CronWorkflow{Schedule: "0 * * * *"}
	_, err := cw.Build()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCronWorkflowNoScheduleFails(t *testing.T) {
	cw := &CronWorkflow{Name: "test"}
	_, err := cw.Build()
	if err == nil {
		t.Fatal("expected error for empty schedule")
	}
}

func TestCronWorkflowConcurrencyPolicy(t *testing.T) {
	cw := &CronWorkflow{
		Name:              "with-policy",
		Schedule:          "*/5 * * * *",
		ConcurrencyPolicy: "Replace",
		Entrypoint:        "main",
		Templates:         []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := cw.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.ConcurrencyPolicy != "Replace" {
		t.Errorf("policy = %q", model.Spec.ConcurrencyPolicy)
	}
}

func TestCronWorkflowSuspend(t *testing.T) {
	suspended := true
	cw := &CronWorkflow{
		Name:       "suspended",
		Schedule:   "0 * * * *",
		Suspend:    &suspended,
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := cw.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Suspend == nil || !*model.Spec.Suspend {
		t.Error("expected suspend to be true")
	}
}

func TestCronWorkflowHistoryLimits(t *testing.T) {
	success := 5
	failed := 3
	cw := &CronWorkflow{
		Name:                       "with-limits",
		Schedule:                   "0 * * * *",
		SuccessfulJobsHistoryLimit: &success,
		FailedJobsHistoryLimit:     &failed,
		Entrypoint:                 "main",
		Templates:                  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := cw.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.SuccessfulJobsHistoryLimit == nil || *model.Spec.SuccessfulJobsHistoryLimit != 5 {
		t.Error("expected success limit = 5")
	}
	if model.Spec.FailedJobsHistoryLimit == nil || *model.Spec.FailedJobsHistoryLimit != 3 {
		t.Error("expected failed limit = 3")
	}
}

func TestCronWorkflowToYAML(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "yaml-cron",
		Schedule:   "0 * * * *",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	y, err := cw.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(y, "kind: CronWorkflow") {
		t.Error("missing kind in YAML")
	}
	if !strings.Contains(y, "schedule:") {
		t.Error("missing schedule in YAML")
	}
}

func TestCronWorkflowToJSON(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "json-cron",
		Schedule:   "0 * * * *",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	j, err := cw.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(j, `"kind": "CronWorkflow"`) {
		t.Error("missing kind in JSON")
	}
}

// --- UserContainer tests (consolidated from user_container_test.go) ---

func TestUserContainerBuild(t *testing.T) {
	uc := &UserContainer{
		Name:    "sidecar",
		Image:   "nginx:latest",
		Command: []string{"nginx", "-g", "daemon off;"},
		Ports:   []ContainerPort{{ContainerPort: 80}},
	}
	m := uc.Build()
	if m.Name != "sidecar" {
		t.Errorf("name = %q", m.Name)
	}
	if m.Image != "nginx:latest" {
		t.Errorf("image = %q", m.Image)
	}
	if len(m.Ports) != 1 || m.Ports[0].ContainerPort != 80 {
		t.Errorf("ports = %v", m.Ports)
	}
}

func TestUserContainerWithImagePullPolicy(t *testing.T) {
	uc := &UserContainer{
		Name:            "test",
		Image:           "alpine",
		ImagePullPolicy: ImagePullIfNotPresent,
	}
	m := uc.Build()
	if m.ImagePullPolicy != "IfNotPresent" {
		t.Errorf("policy = %q", m.ImagePullPolicy)
	}
}

func TestUserContainerWithVolumeMounts(t *testing.T) {
	uc := &UserContainer{
		Name:  "test",
		Image: "alpine",
		VolumeMounts: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "data", MountPath: "/data"}},
		},
	}
	m := uc.Build()
	if len(m.VolumeMounts) != 1 {
		t.Fatalf("mounts = %d", len(m.VolumeMounts))
	}
	if m.VolumeMounts[0].Name != "data" {
		t.Errorf("mount name = %q", m.VolumeMounts[0].Name)
	}
	if m.VolumeMounts[0].MountPath != "/data" {
		t.Errorf("mount path = %q", m.VolumeMounts[0].MountPath)
	}
}

func TestUserContainerWithMultipleVolumeMounts(t *testing.T) {
	uc := &UserContainer{
		Name:  "test",
		Image: "alpine",
		VolumeMounts: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "test1", MountPath: "/test1"}},
			&PVCVolume{BaseVolume: BaseVolume{Name: "test2", MountPath: "/test2"}, Size: "1Gi"},
		},
	}
	m := uc.Build()
	if len(m.VolumeMounts) != 2 {
		t.Fatalf("mounts = %d, want 2", len(m.VolumeMounts))
	}
	if m.VolumeMounts[0].Name != "test1" {
		t.Errorf("mount[0] name = %q", m.VolumeMounts[0].Name)
	}
	if m.VolumeMounts[1].Name != "test2" {
		t.Errorf("mount[1] name = %q", m.VolumeMounts[1].Name)
	}
}

func TestUserContainerAsInitContainer(t *testing.T) {
	initC := &UserContainer{
		Name:    "init",
		Image:   "alpine",
		Command: []string{"sh", "-c"},
		Args:    []string{"echo initializing"},
	}

	c := &Container{
		Name:    "main",
		Image:   "alpine",
		Command: []string{"echo"},
		Args:    []string{"hello"},
	}

	// Verify the init container model can be used in a template's InitContainers
	initModel := initC.Build()
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	tpl.InitContainers = []model.ContainerModel{initModel}

	if len(tpl.InitContainers) != 1 {
		t.Fatalf("initContainers = %d", len(tpl.InitContainers))
	}
	if tpl.InitContainers[0].Name != "init" {
		t.Errorf("init name = %q", tpl.InitContainers[0].Name)
	}
}

func TestUserContainerAsSidecar(t *testing.T) {
	sidecar := &UserContainer{
		Name:    "log-collector",
		Image:   "fluentd:latest",
		Env:     []EnvBuilder{Env{Name: "LOG_LEVEL", Value: "debug"}},
	}

	c := &Container{
		Name:  "main",
		Image: "alpine",
	}

	sidecarModel := sidecar.Build()
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	tpl.Sidecars = []model.ContainerModel{sidecarModel}

	if len(tpl.Sidecars) != 1 {
		t.Fatalf("sidecars = %d", len(tpl.Sidecars))
	}
	if tpl.Sidecars[0].Name != "log-collector" {
		t.Errorf("sidecar name = %q", tpl.Sidecars[0].Name)
	}
	if len(tpl.Sidecars[0].Env) != 1 || tpl.Sidecars[0].Env[0].Name != "LOG_LEVEL" {
		t.Error("expected sidecar env")
	}
}

func TestWorkflowTemplateWithDefaultParam(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "with-default",
		Entrypoint: "main",
		Arguments: []Parameter{
			{Name: "my-arg", Default: ptrStr("foo")},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := wt.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Arguments == nil || len(model.Spec.Arguments.Parameters) != 1 {
		t.Fatal("expected 1 argument")
	}
	p := model.Spec.Arguments.Parameters[0]
	if p.Name != "my-arg" {
		t.Errorf("name = %q", p.Name)
	}
	// AsArgument doesn't include Default, so this tests the argument building path
}

// --- GlobalConfig tests (consolidated from global_config_test.go) ---

func TestGlobalConfigDefaults(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	if cfg.GetImage() != "python:3.11" {
		t.Errorf("default image = %q, want 'python:3.11'", cfg.GetImage())
	}
	if !cfg.VerifySSL {
		t.Error("default VerifySSL should be true")
	}
}

func TestGlobalConfigSetImage(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	cfg.SetImage("alpine:3.18")
	if cfg.GetImage() != "alpine:3.18" {
		t.Errorf("image = %q, want 'alpine:3.18'", cfg.GetImage())
	}
}

func TestGlobalConfigSetNamespace(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	cfg.SetNamespace("workflows")
	if cfg.GetNamespace() != "workflows" {
		t.Errorf("namespace = %q", cfg.GetNamespace())
	}
}

func TestGlobalConfigTemplateHook(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	// Register hook that sets a default image
	cfg.RegisterTemplateHook(func(tpl *model.TemplateModel) {
		if tpl.Container != nil && tpl.Container.Image == "" {
			tpl.Container.Image = "default-image:latest"
		}
	})

	// Build a container with no image
	tpl := &model.TemplateModel{
		Name:      "test",
		Container: &model.ContainerModel{Image: ""},
	}
	cfg.DispatchTemplateHooks(tpl)

	if tpl.Container.Image != "default-image:latest" {
		t.Errorf("image after hook = %q", tpl.Container.Image)
	}
}

func TestGlobalConfigWorkflowHook(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	// Register hook that adds a label
	cfg.RegisterWorkflowHook(func(wf *model.WorkflowModel) {
		if wf.Metadata.Labels == nil {
			wf.Metadata.Labels = make(map[string]string)
		}
		wf.Metadata.Labels["managed-by"] = "forge"
	})

	wf := &model.WorkflowModel{
		Metadata: model.WorkflowMetadata{Name: "test"},
	}
	cfg.DispatchWorkflowHooks(wf)

	if wf.Metadata.Labels["managed-by"] != "forge" {
		t.Errorf("label = %q", wf.Metadata.Labels["managed-by"])
	}
}

func TestGlobalConfigMultipleHooks(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	callOrder := []string{}
	cfg.RegisterTemplateHook(func(tpl *model.TemplateModel) {
		callOrder = append(callOrder, "first")
	})
	cfg.RegisterTemplateHook(func(tpl *model.TemplateModel) {
		callOrder = append(callOrder, "second")
	})

	tpl := &model.TemplateModel{Name: "test"}
	cfg.DispatchTemplateHooks(tpl)

	if len(callOrder) != 2 {
		t.Fatalf("hooks called = %d, want 2", len(callOrder))
	}
	if callOrder[0] != "first" || callOrder[1] != "second" {
		t.Errorf("order = %v, want [first, second]", callOrder)
	}
}

func TestGlobalConfigClearHooks(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	called := false
	cfg.RegisterTemplateHook(func(tpl *model.TemplateModel) {
		called = true
	})
	cfg.ClearHooks()

	tpl := &model.TemplateModel{Name: "test"}
	cfg.DispatchTemplateHooks(tpl)

	if called {
		t.Error("hook should not be called after ClearHooks")
	}
}

func TestGlobalConfigReset(t *testing.T) {
	cfg := GetGlobalConfig()

	cfg.SetImage("custom:v1")
	cfg.SetNamespace("custom-ns")
	cfg.SetHost("https://custom.host")
	cfg.SetToken("secret")

	cfg.Reset()

	if cfg.GetImage() != "python:3.11" {
		t.Errorf("image after reset = %q", cfg.GetImage())
	}
	if cfg.GetNamespace() != "" {
		t.Errorf("namespace after reset = %q", cfg.GetNamespace())
	}
	if cfg.Host != "" {
		t.Errorf("host after reset = %q", cfg.Host)
	}
	if cfg.Token != "" {
		t.Errorf("token after reset = %q", cfg.Token)
	}
}

// --- Status tests (consolidated from status_test.go) ---

func TestParseWorkflowStatus(t *testing.T) {
	tests := []struct {
		input string
		want  WorkflowStatus
		err   bool
	}{
		{"Pending", WorkflowPending, false},
		{"Running", WorkflowRunning, false},
		{"Succeeded", WorkflowSucceeded, false},
		{"Failed", WorkflowFailed, false},
		{"Error", WorkflowError, false},
		{"Terminated", WorkflowTerminated, false},
		{"Unknown", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := model.ParseWorkflowStatus(tt.input)
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

func TestRetryStrategyBuild(t *testing.T) {
	limit := 3
	rs := RetryStrategy{
		Limit:       &limit,
		RetryPolicy: RetryOnFailure,
		Backoff: &Backoff{
			Duration:    "5s",
			Factor:      ptrInt(2),
			MaxDuration: "1m",
		},
	}
	m := rs.Build()
	if m.Limit != "3" {
		t.Errorf("limit = %v, want \"3\"", m.Limit)
	}
	if m.RetryPolicy != "OnFailure" {
		t.Errorf("policy = %q", m.RetryPolicy)
	}
	if m.Backoff == nil || m.Backoff.Duration != "5s" {
		t.Errorf("backoff duration = %v", m.Backoff)
	}
}

func TestMetricStructure(t *testing.T) {
	m := Metric{
		Name: "build_duration",
		Help: "Duration of build step",
		Labels: []Label{{Key: "step", Value: "build"}},
		Gauge:  &Gauge{Value: "{{duration}}", Realtime: ptrBool(true)},
	}
	if m.Name != "build_duration" {
		t.Errorf("name = %q", m.Name)
	}
	if m.Gauge == nil || m.Gauge.Realtime == nil || !*m.Gauge.Realtime {
		t.Error("expected realtime gauge")
	}
}

func TestErrorTypes(t *testing.T) {
	t.Run("InvalidType", func(t *testing.T) {
		err := &model.InvalidType{Expected: "Task", Got: "Step"}
		if err.Error() != "invalid type: expected Task, got Step" {
			t.Errorf("unexpected message: %s", err.Error())
		}
	})
	t.Run("NodeNameConflict", func(t *testing.T) {
		err := &NodeNameConflict{Name: "my-task"}
		if err.Error() != `node name conflict: "my-task" already exists in this context` {
			t.Errorf("unexpected message: %s", err.Error())
		}
	})
	t.Run("InvalidTemplateCall", func(t *testing.T) {
		err := &InvalidTemplateCall{Name: "echo", Context: "Workflow"}
		if err.Error() != `template "echo" is not callable under a Workflow context` {
			t.Errorf("unexpected message: %s", err.Error())
		}
	})
}

// --- Output refs tests (consolidated from output_refs_test.go) ---

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

	wfModel, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}

	// Find the consume task in DAG and verify its argument
	dagTpl := wfModel.Spec.Templates[2]
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
