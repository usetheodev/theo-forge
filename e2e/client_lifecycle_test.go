//go:build e2e

package e2e

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	forge "github.com/usetheodev/theo-forge"
	"github.com/usetheodev/theo-forge/model"
)

// TestE2E_ClientLifecycle exercises the REST client end-to-end against
// the real argo-server. The flow drives every method that mutates or
// observes cluster state:
//
//	Lint → Create → Get → List → Suspend → Resume → Stop → Get → Delete → Get (404)
//
// The chosen workflow uses the suspend template so we have a stable
// "in-progress" phase to suspend/resume against.
func TestE2E_ClientLifecycle(t *testing.T) {
	name := uniqueName("lifecycle")
	cleanupWorkflow(t, name)

	svc := argoClient(t)

	w := &forge.Workflow{
		Name:       name,
		Entrypoint: "main",
		Templates: []forge.Templatable{
			&forge.DAG{
				Name: "main",
				Tasks: []*forge.Task{
					{Name: "pause", Template: "suspend"},
					{Name: "done", Template: "done", Depends: "pause"},
				},
			},
			&forge.Suspend{Name: "suspend", Duration: "30s"},
			&forge.Container{
				Name:    "done",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"resumed"},
			},
		},
	}

	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	ctx := context.Background()

	// 1. Lint — server checks schema before we submit.
	if _, err := svc.LintWorkflow(ctx, w); err != nil {
		t.Fatalf("Lint: %v", err)
	}

	// 2. Create.
	created, err := svc.CreateWorkflowFromModel(ctx, wf, "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Metadata.Name == "" {
		t.Fatalf("server returned empty name")
	}

	// 3. Get — confirm the server stored it under the expected name.
	got, err := svc.GetWorkflow(ctx, name, "")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Metadata.Name != name {
		t.Fatalf("Get returned name=%q, want %q", got.Metadata.Name, name)
	}

	// 4. List — the new workflow must show up.
	list, err := svc.ListWorkflows(ctx, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if !containsWorkflowName(list, name) {
		t.Fatalf("List did not include %q\nfound %d workflows", name, len(list))
	}

	// 5. Wait until the suspend task is actually running, otherwise
	//    Suspend/Resume below are no-ops.
	waitForPhase(t, name, model.WorkflowRunning, 60*time.Second)

	// 6. Suspend — workflow stays Running but is paused.
	if err := svc.SuspendWorkflow(ctx, name, ""); err != nil {
		t.Fatalf("Suspend: %v", err)
	}

	// 7. Resume.
	if err := svc.ResumeWorkflow(ctx, name, ""); err != nil {
		t.Fatalf("Resume: %v", err)
	}

	// 8. Stop — gracefully end the workflow (runs onExit handlers).
	if err := svc.StopWorkflow(ctx, name, ""); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// Wait for any terminal phase (Stop leaves Succeeded/Failed/Terminated/Error).
	waitForTerminal(t, name, 90*time.Second)

	// 9. Delete — server must remove it.
	if err := svc.DeleteWorkflow(ctx, name, ""); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// 10. Get after Delete MUST fail (404 surfaces as APIError; the client
	//     wraps it in an error containing "404" or "NotFound").
	_, err = svc.GetWorkflow(ctx, name, "")
	if err == nil {
		t.Fatalf("Get after Delete returned no error; expected 404")
	}
}

func containsWorkflowName(list []model.WorkflowModel, name string) bool {
	for _, w := range list {
		if w.Metadata.Name == name {
			return true
		}
	}
	return false
}

// waitForPhase polls until the workflow reaches the desired phase
// (or any terminal phase). Used to synchronize before Suspend/Resume.
func waitForPhase(t *testing.T, name string, want model.WorkflowStatus, timeout time.Duration) {
	t.Helper()
	svc := argoClient(t)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		wf, err := svc.GetWorkflow(ctx, name, "")
		cancel()
		if err == nil && wf.Status != nil {
			if wf.Status.Phase == want {
				return
			}
			// If it already finished, stop polling — caller can decide.
			switch wf.Status.Phase {
			case model.WorkflowSucceeded, model.WorkflowFailed,
				model.WorkflowError, model.WorkflowTerminated:
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("workflow %s did not reach phase %s in %s", name, want, timeout)
}

// waitForTerminal blocks until the workflow stops moving.
func waitForTerminal(t *testing.T, name string, timeout time.Duration) {
	t.Helper()
	svc := argoClient(t)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		wf, err := svc.GetWorkflow(ctx, name, "")
		cancel()
		if err == nil && wf.Status != nil {
			switch wf.Status.Phase {
			case model.WorkflowSucceeded, model.WorkflowFailed,
				model.WorkflowError, model.WorkflowTerminated:
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("workflow %s did not reach a terminal phase in %s", name, timeout)
}

// TestE2E_ClientNameValidation is the live counterpart of T2.3 / SEC-003.
// The client MUST reject malformed names BEFORE building a request URL.
func TestE2E_ClientNameValidation(t *testing.T) {
	svc := argoClient(t)
	ctx := context.Background()

	bad := []string{
		"../escape",
		"name with space",
		"UPPERCASE",
		strings.Repeat("a", 254), // >253 chars
	}
	for _, n := range bad {
		_, err := svc.GetWorkflow(ctx, n, "")
		if !errors.Is(err, model.ErrInvalidName) {
			t.Errorf("GetWorkflow(%q): got %v, want ErrInvalidName", n, err)
		}
	}
}
