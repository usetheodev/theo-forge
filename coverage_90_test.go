package forge

import (
	"strings"
	"testing"

	"github.com/usetheo/theo/forge/client"
	"github.com/usetheo/theo/forge/expr"
)

// Cover ToYAML/ToJSON/ToDict error paths (invalid workflow → Build fails → propagates)
func TestWorkflowToYAMLBuildError(t *testing.T) {
	w := &Workflow{Entrypoint: "main"} // no name
	_, err := w.ToYAML()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWorkflowToJSONBuildError(t *testing.T) {
	w := &Workflow{Entrypoint: "main"}
	_, err := w.ToJSON()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWorkflowToDictBuildError(t *testing.T) {
	w := &Workflow{Entrypoint: "main"}
	_, err := w.ToDict()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFromYAMLInvalid(t *testing.T) {
	_, err := FromYAML("{{{{bad yaml")
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover WorkflowTemplate ToYAML error path
func TestWorkflowTemplateToYAMLBuildError(t *testing.T) {
	wt := &WorkflowTemplate{} // no name
	_, err := wt.ToYAML()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover ClusterWorkflowTemplate ToYAML error path
func TestClusterWorkflowTemplateToYAMLBuildError(t *testing.T) {
	cwt := &ClusterWorkflowTemplate{} // no name
	_, err := cwt.ToYAML()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover CronWorkflow ToYAML/ToJSON error paths
func TestCronWorkflowToYAMLBuildError(t *testing.T) {
	cw := &CronWorkflow{} // no name, no schedule
	_, err := cw.ToYAML()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCronWorkflowToJSONBuildError(t *testing.T) {
	cw := &CronWorkflow{}
	_, err := cw.ToJSON()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover Workflow.Build with failing template
func TestWorkflowBuildWithFailingTemplate(t *testing.T) {
	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "", Image: "alpine"}, // no name → fails
		},
	}
	_, err := w.Build()
	if err == nil {
		t.Fatal("expected error from failing template")
	}
}

// Cover WorkflowTemplate.Build with failing template
func TestWorkflowTemplateBuildWithFailingTemplate(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Script{Name: "", Image: "alpine", Command: []string{"sh"}, Source: "echo"}, // no name
		},
	}
	_, err := wt.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover ClusterWorkflowTemplate.Build with failing template
func TestClusterWorkflowTemplateBuildWithFailingTemplate(t *testing.T) {
	cwt := &ClusterWorkflowTemplate{
		Name:       "test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "", Image: "alpine"},
		},
	}
	_, err := cwt.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover CronWorkflow.Build with failing template
func TestCronWorkflowBuildWithFailingTemplate(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "test",
		Schedule:   "0 * * * *",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "", Image: "alpine"},
		},
	}
	_, err := cw.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover WorkflowTemplate.Build with volumes
func TestWorkflowTemplateBuildWithVolumes(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "with-vols",
		Entrypoint: "main",
		Volumes: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "tmp", MountPath: "/tmp"}},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := wt.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Spec.Volumes) != 1 {
		t.Fatalf("volumes = %d", len(model.Spec.Volumes))
	}
}

// Cover CronWorkflow.Build with volumes and arguments
func TestCronWorkflowBuildWithVolsAndArgs(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "full",
		Schedule:   "0 * * * *",
		Entrypoint: "main",
		Arguments:  []Parameter{{Name: "env", Value: ptrStr("prod")}},
		Volumes: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "tmp", MountPath: "/tmp"}},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := cw.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.WorkflowSpec.Arguments == nil {
		t.Fatal("expected arguments")
	}
	if len(model.Spec.WorkflowSpec.Volumes) != 1 {
		t.Fatal("expected 1 volume")
	}
}

// Cover Workflow.ToFile error paths
func TestWorkflowToFileBuildError(t *testing.T) {
	w := &Workflow{Entrypoint: "main"} // no name
	_, err := w.ToFile(t.TempDir(), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover DAG.buildOutputs with output parameters
func TestDAGWithOutputParameters(t *testing.T) {
	dag := &DAG{
		Name: "with-out-params",
		Outputs: []Parameter{
			{Name: "result", ValueFrom: &ValueFrom{Expression: "tasks.final.outputs.result"}},
		},
	}
	tpl, err := dag.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Parameters) != 1 {
		t.Fatal("expected 1 output parameter")
	}
}

// Cover LintWorkflow build error path
func TestServiceLintWorkflowBuildError(t *testing.T) {
	svc := &client.WorkflowsService{Host: "https://argo.example.com", Namespace: "default"}
	w := &Workflow{Entrypoint: "main"} // no name
	_, err := svc.LintWorkflow(nil, w)
	if err == nil {
		t.Fatal("expected build error")
	}
}

// Cover CreateWorkflow build error path
func TestServiceCreateWorkflowBuildError(t *testing.T) {
	svc := &client.WorkflowsService{Host: "https://argo.example.com", Namespace: "default"}
	w := &Workflow{Entrypoint: "main"} // no name
	_, err := svc.CreateWorkflow(nil, w)
	if err == nil {
		t.Fatal("expected build error")
	}
}

// Cover AddParallelGroup with name conflict within group
func TestStepsAddParallelGroupInternalConflict(t *testing.T) {
	steps := &Steps{Name: "test"}
	err := steps.AddParallelGroup(
		&Step{Name: "dup", Template: "a"},
		&Step{Name: "dup", Template: "b"},
	)
	if err == nil {
		t.Fatal("expected name conflict within parallel group")
	}
}

// Cover S3/GCS/HTTP/Git/Raw artifact error paths (80% → 100%)
// The 80% is the base.Build() error path which propagates from Artifact.validate()
func TestS3ArtifactNoNameFails(t *testing.T) {
	a := S3Artifact{Artifact: Artifact{Path: "/tmp"}, Bucket: "b", Key: "k"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGCSArtifactNoNameFails(t *testing.T) {
	a := GCSArtifact{Artifact: Artifact{Path: "/tmp"}, Bucket: "b", Key: "k"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGitArtifactNoNameFails(t *testing.T) {
	a := GitArtifact{Artifact: Artifact{Path: "/tmp"}, Repo: "r"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRawArtifactNoNameFails(t *testing.T) {
	a := RawArtifact{Artifact: Artifact{Path: "/tmp"}, Data: "d"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHTTPArtifactNoNameFails(t *testing.T) {
	a := HTTPArtifact{Artifact: Artifact{Path: "/tmp"}, URL: "u"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover Expr.C default branch
func TestExprConstantDefault(t *testing.T) {
	type custom struct{ X int }
	e := expr.C(custom{X: 42})
	if !strings.Contains(e.String(), "42") {
		t.Errorf("got %q", e.String())
	}
}
