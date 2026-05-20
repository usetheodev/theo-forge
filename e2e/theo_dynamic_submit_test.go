//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	forge "github.com/usetheodev/theo-forge"
	"github.com/usetheodev/theo-forge/model"
)

// TestTheoShape_ToDictDynamicSubmitPath exercises the EXACT submission
// path Theo uses in production (api/internal/argo/workflow_submitter.go
// ~L140 Submit):
//
//	wfObject, _ := wf.ToDict()
//	workflow := &unstructured.Unstructured{Object: wfObject}
//	created, _ := dynamic.Resource(workflowGVR).Namespace(ns).Create(...)
//
// Our previous E2E tests submit via CreateWorkflowFromModel (REST). If
// forge.Workflow.ToDict() ever stops producing a map[string]interface{}
// that the apiserver accepts as an Unstructured Workflow CR, EVERY
// Theo build breaks but our REST-only tests pass green. This test
// closes that gap.
//
// We can't import k8s.io/client-go/dynamic here (depguard policy), so
// we reproduce the path with `kubectl apply -f` of the JSON we get
// out of ToDict. The apiserver validation is identical: SSA accepts
// the same shape Unstructured.Create posts.
func TestTheoShape_ToDictDynamicSubmitPath(t *testing.T) {
	wfName := uniqueName("todict")
	cleanupWorkflow(t, wfName)

	w := &forge.Workflow{
		Name:       wfName,
		Entrypoint: "main",
		Templates: []forge.Templatable{
			&forge.DAG{
				Name:  "main",
				Tasks: []*forge.Task{{Name: "step", Template: "echo"}},
			},
			&forge.Container{
				Name:    "echo",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"submitted via ToDict + dynamic"},
			},
		},
	}

	// Theo's exact path: ToDict → marshal to JSON → apply via dynamic client.
	// We marshal to JSON and `kubectl apply -f` which is byte-equivalent.
	wfDict, err := w.ToDict()
	if err != nil {
		t.Fatalf("ToDict: %v", err)
	}
	raw, err := json.Marshal(wfDict)
	if err != nil {
		t.Fatalf("Marshal dict: %v", err)
	}

	// Sanity: the dict MUST have apiVersion + kind at the top — otherwise
	// the apiserver rejects with "no kind registered".
	if _, ok := wfDict["apiVersion"]; !ok {
		t.Fatalf("ToDict output missing 'apiVersion': %s", raw)
	}
	if k, _ := wfDict["kind"].(string); k != "Workflow" {
		t.Fatalf("ToDict output kind = %q, want Workflow:\n%s", k, raw)
	}

	writeWorkflowYAML(t, wfName, string(raw))
	_ = waitWorkflowSucceeded(t, wfName, defaultWaitTimeout)
}

// TestTheoShape_ConcurrentSubmits proves the SDK is safe under the
// concurrency Theo Cloud hits in production. Theo's BuildReconcileLoop
// + parallel deploy events can submit N workflows simultaneously
// against the SAME WorkflowsService instance. Any race in:
//
//   - Workflow.Build() against shared *config.GlobalConfig
//   - the shared http.Client / TLS handshake state
//   - argoClient(t) re-entry
//
// would surface as flaky CreateWorkflowFromModel errors here.
//
// We submit N workflows in parallel goroutines, wait for ALL to
// reach Succeeded, and assert every error path is empty.
func TestTheoShape_ConcurrentSubmits(t *testing.T) {
	const n = 5 // keep modest — every workflow pulls alpine + starts a pod
	svc := argoClient(t)

	names := make([]string, n)
	for i := range names {
		names[i] = uniqueName(fmt.Sprintf("cc-%d", i))
		cleanupWorkflow(t, names[i])
	}

	var wg sync.WaitGroup
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			w := &forge.Workflow{
				Name:       names[idx],
				Entrypoint: "main",
				Templates: []forge.Templatable{
					&forge.Container{
						Name:    "main",
						Image:   "alpine:3.18",
						Command: []string{"echo"},
						Args:    []string{fmt.Sprintf("concurrent #%d", idx)},
					},
				},
			}
			wf, err := w.Build()
			if err != nil {
				errCh <- fmt.Errorf("worker %d Build: %w", idx, err)
				return
			}
			if _, err := svc.CreateWorkflowFromModel(context.Background(), wf, ""); err != nil {
				errCh <- fmt.Errorf("worker %d Create: %w", idx, err)
				return
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Errorf("concurrent submit error: %v", err)
	}
	if t.Failed() {
		return
	}

	// All accepted — wait every one to Succeed.
	for _, name := range names {
		_ = waitWorkflowSucceeded(t, name, defaultWaitTimeout)
	}
}

// TestTheoShape_ShutdownPatch_Cancel replicates Theo's Cancel path
// (api/internal/argo/workflow_submitter.go ~L187 Cancel):
//
//	patch := {"apiVersion":"argoproj.io/v1alpha1","kind":"Workflow",
//	          "spec":{"shutdown":"Stop"}}
//	client.Resource(workflowGVR).Namespace(ns).Patch(
//	    ctx, workflowID, "application/merge-patch+json", patch, ...)
//
// Theo issues this merge-patch directly via dynamic client. The forge
// REST client has StopWorkflow which is functionally equivalent but
// hits a DIFFERENT server endpoint (/api/v1/workflows/.../stop vs raw
// CRD patch). We test the raw-patch path here.
//
// Submit a slow workflow → patch shutdown=Stop → wait terminal phase.
// The workflow MUST end in a non-Succeeded terminal phase (Failed
// or Terminated) within the test timeout.
func TestTheoShape_ShutdownPatch_Cancel(t *testing.T) {
	wfName := uniqueName("cancel")
	cleanupWorkflow(t, wfName)

	w := &forge.Workflow{
		Name:       wfName,
		Entrypoint: "main",
		Templates: []forge.Templatable{
			// Sleep 5 min so we have time to cancel it before it
			// could possibly succeed on its own.
			&forge.Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"sleep"},
				Args:    []string{"300"},
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

	// Wait until workflow is actually Running before we cancel.
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		current, err := svc.GetWorkflow(ctx, wfName, "")
		cancel()
		if err == nil && current.Status != nil && current.Status.Phase == model.WorkflowRunning {
			break
		}
		time.Sleep(2 * time.Second)
	}

	// Patch shutdown=Stop EXACTLY as Theo does.
	patchJSON := `{"apiVersion":"argoproj.io/v1alpha1","kind":"Workflow","spec":{"shutdown":"Stop"}}`
	kubectl(t, "-n", "argo", "patch", "wf", wfName,
		"--type=merge", "-p", patchJSON)

	// Wait for terminal — MUST not be Succeeded.
	deadline = time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		final, err := svc.GetWorkflow(ctx, wfName, "")
		cancel()
		if err == nil && final.Status != nil {
			switch final.Status.Phase {
			case model.WorkflowFailed, model.WorkflowError, model.WorkflowTerminated:
				return // success — cancellation worked
			case model.WorkflowSucceeded:
				t.Fatalf("workflow Succeeded — cancel did not take effect (we sleep 300s, max test is 60s)")
			}
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("workflow %s never reached terminal phase after shutdown patch\n%s",
		wfName, dumpWorkflow(t, wfName))
}

// TestTheoShape_BackwardCompat_v3DAG_OnV4 is a compat probe: a workflow
// shape that was authored against Argo v3 (the version forge originally
// pinned) MUST still apply cleanly on Argo v4 (Theo Cloud prod).
//
// We construct a workflow that uses ONLY fields present in v3.x spec
// (no v4-only additions) and assert v4 accepts + runs it. This guards
// against the SDK silently emitting v4-only field names that would
// break consumers still on v3.
func TestTheoShape_BackwardCompat_v3DAG_OnV4(t *testing.T) {
	wfName := uniqueName("v3-on-v4")
	cleanupWorkflow(t, wfName)

	// All fields below existed in Argo v3.5; this same workflow must
	// run on v4.0.3 without any changes.
	a := &forge.Task{Name: "a", Template: "echo", Arguments: []forge.Parameter{
		{Name: "msg", Value: forge.Ptr("A")},
	}}
	b := &forge.Task{Name: "b", Template: "echo", Dependencies: []string{"a"},
		Arguments: []forge.Parameter{{Name: "msg", Value: forge.Ptr("B")}}}

	dag := &forge.DAG{Name: "main"}
	dag.AddTasks(a, b)

	w := &forge.Workflow{
		Name:       wfName,
		Entrypoint: "main",
		Templates: []forge.Templatable{
			dag,
			&forge.Container{
				Name:    "echo",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"{{inputs.parameters.msg}}"},
				Inputs:  []forge.Parameter{{Name: "msg"}},
			},
		},
	}
	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	svc := argoClient(t)
	if _, err := svc.CreateWorkflowFromModel(context.Background(), wf, ""); err != nil {
		t.Fatalf("Create: %v — Argo v4 rejected v3-shape workflow", err)
	}
	final := waitWorkflowSucceeded(t, wfName, defaultWaitTimeout)
	if final.Status == nil || final.Status.Phase != model.WorkflowSucceeded {
		t.Errorf("v3-shape did not succeed on v4: %#v", final.Status)
	}
	// Make sure manifest reflects what we asked for (no silent rewrite).
	manifest := dumpWorkflow(t, wfName)
	if !strings.Contains(manifest, "name: a") || !strings.Contains(manifest, "name: b") {
		t.Errorf("controller dropped task names from v3-shape DAG:\n%s", manifest)
	}
}
