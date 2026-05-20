package forge

import (
	"errors"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/config"
	"github.com/usetheodev/theo-forge/model"
)

// T3.3 — Template/TemplateRef/Inline mutual exclusion (code-p4-template-ref-mutual-exclusion).

func TestTask_BuildDAGTask_RequiresOneReference(t *testing.T) {
	task := &Task{Name: "x"} // no Template, TemplateRef, or Inline
	_, err := task.BuildDAGTask()
	if !errors.Is(err, model.ErrTemplateMissing) {
		t.Fatalf("got %v, want ErrTemplateMissing", err)
	}
}

func TestTask_BuildDAGTask_RejectsTemplateAndTemplateRef(t *testing.T) {
	task := &Task{Name: "x", Template: "tpl", TemplateRef: &model.TemplateRef{Name: "wft", Template: "main"}}
	_, err := task.BuildDAGTask()
	if !errors.Is(err, model.ErrTemplateAmbiguous) {
		t.Fatalf("got %v, want ErrTemplateAmbiguous", err)
	}
}

func TestStep_BuildStep_RequiresOneReference(t *testing.T) {
	step := &Step{Name: "x"}
	_, err := step.BuildStep()
	if !errors.Is(err, model.ErrTemplateMissing) {
		t.Fatalf("got %v, want ErrTemplateMissing", err)
	}
}

// T3.4 — ConfigMapName separates volume name from CM name.
func TestConfigMapVolume_UsesConfigMapNameWhenSet(t *testing.T) {
	v := ConfigMapVolume{
		BaseVolume:    BaseVolume{Name: "mount-name", MountPath: "/etc/cfg"},
		ConfigMapName: "real-cm",
	}
	out, err := v.BuildVolume()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.ConfigMap.Name != "real-cm" {
		t.Fatalf("ConfigMap.Name = %q, want \"real-cm\"", out.ConfigMap.Name)
	}
}

func TestConfigMapVolume_FallsBackToMountNameWhenConfigMapNameEmpty(t *testing.T) {
	v := ConfigMapVolume{
		BaseVolume: BaseVolume{Name: "mount", MountPath: "/etc/cfg"},
	}
	out, err := v.BuildVolume()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.ConfigMap.Name != "mount" {
		t.Fatalf("backward compat broken: ConfigMap.Name = %q, want \"mount\"", out.ConfigMap.Name)
	}
}

// T3.5 — Image validation after defaults applied.
func TestContainer_BuildTemplate_RejectsEmptyImageWithoutDefault(t *testing.T) {
	config.ResetGlobalConfigForTest(t)
	cfg := NewConfig()
	cfg.Reset()
	cfg.SetImage("") // no default available
	w := (&Workflow{
		Name:       "wf",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main"},
		},
	}).WithConfig(cfg)
	_, err := w.Build()
	if !errors.Is(err, model.ErrEmptyImage) {
		t.Fatalf("got %v, want ErrEmptyImage", err)
	}
}

// T3.6 — Entrypoint required when templates exist.
func TestWorkflow_RejectsMissingEntrypointWhenTemplatesPresent(t *testing.T) {
	w := &Workflow{
		Name: "wf",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine"},
		},
	}
	_, err := w.Build()
	if !errors.Is(err, model.ErrEntrypointMissing) {
		t.Fatalf("got %v, want ErrEntrypointMissing", err)
	}
}

func TestWorkflow_AcceptsMissingEntrypointWhenWorkflowTemplateRefSet(t *testing.T) {
	w := &Workflow{
		Name:                "wf",
		WorkflowTemplateRef: &model.WorkflowTemplateRef{Name: "wft"},
	}
	if _, err := w.Build(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

// T3.8 — NameLimit on ClusterWorkflowTemplate and CronWorkflow.
func TestClusterWorkflowTemplate_RejectsOversizedName(t *testing.T) {
	cwt := &ClusterWorkflowTemplate{
		Name:       strings.Repeat("a", NameLimit+1),
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	_, err := cwt.Build()
	if err == nil || !strings.Contains(err.Error(), "no more than") {
		t.Fatalf("got %v, want NameLimit error", err)
	}
}

func TestCronWorkflow_RejectsOversizedName(t *testing.T) {
	cw := &CronWorkflow{
		Name:       strings.Repeat("a", NameLimit+1),
		Schedule:   "*/5 * * * *",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	_, err := cw.Build()
	if err == nil || !strings.Contains(err.Error(), "no more than") {
		t.Fatalf("got %v, want NameLimit error", err)
	}
}

// T3.9 — Then() preserves OR precedence by parenthesizing existing Depends.
func TestTask_Then_PreservesOrPrecedence(t *testing.T) {
	A := &Task{Name: "A"}
	B := &Task{Name: "B", Depends: "X || Y"} // pre-existing OR
	A.Then(B)
	want := "(X || Y) && A"
	if B.Depends != want {
		t.Fatalf("Depends = %q, want %q", B.Depends, want)
	}
}

func TestTask_Then_NoOpOnEmptyDepends(t *testing.T) {
	A := &Task{Name: "A"}
	B := &Task{Name: "B"}
	A.Then(B)
	if B.Depends != "A" {
		t.Fatalf("Depends = %q, want \"A\"", B.Depends)
	}
}

func TestTask_Then_NoExtraParensForSimpleAnd(t *testing.T) {
	A := &Task{Name: "A"}
	B := &Task{Name: "B", Depends: "X"}
	A.Then(B)
	if B.Depends != "X && A" {
		t.Fatalf("Depends = %q, want \"X && A\"", B.Depends)
	}
}
