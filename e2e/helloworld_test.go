//go:build e2e

package e2e

import (
	"context"
	"testing"

	forge "github.com/usetheodev/theo-forge"
)

// TestE2E_HelloWorld submits the README quickstart workflow against
// the live Argo control plane and waits for it to reach Succeeded.
//
// This is the single most important E2E test — it proves the SDK's
// happy path round-trips through:
//
//	Workflow{} → Build() → ToYAML → CreateWorkflowFromModel
//	         → workflow-controller → pod scheduled → echo → Succeeded
//
// If this regresses, the README is wrong and every consumer feels it.
func TestE2E_HelloWorld(t *testing.T) {
	name := uniqueName("hello")
	cleanupWorkflow(t, name)

	w := &forge.Workflow{
		Name:       name,
		Entrypoint: "main",
		Templates: []forge.Templatable{
			&forge.Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"hello from theo-forge e2e"},
			},
		},
	}

	svc := argoClient(t)
	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	created, err := svc.CreateWorkflowFromModel(context.Background(), wf, "")
	if err != nil {
		t.Fatalf("CreateWorkflowFromModel: %v", err)
	}
	if created.Metadata.Name == "" {
		t.Fatalf("server returned empty workflow name")
	}

	final := waitWorkflowSucceeded(t, name, defaultWaitTimeout)
	if final.Status == nil || final.Status.Phase != "Succeeded" {
		t.Fatalf("expected Succeeded phase, got %#v\nmanifest:\n%s",
			final.Status, dumpWorkflow(t, name))
	}
}
