// Package docs hosts compile-checked extractions of every Go code block in
// README.md. If this package compiles, the README quickstart is valid Go
// against the current public API.
//
// T1.1 (val-001, val-002, val-003, code-p4-readme-broken-examples,
// completeness-readme-listworkflows-nil, completeness-readme-expr-eq-mismatch)
// + EC-2 (each Test* function calls config.ResetGlobalConfigForTest at top to
// prevent cross-test contamination through the package singleton).
package docs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	forge "github.com/usetheodev/theo-forge"
	"github.com/usetheodev/theo-forge/client"
	"github.com/usetheodev/theo-forge/config"
	"github.com/usetheodev/theo-forge/expr"
)

// resetGlobalConfig is a local wrapper around config.ResetGlobalConfigForTest
// so that README examples remain a single-line addition.
func resetGlobalConfig(t *testing.T) { config.ResetGlobalConfigForTest(t) }

// README §Quick Start — lines 46-75.
func TestReadme_QuickStart(t *testing.T) {
	resetGlobalConfig(t)
	w := &forge.Workflow{
		GenerateName: "hello-",
		Entrypoint:   "main",
		Templates: []forge.Templatable{
			&forge.Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"hello world"},
			},
		},
	}
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML err: %v", err)
	}
	if yaml == "" {
		t.Fatal("ToYAML returned empty string")
	}
}

// README §Diamond DAG — lines 103-130. Uses forge.InputParam + forge.Ptr
// (val-001, val-003).
func TestReadme_DiamondDAG(t *testing.T) {
	resetGlobalConfig(t)
	echoTpl := &forge.Container{
		Name:    "echo",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{forge.InputParam("msg")},
		Inputs:  []forge.Parameter{{Name: "msg"}},
	}
	dag := &forge.DAG{Name: "diamond"}
	A := &forge.Task{Name: "A", Template: "echo", Arguments: []forge.Parameter{{Name: "msg", Value: forge.Ptr("Task A")}}}
	B := &forge.Task{Name: "B", Template: "echo", Arguments: []forge.Parameter{{Name: "msg", Value: forge.Ptr("Task B")}}}
	C := &forge.Task{Name: "C", Template: "echo", Arguments: []forge.Parameter{{Name: "msg", Value: forge.Ptr("Task C")}}}
	D := &forge.Task{Name: "D", Template: "echo", Arguments: []forge.Parameter{{Name: "msg", Value: forge.Ptr("Task D")}}}
	A.Then(B)
	A.Then(C)
	B.Then(D)
	C.Then(D)
	dag.AddTasks(A, B, C, D)

	w := &forge.Workflow{
		GenerateName: "diamond-",
		Entrypoint:   "diamond",
		Templates:    []forge.Templatable{echoTpl, dag},
	}
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML err: %v", err)
	}
	if yaml == "" {
		t.Fatal("ToYAML returned empty string")
	}
}

// README §Conditional Logic (Coinflip) — lines 142-167.
func TestReadme_Coinflip(t *testing.T) {
	resetGlobalConfig(t)
	flip := &forge.Script{
		Name:    "flip-coin",
		Image:   "python:3.11-alpine",
		Command: []string{"python"},
		Source:  `import random; print("heads" if random.randint(0,1) == 0 else "tails")`,
	}
	heads := &forge.Container{
		Name: "heads", Image: "alpine:3.18",
		Command: []string{"echo"}, Args: []string{"it was heads"},
	}
	tails := &forge.Container{
		Name: "tails", Image: "alpine:3.18",
		Command: []string{"echo"}, Args: []string{"it was tails"},
	}
	dag := &forge.DAG{Name: "coinflip"}
	flipTask := &forge.Task{Name: "flip", Template: "flip-coin"}
	headsTask := &forge.Task{Name: "heads", Template: "heads", When: `{{tasks.flip.outputs.result}} == "heads"`}
	tailsTask := &forge.Task{Name: "tails", Template: "tails", When: `{{tasks.flip.outputs.result}} == "tails"`}
	flipTask.Then(headsTask)
	flipTask.Then(tailsTask)
	dag.AddTasks(flipTask, headsTask, tailsTask)
	w := &forge.Workflow{
		GenerateName: "coin-",
		Entrypoint:   "coinflip",
		Templates:    []forge.Templatable{flip, heads, tails, dag},
	}
	if _, err := w.ToYAML(); err != nil {
		t.Fatalf("ToYAML err: %v", err)
	}
}

// README §REST Client — lines 174-191 (val-002, completeness-readme-listworkflows-nil).
// Uses httptest so the snippet runs without an actual Argo server.
func TestReadme_RestClient(t *testing.T) {
	resetGlobalConfig(t)
	// Stand-in Argo server for the README snippet.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"metadata":{},"items":[]}`))
	}))
	t.Cleanup(server.Close)

	ctx := context.Background()
	svc := client.NewWorkflowsService(
		server.URL,
		"my-token",
		"default",
	)
	w := &forge.Workflow{Name: "test-wf", Entrypoint: "main", Templates: []forge.Templatable{
		&forge.Container{Name: "main", Image: "alpine:3.18", Command: []string{"echo"}},
	}}
	if _, err := svc.CreateWorkflow(ctx, w); err != nil {
		t.Errorf("CreateWorkflow err: %v", err)
	}
	if _, err := svc.ListWorkflows(ctx, ""); err != nil {
		t.Errorf("ListWorkflows err: %v", err)
	}
	if _, err := svc.LintWorkflow(ctx, w); err != nil {
		t.Errorf("LintWorkflow err: %v", err)
	}
}

// README §Expression Builder — lines 197-206 (completeness-readme-expr-eq-mismatch).
func TestReadme_ExpressionBuilder(t *testing.T) {
	resetGlobalConfig(t)
	ref := expr.Tasks("my-task").Attr("outputs.result")
	if ref.Tmpl() == "" {
		t.Fatal("ref.Tmpl() returned empty")
	}
	cond := expr.Steps("validate").Attr("outputs.result").Equals(expr.C("success"))
	if cond.String() == "" {
		t.Fatal("Equals returned empty Expr")
	}
}

// README §Default PodAffinity (lines 266-280) — verifies the workflow compiles
// AND that the default affinity is materialized.
func TestReadme_DefaultPodAffinity(t *testing.T) {
	resetGlobalConfig(t)
	w := &forge.Workflow{
		GenerateName: "build-",
		Entrypoint:   "main",
		VolumeClaimTemplates: []forge.PVCVolume{{
			BaseVolume:  forge.BaseVolume{Name: "scratch", MountPath: "/scratch"},
			Size:        "1Gi",
			AccessModes: []forge.AccessMode{forge.ReadWriteOnce},
		}},
		Templates: []forge.Templatable{
			&forge.Container{Name: "main", Image: "alpine:3.18"},
		},
	}
	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build err: %v", err)
	}
	if wf.Spec.Affinity == nil {
		t.Fatal("expected default Affinity to be injected for PVC workflow")
	}
}

// README opt-out (lines 290-296) — DisableDefaultAffinity suppresses injection.
func TestReadme_DisableDefaultAffinity(t *testing.T) {
	resetGlobalConfig(t)
	w := &forge.Workflow{
		Name:                   "fan-out",
		Entrypoint:             "main",
		DisableDefaultAffinity: true,
		VolumeClaimTemplates: []forge.PVCVolume{{
			BaseVolume:  forge.BaseVolume{Name: "scratch", MountPath: "/scratch"},
			Size:        "1Gi",
			AccessModes: []forge.AccessMode{forge.ReadWriteOnce},
		}},
		Templates: []forge.Templatable{
			&forge.Container{Name: "main", Image: "alpine:3.18"},
		},
	}
	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build err: %v", err)
	}
	if wf.Spec.Affinity != nil {
		t.Fatal("DisableDefaultAffinity should suppress injection")
	}
}

// EC-2 contract test: running two README examples back-to-back must not
// contaminate state. We exercise two examples and confirm config.GetGlobal()
// reflects the post-reset defaults each time.
func TestReadme_NoCrossContamination(t *testing.T) {
	// First example: deliberately mutate the singleton.
	t.Run("set-image-once", func(t *testing.T) {
		resetGlobalConfig(t)
		config.GetGlobal().SetImage("contaminated:1")
		if config.GetGlobal().GetImage() != "contaminated:1" {
			t.Fatal("setup failed")
		}
	})
	// Second example: ResetGlobalConfigForTest must restore python:3.11.
	t.Run("clean-state-after", func(t *testing.T) {
		resetGlobalConfig(t)
		if got := config.GetGlobal().GetImage(); got != "python:3.11" {
			t.Fatalf("cross-test contamination: GetImage = %q, want \"python:3.11\"", got)
		}
	})
}
