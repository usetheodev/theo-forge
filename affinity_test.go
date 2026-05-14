package forge

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

func TestWorkflow_WithPodAffinity_Serializes(t *testing.T) {
	w := &Workflow{
		Name:       "affinity-test",
		Entrypoint: "main",
		Affinity: &model.Affinity{
			PodAffinity: &model.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []model.PodAffinityTerm{{
					LabelSelector: &model.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					TopologyKey: "kubernetes.io/hostname",
				}},
			},
		},
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}

	if !strings.Contains(yamlStr, "podAffinity") {
		t.Errorf("YAML missing podAffinity:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "topologyKey: kubernetes.io/hostname") {
		t.Errorf("YAML missing topologyKey:\n%s", yamlStr)
	}
}

func TestNodeAffinity_Roundtrip(t *testing.T) {
	w := &Workflow{
		Name:       "node-affinity-test",
		Entrypoint: "main",
		Affinity: &model.Affinity{
			NodeAffinity: &model.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &model.NodeSelector{
					NodeSelectorTerms: []model.NodeSelectorTerm{{
						MatchExpressions: []model.NodeSelectorRequirement{{
							Key:      "disktype",
							Operator: "In",
							Values:   []string{"ssd"},
						}},
					}},
				},
			},
		},
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}

	wf, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}

	na := wf.Spec.Affinity
	if na == nil || na.NodeAffinity == nil {
		t.Fatal("NodeAffinity is nil after roundtrip")
	}
	req := na.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if req == nil || len(req.NodeSelectorTerms) == 0 {
		t.Fatal("NodeSelectorTerms empty after roundtrip")
	}
	me := req.NodeSelectorTerms[0].MatchExpressions
	if len(me) != 1 || me[0].Key != "disktype" || me[0].Operator != "In" {
		t.Errorf("MatchExpressions = %+v", me)
	}
}

func TestContainerTemplate_WithAffinity(t *testing.T) {
	c := &Container{
		Name:  "with-affinity",
		Image: "alpine:3.18",
		Affinity: &model.Affinity{
			PodAntiAffinity: &model.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []model.WeightedPodAffinityTerm{{
					Weight: 100,
					PodAffinityTerm: model.PodAffinityTerm{
						TopologyKey: "kubernetes.io/hostname",
					},
				}},
			},
		},
	}

	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatalf("BuildTemplate: %v", err)
	}

	if tpl.Affinity == nil || tpl.Affinity.PodAntiAffinity == nil {
		t.Fatal("Affinity not wired in template")
	}
}

// ---------------------------------------------------------------------------
// Default podAffinity for build workflows (issue #11).
// These tests describe the behavior of DefaultPodAffinityFor(w) and the
// resolveAffinity gate inside Workflow.Build().
// ---------------------------------------------------------------------------

func TestDefaultPodAffinityFor_Nil_ReturnsNil(t *testing.T) {
	if got := DefaultPodAffinityFor(nil); got != nil {
		t.Errorf("DefaultPodAffinityFor(nil) = %+v, want nil", got)
	}
}

func TestDefaultPodAffinityFor_NoName_NoGenerateName_ReturnsNil(t *testing.T) {
	w := &Workflow{}
	if got := DefaultPodAffinityFor(w); got != nil {
		t.Errorf("DefaultPodAffinityFor(empty) = %+v, want nil", got)
	}
}

func TestDefaultPodAffinityFor_NameSet_UsesLiteralName(t *testing.T) {
	w := &Workflow{Name: "build-42"}
	got := DefaultPodAffinityFor(w)
	if got == nil || got.PodAffinity == nil {
		t.Fatalf("expected PodAffinity, got %+v", got)
	}
	terms := got.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if len(terms) != 1 {
		t.Fatalf("expected 1 term, got %d", len(terms))
	}
	labels := terms[0].LabelSelector.MatchLabels
	if labels["workflows.argoproj.io/workflow"] != "build-42" {
		t.Errorf("expected literal Name in label, got %q", labels["workflows.argoproj.io/workflow"])
	}
	if terms[0].TopologyKey != "kubernetes.io/hostname" {
		t.Errorf("expected hostname topology, got %q", terms[0].TopologyKey)
	}
}

func TestDefaultPodAffinityFor_GenerateNameOnly_UsesTemplateVar(t *testing.T) {
	w := &Workflow{GenerateName: "build-"}
	got := DefaultPodAffinityFor(w)
	if got == nil || got.PodAffinity == nil {
		t.Fatalf("expected PodAffinity, got %+v", got)
	}
	terms := got.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	labels := terms[0].LabelSelector.MatchLabels
	if labels["workflows.argoproj.io/workflow"] != "{{workflow.name}}" {
		t.Errorf("expected {{workflow.name}} template var, got %q", labels["workflows.argoproj.io/workflow"])
	}
}

func TestBuild_DefaultAffinity_InjectedWhenVCTAndNilAffinity(t *testing.T) {
	w := &Workflow{
		Name:       "with-pvc",
		Entrypoint: "main",
		VolumeClaimTemplates: []PVCVolume{{
			BaseVolume:  BaseVolume{Name: "scratch", MountPath: "/scratch"},
			Size:        "1Gi",
			AccessModes: []AccessMode{ReadWriteOnce},
		}},
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}
	m, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if m.Spec.Affinity == nil || m.Spec.Affinity.PodAffinity == nil {
		t.Fatalf("expected default PodAffinity injected, got Affinity=%+v", m.Spec.Affinity)
	}
	terms := m.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if len(terms) != 1 {
		t.Fatalf("expected 1 term, got %d", len(terms))
	}
	if terms[0].LabelSelector.MatchLabels["workflows.argoproj.io/workflow"] != "with-pvc" {
		t.Errorf("expected label value 'with-pvc', got %q", terms[0].LabelSelector.MatchLabels["workflows.argoproj.io/workflow"])
	}
}

func TestBuild_DefaultAffinity_NotInjectedWhenVCTEmpty(t *testing.T) {
	w := &Workflow{
		Name:       "no-pvc",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}
	m, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if m.Spec.Affinity != nil {
		t.Errorf("expected nil Affinity (no VCT), got %+v", m.Spec.Affinity)
	}
}

func TestBuild_DefaultAffinity_NotInjectedWhenAffinitySet(t *testing.T) {
	custom := &model.Affinity{
		NodeAffinity: &model.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &model.NodeSelector{
				NodeSelectorTerms: []model.NodeSelectorTerm{{
					MatchExpressions: []model.NodeSelectorRequirement{{
						Key: "zone", Operator: "In", Values: []string{"a"},
					}},
				}},
			},
		},
	}
	w := &Workflow{
		Name:       "custom-affinity",
		Entrypoint: "main",
		Affinity:   custom,
		VolumeClaimTemplates: []PVCVolume{{
			BaseVolume:  BaseVolume{Name: "scratch", MountPath: "/scratch"},
			Size:        "1Gi",
			AccessModes: []AccessMode{ReadWriteOnce},
		}},
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}
	m, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if m.Spec.Affinity == nil || m.Spec.Affinity.NodeAffinity == nil {
		t.Fatalf("expected user-supplied NodeAffinity preserved, got %+v", m.Spec.Affinity)
	}
	if m.Spec.Affinity.PodAffinity != nil {
		t.Errorf("expected no PodAffinity injection when user supplied Affinity, got %+v", m.Spec.Affinity.PodAffinity)
	}
}

func TestBuild_DefaultAffinity_NotInjectedWhenOptOutTrue(t *testing.T) {
	w := &Workflow{
		Name:                   "opt-out",
		Entrypoint:             "main",
		DisableDefaultAffinity: true,
		VolumeClaimTemplates: []PVCVolume{{
			BaseVolume:  BaseVolume{Name: "scratch", MountPath: "/scratch"},
			Size:        "1Gi",
			AccessModes: []AccessMode{ReadWriteOnce},
		}},
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}
	m, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if m.Spec.Affinity != nil {
		t.Errorf("expected nil Affinity (opt-out), got %+v", m.Spec.Affinity)
	}
}

func TestColocateByLabel_MatchesTheoJSON(t *testing.T) {
	// The Theo PodSpecPatch wraps the affinity: {"affinity":{...}}.
	// ColocateByLabel returns the inner *Affinity object. We compare the inner part.
	theoAffinityJSON := `{"podAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":[{"labelSelector":{"matchLabels":{"workflows.argoproj.io/workflow":"{{workflow.name}}"}},"topologyKey":"kubernetes.io/hostname"}]}}`

	affinity := ColocateByLabel(
		"workflows.argoproj.io/workflow",
		"{{workflow.name}}",
		"kubernetes.io/hostname",
	)

	got, err := json.Marshal(affinity)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Compare as maps to ignore field ordering
	var expected, actual map[string]interface{}
	if err := json.Unmarshal([]byte(theoAffinityJSON), &expected); err != nil {
		t.Fatalf("unmarshal expected: %v", err)
	}
	if err := json.Unmarshal(got, &actual); err != nil {
		t.Fatalf("unmarshal actual: %v", err)
	}

	expectedBytes, _ := json.Marshal(expected)
	actualBytes, _ := json.Marshal(actual)
	if !bytes.Equal(expectedBytes, actualBytes) {
		t.Errorf("ColocateByLabel output does not match Theo affinity.\nExpected: %s\nGot:      %s", expectedBytes, actualBytes)
	}
}
