package forge

import (
	"strings"
	"testing"
)

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
