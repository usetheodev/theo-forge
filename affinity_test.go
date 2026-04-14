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
