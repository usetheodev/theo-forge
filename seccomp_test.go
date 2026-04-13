package forge

import (
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

func TestSeccompProfile_Roundtrip(t *testing.T) {
	w := &Workflow{
		Name:       "seccomp-test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{
				Name:  "main",
				Image: "alpine:3.18",
				SecurityContext: &model.SecurityContext{
					SeccompProfile: &model.SeccompProfile{Type: "Unconfined"},
				},
			},
		},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}

	if !strings.Contains(yamlStr, "seccompProfile") {
		t.Errorf("YAML missing seccompProfile:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "type: Unconfined") {
		t.Errorf("YAML missing type: Unconfined:\n%s", yamlStr)
	}

	// Roundtrip
	wf, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}
	sc := wf.Spec.Templates[0].Container.SecurityContext
	if sc == nil || sc.SeccompProfile == nil {
		t.Fatal("SecurityContext or SeccompProfile is nil after roundtrip")
	}
	if sc.SeccompProfile.Type != "Unconfined" {
		t.Errorf("SeccompProfile.Type = %q, want Unconfined", sc.SeccompProfile.Type)
	}
}

func TestCapabilities_Roundtrip(t *testing.T) {
	w := &Workflow{
		Name:       "caps-test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{
				Name:  "main",
				Image: "alpine:3.18",
				SecurityContext: &model.SecurityContext{
					Capabilities: &model.Capabilities{
						Drop: []string{"ALL"},
						Add:  []string{"NET_BIND_SERVICE"},
					},
				},
			},
		},
	}

	yamlStr, err := w.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}

	if !strings.Contains(yamlStr, "capabilities") {
		t.Errorf("YAML missing capabilities:\n%s", yamlStr)
	}

	wf, err := FromYAML(yamlStr)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}
	caps := wf.Spec.Templates[0].Container.SecurityContext.Capabilities
	if caps == nil {
		t.Fatal("Capabilities is nil after roundtrip")
	}
	if len(caps.Drop) != 1 || caps.Drop[0] != "ALL" {
		t.Errorf("Capabilities.Drop = %v, want [ALL]", caps.Drop)
	}
	if len(caps.Add) != 1 || caps.Add[0] != "NET_BIND_SERVICE" {
		t.Errorf("Capabilities.Add = %v, want [NET_BIND_SERVICE]", caps.Add)
	}
}
