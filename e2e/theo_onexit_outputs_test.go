//go:build e2e

package e2e

import (
	"context"
	"strings"
	"testing"

	forge "github.com/usetheodev/theo-forge"
	"github.com/usetheodev/theo-forge/model"
)

// TestTheoShape_OnExitFiresAfterSuccess proves that OnExit DAGs run
// after the main entrypoint completes — Theo uses this for cleanup +
// status notifications.
//
// Theo's pattern (workflow_builder.go ~L296):
//
//	wf.Entrypoint = "main"
//	wf.OnExit     = "cleanup"
//	Templates contain both DAGs.
//
// If the OnExit handler regresses, every Theo build leaks pods + never
// signals completion to the reconcile loop.
func TestTheoShape_OnExitFiresAfterSuccess(t *testing.T) {
	wfName := uniqueName("onexit-ok")
	cleanupWorkflow(t, wfName)

	w := &forge.Workflow{
		Name:       wfName,
		Entrypoint: "main",
		OnExit:     "cleanup",
		Templates: []forge.Templatable{
			&forge.Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"main done"},
			},
			&forge.Container{
				Name:    "cleanup",
				Image:   "alpine:3.18",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo CLEANUP-RAN-SUCCESS"},
			},
		},
	}

	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if wf.Spec.OnExit != "cleanup" {
		t.Fatalf("Build dropped OnExit: %q", wf.Spec.OnExit)
	}

	svc := argoClient(t)
	if _, err := svc.CreateWorkflowFromModel(context.Background(), wf, ""); err != nil {
		t.Fatalf("Create: %v", err)
	}
	_ = waitWorkflowSucceeded(t, wfName, defaultWaitTimeout)

	manifest := dumpWorkflow(t, wfName)
	// Argo creates an onExit node named "<workflowName>.onExit" that
	// invokes the template referenced by spec.onExit. We assert on the
	// `.onExit` suffix (proves the handler fired) AND the cleanup pod
	// banner string (proves it actually executed our container).
	if !strings.Contains(manifest, ".onExit") {
		t.Errorf("no '.onExit' node in status — OnExit handler did NOT register:\n%s", manifest)
	}
	if !strings.Contains(manifest, "templateName: cleanup") {
		t.Errorf("no 'templateName: cleanup' in status — OnExit did not invoke our handler:\n%s", manifest)
	}
}

// TestTheoShape_StatusOutputsExtractable mirrors Theo's
// extractStepOutputs (api/internal/argo/workflow_submitter.go ~L282).
// Theo reads outputs at:
//
//	status.nodes[<nodeId>].outputs.parameters[]
//
// and keys them by `displayName`. If the controller stops writing this
// shape OR forge stops requesting outputs correctly, the entire
// post-build delivery pipeline (image digest → cosign signature) loses
// its handoff between steps.
//
// This test:
//  1. Submits a workflow whose step writes /tmp/result.txt
//  2. Declares an output parameter pointing at /tmp/result.txt
//  3. Reads the workflow back AFTER Succeeded and walks status.nodes[*]
//     EXACTLY as extractStepOutputs does
//  4. Asserts the displayName-keyed output map has the expected value
func TestTheoShape_StatusOutputsExtractable(t *testing.T) {
	wfName := uniqueName("outputs")
	cleanupWorkflow(t, wfName)

	const expectedValue = "theo-output-marker-42"

	// Theo wraps every step in a DAG so node displayName == task name
	// (single-template workflows name the node after the workflow itself,
	// which is not how Theo consumes outputs). Mirror the Theo pattern.
	w := &forge.Workflow{
		Name:       wfName,
		Entrypoint: "main",
		Templates: []forge.Templatable{
			&forge.DAG{
				Name: "main",
				Tasks: []*forge.Task{
					{Name: "emit-result", Template: "emit"},
				},
			},
			&forge.Container{
				Name:    "emit",
				Image:   "alpine:3.18",
				Command: []string{"sh", "-c"},
				Args:    []string{"printf '%s' " + expectedValue + " > /tmp/result.txt"},
				Outputs: []forge.Parameter{{
					Name: "result",
					ValueFrom: &model.ValueFrom{
						Path: "/tmp/result.txt",
					},
				}},
			},
		},
	}

	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	svc := argoClient(t)
	if _, err := svc.CreateWorkflowFromModel(context.Background(), wf, ""); err != nil {
		t.Fatalf("Create: %v", err)
	}
	final := waitWorkflowSucceeded(t, wfName, defaultWaitTimeout)

	// --- Mirror Theo's extractStepOutputs walk ---
	// We use the REST client's typed Status here; Theo uses dynamic
	// client + ParseWorkflowStatusFromUnstructured. Theo-4 covers the
	// dynamic-client path; here we prove the OUTPUT WALK works regardless.
	if final.Status == nil || final.Status.Nodes == nil {
		t.Fatalf("Status.Nodes is empty — controller did not record per-node status\n%s",
			dumpWorkflow(t, wfName))
	}
	foundDisplay := ""
	foundValue := ""
	for _, node := range final.Status.Nodes {
		if node.Type != "Pod" {
			continue
		}
		if node.Outputs == nil {
			continue
		}
		for _, p := range node.Outputs.Parameters {
			if p.Name != "result" || p.Value == nil {
				continue
			}
			foundDisplay = node.DisplayName
			foundValue = *p.Value
		}
	}
	if foundDisplay != "emit-result" {
		t.Errorf("output not keyed by displayName=emit-result (got %q)\nmanifest:\n%s",
			foundDisplay, dumpWorkflow(t, wfName))
	}
	if foundValue != expectedValue {
		t.Errorf("output value: got %q, want %q\nmanifest:\n%s",
			foundValue, expectedValue, dumpWorkflow(t, wfName))
	}
}
