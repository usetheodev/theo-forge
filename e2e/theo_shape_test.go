//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	forge "github.com/usetheodev/theo-forge"
	"github.com/usetheodev/theo-forge/model"
)

// --------------------------------------------------------------------------
// Theo-shape coverage — these tests mirror EXACTLY how /home/paulo/theo
// (api/internal/argo/workflow_builder.go) constructs Workflows.
//
// If any of these regress, Theo Cloud builds break in production. They are
// the executable contract between forge and Theo.
// --------------------------------------------------------------------------

// TestTheoShape_FullProductionWorkflow constructs a workflow that uses
// every Theo-specific field at once and verifies each survives Build →
// CreateWorkflow → controller normalization → cluster-side YAML.
//
// Theo-shape inventory (from api/internal/argo/workflow_builder.go ~L221):
//
//	GenerateName             "theo-build-{BuildID}-"
//	Entrypoint, OnExit       (two DAGs)
//	ServiceAccountName       "theo-build"
//	Parallelism              *int
//	TTLStrategy              {AfterCompletion, AfterSuccess, AfterFailure}
//	PodGC                    {Strategy:"OnPodCompletion", DeleteDelayDuration:"5m"}
//	Labels                   {tenant, project, environment, build}
//	Annotations              {"theo.io/correlation-id": ...}
//	VolumeClaimTemplates     [{Name, StorageClassName:"local-path", Size}]
//	Volumes                  [SecretVolume, ConfigMapVolume]
//	SecurityContext          {FSGroup, RunAsUser, RunAsGroup} all 1000
//
// Plus the default podAffinity injected by Build() because
// VolumeClaimTemplates is non-empty (v0.4.0 / PR #12).
func TestTheoShape_FullProductionWorkflow(t *testing.T) {
	name := uniqueName("theo-shape")
	cleanupWorkflow(t, name)

	const (
		buildID       = "build-abc123"
		tenantID      = "tenant-acme"
		projectID     = "proj-saas"
		environment   = "production"
		correlationID = "corr-xyz789"
		uid           = int64(1000)
	)

	w := &forge.Workflow{
		// We use Name (not GenerateName) so we can predict it for cleanup,
		// but the FIELD set + serialization MUST be identical to Theo's
		// GenerateName path. Theo's actual GenerateName usage is covered
		// by TestTheoShape_GenerateName below.
		Name:               name,
		Entrypoint:         "main",
		OnExit:             "cleanup",
		ServiceAccountName: "default",
		Parallelism:        forge.Ptr(5),
		TTLStrategy: &forge.TTLStrategy{
			SecondsAfterCompletion: forge.Ptr(int(14400)), // 4h
			SecondsAfterSuccess:    forge.Ptr(int(14400)),
			SecondsAfterFailure:    forge.Ptr(int(86400)), // 24h
		},
		PodGC: &forge.PodGC{
			Strategy:            "OnPodCompletion",
			DeleteDelayDuration: "5m",
		},
		Labels: map[string]string{
			"theo.io/tenant":      tenantID,
			"theo.io/project":     projectID,
			"theo.io/environment": environment,
			"theo.io/build-id":    buildID,
		},
		Annotations: map[string]string{
			"theo.io/correlation-id": correlationID,
		},
		VolumeClaimTemplates: []forge.PVCVolume{{
			BaseVolume:       forge.BaseVolume{Name: "workspace"},
			Size:             "100Mi",
			StorageClassName: "standard", // kind default — for Theo it's "local-path"
		}},
		Volumes: []forge.VolumeBuilder{
			forge.ConfigMapVolume{
				BaseVolume:    forge.BaseVolume{Name: "kube-root-ca", MountPath: "/etc/ca"},
				ConfigMapName: "kube-root-ca.crt",
			},
		},
		SecurityContext: &model.PodSecurityContext{
			FSGroup:    forge.Ptr(uid),
			RunAsUser:  forge.Ptr(uid),
			RunAsGroup: forge.Ptr(uid),
		},
		Templates: []forge.Templatable{
			&forge.Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo theo-shape && touch /workspace/marker"},
				VolumeMounts: []forge.VolumeBuilder{
					forge.PVCVolume{BaseVolume: forge.BaseVolume{Name: "workspace", MountPath: "/workspace"}},
				},
			},
			&forge.Container{
				Name:    "cleanup",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"cleanup ran on exit"},
			},
		},
	}

	// Build asserts (in-Go, before cluster):
	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if wf.Spec.Affinity == nil || wf.Spec.Affinity.PodAffinity == nil {
		t.Fatalf("Build did NOT inject default podAffinity (v0.4.0 PR #12 regression)")
	}
	if wf.Spec.TTLStrategy == nil || wf.Spec.TTLStrategy.SecondsAfterCompletion == nil {
		t.Fatalf("TTLStrategy lost during Build")
	}
	if wf.Spec.PodGC == nil || wf.Spec.PodGC.Strategy != "OnPodCompletion" {
		t.Fatalf("PodGC lost or wrong: %+v", wf.Spec.PodGC)
	}
	if wf.Spec.SecurityContext == nil || wf.Spec.SecurityContext.FSGroup == nil {
		t.Fatalf("SecurityContext.FSGroup lost")
	}

	// Submit to cluster.
	svc := argoClient(t)
	if _, err := svc.CreateWorkflowFromModel(context.Background(), wf, ""); err != nil {
		t.Fatalf("Create: %v", err)
	}
	_ = waitWorkflowSucceeded(t, name, defaultWaitTimeout)

	// Cluster-side asserts: pull manifest BACK and confirm controller
	// did not strip any field. This is the "Theo Cloud will not fail"
	// guarantee — what we built is what runs.
	manifest := dumpWorkflow(t, name)
	mustContain := []string{
		"theo.io/correlation-id",
		"theo.io/tenant",
		"strategy: OnPodCompletion",
		"deleteDelayDuration: 5m",
		"secondsAfterCompletion: 14400",
		"secondsAfterFailure: 86400",
		"fsGroup: 1000",
		"runAsUser: 1000",
		"workflows.argoproj.io/workflow", // podAffinity term label
		"kubernetes.io/hostname",         // podAffinity topologyKey
	}
	for _, needle := range mustContain {
		if !strings.Contains(manifest, needle) {
			t.Errorf("cluster manifest missing %q — Theo-shape regression\n%s",
				needle, manifest)
		}
	}
}

// TestTheoShape_GenerateName proves Theo's GenerateName pattern works.
// Theo uses `theo-build-{BuildID}-` and Argo appends a random suffix.
// The SDK MUST omit Name and emit only generateName.
func TestTheoShape_GenerateName(t *testing.T) {
	prefix := fmt.Sprintf("theo-build-%s-", uniqueName("g")) // ends in -

	w := &forge.Workflow{
		GenerateName: prefix,
		Entrypoint:   "main",
		Templates: []forge.Templatable{
			&forge.Container{Name: "main", Image: "alpine:3.18", Command: []string{"true"}},
		},
	}
	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if wf.Metadata.Name != "" {
		t.Errorf("Name was set on the model when GenerateName-only was used: %q", wf.Metadata.Name)
	}
	if wf.Metadata.GenerateName != prefix {
		t.Errorf("GenerateName lost: got %q, want %q", wf.Metadata.GenerateName, prefix)
	}

	svc := argoClient(t)
	created, err := svc.CreateWorkflowFromModel(context.Background(), wf, "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Server should have minted a name like "theo-build-XXXX-<random>".
	if !strings.HasPrefix(created.Metadata.Name, prefix) {
		t.Fatalf("server-minted name %q does not start with %q",
			created.Metadata.Name, prefix)
	}
	// Cleanup the minted name (not the prefix).
	cleanupWorkflow(t, created.Metadata.Name)

	_ = waitWorkflowSucceeded(t, created.Metadata.Name, defaultWaitTimeout)
}

// TestTheoShape_AffinityPreservedAcrossSerialize is a regression guard
// against losing the default podAffinity when the workflow goes through
// ToYAML → FromYAML (the path Theo uses when serializing for the
// dynamic client via ToDict).
func TestTheoShape_AffinityPreservedAcrossSerialize(t *testing.T) {
	w := &forge.Workflow{
		Name:       "irrelevant",
		Entrypoint: "main",
		VolumeClaimTemplates: []forge.PVCVolume{{
			BaseVolume: forge.BaseVolume{Name: "ws"},
			Size:       "10Mi",
		}},
		Templates: []forge.Templatable{
			&forge.Container{Name: "main", Image: "alpine:3.18", Command: []string{"true"}},
		},
	}
	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	originalAffinity, _ := json.Marshal(wf.Spec.Affinity)
	if len(originalAffinity) == 0 || string(originalAffinity) == "null" {
		t.Fatalf("Build did not inject affinity for PVC workflow")
	}

	// Theo's ToDict path: marshal to JSON, hand to dynamic client. We
	// emulate by JSON round-tripping the whole WorkflowModel.
	raw, err := json.Marshal(wf)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var back model.WorkflowModel
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	backAffinity, _ := json.Marshal(back.Spec.Affinity)
	if string(backAffinity) != string(originalAffinity) {
		t.Fatalf("affinity changed across JSON round-trip:\nbefore: %s\nafter:  %s",
			originalAffinity, backAffinity)
	}
}
