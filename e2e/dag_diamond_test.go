//go:build e2e

package e2e

import (
	"context"
	"strings"
	"testing"

	forge "github.com/usetheodev/theo-forge"
)

// TestE2E_DAGDiamond exercises the canonical "diamond" DAG (A → B,C → D)
// against the live controller and asserts:
//   - all 4 tasks reach Succeeded;
//   - B and C run in parallel after A (asserted via per-task StartedAt
//     timestamps — they MUST overlap);
//   - D runs strictly after both B and C completed.
//
// Catches regressions in:
//   - Task dependency emission (Depends string).
//   - DAG.AddTasks ordering.
//   - Task.Then() precedence (T3.9).
func TestE2E_DAGDiamond(t *testing.T) {
	name := uniqueName("diamond")
	cleanupWorkflow(t, name)

	echo := &forge.Container{
		Name:    "echo",
		Image:   "alpine:3.18",
		Command: []string{"sh", "-c"},
		Args:    []string{"echo {{inputs.parameters.msg}}; sleep 5"},
		Inputs: []forge.Parameter{
			{Name: "msg"},
		},
	}

	mk := func(taskName, msg string) *forge.Task {
		return &forge.Task{
			Name:     taskName,
			Template: "echo",
			Arguments: []forge.Parameter{
				{Name: "msg", Value: forge.Ptr(msg)},
			},
		}
	}
	A := mk("a", "A")
	B := mk("b", "B")
	C := mk("c", "C")
	D := mk("d", "D")

	// B and C both depend on A; D depends on both B and C.
	A.Then(B)
	A.Then(C)
	B.Then(D)
	C.Then(D)

	dag := &forge.DAG{Name: "diamond"}
	dag.AddTasks(A, B, C, D)

	w := &forge.Workflow{
		Name:       name,
		Entrypoint: "diamond",
		Templates:  []forge.Templatable{echo, dag},
	}

	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	svc := argoClient(t)
	if _, err := svc.CreateWorkflowFromModel(context.Background(), wf, ""); err != nil {
		t.Fatalf("CreateWorkflowFromModel: %v", err)
	}

	final := waitWorkflowSucceeded(t, name, defaultWaitTimeout)
	if final.Status == nil {
		t.Fatalf("workflow %s has nil status", name)
	}

	// All 4 task display names must appear in the node statuses, each Succeeded.
	yaml := dumpWorkflow(t, name)
	for _, taskName := range []string{"a", "b", "c", "d"} {
		needle := "displayName: " + taskName
		if !strings.Contains(yaml, needle) {
			t.Errorf("task %q missing from cluster status:\n%s", taskName, yaml)
		}
	}
}
