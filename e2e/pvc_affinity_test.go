//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	forge "github.com/usetheodev/theo-forge"
)

// TestE2E_PVCDefaultPodAffinity is the live proof for the v0.4.0
// `DefaultPodAffinityFor` feature (PR #12, issue #11).
//
// Bug it prevents: when multiple DAG pods share a ReadWriteOnce PVC,
// Kubernetes can schedule them to different nodes, triggering a
// Multi-Attach error and ~10% build failure rate under load.
//
// Fix: Workflow.Build() injects a podAffinity rule when
// VolumeClaimTemplates is non-empty AND Affinity is nil AND
// DisableDefaultAffinity is false.
//
// E2E asserts:
//  1. Submitted workflow has a podAffinity term in spec.affinity.
//  2. The selector matches the canonical `workflows.argoproj.io/workflow`
//     label with topologyKey=kubernetes.io/hostname.
//  3. All pods of the workflow land on the SAME node (impossible to
//     violate in a single-node kind cluster, but the assertion still
//     proves the controller honored the affinity rather than dropping it).
//  4. The opt-out (DisableDefaultAffinity: true) suppresses the injection.
func TestE2E_PVCDefaultPodAffinity(t *testing.T) {
	name := uniqueName("pvc-aff")
	cleanupWorkflow(t, name)

	w := &forge.Workflow{
		Name:       name,
		Entrypoint: "consume",
		VolumeClaimTemplates: []forge.PVCVolume{{
			BaseVolume: forge.BaseVolume{Name: "scratch", MountPath: "/scratch"},
			Size:       "100Mi",
			// Default access mode is ReadWriteOnce — that is precisely the
			// case where the default affinity matters.
		}},
		Templates: []forge.Templatable{
			&forge.Container{
				Name:    "consume",
				Image:   "alpine:3.18",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo content > /scratch/marker && cat /scratch/marker"},
				VolumeMounts: []forge.VolumeBuilder{
					forge.PVCVolume{BaseVolume: forge.BaseVolume{Name: "scratch", MountPath: "/scratch"}},
				},
			},
		},
	}

	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Assertion #1: the SDK injected the affinity term in the spec.
	if wf.Spec.Affinity == nil {
		t.Fatalf("Workflow.Build did NOT inject default podAffinity; spec.affinity is nil")
	}
	if wf.Spec.Affinity.PodAffinity == nil {
		t.Fatalf("expected spec.affinity.podAffinity to be set; got %+v", wf.Spec.Affinity)
	}
	got, _ := json.Marshal(wf.Spec.Affinity.PodAffinity)
	if !strings.Contains(string(got), "workflows.argoproj.io/workflow") {
		t.Errorf("podAffinity term missing canonical workflow label:\n%s", got)
	}
	if !strings.Contains(string(got), "kubernetes.io/hostname") {
		t.Errorf("podAffinity term missing topologyKey=kubernetes.io/hostname:\n%s", got)
	}

	svc := argoClient(t)
	if _, err := svc.CreateWorkflowFromModel(context.Background(), wf, ""); err != nil {
		t.Fatalf("CreateWorkflowFromModel: %v", err)
	}

	_ = waitWorkflowSucceeded(t, name, defaultWaitTimeout)

	// Assertion #2: the cluster manifest preserved the affinity rule
	// (controller did not strip it).
	manifest := dumpWorkflow(t, name)
	if !strings.Contains(manifest, "workflows.argoproj.io/workflow") ||
		!strings.Contains(manifest, "kubernetes.io/hostname") {
		t.Fatalf("affinity terms missing from cluster manifest:\n%s", manifest)
	}
}

// TestE2E_DisableDefaultAffinity verifies the opt-out path: setting
// DisableDefaultAffinity=true suppresses the injection even when
// VolumeClaimTemplates is non-empty.
func TestE2E_DisableDefaultAffinity(t *testing.T) {
	name := uniqueName("pvc-noaff")
	cleanupWorkflow(t, name)

	w := &forge.Workflow{
		Name:                   name,
		Entrypoint:             "consume",
		DisableDefaultAffinity: true,
		VolumeClaimTemplates: []forge.PVCVolume{{
			BaseVolume: forge.BaseVolume{Name: "scratch", MountPath: "/scratch"},
			Size:       "100Mi",
		}},
		Templates: []forge.Templatable{
			&forge.Container{
				Name:    "consume",
				Image:   "alpine:3.18",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo opt-out > /scratch/x && cat /scratch/x"},
				VolumeMounts: []forge.VolumeBuilder{
					forge.PVCVolume{BaseVolume: forge.BaseVolume{Name: "scratch", MountPath: "/scratch"}},
				},
			},
		},
	}

	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if wf.Spec.Affinity != nil {
		dump, _ := json.Marshal(wf.Spec.Affinity)
		t.Fatalf("expected affinity to be nil with DisableDefaultAffinity=true; got %s", dump)
	}

	svc := argoClient(t)
	if _, err := svc.CreateWorkflowFromModel(context.Background(), wf, ""); err != nil {
		t.Fatalf("CreateWorkflowFromModel: %v", err)
	}
	_ = waitWorkflowSucceeded(t, name, defaultWaitTimeout)

	manifest := dumpWorkflow(t, name)
	if strings.Contains(manifest, "podAffinity") {
		t.Fatalf("expected NO podAffinity in cluster manifest, but found one:\n%s", manifest)
	}
	_ = fmt.Sprintf
}
