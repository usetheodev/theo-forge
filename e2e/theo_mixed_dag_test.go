//go:build e2e

package e2e

import (
	"context"
	"strings"
	"testing"

	forge "github.com/usetheodev/theo-forge"
	"github.com/usetheodev/theo-forge/model"
)

// TestTheoShape_MixedDAG_InlineAndCWT proves the architectural pattern
// that drives the ENTIRE Theo build pipeline: a DAG that mixes
//
//   - tasks calling reusable steps via TemplateRef{ClusterScope:true}
//     (clone, cache, dockerfile-lint, syft, cosign, harbor-mirror)
//   - tasks calling inline templates by Name
//     (install, build, buildkit-per-app)
//
// If forge stops emitting `clusterScope: true` or `templateRef:` correctly,
// EVERY Theo build fails. This is the most important Theo-shape test.
//
// Test flow:
//  1. Create a ClusterWorkflowTemplate in the cluster (mimics Theo's
//     `clone` CWT) — exposes a template named "say".
//  2. Build a forge.Workflow whose DAG has:
//     - Task 'first'  → TemplateRef{Name: "<cwt>", Template: "say", ClusterScope: true}
//     - Task 'second' → Template: "inline-echo" (inline forge.Container)
//  3. Submit, wait Succeeded, dump manifest.
//  4. Assert the cluster's view contains `clusterScope: true` AND the
//     inline template name on the right task.
func TestTheoShape_MixedDAG_InlineAndCWT(t *testing.T) {
	wfName := uniqueName("theo-mixed")
	cleanupWorkflow(t, wfName)

	cwtName := "theo-e2e-shared-" + strings.Split(wfName, "-")[2] // unique cwt name
	// Cleanup the CWT too (it's cluster-scoped, not workflow-scoped).
	t.Cleanup(func() {
		kubectl(t, "delete", "clusterworkflowtemplate", cwtName, "--ignore-not-found")
	})

	// 1. Author + install the CWT as a typed forge.ClusterWorkflowTemplate.
	cwt := &forge.ClusterWorkflowTemplate{
		Name: cwtName,
		Templates: []forge.Templatable{
			&forge.Container{
				Name:    "say",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"hello from CWT"},
			},
		},
	}
	cwtYAML, err := cwt.ToYAML()
	if err != nil {
		t.Fatalf("cwt ToYAML: %v", err)
	}
	writeWorkflowYAML(t, cwtName, cwtYAML)
	// `kubectl apply` above is synchronous; no need to wait — CWT
	// does not expose a `Available` condition. Confirm presence:
	kubectl(t, "get", "clusterworkflowtemplate", cwtName)

	// 2. Build the mixed DAG workflow.
	first := &forge.Task{
		Name: "first",
		TemplateRef: &model.TemplateRef{
			Name:         cwtName,
			Template:     "say",
			ClusterScope: true,
		},
	}
	second := &forge.Task{
		Name:         "second",
		Template:     "inline-echo",
		Dependencies: []string{"first"},
	}
	dag := &forge.DAG{Name: "main"}
	dag.AddTasks(first, second)

	w := &forge.Workflow{
		Name:       wfName,
		Entrypoint: "main",
		Templates: []forge.Templatable{
			dag,
			&forge.Container{
				Name:    "inline-echo",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"hello from inline"},
			},
		},
	}

	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	// In-Go assert: the CWT ref must be serialized with ClusterScope=true.
	if wf.Spec.Templates[0].DAG == nil || len(wf.Spec.Templates[0].DAG.Tasks) != 2 {
		t.Fatalf("DAG template malformed: %+v", wf.Spec.Templates[0].DAG)
	}
	tref := wf.Spec.Templates[0].DAG.Tasks[0].TemplateRef
	if tref == nil || tref.Name != cwtName || tref.Template != "say" || !tref.ClusterScope {
		t.Fatalf("first task TemplateRef wrong: %+v", tref)
	}

	// 3. Submit and wait.
	svc := argoClient(t)
	if _, err := svc.CreateWorkflowFromModel(context.Background(), wf, ""); err != nil {
		t.Fatalf("Create: %v", err)
	}
	_ = waitWorkflowSucceeded(t, wfName, defaultWaitTimeout)

	// 4. Cluster-side: confirm the manifest still says clusterScope and
	//    the CWT name. This is what Theo's controller observes.
	manifest := dumpWorkflow(t, wfName)
	mustContain := []string{
		"clusterScope: true",
		"name: " + cwtName,
		"template: say",
		"template: inline-echo",
	}
	for _, needle := range mustContain {
		if !strings.Contains(manifest, needle) {
			t.Errorf("cluster manifest missing %q — Theo mixed-DAG regression\n%s",
				needle, manifest)
		}
	}
}
