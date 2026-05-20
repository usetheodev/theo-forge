//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	forge "github.com/usetheodev/theo-forge"
	"github.com/usetheodev/theo-forge/model"
)

// TestTheoShape_DynamicClientPolling exercises the EXACT polling path
// Theo's reconcile loop uses (api/internal/argo/workflow_submitter.go
// GetStatus, ~L230):
//
//	wf, _ := dynamic.Resource(workflowGVR).Namespace(ns).Get(...)
//	status, _ := forge.ParseWorkflowStatusFromUnstructured(wf.Object)
//	if status.Phase == "Failed" && forge.AllPodNodesExitedZero(status) {
//	    phase = "Succeeded" // sidecar false-failure override
//	}
//
// We cannot import k8s.io/client-go here (depguard policy in
// .golangci.yml — forge intentionally avoids the k8s.io transitive
// tree). Instead, we drive a `kubectl get -o json` through the shim,
// unmarshal to map[string]interface{} (which is EXACTLY what the
// dynamic client returns: `unstructured.Unstructured.Object`), and
// pass it through the same forge helpers Theo calls.
//
// If ParseWorkflowStatusFromUnstructured stops parsing the controller's
// real output OR AllPodNodesExitedZero changes shape, this catches it.
func TestTheoShape_DynamicClientPolling(t *testing.T) {
	wfName := uniqueName("dyn-poll")
	cleanupWorkflow(t, wfName)

	w := &forge.Workflow{
		Name:       wfName,
		Entrypoint: "main",
		Templates: []forge.Templatable{
			&forge.Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo dynamic-poll; exit 0"},
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

	// --- Phase 1: poll via the SHIM (mirrors Theo's reconcile loop) ---
	//
	// Wait until the workflow reaches a terminal phase using ONLY the
	// dynamic-client style helpers from forge. No REST client involved
	// here — that is the point.
	var lastStatus *model.WorkflowStatusDetail
	deadline := time.Now().Add(defaultWaitTimeout)
	for time.Now().Before(deadline) {
		raw := kubectl(t, "-n", "argo", "get", "wf", wfName, "-o", "json")
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			t.Fatalf("unmarshal kubectl json: %v", err)
		}
		status, err := forge.ParseWorkflowStatusFromUnstructured(obj)
		if err != nil {
			t.Fatalf("ParseWorkflowStatusFromUnstructured: %v", err)
		}
		if status != nil {
			lastStatus = status
			if status.Phase == model.WorkflowSucceeded ||
				status.Phase == model.WorkflowFailed ||
				status.Phase == model.WorkflowError ||
				status.Phase == model.WorkflowTerminated {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	if lastStatus == nil {
		t.Fatalf("polling never produced a non-nil status\n%s", dumpWorkflow(t, wfName))
	}

	// --- Phase 2: assert the parse extracted the fields Theo relies on ---
	if lastStatus.Phase != model.WorkflowSucceeded {
		t.Fatalf("expected Succeeded phase, got %q\n%s",
			lastStatus.Phase, dumpWorkflow(t, wfName))
	}
	if lastStatus.FinishedAt == "" {
		t.Errorf("FinishedAt empty — Theo's `status.FinishedAt` will be missing")
	}
	if len(lastStatus.Nodes) == 0 {
		t.Errorf("Nodes empty — Theo's extractStepOutputs will return nothing")
	}

	// --- Phase 3: AllPodNodesExitedZero on a successful workflow ---
	// MUST return true (every pod main container exited 0).
	if !forge.AllPodNodesExitedZero(lastStatus) {
		t.Errorf("AllPodNodesExitedZero(succeeded workflow) returned false — sidecar override path broken")
	}
}

// TestTheoShape_AllPodNodesExitedZero_OverrideOnFailedWorkflow targets
// the sidecar false-failure defense in Theo (workflow_submitter.go ~L258):
//
//	if phase == "Failed" && forge.AllPodNodesExitedZero(status) {
//	    phase = "Succeeded"
//	}
//
// We construct a synthetic Failed status (because reliably reproducing
// the original sidecar false-fail bug requires a custom sidecar image)
// and feed it through the helper. This is a UNIT-style assertion
// running inside the e2e package because the helper signature mirrors
// the dynamic-client path and we want it co-located with the live
// polling test for traceability.
func TestTheoShape_AllPodNodesExitedZero_OverrideOnFailedWorkflow(t *testing.T) {
	// Synthetic Failed workflow status: main container exited 0,
	// some sidecar exited non-zero, Argo reported "Failed" overall.
	statusJSON := `{
		"phase": "Failed",
		"nodes": {
			"node-A": {
				"id": "node-A",
				"displayName": "main",
				"type": "Pod",
				"phase": "Succeeded",
				"outputs": {
					"exitCode": "0"
				}
			}
		}
	}`
	var rawObj map[string]interface{}
	if err := json.Unmarshal([]byte(`{"status":`+statusJSON+`}`), &rawObj); err != nil {
		t.Fatalf("unmarshal synthetic: %v", err)
	}
	status, err := forge.ParseWorkflowStatusFromUnstructured(rawObj)
	if err != nil {
		t.Fatalf("ParseWorkflowStatusFromUnstructured: %v", err)
	}
	if status == nil {
		t.Fatal("status nil for non-empty input")
	}
	if status.Phase != model.WorkflowFailed {
		t.Errorf("Phase parse lost: got %q", status.Phase)
	}
	if !forge.AllPodNodesExitedZero(status) {
		t.Errorf("AllPodNodesExitedZero should be true when every Pod node exited 0; this is the override Theo relies on")
	}
}
